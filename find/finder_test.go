package find_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/modernice/jotbot/find"
	"github.com/modernice/jotbot/internal/slice"
	"github.com/modernice/jotbot/internal/tests"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func TestFinder_Uncommented(t *testing.T) {
	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "basic")

	tests.WithRepo("basic", root, func(repoFS fs.FS) {
		f := find.New(repoFS)

		result, err := f.Uncommented()
		if err != nil {
			t.Fatal(err)
		}

		tests.ExpectFindings(t, find.Findings{
			"foo.go": {{Path: "foo.go", Identifier: "Foo"}},
			"bar.go": {
				{Path: "bar.go", Identifier: "Bar"},
				{Path: "bar.go", Identifier: "Foo"},
			},
			"baz.go": {
				{Path: "baz.go", Identifier: "*X.Bar"},
				{Path: "baz.go", Identifier: "X"},
				{Path: "baz.go", Identifier: "X.Foo"},
				{Path: "baz.go", Identifier: "Y.Foo"},
			},
		}, result)
	})
}

func TestFinder_Find_onlyGoFiles(t *testing.T) {
	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "only-go-files")

	tests.WithRepo("only-go-files", root, func(repoFS fs.FS) {
		f := find.New(repoFS)

		result, err := f.Uncommented()
		if err != nil {
			t.Fatalf("find uncommented code: %v", err)
		}

		if _, ok := result["bar.ts"]; ok {
			t.Fatalf("only Go files should be returned in findings; got 'bar.ts'")
		}

		if _, ok := result["foo.go"]; !ok {
			t.Fatalf("Go files should be returned in findings; got no 'foo.go'")
		}

		tests.ExpectFindings(t, find.Findings{
			"foo.go": {{Path: "foo.go", Identifier: "Foo"}},
		}, result)
	})
}

func TestFinder_Uncommented_duplicateName(t *testing.T) {
	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "duplicate-name")

	tests.WithRepo("duplicate-name", root, func(repoFS fs.FS) {
		f := find.New(repoFS)

		result, err := f.Uncommented()
		if err != nil {
			t.Fatal(err)
		}

		tests.ExpectFindings(t, find.Findings{
			"foo.go": {
				{Path: "foo.go", Identifier: "Foo"},
				{Path: "foo.go", Identifier: "X"},
				{Path: "foo.go", Identifier: "Y"},
				{Path: "foo.go", Identifier: "X.Foo"},
				{Path: "foo.go", Identifier: "*Y.Foo"},
			},
		}, result)
	})
}

func TestFinder_Uncommented_generic(t *testing.T) {
	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "generic")

	tests.WithRepo("generic", root, func(repoFS fs.FS) {
		f := find.New(repoFS)

		result, err := f.Uncommented()
		if err != nil {
			t.Fatal(err)
		}

		tests.ExpectFindings(t, find.Findings{
			"foo.go": {
				{Path: "foo.go", Identifier: "Foo"},
				{Path: "foo.go", Identifier: "X"},
				{Path: "foo.go", Identifier: "*X.Foo"},
				{Path: "foo.go", Identifier: "*y.AggregateID"},
			},
		}, result)
	})
}

func TestGlob(t *testing.T) {
	all := []string{
		"foo.go",
		"bar.go",
		"baz.go",
		"foo/foo.go",
		"foo/bar.go",
		"foo/baz.go",
		"bar/foo.go",
		"bar/bar.go",
		"bar/baz.go",
		"baz/foo.go",
		"baz/bar.go",
		"baz/baz.go",
	}

	onlyFoo := slice.Filter(all, func(file string) bool {
		return filepath.Base(file) == "foo.go"
	})
	onlyBar := slice.Filter(all, func(file string) bool {
		return filepath.Base(file) == "bar.go"
	})

	cases := []struct {
		name    string
		pattern string
		want    []string
	}{
		{
			name:    "empty pattern",
			pattern: "",
			want:    all,
		},
		{
			name:    "only foo.go",
			pattern: "**/foo.go",
			want:    onlyFoo,
		},
		{
			name:    "only bar.go",
			pattern: "**/bar.go",
			want:    onlyBar,
		},
		{
			name:    "everything within foo/",
			pattern: "foo/*",
			want:    []string{"foo/foo.go", "foo/bar.go", "foo/baz.go"},
		},
		{
			name:    "all ba*.go",
			pattern: "**/ba*.go",
			want: []string{
				"bar.go", "baz.go",
				"foo/bar.go", "foo/baz.go",
				"bar/bar.go", "bar/baz.go",
				"baz/bar.go", "baz/baz.go"},
		},
	}

	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "glob")

	tests.WithRepo("glob", root, func(repoFS fs.FS) {
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				f := find.New(repoFS, find.Glob(tt.pattern))

				result, err := f.Uncommented()
				if err != nil {
					t.Fatal(err)
				}

				got := maps.Keys(result)
				slices.Sort(got)
				slices.Sort(tt.want)

				if len(got) != len(tt.want) {
					t.Fatalf("got findings for %d files; want %d\nwant files: %v\nfound files: %v", len(got), len(tt.want), tt.want, got)
				}

				for _, wantFile := range tt.want {
					if _, ok := result[wantFile]; !ok {
						t.Errorf("file %q not found in findings", wantFile)
					}
				}
			})
		}
	})
}
