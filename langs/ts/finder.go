package ts

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
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

type Position struct {
	Line      int
	Character int
}

type Finder struct {
	symbols []Symbol
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

func (f *Finder) Find(ctx context.Context, code []byte) ([]jotbot.Finding, error) {
	raw, err := f.executeFind(ctx, code)
	if err != nil {
		return nil, err
	}

	var found []jotbot.Finding
	if err := json.Unmarshal(raw, &found); err != nil {
		return nil, fmt.Errorf("unmarshal findings: %w\n%s", err, raw)
	}

	return found, nil
}

func (f *Finder) executeFind(ctx context.Context, code []byte) ([]byte, error) {
	symbols := internal.JoinStrings(slice.Map(f.symbols, unquote[Symbol]), ",")

	var stdout bytes.Buffer

	args := []string{"find", "-v", "--json"}

	if symbols != "" {
		args = append(args, "-s", string(symbols))
	}

	args = append(args, string(code))

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

func (f *Finder) Position(ctx context.Context, identifier string, code []byte) (Position, error) {
	raw, err := f.executePosition(ctx, identifier, code)
	if err != nil {
		return Position{}, err
	}

	var pos Position
	if err := json.Unmarshal(raw, &pos); err != nil {
		return Position{}, fmt.Errorf("unmarshal position: %w\n%s", err, raw)
	}

	return pos, nil
}

func (f *Finder) executePosition(ctx context.Context, identifier string, code []byte) ([]byte, error) {
	var stdout bytes.Buffer

	args := []string{"pos", identifier, string(code)}

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
