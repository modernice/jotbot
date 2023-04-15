package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	opendocs "github.com/modernice/opendocs/go"
	"github.com/modernice/opendocs/go/git"
	"github.com/modernice/opendocs/go/services/openai"
)

type CLI struct {
	Generate struct {
		Root   string `arg:"" default:"." help:"Root directory of the repository."`
		Branch string `default:"opendocs-patch" help:"Branch name to commit changes to. (set to empty string to disable committing)"`
	} `cmd:"" default:"withargs" help:"Generate missing documentation."`

	APIKey string `name:"key" env:"OPENAI_API_KEY" help:"OpenAI API key."`
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

	svc := openai.New(cfg.APIKey)
	repo := opendocs.Repo(cfg.Generate.Root)

	result, err := repo.Generate(context.Background(), svc)
	if err != nil {
		return err
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
