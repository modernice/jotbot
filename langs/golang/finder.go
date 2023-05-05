package golang

import (
	"fmt"
	"go/parser"
	"go/token"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/modernice/jotbot/find"
	"github.com/modernice/jotbot/internal/nodes"
	"golang.org/x/exp/slices"
)

type Finder struct {
	findTests bool
}

type FinderOption func(*Finder)

func FindTests(find bool) FinderOption {
	return func(f *Finder) {
		f.findTests = find
	}
}

func NewFinder(opts ...FinderOption) *Finder {
	var f Finder
	for _, opt := range opts {
		opt(&f)
	}
	return &f
}

func (f *Finder) Find(code []byte) ([]find.Finding, error) {
	var findings []find.Finding

	fset := token.NewFileSet()
	node, err := decorator.ParseFile(fset, "", code, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parse code: %w", err)
	}

	for _, node := range node.Decls {
		var (
			identifier string
			exported   bool
		)

		switch node := node.(type) {
		case *dst.FuncDecl:
			if !f.findTests && isTestFunction(node) {
				break
			}

			if nodes.HasDoc(node.Decs.NodeDecs.Start) {
				break
			}

			identifier, exported = nodes.Identifier(node)
		case *dst.GenDecl:
			if nodes.HasDoc(node.Decs.NodeDecs.Start) {
				break
			}

			if len(node.Specs) == 0 {
				break
			}

			spec := node.Specs[0]

			switch spec := spec.(type) {
			case *dst.TypeSpec:
				if !nodes.HasDoc(spec.Decs.NodeDecs.Start) {
					identifier, exported = nodes.Identifier(spec)
				}

				if isInterface(spec) {
					findings = append(findings, findInterfaceMethods(spec)...)
				}
			case *dst.ValueSpec:
				if !nodes.HasDoc(spec.Decs.NodeDecs.Start) {
					identifier, exported = nodes.Identifier(spec)
				}
			}
		}

		if exported {
			findings = append(findings, find.Finding{
				Identifier: identifier,
			})
		}
	}

	slices.SortFunc(findings, func(a, b find.Finding) bool {
		return a.Identifier <= b.Identifier
	})

	return findings, nil
}

func isInterface(spec *dst.TypeSpec) bool {
	_, ok := spec.Type.(*dst.InterfaceType)
	return ok
}

func findInterfaceMethods(spec *dst.TypeSpec) []find.Finding {
	var findings []find.Finding

	ifaceName := spec.Name.Name
	for _, method := range spec.Type.(*dst.InterfaceType).Methods.List {
		name := method.Names[0].Name
		ident := fmt.Sprintf("func:%s.%s", ifaceName, name)
		if nodes.IsExportedIdentifier(ident) && !nodes.HasDoc(method.Decs.Start) {
			findings = append(findings, find.Finding{
				Identifier: ident,
			})
		}
	}

	return findings
}

func isTestFunction(node *dst.FuncDecl) bool {
	return strings.HasPrefix(node.Name.Name, "Test")
}
