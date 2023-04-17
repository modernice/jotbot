package tests

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/modernice/opendocs/find"
	"github.com/modernice/opendocs/generate"
	"github.com/modernice/opendocs/internal/git"
	"github.com/modernice/opendocs/patch"
	"golang.org/x/exp/slices"
)

var (
	//go:embed testdata/fixtures/basic
	basicFS embed.FS
	//go:embed testdata/fixtures/only-go-files
	onlyGoFilesFS embed.FS
	//go:embed testdata/fixtures/calculator
	calculatorFS embed.FS

	fixtures = map[string]embed.FS{
		"basic":         basicFS,
		"only-go-files": onlyGoFilesFS,
		"calculator":    calculatorFS,
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

func ExpectFindings(t *testing.T, want, got find.Findings) {
	t.Helper()

	if !cmp.Equal(want, got) {
		t.Fatalf("unexpected findings:\n%s", cmp.Diff(want, got))
	}
}

func ExpectPatch(t *testing.T, want *patch.Patch, got *patch.Patch) {
	t.Helper()

	wantDryRun, err := want.DryRun()
	if err != nil {
		t.Fatalf("dry run 'want': %v", err)
	}

	dryRun, err := got.DryRun()
	if err != nil {
		t.Fatalf("dry run 'got': %v", err)
	}

	if !cmp.Equal(wantDryRun, dryRun) {
		t.Fatalf("dry run mismatch:\n%s", cmp.Diff(wantDryRun, dryRun))
	}
}

func ExpectGenerationResult(t *testing.T, want, got *generate.Result) {
	t.Helper()

	wgen := want.Generations
	ggen := got.Generations

	less := func(a, b generate.Generation) bool {
		return fmt.Sprintf("%s@%s", a.Path, a.Identifier) <= fmt.Sprintf("%s@%s", b.Path, b.Identifier)
	}

	slices.SortFunc(wgen, less)
	slices.SortFunc(ggen, less)

	if !cmp.Equal(wgen, ggen) {
		t.Fatalf("unexpected generations:\n%s", cmp.Diff(wgen, ggen))
	}
}
