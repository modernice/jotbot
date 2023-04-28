package git

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/modernice/jotbot/internal"
	"github.com/modernice/jotbot/internal/git"
	"golang.org/x/exp/slog"
)

type Patch interface {
	Apply(ctx context.Context, root string) error
}

type Committer interface {
	Commit() Commit
}

type Repository struct {
	root string
	git  git.Git
	log  *slog.Logger
}

type Option func(*Repository)

func WithLogger(h slog.Handler) Option {
	return func(repo *Repository) {
		repo.log = slog.New(h)
	}
}

func Repo(root string, opts ...Option) *Repository {
	repo := &Repository{
		root: root,
		git:  git.Git(root),
	}
	for _, opt := range opts {
		opt(repo)
	}
	if repo.log == nil {
		repo.log = internal.NopLogger()
	}
	return repo
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

func (r *Repository) Commit(ctx context.Context, p Patch, opts ...CommitOption) error {
	var cfg commit
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.branch == "" {
		cfg.branch = "jotbot-patch"
	}

	_, output, err := r.git.Cmd("rev-parse", "--verify", cfg.branch)
	if err == nil || strings.TrimSpace(string(output)) == "" {
		cfg.branch = fmt.Sprintf("%s_%d", cfg.branch, time.Now().UnixMilli())
	}

	r.log.Info("[git] Committing patch ...", "branch", cfg.branch)

	if _, output, err := r.git.Cmd("checkout", "-b", cfg.branch); err != nil {
		return fmt.Errorf("checkout branch: %w: %s", err, string(output))
	}

	if err := p.Apply(ctx, r.root); err != nil {
		return fmt.Errorf("apply patch to repository %s: %w", r.root, err)
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
