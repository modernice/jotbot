package gittest

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/modernice/opendocs/git"
	igit "github.com/modernice/opendocs/internal/git"
)

type Git igit.Git

func (g Git) Cmd(args ...string) (*exec.Cmd, []byte, error) {
	return igit.Git(g).Cmd(args...)
}

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
