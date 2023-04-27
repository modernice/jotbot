package jotbot

import (
	"context"

	"github.com/modernice/jotbot/generate"
)

type Language[F Finder, P Patcher] interface {
	Finder() F
	Patcher() P
}

type Finder interface {
	Find(context.Context, []byte) ([]Finding, error)
}

type Finding = generate.Finding

type Patcher interface {
	Patch(context.Context, []byte, []Finding) ([]byte, error)
}
