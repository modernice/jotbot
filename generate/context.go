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

// Input returns the Input associated with the genCtx instance.
func (ctx *genCtx) Input() Input {
	return ctx.input
}

// Prompt returns the prompt string associated with the genCtx instance.
func (ctx *genCtx) Prompt() string {
	return ctx.prompt
}

// File returns the code content of the input as a byte slice.
func (ctx *genCtx) File() []byte {
	return ctx.input.Code
}
