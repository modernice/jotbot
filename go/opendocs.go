package opendocs

import (
	"os"

	"github.com/modernice/opendocs/go/git"
	"github.com/modernice/opendocs/go/patch"
)

type Repository struct {
	root string
}

func Repo(root string) *Repository {
	return &Repository{
		root: root,
	}
}

func (r *Repository) Root() string {
	return r.root
}

func (repo *Repository) Patch() *patch.Patcher {
	return patch.New(os.DirFS(repo.root))
}

func (repo *Repository) Generate() (*patch.Patcher, error) {
	patch := repo.Patch()
	return patch, git.Repo(repo.root).Commit(patch)
}
