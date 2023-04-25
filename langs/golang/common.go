package golang

import "golang.org/x/exp/slog"

type Logger slog.Logger

func WithLogger(h slog.Handler) *Logger {
	return (*Logger)(slog.New(h))
}

func (opt *Logger) applyFinder(f *Finder) {
	f.log = (*slog.Logger)(opt)
}

func (opt *Logger) applyPatch(p *Patch) {
	p.log = (*slog.Logger)(opt)
}
