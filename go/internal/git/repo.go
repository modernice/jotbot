package git

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

type Git string

func (g Git) Cmd(args ...string) (*exec.Cmd, []byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = string(g)
	out, err := cmd.Output()
	if err != nil {
		return cmd, out, fmt.Errorf("git: %s (%w)", out, err)
	}
	return cmd, out, nil
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

func (g Git) AssertCommit(t *testing.T, msg, desc string) {
	t.Helper()

	cmd, out, err := g.Cmd("log", "-1", "--pretty=%B")
	if err != nil {
		t.Fatalf("run command: %s", cmd)
	}

	want := msg
	if desc != "" {
		want = fmt.Sprintf("%s\n\n%s", msg, desc)
	}

	if got := strings.TrimSpace(string(out)); got != want {
		t.Fatalf("expected commit message %q; got %q", want, got)
	}
}
