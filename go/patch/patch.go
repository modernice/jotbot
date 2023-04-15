package patch

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Patch struct {
	repo        fs.FS
	fset        *token.FileSet
	files       map[string]*ast.File
	identifiers map[string][]string
}

func New(repo fs.FS) *Patch {
	return &Patch{
		repo:        repo,
		fset:        token.NewFileSet(),
		files:       make(map[string]*ast.File),
		identifiers: make(map[string][]string),
	}
}

func (p *Patch) Identifiers() map[string][]string {
	return p.identifiers
}

func (p *Patch) Comment(file, identifier, comment string) (rerr error) {
	defer func() {
		if rerr == nil {
			p.identifiers[file] = append(p.identifiers[file], identifier)
		}
	}()

	p.identifiers[file] = append(p.identifiers[file], identifier)
	{
		decl, ok, err := p.findFunction(file, identifier)
		if err != nil {
			return fmt.Errorf("look for function %q in %q: %w", identifier, file, err)
		}
		if ok {
			return p.commentFunction(decl, comment)
		}
	}

	{
		spec, decl, ok, err := p.findType(file, identifier)
		if err != nil {
			return fmt.Errorf("look for type %q in %q: %w", identifier, file, err)
		}
		if ok {
			return p.commentType(decl, spec, comment)
		}
	}

	return fmt.Errorf("could not find %q in %q", identifier, file)
}

func (p *Patch) parseFile(path string) (*ast.File, error) {
	if node, ok := p.files[path]; ok {
		return node, nil
	}

	f, err := p.repo.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	defer f.Close()

	code, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", path, err)
	}

	node, err := parser.ParseFile(p.fset, "", code, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parse %q: %w", path, err)
	}
	p.files[path] = node

	return node, nil
}

func (p *Patch) findFunction(file, identifier string) (*ast.FuncDecl, bool, error) {
	node, err := p.parseFile(file)
	if err != nil {
		return nil, false, err
	}
	for _, astDecl := range node.Decls {
		if fd, ok := astDecl.(*ast.FuncDecl); ok && fd.Name.Name == identifier {
			return fd, true, nil
		}
	}
	return nil, false, nil
}

func (p *Patch) commentFunction(decl *ast.FuncDecl, comment string) error {
	if decl.Doc != nil {
		return fmt.Errorf("function %q already has documentation", decl.Name.Name)
	}

	decl.Doc = &ast.CommentGroup{
		List: []*ast.Comment{{
			Text:  formatComment(comment),
			Slash: decl.Pos() - 1,
		}},
	}

	return nil
}

func (p *Patch) findType(file, identifier string) (*ast.TypeSpec, *ast.GenDecl, bool, error) {
	node, err := p.parseFile(file)
	if err != nil {
		return nil, nil, false, err
	}

	for _, astDecl := range node.Decls {
		if decl, ok := astDecl.(*ast.GenDecl); ok {
			for _, spec := range decl.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok && ts.Name.Name == identifier {
					return ts, decl, true, nil
				}
			}
		}
	}
	return nil, nil, false, nil
}

func (p *Patch) commentType(decl *ast.GenDecl, spec *ast.TypeSpec, comment string) error {
	if decl.Doc != nil {
		return fmt.Errorf("type %q already has documentation", spec.Name.Name)
	}

	// INFO(bounoable): ChatGPT said this is the way to go to calculate the
	// slash position, but I don't know if this is really necessary TBH.
	line := p.fset.Position(decl.Pos()).Line - 1
	slash := p.fset.File(decl.Pos()).LineStart(line)

	decl.Doc = &ast.CommentGroup{
		List: []*ast.Comment{{
			Text:  formatComment(comment),
			Slash: slash,
		}},
	}

	return nil
}

func (p *Patch) Apply(repo string) error {
	for path, node := range p.files {
		var buf bytes.Buffer
		if err := format.Node(&buf, p.fset, node); err != nil {
			return fmt.Errorf("format %q in %q: %w", node.Name.Name, path, err)
		}

		path = filepath.Join(repo, path)
		if err := p.patchFile(path, &buf); err != nil {
			return fmt.Errorf("patch %q: %w", path, err)
		}
	}
	return nil
}

func (p *Patch) patchFile(path string, buf *bytes.Buffer) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %q: %w", path, err)
	}
	defer f.Close()

	_, err = io.Copy(f, buf)
	return err
}

func (p *Patch) DryRun() (map[string][]byte, error) {
	result := make(map[string][]byte)

	for path, node := range p.files {
		var buf bytes.Buffer
		if err := format.Node(&buf, p.fset, node); err != nil {
			return nil, fmt.Errorf("format %q in %q: %w", node.Name.Name, path, err)
		}
		result[path] = buf.Bytes()
	}

	return result, nil
}

func formatComment(doc string) string {
	var buf bytes.Buffer
	for _, line := range splitString(doc, 80) {
		buf.WriteString("// ")
		buf.WriteString(line)
		buf.WriteByte('\n')
	}
	return buf.String()
}

func splitString(input string, maxLength int) []string {
	var result []string
	words := strings.Fields(input)
	currentLine := ""

	for _, word := range words {
		if len(currentLine)+len(word)+1 > maxLength {
			result = append(result, currentLine)
			currentLine = word
		} else {
			if currentLine != "" {
				currentLine += " "
			}
			currentLine += word
		}
	}

	if currentLine != "" {
		result = append(result, currentLine)
	}

	return result
}