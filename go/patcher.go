package opendocs

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"os"
	"strings"
)

type Patcher struct {
	ast  *ast.File
	fset *token.FileSet
}

func NewPatcher(code []byte) (*Patcher, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", code, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parse code: %w", err)
	}
	return &Patcher{ast: f, fset: fset}, nil
}

func (p *Patcher) Comment(identifier, comment string) error {
	if decl, ok := p.findFunction(identifier); ok {
		return p.commentFunction(decl, comment)
	}

	if spec, decl, ok := p.findType(identifier); ok {
		return p.commentType(decl, spec, comment)
	}

	return fmt.Errorf("could not find %q", identifier)
}

func (p *Patcher) findFunction(identifier string) (*ast.FuncDecl, bool) {
	for _, astDecl := range p.ast.Decls {
		if fd, ok := astDecl.(*ast.FuncDecl); ok && fd.Name.Name == identifier {
			return fd, true
		}
	}
	return nil, false
}

func (p *Patcher) commentFunction(decl *ast.FuncDecl, comment string) error {
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

func (p *Patcher) findType(identifier string) (*ast.TypeSpec, *ast.GenDecl, bool) {
	for _, astDecl := range p.ast.Decls {
		if decl, ok := astDecl.(*ast.GenDecl); ok {
			for _, spec := range decl.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok && ts.Name.Name == identifier {
					return ts, decl, true
				}
			}
		}
	}
	return nil, nil, false
}

func (p *Patcher) commentType(decl *ast.GenDecl, spec *ast.TypeSpec, comment string) error {
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

func (p *Patcher) Apply(w io.Writer) error {
	if err := format.Node(w, p.fset, p.ast); err != nil {
		return fmt.Errorf("format code: %w", err)
	}
	return nil
}

func (p *Patcher) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := p.Apply(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p *Patcher) Patch(path string) error {
	backup, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("create backup of %q: %w", path, err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %q: %w", path, err)
	}

	if err := p.Apply(f); err != nil {
		f.Close()
		return p.restoreBackup(path, backup, err)
	}

	if err := f.Close(); err != nil {
		return p.restoreBackup(path, backup, err)
	}

	return nil
}

func (p *Patcher) restoreBackup(path string, backup []byte, parentError error) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("restore backup: create %q: %w [originalError=%s]", path, err, parentError)
	}

	if _, err := f.Write(backup); err != nil {
		f.Close()
		return fmt.Errorf("restore backup: %w [originalError=%s]", err, parentError)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("restore backup: %w [originalError=%s]", err, parentError)
	}

	return nil
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
