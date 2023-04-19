package nodes

import (
	"strings"
	"unicode"

	"github.com/dave/dst"
	"github.com/modernice/opendocs/internal/slice"
)

// HasDoc determines whether a Go AST node has any associated documentation
// comments. It takes a Decorations object as input and returns a boolean value
// indicating whether there are any comments present.
func HasDoc(decs dst.Decorations) bool {
	return len(decs.All()) > 0
}

// Doc package provides functions for generating GoDoc documentation for Go
// source code. It includes functions for checking if a node has documentation,
// retrieving documentation for a node, finding a node by identifier, and
// generating an identifier for a method.
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

// Identifier function returns the identifier and a boolean indicating whether
// the identifier is exported or not for a given Go AST node. It supports
// *dst.FuncDecl, *dst.GenDecl, *dst.TypeSpec, and *dst.ValueSpec nodes. If the
// node is a function with a receiver, it returns the method identifier.
func Identifier(node dst.Node) (identifier string, exported bool) {
	switch node := node.(type) {
	case *dst.FuncDecl:
		identifier = node.Name.Name
		exported = isExported(identifier)

		if node.Recv != nil && len(node.Recv.List) > 0 {
			identifier = methodIdentifier(identifier, node.Recv.List[0].Type)
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

		exported = isExported(identifier)
	case *dst.TypeSpec:
		identifier = node.Name.Name
		exported = isExported(identifier)
	case *dst.ValueSpec:
		identifier = node.Names[0].Name
		exported = isExported(identifier)
	}

	return
}

// Find searches for a Go AST node with the given identifier in the provided
// root node. It returns the first node found and a boolean indicating whether
// the node was found. The identifier can be the name of a function, type, or
// variable.
func Find(identifier string, root dst.Node) (dst.Node, bool) {
	return FindT[dst.Node](identifier, root)
}

// FindT searches for a node of type Node with the given identifier in the AST
// rooted at root. It returns the first node found and a boolean indicating
// whether the search was successful. The type of the node to search for is
// specified by the type parameter, which must be a dst.Node type wrapped within
// brackets.
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

func methodIdentifier(identifier string, recv dst.Expr) string {
	switch recv := recv.(type) {
	case *dst.Ident:
		return recv.Name + "." + identifier
	case *dst.StarExpr:
		if ident, ok := recv.X.(*dst.Ident); ok {
			return "*" + ident.Name + "." + identifier
		}
	}
	return identifier
}

func isExported(identifier string) bool {
	if len(identifier) == 0 {
		return false
	}
	runes := []rune(identifier)
	unexported := len(identifier) > 0 &&
		unicode.IsLetter(runes[0]) &&
		strings.ToLower(identifier[:1]) == identifier[:1]
	return !unexported
}
