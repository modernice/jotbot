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

// Finder locates identifiers in Go source code, taking into account options for
// including test functions and documented entities. It analyzes the provided
// code to produce a sorted list of exported names. The search can be customized
// through options to either include or exclude test functions and documented
// identifiers. When examining interface types, it also identifies and includes
// their exported methods. Finder returns a slice of strings representing the
// found identifiers and any errors encountered during the analysis process.
type Finder struct {
	findTests         bool
	includeDocumented bool
}

// FinderOption configures the behavior of a [*Finder] by setting its internal
// options. It is applied when constructing a new [*Finder] instance, allowing
// customization such as whether to include tests or documented entities in the
// search results.
type FinderOption func(*Finder)

// FindTests configures a Finder instance to determine whether it should
// identify test functions during code analysis. If the provided argument is
// true, the Finder will include test functions in its findings; otherwise, it
// will exclude them. This option can be passed to NewFinder to customize the
// Finder's behavior.
func FindTests(find bool) FinderOption {
	return func(f *Finder) {
		f.findTests = find
	}
}

// IncludeDocumented configures a Finder to consider documented entities during
// the search. When set to true, entities with associated documentation will be
// included in the findings; otherwise, they will be excluded. This option is
// used when creating a new Finder instance.
func IncludeDocumented(include bool) FinderOption {
	return func(f *Finder) {
		f.includeDocumented = include
	}
}

// NewFinder constructs a new Finder with optional configurations provided by
// FinderOptions. It returns a pointer to the initialized Finder.
func NewFinder(opts ...FinderOption) *Finder {
	var f Finder
	for _, opt := range opts {
		opt(&f)
	}
	return &f
}

// Find searches through the provided code for identifiers that are eligible
// based on the Finder's configuration. It returns a sorted slice of strings
// containing these identifiers and an error if the code cannot be parsed or
// another issue occurs. Identifiers from function declarations, type
// specifications, and value specifications are included unless they are
// filtered out by the Finder's settings, such as excluding test functions or
// documented identifiers.
func (f *Finder) Find(code []byte) ([]string, error) {
	var findings []string

	fset := token.NewFileSet()
	node, err := decorator.ParseFile(fset, "", code, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parse code: %w", err)
	}

	for _, node := range node.Decls {

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
