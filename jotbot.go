package jotbot

import (
	"context"
	"fmt"
	"io/fs"
	"os"

	"github.com/modernice/jotbot/generate"
	"github.com/modernice/jotbot/internal"
	"github.com/modernice/jotbot/patch"
	"golang.org/x/exp/slog"
)

// Repository represents a repository of code that can be used to generate
// documentation. It contains a generate.Service and a logger. Use New to create
// a new instance of Repository. Call Generate to generate a patch for the
// repository. Call Files to get a channel of files that can be used to generate
// documentation. Call Patch to add comments to the source code files.
type Repository struct {
	svc generate.Service
	log *slog.Logger
}

// Option represents an optional parameter for a Repository. Options can be used
// to configure the behavior of a Repository during generation and patching. An
// Option is a function that receives a pointer to a Repository and modifies it.
// The New function creates a new Repository with the given generate.Service and
// applies the provided Options. The GenerateOption type is an alias for either
// generate.Option or patch.Option, and is used to pass options to the Generate
// function.
type Option func(*Repository)

// WithLogger returns an Option that sets the logger for a Repository. The
// logger is used to log errors and debug information during generation and
// patching.
func WithLogger(h slog.Handler) Option {
	return func(g *Repository) {
		g.log = slog.New(h)
	}
}

// New returns a new Repository that generates documentation for a given
// repository using the provided generate.Service. If WithLogger is passed as an
// Option, it sets the logger of the Repository.
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

// GenerateOption is a type that represents an option for generating
// documentation. It is a function that takes a *generation and modifies it
// using generate.Option values. This type has two implementations:
// GenerateWith, which sets generate.Options, and PatchWith, which sets
// patch.Options.
type GenerateOption func(*generation)

// GenerateWith returns a GenerateOption that appends the given generate.Options
// to the slice of generate.Options passed to Repository.Generate.
func GenerateWith(opts ...generate.Option) GenerateOption {
	return func(g *generation) {
		g.genOpts = append(g.genOpts, opts...)
	}
}

// PatchWith returns a GenerateOption that appends patch options to the
// generation.
func PatchWith(opts ...patch.Option) GenerateOption {
	return func(g *generation) {
		g.patchOpts = append(g.patchOpts, opts...)
	}
}

type generation struct {
	genOpts   []generate.Option
	patchOpts []patch.Option
}

// Generate generates a patch for a repository by generating documentation for
// its files and applying patches to them. It returns the generated patch or an
// error if the generation or patching fails.
func (g *Repository) Generate(ctx context.Context, repo string, opts ...GenerateOption) (*patch.Patch, error) {
	var cfg generation
	for _, opt := range opts {
		opt(&cfg)
	}

	files, ferrs, err := g.Files(ctx, repo, cfg.genOpts...)
	if err != nil {
		return nil, err
	}

	errs := make(chan error)

	pc := make(chan *patch.Patch)
	go func() {
		p, err := g.Patch(ctx, os.DirFS(repo), files, cfg.patchOpts...)
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

// Files retrieves the list of files in a given repository and returns a channel
// of generate.File and a channel of error. It takes a context.Context, a string
// representing the repository path, and optional generate.Options.
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

// Patch applies comments to a file in a file system using the provided options
// [patch.Option]. It returns a *patch.Patch with the applied comments.
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
