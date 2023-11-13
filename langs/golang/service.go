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
	// FileExtensions represents the set of file extensions that are supported by
	// the service for processing.
	FileExtensions = []string{".go"}

	// DefaultMinification represents a sequence of steps used to progressively
	// simplify the code structure, potentially making it more concise while aiming
	// to reduce its token count without altering its functionality. It serves as
	// the initial configuration for services that perform code minification.
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

// Service represents an abstraction for processing and manipulating Go source
// code. It facilitates various operations such as finding specific elements
// within the code, minifying the code to reduce its token count, clearing
// comments from the source, generating prompts based on the provided input, and
// patching the documentation of code entities. The Service can be customized
// with different options that alter its behavior, including setting a custom
// finder for locating elements in the code, specifying a model for
// tokenization, defining minification steps, and deciding whether to clear
// comments during prompt generation. It provides functionality to handle file
// extensions associated with Go source files and performs encoding of the
// source code into tokens using an internal tokenizer. The Service ensures that
// the resultant minified code does not exceed a predefined maximum token count
// and allows for dynamic updates to documentation comments within the source
// code.
type Service struct {
	model         string
	maxTokens     int
	clearComments bool
	codec         tokenizer.Codec
	finder        *Finder
	minifySteps   []nodes.MinifyOptions
}

// Option configures a Service by setting various internal fields such as model,
// finder, and minification steps. It is used in conjunction with the Service
// constructor and other setup functions that accept optional configurations to
// modify the behavior of the Service instance.
type Option func(*Service)

// WithFinder sets the finder to be used by the Service for locating specific
// elements or patterns in code. It accepts a Finder and returns an Option that,
// when applied, configures a Service with the provided Finder.
func WithFinder(f *Finder) Option {
	return func(s *Service) {
		s.finder = f
	}
}

// Model configures the model identifier for a Service. It sets the underlying
// model that the Service will use for operations such as tokenization and code
// analysis.
func Model(m string) Option {
	return func(s *Service) {
		s.model = m
	}
}

// Minify applies a series of transformations to Go source code represented as a
// byte slice to reduce its size, potentially making it more suitable for
// processing within token-based limitations. It returns the minified source
// code as a byte slice and an error if the minification process fails. If the
// resulting code after minification still exceeds the maximum allowed token
// count, an error is returned detailing the token limit and the actual token
// count.
func Minify(steps []nodes.MinifyOptions) Option {
	return func(s *Service) {
		s.minifySteps = steps
	}
}

// ClearComments configures whether a [*Service] should remove comments from the
// code during processing.
func ClearComments(clear bool) Option {
	return func(s *Service) {
		s.clearComments = clear
	}
}

// Must creates a new Service with the provided options, panicking if an error
// occurs during its creation. It ensures that a Service is returned without the
// need to handle errors directly, simplifying initialization in cases where
// failure is not expected or cannot be recovered from. It is intended for use
// when the program cannot continue if the Service cannot be constructed. Must
// returns a [*Service].
func Must(opts ...Option) *Service {
	svc, err := New(opts...)
	if err != nil {
		panic(err)
	}
	return svc
}

// New initializes a new Service with the provided options. It returns a pointer
// to the initialized Service and an error if there is any problem during the
// initialization.
func New(opts ...Option) (*Service, error) {
	svc := Service{minifySteps: DefaultMinification}
	for _, opt := range opts {
		opt(&svc)
	}

	if svc.model == "" {
		svc.model = openai.DefaultModel
	}

	codec, err := internal.OpenAITokenizer(svc.model)
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

// Extensions retrieves a list of file extensions that the service recognizes
// and works with. It returns a fresh copy of the list to avoid modifying the
// original list of supported extensions.
func (svc *Service) Extensions() []string {
	return append([]string{}, FileExtensions...)
}

// Find locates identifiers within the provided source code and returns a list
// of those identifiers along with any errors encountered during the search. It
// defers the actual searching to the associated Finder type within the Service.
// If no identifiers are found or an error occurs, it may return an empty list
// and the corresponding error.
func (svc *Service) Find(code []byte) ([]string, error) {
	return svc.finder.Find(code)
}

// Minify reduces the size of the given Go source code while aiming to preserve
// its functionality. It applies a series of transformations defined by the
// service's configuration to progressively simplify and shrink the code. The
// method stops minifying when the code's size falls below a certain threshold
// measured in tokens or when no further reductions can be made without
// exceeding that limit. If successful, it returns the minified code; otherwise,
// it returns an error indicating why minification failed, such as if the
// resulting code still exceeds the maximum allowed token count.
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

// Prompt prepares the input code by potentially clearing comments and then
// passes the modified input to the underlying Prompt function. If the
// clearComments option is enabled in the Service, it removes all comments from
// the input code before generating a prompt. It returns the generated output as
// a string.
func (svc *Service) Prompt(input generate.PromptInput) string {
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

// Patch applies a documentation string to the declaration identified by the
// specified identifier within the given source code. It updates or adds
// documentation comments in the source code while preserving the original
// structure and formatting. If the identifier does not exist within the source
// code, Patch returns an error indicating that the node was not found. On
// successful application of the documentation string, Patch returns the updated
// source code as a byte slice along with a nil error. If an error occurs during
// parsing or formatting of the source code, Patch will return the error
// encountered.
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
