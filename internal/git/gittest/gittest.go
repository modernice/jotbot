package gittest

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/modernice/jotbot/git"
	igit "github.com/modernice/jotbot/internal/git"
)

// Git provides utilities for verifying the state of a Git repository in test
// cases. It can check the current branch, ensure a branch name has a specific
// prefix, and assert that the latest commit message matches an expected value.
// These assertions are intended to be used within testing functions, where they
// provide helpful error messages upon failure to aid in diagnosing issues with
// repository state during test execution.
type Git igit.Git

// Cmd runs a git command with the provided arguments and returns the command
// along with its output and any encountered error. It wraps the execution of a
// git command in a way that can be easily used for testing purposes, capturing
// both the standard output and potential errors for assertions.
func (g Git) Cmd(args ...string) (*exec.Cmd, []byte, error) {
	return igit.Git(g).Cmd(args...)
}

// AssertBranch confirms that the current git branch matches the specified
// branch name using the provided [*testing.T]. If the current branch does not
// match, it calls [*testing.T]'s Fatal method to immediately fail the test with
// an appropriate error message.
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

// AssertBranchPrefix verifies that the current branch name starts with the
// specified prefix, failing the test if it does not. It logs a fatal error with
// the testing.T if the current branch's name does not match the expected
// prefix. This function is intended to be used in test cases to ensure that
// branch naming conventions are followed.
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

// AssertCommit verifies that the most recent commit in the repository has the
// expected message. It compares the actual commit message with the provided
// commit object and reports an error to the testing context if they do not
// match. This method is intended to be used in test cases to ensure that
// operations affecting the repository's commit history behave as expected.
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
