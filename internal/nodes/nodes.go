package nodes

import (
	"strings"
	"unicode"

	"github.com/dave/dst"
	"github.com/modernice/opendocs/internal/slice"
)

func HasDoc(decs dst.Decorations) bool {
	return len(decs.All()) > 0
}

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

func Identifier(node dst.Node) string {
	var identifier string

	switch node := node.(type) {
	case *dst.FuncDecl:
		identifier = node.Name.Name

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
	case *dst.TypeSpec:
		identifier = node.Name.Name
	case *dst.ValueSpec:
		identifier = node.Names[0].Name
	}

	if identifier == "_" || isUnexported(identifier) {
		identifier = ""
	}

	return identifier
}

func Find(identifier string, root dst.Node) (dst.Node, bool) {
	return FindT[dst.Node](identifier, root)
}

func FindT[Node dst.Node](identifier string, root dst.Node) (Node, bool) {
	var (
		found Node
		ok    bool
	)

	dst.Inspect(root, func(node dst.Node) bool {
		if Identifier(node) == identifier {
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

func isUnexported(identifier string) bool {
	runes := []rune(identifier)
	return len(identifier) > 0 && unicode.IsLetter(runes[0]) && strings.ToLower(identifier[:1]) == identifier[:1]
}
