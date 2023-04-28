package patch

import (
	"context"
	"fmt"
	"io"
	"io/fs"

	"github.com/modernice/jotbot/generate"
)

type LanguageService interface {
	Patch(ctx context.Context, identifier, doc string, code []byte) ([]byte, error)
}

type Patch struct {
	files <-chan generate.File
}

func NewPatch(files <-chan generate.File) *Patch {
	return &Patch{
		files: files,
	}
}

func (p *Patch) Apply(ctx context.Context, repo fs.FS, svc LanguageService) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case file, ok := <-p.files:
			if !ok {
				return nil
			}
			if err := p.applyFile(ctx, repo, svc, file); err != nil {
				return fmt.Errorf("apply patch to %s: %w", file.Path, err)
			}
		}
	}
}

func (p *Patch) applyFile(ctx context.Context, repo fs.FS, svc LanguageService, file generate.File) error {
	code, err := readFile(repo, file.Path)
	if err != nil {
		return err
	}

	for _, doc := range file.Docs {
		if code, err = svc.Patch(ctx, doc.Identifier, doc.Text, code); err != nil {
			return fmt.Errorf("apply patch to %q: %w", doc.Identifier, err)
		}
	}

	return nil
}

func readFile(repo fs.FS, file string) ([]byte, error) {
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
