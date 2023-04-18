package opendocs

import (
	"context"
	"fmt"
	"io/fs"
	"os"

	"github.com/modernice/opendocs/generate"
	"github.com/modernice/opendocs/internal"
	"github.com/modernice/opendocs/patch"
	"golang.org/x/exp/slog"
)

type Repository struct {
	svc generate.Service
	log *slog.Logger
}

type Option func(*Repository)

func WithLogger(h slog.Handler) Option {
	return func(g *Repository) {
		g.log = slog.New(h)
	}
}

func New(svc generate.Service, opts ...Option) *Repository {
	g := &Repository{svc: svc}
	for _, opt := range opts {
		opt(g)
	}
	if g.log == nil {
		g.log = internal.NopLogger()
	}
	return g
}

func (g *Repository) Generate(ctx context.Context, repo string, opts ...generate.Option) (*patch.Patch, error) {
	files, ferrs, err := g.Files(ctx, repo, opts...)
	if err != nil {
		return nil, err
	}

	errs := make(chan error)

	pc := make(chan *patch.Patch)
	go func() {
		p, err := g.Patch(ctx, os.DirFS(repo), files)
		if err != nil {
			select {
			case <-ctx.Done():
			case errs <- err:
			}
			return
		}
		select {
		case <-ctx.Done():
		case pc <- p:
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case err, ok := <-errs:
			if !ok {
				errs = nil
				break
			}
			return nil, err
		case err, ok := <-ferrs:
			if !ok {
				ferrs = nil
				break
			}
			return nil, err
		case p := <-pc:
			return p, nil
		}
	}
}

func (g *Repository) Files(ctx context.Context, repo string, opts ...generate.Option) (<-chan generate.File, <-chan error, error) {
	opts = append([]generate.Option{generate.WithLogger(g.log.Handler())}, opts...)

	repoFS := os.DirFS(repo)
	gen := generate.New(g.svc)

	files, errs, err := gen.Generate(ctx, repoFS, opts...)
	if err != nil {
		return files, errs, fmt.Errorf("generate: %w", err)
	}

	return files, errs, nil
}

func (g *Repository) Patch(ctx context.Context, repo fs.FS, files <-chan generate.File, opts ...patch.Option) (*patch.Patch, error) {
	opts = append([]patch.Option{patch.WithLogger(g.log.Handler())}, opts...)
	p := patch.New(repo, opts...)

	for file := range files {
		for _, gen := range file.Generations {
			if err := p.Comment(gen.File, gen.Identifier, gen.Doc); err != nil {
				return nil, fmt.Errorf("add comment to %s@%s: %w", gen.File, gen.Identifier, err)
			}
		}
	}

	return p, nil
}
