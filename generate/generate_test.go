package generate_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/modernice/jotbot/generate"
	"github.com/modernice/jotbot/generate/mockgenerate"
	"github.com/modernice/jotbot/internal"
	"github.com/modernice/jotbot/langs/golang"
)

func TestGenerator_Generate(t *testing.T) {
	doc := "Add adds two integers."
	svc := mockgenerate.NewMockService()
	svc.GenerateDocFunc.PushReturn(doc, nil)

	code := heredoc.Doc(`
		package foo

		func Add(a, b int) int {
			return a + b
		}
	`)

	g := generate.New(svc, generate.WithLanguage("go", golang.Must()))
	in := generate.PromptInput{
		File: "foo.go",
		Input: generate.Input{
			Code:       []byte(code),
			Language:   "go",
			Identifier: "Add",
		},
	}

	got, err := g.Generate(context.Background(), in)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	if got != doc {
		t.Fatalf("Generate() returned wrong documentation\n%s", cmp.Diff(doc, got))
	}
}

func TestGenerator_Files(t *testing.T) {
	svc := mockgenerate.NewMockService()
	svc.GenerateDocFunc.SetDefaultHook(func(ctx generate.Context) (string, error) {
		switch ctx.Input().Identifier {
		case "func:Foo":
			return "Foo is a function.", nil
		case "var:Foo":
			return "Foo is a variable.", nil
		case "type:Bar":
			return "Bar is a struct.", nil
		default:
			return "", fmt.Errorf("unknown identifier %q", ctx.Input().Identifier)
		}
	})

	g := generate.New(svc, generate.WithLanguage("go", golang.Must()))

	files := map[string][]generate.Input{
		"foo.go": {{Identifier: "func:Foo", Language: "go"}},
		"bar.go": {{Identifier: "var:Foo", Language: "go"}, {Identifier: "type:Bar", Language: "go"}},
	}

	gens, errs, err := g.Files(context.Background(), files)
	if err != nil {
		t.Fatalf("Files() failed: %v", err)
	}

	got := drain(t, gens, errs)

	expectGenerated(t, got, "foo.go", "func:Foo", "Foo is a function.")
	expectGenerated(t, got, "bar.go", "var:Foo", "Foo is a variable.")
	expectGenerated(t, got, "bar.go", "type:Bar", "Bar is a struct.")
}

func TestGenerate_Files_workers(t *testing.T) {
	svc := mockgenerate.NewMockService()
	svc.GenerateDocFunc.SetDefaultHook(func(ctx generate.Context) (string, error) {
		switch ctx.Input().Identifier {
		case "var:Foo":
			return "Foo is a variable.", nil
		default:
			return "", fmt.Errorf("unknown identifier %q", ctx.Input().Identifier)
		}
	})

	g := generate.New(
		svc,
		generate.WithLanguage("go", golang.Must()),
		generate.Workers(2, 1),
	)

	files := map[string][]generate.Input{
		"foo.go":    {{Identifier: "var:Foo", Language: "go"}},
		"bar.go":    {{Identifier: "var:Foo", Language: "go"}},
		"baz.go":    {{Identifier: "var:Foo", Language: "go"}},
		"foobar.go": {{Identifier: "var:Foo", Language: "go"}},
		"barbaz.go": {{Identifier: "var:Foo", Language: "go"}},
	}

	gens, errs, err := g.Files(context.Background(), files)
	if err != nil {
		t.Fatalf("Files() failed: %v", err)
	}

	got := drain(t, gens, errs)

	expectGenerated(t, got, "foo.go", "var:Foo", "Foo is a variable.")
	expectGenerated(t, got, "bar.go", "var:Foo", "Foo is a variable.")
	expectGenerated(t, got, "baz.go", "var:Foo", "Foo is a variable.")
	expectGenerated(t, got, "foobar.go", "var:Foo", "Foo is a variable.")
	expectGenerated(t, got, "barbaz.go", "var:Foo", "Foo is a variable.")
}

func TestFooter(t *testing.T) {
	svc := mockgenerate.NewMockService()
	svc.GenerateDocFunc.PushReturn("Foo is a dummy function.", nil)

	g := generate.New(svc, generate.Footer("This is a footer."), generate.WithLanguage("go", golang.Must()))

	doc, err := g.Generate(context.Background(), generate.PromptInput{
		File: "foo.go",
		Input: generate.Input{
			Code:       []byte("package foo\n\nfunc Foo() {}"),
			Language:   "go",
			Identifier: "Foo",
		},
	})
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	want := "Foo is a dummy function.\n\nThis is a footer."

	if doc != want {
		t.Fatalf("Generate() returned wrong documentation\n%s", cmp.Diff(want, doc))
	}
}

func TestLimit(t *testing.T) {
	svc := mockgenerate.NewMockService()
	svc.GenerateDocFunc.SetDefaultHook(func(ctx generate.Context) (string, error) {
		switch ctx.Input().Identifier {
		case "Foo":
			return "Foo is a dummy.", nil
		case "Bar":
			return "Bar is a dummy.", nil
		case "Baz":
			return "Baz is a dummy.", nil
		default:
			return "", fmt.Errorf("unknown identifier %q", ctx.Input().Identifier)
		}
	})
	g := generate.New(svc, generate.Limit(2), generate.WithLanguage("go", golang.Must()))

	files := map[string][]generate.Input{
		"foo.go": {{Identifier: "Foo", Language: "go"}},
		"bar.go": {{Identifier: "Foo", Language: "go"}, {Identifier: "Bar", Language: "go"}},
		"baz.go": {{Identifier: "Foo", Language: "go"}, {Identifier: "Bar", Language: "go"}, {Identifier: "Baz", Language: "go"}},
	}

	gens, errs, err := g.Files(context.Background(), files)
	if err != nil {
		t.Fatalf("Files() failed: %v", err)
	}

	got := drain(t, gens, errs)

	if n := len(got); n > 2 {
		t.Fatalf("Files() returned %d files; want 2\n%v", n, got)
	}
}

func expectGenerated(t *testing.T, gens []generate.File, file, identifier, doc string) {
	t.Helper()

	var found string
L:
	for _, gen := range gens {
		if file != "" && gen.Path != file {
			continue
		}

		for _, g := range gen.Docs {
			if g.Identifier == identifier && g.Text == doc {
				found = g.Text
				break L
			}
		}
	}

	if found != doc {
		t.Fatalf("unexpected generation for identifier %q in %s\n\nwant:\n%s\n\ngot:\n%s\n\n%v", identifier, file, doc, found, gens)
	}
}

func drain[T any](t *testing.T, vals <-chan T, errs <-chan error) []T {
	t.Helper()

	out, err := internal.Drain(vals, errs)
	if err != nil {
		t.Fatalf("drain generations: %v", err)
	}
	return out
}
