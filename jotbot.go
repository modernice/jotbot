package jotbot

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/modernice/jotbot/find"
	"github.com/modernice/jotbot/generate"
	"github.com/modernice/jotbot/internal"
	"github.com/modernice/jotbot/internal/slice"
	"github.com/modernice/jotbot/patch"
	"github.com/spf13/afero"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
)

// Language is an interface that combines the patch.Language and
// generate.Language interfaces, adding methods to retrieve supported file
// extensions and find identifiers within source code. It is used by JotBot to
// work with different programming languages, allowing for language-specific
// operations such as finding identifiers and generating patches.
type Language interface {
	patch.Language
	generate.Language

	// Extensions returns a slice of file extensions that are supported by the
	// Language. Each extension is associated with a specific language
	// implementation for code generation and patching purposes.
	Extensions() []string

	// Find searches the provided byte slice for identifiers in the language and
	// returns a slice of found identifiers along with any error encountered during
	// the search.
	Find([]byte) ([]string, error)
}

// JotBot is a utility for finding and generating code snippets in various
// programming languages within a project. It supports configurable languages,
// filters, and file systems, and offers patching capabilities for applying
// generated changes.
type JotBot struct {
	root          string
	filters       []*regexp.Regexp
	fs            afero.Fs
	languages     map[string]Language
	extToLanguage map[string]string
	log           *slog.Logger
}

// Option is a function that takes a pointer to a JotBot and modifies its
// configuration. It can be used to set custom languages, loggers, or filters
// for the JotBot instance.
type Option func(*JotBot)

// Finding represents a single code identifier found within a file with its
// associated language.
type Finding struct {
	Identifier string
	File       string
	Language   string
}

// String returns a formatted string representation of the Finding, including
// the file path and identifier.
func (f Finding) String() string {
	return fmt.Sprintf("%s@%s", f.File, f.Identifier)
}

// Patch is a wrapper around a patch.Patch that provides additional
// functionality for applying and dry-running patches using Language
// implementations. It is used to generate, apply, and preview changes to source
// code based on the provided findings.
type Patch struct {
	*patch.Patch

	getLanguage func(string) (Language, error)
}

// WithLanguage configures a JotBot instance to use the specified language for
// processing files with the given name. It associates the language with its
// supported file extensions.
func WithLanguage(name string, lang Language) Option {
	return func(bot *JotBot) {
		bot.ConfigureLanguage(name, lang)
	}
}

// WithLogger configures a JotBot to use the provided slog.Handler for logging.
func WithLogger(h slog.Handler) Option {
	return func(bot *JotBot) {
		bot.log = slog.New(h)
	}
}

// Match appends the provided regular expression filters to the JotBot's
// filters. These filters are used to determine which findings should be
// included in the final result when searching for identifiers in files.
func Match(filters ...*regexp.Regexp) Option {
	return func(bot *JotBot) {
		bot.filters = append(bot.filters, filters...)
	}
}

// New creates a new JotBot instance with the given root directory and options.
// The returned JotBot is configured with the provided languages, filters, and
// logger.
func New(root string, opts ...Option) *JotBot {
	bot := &JotBot{
		root:          root,
		fs:            afero.NewBasePathFs(afero.NewOsFs(), root),
		languages:     make(map[string]Language),
		extToLanguage: make(map[string]string),
	}

	for _, opt := range opts {
		opt(bot)
	}

	if bot.log == nil {
		bot.log = internal.NopLogger()
	}

	return bot
}

// ConfigureLanguage configures a Language for the given name, associating it
// with the provided extensions. The Language is used by JotBot to find and
// generate code for identifiers in files with the associated extensions.
func (bot *JotBot) ConfigureLanguage(name string, lang Language) {
	bot.languages[name] = lang
	for _, ext := range lang.Extensions() {
		bot.extToLanguage[ext] = name
	}
}

// Extensions returns a slice of file extensions that are associated with
// configured languages in the JotBot instance.
func (bot *JotBot) Extensions() []string {
	return maps.Keys(bot.extToLanguage)
}

