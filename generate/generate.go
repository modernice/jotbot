package generate

//go:generate go-mockgen -f github.com/modernice/jotbot/generate -i Service -i Minifier -o ./mockgenerate/generate.go

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/modernice/jotbot/internal"
	"golang.org/x/exp/slog"
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
	Identifier string
	Target     string // e.g. "function 'foo' of class 'Foo'"
}

func (f Input) String() string {
	if f.Target != "" {
		return f.Target
	}
	return f.Identifier
}

type Context interface {
	context.Context

	Identifier() string
	Target() string
	File() []byte
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
	svc       Service
	languages map[string]Language
	limit     int
	footer    string
	log       *slog.Logger
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

func WithLanguage(ext string, lang Language) Option {
	return func(g *Generator) {
		g.languages[ext] = lang
	}
}

func New(svc Service, opts ...Option) *Generator {
	g := &Generator{svc: svc}
	for _, opt := range opts {
		opt(g)
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
		docs := make([]Documentation, 0, len(inputs))
		for _, input := range inputs {
			doc, err := g.Generate(ctx, input)
			if err != nil {
				fail(fmt.Errorf("generate %s: %w", input, err))
				continue
			}
			docs = append(docs, Documentation{Input: input, Text: doc})
		}
		if len(docs) == 0 {
			return true
		}
		return push(File{Path: file, Docs: docs})
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
		workers := runtime.NumCPU()
		if workers > len(files) {
			g.log.Debug(fmt.Sprintf("Setting workers to file count: %d", len(files)))
			workers = len(files)
		}

		g.log.Debug(fmt.Sprintf("Generating using %d workers.", workers))

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
	doc, err := g.svc.GenerateDoc(newCtx(ctx, input))
	if err != nil {
		return "", err
	}

	if g.footer != "" {
		doc = fmt.Sprintf("%s\n\n%s", doc, g.footer)
	}

	return doc, nil
}
