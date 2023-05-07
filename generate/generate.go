package generate

//go:generate go-mockgen -f github.com/modernice/jotbot/generate -i Service -i Minifier -o ./mockgenerate/generate.go

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/modernice/jotbot/internal"
	"golang.org/x/exp/slog"
)

var (
	// DefaultFileWorkers is the default number of concurrent workers for processing
	// files. It is set to the minimum of 4 and the number of available CPU cores.
	DefaultFileWorkers = int(math.Min(4, float64(runtime.NumCPU())))

	DefaultSymbolWorkers = int(math.Min(2, float64(runtime.NumCPU())))
)

// Service is an interface that provides a method to generate documentation
// strings for a given [Context]. It returns the generated documentation string
// and an error if there is any issue during the generation process.
type Service interface {
	// GenerateDoc generates a documentation string for the given context, using the
	// associated Service. It returns the generated documentation string and an
	// error if there is any issue during the generation process.
	GenerateDoc(Context) (string, error)
}

// Language is an interface that defines the Prompt method which returns a
// formatted string using the provided Input for a Language implementation.
type Language interface {
	// Prompt returns a formatted string using the provided Input for a Language
	// implementation.
	Prompt(Input) string
}

// Minifier is an interface that provides a method to minify a byte slice,
// returning the minified byte slice and an error if the minification process
// fails.
type Minifier interface {
	// Minify takes a byte slice, minifies its contents, and returns the minified
	// byte slice and an error if the minification process fails.
	Minify([]byte) ([]byte, error)
}

// Input represents a single unit of code for which documentation needs to be
// generated, including the code itself, its language, and an identifier.
type Input struct {
	Code       []byte
	Language   string
	Identifier string
}

// String returns a formatted string representation of the Input, including its
// Identifier and Language.
func (f Input) String() string {
	return fmt.Sprintf("%s (%s)", f.Identifier, f.Language)
}

// Context is an interface that extends the standard [context.Context] and
// provides additional methods to retrieve the associated [Input] and prompt
// string for the current context. It is used during the documentation
// generation process.
type Context interface {
	context.Context

	// Input returns the Input associated with the Context, which contains code,
	// language, and identifier information.
	Input() Input

	// Prompt returns the prompt string for the current [Context].
	Prompt() string
}

// File represents a documentation file with a path and a slice of
// [Documentation] objects, where each Documentation object contains the
// generated documentation for a code symbol in the file.
type File struct {
	Path string
	Docs []Documentation
}

// Documentation represents a generated documentation string with its associated
// Input, which includes code, language, and identifier information.
type Documentation struct {
	Input

	Text string
}

// Generator is a configurable concurrent documentation generator that uses a
// Service to generate documentation strings for multiple files and languages,
// respecting context cancellation. It supports setting custom slog.Handler for
// logging, limiting the number of files to process, customizing the number of
// workers for file and symbol processing, and registering Language
// implementations for specific file extensions.
type Generator struct {
	svc           Service
	languages     map[string]Language
	limit         int
	fileWorkers   int
	symbolWorkers int
	footer        string
	log           *slog.Logger
}

// Option is a functional option type for configuring a Generator. It allows
// setting custom logger, footer, limits, workers, and languages for the
// Generator, enabling flexible customization while maintaining a concise API.
type Option func(*Generator)

// WithLogger sets a custom slog.Handler for logging in the Generator.
func WithLogger(h slog.Handler) Option {
	return func(g *Generator) {
		g.log = slog.New(h)
	}
}

// Footer sets the footer string for the generated documentation. The provided
// message will be appended to the end of each generated document.
func Footer(msg string) Option {
	return func(g *Generator) {
		g.footer = msg
	}
}

// Limit sets the maximum number of files to generate documentation for. If the
// limit is set to a positive value, only the specified number of files will be
// processed. If the limit is set to a non-positive value, all files will be
// processed.
func Limit(n int) Option {
	return func(g *Generator) {
		g.limit = n
	}
}

// Workers sets the number of file workers and symbol workers for a Generator.
// The fileWorkers argument specifies the number of files to be generated in
// parallel, while the symbolWorkers argument specifies the number of symbols to
// be generated concurrently within each file.
func Workers(files, symbols int) Option {
	return func(g *Generator) {
		g.fileWorkers = files
		g.symbolWorkers = symbols
	}
}

// WithLanguage sets a language for the specified extension in the Generator,
// allowing the generator to recognize and process files with that extension
// using the provided Language implementation.
func WithLanguage(ext string, lang Language) Option {
	return func(g *Generator) {
		g.languages[ext] = lang
	}
}

