package generate

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
	svc Service
}

func New(svc Service) *Generator {
	return &Generator{svc: svc}
}

type Option func(*config)

func Limit(n int) Option {
	return func(g *config) {
		g.limit = n
	}
}

func WithLogger(h slog.Handler) Option {
	return func(g *config) {
		g.log = slog.New(h)
	}
}

func Footer(msg string) Option {
	return func(g *config) {
		g.footer = msg
	}
}

type config struct {
	limit  int
	footer string
	log    *slog.Logger
}

func configure(opts ...Option) (cfg config) {
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.log == nil {
		cfg.log = internal.NopLogger()
	}
	return
}

func (g *Generator) Files(ctx context.Context, files map[string][]Input, opts ...Option) (<-chan File, <-chan error, error) {
	cfg := configure(opts...)

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

	work, done := g.distributeWork(files, cfg)
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

func (g *Generator) distributeWork(files map[string][]Input, cfg config) (func(context.Context, func(string, []Input) bool), <-chan struct{}) {
	done := make(chan struct{})
	return func(ctx context.Context, work func(string, []Input) bool) {
		workers := runtime.NumCPU()
		if workers > len(files) {
			cfg.log.Debug(fmt.Sprintf("Setting workers to file count: %d", len(files)))
			workers = len(files)
		}

		cfg.log.Debug(fmt.Sprintf("Generating using %d workers.", workers))

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
					if cfg.limit > 0 {
						n := nFiles.Load()
						if n >= int64(cfg.limit) {
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

func (g *Generator) Generate(ctx context.Context, input Input, opts ...Option) (string, error) {
	cfg := configure(opts...)

	doc, err := g.svc.GenerateDoc(newCtx(ctx, input))
	if err != nil {
		return "", err
	}

	if cfg.footer != "" {
		doc = fmt.Sprintf("%s\n\n%s", doc, cfg.footer)
	}

	return doc, nil
}
