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

	"github.com/modernice/opendocs/go/internal"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
)

type Finder struct {
	repo fs.FS
	log  *slog.Logger
}

type Finding struct {
	Path       string
	Identifier string
}

type Findings map[string][]Finding

type Option func(*Finder)

func WithLogger(h slog.Handler) Option {
	return func(f *Finder) {
		f.log = slog.New(h)
	}
}

func New(repo fs.FS, opts ...Option) *Finder {
	f := &Finder{repo: repo}
	for _, opt := range opts {
		opt(f)
	}
	if f.log == nil {
		f.log = internal.NopLogger()
	}
	return f
}

func (f *Finder) Uncommented() (Findings, error) {
	f.log.Info("Searching for uncommented code in repository ...", "repo", f.repo)

	allFindings := make(Findings)

	if err := fs.WalkDir(f.repo, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if d.Name() != "." && strings.HasPrefix(d.Name(), ".") {
				f.log.Debug("Skipping directory", "path", path, "reason", "hidden directory")
				return filepath.SkipDir
			}
			if d.Name() == "testdata" {
				f.log.Debug("Skipping directory", "path", path, "reason", "testdata directory")
				return filepath.SkipDir
			}
			return nil
		}

		if !isGoFile(d) {
			f.log.Debug("Skipping file", "path", path, "reason", "not a Go file")
			return nil
		}

		if isTestFile(d) {
			f.log.Debug("Skipping file", "path", path, "reason", "test file")
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
	f.log.Info(fmt.Sprintf("Searching for uncommented code in %q ...", path))

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

	for _, node := range node.Decls {
		var identifier string

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
				break
			}

			spec := node.Specs[0]

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
	}

	slices.SortFunc(findings, func(a, b Finding) bool {
		if a.Path < b.Path {
			return true
		}

		if a.Path == b.Path {
			return a.Identifier <= b.Identifier
		}

		return false
	})

	f.log.Info(fmt.Sprintf("Found %d uncommented types/functions in %q", len(findings), path))

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
