package git

import (
	"fmt"

	"github.com/modernice/opendocs/go/internal/git"
)

type Patch interface {
	Apply(root string) error
}

type IdentifierProvider interface {
	Identifiers() map[string][]string
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

type CommitOption func(*commit)

func Branch(branch string) CommitOption {
	return func(c *commit) {
		c.branch = branch
	}
}

type commit struct {
	branch string
}

func (r *Repository) Commit(p Patch, opts ...CommitOption) error {
	var cfg commit
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.branch == "" {
		cfg.branch = "opendocs-patch"
	}

	if _, _, err := r.git.Cmd("checkout", "-b", cfg.branch); err != nil {
		return fmt.Errorf("checkout branch: %w", err)
	}

	if err := p.Apply(r.root); err != nil {
		return fmt.Errorf("apply patch to repository %q: %w", r.root, err)
	}

	if _, _, err := r.git.Cmd("add", "."); err != nil {
		return fmt.Errorf("add changes: %w", err)
	}

	msgs := []string{"docs: add missing documentation"}

	if fp, ok := p.(IdentifierProvider); ok {
		msgs = append(msgs, CommitDescription(fp.Identifiers()))
	}

	args := []string{"commit"}
	for _, msg := range msgs {
		args = append(args, "-m", msg)
	}

	if _, _, err := r.git.Cmd(args...); err != nil {
		return fmt.Errorf("commit patch: %w", err)
	}

	return nil
}