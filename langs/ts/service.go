package ts

import (
	"context"
	"fmt"
	"strings"

	"github.com/modernice/jotbot/find"
	"github.com/modernice/jotbot/generate"
	"github.com/modernice/jotbot/internal/slice"
)

var (
	FileExtensions = []string{".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs"}
)

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
	return FileExtensions
}

func (svc *Service) Find(code []byte) ([]find.Finding, error) {
	return svc.finder.Find(context.Background(), code)
}

func (svc *Service) Prompt(input generate.Input) string {
	return Prompt(input)
}

func (svc *Service) Patch(ctx context.Context, identifier, doc string, code []byte) ([]byte, error) {
	pos, err := svc.finder.Position(ctx, identifier, code)
	if err != nil {
		return nil, fmt.Errorf("find position of %q in code: %w", identifier, err)
	}

	if doc != "" {
		doc = formatDoc(doc, pos.Character)
	}

	return InsertComment(doc, code, pos)
}

func formatDoc(doc string, indent int) string {
	doc = normalizeGeneratedComment(doc)

	lines := splitString(doc, 77-indent)
	if len(lines) == 1 {
		return fmt.Sprintf("/** %s */\n", lines[0])
	}

	lines = slice.Map(lines, func(s string) string {
		return " * " + s
	})

	return "/**\n" + strings.Join(lines, "\n") + "\n */\n"
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

func normalizeGeneratedComment(doc string) string {
	doc = strings.TrimSpace(doc)
	doc = strings.TrimPrefix(doc, "/**")
	doc = strings.TrimSuffix(doc, "*/")
	lines := strings.Split(doc, "\n")
	lines = slice.Map(lines, func(l string) string {
		return strings.TrimPrefix(l, " * ")
	})
	return strings.Join(lines, "\n")
}
