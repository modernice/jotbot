package opendocs_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	opendocs "github.com/modernice/opendocs/go"
	"github.com/modernice/opendocs/go/internal"
)

func TestFinder_Uncommented(t *testing.T) {
	root := filepath.Join(internal.Must(os.Getwd()), "testdata", "gen", "basic")

	internal.WithRepo("basic", root, func(repoFS fs.FS) {
		f := opendocs.NewFinder(repoFS)

		result, err := f.Uncommented()
		if err != nil {
			t.Fatal(err)
		}

		internal.AssertFindings(t, opendocs.Findings{
			"foo.go": {{Path: "foo.go", Identifier: "Foo"}},
			"bar.go": {
				{Path: "bar.go", Identifier: "Foo"},
				{Path: "bar.go", Identifier: "Bar"},
			},
		}, result)
	})
}

func TestFinder_Find_onlyGoFiles(t *testing.T) {
	root := filepath.Join(internal.Must(os.Getwd()), "testdata", "gen", "only-go-files")

	internal.WithRepo("only-go-files", root, func(repoFS fs.FS) {
		f := opendocs.NewFinder(repoFS)

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

		internal.AssertFindings(t, opendocs.Findings{
			"foo.go": {{Path: "foo.go", Identifier: "Foo"}},
		}, result)
	})
}
