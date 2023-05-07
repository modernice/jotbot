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

// Language is an interface that defines the Patch method for applying patches
// to code documentation. It takes a context, an identifier, a documentation
// string, and a byte slice of code, and returns a patched byte slice of code
// along with any errors encountered during the process.
type Language interface {
	// Patch applies a documentation patch to the given code using the provided
	// Language implementation. It takes a context, an identifier, a doc string, and
	// a slice of bytes representing the code. It returns a slice of bytes with the
	// patched code or an error if something went wrong.
	Patch(ctx context.Context, identifier, doc string, code []byte) ([]byte, error)
}

// Patch applies generated documentation to source code files by processing
// those files and applying the appropriate language-specific patches. It can be
// configured with a custom error channel and logger, and provides methods for
// dry run and actual application of patches. The patching process is based on
// the provided Language interface implementation for each file extension.
type Patch struct {
	files <-chan generate.File
	errs  <-chan error
	log   *slog.Logger
}

// Option is a function that modifies a Patch configuration. It is used as an
// argument for the New function to customize the behavior of the Patch
// instance. Common options include WithErrors to provide an error channel and
// WithLogger to provide a custom logger.
type Option func(*Patch)

// WithErrors configures a Patch to receive errors from the provided error
// channel. These errors are logged as warnings and don't cause the patching
// process to fail.
func WithErrors(errs <-chan error) Option {
	return func(p *Patch) {
		p.errs = errs
	}
}

// WithLogger sets the logger for a Patch instance using the provided
// slog.Handler.
func WithLogger(h slog.Handler) Option {
	return func(p *Patch) {
		p.log = slog.New(h)
	}
}

// New creates a new Patch instance with the provided generate.File channel and
// optional configuration options. The returned Patch can be used to apply
// generated documentation patches to code files.
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

// DryRun performs a dry run of applying patches to the given repository,
// returning a map of file paths to their resulting patched content without
// writing any changes. It uses the provided function to get a language-specific
// patching service for each file based on its extension.
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

// Apply applies patches to the generated files provided by the generate.File
// channel. It reads each file, applies patches using the provided Language
// service, and writes the patched content back to the file. Errors encountered
// during patching are logged and do not stop the process. The function returns
// when all files have been processed or the context is done.
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
			p.log.Warn(fmt.Sprintf("Failed to generate doc: %v", err))
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
