package find

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"golang.org/x/exp/slices"
)

type Finder struct {
	repo fs.FS
}

type Finding struct {
	Path       string
	Identifier string
}

type Findings map[string][]Finding

func New(repo fs.FS) *Finder {
	return &Finder{repo}
}

func (f *Finder) Uncommented() (Findings, error) {
	allFindings := make(Findings)

	if err := fs.WalkDir(f.repo, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if d.Name() != "." && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			if d.Name() == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}

		if !isGoFile(d) || isTestFile(d) {
			return nil
		}

		findings, err := f.findUncommented(path)
		if err != nil {
			return fmt.Errorf("find uncommented code in %q: %w", path, err)
		}

		allFindings[path] = append(allFindings[path], findings...)

		return nil
	}); err != nil {
		return nil, err
	}

	return allFindings, nil
}

func (f *Finder) findUncommented(path string) ([]Finding, error) {
	var findings []Finding

	codeFile, err := f.repo.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", path, err)
	}
	defer codeFile.Close()

	code, err := io.ReadAll(codeFile)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", path, err)
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", code, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parse code: %w", err)
	}

	ast.Inspect(node, func(node ast.Node) bool {
		var (
			identifier string
			cont       = true
		)

		switch node := node.(type) {
		case *ast.FuncDecl:
			if node.Doc != nil {
				break
			}

			identifier = node.Name.Name

			if node.Recv != nil && len(node.Recv.List) > 0 {
				identifier = methodIdentifier(identifier, node.Recv.List[0].Type)
			}
		case *ast.GenDecl:
			if node.Doc != nil {
				break
			}

			if len(node.Specs) == 0 {
				return true
			}

			spec := node.Specs[0]
			cont = false

			switch spec := spec.(type) {
			case *ast.TypeSpec:
				if node.Doc == nil {
					identifier = spec.Name.Name
				}
			case *ast.ValueSpec:
				if node.Doc == nil {
					identifier = spec.Names[0].Name
				}
			}
		}

		if identifier != "" && identifier != "_" {
			findings = append(findings, Finding{
				Path:       path,
				Identifier: identifier,
			})
		}

		return cont
	})

	slices.SortFunc(findings, func(a, b Finding) bool {
		if a.Path < b.Path {
			return true
		}

		if a.Path == b.Path {
			return a.Identifier <= b.Identifier
		}

		return false
	})

	return findings, nil
}

func isGoFile(d fs.DirEntry) bool {
	return filepath.Ext(d.Name()) == ".go"
}

func isTestFile(d fs.DirEntry) bool {
	return strings.HasSuffix(d.Name(), "_test.go")
}

func methodIdentifier(identifier string, recv ast.Expr) string {
	switch recv := recv.(type) {
	case *ast.Ident:
		return recv.Name + "." + identifier
	case *ast.StarExpr:
		if ident, ok := recv.X.(*ast.Ident); ok {
			return "*" + ident.Name + "." + identifier
		}
	}
	return identifier
}