// New creates a new Generator with the given Service and options. The Generator
// can generate documentation for multiple files concurrently, using a
// configurable number of workers for file and symbol processing.
func New(svc Service, opts ...Option) *Generator {
	g := &Generator{
		svc:           svc,
		fileWorkers:   DefaultFileWorkers,
		symbolWorkers: DefaultSymbolWorkers,
		languages:     make(map[string]Language),
	}
	for _, opt := range opts {
		opt(g)
	}
	if g.fileWorkers <= 0 {
		g.fileWorkers = 1
	}
	if g.symbolWorkers <= 0 {
		g.symbolWorkers = 1
	}
	if g.log == nil {
		g.log = internal.NopLogger()
	}
	return g
}

// Files generates documentation for the provided files concurrently, returning
// a channel of generated File objects and a channel of errors. It respects the
// context cancellation and returns an error if the context is canceled.
func (g *Generator) Files(ctx context.Context, files map[string][]Input) (<-chan File, <-chan error, error) {
	out, errs := make(chan File), make(chan error)

	push := func(f File) bool {
		select {
		case <-ctx.Done():
			return false
		case out <- f:
			return true
		}
	}

	fail := func(err error) {
		select {
		case <-ctx.Done():
		case errs <- err:
		}
	}

	work, done := g.distributeWork(files)
	go work(ctx, func(file string, inputs []Input) bool {
		docs := make(chan Documentation)

		queue := make(chan Input)
		go func() {
			defer close(queue)
			for _, input := range inputs {
				select {
				case <-ctx.Done():
					return
				case queue <- input:
				}
			}
		}()

		var wg sync.WaitGroup
		wg.Add(g.symbolWorkers)
		go func() {
			wg.Wait()
			close(docs)
		}()

		for i := 0; i < g.symbolWorkers; i++ {
			go func() {
				defer wg.Done()
				for input := range queue {
					g.log.Info(fmt.Sprintf("Generating %s ...", input))

					doc, err := g.Generate(ctx, input)
					if err != nil {
						fail(fmt.Errorf("generate %q: %w", input.Identifier, err))
						continue
					}

					select {
					case <-ctx.Done():
						return
					case docs <- Documentation{Input: input, Text: doc}:
					}
				}
			}()
		}

		result, err := internal.Drain(docs, nil)
		if err != nil {
			fail(err)
			return true
		}

		return push(File{Path: file, Docs: result})
	})

	go func() {
		<-done
		close(out)
		close(errs)
	}()

	return out, errs, nil
}

func (g *Generator) distributeWork(files map[string][]Input) (func(context.Context, func(string, []Input) bool), <-chan struct{}) {
	done := make(chan struct{})
	return func(ctx context.Context, work func(string, []Input) bool) {
		workers := g.fileWorkers
		if workers < 1 {
			g.log.Debug(fmt.Sprintf("Invalid worker count %d. Setting workers to 1", workers))
			workers = 1
		}
		if workers > len(files) {
			g.log.Debug(fmt.Sprintf("Setting workers to file count: %d", len(files)))
			workers = len(files)
		}
		if g.limit > 0 && workers > g.limit {
			g.log.Debug(fmt.Sprintf("Setting workers to file limit: %d", g.limit))
			workers = g.limit
		}

		if workers > 1 {
			g.log.Debug(fmt.Sprintf("Generating %d files in parallel.", workers))
		}

		type job struct {
			file   string
			inputs []Input
		}

		queue := make(chan job)
		go func() {
			defer close(queue)
			for file, inputs := range files {
				select {
				case <-ctx.Done():
					return
				case <-done:
					return
				case queue <- job{file: file, inputs: inputs}:
				}
			}
		}()

		var wg sync.WaitGroup
		wg.Add(workers)

		var nFiles atomic.Int64

		for i := 0; i < workers; i++ {
			go func() {
				defer wg.Done()
				select {
				case <-ctx.Done():
					return
				case job, ok := <-queue:
					if !ok {
						return
					}

					g.log.Info(fmt.Sprintf("Generating %s ...", job.file))

					if g.limit > 0 {
						n := nFiles.Load()
						if n >= int64(g.limit) {
							return
						}
						nFiles.Add(1)
					}

					if !work(job.file, job.inputs) {
						return
					}
				}
			}()
		}

		go func() {
			wg.Wait()
			close(done)
		}()
	}, done
}

// Generate creates a documentation string for the provided Input using the
// associated Language and Service. It returns an error if the specified
// language is not registered or if there is an issue generating the
// documentation.
func (g *Generator) Generate(ctx context.Context, input Input) (string, error) {
	lang, ok := g.languages[input.Language]
	if !ok {
		return "", fmt.Errorf("unknown language %q", input.Language)
	}

	if min, ok := lang.(Minifier); ok {
		code, err := min.Minify(input.Code)
		if err != nil {
			return "", fmt.Errorf("minify code: %w", err)
		}
		input.Code = code
	}

	genCtx := newCtx(ctx, input, lang.Prompt(input))

	doc, err := g.svc.GenerateDoc(genCtx)
	if err != nil {
		return "", fmt.Errorf("service: %w", err)
	}

	if g.footer != "" {
		doc = fmt.Sprintf("%s\n\n%s", doc, g.footer)
	}

	return doc, nil
}
