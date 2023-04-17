package git

import (
	"fmt"
	"os/exec"
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
