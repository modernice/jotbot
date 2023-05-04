package jotbot

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/modernice/jotbot/find"
	"github.com/modernice/jotbot/generate"
	"github.com/modernice/jotbot/internal"
	"github.com/modernice/jotbot/internal/slice"
	"github.com/modernice/jotbot/langs/golang"
	"github.com/modernice/jotbot/patch"
	"github.com/spf13/afero"
	"golang.org/x/exp/slog"
)

var (
	DefaultLanguages = map[string]Language{
		".go": golang.New(golang.NewFinder()),
	}
)

type Language interface {
	patch.Language

	Find([]byte) ([]find.Finding, error)
}

type JotBot struct {
	root      string
	fs        afero.Fs
	languages map[string]Language
	log       *slog.Logger
}

type Option func(*JotBot)

type Finding struct {
	find.Finding

	File string
}

type Patch struct {
	*patch.Patch

	getLanguage func(string) (Language, error)
}

func New(root string, opts ...Option) *JotBot {
	bot := &JotBot{
		root: root,
		fs:   afero.NewBasePathFs(afero.NewOsFs(), root),
	}
	for _, opt := range opts {
		opt(bot)
	}
	if bot.languages == nil {
		bot.languages = DefaultLanguages
	}
	if bot.log == nil {
		bot.log = internal.NopLogger()
	}
	return bot
}

func (bot *JotBot) ConfigureLanguage(ext string, svc Language) {
	bot.languages[ext] = svc
}

func (bot *JotBot) Find(ctx context.Context, opts ...find.Option) ([]Finding, error) {
	repo := os.DirFS(bot.root)
	files, err := find.Files(ctx, repo, opts...)
	if err != nil {
		return nil, err
	}

	var out []Finding
	for _, file := range files {
		ext := filepath.Ext(file)
		lang, err := bot.language(ext)
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

		out = append(out, slice.Map(findings, func(f find.Finding) Finding {
			return Finding{f, file}
		})...)
	}

	return out, nil
}

func (bot *JotBot) language(ext string) (Language, error) {
	if lang, ok := bot.languages[ext]; ok {
		return lang, nil
	}
	return nil, fmt.Errorf("no language service defined for file extension %q", ext)
}

func (bot *JotBot) Generate(ctx context.Context, findings []Finding, svc generate.Service, opts ...generate.Option) (*Patch, error) {
	g := generate.New(svc)

	files := make(map[string][]generate.Input)
	for _, finding := range findings {
		input, err := bot.makeInput(ctx, finding)
		if err != nil {
			return nil, fmt.Errorf("prepare generator input for %q: %w", finding, err)
		}
		files[finding.File] = append(files[finding.File], input)
	}

	generated, errs, err := g.Files(ctx, files, opts...)
	if err != nil {
		return nil, err
	}

	return &Patch{
		Patch: patch.New(generated, patch.WithErrors(errs)),
		// repo:        bot.fs,
		getLanguage: bot.language,
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
		Identifier: finding.Identifier,
		Target:     finding.Target,
	}

	return input, nil
}

func (p *Patch) Apply(ctx context.Context, root string) error {
	return p.Patch.Apply(ctx, afero.NewBasePathFs(afero.NewOsFs(), root), func(s string) (patch.Language, error) {
		return p.getLanguage(s)
	})
}
