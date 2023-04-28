package jotbot

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/modernice/jotbot/find"
	"github.com/modernice/jotbot/generate"
	"github.com/modernice/jotbot/internal/slice"
	"github.com/modernice/jotbot/langs/golang"
	"github.com/modernice/jotbot/patch"
	"golang.org/x/exp/slog"
)

var (
	Default = Options{
		Languages: map[string]LanguageService{
			".go": golang.New(golang.NewFinder()),
		},
	}
)

type LanguageService interface {
	Find(context.Context, []byte) ([]find.Finding, error)
	// Patch(context.Context, string, []byte) ([]byte, error)
}

type Options struct {
	Languages map[string]LanguageService
}

type JotBot struct {
	root      string
	languages map[string]LanguageService
	find      find.Options
	log       *slog.Logger
}

type Option func(*JotBot)

type Finding struct {
	find.Finding

	File string
}

func FindWith(opts find.Options) Option {
	return func(bot *JotBot) {
		bot.find = opts
	}
}

func (opts Options) New(root string) *JotBot {
	return &JotBot{
		root:      root,
		languages: opts.Languages,
		log:       slog.New(slog.NewTextHandler(os.Stdout)),
	}
}

func (bot *JotBot) ConfigureLanguage(ext string, svc LanguageService) {
	bot.languages[ext] = svc
}

func (bot *JotBot) Find(ctx context.Context) ([]Finding, error) {
	repo := os.DirFS(bot.root)
	files, err := bot.find.Find(ctx, repo)
	if err != nil {
		return nil, err
	}

	var out []Finding
	for _, file := range files {
		ext := filepath.Ext(file)
		lang, ok := bot.languages[ext]
		if !ok {
			bot.log.Warn(fmt.Sprintf("no language service for extension %q", ext))
			continue
		}

		path := filepath.Clean(filepath.Join(bot.root, file))

		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read file %s: %w", path, err)
		}

		findings, err := lang.Find(ctx, b)
		if err != nil {
			return nil, fmt.Errorf("find in %s: %w", path, err)
		}

		out = append(out, slice.Map(findings, func(f find.Finding) Finding {
			return Finding{f, file}
		})...)
	}

	return out, nil
}

func (bot *JotBot) Generate(ctx context.Context, findings []find.Finding, svc generate.Service, opts ...generate.Option) (*patch.Patch, error) {
	return nil, nil
	// gen := generate.New(svc)

	// results, errs, err := gen.Generate(ctx, os.DirFS(bot.root), opts...)
	// if err != nil {
	// 	return nil, err
	// }

	// files, err := internal.Drain(results, errs)
	// if err != nil {
	// 	return nil, err
	// }

	// patch := make(generate.Patch)
	// for _, file := range files {
	// 	patch[file.Path] = append(patch[file.Path], file.Generations...)
	// }

	// return patch, nil
}
