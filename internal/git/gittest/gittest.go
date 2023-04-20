package gittest

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/modernice/jotbot/git"
	igit "github.com/modernice/jotbot/internal/git"
)

// Git represents a Git repository, and provides methods to assert the current
// branch and latest commit.
type Git igit.Git

// Cmd returns a pointer to an exec.Cmd, the output of that command as []byte,
// and an error. It executes the provided arguments as a git command. [exec.Cmd]
func (g Git) Cmd(args ...string) (*exec.Cmd, []byte, error) {
	return igit.Git(g).Cmd(args...)
}

// AssertBranch asserts that the current Git branch is the specified branch. It
// takes a testing.T and a string representing the expected branch name. If the
// current branch is not the expected branch, AssertBranch will fail the test.
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

// AssertBranchPrefix asserts that the current Git branch has a specific prefix.
// The method takes a testing instance and a string prefix as arguments.
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

// AssertCommit asserts that the latest commit message matches the commit
// message of the given git.Commit.
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
