package generate

import "context"

var _ Context = (*genCtx)(nil)

type genCtx struct {
	context.Context

	input  PromptInput
	prompt string
}

func newCtx(parent context.Context, input PromptInput, prompt string) *genCtx {
	return &genCtx{
		Context: parent,
		input:   input,
		prompt:  prompt,
	}
}

// Input returns the PromptInput associated with the genCtx instance. This is
// typically used to retrieve user input that has been stored within a context.
func (ctx *genCtx) Input() PromptInput {
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
