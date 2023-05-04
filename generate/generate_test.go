package generate_test

import (
	"context"
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/modernice/jotbot/generate"
	"github.com/modernice/jotbot/internal"
	igen "github.com/modernice/jotbot/internal/generate"
)

func TestGenerator_Generate(t *testing.T) {
	doc := "Add adds two integers."
	svc := igen.MockService().WithDoc("Add", doc)

	code := heredoc.Doc(`
		package foo

		func Add(a, b int) int {
			return a + b
		}
	`)

	g := generate.New(svc)
	in := generate.Input{
		Code:       []byte(code),
		Identifier: "Add",
		Target:     "function 'Add'",
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
	svc := igen.MockService().
		WithDoc("Foo", "Foo is a function.").
		WithDoc("Bar", "Bar is a struct.")

	g := generate.New(svc)

	files := map[string][]generate.Input{
		"foo.go": {{Identifier: "Foo"}},
		"bar.go": {{Identifier: "Foo"}, {Identifier: "Bar"}},
	}

	gens, errs, err := g.Files(context.Background(), files)
	if err != nil {
		t.Fatalf("Files() failed: %v", err)
	}

	got := drain(t, gens, errs)

	expectGenerated(t, got, "Foo", "Foo is a function.")
	expectGenerated(t, got, "Bar", "Bar is a struct.")
}

func TestFooter(t *testing.T) {
	svc := igen.MockService().WithDoc("Foo", "Foo is a dummy function.")
	g := generate.New(svc, generate.Footer("This is a footer."))

	doc, err := g.Generate(context.Background(), generate.Input{
		Code:       []byte("package foo\n\nfunc Foo() {}"),
		Identifier: "Foo",
		Target:     "function 'Foo'",
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
	svc := igen.MockService().
		WithDoc("Foo", "Foo is a dummy function.").
		WithDoc("Bar", "Bar is a dummy struct.")
	g := generate.New(svc, generate.Limit(2))

	files := map[string][]generate.Input{
		"foo.go": {{Identifier: "Foo"}},
		"bar.go": {{Identifier: "Foo"}, {Identifier: "Bar"}},
		"baz.go": {{Identifier: "Foo"}, {Identifier: "Bar"}, {Identifier: "Baz"}},
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

// func TestWithLanguage_Prompt(t *testing.T) {
// 	svc := mockgenerate.NewMockService()
// 	svc.GenerateDocFunc.PushHook(func(ctx generate.Context) (string, error) {
// 	})

// 	g := generate.New(svc, generate.WithLanguage(".go", golang.New()))

// 	doc, err := g.Generate(context.Background(), generate.Input{
// 		Code:       []byte("package foo\n\nfunc Foo() {}"),
// 		Identifier: "Foo",
// 		Target:     "function 'Foo'",
// 	})
// 	if err != nil {
// 		t.Fatalf("Generate() failed: %v", err)
// 	}
// }

func expectGenerated(t *testing.T, gens []generate.File, identifier, doc string) {
	t.Helper()

	var found string
L:
	for _, gen := range gens {
		for _, g := range gen.Docs {
			if g.Identifier == identifier && g.Text == doc {
				found = g.Text
				break L
			}
		}
	}

	if found != doc {
		t.Fatalf("unexpected generation for identifier %q\n\nwant:\n%s\n\ngot:\n%s\n\n%v", identifier, doc, found, gens)
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
