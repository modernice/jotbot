package ts

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/modernice/jotbot/generate"
	"github.com/modernice/jotbot/internal"
	"github.com/modernice/jotbot/internal/slice"
	"github.com/modernice/jotbot/services/openai"
)

// FileExtensions is a slice of supported file extensions for TypeScript and
// JavaScript files, including ".ts", ".tsx", ".js", ".jsx", ".mjs", and ".cjs".
var (
	FileExtensions = []string{".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs"}
)

// Service is a type that provides utility functions for TypeScript and
// JavaScript files, such as finding identifiers, minifying code, generating
// prompts, and patching comments into the code. It can be configured with
// different file finders and OpenAI models using the provided options.
type Service struct {
	finder *Finder
	model  string
}

// Option is a functional option type for configuring a [Service] instance. Use
// available options like WithFinder and Model to customize the behavior of the
// [Service].
type Option func(*Service)

// WithFinder sets the Finder for a Service. It is an Option for configuring a
// Service instance.
func WithFinder(f *Finder) Option {
	return func(s *Service) {
		s.finder = f
	}
}

// Model sets the model to be used by the Service in the given Option.
func Model(model string) Option {
	return func(s *Service) {
		s.model = model
	}
}

// New creates a new Service instance with the provided options. If no model or
// finder is specified, it uses the default openai model and a new Finder
// instance.
func New(opts ...Option) *Service {
	var svc Service
	for _, opt := range opts {
		opt(&svc)
	}
	if svc.model == "" {
		svc.model = openai.DefaultModel
	}
	if svc.finder == nil {
		svc.finder = NewFinder()
	}
	return &svc
}

// Extensions returns a slice of supported file extensions for the Service, such
// as ".ts", ".tsx", ".js", ".jsx", ".mjs", and ".cjs".
func (svc *Service) Extensions() []string {
	return FileExtensions
}

// Find searches the given code for TypeScript and JavaScript identifiers and
// returns a slice of found identifiers. It may return an error if the search
// fails.
func (svc *Service) Find(code []byte) ([]string, error) {
	return svc.finder.Find(context.Background(), code)
}

// Minify takes a byte slice of code and returns a minified version of the code
// as a byte slice, using the associated model. It returns an error if the
// minification process fails.
func (svc *Service) Minify(code []byte) ([]byte, error) {
	args := []string{"minify", "-m", svc.model, string(code)}

	cmd := exec.Command(jotbotTSPath, args...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w:\n%s", err, out)
	}

	return out, nil
}

// Prompt generates a prompt string from the provided generate.Input, which can
// be used in a code generation service request.
func (svc *Service) Prompt(input generate.Input) string {
	return Prompt(input)
}

// Patch inserts a formatted documentation comment for the given identifier in
// the provided code and returns the modified code. The generated comment is
// based on the provided doc string and inserted at the position determined by
// the finder service.
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
	doc = NormalizeGeneratedComment(doc)

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

// NormalizeGeneratedComment removes leading and trailing whitespaces, removes
// comment prefixes, and collapses consecutive whitespaces in the given
// documentation string. It also ensures that comment tokens are not
// accidentally closed within the comment text.
func NormalizeGeneratedComment(doc string) string {
	doc = strings.TrimSpace(doc)
	doc = strings.TrimPrefix(doc, "/**")
	doc = strings.TrimSuffix(doc, "*/")
	doc = strings.ReplaceAll(doc, "*/", "*\\/")

	lines := strings.Split(doc, "\n")

	lines = slice.Map(lines, func(l string) string {
		return commentLinePrefixRE.ReplaceAllString(l, "")
	})

	lines = slice.Filter(lines, func(l string) bool {
		return !strings.HasPrefix(l, "@")
	})

	return internal.RemoveColumns(strings.TrimSpace(strings.Join(lines, "\n")))
}
