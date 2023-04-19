package generate

import (
	"context"
	"fmt"
	"io/fs"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/modernice/opendocs/find"
	"github.com/modernice/opendocs/internal"
	"golang.org/x/exp/slog"
)

// Service represents a service that can generate documentation for a given
// identifier. It has a single method, GenerateDoc, which takes a Context and
// returns the generated documentation as a string or an error. The Context
// interface provides methods to retrieve information about the identifier and
// the file(s) being documented.
type Service interface {
	GenerateDoc(Context) (string, error)
}

// Context is an interface that extends the context.Context interface. It
// provides additional methods for generating documentation, including
// Identifier, File, Files, and Read.
type Context interface {
	context.Context

	Identifier() string
	File() string
	Files() []string
	Read(file string) ([]byte, error)
}

// File represents a file and its generated documentation. It contains the
// file's path and a slice of Generations, which represent the documentation
// generated for each identifier found in the file.
type File struct {
	Path        string
	Generations []Generation
}

// Generation is a package that provides a Generator type for generating
// documentation for Go code. The Generator type has a Generate method that
// takes a context, a file system, and options. It returns two channels: one for
// File objects and one for errors. The File object represents a file that was
// processed and contains a slice of Generation objects. The Generation object
// represents a generated documentation for a specific identifier in a file. The
// Generator type also has several Option functions that can be used to
// configure the generation process, such as Limit, FileLimit, WithLogger,
// Footer, and FindWith.
type Generation struct {
	File       string
	Identifier string
	Doc        string
}

// Generator is a type that generates documentation for Go code. It has a
// Generate method that takes a context, a file system, and options. The
// Generate method returns two channels: one for the generated files and one for
// errors. The Generator type also has a New function that creates a new
// instance of the type. The Generator type uses a Service interface to generate
// documentation.
type Generator struct {
	svc Service
}

// New returns a new *Generator that uses the given Service to generate
// documentation.
func New(svc Service) *Generator {
	return &Generator{svc: svc}
}

// Option is a type that represents a configuration option for a Generator. It
// is used to modify the behavior of the Generator when generating
// documentation. Option is implemented as a function that takes a pointer to a
// generation struct and modifies its fields. The available options are Limit,
// FileLimit, WithLogger, Footer, and FindWith.
type Option func(*generation)

// Limit sets the maximum number of generated files. It returns an Option.
func Limit(n int) Option {
	return func(g *generation) {
		g.limit = n
	}
}

// FileLimit is an Option that sets the maximum number of files to generate
// documentation for. It is used with the Generator type to limit the number of
// files that are processed during documentation generation.
func FileLimit(n int) Option {
	return func(g *generation) {
		g.fileLimit = n
	}
}

// WithLogger is an Option for Generator that sets the logger for the generation
// process. The logger is used to log debug, info, and error messages during the
// generation process. The logger must implement the slog.Handler interface.
func WithLogger(h slog.Handler) Option {
	return func(g *generation) {
		g.log = slog.New(h)
	}
}

// Footer is an Option for Generator that sets a message to be appended to the
// end of each generated documentation. It takes a string as its argument.
func Footer(msg string) Option {
	return func(g *generation) {
		g.footer = msg
	}
}

func Override(override bool) Option {
	return func(g *generation) {
		g.override = override
	}
}

// FindWith adds options to a Generator that are used to configure the search
// for code to generate documentation for. It takes one or more find.Options and
// returns an Option that can be passed to a Generator.
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

// Generate generates documentation for the code found in a file system. It
// takes a context, a file system, and options. It returns two channels: one for
// the generated files and one for errors. The generated files are of type File,
// which contains the path to the file and a slice of Generations. Each
// Generation contains the file name, identifier, and documentation. The options
// include Limit, FileLimit, WithLogger, Footer, and FindWith.
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
