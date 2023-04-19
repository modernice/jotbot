package gittest

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/modernice/opendocs/git"
	igit "github.com/modernice/opendocs/internal/git"
)

// Git is a type that represents a Git repository. It has methods to execute Git
// commands and assert the current branch and commit.
type Git igit.Git

// Cmd returns a pointer to an exec.Cmd struct, the output of the command as a
// byte slice, and an error. It takes a variadic argument of strings that
// represent the command and its arguments.
func (g Git) Cmd(args ...string) (*exec.Cmd, []byte, error) {
	return igit.Git(g).Cmd(args...)
}

// AssertBranch asserts that the current branch of a Git repository is equal to
// the given branch name. If the current branch is not equal to the given branch
// name, it will fail the test with a message indicating the expected and actual
// branch names.
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

// AssertCommit asserts that the latest commit message in the current branch of
// a Git repository matches the commit message of the given [git.Commit]. If the
// commit messages do not match, AssertCommit will fail the test.
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
