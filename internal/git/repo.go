package git

import (
	"fmt"
	"os/exec"
)

// Git is a type that represents a local Git repository. It has a method, Cmd,
// that executes a Git command with the given arguments and returns the
// resulting output as a byte slice.
type Git string

// Cmd executes a Git command with the given arguments in the context of a Git
// repository. It returns a pointer to an exec.Cmd struct, the output of the
// command as a byte slice, and an error if any occurred. [exec.Cmd]
func (g Git) Cmd(args ...string) (*exec.Cmd, []byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = string(g)
	out, err := cmd.Output()
	if err != nil {
		return cmd, out, fmt.Errorf("git: %s (%w)", out, err)
	}
	return cmd, out, nil
}
