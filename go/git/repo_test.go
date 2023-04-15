package git_test

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/modernice/opendocs/go/git"
	igit "github.com/modernice/opendocs/go/internal/git"
	"github.com/modernice/opendocs/go/internal/tests"
	"github.com/modernice/opendocs/go/patch"
	"github.com/psanford/memfs"
)

var (
	repoRoot = filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "repo")
	g        = igit.Git(repoRoot)
)

func TestRepo_Commit(t *testing.T) {
	tests.WithRepo("basic", repoRoot, func(repoFS fs.FS) {
		codeFile, err := repoFS.Open("foo.go")
		if err != nil {
			t.Fatal(err)
		}

		code, err := io.ReadAll(codeFile)
		if err != nil {
			t.Fatal(err)
		}

		sourceFS := memfs.New()
		if err := sourceFS.WriteFile("foo.go", code, fs.ModePerm); err != nil {
			t.Fatal(err)
		}

		p := patch.New(sourceFS)
		if err := p.Comment("foo.go", "Foo", `Foo is a function that returns a "foo" error.`); err != nil {
			t.Fatal(err)
		}

		repo := git.Repo(repoRoot)

		g.AssertBranch(t, "main")

		if err := repo.Commit(p); err != nil {
			t.Fatal(err)
		}

		g.AssertBranch(t, "opendocs-patch")
		g.AssertCommit(t, "docs: add missing documentation")

		dryRun, err := p.DryRun()
		if err != nil {
			t.Fatal(err)
		}
		wantCode, ok := dryRun["foo.go"]
		if !ok {
			t.Fatal("no code for foo.go in dry run result")
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
