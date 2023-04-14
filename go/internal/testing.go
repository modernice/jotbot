package internal

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	//go:embed testdata/fixtures/repo
	fixtures embed.FS
)

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func InitRepo(root string) error {
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		if err := os.RemoveAll(root); err != nil {
			return fmt.Errorf("remove existing repository directory: %w", err)
		}
	}

	if err := os.MkdirAll(root, 0755); err != nil {
		return fmt.Errorf("create repository directory: %w", err)
	}

	g := Git(root)

	if err := fs.WalkDir(fixtures, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		copy, err := os.Create(filepath.Join(root, entry.Name()))
		if err != nil {
			return fmt.Errorf("create file %q: %w", entry.Name(), err)
		}
		defer copy.Close()

		f, err := fixtures.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(copy, f)

		return err
	}); err != nil {
		return err
	}

	if _, _, err := g.Cmd("init"); err != nil {
		return err
	}

	if _, _, err := g.Cmd("add", "."); err != nil {
		return err
	}

	if _, _, err := g.Cmd("commit", "-m", "test commit"); err != nil {
		return err
	}

	return nil
}

type Git string

func (g Git) Cmd(args ...string) (*exec.Cmd, []byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = string(g)
	out, err := cmd.Output()
	if err != nil {
		return cmd, out, fmt.Errorf("git: %w", err)
	}
	return cmd, out, nil
}
