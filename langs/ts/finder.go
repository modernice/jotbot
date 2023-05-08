package ts

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/modernice/jotbot/internal"
	"github.com/modernice/jotbot/internal/slice"
	"golang.org/x/exp/slog"
)

const (
	// Var represents a variable symbol in the code.
	Var = Symbol("var")

	Class     = Symbol("class")
	Interface = Symbol("iface")
	Func      = Symbol("func")
	Method    = Symbol("method")
	Property  = Symbol("prop")
)

// Symbol represents a type of code construct, such as variables, classes,
// interfaces, functions, methods, or properties. It is used by the Finder to
// search for and filter specific types of symbols within code.
type Symbol string

// Position represents a specific location within a text document, identified by
// its line number and character position on that line.
type Position struct {
	Line      int
	Character int
}

// Finder is a utility that searches for specified [Symbol]s within provided
// code and optionally includes documented symbols. It can also find the
// position of a given identifier within the code.
type Finder struct {
	symbols           []Symbol
	includeDocumented bool
	log               *slog.Logger
}

// FinderOption is a function that configures a [Finder] by modifying its
// fields. Common options include Symbols, IncludeDocumented, and WithLogger.
type FinderOption func(*Finder)

// Symbols returns a FinderOption that appends the provided symbols to the list
// of symbols the Finder should search for in the code.
func Symbols(symbols ...Symbol) FinderOption {
	return func(f *Finder) {
		f.symbols = append(f.symbols, symbols...)
	}
}

// IncludeDocumented is a FinderOption that configures a Finder to include or
// exclude documented symbols in its search results based on the provided
// boolean value.
func IncludeDocumented(include bool) FinderOption {
	return func(f *Finder) {
		f.includeDocumented = include
	}
}

// WithLogger sets the logger for a Finder. It accepts an *slog.Logger and
// returns a FinderOption that configures the Finder to use the provided logger.
func WithLogger(log *slog.Logger) FinderOption {
	return func(f *Finder) {
		f.log = log
	}
}

// NewFinder creates a new Finder with the provided options. A Finder searches
// for symbols in TypeScript code and returns their positions.
func NewFinder(opts ...FinderOption) *Finder {
	var f Finder
	for _, opt := range opts {
		opt(&f)
	}
	if f.log == nil {
		f.log = internal.NopLogger()
	}
	return &f
}

// Find searches the provided code for symbols specified in the Finder
// configuration, and returns a slice of strings containing the found symbols.
// It respects the includeDocumented flag in the Finder configuration. The
// search is performed within the provided context.
func (f *Finder) Find(ctx context.Context, code []byte) ([]string, error) {
	raw, err := f.executeFind(ctx, code)
	if err != nil {
		return nil, err
	}

	var found []string
	if err := json.Unmarshal(raw, &found); err != nil {
		return nil, fmt.Errorf("unmarshal findings: %w\n%s", err, raw)
	}

	return found, nil
}

func (f *Finder) executeFind(ctx context.Context, code []byte) ([]byte, error) {
	args := []string{"find", "--json"}

	if len(f.symbols) > 0 {
		symbols := internal.JoinStrings(slice.Map(f.symbols, unquote[Symbol]), ",")
		args = append(args, "-s", string(symbols))
	}

	if f.includeDocumented {
		args = append(args, "--documented")
	}

	args = append(args, string(code))

	cmd := exec.CommandContext(ctx, jotbotTSPath, args...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w:\n%s", err, out)
	}

	return out, nil
}

// Position determines the line and character position of the specified
// identifier within the provided code. It returns a Position struct containing
// the line and character information or an error if the operation fails.
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
	args := []string{"pos", identifier, string(code)}

	cmd := exec.CommandContext(ctx, jotbotTSPath, args...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w:\n%s", err, out)
	}

	return out, nil
}

func unquote[S ~string](s S) S {
	return S(strings.Trim(string(s), `"`))
}

var jotbotTSPath = os.Getenv("JOTBOT_TS_PATH")

func init() {
	if jotbotTSPath == "" {
		jotbotTSPath = "jotbot-ts"
	}
}
