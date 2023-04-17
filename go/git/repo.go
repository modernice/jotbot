package git

import (
	"fmt"

	"github.com/modernice/opendocs/go/internal/git"
)

type Patch interface {
	Apply(root string) error
}

type Committer interface {
	Commit() Commit
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

	c := DefaultCommit()
	if com, ok := p.(Committer); ok {
		c = com.Commit()
	}

	paras := c.Paragraphs()

	args := []string{"commit"}
	for _, p := range paras {
		args = append(args, "-m", p)
	}

	if _, _, err := r.git.Cmd(args...); err != nil {
		return fmt.Errorf("commit patch: %w", err)
	}

	return nil
}

// func quote(s string) string {
// 	return fmt.Sprintf("%q", s)
// }
