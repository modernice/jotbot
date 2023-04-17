package opendocs

import (
	"context"
	"os"

	"github.com/modernice/opendocs/generate"
)

type Repo string

func (repo Repo) Generate(ctx context.Context, svc generate.Service, opts ...generate.Option) (*generate.Result, error) {
	g := generate.New(svc, opts...)
	result, err := g.Generate(ctx, os.DirFS(string(repo)), opts...)
	if err != nil {
		return result, err
	}
	return result, nil
}
