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

// func TestFinder_Uncommented_duplicateName(t *testing.T) {
// 	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "duplicate-name")

// 	tests.WithRepo("duplicate-name", root, func(repoFS fs.FS) {
// 		f := golang.NewFinder(repoFS)

// 		result, err := f.Uncommented()
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		tests.ExpectFindings(t, golang.Findings{
// 			"foo.go": {
// 				{Path: "foo.go", Identifier: "Foo"},
// 				{Path: "foo.go", Identifier: "X"},
// 				{Path: "foo.go", Identifier: "Y"},
// 				{Path: "foo.go", Identifier: "X.Foo"},
// 				{Path: "foo.go", Identifier: "*Y.Foo"},
// 			},
// 		}, result)
// 	})
// }

// func TestFinder_Uncommented_generic(t *testing.T) {
// 	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "generic")

// 	tests.WithRepo("generic", root, func(repoFS fs.FS) {
// 		f := golang.NewFinder(repoFS)

// 		result, err := f.Uncommented()
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		tests.ExpectFindings(t, golang.Findings{
// 			"foo.go": {
// 				{Path: "foo.go", Identifier: "Foo"},
// 				{Path: "foo.go", Identifier: "X"},
// 				{Path: "foo.go", Identifier: "*X.Foo"},
// 				{Path: "foo.go", Identifier: "*y.Foo"},
// 			},
// 		}, result)
// 	})
// }

// func TestGlob(t *testing.T) {
// 	all := []string{
// 		"foo.go",
// 		"bar.go",
// 		"baz.go",
// 		"foo/foo.go",
// 		"foo/bar.go",
// 		"foo/baz.go",
// 		"bar/foo.go",
// 		"bar/bar.go",
// 		"bar/baz.go",
// 		"baz/foo.go",
// 		"baz/bar.go",
// 		"baz/baz.go",
// 	}

// 	onlyFoo := slice.Filter(all, func(file string) bool {
// 		return filepath.Base(file) == "foo.go"
// 	})
// 	onlyBar := slice.Filter(all, func(file string) bool {
// 		return filepath.Base(file) == "bar.go"
// 	})

// 	cases := []struct {
// 		name    string
// 		pattern string
// 		want    []string
// 	}{
// 		{
// 			name:    "empty pattern",
// 			pattern: "",
// 			want:    all,
// 		},
// 		{
// 			name:    "only foo.go",
// 			pattern: "**/foo.go",
// 			want:    onlyFoo,
// 		},
// 		{
// 			name:    "only bar.go",
// 			pattern: "**/bar.go",
// 			want:    onlyBar,
// 		},
// 		{
// 			name:    "everything within foo/",
// 			pattern: "foo/*",
// 			want:    []string{"foo/foo.go", "foo/bar.go", "foo/baz.go"},
// 		},
// 		{
// 			name:    "all ba*.go",
// 			pattern: "**/ba*.go",
// 			want: []string{
// 				"bar.go", "baz.go",
// 				"foo/bar.go", "foo/baz.go",
// 				"bar/bar.go", "bar/baz.go",
// 				"baz/bar.go", "baz/baz.go"},
// 		},
// 	}

// 	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "glob")

// 	tests.WithRepo("glob", root, func(repoFS fs.FS) {
// 		for _, tt := range cases {
// 			t.Run(tt.name, func(t *testing.T) {
// 				f := golang.NewFinder(repoFS, golang.Glob(tt.pattern))

// 				result, err := f.Uncommented()
// 				if err != nil {
// 					t.Fatal(err)
// 				}

// 				got := maps.Keys(result)
// 				slices.Sort(got)
// 				slices.Sort(tt.want)

// 				if len(got) != len(tt.want) {
// 					t.Fatalf("got findings for %d files; want %d\nwant files: %v\nfound files: %v", len(got), len(tt.want), tt.want, got)
// 				}

// 				for _, wantFile := range tt.want {
// 					if _, ok := result[wantFile]; !ok {
// 						t.Errorf("file %q not found in findings", wantFile)
// 					}
// 				}
// 			})
// 		}
// 	})
// }
