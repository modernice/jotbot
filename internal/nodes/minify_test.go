package nodes_test

import (
	"testing"

	"github.com/dave/dst/decorator"
	"github.com/modernice/opendocs/internal/nodes"
)

var minifyInput = `// Package foo is super nice.
package foo

// Foo is a function.
func Foo() error {
	return foo()
}

type X struct{}

// Bar is a method.
func (X) Bar() error {
	return bar()
}

// foo is a function.
func foo() error {
	return bar()
}

// bar is a function, too.
func bar() error {
	return errors.New("bar")
}
`

var wantUnexportedFuncBody = `// Package foo is super nice.
package foo

// Foo is a function.
func Foo() error {
	return foo()
}

type X struct{}

// Bar is a method.
func (X) Bar() error {
	return bar()
}

// foo is a function.
func foo() error

// bar is a function, too.
func bar() error
`

var wantUnexportedFuncComment = `// Package foo is super nice.
package foo

// Foo is a function.
func Foo() error {
	return foo()
}

type X struct{}

// Bar is a method.
func (X) Bar() error {
	return bar()
}

func foo() error {
	return bar()
}

func bar() error {
	return errors.New("bar")
}
`

var wantUnexportedFunc = `// Package foo is super nice.
package foo

// Foo is a function.
func Foo() error {
	return foo()
}

type X struct{}

// Bar is a method.
func (X) Bar() error {
	return bar()
}

func foo() error

func bar() error
`

var wantExportedFuncBody = `// Package foo is super nice.
package foo

// Foo is a function.
func Foo() error

type X struct{}

// Bar is a method.
func (X) Bar() error

// foo is a function.
func foo() error {
	return bar()
}

// bar is a function, too.
func bar() error {
	return errors.New("bar")
}
`

var wantExportedFuncComment = `// Package foo is super nice.
package foo

func Foo() error {
	return foo()
}

type X struct{}

func (X) Bar() error {
	return bar()
}

// foo is a function.
func foo() error {
	return bar()
}

// bar is a function, too.
func bar() error {
	return errors.New("bar")
}
`

var wantExportedFuncAll = `// Package foo is super nice.
package foo

func Foo() error

type X struct{}

func (X) Bar() error

// foo is a function.
func foo() error {
	return bar()
}

// bar is a function, too.
func bar() error {
	return errors.New("bar")
}
`

var wantPackageComment = `package foo

// Foo is a function.
func Foo() error {
	return foo()
}

type X struct{}

// Bar is a method.
func (X) Bar() error {
	return bar()
}

// foo is a function.
func foo() error {
	return bar()
}

// bar is a function, too.
func bar() error {
	return errors.New("bar")
}
`

func TestMinify(t *testing.T) {
	tests := []struct {
		name string
		opts nodes.MinifyOptions
		want string
	}{
		{
			name: "Unexported.Body",
			opts: nodes.MinifyOptions{
				Unexported: nodes.MinifyFuncBody,
			},
			want: wantUnexportedFuncBody,
		},
		{
			name: "Unexported.Comment",
			opts: nodes.MinifyOptions{
				Unexported: nodes.MinifyFuncComment,
			},
			want: wantUnexportedFuncComment,
		},
		{
			name: "Unexported.All",
			opts: nodes.MinifyOptions{
				Unexported: nodes.MinifyFuncComment | nodes.MinifyFuncBody,
			},
			want: wantUnexportedFunc,
		},
		{
			name: "Exported.Body",
			opts: nodes.MinifyOptions{
				Exported: nodes.MinifyFuncBody,
			},
			want: wantExportedFuncBody,
		},
		{
			name: "Exported.Comment",
			opts: nodes.MinifyOptions{
				Exported: nodes.MinifyFuncComment,
			},
			want: wantExportedFuncComment,
		},
		{
			name: "Exported.All",
			opts: nodes.MinifyOptions{
				Exported: nodes.MinifyFuncComment | nodes.MinifyFuncBody,
			},
			want: wantExportedFuncAll,
		},
		{
			name: "PackageComment",
			opts: nodes.MinifyOptions{
				PackageComment: true,
			},
			want: wantPackageComment,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := decorator.Parse([]byte(minifyInput))
			if err != nil {
				t.Fatal(err)
			}

			minified := nodes.Minify(node, tt.opts)

			code, err := nodes.Format(minified)
			if err != nil {
				t.Fatal(err)
			}

			if string(code) != tt.want {
				t.Fatalf("unexpected minified code\n\nwant:\n%s\n\ngot:\n%s", tt.want, string(code))
			}
		})
	}
}
