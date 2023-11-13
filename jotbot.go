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

// Language represents a programming language with the ability to describe
// itself, such as providing its file extensions and identifying code patterns
// within its syntax. It integrates with patching and code generation systems,
// allowing for modifications and enhancements of code written in the language
// it represents. It also offers methods to locate identifiers within a body of
// text, aiding in various analysis and automation tasks.
type Language interface {
	patch.Language
	generate.Language

	// Extensions reports the file extensions associated with a language. These
	// extensions typically do not include the leading dot.
	Extensions() []string

	// Find locates and returns all identifiers within a given byte slice according
	// to the rules of the implementing language, or an error if the search cannot
	// be completed. It returns a slice of strings representing the found
	// identifiers and an error, if any occurred during the search process.
	Find([]byte) ([]string, error)
}

// JotBot orchestrates the process of searching, analyzing, and transforming
// code across multiple programming languages within a specified directory
// structure. It leverages configurable language-specific behaviors to locate
// identifiers, generate updates, and apply patches based on findings. JotBot
// allows for flexible extension by registering different languages and can be
// customized through various options to match specific project needs. It
// provides methods to configure languages, find code patterns, generate code
// changes, and apply those changes as patches. Additionally, it supports
// logging for traceability of operations.
type JotBot struct {
	root          string
	filters       []*regexp.Regexp
	fs            afero.Fs
	languages     map[string]Language
	extToLanguage map[string]string
	log           *slog.Logger
}

// Option configures a [*JotBot] instance with custom settings, such as
// specifying languages to recognize, logging behavior, or file matching
// patterns. It is used when creating a new [*JotBot] or adjusting its
// configuration at runtime. Each Option is a function that applies a specific
// configuration to the [*JotBot].
type Option func(*JotBot)

// Finding represents a discovered identifier within a particular file and
// programming language. It holds the unique identifier found, the file in which
// it was found, and the language of that file. The Finding type provides a way
// to reference specific code elements that have been identified during analysis
// or processing across various files and languages.
type Finding struct {
	Identifier string
	File       string
	Language   string
}

// String provides a human-readable representation of a Finding, combining its
// File and Identifier properties in a formatted string.
func (f Finding) String() string {
	return fmt.Sprintf("%s@%s", f.File, f.Identifier)
}

// Patch applies modifications across a collection of files within a specified
// root directory. It leverages a provided callback to determine the
// language-specific behaviors required for each file based on its extension,
// ensuring that patches are applied correctly according to language rules and
// syntax. Patch also supports a dry run mode that simulates the patch
// application, returning a map of the proposed changes without altering the
// original files, allowing for pre-application review and validation.
type Patch struct {
	*patch.Patch

	getLanguage func(string) (Language, error)
}

// WithLanguage configures a JotBot instance to use a specified language with an
// associated name. It allows the JotBot to recognize and handle files that
// pertain to the given language during its operations. This configuration is
// done through an option that can be passed to the JotBot constructor.
func WithLanguage(name string, lang Language) Option {
	return func(bot *JotBot) {
		bot.ConfigureLanguage(name, lang)
	}
}

// WithLogger configures a JotBot instance to use the provided slog.Handler for
// logging operations. It returns an Option which, when applied, sets up the
// internal logger of the JotBot with the specified handler.
func WithLogger(h slog.Handler) Option {
	return func(bot *JotBot) {
		bot.log = slog.New(h)
	}
}

// Match configures a JotBot with custom filters for identifying relevant
// findings. It accepts a variable number of regular expressions that are used
// to filter the search results when finding identifiers within files. The
// provided filters are appended to any existing filters the JotBot may have.
func Match(filters ...*regexp.Regexp) Option {
	return func(bot *JotBot) {
		bot.filters = append(bot.filters, filters...)
	}
}

// New initializes and returns a new instance of JotBot configured with the
// provided root directory and options.
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

// ConfigureLanguage associates a given language with its name and file
// extensions within the JotBot instance. It enables the JotBot to recognize
// files of this language by their extensions when performing operations such as
// finding identifiers or generating patches.
func (bot *JotBot) ConfigureLanguage(name string, lang Language) {
	bot.languages[name] = lang
	for _, ext := range lang.Extensions() {
		bot.extToLanguage[ext] = name
	}
}

// Extensions returns a slice of all file extensions that are associated with
// configured languages within the JotBot instance. These extensions can be used
// to filter files for processing based on the languages that the JotBot is
// capable of handling.
func (bot *JotBot) Extensions() []string {
	return maps.Keys(bot.extToLanguage)
}

// Find performs a search for identifiers within the files of a repository based
// on the configured languages and file extensions. It accepts a context and
// variadic find options to customize the search behavior. The function returns
// a slice of Findings, which contain the identifier, file, and language of each
// found item, or an error if the search could not be completed. The Findings
// are sorted by file and then by identifier. If filters are configured, only
// findings matching those filters are included in the results.
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

	slices.SortFunc(out, func(a, b Finding) int {
		if a.File < b.File || (a.File == b.File && a.Identifier < b.Identifier) {
			return -1
		}
		return 1
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

// Generate creates a patch based on the provided findings and generation
// service, applying additional options if specified. It processes each finding
// to prepare the input for the generator, then invokes the generator to create
// file patches. On success, it returns a [*Patch] that encapsulates the
// generated patches along with any errors that occurred during generation. If
// an error is encountered during the preparation of inputs or generation
// process, it returns an error detailing the failure.
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

// Apply applies the patch to the files within the specified root directory. It
// takes a context and a string representing the root directory path as
// arguments and returns an error if the patch cannot be applied. The operation
// respects the context cancellation and will abort if the context is canceled.
func (p *Patch) Apply(ctx context.Context, root string) error {
	return p.Patch.Apply(ctx, afero.NewBasePathFs(afero.NewOsFs(), root), func(s string) (patch.Language, error) {
		return p.getLanguage(s)
	})
}

// DryRun simulates the application of the patch to the given root directory
// without making actual changes, and returns a map of file paths to their new
// content as it would appear after applying the patch. It accepts a context for
// cancellation and deadline control, and requires the root directory as an
// argument. The function returns an error if any issues occur during the dry
// run process.
func (p *Patch) DryRun(ctx context.Context, root string) (map[string][]byte, error) {
	return p.Patch.DryRun(ctx, afero.NewBasePathFs(afero.NewOsFs(), root), func(s string) (patch.Language, error) {
		return p.getLanguage(s)
	})
}
