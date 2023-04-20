package patch_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/google/go-cmp/cmp"
	"github.com/modernice/jotbot/git"
	"github.com/modernice/jotbot/internal/tests"
	"github.com/modernice/jotbot/patch"
	"github.com/psanford/memfs"
)

var _ interface {
	git.Patch
	git.Committer
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
	{
		name:       "const",
		comment:    `Foo is a constant that is used to do things.`,
		identifier: "Foo",
		input: func(f *jen.File) string {
			f.Const().Id("Foo").Op("=").Lit("foo")
			return f.GoString()
		},
	},
	{
		name:       "var",
		comment:    `Foo is a variable that is used to do things.`,
		identifier: "Foo",
		input: func(f *jen.File) string {
			f.Var().Id("Foo").Op("=").Lit("foo")
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
				t.Log(input)
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

var splitStringInput = strings.TrimSpace(`
So the LORD God said to the serpent, Because you have done this, cursed are you among all animals and among all animals and among all animals and among all animals and among all animals and among all wild creatures; upon your belly you shall go, and dust you shall go, and dust you shall return.

Out of the ground the LORD God commanded the man, You may freely eat of every kind, and everything that has the breath of life, I have given you every plant yielding seed that is upon the face of the first is Pishon; it is the one that flows around the whole face of all the wild animals of the evening breeze, and the gold of that land is good; bdellium and onyx stone are there.

I will greatly increase your pangs in childbearing; in pain you shall not eat, for in the day and the tree of which I commanded you, 'You shall not eat of the field had yet sprung up - for the man there was not found a helper as his partner. For God knows that when you eat of every tree of the tree that is in the middle of the earth of every kind. And it was so.
`)

var splitStringWant = strings.TrimSpace(`
So the LORD God said to the serpent, Because you have done this, cursed are
you among all animals and among all animals and among all animals and among
all animals and among all animals and among all wild creatures; upon your
belly you shall go, and dust you shall go, and dust you shall return.

Out of the ground the LORD God commanded the man, You may freely eat of every
kind, and everything that has the breath of life, I have given you every
plant yielding seed that is upon the face of the first is Pishon; it is the
one that flows around the whole face of all the wild animals of the evening
breeze, and the gold of that land is good; bdellium and onyx stone are there.

I will greatly increase your pangs in childbearing; in pain you shall not
eat, for in the day and the tree of which I commanded you, 'You shall not eat
of the field had yet sprung up - for the man there was not found a helper as
his partner. For God knows that when you eat of every tree of the tree that
is in the middle of the earth of every kind. And it was so.
`)

func TestPatch_Comment_splitString(t *testing.T) {
	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "split-string")
	tests.WithRepo("basic", root, func(repoFS fs.FS) {
		p := patch.New(repoFS)

		if err := p.Comment("foo.go", "Foo", splitStringInput); err != nil {
			t.Fatal(err)
		}

		b, err := p.File("foo.go")
		if err != nil {
			t.Fatal(err)
		}

		tests.ExpectComment(t, "Foo", splitStringWant, strings.NewReader(string(b)))
	})
}

func TestPatch_Comment_duplicateName(t *testing.T) {
	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "duplicate-name")
	tests.WithRepo("duplicate-name", root, func(repoFS fs.FS) {
		p := patch.New(repoFS)

		if err := p.Comment("foo.go", "Foo", "Foo is a function."); err != nil {
			t.Fatal(err)
		}

		if err := p.Comment("foo.go", "X.Foo", "Foo is a method of X."); err != nil {
			t.Fatal(err)
		}

		if err := p.Comment("foo.go", "*Y.Foo", "Foo is a method of *Y."); err != nil {
			t.Fatal(err)
		}

		b, err := p.File("foo.go")
		if err != nil {
			t.Fatal(err)
		}

		tests.ExpectComment(t, "Foo", "Foo is a function.", strings.NewReader(string(b)))
		tests.ExpectComment(t, "X.Foo", "Foo is a method of X.", strings.NewReader(string(b)))
		tests.ExpectComment(t, "*Y.Foo", "Foo is a method of *Y.", strings.NewReader(string(b)))
	})
}
