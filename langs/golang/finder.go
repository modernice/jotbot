package golang

import (
	"fmt"
	"go/parser"
	"go/token"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/modernice/jotbot/internal/nodes"
	"golang.org/x/exp/slices"
)

type Finder struct {
	findTests         bool
	includeDocumented bool
}

type FinderOption func(*Finder)

func FindTests(find bool) FinderOption {
	return func(f *Finder) {
		f.findTests = find
	}
}

func IncludeDocumented(include bool) FinderOption {
	return func(f *Finder) {
		f.includeDocumented = include
	}
}

func NewFinder(opts ...FinderOption) *Finder {
	var f Finder
	for _, opt := range opts {
		opt(&f)
	}
	return &f
}

func (f *Finder) Find(code []byte) ([]string, error) {
	var findings []string

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

			if !f.includeDocumented && nodes.HasDoc(node.Decs.NodeDecs.Start) {
				break
			}

			identifier, exported = nodes.Identifier(node)
		case *dst.GenDecl:
			if !f.includeDocumented && nodes.HasDoc(node.Decs.NodeDecs.Start) {
				break
			}

			if len(node.Specs) == 0 {
				break
			}

			spec := node.Specs[0]

			switch spec := spec.(type) {
			case *dst.TypeSpec:
				if f.includeDocumented || !nodes.HasDoc(spec.Decs.NodeDecs.Start) {
					identifier, exported = nodes.Identifier(spec)
				}

				if isInterface(spec) {
					findings = append(findings, f.findInterfaceMethods(spec)...)
				}
			case *dst.ValueSpec:
				if f.includeDocumented || !nodes.HasDoc(spec.Decs.NodeDecs.Start) {
					identifier, exported = nodes.Identifier(spec)
				}
			}
		}

		if exported {
			findings = append(findings, identifier)
		}
	}

	slices.Sort(findings)

	return findings, nil
}

func (f *Finder) findInterfaceMethods(spec *dst.TypeSpec) []string {
	var findings []string

	ifaceName := spec.Name.Name
	for _, method := range spec.Type.(*dst.InterfaceType).Methods.List {
		if len(method.Names) == 0 {
			continue
		}
		name := method.Names[0].Name
		ident := fmt.Sprintf("func:%s.%s", ifaceName, name)
		if nodes.IsExportedIdentifier(ident) && (f.includeDocumented || !nodes.HasDoc(method.Decs.Start)) {
			findings = append(findings, ident)
		}
	}

	return findings
}

func isInterface(spec *dst.TypeSpec) bool {
	_, ok := spec.Type.(*dst.InterfaceType)
	return ok
}

func isTestFunction(node *dst.FuncDecl) bool {
	return strings.HasPrefix(node.Name.Name, "Test")
}
