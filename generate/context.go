package generate

import "context"

var _ Context = (*genCtx)(nil)

type genCtx struct {
	context.Context

	input  Input
	prompt string
}

func newCtx(parent context.Context, input Input, prompt string) *genCtx {
	return &genCtx{
		Context: parent,
		input:   input,
		prompt:  prompt,
	}
}

func (ctx *genCtx) Input() Input {
	return ctx.input
}

func (ctx *genCtx) Prompt() string {
	return ctx.prompt
}

func (ctx *genCtx) File() []byte {
	return ctx.input.Code
}
