package git_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/modernice/jotbot/git"
	"github.com/modernice/jotbot/internal/git/gittest"
	"github.com/modernice/jotbot/internal/patch"
	"github.com/modernice/jotbot/internal/tests"
)

var (
	repoRoot = filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "repo")
	g        = gittest.Git(repoRoot)
)

func TestRepo_Commit(t *testing.T) {
	repo := git.Repo(repoRoot)

	p := patch.Mock(map[string]string{
		"foo.go": heredoc.Doc(`
			package foo

			// Foo does nothing.
			func Foo() {}
		`),
	})

	if err := repo.Commit(context.Background(), p); err != nil {
		t.Fatal(err)
	}

	g.AssertBranchPrefix(t, "jotbot-patch")
	g.AssertCommit(t, git.Commit{
		Msg: "docs: add missing documentation",
		Desc: []string{
			"Updated docs:",
			"  - foo.go@Foo",
		},
		Footer: "This commit was created by jotbot.",
	})

	repoFS := os.DirFS(repoRoot)

	f, err := repoFS.Open("foo.go")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	tests.ExpectComment(t, "Foo", "Foo does nothing.", f)
}
