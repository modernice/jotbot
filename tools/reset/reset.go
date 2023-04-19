package reset

import "github.com/dave/dst"

// Comments function removes all comments from the source code represented by 
// the given dst.Node. It traverses the AST and clears comments associated with 
// the package declaration, declarations, functions, variables, and types.
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
