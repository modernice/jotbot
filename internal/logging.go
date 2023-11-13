package internal

import (
	"context"
	"fmt"
	"os"
	"strings"

	"golang.org/x/exp/slog"
)

// LogLevelNaked represents a logging level that is not associated with any
// predefined severity, allowing for custom handling in logging operations. It
// is used to indicate a log entry that should not be adorned with a standard
// level icon or prefix when processed by a custom log handler such as
// PrettyLogger.
const LogLevelNaked = slog.Level(-1)

var nop (slog.Handler) = nopLogger{}

// NopLogger returns a new logger instance that will ignore all log events,
// effectively performing no operations when handling logs. It provides a
// [*slog.Logger] that can be used where logging is necessary but not desired to
// have any output or effect.
func NopLogger() *slog.Logger {
	return slog.New(nop)
}

type nopLogger struct{}

// Enabled reports whether the logger is enabled for messages at the specified
// level within the given context. It always returns false for nopLogger,
// indicating that no logging will occur regardless of the level or context
// provided.
func (nopLogger) Enabled(context.Context, slog.Level) bool { return false }

// Handle silently ignores the incoming log records without performing any
// operations or emitting logs. It always returns nil, indicating no error
// occurred during its invocation.
func (nopLogger) Handle(context.Context, slog.Record) error { return nil }

// WithAttrs appends a given set of attributes to the logger and returns a new
// logger instance with those attributes.
func (nopLogger) WithAttrs([]slog.Attr) slog.Handler { return nop }

// WithGroup associates a named group with the logger, returning a new logger
// instance that is functionally identical to the original as this method is
// implemented by a no-operation logger.
func (nopLogger) WithGroup(string) slog.Handler { return nop }

type prettyLogger struct {
	slog.Handler
}

// PrettyLogger wraps a given slog.Handler to enhance logging output with
// visually distinct icons based on the log level and formats attributes for
// improved readability. It returns a slog.Handler that can be used to handle
// log records in a more user-friendly manner.
func PrettyLogger(h slog.Handler) slog.Handler {
	return &prettyLogger{h}
}

// Handle processes a log record by outputting a formatted message to standard
// output. It includes an icon representing the log level, the log message
// itself, and any associated attributes. It does not filter any log levels and
// does not return errors under normal operation.
func (l *prettyLogger) Handle(ctx context.Context, r slog.Record) error {
	var icon rune

	switch r.Level {
	case slog.LevelDebug:
		icon = '⚙'
	case slog.LevelInfo:
		icon = 'ℹ'
	case slog.LevelWarn:
		icon = '⚠'
	case slog.LevelError:
		icon = '✖'
	case LogLevelNaked:
	}

	fmt.Fprint(os.Stdout, strings.TrimLeft(fmt.Sprintf("%c %s", icon, r.Message), " "))

	r.Attrs(func(a slog.Attr) bool {
		fmt.Fprintf(os.Stdout, " %s=%v", a.Key, a.Value)
		return true
	})

	fmt.Fprintln(os.Stdout)

	return nil
}
