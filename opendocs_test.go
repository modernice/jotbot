package opendocs_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	opendocs "github.com/modernice/opendocs"
	igen "github.com/modernice/opendocs/internal/generate"
	"github.com/modernice/opendocs/internal/tests"
)

func TestRepository_Generate(t *testing.T) {
	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "opendocs-generate")
	tests.WithRepo("basic", root, func(fs.FS) {
		repo := opendocs.Repo(root)

		svc := igen.MockService().
			WithDoc("foo.go", "Foo", "Foo is a function that always returns an error.").
			WithDoc("bar.go", "Foo", `Foo is always "foo".`).
			WithDoc("bar.go", "Bar", "Bar is an empty struct.").
			WithDoc("baz.go", "X", "X is an empty struct.").
			WithDoc("baz.go", "X.Foo", "Foo is a method of X.").
			WithDoc("baz.go", "*X.Bar", "Bar is a method of X.").
			WithDoc("baz.go", "Y.Foo", "Foo is a method of Y.")

		result, err := repo.Generate(context.Background(), svc)
		if err != nil {
			t.Fatalf("generate documentation: %v", err)
		}

		tests.ExpectGenerations(t, svc.Generations(), result)
	})
}
