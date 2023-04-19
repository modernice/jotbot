package tests

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/modernice/opendocs/internal/git"
)

var (
	//go:embed testdata/fixtures/basic
	basicFS embed.FS
	//go:embed testdata/fixtures/only-go-files
	onlyGoFilesFS embed.FS
	//go:embed testdata/fixtures/calculator
	calculatorFS embed.FS
	//go:embed testdata/fixtures/minify
	minifyFS embed.FS
	//go:embed testdata/fixtures/glob
	globFS embed.FS

	fixtures = map[string]fs.FS{
		"basic":         Must(fs.Sub(basicFS, "testdata/fixtures/basic")),
		"only-go-files": Must(fs.Sub(onlyGoFilesFS, "testdata/fixtures/only-go-files")),
		"calculator":    Must(fs.Sub(calculatorFS, "testdata/fixtures/calculator")),
		"minify":        Must(fs.Sub(minifyFS, "testdata/fixtures/minify")),
		"glob":          Must(fs.Sub(globFS, "testdata/fixtures/glob")),
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

	g := git.Git(root)

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

		targetPath := filepath.Join(root, path)
		targetDir := filepath.Dir(targetPath)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("create directory %q: %w", targetDir, err)
		}

		copy, err := os.Create(targetPath)
		if err != nil {
			return fmt.Errorf("create %q: %w", targetPath, err)
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
