package patch_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/google/go-cmp/cmp"
	"github.com/modernice/opendocs/go/git"
	"github.com/modernice/opendocs/go/internal/tests"
	"github.com/modernice/opendocs/go/patch"
	"github.com/psanford/memfs"
)

var _ interface {
	git.Patch
	git.IdentifierProvider
} = (*patch.Patch)(nil)

var dryRunTests = []struct {
	name       string
	comment    string
	identifier string
	input      func(f *jen.File) string
}{
	{
		name:       "function",
		comment:    `Foo is a function that returns a "foo" error.`,
		identifier: "Foo",
		input: func(f *jen.File) string {
			f.Func().Id("Foo").Params().Error()
			return f.GoString()
		},
	},
	{
		name:       "struct",
		comment:    `Foo is a struct that does things.`,
		identifier: "Foo",
		input: func(f *jen.File) string {
			f.Type().Id("Foo").Struct()
			return f.GoString()
		},
	},
	{
		name:       "interface",
		comment:    `Foo is an interface that can be implemented.`,
		identifier: "Foo",
		input: func(f *jen.File) string {
			f.Type().Id("Foo").Interface()
			return f.GoString()
		},
	},
	{
		name:       "struct method",
		comment:    `Foo is a method of a struct.`,
		identifier: "X.Foo",
		input: func(f *jen.File) string {
			f.Func().Parens(jen.Id("X")).Id("Foo").Params().Error()
			return f.GoString()
		},
	},
	{
		name:       "pointer method",
		comment:    `Foo is a method of a struct pointer.`,
		identifier: "*X.Foo",
		input: func(f *jen.File) string {
			f.Func().Parens(jen.Id("*X")).Id("Foo").Params().Error()
			return f.GoString()
		},
	},
}

func TestPatch_DryRun(t *testing.T) {
	for _, tt := range dryRunTests {
		t.Run(tt.name, func(t *testing.T) {
			f := jen.NewFile("foo")
			input := tt.input(f)

			f2 := jen.NewFile("foo")
			f2.Comment(tt.comment)
			want := tt.input(f2)

			sourceFS := memfs.New()
			if err := sourceFS.WriteFile("foo.go", []byte(input), fs.ModePerm); err != nil {
				t.Fatal(err)
			}

			p := patch.New(sourceFS)

			if err := p.Comment("foo.go", tt.identifier, tt.comment); err != nil {
				t.Fatal(err)
			}

			result, err := p.DryRun()
			if err != nil {
				t.Fatal(err)
			}

			got := string(result["foo.go"])

			t.Logf("Input:\n%s", input)
			t.Logf("Want:\n%s", want)
			t.Logf("Got:\n%s", got)

			if got != want {
				t.Fatal(cmp.Diff(want, got))
			}
		})
	}
}

func TestPatch_DryRun_realFiles(t *testing.T) {
	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "patch")
	tests.WithRepo("basic", root, func(repoFS fs.FS) {
		p := patch.New(repoFS)

		if err := p.Comment("foo.go", "Foo", "Foo is a function that returns a \"foo\" error."); err != nil {
			t.Fatal(err)
		}

		if err := p.Comment("baz.go", "X.Foo", "Foo is a method of a struct."); err != nil {
			t.Fatal(err)
		}

		if err := p.Apply(root); err != nil {
			t.Fatalf("apply patch: %v", err)
		}

		result, err := p.DryRun()
		if err != nil {
			t.Fatal(err)
		}

		for path, content := range result {
			t.Logf("%s:\n%s", path, content)
		}

		// TODO(bounoable): Implement tests.ExpectComment
		// tests.ExpectComment(t, filepath.Join(root, "foo.go"), "Foo", "Foo is a function that returns a \"foo\" error.")
		// tests.ExpectComment(t, filepath.Join(root, "baz.go"), "*X.Bar", "Bar is a method of a struct pointer.")
	})
}
