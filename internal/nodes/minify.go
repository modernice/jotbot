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

// MinifyOptions represents the options for minifying Go AST nodes. It includes
// fields for controlling the minification of package comments, function
// comments, function bodies, and struct comments, as well as whether to minify
// only exported identifiers. MinifyOptions also has a Minify method that takes
// a Go AST node and returns the minified version according to the options
// specified. The Minify function is a convenience wrapper around
// MinifyOptions.Minify that allows specifying the type of the input and output
// nodes.
type MinifyOptions struct {
	PackageComment bool
	FuncComment    bool
	FuncBody       bool
	StructComment  bool
	Exported       bool
}

// MinifyOptions.Minify [Node dst.Node](node Node, opts MinifyOptions) Node
//
// Minify applies the minification options specified in the MinifyOptions struct
// to the provided AST node of type "Node" and returns the resulting node. The
// minification options include removing function bodies, comments, or struct
// comments based on the values of corresponding fields in the MinifyOptions
// struct. If a field is set to true, then that part of the code will be
// minified. The "Exported" field determines whether to apply the minification
// options to exported or unexported identifiers.
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

// Minify is a function that takes a Node and a MinifyOptions as input and
// returns a Node. It removes unnecessary code from the given Node based on the
// options provided in MinifyOptions. MinifyOptions is a struct type that
// specifies which parts of the code to minify. There are four pre-defined
// MinifyOptions: MinifyNone, MinifyUnexported, MinifyExported, MinifyComments,
// and MinifyAll. The options include package comments, function comments,
// function bodies, and struct comments. The function also checks whether
// identifiers are exported or not before applying the patch.
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
