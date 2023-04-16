package patch_test

import (
	"io/fs"
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/google/go-cmp/cmp"
	"github.com/modernice/opendocs/go/git"
	"github.com/modernice/opendocs/go/patch"
	"github.com/psanford/memfs"
)

var _ interface {
	git.Patch
	git.IdentifierProvider
} = (*patch.Patch)(nil)

var tests = []struct {
	name    string
	comment string
	input   func(f *jen.File) string
}{
	{
		name:    "function",
		comment: `Foo is a function that returns a "foo" error.`,
		input: func(f *jen.File) string {
			f.Func().Id("Foo").Params().Error()
			return f.GoString()
		},
	},
	{
		name:    "struct",
		comment: `Foo is a struct that does things.`,
		input: func(f *jen.File) string {
			f.Type().Id("Foo").Struct()
			return f.GoString()
		},
	},
	{
		name:    "interface",
		comment: `Foo is an interface that can be implemented.`,
		input: func(f *jen.File) string {
			f.Type().Id("Foo").Interface()
			return f.GoString()
		},
	},
	{
		name:    "struct method",
		comment: `Foo is a method of a struct.`,
		input: func(f *jen.File) string {
			f.Func().Parens(jen.Id("X")).Id("Foo").Params().Error()
			return f.GoString()
		},
	},
	{
		name:    "pointer method",
		comment: `Foo is a method of a struct pointer.`,
		input: func(f *jen.File) string {
			f.Func().Parens(jen.Id("*X")).Id("Foo").Params().Error()
			return f.GoString()
		},
	},
}

func TestPatch_DryRun(t *testing.T) {
	for _, tt := range tests {
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

			if err := p.Comment("foo.go", "Foo", tt.comment); err != nil {
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
