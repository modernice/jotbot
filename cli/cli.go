package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"

	"github.com/alecthomas/kong"
	"github.com/modernice/jotbot"
	"github.com/modernice/jotbot/find"
	"github.com/modernice/jotbot/generate"
	"github.com/modernice/jotbot/langs/golang"
	"github.com/modernice/jotbot/langs/ts"
	"github.com/modernice/jotbot/services/openai"
	"golang.org/x/exp/slog"
)

type CLI struct {
	Generate struct {
		Root      string      `arg:"" default:"." help:"Root directory of the repository."`
		Include   []string    `name:"include" short:"i" env:"JOTBOT_INCLUDE" help:"Glob pattern(s) to include files."`
		Exclude   []string    `name:"exclude" short:"e" env:"JOTBOT_EXCLUDE" help:"Glob pattern(s) to exclude files."`
		Match     []string    `name:"match" short:"m" env:"JOTBOT_MATCH" help:"Regular expression(s) to match identifiers."`
		Symbols   []ts.Symbol `name:"symbol" short:"s" env:"JOTBOT_SYMBOLS" help:"Symbol(s) to search for in code (only for TS/JS)."`
		Branch    string      `env:"JOTBOT_BRANCH" help:"Branch name to commit changes to. Leave empty to not commit changes."`
		Limit     int         `default:"0" env:"JOTBOT_LIMIT" help:"Limit the number of files to generate documentation for."`
		DryRun    bool        `name:"dry" default:"false" env:"JOTBOT_DRY_RUN" help:"Print the changes without applying them."`
		Model     string      `default:"gpt-3.5-turbo" env:"JOTBOT_MODEL" help:"OpenAI model to use."`
		MaxTokens int         `default:"512" env:"JOTBOT_MAX_TOKENS" help:"Maximum number of tokens to generate for a single documentation."`
		Workers   int         `default:"1" env:"JOTBOT_WORKERS" help:"Number of workers to use per file."`
		// Override bool     `name:"override" short:"o" env:"JOTBOT_OVERRIDE" help:"Override existing documentation."`
		// Clear    bool     `name:"clear" short:"c" env:"JOTBOT_CLEAR" help:"Clear existing documentation."`
	} `cmd:"" help:"Generate missing documentation."`

	APIKey  string `name:"key" env:"OPENAI_API_KEY" help:"OpenAI API key."`
	Verbose bool   `name:"verbose" short:"v" env:"JOTBOT_VERBOSE" help:"Enable verbose logging."`
}

func (cfg *CLI) Run(kctx *kong.Context) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

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

	goFinder := golang.NewFinder(golang.FindTests(false))
	gosvc, err := golang.New(golang.WithFinder(goFinder), golang.Model(cfg.Generate.Model))
	if err != nil {
		return fmt.Errorf("create Go language service: %w", err)
	}

	tsFinder := ts.NewFinder(ts.Symbols(cfg.Generate.Symbols...))
	tssvc := ts.New(ts.WithFinder(tsFinder))

	matchers, err := parseMatchers(cfg.Generate.Match)
	if err != nil {
		return fmt.Errorf("parse matchers: %w", err)
	}

	bot := jotbot.New(
		cfg.Generate.Root,
		jotbot.WithLogger(logHandler),
		jotbot.WithLanguage("go", gosvc),
		jotbot.WithLanguage("ts", tssvc),
		jotbot.Match(matchers...),
	)

	openaiOpts := []openai.Option{
		openai.Model(cfg.Generate.Model),
		openai.MaxTokens(cfg.Generate.MaxTokens),
		openai.WithLogger(logHandler),
	}

	oai, err := openai.New(cfg.APIKey, openaiOpts...)
	if err != nil {
		return fmt.Errorf("create OpenAI service: %w", err)
	}

	findings, err := bot.Find(
		ctx,
		find.Include(cfg.Generate.Include...),
		find.Exclude(cfg.Generate.Exclude...),
	)
	if err != nil {
		return fmt.Errorf("find uncommented code: %w", err)
	}

	patch, err := bot.Generate(ctx, findings, oai, generate.Limit(cfg.Generate.Limit), generate.Workers(cfg.Generate.Workers))
	if err != nil {
		return fmt.Errorf("generate documentation: %w", err)
	}

	patched, err := patch.DryRun(ctx, cfg.Generate.Root)
	if err != nil {
		return fmt.Errorf("dry run: %w", err)
	}

	for file, code := range patched {
		log.Printf("Patched %q:\n\n%s\n", file, code)
	}

	// if cfg.Generate.Clear {
	// 	openaiOpts = append(openaiOpts, openai.MinifyWith([]nodes.MinifyOptions{
	// 		nodes.MinifyComments,
	// 		nodes.MinifyExported,
	// 		nodes.MinifyAll,
	// 	}, true))
	// }

	// svc := openai.New(cfg.APIKey, openaiOpts...)

	// opts := []generate.Option{
	// 	generate.Limit(cfg.Generate.Limit),
	// 	generate.FileLimit(cfg.Generate.FileLimit),
	// 	generate.Override(cfg.Generate.Override),
	// }
	// if len(cfg.Generate.Filter) > 0 {
	// 	// opts = append(opts, generate.FindWith(golang.Glob(cfg.Generate.Filter...)))
	// }

	// docs := jotbot.New(svc, jotbot.WithLogger(logHandler))

	// patch, err := docs.Generate(
	// 	context.Background(),
	// 	cfg.Generate.Root,
	// 	jotbot.GenerateWith(opts...),
	// 	jotbot.PatchWith(golang.Override(cfg.Generate.Override)),
	// )
	// if err != nil {
	// 	return err
	// }

	// if cfg.Generate.DryRun {
	// 	patchResult, err := patch.DryRun()
	// 	if err != nil {
	// 		return fmt.Errorf("dry run: %w", err)
	// 	}
	// 	printDryRun(patchResult)
	// 	return nil
	// }

	// if cfg.Generate.Commit {
	// 	grepo := git.Repo(cfg.Generate.Root, git.WithLogger(logHandler))
	// 	if err := grepo.Commit(patch, git.Branch(cfg.Generate.Branch)); err != nil {
	// 		return fmt.Errorf("commit patch: %w", err)
	// 	}

	// 	log.Info("Done.")

	// 	return nil
	// }

	// if err := patch.Apply(cfg.Generate.Root); err != nil {
	// 	return fmt.Errorf("apply patch: %w", err)
	// }

	return nil
}

func New() *kong.Context {
	if len(os.Args) < 1 {
		os.Args = append(os.Args, "generate")
	}

	var cfg CLI
	return kong.Parse(&cfg)
}

func parseMatchers(raw []string) ([]*regexp.Regexp, error) {
	out := make([]*regexp.Regexp, len(raw))
	var err error
	for i, r := range raw {
		if out[i], err = regexp.Compile(r); err != nil {
			return out, fmt.Errorf("compile regular expression %q: %w", r, err)
		}
	}
	return out, nil
}
