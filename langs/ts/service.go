package ts

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/modernice/jotbot/generate"
	"github.com/modernice/jotbot/internal"
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

func (svc *Service) Find(code []byte) ([]string, error) {
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

	lines := internal.Columns(doc, 77-indent)

	if len(lines) == 1 {
		return fmt.Sprintf("/** %s */\n", strings.TrimSpace(lines[0]))
	}

	lines = slice.Map(lines, func(l string) string {
		return " * " + l
	})

	return "/**\n" + strings.Join(lines, "\n") + "\n */\n"
}

var commentLinePrefixRE = regexp.MustCompile(`^\s\*\s?`)

func normalizeGeneratedComment(doc string) string {
	doc = strings.TrimSpace(doc)
	doc = strings.TrimPrefix(doc, "/**")
	doc = strings.TrimSuffix(doc, "*/")
	doc = strings.ReplaceAll(doc, "*/", "*\\/")

	lines := strings.Split(doc, "\n")
	lines = slice.Map(lines, func(l string) string {
		return commentLinePrefixRE.ReplaceAllString(l, "")
	})

	return strings.TrimSpace(strings.Join(lines, "\n"))
}
