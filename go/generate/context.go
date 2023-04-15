package generate

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"golang.org/x/exp/slices"
)

var _ Context = (*ctx)(nil)

type ctx struct {
	context.Context

	file       string
	identifier string

	repo  fs.FS
	files []string
}

func newCtx(parent context.Context, repo fs.FS, file, identifier string) (*ctx, error) {
	c := &ctx{
		Context:    parent,
		file:       file,
		identifier: identifier,
		repo:       repo,
	}
	if err := c.buildFileList(); err != nil {
		return nil, fmt.Errorf("build file list: %w", err)
	}
	return c, nil
}

func (ctx *ctx) File() string {
	return ctx.file
}

func (ctx *ctx) Files() []string {
	return ctx.files
}

func (ctx *ctx) Identifier() string {
	return ctx.identifier
}

func (ctx *ctx) Read(file string) ([]byte, error) {
	f, err := ctx.repo.Open(file)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", file, err)
	}
	defer f.Close()
	return io.ReadAll(f)
}

func (ctx *ctx) buildFileList() error {
	var files []string
	if err := fs.WalkDir(ctx.repo, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if d.Name() != "." && strings.HasPrefix(d.Name(), ".") {
				return fs.SkipDir
			}
			return nil
		}

		files = append(files, path)

		return nil
	}); err != nil {
		return err
	}

	slices.Sort(files)
	ctx.files = files

	return nil
}
