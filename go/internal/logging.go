package internal

import (
	"context"

	"golang.org/x/exp/slog"
)

var nop (slog.Handler) = nopLogger{}

func NopLogger() *slog.Logger {
	return slog.New(nop)
}

type nopLogger struct{}

func (nopLogger) Enabled(context.Context, slog.Level) bool { return false }

func (nopLogger) Handle(context.Context, slog.Record) error { return nil }

func (nopLogger) WithAttrs([]slog.Attr) slog.Handler { return nop }

func (nopLogger) WithGroup(string) slog.Handler { return nop }
