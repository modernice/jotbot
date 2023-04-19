package git

import (
	"fmt"
	"os/exec"
)

// Git represents a Git repository. It has a Cmd method that executes a Git
// command with the given arguments and returns the resulting output as a byte
// slice. Any errors encountered during execution are returned as an error.
type Git string

// Cmd returns a pointer to an exec.Cmd struct, the output of the command as a
// byte slice, and an error. It executes a Git command with the given arguments
// on the Git repository specified by the receiver Git.
func (g Git) Cmd(args ...string) (*exec.Cmd, []byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = string(g)
	out, err := cmd.Output()
	if err != nil {
		return cmd, out, fmt.Errorf("git: %s (%w)", out, err)
	}
	return cmd, out, nil
}
