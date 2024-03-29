package openai

import (
	"fmt"
	"strings"
	"sync"

	"github.com/modernice/jotbot/internal"
	"github.com/sashabaranov/go-openai"
	"github.com/tiktoken-go/tokenizer"
)

var (
	codecsMux sync.Mutex
	codecs    = make(map[string]tokenizer.Codec)
)

// ChatTokens calculates the total number of tokens that would be generated by
// encoding a sequence of chat completion messages using a specified model's
// tokenizer. It takes into account any additional tokens required for the
// message names and content, based on the model's specific tokenization rules.
// The function returns the calculated token count alongside any error
// encountered during the tokenization process. If an error occurs, it includes
// additional context to clarify that the error happened while encoding a
// message.
func ChatTokens(model string, messages []openai.ChatCompletionMessage) (int, error) {
	codec, err := getCodec(model)
	if err != nil {
		return 0, err
	}

	var (
		perMessage = 3
		perName    int
	)

	if strings.HasPrefix(model, "gpt-4") {
		perName = 1
	} else if strings.HasPrefix(model, "gpt-3.5") {
		perName = -1
	}

	tokens := 3
	for _, message := range messages {
		tokens += perMessage
		if message.Name != "" {
			tokens += perName
		}

		toks, _, err := codec.Encode(message.Content)
		if err != nil {
			return tokens, fmt.Errorf("encode message: %w", err)
		}
		tokens += len(toks)
	}

	return tokens + 1, nil
}

func getCodec(model string) (tokenizer.Codec, error) {
	codecsMux.Lock()
	defer codecsMux.Unlock()

	codec, ok := codecs[model]
	if !ok {
		var err error
		if codec, err = internal.OpenAITokenizer(model); err != nil {
			return nil, err
		}
		codecs[model] = codec
	}

	return codec, nil
}

// PromptTokens calculates the number of tokens in a given prompt using the
// specified model's tokenizer. It returns the token count and any error
// encountered during the tokenization process. If an error occurs, the token
// count is not guaranteed to be accurate.
func PromptTokens(model string, prompt string) (int, error) {
	codec, err := getCodec(model)
	if err != nil {
		return 0, err
	}
	toks, _, err := codec.Encode(prompt)
	return len(toks), err
}
