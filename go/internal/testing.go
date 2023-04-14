package internal

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	opendocs "github.com/modernice/opendocs/go"
)

var (
	//go:embed testdata/fixtures/basic
	basicFS embed.FS
	//go:embed testdata/fixtures/only-go-files
	onlyGoFilesFS embed.FS

	fixtures = map[string]embed.FS{
		"basic":         basicFS,
		"only-go-files": onlyGoFilesFS,
	}
)

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func WithRepo(name string, root string, fn func(fs.FS)) {
	if err := InitRepo(name, root); err != nil {
		panic(err)
	}
	fn(os.DirFS(root))
}

func InitRepo(name, root string) error {
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		if err := os.RemoveAll(root); err != nil {
			return fmt.Errorf("remove existing repository directory: %w", err)
		}
	}

	if err := os.MkdirAll(root, 0755); err != nil {
		return fmt.Errorf("create repository directory: %w", err)
	}

	g := Git(root)

	fixture, ok := fixtures[name]
	if !ok {
		panic(fmt.Errorf("unknown fixture %q", name))
	}

	if err := fs.WalkDir(fixture, ".", func(path string, entry fs.DirEntry, err error) error {
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

		f, err := fixture.Open(path)
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

func AssertFindings(t *testing.T, want, got opendocs.Findings) {
	t.Helper()
	if !cmp.Equal(want, got) {
		t.Fatalf("unexpected findings:\n%s", cmp.Diff(want, got))
	}
}
