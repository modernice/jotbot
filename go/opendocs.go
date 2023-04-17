package opendocs

import (
	"context"
	"os"

	"github.com/modernice/opendocs/go/generate"
)

type Repository struct {
	root string
}

func Repo(root string) *Repository {
	return &Repository{root}
}

func (r *Repository) Root() string {
	return r.root
}

func (repo *Repository) Generate(ctx context.Context, svc generate.Service, opts ...generate.Option) (*generate.Result, error) {
	g := generate.New(svc, opts...)
	result, err := g.Generate(ctx, os.DirFS(repo.root), opts...)
	if err != nil {
		return result, err
	}
	return result, nil
}
