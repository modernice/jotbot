package internal

import (
	"context"

	"golang.org/x/exp/slog"
)

var nop (slog.Handler) = nopLogger{}

// NopLogger returns a *slog.Logger that discards all log records. It is
// implemented by the nopLogger struct, which satisfies the slog.Handler
// interface.
func NopLogger() *slog.Logger {
	return slog.New(nop)
}

type nopLogger struct{}

// Enabled returns a boolean value indicating whether the logger is enabled for
// a given context and log level.
func (nopLogger) Enabled(context.Context, slog.Level) bool { return false }

// Handle is a method of the nopLogger type. It implements the Handle method of
// the slog.Handler interface. It takes a context.Context and a slog.Record as
// arguments, and returns an error. It always returns nil, indicating that the
// logging operation was successful.
func (nopLogger) Handle(context.Context, slog.Record) error { return nil }

// WithAttrs returns a slog.Handler that discards all log records and their
// attributes.
func (nopLogger) WithAttrs([]slog.Attr) slog.Handler { return nop }

// WithGroup returns a slog.Handler that discards all log records. The returned
// handler is equivalent to the input handler.
func (nopLogger) WithGroup(string) slog.Handler { return nop }
