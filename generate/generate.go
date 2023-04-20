package generate

import (
	"context"
	"fmt"
	"io/fs"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/modernice/jotbot/find"
	"github.com/modernice/jotbot/internal"
	"golang.org/x/exp/slog"
)

// Service represents a code documentation generator. It defines a single
// method, GenerateDoc, which takes a Context and returns a string of generated
// documentation or an error. The Context interface provides methods for
// retrieving information about the code being documented, including the
// identifier and file path.
type Service interface {
	GenerateDoc(Context) (string, error)
}

// Context defines the interface for a context used in generating documentation.
// It extends the standard library's context.Context interface and adds methods
// for retrieving the identifier and file associated with the context, as well
// as methods for reading the contents of a file.
type Context interface {
	context.Context

	Identifier() string
	File() string
	Files() []string
	Read(file string) ([]byte, error)
}

// File represents a file and its generated documentation.
type File struct {
	Path        string
	Generations []Generation
}

// Generation represents a code generation service. It provides a Generate
// method that generates documentation for a given repository, and returns a
// channel of generated files and errors. It also provides options to limit the
// number of files or generations that are generated, to override existing
// documentation, to add a footer to each generated document, and to configure
// the logger.
type Generation struct {
	File       string
	Identifier string
	Doc        string
}

// Generator generates documentation for code identified by an identifier in a
// given file. It takes a Service, which defines how to generate documentation
// for a given identifier, and produces a channel of Files and errors. The
// generator can be configured with various Options, such as a limit on the
// number of generated files, a limit on the number of workers, and a custom
// footer to append to each generated doc.
type Generator struct {
	svc Service
}

// New creates a new Generator that generates documentation using a Service.
func New(svc Service) *Generator {
	return &Generator{svc: svc}
}

// Option is a function type that modifies the behavior of a Generator. It is
// used as an argument to the Generate method of a Generator to customize the
// generation process. A few predefined Options are available, such as Limit,
// FileLimit, WithLogger, Footer, Override, and FindWith. The Limit option
// limits the number of generated files, while the FileLimit option limits the
// number of concurrently processed files. The WithLogger option sets a logger
// for the generator, while the Footer option adds a footer to each generated
// file. The Override option determines whether to generate documentation for
// commented and uncommented identifiers, and the FindWith option is used to
// pass options to the find package that is used in the generation process.
type Option func(*generation)

// Limit sets the maximum number of Generations to generate per worker.
func Limit(n int) Option {
	return func(g *generation) {
		g.limit = n
	}
}

// FileLimit specifies the maximum number of files that should be processed
// during code generation for a given identifier.
func FileLimit(n int) Option {
	return func(g *generation) {
		g.fileLimit = n
	}
}

// WithLogger sets the logger for the Generator. It returns an Option that can
// be passed to New or Generate methods.
func WithLogger(h slog.Handler) Option {
	return func(g *generation) {
		g.log = slog.New(h)
	}
}

// Footer adds a footer message to the generated documentation. It is an Option
// for the Generator type.
func Footer(msg string) Option {
	return func(g *generation) {
		g.footer = msg
	}
}

// Override sets the override flag to either true or false.
func Override(override bool) Option {
	return func(g *generation) {
		g.override = override
	}
}

// FindWith returns an Option that appends find.Options to the generation's
// findOpts slice. This slice is used in the Generation's GenerateFile method to
// configure Findings with identifiers and files to generate documentation for.
func FindWith(opts ...find.Option) Option {
	return func(g *generation) {
		g.findOpts = append(g.findOpts, opts...)
	}
}

type generation struct {
	limit     int
	fileLimit int
	footer    string
	override  bool
	findOpts  []find.Option
	log       *slog.Logger
}

