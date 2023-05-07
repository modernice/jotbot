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

// Patch represents an interface for applying changes to a repository in the
// form of patches. It provides a method to apply the patch to a given root
// directory within a context.
type Patch interface {
	// Apply applies the given patch to the repository at the specified root
	// directory within the provided context. Returns an error if the patch
	// application fails.
	Apply(ctx context.Context, root string) error
}

// Committer is an interface that provides a method to generate a Commit object,
// which represents a git commit with a message and optional metadata. It is
// used when applying patches to repositories.
type Committer interface {
	Commit() Commit
}

// Repository represents a Git repository, providing methods for applying
// patches and committing changes. It uses the [git.Git] interface for executing
// Git commands and supports custom logging with an optional [slog.Logger].
type Repository struct {
	root string
	git  git.Git
	log  *slog.Logger
}

// Option is a function that configures a [Repository] by applying specific
// settings or customizations. It allows for optional and flexible configuration
// of a [Repository] instance without modifying its core implementation.
type Option func(*Repository)

// WithLogger returns an Option that sets the logger of a Repository to the
// provided slog.Handler.
func WithLogger(h slog.Handler) Option {
	return func(repo *Repository) {
		repo.log = slog.New(h)
	}
}

// Repo creates a new Repository instance with the given root directory and
// applies the provided options. It returns a pointer to the created Repository.
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

// Root returns the root directory of the Repository.
func (repo *Repository) Root() string {
	return repo.root
}

// CommitOption is a function that modifies the configuration of a commit
// operation in a Repository. It is used as an optional argument for the Commit
// method to customize the commit behavior, such as specifying a branch to
// commit the changes to.
type CommitOption func(*commit)

// Branch creates a CommitOption that sets the branch name for a commit
// operation in a Repository. If the specified branch already exists, a unique
// branch name is generated by appending the current Unix time in milliseconds.
func Branch(branch string) CommitOption {
	return func(c *commit) {
		c.branch = branch
	}
}

type commit struct {
	branch string
}

// Commit applies the provided Patch to the Repository, creates a new branch,
// and commits the changes with the specified CommitOptions. It returns an error
// if any step in the process fails.
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
