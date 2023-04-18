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

type File struct {
	Path        string
	Generations []Generation
}

type Generation struct {
	File       string
	Identifier string
	Doc        string
}

type Generator struct {
	svc Service
}

func New(svc Service) *Generator {
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

func (g *Generator) Generate(ctx context.Context, repo fs.FS, opts ...Option) (<-chan File, <-chan error, error) {
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

	files, errs := make(chan File), make(chan error)

	ctx, cancel := context.WithCancel(ctx)

	var (
		nFiles     int64
		nGenerated int64
	)

	canGenerate := func() bool {
		nf := atomic.LoadInt64(&nFiles)
		if cfg.fileLimit > 0 && nf >= int64(cfg.fileLimit) {
			return false
		}

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
