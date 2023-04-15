package generate

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/modernice/opendocs/go/find"
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

func (g *Generator) Generate(ctx context.Context, repo fs.FS) (Result, error) {
	out := NewResult(repo)

	result, err := find.New(repo).Uncommented()
	if err != nil {
		return out, fmt.Errorf("find uncommented code: %w", err)
	}

	for _, findings := range result {
		for _, finding := range findings {
			ctx, err := newCtx(ctx, repo, finding.Path, finding.Identifier)
			if err != nil {
				return out, fmt.Errorf("create generation context: %w", err)
			}

			doc, err := g.svc.GenerateDoc(ctx)
			if err != nil {
				return out, fmt.Errorf("generate doc for %q in %q: %w", finding.Identifier, finding.Path, err)
			}

			out.generations = append(out.generations, Generation{
				Path:       finding.Path,
				Identifier: finding.Identifier,
				Doc:        doc,
			})
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
