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
	"github.com/modernice/jotbot/tools/reset"
	"github.com/tiktoken-go/tokenizer"
)

var (
	// FileExtensions is a slice of strings representing the file extensions
	// supported by the package, such as ".go" for Go source files.
	FileExtensions = []string{".go"}

	// DefaultMinification is a predefined set of minification options used by the
	// Service to minify Go source code. The options include minifying unexported
	// elements, exporting functions with their bodies, and minifying all elements.
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

// Service is a type that provides functionality for parsing, minifying, and
// patching Go source code. It can find identifiers, minify source code based on
// configured steps, generate prompts, and apply documentation patches to
// specified identifiers. The Service can be customized with various options,
// such as setting a Finder for searching elements in the source code,
// specifying a model for the tokenizer, and configuring minification steps.
type Service struct {
	model         string
	maxTokens     int
	clearComments bool
	codec         tokenizer.Codec
	finder        *Finder
	minifySteps   []nodes.MinifyOptions
}

// Option is a function that modifies a [Service] instance. It is used to
// provide optional configurations for the [Service] during its creation, such
// as setting the finder, model, or minification steps.
type Option func(*Service)

// WithFinder sets the Finder for a Service, which is used to search for
// specific elements in the source code.
func WithFinder(f *Finder) Option {
	return func(s *Service) {
		s.finder = f
	}
}

// Model sets the model name for a Service. The provided model name is used to
// create a tokenizer.Codec and determine the maximum number of tokens allowed
// for code minification.
func Model(m string) Option {
	return func(s *Service) {
		s.model = m
	}
}

// Minify is an Option that sets the minification steps for a Service,
// specifying how the source code should be minified in each step.
func Minify(steps []nodes.MinifyOptions) Option {
	return func(s *Service) {
		s.minifySteps = steps
	}
}

// ClearComments is an Option that sets whether or not to remove comments from
// the code when generating a prompt for the Service. If clear is true, comments
// will be removed; otherwise, they will be preserved.
func ClearComments(clear bool) Option {
	return func(s *Service) {
		s.clearComments = clear
	}
}

// Must creates a new Service with the provided options and panics if an error
// occurs. It is useful for creating a Service when it is known that the
// provided options are valid and an error will not occur.
func Must(opts ...Option) *Service {
	svc, err := New(opts...)
	if err != nil {
		panic(err)
	}
	return svc
}

// New creates a new Service with the provided options, initializes its
// tokenizer based on the model, and sets the maximum tokens allowed. If no
// finder is provided, it creates a new Finder. Returns an error if there's an
// issue creating the tokenizer.
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

// Extensions returns a copy of the slice containing file extensions supported
// by the Service.
func (svc *Service) Extensions() []string {
	return append([]string{}, FileExtensions...)
}

// Find returns a slice of strings representing the identifiers found in the
// provided code byte slice. An error is returned if the search for identifiers
// fails.
func (svc *Service) Find(code []byte) ([]string, error) {
	return svc.finder.Find(code)
}

// Minify takes a byte slice of code and returns a minified version of the code,
// removing unnecessary elements based on the minification steps configured for
// the Service. If the minified code's token count is within the limit, it
// returns the minified code; otherwise, it returns an error.
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

// Prompt takes an input of type [generate.Input] and returns a string. If the
// [Service] is configured to clear comments, it clears the comments from the
// input code before generating the prompt.
func (svc *Service) Prompt(input generate.Input) string {
	if svc.clearComments {
		if node, err := nodes.Parse(input.Code); err == nil {
			reset.Comments(node)
			if code, err := nodes.Format(node); err == nil {
				input.Code = code
			}
		}
	}
	return Prompt(input)
}

// Patch applies a given documentation string to the specified identifier in the
// provided code, and returns the updated code. It updates the documentation for
// functions, general declarations, type specifications, value specifications,
// and fields. If the specified identifier is not found in the code, an error is
// returned.
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
	return internal.RemoveColumns(strings.ReplaceAll(doc, "// ", ""))
}

func updateDoc(decs *dst.Decorations, doc string) {
	decs.Clear()
	if doc != "" {
		decs.Append(formatDoc(doc))
	}
}
