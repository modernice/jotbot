package git

import (
	"fmt"
	"strings"
	"time"

	"github.com/modernice/jotbot/internal"
	"github.com/modernice/jotbot/internal/git"
	"golang.org/x/exp/slog"
)

// Patch represents a patch that can be applied to a Git repository. It is an
// interface with a single method, Apply, which applies the patch to the
// repository.
type Patch interface {
	Apply(root string) error
}

// Committer represents an interface for a type that can return a Commit. The
// Commit method signature is not defined in this interface, but it is expected
// to be implemented by types that satisfy it.
type Committer interface {
	Commit() Commit
}

// Repository represents a Git repository. It provides methods for committing
// patches to the repository. Use Repo to create a new Repository.
type Repository struct {
	root string
	git  git.Git
	log  *slog.Logger
}

// Option is a function type that takes a pointer to a Repository and applies an
// option to it. The Repo function creates a new Repository with the given root
// directory and applies the given options. The WithLogger option sets the
// logger for the Repository. The Branch option sets the branch name for a
// Commit operation.
type Option func(*Repository)

// WithLogger returns an Option that sets the logger of a Repository to the
// provided slog.Handler.
func WithLogger(h slog.Handler) Option {
	return func(repo *Repository) {
		repo.log = slog.New(h)
	}
}

// Repo represents a Git repository. It provides methods to commit a patch and
// to get the root directory of the repository. It also accepts options, such as
// logging, that can be used to customize its behavior.
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

// Root returns the root directory of the git repository.
func (repo *Repository) Root() string {
	return repo.root
}

// CommitOption is a type that represents an option for the Commit method of
// Repository. It is a function that takes a pointer to a commit struct and
// modifies its fields. The Branch function returns a CommitOption that sets the
// branch field of the commit struct.
type CommitOption func(*commit)

// Branch returns a CommitOption that sets the branch to commit to. It takes a
// string argument representing the branch name.
func Branch(branch string) CommitOption {
	return func(c *commit) {
		c.branch = branch
	}
}

type commit struct {
	branch string
}

// Commit creates a new commit in the repository. It takes a Patch and optional
// CommitOptions. If the CommitOption Branch is not provided, it defaults to
// "jotbot-patch". If a branch with that name already exists, it appends the
// current Unix timestamp to the branch name. If Patch implements the Committer
// interface, its Commit method is used to generate the commit message.
// Otherwise, DefaultCommit is used.
func (r *Repository) Commit(p Patch, opts ...CommitOption) error {
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

	if err := p.Apply(r.root); err != nil {
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
