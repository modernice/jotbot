package generate

import (
	"context"
	"fmt"
	"io"
	"io/fs"

	"golang.org/x/exp/slices"
)

var _ Context = (*genCtx)(nil)

type genCtx struct {
	context.Context

	file       string
	identifier string

	repo  fs.FS
	files []string

	// // shared with all child instances that are created with ctx.new()
	// mux       *sync.RWMutex
	// fileCache map[string][]byte
}

func newCtx(parent context.Context, repo fs.FS, file, identifier string) (*genCtx, error) {
	ctx := &genCtx{
		Context:    parent,
		file:       file,
		identifier: identifier,
		// mux:        &sync.RWMutex{},
		repo: repo,
		// fileCache:  make(map[string][]byte),
	}
	if err := ctx.buildFileList(); err != nil {
		return nil, fmt.Errorf("build file list: %w", err)
	}
	return ctx, nil
}

func (ctx *genCtx) new(parent context.Context, file, identifier string) *genCtx {
	return &genCtx{
		Context:    parent,
		file:       file,
		identifier: identifier,
		// mux:        ctx.mux,
		repo:  ctx.repo,
		files: ctx.files,
		// fileCache:  ctx.fileCache,
	}
}

func (ctx *genCtx) File() string {
	return ctx.file
}

func (ctx *genCtx) Files() []string {
	return ctx.files
}

func (ctx *genCtx) Identifier() string {
	return ctx.identifier
}

func (ctx *genCtx) Read(file string) ([]byte, error) {
	f, err := ctx.repo.Open(file)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", file, err)
	}
	defer f.Close()
	return io.ReadAll(f)
}

func (ctx *genCtx) buildFileList() error {
	var files []string
	if err := fs.WalkDir(ctx.repo, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
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
