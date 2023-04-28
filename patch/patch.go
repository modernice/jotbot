package patch

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/modernice/jotbot/generate"
	"github.com/spf13/afero"
)

type LanguageService interface {
	Patch(ctx context.Context, identifier, doc string, code []byte) ([]byte, error)
}

type Patch struct {
	files <-chan generate.File
	errs  <-chan error
}

type Option func(*Patch)

func WithErrors(errs <-chan error) Option {
	return func(p *Patch) {
		p.errs = errs
	}
}

func New(files <-chan generate.File, opts ...Option) *Patch {
	p := &Patch{files: files}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *Patch) Apply(ctx context.Context, repo afero.Fs, getLanguage func(string) (LanguageService, error)) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-p.errs:
			return err
		case file, ok := <-p.files:
			if !ok {
				return nil
			}

			ext := filepath.Ext(file.Path)
			svc, err := getLanguage(ext)
			if err != nil {
				return fmt.Errorf("get language service for %q files: %w", ext, err)
			}

			if err := p.applyFile(ctx, repo, svc, file); err != nil {
				return fmt.Errorf("apply patch to %s: %w", file.Path, err)
			}
		}
	}
}

func (p *Patch) applyFile(ctx context.Context, repo afero.Fs, svc LanguageService, file generate.File) error {
	code, err := readFile(repo, file.Path)
	if err != nil {
		return err
	}

	for _, doc := range file.Docs {
		if code, err = svc.Patch(ctx, doc.Identifier, doc.Text, code); err != nil {
			return fmt.Errorf("apply patch to %q: %w", doc.Identifier, err)
		}
	}

	f, err := repo.Create(file.Path)
	if err != nil {
		return fmt.Errorf("create %s: %w", file.Path, err)
	}
	defer f.Close()

	if _, err := f.Write(code); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	return nil
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
