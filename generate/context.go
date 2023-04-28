package generate

import (
	"context"
)

var _ Context = (*genCtx)(nil)

type genCtx struct {
	context.Context

	input Input
}

func newCtx(parent context.Context, input Input) *genCtx {
	return &genCtx{
		Context: parent,
		input:   input,
	}
}

func (ctx *genCtx) Identifier() string {
	return ctx.input.Identifier
}

func (ctx *genCtx) Target() string {
	return ctx.input.NaturalLanguageIdentifier
}

func (ctx *genCtx) File() []byte {
	return ctx.input.Code
}
