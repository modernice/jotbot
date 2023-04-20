package nodes_test

import (
	"testing"

	"github.com/dave/dst/decorator"
	"github.com/modernice/jotbot/internal/nodes"
)

var minifyInput = `// Package foo is super nice.
package foo

// Foo is a function.
func Foo() error {
	return foo()
}

// X is a struct.
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

// y is a struct.
type y struct{}
`

var wantUnexportedFuncBody = `// Package foo is super nice.
package foo

// Foo is a function.
func Foo() error {
	return foo()
}

// X is a struct.
type X struct{}

// Bar is a method.
func (X) Bar() error {
	return bar()
}

// foo is a function.
func foo() error

// bar is a function, too.
func bar() error

// y is a struct.
type y struct{}
`

var wantUnexportedFuncComment = `// Package foo is super nice.
package foo

// Foo is a function.
func Foo() error {
	return foo()
}

// X is a struct.
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

// y is a struct.
type y struct{}
`

var wantUnexportedFunc = `// Package foo is super nice.
package foo

// Foo is a function.
func Foo() error {
	return foo()
}

// X is a struct.
type X struct{}

// Bar is a method.
func (X) Bar() error {
	return bar()
}

func foo() error

func bar() error

// y is a struct.
type y struct{}
`

var wantUnexportedStructComment = `// Package foo is super nice.
package foo

// Foo is a function.
func Foo() error {
	return foo()
}

// X is a struct.
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

type y struct{}
`

var wantExportedFuncBody = `// Package foo is super nice.
package foo

// Foo is a function.
func Foo() error

// X is a struct.
type X struct{}

// Bar is a method.
func (X) Bar() error

// foo is a function.
func foo() error

// bar is a function, too.
func bar() error

// y is a struct.
type y struct{}
`

var wantExportedFuncComment = `// Package foo is super nice.
package foo

func Foo() error {
	return foo()
}

// X is a struct.
type X struct{}

func (X) Bar() error {
	return bar()
}

func foo() error {
	return bar()
}

func bar() error {
	return errors.New("bar")
}

// y is a struct.
type y struct{}
`

var wantExportedFunc = `// Package foo is super nice.
package foo

func Foo() error

// X is a struct.
type X struct{}

func (X) Bar() error

func foo() error

func bar() error

// y is a struct.
type y struct{}
`

var wantExportedStructComment = `// Package foo is super nice.
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

type y struct{}
`

var wantPackageComment = `package foo

// Foo is a function.
func Foo() error {
	return foo()
}

// X is a struct.
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

// y is a struct.
type y struct{}
`

var wantAll = `package foo

func Foo() error

type X struct{}

func (X) Bar() error

func foo() error

func bar() error

type y struct{}
`

func TestMinify(t *testing.T) {
	tests := []struct {
		name string
		opts nodes.MinifyOptions
		want string
	}{
		{
			name: "Unexported.FuncBody",
			opts: nodes.MinifyOptions{
				FuncBody: true,
			},
			want: wantUnexportedFuncBody,
		},
		{
			name: "Unexported.FuncComment",
			opts: nodes.MinifyOptions{
				FuncComment: true,
			},
			want: wantUnexportedFuncComment,
		},
		{
			name: "Unexported.Func",
			opts: nodes.MinifyOptions{
				FuncComment: true,
				FuncBody:    true,
			},
			want: wantUnexportedFunc,
		},
		{
			name: "Unexported.StructComment",
			opts: nodes.MinifyOptions{
				StructComment: true,
			},
			want: wantUnexportedStructComment,
		},
		{
			name: "Exported.FuncBody",
			opts: nodes.MinifyOptions{
				FuncBody: true,
				Exported: true,
			},
			want: wantExportedFuncBody,
		},
		{
			name: "Exported.FuncComment",
			opts: nodes.MinifyOptions{
				FuncComment: true,
				Exported:    true,
			},
			want: wantExportedFuncComment,
		},
		{
			name: "Exported.Func",
			opts: nodes.MinifyOptions{
				FuncComment: true,
				FuncBody:    true,
				Exported:    true,
			},
			want: wantExportedFunc,
		},
		{
			name: "Exported.StructComment",
			opts: nodes.MinifyOptions{
				StructComment: true,
				Exported:      true,
			},
			want: wantExportedStructComment,
		},
		{
			name: "PackageComment",
			opts: nodes.MinifyOptions{
				PackageComment: true,
			},
			want: wantPackageComment,
		},
		{
			name: "All",
			opts: nodes.MinifyOptions{
				PackageComment: true,
				FuncComment:    true,
				FuncBody:       true,
				StructComment:  true,
				Exported:       true,
			},
			want: wantAll,
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