// Generate generates documentation for the given identifier. It takes a
// context, a file system, and options as input. It returns a channel of Files
// and a channel of errors. Use the File struct to access the Path and
// Generations of the generated documentation. Use the Generation struct to
// access the File, Identifier, and Doc of each generated documentation.
func (g *Generator) Generate(ctx context.Context, repo fs.FS, opts ...Option) (<-chan File, <-chan error, error) {
	var cfg generation
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.log == nil {
		cfg.log = internal.NopLogger()
	}

	findOpts := append([]find.Option{find.WithLogger(cfg.log.Handler())}, cfg.findOpts...)

	var (
		result find.Findings
		err    error
	)

	if cfg.override {
		result, err = find.New(repo, findOpts...).All()
	} else {
		result, err = find.New(repo, findOpts...).Uncommented()
	}
	if err != nil {
		return nil, nil, fmt.Errorf("find identifiers: %w", err)
	}

	files, errs := make(chan File), make(chan error)

	ctx, cancel := context.WithCancel(ctx)

	var (
		nFiles     int64
		nGenerated int64
	)

	canGenerate := func() bool {
		ng := atomic.LoadInt64(&nGenerated)
		if cfg.limit > 0 && ng >= int64(cfg.limit) {
			return false
		}
		return true
	}

	onGenerated := func() { atomic.AddInt64(&nGenerated, 1) }

	push := func(file string, gens []Generation) bool {
		select {
		case <-ctx.Done():
			return false
		case files <- File{file, gens}:
			atomic.AddInt64(&nFiles, 1)
			return true
		}
	}

	fail := func(err error) {
		select {
		case <-ctx.Done():
		case errs <- err:
		}
	}

	workers := runtime.NumCPU()
	if cfg.fileLimit > 0 && cfg.fileLimit < workers {
		cfg.log.Debug(fmt.Sprintf("File limit (%d) is lower than number of workers (%d). Reducing workers to %d.", cfg.fileLimit, workers, cfg.fileLimit))
		workers = cfg.fileLimit
	}

	queue := make(chan string)
	var wg sync.WaitGroup
	wg.Add(workers)

	go func() {
		defer cancel()
		wg.Wait()
		close(files)
		close(errs)
	}()

	go func() {
		defer close(queue)
		for file := range result {
			select {
			case <-ctx.Done():
				return
			case queue <- file:
			}
		}
	}()

	cfg.log.Debug(fmt.Sprintf("Generating docs using %d workers ...", workers))

	// root context is only used to create child contexts
	rootGenCtx, err := newCtx(ctx, repo, "", "")
	if err != nil {
		return nil, nil, fmt.Errorf("create generation context: %w", err)
	}

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for file := range queue {
				gens, err := g.generateFile(rootGenCtx, ctx, file, result[file], repo, cfg, canGenerate, onGenerated)
				if err != nil {
					fail(fmt.Errorf("generate %s: %w", file, err))
					return
				}

				if !push(file, gens) {
					return
				}

				if n := atomic.LoadInt64(&nFiles); cfg.fileLimit > 0 && n >= int64(cfg.fileLimit) {
					return
				}
			}
		}()
	}

	return files, errs, nil
}

func (g *Generator) generateFile(
	ctx *genCtx,
	parent context.Context,
	file string,
	findings []find.Finding,
	repo fs.FS,
	cfg generation,
	canGenerate func() bool,
	onGenerated func(),
) ([]Generation, error) {
	var generations []Generation

	var err error
	for _, finding := range findings {
		if !canGenerate() {
			break
		}

		if ctx, err = g.context(ctx, parent, repo, finding); err != nil {
			return generations, fmt.Errorf("create context for %s: %w", finding, err)
		}

		gen, err := g.generateFinding(ctx, cfg, finding)
		if err != nil {
			return generations, err
		}

		if canGenerate() {
			generations = append(generations, gen)
		}

		onGenerated()
	}

	return generations, nil
}

func (g *Generator) context(ctx *genCtx, parent context.Context, repo fs.FS, finding find.Finding) (*genCtx, error) {
	if ctx == nil {
		return newCtx(parent, repo, finding.Path, finding.Identifier)
	}
	return ctx.new(parent, finding.Path, finding.Identifier), nil
}

func (g *Generator) generateFinding(ctx *genCtx, cfg generation, finding find.Finding) (Generation, error) {
	cfg.log.Info("Generating docs ...", "path", finding.Path, "identifier", finding.Identifier)

	doc, err := g.svc.GenerateDoc(ctx)
	if err != nil {
		return Generation{}, fmt.Errorf("generate %s: generate doc: %w", finding, err)
	}

	if cfg.footer != "" {
		doc = fmt.Sprintf("%s\n\n%s", doc, cfg.footer)
	}

	return Generation{
		File:       finding.Path,
		Identifier: finding.Identifier,
		Doc:        doc,
	}, nil
}
