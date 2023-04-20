package git

import (
	"fmt"
	"os/exec"
)

// Package git provides a Git client for executing Git commands.
//
// Git type represents a local Git repository path. The Cmd method on Git
// executes the given Git command with the specified arguments and returns the
// output as bytes, along with any error occurred during execution.
type Git string

// Cmd executes a git command with the given arguments on the Git repository. It
// returns the *exec.Cmd object representing the command, the combined output of
// stdout and stderr as a []byte, and an error if the command failed. [exec.Cmd]
func (g Git) Cmd(args ...string) (*exec.Cmd, []byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = string(g)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return cmd, out, fmt.Errorf("git: %s (%w)", out, err)
	}
	return cmd, out, nil
}
