package nodes

import (
	"go/parser"
	"go/token"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/modernice/jotbot/internal/slice"
)

// IsExportedIdentifier checks if the given identifier string represents an
// exported identifier (i.e., starts with an uppercase letter) and returns true
// if it does, otherwise false. It also handles identifiers with a package
// prefix or type/method separator.
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

// StripIdentifierPrefix removes the prefix (the part before the colon) from an
// identifier string, if present, and returns the remaining identifier.
func StripIdentifierPrefix(identifier string) string {
	if parts := strings.Split(identifier, ":"); len(parts) > 1 {
		return parts[1]
	}
	return identifier
}

// HasDoc checks if the provided decorations contain any documentation comments.
func HasDoc(decs dst.Decorations) bool {
	return len(decs.All()) > 0
}

// Doc returns the documentation string associated with the given [dst.Node]. If
// removeSlash is true, leading slashes and spaces are removed from the
// documentation lines.
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

// Parse takes a code input and returns a parsed *dst.File along with any error
// encountered. The input code can be either a string or a byte slice. The
// function uses decorator.ParseFile to parse the code with ParseComments and
// SkipObjectResolution flags.
func Parse[Code ~string | ~[]byte](code Code) (*dst.File, error) {
	return decorator.ParseFile(token.NewFileSet(), "", code, parser.ParseComments|parser.SkipObjectResolution)
}

// MustParse parses the provided code into a dst.Node, panicking if an error
// occurs during parsing.
func MustParse[Code ~string | ~[]byte](code Code) dst.Node {
	node, err := Parse(code)
	if err != nil {
		panic(err)
	}
	return node
}

// Identifier returns the identifier string and a boolean indicating whether the
// identifier is exported or not for the given node. The returned identifier is
// in the format "kind:name", where kind can be "func", "type", or "var".
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

// Find searches for a specified identifier within a given root node, returning
// the associated Spec, Node, and a boolean indicating if the identifier was
// found. It supports finding functions, types, and variables.
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

// FindFunc searches for a function with the specified identifier within the
// provided root node and returns the found *dst.FuncDecl and a boolean
// indicating whether the function was found or not.
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

// FindInterfaceMethod searches for a method with the specified identifier in
// interface types within a given root node. It returns the found method as a
// *dst.Field and a boolean indicating whether the method was found or not.
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

// FindValue searches for a value with the given identifier in the provided root
// node, returning the matching ValueSpec, its parent GenDecl, and a boolean
// indicating whether the value was found.
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

// FindType searches for a type declaration with the given identifier in the
// provided root node. It returns the found TypeSpec, the containing GenDecl,
// and a boolean indicating whether the type was found.
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

// CommentTarget determines the appropriate node to attach a comment to, given a
// specification and an outer node. If the spec is nil or if the spec is a
// single TypeSpec or ValueSpec within a GenDecl, the outer node is returned.
// Otherwise, the spec is returned.
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
