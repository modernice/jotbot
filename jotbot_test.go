package jotbot_test

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/modernice/jotbot"
	"github.com/modernice/jotbot/find"
	"github.com/modernice/jotbot/generate"
	"github.com/modernice/jotbot/generate/mockgenerate"
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
		{File: "foo.go", Finding: find.Finding{Identifier: "Foo"}, Language: "go"},
		{File: "bar.go", Finding: find.Finding{Identifier: "Foo"}, Language: "go"},
		{File: "bar.go", Finding: find.Finding{Identifier: "Bar"}, Language: "go"},
		{File: "baz.go", Finding: find.Finding{Identifier: "X"}, Language: "go"},
		{File: "baz.go", Finding: find.Finding{Identifier: "X.Foo"}, Language: "go"},
		{File: "baz.go", Finding: find.Finding{Identifier: "(*X).Bar"}, Language: "go"},
		{File: "baz.go", Finding: find.Finding{Identifier: "Y.Foo"}, Language: "go"},
	}, findings)
}

func TestJotBot_Generate(t *testing.T) {
	svc := mockgenerate.NewMockService()
	svc.GenerateDocFunc.SetDefaultHook(func(ctx generate.Context) (string, error) {
		switch ctx.Input().Identifier {
		case "Foo":
			return "Foo is a foo.", nil
		case "Bar":
			return "Bar is a bar.", nil
		default:
			return "", fmt.Errorf("unknown identifier %q", ctx.Input().Identifier)
		}
	})

	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "generate")
	tests.WithRepo("basic", root, func(repo fs.FS) {
		bot := jotbot.New(root)

		findings := append(
			makeFindings(
				"foo.go",
				find.Finding{Identifier: "Foo"},
			),
			makeFindings(
				"bar.go",
				find.Finding{Identifier: "Foo"},
				find.Finding{Identifier: "Bar"},
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
		out[i] = jotbot.Finding{File: file, Finding: f, Language: "go"}
	}
	return out
}
