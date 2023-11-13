package nodes

import (
	"go/token"

	"github.com/dave/dst"
)

var (
	// MinifyNone represents the default state where no minification options are
	// applied.
	MinifyNone MinifyOptions

	// MinifyUnexported specifies options to minify unexported function bodies and
	// comments as well as structure comments.
	MinifyUnexported = MinifyOptions{
		FuncComment:   true,
		FuncBody:      true,
		StructComment: true,
	}

	// MinifyExported applies minification settings to exported Go code elements,
	// including comments and function bodies, while preserving the visibility of
	// these elements.
	MinifyExported = MinifyOptions{
		FuncComment:   true,
		FuncBody:      true,
		StructComment: true,
		Exported:      true,
	}

	// MinifyComments specifies options to remove comments from package
	// declarations, function declarations, and struct declarations for exported
	// identifiers.
	MinifyComments = MinifyOptions{
		PackageComment: true,
		FuncComment:    true,
		StructComment:  true,
		Exported:       true,
	}

	// MinifyAll represents the most aggressive reduction of a DST, stripping all
	// comments and function bodies, and applying these changes to both unexported
	// and exported entities.
	MinifyAll = MinifyOptions{
		PackageComment: true,
		FuncComment:    true,
		FuncBody:       true,
		StructComment:  true,
		Exported:       true,
	}
)

// MinifyOptions represents a set of configurable behaviors to control the
// minification process of Go source code represented by the [dst.Node]
// structure. It allows specification of which parts of the code to minify, such
// as comments and function bodies, and whether to restrict minification to
// exported identifiers only. The zero value for MinifyOptions disables all
// minification features.
type MinifyOptions struct {
	PackageComment bool
	FuncComment    bool
	FuncBody       bool
	StructComment  bool
	Exported       bool
}

// Minify applies the specified minification options to the given syntax tree
// node and returns a modified version of that node. The minification process
// may remove comments, function bodies, or alter the visibility of declarations
// based on the settings in the provided MinifyOptions. The function operates
// recursively on all applicable child nodes of the input node. It returns a new
// node with the modifications applied, while leaving the original node
// unchanged.
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

// Minify applies transformations to a given DST node based on the provided
// MinifyOptions, stripping parts like function bodies, comments, and optionally
// affecting exported identifiers only. It returns the transformed DST node.
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
