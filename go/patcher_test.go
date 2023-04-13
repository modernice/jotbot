package opendocs_test

import (
	"strings"
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/google/go-cmp/cmp"
	opendocs "github.com/modernice/opendocs/go"
)

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
}

func TestPatcher(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := jen.NewFile("foo")
			input := tt.input(f)

			f2 := jen.NewFile("foo")
			f2.Comment(tt.comment)
			want := tt.input(f2)

			p, err := opendocs.NewPatcher(strings.NewReader(input))
			if err != nil {
				t.Fatal(err)
			}

			if err := p.Update("Foo", tt.comment); err != nil {
				t.Fatal(err)
			}

			b, err := p.Result()
			if err != nil {
				t.Fatal(err)
			}

			result := string(b)

			t.Logf("Input:\n%s", input)
			t.Logf("Want:\n%s", want)
			t.Logf("Got:\n%s", result)

			if result != want {
				t.Fatal(cmp.Diff(want, result))
			}
		})
	}
}
