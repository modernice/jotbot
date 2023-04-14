package git_test

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	opendocs "github.com/modernice/opendocs/go"
	"github.com/modernice/opendocs/go/git"
	"github.com/modernice/opendocs/go/internal"
)

var (
	repoRoot = filepath.Join(must(os.Getwd()), "testdata", "gen", "repo")
	g        = internal.Git(repoRoot)
)

func TestRepo_Commit(t *testing.T) {
	internal.WithRepo("basic", repoRoot, func(repoFS fs.FS) {
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

		_, output, err := g.Cmd("branch", "--show-current")
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

		_, output, err = g.Cmd("branch", "--show-current")
		if err != nil {
			t.Fatal(err)
		}
		branch = strings.TrimSpace(string(output))

		if branch != "opendocs-patch" {
			t.Fatalf("unexpected branch %q; want %q", branch, "opendocs-test")
		}

		_, output, err = g.Cmd("log", "-1", "--pretty=%B")
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
	})
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
