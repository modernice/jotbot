package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/modernice/jotbot"
	"github.com/modernice/jotbot/find"
	"github.com/modernice/jotbot/generate"
	"github.com/modernice/jotbot/git"
	"github.com/modernice/jotbot/internal/nodes"
	"github.com/modernice/jotbot/patch"
	"github.com/modernice/jotbot/services/openai"
	"golang.org/x/exp/slog"
)

// CLI represents a command-line interface for generating missing documentation.
// It has options to configure the generation process, such as file filtering,
// committing changes to Git, and limiting the number of documentations
// generated. It also allows overriding or clearing existing documentation. The
// API key for OpenAI is provided as an option. Use New() to create a new CLI
// instance and Run() to execute it.
type CLI struct {
	Generate struct {
		Root      string   `arg:"" default:"." help:"Root directory of the repository."`
		Filter    []string `name:"filter" short:"f" env:"JOTBOT_FILTER" help:"Glob pattern(s) to filter files."`
		Commit    bool     `name:"commit" default:"true" env:"JOTBOT_COMMIT" help:"Commit changes to Git."`
		Branch    string   `default:"jotbot-patch" env:"JOTBOT_BRANCH" help:"Branch name to commit changes to."`
		Limit     int      `default:"0" env:"JOTBOT_LIMIT" help:"Limit the number of documentations to generate."`
		FileLimit int      `default:"0" env:"JOTBOT_FILE_LIMIT" help:"Limit the number of files to generate documentations for."`
		DryRun    bool     `name:"dry" default:"false" env:"JOTBOT_DRY_RUN" help:"Print the changes without applying them."`
		Model     string   `default:"gpt-3.5-turbo" env:"JOTBOT_MODEL" help:"OpenAI model to use."`
		Override  bool     `name:"override" short:"o" env:"JOTBOT_OVERRIDE" help:"Override existing documentation."`
		Clear     bool     `name:"clear" short:"c" env:"JOTBOT_CLEAR" help:"Clear existing documentation."`
	} `cmd:"" help:"Generate missing documentation."`

	APIKey  string `name:"key" env:"OPENAI_API_KEY" help:"OpenAI API key."`
	Verbose bool   `name:"verbose" short:"v" env:"JOTBOT_VERBOSE" help:"Enable verbose logging."`
}

// Run generates missing documentation. It uses OpenAI's GPT to generate
// documentation for exported functions and types in Go source files. It takes a
// Kong context as an argument and returns an error if one occurred during
// execution.
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
		openaiOpts = append(openaiOpts, openai.MinifyWith([]nodes.MinifyOptions{
			nodes.MinifyComments,
			nodes.MinifyExported,
			nodes.MinifyAll,
		}, true))
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

	docs := jotbot.New(svc, jotbot.WithLogger(logHandler))

	patch, err := docs.Generate(
		context.Background(),
		cfg.Generate.Root,
		jotbot.GenerateWith(opts...),
		jotbot.PatchWith(patch.Override(cfg.Generate.Override)),
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

// New creates a new *kong.Context and returns a pointer to it. The
// *kong.Context is used to parse command-line arguments for the "jotbot" CLI
// tool.
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
