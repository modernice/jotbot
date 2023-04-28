package jotbot_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/modernice/jotbot"
	"github.com/modernice/jotbot/find"
	igen "github.com/modernice/jotbot/internal/generate"
	"github.com/modernice/jotbot/internal/tests"
)

func TestJotBot_Find(t *testing.T) {
	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "find")
	tests.InitRepo("basic", root)

	bot := jotbot.New(root)

	findings, err := bot.Find(context.Background())
	if err != nil {
		t.Fatalf("Find() failed: %v", err)
	}

	tests.ExpectFound(t, []jotbot.Finding{
		{File: "foo.go", Finding: find.Finding{Identifier: "Foo", Target: "function 'Foo'"}},
		{File: "bar.go", Finding: find.Finding{Identifier: "Foo", Target: "const 'Foo'"}},
		{File: "bar.go", Finding: find.Finding{Identifier: "Bar", Target: "type 'Bar'"}},
		{File: "baz.go", Finding: find.Finding{Identifier: "X", Target: "type 'X'"}},
		{File: "baz.go", Finding: find.Finding{Identifier: "X.Foo", Target: "method 'X.Foo'"}},
		{File: "baz.go", Finding: find.Finding{Identifier: "(*X).Bar", Target: "method '(*X).Bar'"}},
		{File: "baz.go", Finding: find.Finding{Identifier: "Y.Foo", Target: "method 'Y.Foo'"}},
	}, findings)
}

func TestJotBot_Generate(t *testing.T) {
	svc := igen.MockService().
		WithDoc("Foo", "Foo is a foo.").
		WithDoc("Bar", "Bar is a bar.")

	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "generate")
	tests.WithRepo("basic", root, func(repo fs.FS) {
		bot := jotbot.New(root)

		findings := append(
			makeFindings(
				"foo.go",
				find.Finding{
					Identifier: "Foo",
					Target:     "function 'Foo'",
				},
			),
			makeFindings(
				"bar.go",
				find.Finding{
					Identifier: "Foo",
					Target:     "const 'Foo'",
				},
				find.Finding{
					Identifier: "Bar",
					Target:     "type 'Bar'",
				},
			)...,
		)

		patch, err := bot.Generate(context.Background(), findings, svc)
		if err != nil {
			t.Fatalf("Generate() failed: %v", err)
		}

		if err := patch.Apply(context.Background(), root); err != nil {
			t.Fatalf("patch.Apply() failed: %v", err)
		}

		tests.ExpectCommentIn(t, repo, "foo.go", "Foo", "Foo is a foo.")
		tests.ExpectCommentIn(t, repo, "bar.go", "Foo", "Foo is a foo.")
		tests.ExpectCommentIn(t, repo, "bar.go", "Bar", "Bar is a bar.")
	})
}

func makeFindings(file string, findings ...find.Finding) []jotbot.Finding {
	out := make([]jotbot.Finding, len(findings))
	for i, f := range findings {
		out[i] = jotbot.Finding{File: file, Finding: f}
	}
	return out
}
