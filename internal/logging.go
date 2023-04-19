package internal

import (
	"context"

	"golang.org/x/exp/slog"
)

var nop (slog.Handler) = nopLogger{}

// NopLogger provides a logger that discards all log records. It implements the
// slog.Handler interface from the golang.org/x/exp/slog package, and can be
// used as a drop-in replacement for the default logger.
func NopLogger() *slog.Logger {
	return slog.New(nop)
}

type nopLogger struct{}

// Enabled checks whether the logger is enabled for a given logging level. It
// takes a context.Context and a slog.Level as input parameters, and returns a
// boolean value indicating whether the logger is enabled or not.
func (nopLogger) Enabled(context.Context, slog.Level) bool { return false }

// Handle is a method of nopLogger that implements the slog.Handler interface.
// It takes a context and a slog.Record as parameters, and returns nil. This
// method does not perform any logging or record handling.
func (nopLogger) Handle(context.Context, slog.Record) error { return nil }

// WithAttrs returns a logger handler which discards all log records and
// attributes passed to it. It takes a slice of slog.Attr as its argument, and
// returns a slog.Handler.
func (nopLogger) WithAttrs([]slog.Attr) slog.Handler { return nop }

// WithGroup returns a slog.Handler that discards all log records. The returned
// handler is equivalent to the input handler.
func (nopLogger) WithGroup(string) slog.Handler { return nop }
