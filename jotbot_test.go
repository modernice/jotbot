package jotbot_test

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/modernice/jotbot"
	"github.com/modernice/jotbot/generate"
	"github.com/modernice/jotbot/generate/mockgenerate"
	"github.com/modernice/jotbot/internal/tests"
	"github.com/modernice/jotbot/langs/golang"
)

func TestJotBot_Find(t *testing.T) {
	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "find")
	tests.InitRepo("basic", root)

	bot := newJotBot(root)

	findings, err := bot.Find(context.Background())
	if err != nil {
		t.Fatalf("Find() failed: %v", err)
	}

	tests.ExpectFound(t, []jotbot.Finding{
		{File: "foo.go", Identifier: "func:Foo", Language: "go"},
		{File: "bar.go", Identifier: "var:Foo", Language: "go"},
		{File: "bar.go", Identifier: "type:Bar", Language: "go"},
		{File: "baz.go", Identifier: "type:X", Language: "go"},
		{File: "baz.go", Identifier: "func:X.Foo", Language: "go"},
		{File: "baz.go", Identifier: "func:(*X).Bar", Language: "go"},
		{File: "baz.go", Identifier: "func:Y.Foo", Language: "go"},
	}, findings)
}

func TestJotBot_Generate(t *testing.T) {
	svc := mockgenerate.NewMockService()
	svc.GenerateDocFunc.SetDefaultHook(func(ctx generate.Context) (string, error) {
		switch ctx.Input().Identifier {
		case "func:Foo":
			return "Foo is a foo.", nil
		case "var:Foo":
			return "Foo is a foo.", nil
		case "type:Bar":
			return "Bar is a bar.", nil
		default:
			return "", fmt.Errorf("unknown identifier %q", ctx.Input().Identifier)
		}
	})

	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "generate")
	tests.WithRepo("basic", root, func(repo fs.FS) {
		bot := newJotBot(root)

		findings := append(
			makeFindings(
				"foo.go",
				"func:Foo",
			),
			makeFindings(
				"bar.go",
				"var:Foo",
				"type:Bar",
			)...,
		)

		patch, err := bot.Generate(context.Background(), findings, svc)
		if err != nil {
			t.Fatalf("Generate() failed: %v", err)
		}

		if err := patch.Apply(context.Background(), root); err != nil {
			t.Fatalf("patch.Apply() failed: %v", err)
		}

		tests.ExpectCommentIn(t, repo, "foo.go", "func:Foo", "Foo is a foo.")
		tests.ExpectCommentIn(t, repo, "bar.go", "var:Foo", "Foo is a foo.")
		tests.ExpectCommentIn(t, repo, "bar.go", "type:Bar", "Bar is a bar.")
	})
}

func TestMatch(t *testing.T) {
	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "filter")
	tests.InitRepo("basic", root)

	filters := []*regexp.Regexp{
		regexp.MustCompile(`^(var:|type:)`),
		regexp.MustCompile(`^func:\(\*X\)\.Bar$`),
	}

	bot := newJotBot(root, jotbot.Match(filters...))

	findings, err := bot.Find(context.Background())
	if err != nil {
		t.Fatalf("Find() failed: %v", err)
	}

	tests.ExpectFound(t, []jotbot.Finding{
		{Identifier: "var:Foo", Language: "go", File: "bar.go"},
		{Identifier: "type:Bar", Language: "go", File: "bar.go"},
		{Identifier: "type:X", Language: "go", File: "baz.go"},
		{Identifier: "func:(*X).Bar", Language: "go", File: "baz.go"},
	}, findings)
}

func makeFindings(file string, findings ...string) []jotbot.Finding {
	out := make([]jotbot.Finding, len(findings))
	for i, id := range findings {
		out[i] = jotbot.Finding{File: file, Identifier: id, Language: "go"}
	}
	return out
}

func newJotBot(root string, opts ...jotbot.Option) *jotbot.JotBot {
	bot := jotbot.New(root, opts...)
	bot.ConfigureLanguage("go", golang.Must())
	return bot
}
