package golang

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/modernice/jotbot/find"
	"github.com/modernice/jotbot/generate"
	"github.com/modernice/jotbot/internal/nodes"
	"github.com/modernice/jotbot/internal/slice"
)

var FileExtensions = []string{".go"}

type Service struct {
	finder *Finder
}

type Option func(*Service)

func WithFinder(f *Finder) Option {
	return func(s *Service) {
		s.finder = f
	}
}

func New(opts ...Option) *Service {
	var svc Service
	for _, opt := range opts {
		opt(&svc)
	}
	if svc.finder == nil {
		svc.finder = NewFinder()
	}
	return &svc
}

func (svc *Service) Extensions() []string {
	return append([]string{}, FileExtensions...)
}

func (svc *Service) Find(code []byte) ([]find.Finding, error) {
	return svc.finder.Find(code)
}

func (svc *Service) Prompt(input generate.Input) string {
	return Prompt(input)
}

func (svc *Service) Patch(ctx context.Context, identifier, doc string, code []byte) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := decorator.ParseFile(fset, "", code, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parse code: %w", err)
	}
	return svc.patch(file, identifier, doc)
}

func (svc *Service) patch(file *dst.File, identifier, doc string) ([]byte, error) {
	file = dst.Clone(file).(*dst.File)

	identifier = nodes.StripIdentifierPrefix(identifier)

	if recv, method, isMethod := extractMethod(identifier); isMethod {
		if err := svc.patchMethod(file, recv, method, doc); err != nil {
			return nil, fmt.Errorf("patch method %q: %w", identifier, err)
		}
	} else if decl, ok := svc.findFunction(file, identifier); ok {
		svc.updateFunctionDoc(decl, doc)
	} else if spec, decl, ok := svc.findVar(file, identifier); ok {
		svc.updateGeneralDeclarationDoc(decl, spec.Names[0].Name, doc)
	} else if spec, decl, ok := svc.findType(file, identifier); ok {
		svc.updateGeneralDeclarationDoc(decl, spec.Name.Name, doc)
	} else {
		return nil, fmt.Errorf("node %q not found", identifier)
	}

	return nodes.Format(file)
}

func (svc *Service) findFunction(file *dst.File, identifier string) (*dst.FuncDecl, bool) {
	for _, astDecl := range file.Decls {
		if fd, ok := astDecl.(*dst.FuncDecl); ok && fd.Name.Name == identifier && fd.Recv == nil {
			return fd, true
		}
	}
	return nil, false
}

func (svc *Service) findVar(file *dst.File, identifier string) (*dst.ValueSpec, *dst.GenDecl, bool) {
	var (
		spec *dst.ValueSpec
		decl *dst.GenDecl
	)

	dst.Inspect(file, func(node dst.Node) bool {
		switch node := node.(type) {
		case *dst.GenDecl:
			for _, sp := range node.Specs {
				if vs, ok := sp.(*dst.ValueSpec); ok {
					for _, ident := range vs.Names {
						if ident.Name == identifier {
							spec = vs
							decl = node
							return false
						}
					}
				}
			}
		}
		return true
	})

	return spec, decl, spec != nil
}

func (svc *Service) findType(file *dst.File, identifier string) (*dst.TypeSpec, *dst.GenDecl, bool) {
	for _, astDecl := range file.Decls {
		if decl, ok := astDecl.(*dst.GenDecl); ok {
			for _, spec := range decl.Specs {
				if ts, ok := spec.(*dst.TypeSpec); ok && ts.Name.Name == identifier {
					return ts, decl, true
				}
			}
		}
	}
	return nil, nil, false
}

func (svc *Service) updateGeneralDeclarationDoc(decl *dst.GenDecl, identifier, doc string) {
	decl.Decs.Start.Clear()
	if doc != "" {
		decl.Decs.Start.Append(formatDoc(doc))
	}
}

func (svc *Service) patchMethod(file *dst.File, recv, method, doc string) error {
	decl, ok, err := svc.findMethod(file, recv, method)
	if err != nil {
		return fmt.Errorf("find method: %w", err)
	}
	if !ok {
		return fmt.Errorf("could not find method")
	}
	svc.updateFunctionDoc(decl, doc)
	return nil
}

func (svc *Service) findMethod(file *dst.File, recv, method string) (*dst.FuncDecl, bool, error) {
	recv = strings.TrimPrefix(recv, "*")
	if recv == "" {
		return nil, false, fmt.Errorf("empty receiver name")
	}

	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*dst.FuncDecl)
		if !ok {
			continue
		}

		if funcDecl.Name.Name != method {
			continue
		}

		if funcDecl.Recv == nil {
			continue
		}

		if len(funcDecl.Recv.List) == 0 {
			continue
		}

		rcv := funcDecl.Recv.List[0].Type

		if star, ok := rcv.(*dst.StarExpr); ok {
			rcv = star.X
		}

		if list, ok := rcv.(*dst.IndexListExpr); ok {
			rcv = list.X
		}

		if idx, ok := rcv.(*dst.IndexExpr); ok {
			rcv = idx.X
		}

		if ident, ok := rcv.(*dst.Ident); ok && ident.Name == recv {
			return funcDecl, true, nil
		}
	}

	return nil, false, nil
}

func (svc *Service) updateFunctionDoc(decl *dst.FuncDecl, doc string) {
	decl.Decs.Start.Clear()
	if doc != "" {
		decl.Decs.Start.Append(formatDoc(doc))
	}
}

func extractMethod(identifier string) (recv, method string, ok bool) {
	parts := strings.Split(identifier, ".")
	if len(parts) != 2 {
		return "", "", false
	}
	recv = parts[0]
	if strings.HasPrefix(recv, "(*") {
		recv = recv[2 : len(recv)-1]
	}
	return recv, parts[1], true
}

func formatDoc(doc string) string {
	lines := splitString(doc, 77)
	lines = slice.Map(lines, func(s string) string {
		return "// " + s
	})
	return strings.Join(lines, "\n")
}

func splitString(str string, maxLen int) []string {
	var out []string

	paras := strings.Split(str, "\n\n")
	for i, para := range paras {
		lines := splitByWords(para, maxLen)
		out = append(out, lines...)
		if i < len(paras)-1 {
			out = append(out, "")
		}
	}

	return out
}

func splitByWords(str string, maxLen int) []string {
	words := strings.Split(str, " ")

	var lines []string
	var line string
	for _, word := range words {
		if len(line)+len(word) > maxLen {
			lines = append(lines, line)
			line = ""
		}
		line += word + " "
	}
	lines = append(lines, line)

	return lines
}
