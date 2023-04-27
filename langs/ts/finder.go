package ts

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modernice/jotbot"
	"github.com/modernice/jotbot/internal"
	"github.com/modernice/jotbot/internal/slice"
	"golang.org/x/exp/slog"
)

var _ (jotbot.Finder) = (*Finder)(nil)

const (
	Var      = Symbol("var")
	Class    = Symbol("class")
	Func     = Symbol("func")
	Method   = Symbol("method")
	Property = Symbol("property")
)

type Symbol string

type Finder struct {
	symbols []Symbol
	include []string
	exclude []string
	log     *slog.Logger
}

type FinderOption interface {
	applyFinder(*Finder)
}

type finderOptionFunc func(*Finder)

func (opt finderOptionFunc) applyFinder(s *Finder) {
	opt(s)
}

func Symbols(symbols ...Symbol) FinderOption {
	return finderOptionFunc(func(f *Finder) {
		f.symbols = append(f.symbols, symbols...)
	})
}

func Include(include ...string) FinderOption {
	return finderOptionFunc(func(f *Finder) {
		f.include = append(f.include, include...)
	})
}

func Exclude(exclude ...string) FinderOption {
	return finderOptionFunc(func(f *Finder) {
		f.exclude = append(f.exclude, exclude...)
	})
}

func WithLogger(log *slog.Logger) FinderOption {
	return finderOptionFunc(func(f *Finder) {
		f.log = log
	})
}

func NewFinder(opts ...FinderOption) *Finder {
	var f Finder
	for _, opt := range opts {
		opt.applyFinder(&f)
	}
	if f.log == nil {
		f.log = internal.NopLogger()
	}
	return &f
}

type findings map[string][]jotbot.Finding

func (f *Finder) Find(ctx context.Context, code []byte) ([]jotbot.Finding, error) {
	dir, err := f.createTempFile(code)
	if err != nil {
		return nil, err
	}

	raw, err := f.execute(ctx, dir)
	if err != nil {
		return nil, err
	}

	if err := os.RemoveAll(dir); err != nil {
		return nil, fmt.Errorf("delete working directory: %w", err)
	}

	var found findings
	if err := json.Unmarshal(raw, &found); err != nil {
		return nil, fmt.Errorf("unmarshal findings: %w\n%s", err, raw)
	}

	for _, v := range found {
		return v, nil
	}

	return nil, nil
}

func (f *Finder) createTempFile(code []byte) (string, error) {
	dir, err := os.MkdirTemp("", "jotbot-*")
	if err != nil {
		return "", fmt.Errorf("create temporary directory: %w", err)
	}

	tmpFile, err := os.CreateTemp(dir, "code.ts")
	if err != nil {
		return dir, fmt.Errorf("create temporary code file: %w", err)
	}

	if _, err := tmpFile.Write(code); err != nil {
		return "", fmt.Errorf("write code to temporary file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("close temporary code file: %w", err)
	}

	path := tmpFile.Name()
	withoutExtension := strings.TrimSuffix(path, filepath.Ext(path))
	name := fmt.Sprintf("%s%s", withoutExtension, ".ts")

	if err := os.Rename(path, name); err != nil {
		return "", fmt.Errorf("rename temporary code file: %w", err)
	}

	return dir, nil
}

func (f *Finder) execute(ctx context.Context, dir string) ([]byte, error) {
	symbols := internal.JoinStrings(slice.Map(f.symbols, unquote[Symbol]), ",")
	include := internal.JoinStrings(slice.Map(f.include, unquote[string]), ",")
	exclude := internal.JoinStrings(slice.Map(f.exclude, unquote[string]), ",")

	var stdout bytes.Buffer

	args := []string{"find", "--json"}

	if symbols != "" {
		args = append(args, "-s", string(symbols))
	}

	if include != "" {
		args = append(args, "-i", include)
	}

	if exclude != "" {
		args = append(args, "-e", exclude)
	}

	args = append(args, dir)

	cmd := exec.CommandContext(ctx, "jotbot-es", args...)

	cmd.Stdout = &stdout
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("pipe stderr: %w", err)
	}

	done := f.logErrors(stderr)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start jotbot-es: %w", err)
	}

	<-done
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("wait for jotbot-es: %w", err)
	}

	return stdout.Bytes(), nil
}

func (f *Finder) logErrors(r io.Reader) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			l := scanner.Text()
			f.log.Info(l)
		}
	}()
	return done
}

func unquote[S ~string](s S) S {
	return S(strings.Trim(string(s), `"`))
}
