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

// Repository represents a repository of source code files that can be used to
// generate documentation. It provides methods for generating a patch file
// containing comments for the identified functions or types, and for retrieving
// the files and errors generated during the process. It can be configured with
// a logger using the WithLogger option.
type Repository struct {
	svc generate.Service
	log *slog.Logger
}

// Option is a type that represents a configuration option for a Repository. It
// is a function that takes a pointer to a Repository and modifies it. The New
// function takes an optional list of Option functions that are applied to the
// new Repository. The WithLogger function returns an Option that sets the
// logger for the Repository.
type Option func(*Repository)

// WithLogger returns an Option that sets the logger for a Repository. The
// logger is used to log messages during the generation and patching of files.
// The logger must implement the slog.Handler interface from the
// golang.org/x/exp/slog package.
func WithLogger(h slog.Handler) Option {
	return func(g *Repository) {
		g.log = slog.New(h)
	}
}

// New creates a new *Repository with the given generate.Service and options. If
// no options are provided, a default logger is used.
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

// Generate generates a patch for the Go source code repository located at repo.
// It returns the generated patch or an error.
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

// Files returns a channel of
// [generate.File](https://pkg.go.dev/github.com/modernice/opendocs/generate#File)
// and a channel of errors that occur during generation. It generates
// documentation files for the repository located at the given path, using the
// [generate.Service](https://pkg.go.dev/github.com/modernice/opendocs/generate#Service)
// specified in the Repository. Additional options can be passed to the
// generate.Service using the variadic `opts` parameter.
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

// Patch applies comments to a file system [fs.FS] using the comments generated
// by Generate. It returns a *patch.Patch that represents the changes made to
// the file system.
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
