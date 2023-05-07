package patch

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/modernice/jotbot/generate"
	"github.com/modernice/jotbot/internal"
	"github.com/spf13/afero"
	"golang.org/x/exp/slog"
)

type Language interface {
	Patch(ctx context.Context, identifier, doc string, code []byte) ([]byte, error)
}

type Patch struct {
	files <-chan generate.File
	errs  <-chan error
	log   *slog.Logger
}

type Option func(*Patch)

func WithErrors(errs <-chan error) Option {
	return func(p *Patch) {
		p.errs = errs
	}
}

func WithLogger(h slog.Handler) Option {
	return func(p *Patch) {
		p.log = slog.New(h)
	}
}

func New(files <-chan generate.File, opts ...Option) *Patch {
	p := &Patch{files: files}
	for _, opt := range opts {
		opt(p)
	}
	if p.log == nil {
		p.log = internal.NopLogger()
	}
	return p
}

func (p *Patch) DryRun(ctx context.Context, repo afero.Fs, getLanguage func(string) (Language, error)) (map[string][]byte, error) {
	files, err := internal.Drain(p.files, p.errs)
	if err != nil {
		return nil, err
	}

	out := make(map[string][]byte, len(files))
	for _, file := range files {
		ext := filepath.Ext(file.Path)
		svc, err := getLanguage(ext)
		if err != nil {
			return out, fmt.Errorf("get language service for %q files: %w", ext, err)
		}

		code, err := p.applyFile(ctx, repo, svc, file, false)
		if err != nil {
			return out, fmt.Errorf("apply patch to %q: %w", file.Path, err)
		}
		out[file.Path] = code
	}

	return out, nil
}

func (p *Patch) Apply(ctx context.Context, repo afero.Fs, getLanguage func(string) (Language, error)) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err, ok := <-p.errs:
			if !ok {
				p.errs = nil
				continue
			}
			p.log.Warn(fmt.Sprintf("Failed to generate file: %v", err))
			continue
		case file, ok := <-p.files:
			if !ok {
				return nil
			}

			p.log.Info(fmt.Sprintf("Patching %s ...", file.Path))

			ext := filepath.Ext(file.Path)
			svc, err := getLanguage(ext)
			if err != nil {
				p.log.Warn(fmt.Sprintf("Get language service for %q files: %v", ext, err), "file", file.Path)
				break
			}

			if _, err := p.applyFile(ctx, repo, svc, file, true); err != nil {
				p.log.Warn(fmt.Sprintf("Failed to apply patch: %v", err), "file", file.Path)
			}
		}
	}
}

func (p *Patch) applyFile(ctx context.Context, repo afero.Fs, svc Language, file generate.File, write bool) ([]byte, error) {
	code, err := readFile(repo, file.Path)
	if err != nil {
		return code, err
	}

	for _, doc := range file.Docs {
		if patched, err := svc.Patch(ctx, doc.Identifier, doc.Text, code); err != nil {
			p.log.Debug(fmt.Sprintf("failed to patch %q: %v", doc.Identifier, err), "documentation", doc.Text)
			return code, fmt.Errorf("apply patch to %q: %w", doc.Identifier, err)
		} else {
			code = patched
		}
	}

	if !write {
		return code, nil
	}

	f, err := repo.Create(file.Path)
	if err != nil {
		return code, fmt.Errorf("create %s: %w", file.Path, err)
	}
	defer f.Close()

	if _, err := f.Write(code); err != nil {
		return code, err
	}

	if err := f.Close(); err != nil {
		return code, err
	}

	return code, nil
}

func readFile(repo afero.Fs, file string) ([]byte, error) {
	f, err := repo.Open(file)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", file, err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", file, err)
	}

	return b, nil
}
