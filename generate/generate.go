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

// Service is an interface that defines the method GenerateDoc. A type that
// implements Service can generate documentation for a Go code identifier. The
// documentation is returned as a string and an error. The Context argument
// passed to GenerateDoc specifies the identifier for which the documentation
// should be generated, as well as the files in which it appears.
type Service interface {
	GenerateDoc(Context) (string, error)
}

// Context [Context](https://golang.org/pkg/context/) is an interface that
// extends the `context.Context` interface. It represents a context for
// generating documentation for a specific identifier in a file. It provides
// methods to retrieve the identifier and file name, as well as to read the
// contents of a file.
type Context interface {
	context.Context

	Identifier() string
	File() string
	Files() []string
	Read(file string) ([]byte, error)
}

// File represents a file and its generated documentation. It contains the path
// to the file and a slice of Generations, which represent the generated
// documentation for each identifier found in the file. A Generator can be used
// to generate documentation for one or more files using the Generate method.
// The documentation can be customized using various Option functions, such as
// Limit, FileLimit, WithLogger, Footer, Override, and FindWith. The generated
// documentation is sent to a channel of Files, which can be read from until it
// is closed.
type File struct {
	Path        string
	Generations []Generation
}

// Generation represents a generated documentation for a given identifier. It
// contains information about the file in which the identifier was found, the
// identifier itself, and the generated documentation.
type Generation struct {
	File       string
	Identifier string
	Doc        string
}

// Generator is a type that generates documentation for Go code. It provides a
// method to generate documentation for a given context by calling the
// GenerateDoc method on a Service. The Generate method takes a context, a file
// system, and optional generation options, and returns channels of generated
// files and errors.
type Generator struct {
	svc Service
}

// New returns a new *Generator that uses the provided Service to generate
// documentation.
func New(svc Service) *Generator {
	return &Generator{svc: svc}
}

// Option is a type that represents a configuration option for generating
// documentation. It is used as an argument to the Generate method of Generator.
// The following functions return Option values: Limit, FileLimit, WithLogger,
// Footer, Override, FindWith.
type Option func(*generation)

// Limit returns an Option that sets the maximum number of generated files. If n
// is 0, no limit is set. The default is no limit.
func Limit(n int) Option {
	return func(g *generation) {
		g.limit = n
	}
}

// FileLimit specifies an Option for limiting the number of files that should be
// processed by a Generator. It is used in conjunction with New to create a new
// Generator instance.
func FileLimit(n int) Option {
	return func(g *generation) {
		g.fileLimit = n
	}
}

// WithLogger returns an Option that sets the logger that WithLogger will use
// for logging. The logger must implement slog.Handler.
func WithLogger(h slog.Handler) Option {
	return func(g *generation) {
		g.log = slog.New(h)
	}
}

// Footer is an Option that can be passed to the New function of a Generator. It
// sets a message that will be appended to the generated documentation for each
// identifier.
func Footer(msg string) Option {
	return func(g *generation) {
		g.footer = msg
	}
}

// Override sets the option to override existing documentation. It is a function
// that takes a boolean value and returns an Option.
func Override(override bool) Option {
	return func(g *generation) {
		g.override = override
	}
}

// FindWith returns an Option that can be passed to Generator.Generate to
// configure the Find operation used to identify code elements for documentation
// generation. The Option takes one or more find.Options from the
// "github.com/modernice/jotbot/find" package.
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

// Generate generates documentation for Go code by analyzing the code in a given
// file system [fs.FS] and generating documentation for each identifier that is
// found. The generated documentation is based on the result of calling the
// GenerateDoc method of the service [Service] provided to the Generator. The
// identifiers are found using the find package, which allows for various
// options such as filtering out commented code or limiting the number of
// generated files. The generation can be further customized using options
// [Option], such as setting a limit on the number of generated identifiers or
// providing a custom footer for each generated document. The method returns two
// channels: one that streams all generated files [<-chan File] and one that
// streams any errors that occurred during generation [<-chan error].
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
					cfg.log.Debug(fmt.Sprintf("generate %s: %v", file, err))
					cfg.log.Warn(fmt.Sprintf("Generation of %s failed. Skipping.", file))
					continue
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
