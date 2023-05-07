package reset

import "github.com/dave/dst"

// Comments removes all comments from the specified [dst.Node] and its
// descendants in the abstract syntax tree, including package, file, and
// declaration-level comments.
func Comments(node dst.Node) {
	dst.Inspect(node, func(node dst.Node) bool {
		switch node := node.(type) {
		case *dst.File:
			node.Decs.Package.Clear()
			node.Decs.Start.Clear()
		case *dst.GenDecl:
			node.Decs.Start.Clear()
		case *dst.FuncDecl:
			node.Decs.Start.Clear()
		case *dst.ValueSpec:
			node.Decs.Start.Clear()
		case *dst.TypeSpec:
			node.Decs.Start.Clear()
		}
		return true
	})
}
