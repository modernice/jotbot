package nodes

import (
	"go/token"

	"github.com/dave/dst"
)

// MinifyNone is a predefined variable of type MinifyOptions that represents the
// option to not minify any part of the code. It is one of the four options
// available in the package nodes/internal/nodes/minify.go.
var (
	MinifyNone MinifyOptions

	MinifyUnexported = MinifyOptions{
		FuncComment:   true,
		FuncBody:      true,
		StructComment: true,
	}

	MinifyExported = MinifyOptions{
		FuncComment:   true,
		FuncBody:      true,
		StructComment: true,
		Exported:      true,
	}

	MinifyComments = MinifyOptions{
		PackageComment: true,
		FuncComment:    true,
		StructComment:  true,
		Exported:       true,
	}

	MinifyAll = MinifyOptions{
		PackageComment: true,
		FuncComment:    true,
		FuncBody:       true,
		StructComment:  true,
		Exported:       true,
	}
)

// MinifyOptions represents the options for minifying Go code. It is a struct
// type that contains boolean fields for PackageComment, FuncComment, FuncBody,
// StructComment, and Exported. The Minify method of MinifyOptions receives a
// dst.Node and returns a minified dst.Node according to the options specified.
// The function Minify[Node dst.Node] takes a Node and MinifyOptions as
// arguments and returns a minified Node. MinifyNone, MinifyUnexported,
// MinifyExported, MinifyComments, and MinifyAll are predefined MinifyOptions
// variables in the package nodes/internal/nodes/minify.go.
type MinifyOptions struct {
	PackageComment bool
	FuncComment    bool
	FuncBody       bool
	StructComment  bool
	Exported       bool
}

// Minify[Node dst.Node] minifies a given AST (abstract syntax tree) node
// according to the specified MinifyOptions. It returns the resulting AST node
// of the same type as the input. The MinifyOptions control which parts of the
// code are minified, including package comments, function comments, function
// bodies, and struct comments. The MinifyAll, MinifyExported, MinifyUnexported,
// and MinifyNone variables in this package are predefined options for common
// use cases.
func (opts MinifyOptions) Minify(node dst.Node) dst.Node {
	out := dst.Clone(node)

	patch := func(node dst.Node) {
		switch node := node.(type) {
		case *dst.FuncDecl:
			if opts.FuncBody {
				node.Body = nil
			}

			if opts.FuncComment {
				node.Decs.Start.Clear()
			}
		case *dst.GenDecl:
			if opts.StructComment && isStruct(node) {
				node.Decs.Start.Clear()
			}
		}
	}

	dst.Inspect(out, func(node dst.Node) bool {
		if _, exported := Identifier(node); exported && opts.Exported {
			patch(node)
		} else if !exported {
			patch(node)
		}
		return true
	})

	if opts.PackageComment {
		if file, ok := out.(*dst.File); ok {
			file.Decs.Package.Clear()
			file.Decs.Start.Clear()
		}
	}

	return out
}

// Minify is a function that takes a Node and MinifyOptions as input, and
// returns a minified version of the Node according to the specified options.
// The MinifyOptions struct specifies which parts of the code to minify,
// including package comments, function comments, function bodies, and struct
// comments. The MinifyNone variable represents an option to not minify any part
// of the code, and there are three other predefined MinifyOptions variables
// with different levels of minification.
func Minify[Node dst.Node](node Node, opts MinifyOptions) Node {
	return opts.Minify(node).(Node)
}

func isStruct(decl *dst.GenDecl) bool {
	if decl.Tok != token.TYPE {
		return false
	}

	if len(decl.Specs) < 1 {
		return false
	}

	_, ok := decl.Specs[0].(*dst.TypeSpec)

	return ok
}
