package nodes

import (
	"go/parser"
	"go/token"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/modernice/jotbot/internal/slice"
)

// IsExportedIdentifier determines whether the given identifier is exported in
// Go. An identifier is considered exported if it begins with an uppercase
// letter and is not the blank identifier "_". This function can be used to
// check if an identifier would be accessible from other packages based on Go's
// visibility rules.
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

// StripIdentifierPrefix removes any prefix before a colon in the given
// identifier and returns the resulting string without the prefix. If no colon
// is present, it returns the original identifier unchanged.
func StripIdentifierPrefix(identifier string) string {
	if parts := strings.Split(identifier, ":"); len(parts) > 1 {
		return parts[1]
	}
	return identifier
}

// HasDoc reports whether the provided decorations contain any documentation
// comments.
func HasDoc(decs dst.Decorations) bool {
	return len(decs.All()) > 0
}

// Doc extracts the leading comment from the specified node, concatenating all
// lines of the comment into a single string. If removeSlash is true, the "//"
// prefix is removed from each line of the comment before concatenation.
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

// Parse reads the given source code and constructs an abstract syntax tree for
// that code. It accepts a string or a byte slice as input and returns a pointer
// to a [*dst.File] representing the parsed code, along with any error
// encountered during parsing. The function also considers comments in the
// source code as part of the parsing process.
func Parse[Code ~string | ~[]byte](code Code) (*dst.File, error) {
	return decorator.ParseFile(token.NewFileSet(), "", code, parser.ParseComments|parser.SkipObjectResolution)
}

// MustParse parses the given code and returns the corresponding DST node. It
// panics if an error occurs during parsing. This convenience function is
// intended for use when the code is guaranteed to be valid and any parse error
// would be considered exceptional. The function accepts a type parameter that
// can be a string or a byte slice, representing the source code to parse. The
// returned DST node can then be used for further manipulation or analysis of
// the parsed code structure.
func MustParse[Code ~string | ~[]byte](code Code) dst.Node {
	node, err := Parse(code)
	if err != nil {
		panic(err)
	}
	return node
}

// Identifier extracts the name and determines the export status of the given
// node. It returns an identifier string with a prefix indicating the kind of
// node, such as "func:", "type:", or "var:", along with a boolean indicating
// whether the identifier is exported.
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

// Find locates a declaration within the abstract syntax tree of Go code given
// an identifier and the root node. It returns the associated specification, the
// parent node if applicable, and a boolean indicating whether the declaration
// was found. The identifier is expected to be prefixed with a domain specifying
// the type of declaration to search for, such as "func", "type", or "var". This
// function does not traverse nested scopes and is limited to top-level
// declarations or methods on top-level types.
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

// FindFunc locates a function declaration within the given root node of the
// abstract syntax tree that matches the specified identifier. It returns the
// function declaration if found and a boolean indicating whether the search was
// successful.
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

// FindInterfaceMethod locates a method of a specified interface within the
// given abstract syntax tree node. It returns the method declaration as a
// [*dst.Field] and a boolean indicating whether the method was found. The
// identifier used to specify the method should be in the format
// "interfaceName.methodName". If the method is not found, the returned
// [*dst.Field] will be nil and the boolean will be false.
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

// FindValue locates a variable declaration within the abstract syntax tree
// rooted at the specified node, matching the provided identifier. It returns
// the corresponding value specification, the enclosing general declaration if
// present, and a boolean indicating whether the variable was found.
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

// FindType locates a type declaration within the abstract syntax tree of a Go
// source file, given its identifier and the root node of the tree. It returns
// the corresponding type specification, the enclosing general declaration if
// applicable, and a boolean indicating whether the type was found.
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

// CommentTarget determines the appropriate node for attaching a comment within
// a Go abstract syntax tree. It resolves between a specification and its outer
// declaration node to find where a comment should be associated, typically
// returning the node that represents the closest syntactic construct to which
// the comment applies. If no specific association is found, it defaults to
// using the provided outer node.
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
