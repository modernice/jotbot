package generate

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/modernice/opendocs/go/find"
	"github.com/modernice/opendocs/go/git"
	"github.com/modernice/opendocs/go/internal"
	"github.com/modernice/opendocs/go/patch"
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

type Result struct {
	repo        fs.FS
	generations []Generation
}

func NewResult(repo fs.FS, generations ...Generation) Result {
	return Result{
		repo:        repo,
		generations: generations,
	}
}

type Generation struct {
	Path       string
	Identifier string
	Doc        string
}

type Generator struct {
	svc   Service
	limit int
	log   *slog.Logger
}

type Option func(*Generator)

func Limit(n int) Option {
	return func(g *Generator) {
		g.limit = n
	}
}

func WithLogger(h slog.Handler) Option {
	return func(g *Generator) {
		g.log = slog.New(h)
	}
}

func New(svc Service, opts ...Option) *Generator {
	gen := &Generator{svc: svc}
	for _, opt := range opts {
		opt(gen)
	}
	if gen.log == nil {
		gen.log = internal.NopLogger()
	}
	return gen
}

func (g *Generator) Generate(ctx context.Context, repo fs.FS, opts ...Option) (Result, error) {
	out := NewResult(repo)

	result, err := find.New(repo, find.WithLogger(g.log.Handler())).Uncommented()
	if err != nil {
		return out, fmt.Errorf("find uncommented code: %w", err)
	}

	var (
		generateCtx *genCtx
		nGenerated  int
	)

	for _, findings := range result {
		for _, finding := range findings {
			g.log.Info("Generating docs ...", "path", finding.Path, "identifier", finding.Identifier)

			if generateCtx == nil {
				if generateCtx, err = newCtx(ctx, repo, finding.Path, finding.Identifier); err != nil {
					return out, fmt.Errorf("create generation context: %w", err)
				}
			} else {
				generateCtx = generateCtx.new(ctx, finding.Path, finding.Identifier)
			}

			doc, err := g.svc.GenerateDoc(generateCtx)
			if err != nil {
				return out, fmt.Errorf("generate doc for %q in %q: %w", finding.Identifier, finding.Path, err)
			}

			out.generations = append(out.generations, Generation{
				Path:       finding.Path,
				Identifier: finding.Identifier,
				Doc:        doc,
			})

			nGenerated++

			if g.limit > 0 && nGenerated >= g.limit {
				g.log.Debug(fmt.Sprintf("Limit of %d generations reached. Stopping.", g.limit))
				return out, nil
			}
		}
	}

	return out, nil
}

func (r Result) Generations() []Generation {
	return r.generations
}

func (r Result) Patch(opts ...patch.Option) *patch.Patch {
	p := patch.New(r.repo, opts...)
	for _, gen := range r.generations {
		p.Comment(gen.Path, gen.Identifier, gen.Doc)
	}
	return p
}

func (r Result) Commit(root string, opts ...git.CommitOption) (*git.Repository, error) {
	repo := git.Repo(root)
	return repo, repo.Commit(r.Patch(), opts...)
}
