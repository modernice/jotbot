package generate

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/modernice/jotbot/internal"
	"golang.org/x/exp/slog"
)

var (
	// DefaultFileWorkers represents the default number of workers that process
	// files concurrently in the generation service. This value is used to initiate
	// the number of goroutines handling file-level tasks when no specific worker
	// count is provided. It is set based on system resources, ensuring efficient
	// utilization without overwhelming the host machine.
	DefaultFileWorkers = int(math.Min(4, float64(runtime.NumCPU())))

	// DefaultSymbolWorkers represents the default number of concurrent workers that
	// process symbols within files during documentation generation. This number is
	// determined based on the lesser of a predefined value or the number of CPU
	// cores available, ensuring efficient parallel processing without overloading
	// the system.
	DefaultSymbolWorkers = int(math.Min(2, float64(runtime.NumCPU())))
)

// Service represents the core functionality of generating documentation based
// on provided context, encapsulating the complexities of the documentation
// generation process. It accepts a context which carries metadata and
// configuration for the generation task and returns the generated documentation
// as a string along with any error that occurred during the process. This
// interface abstracts away the details of how documentation is generated,
// allowing different implementations to provide their own logic for converting
// code and associated information into human-readable documentation.
type Service interface {
	// GenerateDoc initiates the documentation generation process for a given
	// context. It returns the generated documentation as a string, or an error if
	// the generation fails. The context must provide the necessary information such
	// as input code, language specifics, and file details through the [Context]
	// interface. The actual generation logic is implemented by the [Service]
	// interface that this function is a part of.
	GenerateDoc(Context) (string, error)
}

// Language represents a mechanism for generating textual prompts based on
// structured input. It operates on the given input to produce a string that can
// be used as a directive or guide in subsequent operations. This interface is
// typically implemented by types that understand how to interpret and process
// language-specific information within the context of code generation,
// documentation automation, or similar domains where dynamic text generation is
// required.
type Language interface {
	// Prompt initiates a text-based interaction using the provided input parameters
	// and returns the resulting string. The interaction is defined by the
	// implementing language's logic, which determines how the input influences the
	// output.
	Prompt(PromptInput) string
}

// Minifier reduces the size of a given byte slice by removing unnecessary
// characters and formatting, potentially optimizing it for network transmission
// or storage. It returns the minified byte slice along with any error
// encountered during the process.
type Minifier interface {
	// Minify compresses the provided byte slice by removing unnecessary characters
	// without changing its functionality. It returns the minified byte slice and
	// any error encountered during the minification process.
	Minify([]byte) ([]byte, error)
}

// Input represents a unit of source code to be processed for documentation
// generation. It includes the raw code, the programming language of the code,
// and an identifier for referencing the specific piece of code within a larger
// context or system.
type Input struct {
	Code       []byte
	Language   string
	Identifier string
}

// String returns a formatted string representation of the Input, which includes
// its identifier and language.
func (f Input) String() string {
	return fmt.Sprintf("%s (%s)", f.Identifier, f.Language)
}

// PromptInput represents the input data required by a language-specific prompt
// to generate documentation. It encapsulates code, language settings, and file
// association details necessary for the documentation process.
type PromptInput struct {
	Input
	File string
}

// Context provides an interface for carrying deadlines, cancellation signals,
// and other request-scoped values across API boundaries and between processes.
// It extends the standard context.Context interface with additional methods to
// access specific input data and generate prompts based on that data,
// facilitating a more tailored request handling in the context of documentation
// generation.
type Context interface {
	context.Context

	// Input retrieves the current prompt input from the context. It returns a
	// [PromptInput] that contains both the code to be processed and metadata such
	// as the language and identifier of the code, as well as the associated file.
	Input() PromptInput

	// Prompt retrieves the prompt string for code generation based on the current
	// context and input parameters. It is typically used to obtain a starting point
	// or a question that guides the code generation process. The prompt is derived
	// from the contextual information encapsulated within the Context interface,
	// including details about the input source code, programming language, and file
	// associated with the operation. The resulting string can be leveraged by
	// generation services or other processes that require contextual prompts to
	// function effectively.
	Prompt() string
}

// File represents a collection of documentation entries associated with a
// specific file path. It contains the file path and a slice of Documentation,
// which encapsulates the content and context of generated documentation for
// each code symbol within the file.
type File struct {
	Path string
	Docs []Documentation
}

