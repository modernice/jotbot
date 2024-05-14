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

// ChatTokens calculates the total number of tokens used by a list of
// [openai.ChatCompletionMessage] for a given model.
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

// PromptTokens calculates the number of tokens in a given prompt based on the
// specified model and returns the token count along with any error encountered
// during encoding.
func PromptTokens(model string, prompt string) (int, error) {
	codec, err := getCodec(model)
	if err != nil {
		return 0, err
	}
	toks, _, err := codec.Encode(prompt)
	return len(toks), err
}
