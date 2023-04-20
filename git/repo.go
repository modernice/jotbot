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
// interface that defines the Apply method. The Apply method takes a root string
// as argument, which specifies the root directory of the Git repository, and
// returns an error if the patch application fails.
type Patch interface {
	Apply(root string) error
}

// Committer represents an interface for a type that can commit changes to a Git
// repository. Any type implementing this interface must have a Commit method
// which returns a Commit. The Repository type has a Commit method which takes a
// Patch and optional CommitOptions and applies the patch to the repository,
// creates a new branch if necessary, adds the changes to the index, and commits
// them with the specified commit message. If the provided patch implements the
// Committer interface, its Commit method will be used to create the commit
// message.
type Committer interface {
	Commit() Commit
}

// Repository represents a Git repository. It provides methods for committing
// patches and accessing the root directory of the repository. The Repository
// type implements the Patch and Committer interfaces, allowing for patches to
// be applied to the repository and commits to be made on behalf of the
// repository, respectively. It can be initialized with options using Repo
// function.
type Repository struct {
	root string
	git  git.Git
	log  *slog.Logger
}

// Option is a type that represents an option for a Repository. It is used to
// configure Repository instances with optional parameters. Option functions
// modify the Repository instance passed to them.
type Option func(*Repository)

// WithLogger returns an Option that sets the logger for a Repository. The
// logger is used to log information about the git commands that are run during
// patching.
func WithLogger(h slog.Handler) Option {
	return func(repo *Repository) {
		repo.log = slog.New(h)
	}
}

// Repo represents a Git repository. It provides methods for committing patches
// to the repository. Use the Repo function to create a new Repository.
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

// Root returns the root directory of the Repository [Repository].
func (repo *Repository) Root() string {
	return repo.root
}

// CommitOption is a type representing a function that modifies the behavior of
// the Commit function in Repository. It can be used as an optional argument in
// the Commit function to specify additional options such as the branch to
// commit to. The Branch function returns a CommitOption that sets the branch
// name for the commit.
type CommitOption func(*commit)

// Branch is a function that returns a CommitOption. The returned option sets
// the name of the branch to commit to when committing a patch. If not set, the
// branch name defaults to "jotbot-patch".
func Branch(branch string) CommitOption {
	return func(c *commit) {
		c.branch = branch
	}
}

type commit struct {
	branch string
}

// Commit commits a patch to the repository. It takes a Patch and optional
// CommitOptions. If no branch is specified in the options, "jotbot-patch" is
// used. If the branch already exists, the name will be suffixed with the
// current Unix timestamp in milliseconds. The patch is applied to the
// repository, and all changes are added and committed. If the Patch implements
// Committer, its Commit method will be called to obtain commit information.
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
