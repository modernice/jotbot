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
	DefaultFileWorkers   = int(math.Min(4, float64(runtime.NumCPU())))
	DefaultSymbolWorkers = int(math.Min(2, float64(runtime.NumCPU())))
)

type Service interface {
	GenerateDoc(Context) (string, error)
}

type Language interface {
	Prompt(Input) string
}

type Minifier interface {
	Minify([]byte) ([]byte, error)
}

type Input struct {
	Code       []byte
	Language   string
	Identifier string
}

func (f Input) String() string {
	return fmt.Sprintf("%s (%s)", f.Identifier, f.Language)
}

type Context interface {
	context.Context

	Input() Input
	Prompt() string
}

type File struct {
	Path string
	Docs []Documentation
}

type Documentation struct {
	Input

	Text string
}

type Generator struct {
	svc           Service
	languages     map[string]Language
	limit         int
	fileWorkers   int
	symbolWorkers int
	footer        string
	log           *slog.Logger
}

type Option func(*Generator)

func WithLogger(h slog.Handler) Option {
	return func(g *Generator) {
		g.log = slog.New(h)
	}
}

func Footer(msg string) Option {
	return func(g *Generator) {
		g.footer = msg
	}
}

func Limit(n int) Option {
	return func(g *Generator) {
		g.limit = n
	}
}

func Workers(files, symbols int) Option {
	return func(g *Generator) {
		g.fileWorkers = files
		g.symbolWorkers = symbols
	}
}

func WithLanguage(ext string, lang Language) Option {
	return func(g *Generator) {
		g.languages[ext] = lang
	}
}

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
					// log.Println(input)
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
			return false
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
