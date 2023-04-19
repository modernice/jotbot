package nodes

import (
	"go/token"

	"github.com/dave/dst"
)

// MinifyNone is a variable of type MinifyOptions that represents the option to
// not minify any part of the code. It is one of the four predefined
// MinifyOptions variables in the package nodes/internal/nodes/minify.go.
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

// MinifyOptions is a struct that contains boolean fields that determine which
// parts of a given dst.Node should be minified. The fields include
// PackageComment, FuncComment, FuncBody, StructComment, and Exported. The
// Minify method takes a dst.Node and returns a minified version of it based on
// the options specified in the MinifyOptions struct. The Minify function takes
// a dst.Node and a MinifyOptions struct and returns the minified version of the
// node.
type MinifyOptions struct {
	PackageComment bool
	FuncComment    bool
	FuncBody       bool
	StructComment  bool
	Exported       bool
}

// MinifyOptions.Minify is a method that takes a dst.Node and returns a minified
// version of it based on the MinifyOptions struct. The MinifyOptions struct
// contains boolean fields that determine which parts of the node should be
// minified. If the Exported field is true, only exported identifiers will be
// minified. If the PackageComment field is true, the package comment will be
// removed. If the FuncComment field is true, function comments will be removed.
// If the FuncBody field is true, function bodies will be removed. If the
// StructComment field is true, struct comments will be removed.
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

// Minify is a function that takes a dst.Node and MinifyOptions as input, and
// returns a dst.Node. It removes unnecessary comments and function bodies from
// the input node based on the options specified in MinifyOptions. The
// MinifyOptions struct contains boolean fields that determine which parts of
// the node to minify. The Minify function uses the MinifyOptions to patch the
// input node and return the minified version.
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
