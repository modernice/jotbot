package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/modernice/opendocs"
	"github.com/modernice/opendocs/find"
	"github.com/modernice/opendocs/generate"
	"github.com/modernice/opendocs/git"
	"github.com/modernice/opendocs/internal/nodes"
	"github.com/modernice/opendocs/patch"
	"github.com/modernice/opendocs/services/openai"
	"golang.org/x/exp/slog"
)

// CLI is a command-line interface that generates missing documentation. It uses
// OpenAI to generate documentation for functions and types that lack
// documentation. CLI takes in options such as the root directory of the
// repository, a filter for files, the number of documentations to generate, and
// more. It provides options to override or clear existing documentation, and
// can commit changes to Git.
type CLI struct {
	Generate struct {
		Root      string   `arg:"" default:"." help:"Root directory of the repository."`
		Filter    []string `name:"filter" short:"f" env:"OPENDOCS_FILTER" help:"Glob pattern(s) to filter files."`
		Commit    bool     `name:"commit" default:"true" env:"OPENDOCS_COMMIT" help:"Commit changes to Git."`
		Branch    string   `default:"opendocs-patch" env:"OPENDOCS_BRANCH" help:"Branch name to commit changes to."`
		Limit     int      `default:"0" env:"OPENDOCS_LIMIT" help:"Limit the number of documentations to generate."`
		FileLimit int      `default:"0" env:"OPENDOCS_FILE_LIMIT" help:"Limit the number of files to generate documentations for."`
		DryRun    bool     `name:"dry" default:"false" env:"OPENDOCS_DRY_RUN" help:"Just print the changes without applying them."`
		Model     string   `default:"gpt-3.5-turbo" env:"OPENDOCS_MODEL" help:"OpenAI model to use."`
		Override  bool     `name:"override" short:"o" env:"OPENDOCS_OVERRIDE" help:"Override existing documentation."`
		Clear     bool     `name:"clear" short:"c" env:"OPENDOCS_CLEAR" help:"Clear existing documentation."`
	} `cmd:"" default:"withargs" help:"Generate missing documentation."`

	APIKey  string `name:"key" env:"OPENAI_API_KEY" help:"OpenAI API key."`
	Verbose bool   `name:"verbose" short:"v" env:"OPENDOCS_VERBOSE" help:"Enable verbose logging."`
}

// Run executes the CLI application with the given configuration. It generates
// missing documentation for a repository by using OpenAI's GPT-3 model to write
// documentation for functions and types in GoDoc format. It takes into account
// various configuration options such as filtering files, committing changes to
// Git, limiting the number of documentations to generate, and more.
func (cfg *CLI) Run(ctx *kong.Context) error {
	if cfg.Generate.Root != "." {
		if !filepath.IsAbs(cfg.Generate.Root) {
			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			cfg.Generate.Root = filepath.Join(wd, cfg.Generate.Root)
		}
	}

	var level slog.Level
	if cfg.Verbose {
		level = slog.LevelDebug
	}
	logHandler := slog.HandlerOptions{Level: level}.NewTextHandler(os.Stdout)
	log := slog.New(logHandler)

	openaiOpts := []openai.Option{
		openai.WithLogger(logHandler),
		openai.Model(cfg.Generate.Model),
	}

	if cfg.Generate.Clear {
		openaiOpts = append(openaiOpts, openai.MinifyWith([]nodes.MinifyOptions{nodes.MinifyComments}, true))
	}

	svc := openai.New(cfg.APIKey, openaiOpts...)

	opts := []generate.Option{
		generate.Limit(cfg.Generate.Limit),
		generate.FileLimit(cfg.Generate.FileLimit),
		generate.Override(cfg.Generate.Override),
	}
	if len(cfg.Generate.Filter) > 0 {
		opts = append(opts, generate.FindWith(find.Glob(cfg.Generate.Filter...)))
	}

	docs := opendocs.New(svc, opendocs.WithLogger(logHandler))

	patch, err := docs.Generate(
		context.Background(),
		cfg.Generate.Root,
		opendocs.GenerateWith(opts...),
		opendocs.PatchWith(patch.Override(cfg.Generate.Override)),
	)
	if err != nil {
		return err
	}

	if cfg.Generate.DryRun {
		patchResult, err := patch.DryRun()
		if err != nil {
			return fmt.Errorf("dry run: %w", err)
		}
		printDryRun(patchResult)
		return nil
	}

	if cfg.Generate.Commit {
		grepo := git.Repo(cfg.Generate.Root, git.WithLogger(logHandler))
		if err := grepo.Commit(patch, git.Branch(cfg.Generate.Branch)); err != nil {
			return fmt.Errorf("commit patch: %w", err)
		}

		log.Info("Done.")

		return nil
	}

	if err := patch.Apply(cfg.Generate.Root); err != nil {
		return fmt.Errorf("apply patch: %w", err)
	}

	return nil
}

// New returns a new *kong.Context for parsing command line arguments. It
// creates a CLI object, which can be used to generate or apply documentation
// patches.
func New() *kong.Context {
	if len(os.Args) < 1 {
		os.Args = append(os.Args, "generate")
	}

	var cfg CLI
	return kong.Parse(&cfg)
}

func printDryRun(result map[string][]byte) {
	for path, content := range result {
		log.Printf("\n%s:\n\n%s", path, content)
	}
}
