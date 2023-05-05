package nodes

import (
	"go/parser"
	"go/token"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/modernice/jotbot/internal/slice"
)

func IsExportedIdentifier(identifier string) bool {
	if parts := strings.Split(identifier, ":"); len(parts) > 1 {
		identifier = parts[1]
	}
	if parts := strings.Split(identifier, "."); len(parts) > 1 {
		identifier = parts[1]
	}
	return len(identifier) > 0 && identifier != "_" &&
		strings.ToUpper(identifier[:1]) == identifier[:1]
}

func StripIdentifierPrefix(identifier string) string {
	if parts := strings.Split(identifier, ":"); len(parts) > 1 {
		return parts[1]
	}
	return identifier
}

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

func Parse[Code ~string | ~[]byte](code Code) (*dst.File, error) {
	return decorator.ParseFile(token.NewFileSet(), "", code, parser.ParseComments|parser.SkipObjectResolution)
}

func MustParse[Code ~string | ~[]byte](code Code) dst.Node {
	node, err := Parse(code)
	if err != nil {
		panic(err)
	}
	return node
}

func Identifier(node dst.Node) (identifier string, exported bool) {
	switch node := node.(type) {
	case *dst.FuncDecl:
		identifier = node.Name.Name

		if node.Recv != nil && len(node.Recv.List) > 0 {
			identifier, _ = methodIdentifier(identifier, node.Recv.List[0].Type)
		}

		identifier = "func:" + identifier
	case *dst.GenDecl:
		if len(node.Specs) == 0 {
			break
		}

		spec := node.Specs[0]

		switch spec := spec.(type) {
		case *dst.TypeSpec:
			identifier = "type:" + spec.Name.Name
		case *dst.ValueSpec:
			identifier = "var:" + spec.Names[0].Name
		}
	case *dst.TypeSpec:
		identifier = "type:" + node.Name.Name
	case *dst.ValueSpec:
		identifier = "var:" + node.Names[0].Name
	}

	return identifier, IsExportedIdentifier(identifier)
}

func Find(identifier string, root dst.Node) (dst.Spec, dst.Node, bool) {
	parts := strings.Split(identifier, ":")
	if len(parts) != 2 {
		return nil, nil, false
	}

	switch parts[0] {
	case "func":
		if decl, ok := FindFunc(identifier, root); ok {
			return nil, decl, ok
		}
		if decl, ok := FindInterfaceMethod(identifier, root); ok {
			return nil, decl, ok
		}
		return nil, nil, false
	case "type":
		return FindType(identifier, root)
	case "var":
		return FindValue(identifier, root)
	default:
		return nil, nil, false
	}
}

func FindFunc(identifier string, root dst.Node) (fn *dst.FuncDecl, found bool) {
	dst.Inspect(root, func(node dst.Node) bool {
		switch node := node.(type) {
		case *dst.FuncDecl:
			if ident, _ := Identifier(node); ident == identifier {
				fn = node
				found = true
				return false
			}
		}
		return true
	})
	return
}

func FindInterfaceMethod(identifier string, root dst.Node) (method *dst.Field, found bool) {
	parts := strings.Split(identifier, ":")
	if len(parts) == 2 {
		identifier = parts[1]
	}

	parts = strings.Split(identifier, ".")
	if len(parts) != 2 {
		return nil, false
	}

	owner := parts[0]
	met := parts[1]

	dst.Inspect(root, func(node dst.Node) bool {
		switch node := node.(type) {
		case *dst.TypeSpec:
			if node.Name.Name != owner {
				break
			}

			if iface, ok := node.Type.(*dst.InterfaceType); ok {
				for _, field := range iface.Methods.List {
					if len(field.Names) == 0 {
						continue
					}

					if field.Names[0].Name == met {
						method = field
						found = true
						return false
					}
				}
			}
		}
		return true
	})
	return
}

func FindValue(identifier string, root dst.Node) (spec *dst.ValueSpec, decl *dst.GenDecl, found bool) {
	dst.Inspect(root, func(node dst.Node) bool {
		switch node := node.(type) {
		case *dst.GenDecl:
			if len(node.Specs) == 0 {
				break
			}

			for _, s := range node.Specs {
				switch s := s.(type) {
				case *dst.ValueSpec:
					if ident, _ := Identifier(s); ident == identifier {
						spec = s
						decl = node
						found = true
						return false
					}
				}
			}
		}
		return true
	})
	return
}

func FindType(identifier string, root dst.Node) (spec *dst.TypeSpec, decl *dst.GenDecl, found bool) {
	dst.Inspect(root, func(node dst.Node) bool {
		switch node := node.(type) {
		case *dst.GenDecl:
			for _, s := range node.Specs {
				switch s := s.(type) {
				case *dst.TypeSpec:
					if ident, _ := Identifier(s); ident == identifier {
						spec = s
						decl = node
						found = true
						return false
					}
				}
			}
		}
		return true
	})
	return
}

func CommentTarget(spec dst.Spec, outer dst.Node) dst.Node {
	if spec == nil {
		return outer
	}

	switch spec := spec.(type) {
	case *dst.TypeSpec:
		if decl, ok := outer.(*dst.GenDecl); ok && len(decl.Specs) == 1 {
			return decl
		}
		return spec
	case *dst.ValueSpec:
		if decl, ok := outer.(*dst.GenDecl); ok && len(decl.Specs) == 1 {
			return decl
		}
		return spec
	}

	return outer
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
