package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/alecthomas/kong"
	"github.com/modernice/jotbot"
	"github.com/modernice/jotbot/find"
	"github.com/modernice/jotbot/generate"
	"github.com/modernice/jotbot/git"
	"github.com/modernice/jotbot/internal"
	"github.com/modernice/jotbot/langs/golang"
	"github.com/modernice/jotbot/langs/ts"
	"github.com/modernice/jotbot/services/openai"
	"golang.org/x/exp/slog"
)

const internalDirectoriesGlob = "**/internal/**/*.go"

type Config struct {
	Generate struct {
		Root            string      `arg:"" default:"." help:"Root directory of the repository."`
		Include         []string    `name:"include" short:"i" env:"JOTBOT_INCLUDE" help:"Glob pattern(s) to include files."`
		Exclude         []string    `name:"exclude" short:"e" env:"JOTBOT_EXCLUDE" help:"Glob pattern(s) to exclude files."`
		ExcludeInternal bool        `name:"exclude-internal" short:"E" default:"true" env:"JOTBOT_EXCLUDE_INTERNAL" help:"Exclude 'internal' directories (Go-specific)"`
		Match           []string    `name:"match" short:"m" env:"JOTBOT_MATCH" help:"Regular expression(s) to match identifiers."`
		Symbols         []ts.Symbol `name:"symbol" short:"s" env:"JOTBOT_SYMBOLS" help:"Symbol(s) to search for in code (TS/JS-specific)."`
		Branch          string      `name:"branch" env:"JOTBOT_BRANCH" help:"Branch name to commit changes to. Leave empty to not commit changes."`
		Limit           int         `name:"limit" default:"0" env:"JOTBOT_LIMIT" help:"Limit the number of files to generate documentation for."`
		DryRun          bool        `name:"dry" default:"false" env:"JOTBOT_DRY_RUN" help:"Print the changes without applying them."`
		Model           string      `name:"model" default:"gpt-3.5-turbo" env:"JOTBOT_MODEL" help:"OpenAI model used to generate documentation."`
		MaxTokens       int         `name:"maxTokens" default:"${maxTokens=512}" env:"JOTBOT_MAX_TOKENS" help:"Maximum number of tokens to generate for a single documentation."`
		Parallel        int         `name:"parallel" short:"p" default:"${parallel=4}" env:"JOTBOT_PARALLEL" help:"Number of files to handle concurrently."`
		Workers         int         `name:"workers" default:"${workers=2}" env:"JOTBOT_WORKERS" help:"Number of workers to use per file."`
		Override        bool        `name:"override" short:"o" env:"JOTBOT_OVERRIDE" help:"Override existing documentation."`
	} `cmd:"" help:"Generate missing documentation."`

	APIKey  string `name:"key" env:"OPENAI_API_KEY" help:"OpenAI API key."`
	Verbose bool   `name:"verbose" short:"v" env:"JOTBOT_VERBOSE" help:"Enable verbose logging."`
}

func (cfg *Config) Run(kctx *kong.Context) error {
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
	logHandler := internal.PrettyLogger(slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Key = ""
			}
			return a
		},
	}.NewTextHandler(os.Stdout))
	logger := slog.New(logHandler)

	goFinder := golang.NewFinder(golang.FindTests(false))
	gosvc, err := golang.New(golang.WithFinder(goFinder), golang.Model(cfg.Generate.Model))
	if err != nil {
		return fmt.Errorf("create Go language service: %w", err)
	}

	tsFinder := ts.NewFinder(
		ts.Symbols(cfg.Generate.Symbols...),
		ts.IncludeDocumented(cfg.Generate.Override),
	)
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

	if cfg.Generate.ExcludeInternal {
		cfg.Generate.Exclude = append(cfg.Generate.Exclude, internalDirectoriesGlob)
	}

	start := time.Now()

	findings, err := bot.Find(
		ctx,
		find.Include(cfg.Generate.Include...),
		find.Exclude(cfg.Generate.Exclude...),
	)
	if err != nil {
		return fmt.Errorf("find uncommented code: %w", err)
	}

	patch, err := bot.Generate(
		ctx,
		findings,
		oai,
		generate.Limit(cfg.Generate.Limit),
		generate.Workers(cfg.Generate.Parallel, cfg.Generate.Workers),
	)
	if err != nil {
		return fmt.Errorf("generate documentation: %w", err)
	}

	if cfg.Generate.DryRun {
		patched, err := patch.DryRun(ctx, cfg.Generate.Root)
		if err != nil {
			return fmt.Errorf("dry run: %w", err)
		}

		for file, code := range patched {
			fmt.Printf("Patched %q:\n\n%s\n", file, code)
		}

		took := time.Since(start)
		logger.Info(fmt.Sprintf("Done in %s.", took))

		return nil
	}

	if cfg.Generate.Branch == "" {
		return patch.Apply(ctx, cfg.Generate.Root)
	}

	repo := git.Repo(cfg.Generate.Root, git.WithLogger(logHandler))
	if err := repo.Commit(ctx, patch, git.Branch(cfg.Generate.Branch)); err != nil {
		return fmt.Errorf("commit patch: %w", err)
	}

	took := time.Since(start)
	logger.Info(fmt.Sprintf("Done in %s.", took))

	return nil
}

func New() *kong.Context {
	if len(os.Args) < 1 {
		os.Args = append(os.Args, "generate")
	}
	var cfg Config
	return kong.Parse(&cfg, kong.Vars{
		"maxTokens": strconv.Itoa(openai.DefaultMaxTokens),
		"parallel":  strconv.Itoa(generate.DefaultFileWorkers),
		"workers":   strconv.Itoa(generate.DefaultSymbolWorkers),
	})
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
