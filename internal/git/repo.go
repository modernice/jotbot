package git

import (
	"fmt"
	"os/exec"
)

// Git represents a local Git repository path and provides a method to execute
// Git commands within that repository. The Cmd method constructs and runs a Git
// command with the given arguments, returning the underlying command execution
// details, combined standard output and error output, and any execution error
// that occurred.
type Git string

// Cmd executes a git command with the specified arguments within the
// repository's directory, returning the underlying [*exec.Cmd], combined
// standard output and standard error as a []byte, and an error if one occurred
// during command execution.
func (g Git) Cmd(args ...string) (*exec.Cmd, []byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = string(g)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return cmd, out, fmt.Errorf("git: %s (%w)", out, err)
	}
	return cmd, out, nil
}
