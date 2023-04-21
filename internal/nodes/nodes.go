package nodes

import (
	"strings"
	"unicode"

	"github.com/dave/dst"
	"github.com/modernice/jotbot/internal/slice"
)

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

func methodIdentifier(identifier string, recv dst.Expr) string {
	switch recv := recv.(type) {
	case *dst.Ident:
		return recv.Name + "." + identifier
	case *dst.StarExpr:
		var ident *dst.Ident

		switch recv := recv.X.(type) {
		case *dst.Ident:
			ident = recv
		case *dst.IndexExpr:
			ident, _ = recv.X.(*dst.Ident)
		}

		if ident != nil {
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
