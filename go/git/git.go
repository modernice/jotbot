package git

import (
	"fmt"

	"github.com/modernice/opendocs/go/internal"
)

type Patch interface {
	Apply(root string) error
}

type Repository struct {
	root string
	git  internal.Git
}

func Repo(root string) *Repository {
	return &Repository{
		root: root,
		git:  internal.Git(root),
	}
}

func (r *Repository) Commit(identifier, path string, p Patch) error {
	if _, _, err := r.git.Cmd("checkout", "-b", "opendocs-patch"); err != nil {
		return fmt.Errorf("checkout branch: %w", err)
	}

	if err := p.Apply(r.root); err != nil {
		return fmt.Errorf("apply patch to %q: %w", path, err)
	}

	if _, _, err := r.git.Cmd("add", path); err != nil {
		return fmt.Errorf("git add %q: %w", path, err)
	}

	if _, _, err := r.git.Cmd("commit", "-m", "opendocs: add `"+identifier+"` comment"); err != nil {
		return fmt.Errorf("commit patch: %w", err)
	}

	return nil
}
