package nodes

import (
	"go/token"

	"github.com/dave/dst"
)

var (
	// MinifyNone is a variable of MinifyOptions type that represents the default,
	// no-op configuration where no minification actions are performed on the input
	// node.
	MinifyNone MinifyOptions

	// MinifyUnexported is a preconfigured set of MinifyOptions that minifies
	// function comments, function bodies, and struct comments for unexported
	// elements only.
	MinifyUnexported = MinifyOptions{
		FuncComment:   true,
		FuncBody:      true,
		StructComment: true,
	}

	// MinifyExported is a variable of MinifyOptions type that configures
	// minification to remove function comments, function bodies, and struct
	// comments for exported elements in the Go source code.
	MinifyExported = MinifyOptions{
		FuncComment:   true,
		FuncBody:      true,
		StructComment: true,
		Exported:      true,
	}

	// MinifyComments is a predefined configuration of MinifyOptions that minifies
	// package, function, and struct comments for both exported and unexported
	// elements.
	MinifyComments = MinifyOptions{
		PackageComment: true,
		FuncComment:    true,
		StructComment:  true,
		Exported:       true,
	}

	// MinifyAll is a variable of MinifyOptions type representing the most
	// aggressive minification configuration, which removes package comments,
	// function comments and bodies, and struct comments for both exported and
	// unexported elements in the input node.
	MinifyAll = MinifyOptions{
		PackageComment: true,
		FuncComment:    true,
		FuncBody:       true,
		StructComment:  true,
		Exported:       true,
	}
)

// MinifyOptions is a configuration struct that sets options for minifying the
// different parts of a Go source code file, such as package comments, function
// comments and bodies, and struct comments. It can also be configured to only
// minify exported or unexported elements.
type MinifyOptions struct {
	PackageComment bool
	FuncComment    bool
	FuncBody       bool
	StructComment  bool
	Exported       bool
}

// Minify applies the specified MinifyOptions to the given dst.Node, removing
// comments, function bodies, and/or struct comments based on the options set.
// If Exported is true, only exported nodes are affected; otherwise, all nodes
// are affected. The modified node is returned.
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

// Minify applies the specified MinifyOptions to the given dst.Node, removing
// elements such as comments and function bodies to reduce its size. The
// resulting node is a clone of the original with the specified changes applied.
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
