package internal

import (
	"errors"

	"github.com/tiktoken-go/tokenizer"
)

// OpenAITokenizer returns a tokenizer.Codec and an error. It uses the provided
// model string to create a tokenizer for the specified model. If the model is
// not supported, it falls back to the tokenizer.Cl100kBase model.
func OpenAITokenizer(model string) (tokenizer.Codec, error) {
	codec, err := tokenizer.ForModel(tokenizer.Model(model))
	if errors.Is(err, tokenizer.ErrModelNotSupported) {
		return tokenizer.Get(tokenizer.Cl100kBase)
	}
	return codec, err
}
