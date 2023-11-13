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

// Patch represents an operation that can be applied to a repository to modify
// its contents. It encapsulates the changes that should be made within the
// context of a given directory, and it is responsible for executing those
// changes when prompted. The Apply method is used to trigger the application of
// the patch to the repository, effectively altering the state of the
// repository's files and directories as intended by the patch.
type Patch interface {
	// Apply applies the patch to the repository at the given root path, using the
	// provided context for any necessary operations. It returns an error if the
	// patch cannot be applied successfully.
	Apply(ctx context.Context, root string) error
}

// Committer represents an entity capable of producing a commit, which
// encapsulates changes to be recorded in a version control system. It provides
// a way to generate a [Commit] that describes the modifications made.
type Committer interface {
	// Commit creates a new commit in the repository to which the committer belongs,
	// incorporating changes from a provided patch. It applies the patch, stages the
	// changes, and then generates a commit with a message derived from the patch's
	// content. The function allows for specifying additional commit options, such
	// as the target branch. If any step of the process fails, an error detailing
	// the issue is returned.
	Commit() Commit
}

// Repository represents a version-controlled workspace where changes to files
// are tracked. It provides an interface to commit changesets, represented by
// the [Patch] interface, to the underlying version control system. It supports
// custom logging and branch naming through various options that can be passed
// during the creation or committing process. Additionally, it allows clients to
// retrieve the root directory of the repository.
type Repository struct {
	root string
	git  git.Git
	log  *slog.Logger
}

// Option configures a [*Repository] by setting its properties or initializing
// resources needed by the repository. It is typically passed to the Repo
// function to customize the returned [*Repository] instance.
type Option func(*Repository)

// WithLogger returns an Option that sets the logger of a Repository to a new
// logger using the provided slog.Handler.
func WithLogger(h slog.Handler) Option {
	return func(repo *Repository) {
		repo.log = slog.New(h)
	}
}

// Repo initializes a new instance of a [*Repository] with the provided root
// directory and applies any provided options. If no logger is provided in the
// options, a no-op logger is used by default. It returns the newly created
// [*Repository].
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

// Root retrieves the root directory path associated with the repository. It
// returns a string representing the filesystem path to the repository's root
// directory.
func (repo *Repository) Root() string {
	return repo.root
}

// CommitOption represents a configuration modifier that customizes the behavior
// of a commit operation within a repository. It allows for setting various
// commit-related properties or parameters before finalizing the commit. This
// modifier is generally used when performing a commit to specify options such
// as the target branch, author information, or commit message, ensuring that
// the commit reflects the desired state and metadata.
type CommitOption func(*commit)

// Branch configures the branch on which a commit should be made.
func Branch(branch string) CommitOption {
	return func(c *commit) {
		c.branch = branch
	}
}

type commit struct {
	branch string
}

// Commit applies the provided [Patch] to the repository, creating a new commit
// on a branch specified by the CommitOptions. If no branch is specified, a
// default one is created. The function records changes in the repository and
// logs the commit process. In case of failure during any step of the commit
// process, an error is returned detailing the issue encountered.
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
