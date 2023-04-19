package nodes_test

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/dave/dst/decorator"
	"github.com/modernice/opendocs/internal/nodes"
	"github.com/modernice/opendocs/internal/tests"
)

var wantMinified = `package fixture

import "errors"

type Foo struct{}

// Foo is a method.
func (f *Foo) Foo() error {
	return f.foo()
}

func (f *Foo) foo() error

func (f *Foo) bar() error
`

func TestMinify(t *testing.T) {
	root := filepath.Join(tests.Must(os.Getwd()), "testdata", "gen", "minify")
	tests.WithRepo("minify", root, func(repo fs.FS) {
		f, err := repo.Open("foo.go")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		node, err := decorator.Parse(f)
		if err != nil {
			t.Fatal(err)
		}

		node = nodes.Minify(node)

		var buf bytes.Buffer
		if err := decorator.Fprint(&buf, node); err != nil {
			t.Fatal(err)
		}

		got := buf.String()

		if got != wantMinified {
			t.Fatalf("unpected minified code\n\nwant:\n%s\n\ngot:\n%s", wantMinified, got)
		}
	})
}
