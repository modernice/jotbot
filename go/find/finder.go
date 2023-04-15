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

	for _, decl := range node.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if decl.Doc == nil {
				findings = append(findings, Finding{
					Path:       path,
					Identifier: decl.Name.Name,
				})
			}
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch spec := spec.(type) {
				case *ast.TypeSpec:
					if decl.Doc == nil {
						findings = append(findings, Finding{
							Path:       path,
							Identifier: spec.Name.Name,
						})
					}
				case *ast.ValueSpec:
					if decl.Doc == nil {
						findings = append(findings, Finding{
							Path:       path,
							Identifier: spec.Names[0].Name,
						})
					}
				}
			}
		}
	}

	return findings, nil
}

func isGoFile(d fs.DirEntry) bool {
	return filepath.Ext(d.Name()) == ".go"
}

func isTestFile(d fs.DirEntry) bool {
	return strings.HasSuffix(d.Name(), "_test.go")
}
