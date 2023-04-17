package find

import (
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/modernice/opendocs/internal"
	"github.com/modernice/opendocs/internal/nodes"
	"github.com/modernice/opendocs/internal/slice"
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
			return fmt.Errorf("find uncommented code in %s: %w", path, err)
		}

		allFindings[path] = append(allFindings[path], findings...)

		return nil
	}); err != nil {
		return nil, err
	}

	return allFindings, nil
}

func (f *Finder) findUncommented(path string) ([]Finding, error) {
	f.log.Info(fmt.Sprintf("Searching for uncommented code in %s ...", path))

	var findings []Finding

	codeFile, err := f.repo.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	defer codeFile.Close()

	code, err := io.ReadAll(codeFile)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	fset := token.NewFileSet()
	node, err := decorator.ParseFile(fset, "", code, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parse code: %w", err)
	}

	for _, node := range node.Decls {
		var identifier string

		switch node := node.(type) {
		case *dst.FuncDecl:
			if nodes.HasDoc(node.Decs.NodeDecs.Start) {
				break
			}

			identifier = nodes.Identifier(node)
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
					identifier = nodes.Identifier(spec)
				}
			case *dst.ValueSpec:
				if !nodes.HasDoc(spec.Decs.NodeDecs.Start) {
					identifier = nodes.Identifier(spec)
				}
			}
		}

		if identifier != "" {
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

	idents := slice.Map(findings, func(f Finding) string { return f.Identifier })

	f.log.Info(fmt.Sprintf("Found %d uncommented types/functions in %s", len(findings), path), "identifiers", idents)

	return findings, nil
}

func isGoFile(d fs.DirEntry) bool {
	return filepath.Ext(d.Name()) == ".go"
}

func isTestFile(d fs.DirEntry) bool {
	return strings.HasSuffix(d.Name(), "_test.go")
}
