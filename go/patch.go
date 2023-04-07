package opendocs

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"os"
	"strings"
)

type Patcher struct {
	filepath string
	file     []byte
}

func NewPatcher(filepath string) (*Patcher, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("open %q: %v", filepath, err)
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read %q: %v", filepath, err)
	}

	return &Patcher{
		filepath: filepath,
		file:     b,
	}, nil
}

func (p *Patcher) Patch(typeName, documentation string) error {
	documentation = strings.Trim(documentation, "\"")

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, "", p.file, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse file: %v", err)
	}

	var decl *ast.GenDecl
L:
	for _, topDecl := range f.Decls {
		// Filter out non-type declarations.
		genDecl, ok := topDecl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			if ts, ok := spec.(*ast.TypeSpec); ok && ts.Name.Name == typeName {
				decl = genDecl
				break L
			}
		}
	}

	if decl == nil {
		return fmt.Errorf("type %q not found", typeName)
	}

	if decl.Doc != nil {
		return fmt.Errorf("type %q already has documentation", typeName)
	}

	decl.Doc = &ast.CommentGroup{
		List: []*ast.Comment{{
			Text:  formatComment(documentation),
			Slash: decl.Pos() - 1,
		}},
	}

	var output bytes.Buffer
	err = printer.Fprint(&output, fset, f)
	if err != nil {
		return err
	}

	formatted, err := format.Source(output.Bytes())
	if err != nil {
		return fmt.Errorf("format file: %v", err)
	}

	err = os.WriteFile(p.filepath, formatted, 0644)
	if err != nil {
		return fmt.Errorf("write file: %v", err)
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
