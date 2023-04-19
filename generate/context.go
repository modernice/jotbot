package generate

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"sync"

	"golang.org/x/exp/slices"
)

var _ Context = (*genCtx)(nil)

type genCtx struct {
	context.Context

	file       string
	identifier string

	repo  fs.FS
	files []string

	// shared with all child instances that are created with ctx.new()
	mux       *sync.RWMutex
	fileCache map[string][]byte
}

func newCtx(parent context.Context, repo fs.FS, file, identifier string) (*genCtx, error) {
	ctx := &genCtx{
		Context:    parent,
		file:       file,
		identifier: identifier,
		mux:        &sync.RWMutex{},
		repo:       repo,
		fileCache:  make(map[string][]byte),
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
		mux:        ctx.mux,
		repo:       ctx.repo,
		files:      ctx.files,
		fileCache:  ctx.fileCache,
	}
}

// File returns the name of the file being processed by the generator's context
// [*genCtx.File].
func (ctx *genCtx) File() string {
	return ctx.file
}

// Files returns a list of file names, including their paths relative to the
// root directory, that can be accessed by the context.
func (ctx *genCtx) Files() []string {
	return ctx.files
}

// Identifier returns the identifier of the current genCtx instance.
func (ctx *genCtx) Identifier() string {
	return ctx.identifier
}

// Read reads the contents of a file from the file system and returns it as a
// byte slice. If the file has already been read, it returns the cached result.
// If the file does not exist or cannot be read, an error is returned.
func (ctx *genCtx) Read(file string) ([]byte, error) {
	if b, ok := ctx.cached(file); ok {
		return b, nil
	}

	ctx.mux.Lock()
	defer ctx.mux.Unlock()

	f, err := ctx.repo.Open(file)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", file, err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return b, err
	}
	ctx.fileCache[file] = b

	return b, nil
}

func (ctx *genCtx) cached(file string) ([]byte, bool) {
	ctx.mux.RLock()
	defer ctx.mux.RUnlock()
	b, ok := ctx.fileCache[file]
	return b, ok
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
