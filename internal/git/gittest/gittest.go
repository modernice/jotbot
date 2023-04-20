package gittest

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/modernice/jotbot/git"
	igit "github.com/modernice/jotbot/internal/git"
)

// Git is a type that represents an instance of the Git version control system.
// It provides methods for executing Git commands and asserting the state of the
// repository.
type Git igit.Git

// Cmd returns a *exec.Cmd that can be used to execute Git commands and captures
// the command's output. It takes a variadic list of strings, which are passed
// as arguments to the Git command. This method is a part of type Git [igit.Git]
// in the gittest package [gittest].
func (g Git) Cmd(args ...string) (*exec.Cmd, []byte, error) {
	return igit.Git(g).Cmd(args...)
}

// AssertBranch asserts that the current branch of a Git repository matches the
// given branch name [testing.T].
func (g Git) AssertBranch(t *testing.T, branch string) {
	t.Helper()

	_, output, err := g.Cmd("branch", "--show-current")
	if err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(string(output))

	if got != branch {
		t.Fatalf("expected to be in branch %q; branch is %q", branch, got)
	}
}

// AssertBranchPrefix asserts that the current Git branch has the specified
// prefix. It takes a testing.T instance and a string as arguments. If the
// prefix is not found in the current branch name, AssertBranchPrefix will fail
// the test with a descriptive error message.
func (g Git) AssertBranchPrefix(t *testing.T, prefix string) {
	t.Helper()

	_, output, err := g.Cmd("branch", "--show-current")
	if err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(string(output))

	if !strings.HasPrefix(got, prefix) {
		t.Fatalf("expected branch %q to have prefix %q", got, prefix)
	}
}

// AssertCommit asserts that the last commit message matches the provided
// `git.Commit`. It takes a `*testing.T` and a `git.Commit` as arguments. If the
// assertion fails, it will fail the test.
func (g Git) AssertCommit(t *testing.T, c git.Commit) {
	t.Helper()

	cmd, out, err := g.Cmd("log", "-1", "--pretty=%B")
	if err != nil {
		t.Fatalf("run command: %s", cmd)
	}

	want := c.String()

	if got := strings.TrimSpace(string(out)); got != want {
		t.Fatalf("unexpected commit message\n%s\n\nwant:\n%s\n\ngot:\n%s", cmp.Diff(want, got), want, got)
	}
}
