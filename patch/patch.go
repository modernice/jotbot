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

// Language represents the ability to update source code with documentation. It
// takes a context for cancellation, an identifier denoting something of
// interest within the code, a documentation string that should be associated
// with the identifier, and the original source code as a byte slice. It returns
// the updated source code as a byte slice along with any error encountered
// during the patching process. This interface is intended to be implemented by
// various programming language-specific services that know how to integrate
// documentation into their respective code formats.
type Language interface {
	// Patch updates the source code in the given language by inserting or altering
	// documentation strings identified by an identifier. It takes a context, an
	// identifier for the documentation to patch, the documentation text, and the
	// original source code as input. It returns the updated source code or an error
	// if the patching fails.
	Patch(ctx context.Context, identifier, doc string, code []byte) ([]byte, error)
}

// Patch represents a process for modifying files with documentation updates. It
// listens for file generation events and applies text patches to the content of
// these files based on language-specific rules provided by a Language service.
// Patch supports both dry-run operations, which do not alter the filesystem but
// return a map of the intended changes, and actual application of changes to
// the filesystem. It can be configured with various options such as error
// channels and logging handlers to tailor its behavior during the patching
// process.
type Patch struct {
	files <-chan generate.File
	errs  <-chan error
	log   *slog.Logger
}

// Option configures a [*Patch] by setting optional parameters.
type Option func(*Patch)

// WithErrors specifies an error channel to be used by Patch for error
// reporting. It modifies the provided Patch instance to receive and handle
// errors during its operations. This option allows the caller to monitor and
// respond to errors that occur while Patch is processing files.
func WithErrors(errs <-chan error) Option {
	return func(p *Patch) {
		p.errs = errs
	}
}

// WithLogger configures a Patch instance with a specific slog.Logger to handle
// logging output. It accepts a slog.Handler which is used to create the new
// logger. This function is intended to be passed as an option when creating a
// new Patch instance.
func WithLogger(h slog.Handler) Option {
	return func(p *Patch) {
		p.log = slog.New(h)
	}
}

// New initializes a new Patch with provided file channel and optional
// configurations. It ensures the presence of a logger, either provided through
// options or a no-operation logger by default. It returns the initialized
// Patch.
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

// DryRun simulates the patching process by applying modifications to files in
// memory and returns the resulting content without writing changes to the file
// system. It accepts a context, a file system abstraction, and a function to
// retrieve language-specific patching services based on file extensions. On
// success, it returns a map associating file paths with their modified content.
// If an error occurs during the simulation, it returns the partial results
// along with the encountered error.
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

// Apply processes a series of files intended for patching, applying the changes
// defined within them to the corresponding files in the provided filesystem
// repository. It takes a context for cancellation and timeout control, a
// filesystem interface to access file data, and a function to retrieve
// language-specific patching services based on file extensions. This method
// logs informational messages about its progress and warnings if any issues
// arise during the patching process. If an error occurs that prevents a file
// from being patched, Apply continues with the next file without terminating
// the entire operation. It returns an error only if the context is canceled or
// closed.
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
