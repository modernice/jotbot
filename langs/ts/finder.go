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
	// Var represents the symbolic constant for identifying variable declarations in
	// TypeScript code within the context of the Finder's operations.
	Var = Symbol("var")

	// Class represents a TypeScript class symbol for use in code analysis and
	// manipulation tasks within the package. It identifies class declarations when
	// performing operations such as finding symbols or determining their positions
	// in source code.
	Class = Symbol("class")

	// Interface represents the TypeScript interface symbol that can be used to
	// filter TypeScript entities during code analysis.
	Interface = Symbol("iface")

	// Func represents the identifier for a TypeScript function symbol.
	Func = Symbol("func")

	// Method represents a TypeScript method symbol.
	Method = Symbol("method")

	// Property represents a TypeScript object property symbol used for identifying
	// such properties within source code during static analysis.
	Property = Symbol("prop")
)

// Symbol represents a distinct element or token in the TypeScript language that
// can be targeted for identification within the source code. It encapsulates a
// classification of TypeScript entities such as variables, classes, interfaces,
// functions, methods, and properties. Each instance of Symbol corresponds to a
// specific kind of these language constructs, allowing for operations like
// searching and analysis to be performed on the corresponding code segments
// that define or reference these entities.
type Symbol string

// Position represents the location within a text document, such as source code.
// It holds information about a specific point in the document, typically
// identified by a line number and a character offset on that line. This can be
// used to pinpoint exact locations, such as where a syntax error occurs or
// where a particular identifier is defined.
type Position struct {
	Line      int
	Character int
}

// Finder provides functionality for locating specific symbols within TypeScript
// code. It supports customization through options that can specify which
// symbols to look for, whether to include documented symbols in the search, and
// an optional logger for logging purposes. Finder offers methods to execute
// searches that return either a list of found symbol names or the position of a
// particular symbol. It handles the execution context and potential errors,
// returning structured results based on the TypeScript code provided.
type Finder struct {
	symbols           []Symbol
	includeDocumented bool
	log               *slog.Logger
}

// FinderOption configures a [Finder] instance, allowing customization of its
// behavior during the symbol searching process. It's used to specify which
// symbols to include, whether to include documented symbols, and to set a
// logger for outputting information during the search operations.
type FinderOption func(*Finder)

// Symbols specifies a set of TypeScript symbols to be included in the search by
// a Finder. It configures a Finder to consider only the provided symbols when
// performing code analysis operations.
func Symbols(symbols ...Symbol) FinderOption {
	return func(f *Finder) {
		f.symbols = append(f.symbols, symbols...)
	}
}

// IncludeDocumented configures a Finder to consider documented symbols in its
// search. If set to true, the Finder will include symbols with associated
// documentation in its results; otherwise, it will exclude them. This option
// can be passed to NewFinder when creating a new instance of Finder.
func IncludeDocumented(include bool) FinderOption {
	return func(f *Finder) {
		f.includeDocumented = include
	}
}

// WithLogger configures a Finder with a specified logger. It allows for logging
// within the Finder's operations, utilizing the provided [*slog.Logger]. This
// option can be passed to NewFinder to influence its logging behavior.
func WithLogger(log *slog.Logger) FinderOption {
	return func(f *Finder) {
		f.log = log
	}
}

// NewFinder constructs a new Finder instance with the provided options. It
// returns a pointer to the created Finder. If no logger is provided via
// options, it assigns a no-operation logger by default. Options can be used to
// specify which symbols to look for and whether to include documented symbols
// in the search results.
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

// Find searches for specified symbols in the provided TypeScript code and
// returns a list of findings. It respects the configured symbols and
// documentation inclusion settings of the Finder instance. If an error occurs
// during the search, it is returned along with an empty list. The context
// parameter allows the search to be canceled or have a deadline.
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

// Position locates the position of a specified identifier within a given body
// of code and returns its location as a [Position]. If the identifier cannot be
// found or another error occurs, an error is returned instead. The search is
// conducted within the provided context for cancellation and timeout handling.
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
