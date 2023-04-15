package generate

import (
	"context"
	"fmt"
	"io/fs"
	"log"

	"github.com/modernice/opendocs/go/find"
	"github.com/modernice/opendocs/go/git"
	"github.com/modernice/opendocs/go/patch"
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

type Generation struct {
	Path       string
	Identifier string
	Doc        string
}

type Generator struct {
	svc Service
}

func NewResult(repo fs.FS, generations ...Generation) Result {
	return Result{
		repo:        repo,
		generations: generations,
	}
}

func New(svc Service) *Generator {
	return &Generator{
		svc: svc,
	}
}

type Option func(*generation)

func Limit(n int) Option {
	return func(g *generation) {
		g.limit = n
	}
}

type generation struct {
	limit int
}

func (g *Generator) Generate(ctx context.Context, repo fs.FS, opts ...Option) (Result, error) {
	var cfg generation
	for _, opt := range opts {
		opt(&cfg)
	}

	out := NewResult(repo)

	result, err := find.New(repo).Uncommented()
	if err != nil {
		return out, fmt.Errorf("find uncommented code: %w", err)
	}

	var (
		generateCtx *genCtx
		nGenerated  int
	)

	for _, findings := range result {
		for _, finding := range findings {
			log.Printf("Generating docs for %s@%s", finding.Path, finding.Identifier)

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

			log.Printf("Generated docs for %s@%s", finding.Path, finding.Identifier)

			if cfg.limit > 0 && nGenerated >= cfg.limit {
				log.Printf("Limit of %d generations reached. Stopping.", cfg.limit)
				return out, nil
			}
		}
	}

	return out, nil
}

func (r Result) Generations() []Generation {
	return r.generations
}

func (r Result) Patch() *patch.Patch {
	p := patch.New(r.repo)
	for _, gen := range r.generations {
		p.Comment(gen.Path, gen.Identifier, gen.Doc)
	}
	return p
}

func (r Result) Commit(root string, opts ...git.CommitOption) (*git.Repository, error) {
	repo := git.Repo(root)
	return repo, repo.Commit(r.Patch(), opts...)
}
