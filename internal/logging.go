package internal

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/exp/slog"
)

var nop (slog.Handler) = nopLogger{}

// NopLogger [func] returns a new logger that discards all log messages. It
// implements the slog.Handler interface and is intended to be used in tests or
// when logging is not desired.
func NopLogger() *slog.Logger {
	return slog.New(nop)
}

type nopLogger struct{}

// Enabled returns whether logging is enabled for the specified level. It is a
// method of the nopLogger type, which implements the slog.Handler interface.
func (nopLogger) Enabled(context.Context, slog.Level) bool { return false }

// Handle is a function implemented by the nopLogger type. It satisfies the
// slog.Handler interface by discarding any log records passed to it and
// returning nil.
func (nopLogger) Handle(context.Context, slog.Record) error { return nil }

// WithAttrs returns a new logger handler with the given attributes [slog.Attr]
// added to its context. The returned handler is a no-op logger and does not
// perform any logging.
func (nopLogger) WithAttrs([]slog.Attr) slog.Handler { return nop }

// WithGroup returns a new slog.Handler that is a copy of the receiver, but with
// the group field set to the provided value. The group field is used to
// categorize log handlers into logical groups.
func (nopLogger) WithGroup(string) slog.Handler { return nop }

type prettyLogger struct {
	slog.Handler
}

func PrettyLogger(h slog.Handler) slog.Handler {
	return &prettyLogger{h}
}

func (l *prettyLogger) Handle(ctx context.Context, r slog.Record) error {
	if r.Level == slog.LevelInfo {
		fmt.Fprintf(os.Stdout, "%s\n", r.Message)
		return nil
	}
	return l.Handler.Handle(ctx, r)
}
