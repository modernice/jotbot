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
	Repo        fs.FS
	Generations []Generation
	Logger      slog.Handler
}

func NewResult(repo fs.FS, generations ...Generation) *Result {
	return &Result{
		Repo:        repo,
		Generations: generations,
	}
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
	limit  int
	footer string
	log    *slog.Logger
}

func (g *Generator) Generate(ctx context.Context, repo fs.FS, opts ...Option) (*Result, error) {
	var cfg generation
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.log == nil {
		cfg.log = internal.NopLogger()
	}

	out := NewResult(repo)
	out.Logger = cfg.log.Handler()

	result, err := find.New(repo, find.WithLogger(cfg.log.Handler())).Uncommented()
	if err != nil {
		return out, fmt.Errorf("find uncommented code: %w", err)
	}

	var (
		generateCtx *genCtx
		nGenerated  int
	)

	for _, findings := range result {
		for _, finding := range findings {
			cfg.log.Info("Generating docs ...", "path", finding.Path, "identifier", finding.Identifier)

			if generateCtx == nil {
				if generateCtx, err = newCtx(ctx, repo, finding.Path, finding.Identifier); err != nil {
					return out, fmt.Errorf("create generation context: %w", err)
				}
			} else {
				generateCtx = generateCtx.new(ctx, finding.Path, finding.Identifier)
			}

			doc, err := g.svc.GenerateDoc(generateCtx)
			if err != nil {
				return out, fmt.Errorf("generate doc for %s in %s: %w", finding.Identifier, finding.Path, err)
			}

			if cfg.footer != "" {
				doc = fmt.Sprintf("%s\n\n%s", doc, cfg.footer)
			}

			out.Generations = append(out.Generations, Generation{
				Path:       finding.Path,
				Identifier: finding.Identifier,
				Doc:        doc,
			})

			nGenerated++

			if cfg.limit > 0 && nGenerated >= cfg.limit {
				cfg.log.Debug(fmt.Sprintf("Limit of %d generations reached. Stopping.", cfg.limit))
				return out, nil
			}
		}
	}

	return out, nil
}

func (r *Result) Patch() *patch.Patch {
	opts := []patch.Option{patch.WithLogger(r.Logger)}
	p := patch.New(r.Repo, opts...)
	for _, gen := range r.Generations {
		p.Comment(gen.Path, gen.Identifier, gen.Doc)
	}
	return p
}

func (r *Result) Commit(root string, opts ...git.CommitOption) (*git.Repository, error) {
	repo := git.Repo(root, git.WithLogger(r.Logger))
	return repo, repo.Commit(r.Patch(), opts...)
}
