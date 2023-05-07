package golang

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/modernice/jotbot/generate"
	"github.com/modernice/jotbot/internal"
	"github.com/modernice/jotbot/internal/nodes"
	"github.com/modernice/jotbot/internal/slice"
	"github.com/modernice/jotbot/services/openai"
	"github.com/tiktoken-go/tokenizer"
)

var (
	FileExtensions      = []string{".go"}
	DefaultMinification = []nodes.MinifyOptions{
		nodes.MinifyUnexported,
		{
			FuncBody: true,
			Exported: true,
		},
		nodes.MinifyExported,
		nodes.MinifyAll,
	}
)

type Service struct {
	model       string
	maxTokens   int
	codec       tokenizer.Codec
	finder      *Finder
	minifySteps []nodes.MinifyOptions
}

type Option func(*Service)

func WithFinder(f *Finder) Option {
	return func(s *Service) {
		s.finder = f
	}
}

func Model(m string) Option {
	return func(s *Service) {
		s.model = m
	}
}

func Minify(steps []nodes.MinifyOptions) Option {
	return func(s *Service) {
		s.minifySteps = steps
	}
}

func Must(opts ...Option) *Service {
	svc, err := New(opts...)
	if err != nil {
		panic(err)
	}
	return svc
}

func New(opts ...Option) (*Service, error) {
	svc := Service{minifySteps: DefaultMinification}
	for _, opt := range opts {
		opt(&svc)
	}

	if svc.model == "" {
		svc.model = openai.DefaultModel
	}

	codec, err := tokenizer.ForModel(tokenizer.Model(svc.model))
	if err != nil {
		return nil, fmt.Errorf("create tokenizer: %w", err)
	}
	svc.codec = codec

	svc.maxTokens = openai.MaxTokensForModel(string(svc.model))

	if svc.finder == nil {
		svc.finder = NewFinder()
	}

	return &svc, err
}

func (svc *Service) Extensions() []string {
	return append([]string{}, FileExtensions...)
}

func (svc *Service) Find(code []byte) ([]string, error) {
	return svc.finder.Find(code)
}

func (svc *Service) Minify(code []byte) ([]byte, error) {
	if len(code) == 0 {
		return code, nil
	}

	if len(svc.minifySteps) == 0 {
		return code, nil
	}

	node, err := nodes.Parse(code)
	if err != nil {
		return nil, fmt.Errorf("parse code: %w", err)
	}

	var tokens []uint
	for _, step := range svc.minifySteps {
		formatted, err := nodes.Format(node)
		if err != nil {
			return nil, fmt.Errorf("format code: %w", err)
		}

		tokens, _, err = svc.codec.Encode(string(formatted))
		if err != nil {
			return nil, fmt.Errorf("encode code: %w", err)
		}

		if len(tokens) <= svc.maxTokens {
			return formatted, nil
		}

		node = nodes.Minify(node, step)

		minified, err := nodes.Format(node)
		if err != nil {
			return nil, fmt.Errorf("format minified code: %w", err)
		}

		tokens, _, err = svc.codec.Encode(string(minified))
		if err != nil {
			return nil, fmt.Errorf("encode minified code: %w", err)
		}

		if len(tokens) <= svc.maxTokens {
			return minified, nil
		}
	}

	return nil, fmt.Errorf("minified code exceeds %d tokens (%d tokens)", svc.maxTokens, len(tokens))
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
	spec, decl, ok := nodes.Find(identifier, file)
	if !ok {
		return nil, fmt.Errorf("node %q not found", identifier)
	}

	target := nodes.CommentTarget(spec, decl)

	switch target := target.(type) {
	case *dst.FuncDecl:
		updateDoc(&target.Decs.Start, doc)
		target.Decs.After = dst.EmptyLine
	case *dst.GenDecl:
		updateDoc(&target.Decs.Start, doc)
		target.Decs.After = dst.EmptyLine
	case *dst.TypeSpec:
		updateDoc(&target.Decs.Start, doc)
		target.Decs.After = dst.EmptyLine
	case *dst.ValueSpec:
		updateDoc(&target.Decs.Start, doc)
		target.Decs.After = dst.EmptyLine
	case *dst.Field:
		updateDoc(&target.Decs.Start, doc)
		target.Decs.After = dst.EmptyLine
	}

	return nodes.Format(file)
}

func formatDoc(doc string) string {
	doc = normalizeGeneratedComment(doc)

	lines := internal.Columns(doc, 77)
	lines = slice.Map(lines, func(s string) string {
		return "// " + s
	})
	return strings.Join(lines, "\n")
}

func normalizeGeneratedComment(doc string) string {
	return strings.ReplaceAll(doc, "// ", "")
}

func updateDoc(decs *dst.Decorations, doc string) {
	decs.Clear()
	if doc != "" {
		decs.Append(formatDoc(doc))
	}
}
