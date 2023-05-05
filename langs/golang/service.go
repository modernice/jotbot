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

	spec, decl, ok := nodes.Find(identifier, file)
	if !ok {
		return nil, fmt.Errorf("node %q not found", identifier)
	}

	target := nodes.CommentTarget(spec, decl)

	doc = formatDoc(doc)

	switch target := target.(type) {
	case *dst.FuncDecl:
		target.Decs.Start.Clear()
		if doc != "" {
			target.Decs.Start.Append(doc)
		}
	case *dst.GenDecl:
		target.Decs.Start.Clear()
		if doc != "" {
			target.Decs.Start.Append(doc)
		}
	case *dst.TypeSpec:
		target.Decs.Start.Clear()
		if doc != "" {
			target.Decs.Start.Append(doc)
		}
	case *dst.ValueSpec:
		target.Decs.Start.Clear()
		if doc != "" {
			target.Decs.Start.Append(doc)
		}
	}

	return nodes.Format(file)
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
