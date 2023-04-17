package generate_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/modernice/opendocs/go/generate"
	igen "github.com/modernice/opendocs/go/internal/generate"
	"github.com/modernice/opendocs/go/internal/tests"
	"github.com/modernice/opendocs/go/patch"
)

func mockService(repo fs.FS) (*igen.Service, *generate.Result) {
	svc := igen.MockService().
		WithDoc("operations.go", "Add", "Add adds numbers together.").
		WithDoc("operations.go", "Sub", "Sub subtracts numbers.").
		WithDoc("operations.go", "Mul", "Mul multiplies numbers.").
		WithDoc("operations.go", "Div", "Div divides numbers.")
	return svc, generate.NewResult(
		repo,
		generate.Generation{
			Path:       "operations.go",
			Identifier: "Add",
			Doc:        "Add adds numbers together.",
		},
		generate.Generation{
			Path:       "operations.go",
			Identifier: "Sub",
			Doc:        "Sub subtracts numbers.",
		},
		generate.Generation{
			Path:       "operations.go",
			Identifier: "Mul",
			Doc:        "Mul multiplies numbers.",
		},
		generate.Generation{
			Path:       "operations.go",
			Identifier: "Div",
			Doc:        "Div divides numbers.",
		},
	)
}

func TestGenerator_Generate(t *testing.T) {
	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "calculator")
	tests.WithRepo("calculator", root, func(repoFS fs.FS) {
		svc, want := mockService(repoFS)
		g := generate.New(svc)

		result, err := g.Generate(context.Background(), repoFS)
		if err != nil {
			t.Fatalf("Generate() should not return an error; got %q", err)
		}

		tests.ExpectGenerationResult(t, want, result)
	})
}

func TestResult_Patch(t *testing.T) {
	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "calculator")
	tests.WithRepo("calculator", root, func(repoFS fs.FS) {
		svc, mockResult := mockService(repoFS)
		g := generate.New(svc)

		result, err := g.Generate(context.Background(), repoFS)
		if err != nil {
			t.Fatalf("Generate() should not return an error; got %q", err)
		}

		p := result.Patch()

		want := patch.New(repoFS)
		for _, gen := range mockResult.Generations {
			want.Comment(gen.Path, gen.Identifier, gen.Doc)
		}

		tests.ExpectPatch(t, want, p)
	})
}
