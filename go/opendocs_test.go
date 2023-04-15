package opendocs_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	opendocs "github.com/modernice/opendocs/go"
	"github.com/modernice/opendocs/go/generate"
	igen "github.com/modernice/opendocs/go/internal/generate"
	"github.com/modernice/opendocs/go/internal/tests"
)

func TestRepository_Generate(t *testing.T) {
	sourceRoot := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "opendocs-generate-source")
	tests.InitRepo("basic", sourceRoot)

	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "opendocs-generate")
	tests.WithRepo("basic", root, func(repoFS fs.FS) {
		repo := opendocs.Repo(root)

		svc := igen.MockService().
			WithDoc("foo.go", "Foo", "Foo is a function that always returns an error.").
			WithDoc("bar.go", "Foo", `Foo is always "foo".`).
			WithDoc("bar.go", "Bar", "Bar is an empty struct.")

		result, err := repo.Generate(context.Background(), svc)
		if err != nil {
			t.Fatalf("generate documentation: %v", err)
		}

		tests.ExpectGenerationResult(t, generate.NewResult(repoFS, svc.Generations()...), result)
	})
}
