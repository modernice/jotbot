package nodes

import (
	"strings"

	"github.com/dave/dst"
	"github.com/modernice/jotbot/internal/slice"
)

func IsExportedIdentifier(identifier string) bool {
	return len(identifier) > 0 && identifier != "_" && strings.ToUpper(identifier[:1]) == identifier[:1]
}

// HasDoc [func] returns a boolean indicating whether a given Go AST node has
// any associated documentation comments. It takes the Decorations of the node
// as input.
func HasDoc(decs dst.Decorations) bool {
	return len(decs.All()) > 0
}

// Doc package provides functions for working with Go code documentation. The
// `HasDoc` function checks whether a node has any associated comments. The
// `Doc` function retrieves the comment text associated with a node. The
// `Identifier` function returns the identifier and export status of a node. The
// `Find` function finds the first node with a matching identifier in the AST
// rooted at the given node. The `FindT` function is a type-safe version of
// `Find`.
func Doc(n dst.Node, removeSlash bool) string {
	lines := n.Decorations().Start.All()
	if removeSlash {
		lines = slice.Map(lines, trimSlash)
	}
	return strings.Join(lines, "")
}

func trimSlash(s string) string {
	return strings.TrimLeft(strings.TrimPrefix(s, "//"), " ")
}

// Identifier function in the nodes package returns the identifier and export
// status of a Go AST node. It takes a dst.Node as input and returns the
// identifier string and a boolean indicating whether the identifier is exported
// or not.
func Identifier(node dst.Node) (identifier string, exported bool) {
	switch node := node.(type) {
	case *dst.FuncDecl:
		identifier = node.Name.Name

		if node.Recv != nil && len(node.Recv.List) > 0 {
			// if identifier == "Foo" {
			// 	log.Printf("%#v", node.Recv.List[0].Type)
			// }
			// if identifier == "Bar" {
			// 	log.Printf("%#v", node.Recv.List[0].Type)
			// }
			// ident, ok := getIdent(node.Recv.List[0].Type)
			identifier, _ = methodIdentifier(identifier, node.Recv.List[0].Type)
		}
	case *dst.GenDecl:
		if len(node.Specs) == 0 {
			break
		}

		spec := node.Specs[0]

		switch spec := spec.(type) {
		case *dst.TypeSpec:
			identifier = spec.Name.Name
		case *dst.ValueSpec:
			identifier = spec.Names[0].Name
		}
	case *dst.TypeSpec:
		identifier = node.Name.Name
	case *dst.ValueSpec:
		identifier = node.Names[0].Name
	}

	return identifier, IsExportedIdentifier(identifier)
}

// Find searches for a Go AST [dst.Node] with the given identifier in the root
// node and its children. If found, it returns the first matching node and true;
// otherwise, it returns nil and false.
func Find(identifier string, root dst.Node) (dst.Node, bool) {
	return FindT[dst.Node](identifier, root)
}

// FindT[Node dst.Node] searches for the first node in the AST rooted at root
// with an identifier that matches the given identifier, and returns it as a
// Node of type Node. If no such node is found, it returns nil and false.
func FindT[Node dst.Node](identifier string, root dst.Node) (Node, bool) {
	var (
		found Node
		ok    bool
	)

	dst.Inspect(root, func(node dst.Node) bool {
		if ident, _ := Identifier(node); ident == identifier {
			found, ok = node.(Node)
			return false
		}
		return true
	})

	return found, ok
}

func methodIdentifier(identifier string, recv dst.Expr) (string, bool) {
	switch recv := recv.(type) {
	case *dst.StarExpr:
		if ident, ok := getIdent(recv.X); ok {
			return "(*" + ident.Name + ")." + identifier, true
		}
	default:
		if ident, ok := getIdent(recv); ok {
			return ident.Name + "." + identifier, true
		}
	}
	return "", false
}

func getIdent(expr dst.Expr) (*dst.Ident, bool) {
	var ident *dst.Ident

	switch e := expr.(type) {
	case *dst.Ident:
		ident = e
	case *dst.IndexListExpr:
		ident, _ = e.X.(*dst.Ident)
	case *dst.IndexExpr:
		ident, _ = e.X.(*dst.Ident)
	}

	if ident == nil {
		return nil, false
	}

	return ident, true
}
