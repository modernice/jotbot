package internal

import (
	"errors"

	"github.com/tiktoken-go/tokenizer"
)

// OpenAITokenizer initializes and returns a tokenizer codec based on the
// specified model. If the model is not supported, it falls back to the default
// CL100kBase tokenizer. It may return an error if initializing the tokenizer
// fails for reasons other than an unsupported model. The returned codec can be
// used to tokenize or detokenize text according to the OpenAI specifications
// associated with the model.
func OpenAITokenizer(model string) (tokenizer.Codec, error) {
	codec, err := tokenizer.ForModel(tokenizer.Model(model))
	if errors.Is(err, tokenizer.ErrModelNotSupported) {
		return tokenizer.Get(tokenizer.Cl100kBase)
	}
	return codec, err
}
