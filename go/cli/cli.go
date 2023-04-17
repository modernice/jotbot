package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	opendocs "github.com/modernice/opendocs/go"
	"github.com/modernice/opendocs/go/generate"
	"github.com/modernice/opendocs/go/git"
	"github.com/modernice/opendocs/go/services/openai"
	"golang.org/x/exp/slog"
)

type CLI struct {
	Generate struct {
		Root   string `arg:"" default:"." help:"Root directory of the repository."`
		Branch string `default:"opendocs-patch" env:"OPENDOCS_BRANCH" help:"Branch name to commit changes to. (set to empty string to disable committing)"`
		Limit  int    `default:"0" env:"OPENDOCS_LIMIT" help:"Limit the number of documentations to generate."`
		DryRun bool   `name:"dry" default:"false" env:"OPENDOCS_DRY_RUN" help:"Just print the changes without applying them."`
	} `cmd:"" default:"withargs" help:"Generate missing documentation."`

	APIKey  string `name:"key" env:"OPENAI_API_KEY" help:"OpenAI API key."`
	Verbose bool   `name:"verbose" short:"v" env:"OPENDOCS_VERBOSE" help:"Enable verbose logging."`
}

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

	svc := openai.New(cfg.APIKey, openai.WithLogger(logHandler))
	repo := opendocs.Repo(cfg.Generate.Root)

	opts := []generate.Option{generate.Limit(cfg.Generate.Limit)}

	opts = append(opts, generate.WithLogger(logHandler))

	result, err := repo.Generate(context.Background(), svc, opts...)
	if err != nil {
		return err
	}

	if cfg.Generate.DryRun {
		patch := result.Patch()
		patchResult, err := patch.DryRun()
		if err != nil {
			return fmt.Errorf("dry run: %w", err)
		}
		printDryRun(patchResult)
		return nil
	}

	if cfg.Generate.Branch != "" {
		if _, err := result.Commit(cfg.Generate.Root, git.Branch(cfg.Generate.Branch)); err != nil {
			return fmt.Errorf("commit changes: %w", err)
		}
		return nil
	}

	if err := result.Patch().Apply(cfg.Generate.Root); err != nil {
		return fmt.Errorf("apply patch: %w", err)
	}

	return nil
}

func New() *kong.Context {
	if len(os.Args) < 1 {
		os.Args = append(os.Args, "generate")
	}

	var cfg CLI
	return kong.Parse(&cfg)
}

func printDryRun(result map[string][]byte) {
	for path, content := range result {
		log.Printf("%s:\n\n%s", path, content)
	}
}
