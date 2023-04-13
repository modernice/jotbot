package git_test

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	opendocs "github.com/modernice/opendocs/go"
	"github.com/modernice/opendocs/go/git"
)

var (
	repoRoot = filepath.Join(must(os.Getwd()), "testdata", "gen", "repo")
	repoFS   = os.DirFS(repoRoot)

	//go:embed testdata/fixtures/repo
	fixtures embed.FS
)

func init() {
	createRepository()
}

func TestRepo_Commit(t *testing.T) {
	codeFile, err := repoFS.Open("foo.go")
	if err != nil {
		t.Fatal(err)
	}

	code, err := io.ReadAll(codeFile)
	if err != nil {
		t.Fatal(err)
	}

	p, err := opendocs.NewPatcher(code)
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Comment("Foo", `Foo is a function that returns a "foo" error.`); err != nil {
		t.Fatal(err)
	}

	wantCode, err := p.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	repo := git.Repo(repoRoot)

	_, output, err := gitCmd("branch", "--show-current")
	if err != nil {
		t.Fatal(err)
	}
	branch := strings.TrimSpace(string(output))

	if branch != "main" {
		t.Fatalf("unexpected branch %q; want %q", branch, "main")
	}

	if err := repo.Commit("Foo", "foo.go", p); err != nil {
		t.Fatal(err)
	}

	_, output, err = gitCmd("branch", "--show-current")
	if err != nil {
		t.Fatal(err)
	}
	branch = strings.TrimSpace(string(output))

	if branch != "opendocs-patch" {
		t.Fatalf("unexpected branch %q; want %q", branch, "opendocs-test")
	}

	_, output, err = gitCmd("log", "-1", "--pretty=%B")
	if err != nil {
		t.Fatalf("get last commit message: %v", err)
	}
	msg := strings.TrimSpace(string(output))

	wantMsg := "opendocs: add `Foo` comment"
	if msg != wantMsg {
		t.Fatalf("unexpected commit message %q; want %q", msg, wantMsg)
	}

	gotCodeFile, err := repoFS.Open("foo.go")
	if err != nil {
		t.Fatal(err)
	}

	gotCode, err := io.ReadAll(gotCodeFile)
	if err != nil {
		t.Fatal(err)
	}

	if string(gotCode) != string(wantCode) {
		t.Fatalf("unexpected code\n%s", cmp.Diff(string(wantCode), string(gotCode)))
	}
}

func createRepository() error {
	if _, err := os.Stat(repoRoot); !os.IsNotExist(err) {
		if err := os.RemoveAll(repoRoot); err != nil {
			return fmt.Errorf("remove existing repository directory: %w", err)
		}
	}

	if err := os.MkdirAll(repoRoot, 0755); err != nil {
		return fmt.Errorf("create repository directory: %w", err)
	}

	if err := fs.WalkDir(fixtures, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		copy, err := os.Create(filepath.Join(repoRoot, entry.Name()))
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

	if _, _, err := gitCmd("init"); err != nil {
		return err
	}

	if _, _, err := gitCmd("add", "."); err != nil {
		return err
	}

	if _, _, err := gitCmd("commit", "-m", "test commit"); err != nil {
		return err
	}

	return nil
}

func gitCmd(args ...string) (*exec.Cmd, []byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return cmd, out, fmt.Errorf("git: %w", err)
	}
	return cmd, out, nil
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
