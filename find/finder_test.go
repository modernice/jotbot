package find_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/modernice/opendocs/find"
	"github.com/modernice/opendocs/internal/tests"
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
