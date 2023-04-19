package git

import (
	"fmt"
	"strings"
	"time"

	"github.com/modernice/opendocs/internal"
	"github.com/modernice/opendocs/internal/git"
	"golang.org/x/exp/slog"
)

// Patch is an interface that defines the Apply method. Apply takes a string
// argument representing the root directory of a repository and applies the
// patch to the repository. The Commit method of Repository takes a Patch as its
// first argument and applies the patch to the repository. If the Patch also
// implements the Committer interface, the Commit method will use the Commit
// method of the Patch to generate a commit message.
type Patch interface {
	Apply(root string) error
}

// Committer is an interface that defines the Commit method. Any type that
// implements Committer can be used as a parameter to the Commit method of
// Repository. The Commit method creates a new branch, applies a patch to the
// repository, stages the changes, and commits them with a commit message
// generated from the Commit method of the provided patch. If the provided patch
// does not implement Committer, a default commit message is used.
type Committer interface {
	Commit() Commit
}

// Repository represents a Git repository. It provides methods for committing
// patches to the repository. A Repository can be created using the Repo
// function, which takes a root directory and optional options. The Root method
// returns the root directory of the repository. The Commit method takes a Patch
// and optional CommitOptions and applies the patch to the repository,
// committing the changes with a commit message generated from the Patch's
// Commit method or from the CommitOptions.
type Repository struct {
	root string
	git  git.Git
	log  *slog.Logger
}

// Option is a type that represents a configuration option for a Repository. It
// is a function that takes a pointer to a Repository and modifies it. The Repo
// function creates a new Repository with the given root and applies the given
// options to it. WithLogger is an example of an Option that sets the logger for
// the Repository.
type Option func(*Repository)

// WithLogger returns an Option that sets the logger of a *Repository to the
// provided slog.Handler. The logger is used to log information about the
// repository's actions.
func WithLogger(h slog.Handler) Option {
	return func(repo *Repository) {
		repo.log = slog.New(h)
	}
}

// Repo represents a Git repository. It contains a root directory, a Git
// instance, and a logger. It provides methods for committing patches to the
// repository. The Root method returns the root directory of the repository. The
// Commit method applies a patch to the repository, adds the changes, and
// commits them with a commit message generated from the patch's Commit method
// or from a default commit message. The Commit method also allows for optional
// configuration via CommitOption functions. The Repo function creates a new
// Repository instance with optional configuration via Option functions.
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

// Root returns the root directory of the repository.
func (repo *Repository) Root() string {
	return repo.root
}

// CommitOption is a type that represents an option for the Commit method of
// Repository. It is a function that takes a pointer to a commit and modifies
// it. The Branch function returns a CommitOption that sets the branch of the
// commit.
type CommitOption func(*commit)

// Branch is a function that returns a CommitOption. The CommitOption sets the
// branch name for a commit. If no branch name is provided, the default branch
// name is "opendocs-patch".
func Branch(branch string) CommitOption {
	return func(c *commit) {
		c.branch = branch
	}
}

type commit struct {
	branch string
}

// Commit applies a patch to the repository and creates a new commit. It takes a
// Patch as its first argument and optional CommitOptions as subsequent
// arguments. If no branch is specified in the options, it defaults to
// "opendocs-patch". If the branch already exists, it appends the current Unix
// time in milliseconds to the branch name. The commit message is generated from
// the Paragraphs of the Commit returned by the Committer, if the Patch
// implements Committer.
func (r *Repository) Commit(p Patch, opts ...CommitOption) error {
	var cfg commit
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.branch == "" {
		cfg.branch = "opendocs-patch"
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
