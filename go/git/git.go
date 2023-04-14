package git

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/modernice/opendocs/go/internal"
)

type Patch interface {
	Apply(io.Writer) error
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

	path = filepath.Join(r.root, path)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %q: %w", path, err)
	}
	defer f.Close()

	if err := p.Apply(f); err != nil {
		return fmt.Errorf("apply patch to %q: %w", path, err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("close %q: %w", path, err)
	}

	if _, _, err := r.git.Cmd("add", path); err != nil {
		return fmt.Errorf("git add %q: %w", path, err)
	}

	if _, _, err := r.git.Cmd("commit", "-m", "opendocs: add `"+identifier+"` comment"); err != nil {
		return fmt.Errorf("commit patch: %w", err)
	}

	return nil
}
