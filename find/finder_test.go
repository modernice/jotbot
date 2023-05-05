package find_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/modernice/jotbot/find"
	"github.com/modernice/jotbot/internal/tests"
)

func TestOptions_Find_defaultOptions(t *testing.T) {
	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "basic")

	tests.WithRepo("basic", root, func(repoFS fs.FS) {
		got, err := find.Files(context.Background(), repoFS)
		if err != nil {
			t.Fatal(err)
		}

		tests.ExpectFiles(t, []string{
			"foo.go",
			"bar.go",
			"baz.go",
		}, got)
	})
}

func TestOptions_Extensions(t *testing.T) {
	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "extensions")

	tests.WithRepo("extensions", root, func(repoFS fs.FS) {
		cases := map[string][]string{
			".go": {"foo.go", "bar.go"},
			".ts": {"bar.ts", "baz.ts"},
		}

		for ext, want := range cases {
			t.Run(ext, func(t *testing.T) {
				got, err := find.Options{Extensions: []string{ext}}.Find(context.Background(), repoFS)
				if err != nil {
					t.Fatal(err)
				}

				tests.ExpectFiles(t, want, got)
			})
		}
	})
}

func TestOptions_Include(t *testing.T) {
	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "include")

	tests.WithRepo("glob", root, func(repoFS fs.FS) {
		got, err := find.Options{
			Include: []string{"**/{foo,baz}.go"},
		}.Find(context.Background(), repoFS)

		if err != nil {
			t.Fatal(err)
		}

		tests.ExpectFiles(t, []string{
			"foo.go",
			"baz.go",
			"foo/foo.go",
			"foo/baz.go",
			"bar/foo.go",
			"bar/baz.go",
			"baz/foo.go",
			"baz/baz.go",
		}, got)
	})
}

func TestOptions_Exclude(t *testing.T) {
	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "exclude")

	tests.WithRepo("glob", root, func(repoFS fs.FS) {
		got, err := find.Options{
			Exclude: []string{"**/{foo,baz}.go"},
		}.Find(context.Background(), repoFS)

		if err != nil {
			t.Fatal(err)
		}

		tests.ExpectFiles(t, []string{
			"bar.go",
			"foo/bar.go",
			"bar/bar.go",
			"baz/bar.go",
		}, got)
	})
}
