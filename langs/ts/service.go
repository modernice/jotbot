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

// FileExtensions holds a list of recognized file extensions for TypeScript and
// JavaScript source files. It includes extensions for both standard and
// module-specific file types.
var (
	FileExtensions = []string{".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs"}
)

// Service provides a set of operations for working with TypeScript code. It
// allows for the retrieval of supported file extensions, finding specific
// elements within code, minifying TypeScript source code, generating prompts
// for code suggestions, and patching existing code with documentation comments.
// Customization of the service can be achieved using provided options such as
// specifying a custom finder or model. The service encapsulates functionality
// that leverages external tools and libraries to process and enhance TypeScript
// code in various ways.
type Service struct {
	finder *Finder
	model  string
}

// Option represents a configuration function used to customize the behavior of
// a [Service]. Each Option takes a pointer to a [Service] and applies settings
// or parameters to it, allowing for flexible and composable service
// configuration. Options are typically passed to the New function to construct
// a new [Service] instance with the desired customizations.
type Option func(*Service)

// WithFinder specifies a finder to be used by the service for operations such
// as searching and position finding within code. It accepts a [*Finder] and
// returns an [Option] to configure a [Service].
func WithFinder(f *Finder) Option {
	return func(s *Service) {
		s.finder = f
	}
}

// Model configures a Service with the provided model identifier. It returns an
// Option that, when applied to a Service, sets its internal model field to the
// given string. This model identifier is typically used by the Service for
// operations that require knowledge of a specific model, such as data
// processing or interaction with external APIs that use different models.
func Model(model string) Option {
	return func(s *Service) {
		s.model = model
	}
}

// New initializes a new Service with the provided options. If no model is
// specified through the options, it uses the default model. If no Finder is
// provided, it initializes a new default Finder. It returns an initialized
// [*Service].
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

// Extensions retrieves the list of file extensions that the Service recognizes
// and is capable of processing. It returns a slice of strings, each
// representing a supported file extension.
func (svc *Service) Extensions() []string {
	return FileExtensions
}

// Find retrieves a list of strings based on the provided code slice. It
// utilizes the service's internal finder to perform the search within a default
// context. If the search is successful, it returns the results along with a nil
// error. If it fails, it returns an empty slice and an error detailing what
// went wrong.
func (svc *Service) Find(code []byte) ([]string, error) {
	return svc.finder.Find(context.Background(), code)
}

// Minify reduces the size of TypeScript code by removing unnecessary characters
// without changing its functionality and returns the minified code or an error
// if the minification fails.
func (svc *Service) Minify(code []byte) ([]byte, error) {
	args := []string{"minify", "-m", svc.model, string(code)}

	cmd := exec.Command(jotbotTSPath, args...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w:\n%s", err, out)
	}

	return out, nil
}

// Prompt invokes the generation of a prompt based on the provided input and
// returns the generated content as a string. It utilizes the underlying prompt
// generation logic to transform the input into a textual representation
// suitable for various applications.
func (svc *Service) Prompt(input generate.PromptInput) string {
	return Prompt(input)
}

// Patch applies a documentation patch to the source code at the location of a
// specified identifier. It creates or updates existing documentation based on
// the provided doc string. If the identifier cannot be located or if any errors
// occur during the process, an error is returned along with a nil byte slice.
// Otherwise, it returns the patched source code as a byte slice. The operation
// is context-aware and can be cancelled through the provided context.Context.
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

// NormalizeGeneratedComment ensures the consistency and readability of a
// generated comment by trimming excess whitespace, removing leading asterisks
// commonly used in block comments, and stripping any trailing comment
// terminators. It also filters out lines beginning with an "@" symbol, which
// are often used for annotations in documentation comments. The result is a
// clean, normalized string ready for further processing or insertion into code.
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
