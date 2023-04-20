package reset

import "github.com/dave/dst"

// Comments [func Comments(node dst.Node)] is a function that removes comments 
// from the specified Go Abstract Syntax Tree (AST) node. This function accepts 
// a node of type dst.Node and recursively traverses the AST to remove all 
// comments associated with the specified node. The types of nodes that have 
// their comments removed include *dst.File, *dst.GenDecl, *dst.FuncDecl, 
// *dst.ValueSpec, and *dst.TypeSpec.
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
