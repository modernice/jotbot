package jotbot_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/modernice/jotbot"
	"github.com/modernice/jotbot/generate"
	igen "github.com/modernice/jotbot/internal/generate"
	"github.com/modernice/jotbot/internal/tests"
)

func TestRepository_Generate(t *testing.T) {
	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "generate")

	tests.WithRepo("basic", root, func(repoFS fs.FS) {
		svc := igen.MockService().
			WithDoc("foo.go", "Foo", "Foo is a function that returns a \"foo\" error.").
			WithDoc("bar.go", "Bar", "Bar is an empty struct.").
			WithDoc("baz.go", "Y.Foo", "Foo is a method of Y.")
		svc.Fallbacks = true

		repo := jotbot.New(svc)

		p, err := repo.Generate(context.Background(), root)
		if err != nil {
			t.Fatalf("Generate() should not return error; got %q", err)
		}

		want := generate.Patch(repoFS, svc.Generations())

		tests.ExpectPatch(t, want, p)
	})
}
