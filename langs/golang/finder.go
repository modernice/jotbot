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

// Finder is a utility that searches for exported identifiers in Go code, with
// options to include or exclude test functions and documented identifiers. It
// can be configured using FinderOption functions like FindTests and
// IncludeDocumented. The Find method takes a byte slice of Go code and returns
// a sorted slice of strings representing the found exported identifiers.
type Finder struct {
	findTests         bool
	includeDocumented bool
}

// FinderOption is a function type that modifies the behavior of a Finder, which
// searches for exported identifiers in Go code. Common options include
// FindTests and IncludeDocumented.
type FinderOption func(*Finder)

// FindTests is a function that returns a FinderOption which sets the findTests
// field of a Finder. The findTests field determines whether or not test
// functions should be included in the search results.
func FindTests(find bool) FinderOption {
	return func(f *Finder) {
		f.findTests = find
	}
}

// IncludeDocumented modifies a Finder to include documented declarations if the
// passed boolean is true. If false, only undocumented declarations will be
// considered by the Finder.
func IncludeDocumented(include bool) FinderOption {
	return func(f *Finder) {
		f.includeDocumented = include
	}
}

// NewFinder creates a new Finder with the provided options. The Finder can be
// used to find exported identifiers in Go code, optionally including test
// functions and documented identifiers.
func NewFinder(opts ...FinderOption) *Finder {
	var f Finder
	for _, opt := range opts {
		opt(&f)
	}
	return &f
}

// Find searches the provided code for exported identifiers, such as functions,
// types, and variables. It returns a sorted slice of strings containing the
// found identifiers. The search can be configured to include or exclude test
// functions and documented identifiers with FinderOptions.
func (f *Finder) Find(code []byte) ([]string, error) {
	var findings []string

	fset := token.NewFileSet()
	node, err := decorator.ParseFile(fset, "", code, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parse code: %w", err)
	}

	for _, node := range node.Decls {
		// var (
		// 	identifier string
		// 	exported   bool
		// )

		switch node := node.(type) {
		case *dst.FuncDecl:
			if !f.findTests && isTestFunction(node) {
				break
			}

			if !f.includeDocumented && nodes.HasDoc(node.Decs.NodeDecs.Start) {
				break
			}

			if identifier, exported := nodes.Identifier(node); exported {
				findings = append(findings, identifier)
			}
		case *dst.GenDecl:
			if !f.includeDocumented && nodes.HasDoc(node.Decs.NodeDecs.Start) {
				break
			}

			if len(node.Specs) == 0 {
				break
			}

			for _, spec := range node.Specs {
				switch spec := spec.(type) {
				case *dst.TypeSpec:
					if f.includeDocumented || !nodes.HasDoc(spec.Decs.NodeDecs.Start) {
						if identifier, exported := nodes.Identifier(spec); exported {
							findings = append(findings, identifier)
						}
					}

					if isInterface(spec) {
						findings = append(findings, f.findInterfaceMethods(spec)...)
					}
				case *dst.ValueSpec:
					if f.includeDocumented || !nodes.HasDoc(spec.Decs.NodeDecs.Start) {
						if identifier, exported := nodes.Identifier(spec); exported {
							findings = append(findings, identifier)
						}
					}
				}
			}
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
