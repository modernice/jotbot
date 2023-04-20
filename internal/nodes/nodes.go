package nodes

import (
	"strings"
	"unicode"

	"github.com/dave/dst"
	"github.com/modernice/jotbot/internal/slice"
)

// HasDoc checks if a Go syntax tree node has any associated comments in its
// Decorations [Decorations]. It returns true if the node has comments and false
// otherwise.
func HasDoc(decs dst.Decorations) bool {
	return len(decs.All()) > 0
}

// Doc package provides utility functions for working with Go AST nodes'
// documentation. It includes functions for detecting if a node has any
// associated documentation, retrieving the documentation for a node, and
// finding a node by its identifier within an AST.
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

// Identifier provides a way to extract the identifier and exported status of a
// given Go AST node. The function returns the node's identifier as a string and
// a boolean indicating if the identifier is exported (starts with an uppercase
// letter). The input node can be any of the following types: *dst.FuncDecl,
// *dst.GenDecl, *dst.TypeSpec, or *dst.ValueSpec. If the input node is a
// function with a receiver, the returned identifier will include the receiver's
// type name.
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

// Find searches a given [identifier] in the AST rooted at [root] and returns
// the first node that has a matching identifier along with a boolean indicating
// whether it was found or not.
func Find(identifier string, root dst.Node) (dst.Node, bool) {
	return FindT[dst.Node](identifier, root)
}

// FindT searches a DST (Go AST) tree for a node of type Node with the given
// identifier. It returns the found node and a boolean indicating whether it was
// found or not.
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
