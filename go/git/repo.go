package git

import (
	"fmt"

	"github.com/modernice/opendocs/go/internal/git"
)

type Patch interface {
	Apply(root string) error
}

type Repository struct {
	root string
	git  git.Git
}

func Repo(root string) *Repository {
	return &Repository{
		root: root,
		git:  git.Git(root),
	}
}

func (repo *Repository) Root() string {
	return repo.root
}

func (r *Repository) Commit(p Patch) error {
	if _, _, err := r.git.Cmd("checkout", "-b", "opendocs-patch"); err != nil {
		return fmt.Errorf("checkout branch: %w", err)
	}

	if err := p.Apply(r.root); err != nil {
		return fmt.Errorf("apply patch to repository %q: %w", r.root, err)
	}

	if _, _, err := r.git.Cmd("add", "."); err != nil {
		return fmt.Errorf("add changes: %w", err)
	}

	if _, _, err := r.git.Cmd("commit", "-m", "docs: add missing documentation"); err != nil {
		return fmt.Errorf("commit patch: %w", err)
	}

	return nil
}