// Documentation represents the written explanation or clarification of the
// source code, which is intended to help developers understand the code's
// purpose and functionality. It includes both the original input source such as
// the code snippet, language identifier, and any additional text that further
// elucidates the code. This text may be generated through automated processes
// or manually composed to provide insight into complex or non-obvious aspects
// of the code. The Documentation type is typically associated with a specific
// segment of source code, making it easier for others to comprehend and
// maintain that code in the future.
type Documentation struct {
	Input

	Text string
}

// Generator orchestrates the generation of documentation across multiple files
// and programming languages concurrently. It manages work distribution,
// integrates with services for documentation generation, supports minification
// through language-specific implementations, and allows for customization
// through options such as custom loggers, footers, concurrency limits, and
// language handlers. Generator emits generated documentation along with any
// errors encountered during the process.
type Generator struct {
	svc           Service
	languages     map[string]Language
	limit         int
	fileWorkers   int
	symbolWorkers int
	footer        string
	log           *slog.Logger
}

// Option configures a Generator by setting various parameters such as the
// logger, languages, and processing limits. Each Option is a function that
// applies a specific configuration to the Generator instance. These options
// enable customization of the documentation generation process to suit
// different needs and preferences.
type Option func(*Generator)

// WithLogger configures a new logger for the generator using the provided
// logging handler. It returns an option that, when applied, initializes the
// internal logger of the generator with the specified handler.
func WithLogger(h slog.Handler) Option {
	return func(g *Generator) {
		g.log = slog.New(h)
	}
}

// Footer sets a custom footer text that is appended to the generated
// documentation by a Generator instance. The text is provided as an argument
// and can be used to include additional information or a signature at the end
// of documentation output.
func Footer(msg string) Option {
	return func(g *Generator) {
		g.footer = msg
	}
}

// Limit applies a cap on the number of concurrent file processing workers in a
// Generator. It accepts an integer that specifies the maximum number of files
// to be processed at the same time. If the provided limit is less than one, it
// will not impose any restriction. This option allows for control over resource
// utilization during document generation tasks.
func Limit(n int) Option {
	return func(g *Generator) {
		g.limit = n
	}
}

// Workers configures the number of workers for processing files and symbols
// within a Generator. It accepts two integers representing the desired number
// of file workers and symbol workers, respectively. If either argument is less
// than 1, it will be set to 1 within the Generator. This function returns an
// Option which can be passed to New when creating a new Generator instance.
func Workers(files, symbols int) Option {
	return func(g *Generator) {
		g.fileWorkers = files
		g.symbolWorkers = symbols
	}
}

// WithLanguage associates a language implementation with a given file extension
// within the Generator. It accepts an extension string and a Language interface
// implementation, registering them so that the Generator can use the
// appropriate language behavior when processing files with that extension. This
// function is intended to be used in conjunction with other configuration
// options when initializing a new Generator.
func WithLanguage(ext string, lang Language) Option {
	return func(g *Generator) {
		g.languages[ext] = lang
	}
}

// New creates a new Generator using the provided Service and applies any
// additional options supplied. It initializes a Generator with default file and
// symbol workers based on CPU availability, a no-operation logger, and an empty
// language map. Options can be used to customize the Generator's behavior, such
// as setting custom loggers, footers, work limits, and language-specific
// functionality. Returns a pointer to the initialized Generator.
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

// Files initiates the generation of documentation for multiple files
// concurrently, using the provided context and a map of file paths to
// corresponding inputs. It returns two channels: one for receiving generated
// documentation encapsulated in [File] structs, and another for errors that may
// occur during the generation process. The operation can be cancelled through
// the context, and an error is returned if the initialization fails.
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
			go func(file string) {
				defer wg.Done()
				for input := range queue {
					g.log.Info(fmt.Sprintf("Generating %s ...", input))

					doc, err := g.Generate(ctx, PromptInput{
						Input: input,
						File:  file,
					})
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
			}(file)
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
				for {
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
								g.log.Debug(fmt.Sprintf("Reached file limit of %d files. Stopping file worker.", g.limit))
								return
							}
							nFiles.Add(1)
						}

						if !work(job.file, job.inputs) {
							g.log.Debug("Stopping file worker.")
							return
						}
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

// Generate orchestrates the creation of documentation for a given input within
// the context. It resolves the appropriate language handler, optionally
// minifies the code if supported, and invokes the associated service to produce
// documentation. The result is post-processed with any configured footer before
// being returned. If an unknown language is specified or a service error
// occurs, Generate will return an error detailing the failure.
func (g *Generator) Generate(ctx context.Context, input PromptInput) (string, error) {
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

	doc = strings.Trim(doc, `"' `)

	if g.footer != "" {
		doc = fmt.Sprintf("%s\n\n%s", doc, g.footer)
	}

	return doc, nil
}
