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

// Language is an interface that combines the functionality of patch.Language
// and generate.Language, providing methods for file extension handling and
// finding identifiers in source code. It extends the capabilities of patching
// and generating code by allowing customization for different programming
// languages.
type Language interface {
	patch.Language
	generate.Language

	// Extensions returns a slice of file extensions supported by the configured
	// languages in the Language interface.
	Extensions() []string

	// Find searches the provided byte slice for language-specific identifiers and
	// returns a slice of found identifiers and an error if any occurs.
	Find([]byte) ([]string, error)
}

// JotBot is a tool for finding, generating, and patching code snippets in
// multiple programming languages. It supports customizable language
// configurations and can be extended with additional languages. JotBot can
// filter findings based on regular expressions, generate new code snippets
// using a provided generate.Service, and apply or dry-run patches to modify the
// source code.
type JotBot struct {
	root          string
	filters       []*regexp.Regexp
	fs            afero.Fs
	languages     map[string]Language
	extToLanguage map[string]string
	log           *slog.Logger
}

// Option is a function that configures a JotBot instance. It takes a JotBot
// pointer as an argument and modifies its properties according to the desired
// configuration. Common options include WithLanguage, which associates a
// language with JotBot, and WithLogger, which sets up a logger for the JotBot
// instance.
type Option func(*JotBot)

// Finding represents an identifier found within a file and its associated
// language.
type Finding struct {
	Identifier string
	File       string
	Language   string
}

// String returns a formatted string representation of the Finding, containing
// the file name and identifier separated by an '@' symbol.
func (f Finding) String() string {
	return fmt.Sprintf("%s@%s", f.File, f.Identifier)
}

// Patch represents a set of changes to be applied to source code files. It
// provides methods for applying the changes directly or performing a dry run to
// preview the resulting files without modifying them. A Patch is created by
// generating code from a list of findings using a generate.Service and various
// generate.Option values.
type Patch struct {
	*patch.Patch

	getLanguage func(string) (Language, error)
}

// WithLanguage configures a JotBot instance to use the specified Language
// implementation with the given name. The Language implementation is used for
// finding identifiers, generating code, and patching files with the
// corresponding file extension.
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

// Match adds the provided regular expression filters to the JotBot instance,
// which are then used to filter identifiers found in files during the Find
// operation.
func Match(filters ...*regexp.Regexp) Option {
	return func(bot *JotBot) {
		bot.filters = append(bot.filters, filters...)
	}
}

// New creates a new JotBot instance with the specified root directory and
// optional configurations. The returned JotBot can be used to find and generate
// code snippets based on configured languages and filters.
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

// ConfigureLanguage configures a Language for the given name and adds the
// language's file extensions to the mapping of extensions to languages.
func (bot *JotBot) ConfigureLanguage(name string, lang Language) {
	bot.languages[name] = lang
	for _, ext := range lang.Extensions() {
		bot.extToLanguage[ext] = name
	}
}

// Extensions returns a slice of file extensions for which a language is
// configured in the JotBot instance.
func (bot *JotBot) Extensions() []string {
	return maps.Keys(bot.extToLanguage)
}

// Find searches for files in the root directory of a JotBot instance, filters
// them by configured languages, and returns a slice of findings. Each finding
// includes the identifier, file path, and language name associated with the
// matched file. The search can be further customized using context and find
// options.
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

// Generate creates a Patch by generating code for the given Findings using the
// provided generate.Service and options. It returns an error if there is any
// issue during code generation.
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

// Apply applies the patch to the files in the provided root directory. It
// returns an error if there is an issue applying the patch or accessing the
// file system.
func (p *Patch) Apply(ctx context.Context, root string) error {
	return p.Patch.Apply(ctx, afero.NewBasePathFs(afero.NewOsFs(), root), func(s string) (patch.Language, error) {
		return p.getLanguage(s)
	})
}

// DryRun applies the patch in a simulated environment without making actual
// changes to the files in the specified root directory, returning a map of
// filenames to their updated contents after applying the patch. It is useful
// for previewing changes before applying them.
func (p *Patch) DryRun(ctx context.Context, root string) (map[string][]byte, error) {
	return p.Patch.DryRun(ctx, afero.NewBasePathFs(afero.NewOsFs(), root), func(s string) (patch.Language, error) {
		return p.getLanguage(s)
	})
}
