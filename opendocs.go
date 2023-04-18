package opendocs

import (
	"context"
	"io/fs"
	"os"

	"github.com/modernice/opendocs/generate"
	"github.com/modernice/opendocs/internal"
)

type Repo string

func (r Repo) FS() fs.FS {
	return os.DirFS(string(r))
}

func (repo Repo) Generate(ctx context.Context, svc generate.Service, opts ...generate.Option) ([]generate.Generation, error) {
	g := generate.New(svc, opts...)
	gens, errs, err := g.Generate(ctx, repo.FS(), opts...)
	if err != nil {
		return nil, err
	}
	return internal.Drain(gens, errs)
}
