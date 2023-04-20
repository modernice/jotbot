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

// Repository represents a Jotbot repository. It provides methods for generating
// and patching files in the repository. Use the New function to create a new
// Repository instance. The Generate method generates files in the repository
// using a generate.Service, while the Patch method patches files in the
// repository using a patch.Patch. The Files method returns a channel of
// generated files and any errors encountered during generation.
type Repository struct {
	svc generate.Service
	log *slog.Logger
}

// Option is a type that represents an optional configuration for a Repository.
// It is a functional option that can modify the behavior of the Repository
// constructor. Option values can be passed to the constructor as variadic
// arguments and applied to the Repository using the Option function. The
// WithLogger Option sets the logger of the Repository.
type Option func(*Repository)

// WithLogger returns an Option that sets the logger for a Repository. The
// logger is used to log messages during generation and patching. The logger
// must implement slog.Handler.
func WithLogger(h slog.Handler) Option {
	return func(g *Repository) {
		g.log = slog.New(h)
	}
}

// New returns a new *Repository that uses the provided generate.Service.
// Options can be passed to configure the Repository. Use WithLogger to provide
// a slog.Handler for logging.
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

// GenerateOption is a type that represents an option for generating files. It
// is used as an argument to the Generate method of Repository to customize file
// generation. Two functions are provided to create GenerateOptions:
// GenerateWith and PatchWith. The former takes a variadic parameter of
// generate.Options, while the latter takes a variadic parameter of
// patch.Options.
type GenerateOption func(*generation)

// GenerateWith is an option for generating code with given options. It returns
// a GenerateOption that appends the given options to the generation.genOpts
// slice. The slice is later used in generating code.
func GenerateWith(opts ...generate.Option) GenerateOption {
	return func(g *generation) {
		g.genOpts = append(g.genOpts, opts...)
	}
}

// PatchWith returns a GenerateOption that appends patch options to the
// generation. These options are applied when generating a patch for a
// repository.
func PatchWith(opts ...patch.Option) GenerateOption {
	return func(g *generation) {
		g.patchOpts = append(g.patchOpts, opts...)
	}
}

type generation struct {
	genOpts   []generate.Option
	patchOpts []patch.Option
}

// Generate generates a patch for a repository at path repo. It accepts a
// context.Context and a GenerateOption variadic parameter that is used to
// configure the generation process. It returns the resulting *patch.Patch and
// an error (if any).
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

// Files returns a channel of
// [generate.File](https://pkg.go.dev/github.com/modernice/jotbot/generate#File)
// and a channel of errors, and an error. It generates files for the repository
// located at the given `repo` path, using the configured
// [generate.Service](https://pkg.go.dev/github.com/modernice/jotbot/generate#Service)
// and options. If an error occurs during generation, it is returned
// immediately. The returned channels are closed when there are no more files or
// errors to be sent.
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

// Patch applies comments to the source code files in a file system. It takes a
// context.Context, an fs.FS, a <-chan generate.File and zero or more
// patch.Option arguments. It returns a *patch.Patch and an error.
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
