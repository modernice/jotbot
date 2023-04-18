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

type Service interface {
	GenerateDoc(Context) (string, error)
}

type Context interface {
	context.Context

	Identifier() string
	File() string
	Files() []string
	Read(file string) ([]byte, error)
}

type Generation struct {
	Path       string
	Identifier string
	Doc        string
}

type Generator struct {
	svc Service
}

func New(svc Service, opts ...Option) *Generator {
	return &Generator{svc: svc}
}

type Option func(*generation)

func Limit(n int) Option {
	return func(g *generation) {
		g.limit = n
	}
}

func FileLimit(n int) Option {
	return func(g *generation) {
		g.fileLimit = n
	}
}

func WithLogger(h slog.Handler) Option {
	return func(g *generation) {
		g.log = slog.New(h)
	}
}

func Footer(msg string) Option {
	return func(g *generation) {
		g.footer = msg
	}
}

type generation struct {
	limit     int
	fileLimit int
	footer    string
	log       *slog.Logger
}

func (g *Generator) Generate(ctx context.Context, repo fs.FS, opts ...Option) (<-chan Generation, <-chan error, error) {
	var cfg generation
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.log == nil {
		cfg.log = internal.NopLogger()
	}

	result, err := find.New(repo, find.WithLogger(cfg.log.Handler())).Uncommented()
	if err != nil {
		return nil, nil, fmt.Errorf("find uncommented code: %w", err)
	}

	gens, errs := make(chan Generation), make(chan error)

	ctx, cancel := context.WithCancel(ctx)

	var nGenerated int
	push := func(g Generation) bool {
		select {
		case <-ctx.Done():
			return false
		case gens <- g:
			nGenerated++
			if cfg.limit > 0 && nGenerated >= cfg.limit {
				cancel()
			}
			return true
		}
	}

	fail := func(err error) {
		select {
		case <-ctx.Done():
		case errs <- err:
		}
	}

	var nFiles int64
	fileDone := func() {
		if cfg.limit <= 0 {
			return
		}

		n := atomic.AddInt64(&nFiles, 1)
		if n >= int64(cfg.fileLimit) {
			cancel()
		}
	}

	workers := runtime.NumCPU()
	if cfg.fileLimit > 0 && cfg.fileLimit < workers {
		cfg.log.Debug(fmt.Sprintf("File limit is lower than number of workers. Reducing workers to %d.", cfg.fileLimit))
		workers = cfg.fileLimit
	}

	queue := make(chan string)
	var wg sync.WaitGroup
	wg.Add(workers)

	go func() {
		defer cancel()
		wg.Wait()
		close(gens)
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

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for file := range queue {
				findings := result[file]
				for _, finding := range findings {
					var generateCtx *genCtx
					var err error
					if generateCtx, err = g.context(generateCtx, ctx, repo, finding); err != nil {
						fail(fmt.Errorf("generate %s: create context: %w", finding, err))
						return
					}

					gen, err := g.generate(generateCtx, cfg, finding)
					if err != nil {
						fail(err)
						return
					}

					if !push(gen) {
						return
					}
				}

				fileDone()
			}
		}()
	}

	return gens, errs, nil
}

func (g *Generator) context(ctx *genCtx, parent context.Context, repo fs.FS, finding find.Finding) (*genCtx, error) {
	if ctx == nil {
		return newCtx(parent, repo, finding.Path, finding.Identifier)
	}
	return ctx.new(parent, finding.Path, finding.Identifier), nil
}

func (g *Generator) generate(ctx *genCtx, cfg generation, finding find.Finding) (Generation, error) {
	cfg.log.Info("Generating docs ...", "path", finding.Path, "identifier", finding.Identifier)

	doc, err := g.svc.GenerateDoc(ctx)
	if err != nil {
		return Generation{}, fmt.Errorf("generate %s: generate doc: %w", finding, err)
	}

	if cfg.footer != "" {
		doc = fmt.Sprintf("%s\n\n%s", doc, cfg.footer)
	}

	return Generation{
		Path:       finding.Path,
		Identifier: finding.Identifier,
		Doc:        doc,
	}, nil
}