// Find searches the JotBot's root directory for files matching the provided
// options, extracts identifiers from those files using the configured
// languages, and returns a slice of Findings containing the extracted
// identifiers, file paths, and language names. It filters the extracted
// identifiers based on the JotBot's configured filters. If any error occurs
// during the process, it returns the error.
func (bot *JotBot) Find(ctx context.Context, opts ...find.Option) ([]Finding, error) {
	bot.log.Info(fmt.Sprintf("Searching for files in %s ...", bot.root))

	opts = append([]find.Option{find.Extensions(bot.Extensions()...)}, opts...)

	repo := os.DirFS(bot.root)
	files, err := find.Files(ctx, repo, opts...)
	if err != nil {
		return nil, err
	}

	var out []Finding
	for _, file := range files {
		ext := filepath.Ext(file)
		langName, ok := bot.extToLanguage[ext]
		if !ok {
			bot.log.Warn(fmt.Sprintf("no language configured for file extension %q", ext))
			continue
		}

		lang, err := bot.languageForExtension(ext)
		if err != nil {
			bot.log.Warn(err.Error())
			continue
		}

		path := filepath.Clean(filepath.Join(bot.root, file))

		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read file %s: %w", path, err)
		}

		findings, err := lang.Find(b)
		if err != nil {
			return nil, fmt.Errorf("find in %s: %w", path, err)
		}

		findings = bot.filterFindings(findings)

		out = append(out, slice.Map(findings, func(id string) Finding {
			return Finding{
				Identifier: id,
				File:       file,
				Language:   langName,
			}
		})...)
	}

	slices.SortFunc(out, func(a, b Finding) bool {
		return a.File < b.File || (a.File == b.File && a.Identifier < b.Identifier)
	})

	if len(out) == 0 {
		bot.log.Info("No identifiers found in files.")
	} else {
		bot.log.Info(fmt.Sprintf("Found %d identifiers:", len(out)))
	}

	for _, finding := range out {
		bot.log.Log(ctx, internal.LogLevelNaked, fmt.Sprintf("- %s", finding))
	}

	return out, nil
}

func (bot *JotBot) filterFindings(findings []string) []string {
	if len(bot.filters) == 0 {
		return findings
	}
	return slice.Filter(findings, func(id string) bool {
		for _, filter := range bot.filters {
			if filter.MatchString(id) {
				return true
			}
		}
		return false
	})
}

func (bot *JotBot) languageForExtension(ext string) (Language, error) {
	if name, ok := bot.extToLanguage[ext]; ok {
		if lang, ok := bot.languages[name]; ok {
			return lang, nil
		}
	}
	return nil, fmt.Errorf("no language configured for file extension %q", ext)
}

// Generate creates a Patch by generating code snippets for the given Findings
// using the provided generate.Service and options. It returns a Patch that can
// be applied to the project or used in a dry run.
func (bot *JotBot) Generate(ctx context.Context, findings []Finding, svc generate.Service, opts ...generate.Option) (*Patch, error) {
	baseOpts := []generate.Option{generate.WithLogger(bot.log.Handler())}
	for name, lang := range bot.languages {
		baseOpts = append(baseOpts, generate.WithLanguage(name, lang))
	}
	opts = append(baseOpts, opts...)

	g := generate.New(svc, opts...)

	files := make(map[string][]generate.Input)
	for _, finding := range findings {
		input, err := bot.makeInput(ctx, finding)
		if err != nil {
			return nil, fmt.Errorf("prepare generator input for %q: %w", finding, err)
		}
		files[finding.File] = append(files[finding.File], input)
	}

	generated, errs, err := g.Files(ctx, files)
	if err != nil {
		return nil, err
	}

	return &Patch{
		Patch:       patch.New(generated, patch.WithErrors(errs), patch.WithLogger(bot.log.Handler())),
		getLanguage: bot.languageForExtension,
	}, nil
}

func (bot *JotBot) makeInput(ctx context.Context, finding Finding) (generate.Input, error) {
	f, err := bot.fs.Open(finding.File)
	if err != nil {
		return generate.Input{}, err
	}
	defer f.Close()

	code, err := io.ReadAll(f)
	if err != nil {
		return generate.Input{}, err
	}

	input := generate.Input{
		Code:       code,
		Language:   finding.Language,
		Identifier: finding.Identifier,
	}

	return input, nil
}

// Apply applies the patch to the filesystem rooted at the given root path,
// using the configured languages for applying changes. The operation can be
// canceled through the provided context. It returns an error if any issues are
// encountered during the application process.
func (p *Patch) Apply(ctx context.Context, root string) error {
	return p.Patch.Apply(ctx, afero.NewBasePathFs(afero.NewOsFs(), root), func(s string) (patch.Language, error) {
		return p.getLanguage(s)
	})
}

// DryRun applies the patch to a virtual filesystem rooted at the provided root
// path, returning a map of file paths to their new contents without modifying
// the actual filesystem. The operation can be canceled through the provided
// context.
func (p *Patch) DryRun(ctx context.Context, root string) (map[string][]byte, error) {
	return p.Patch.DryRun(ctx, afero.NewBasePathFs(afero.NewOsFs(), root), func(s string) (patch.Language, error) {
		return p.getLanguage(s)
	})
}
