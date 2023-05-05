package openai

import (
	"fmt"
	"sync"

	"github.com/sashabaranov/go-openai"
	"github.com/tiktoken-go/tokenizer"
)

var (
	codecsMux sync.Mutex
	codecs    = make(map[string]tokenizer.Codec)
)

func ChatTokens(model string, messages []openai.ChatCompletionMessage) (int, error) {
	codec, err := getCodec(model)
	if err != nil {
		return 0, err
	}

	var (
		perMessage int
		perName    int
	)
	switch tokenizer.Model(model) {
	case tokenizer.GPT4:
		perMessage = 3
		perName = 1
		break
	case tokenizer.GPT35Turbo:
		perMessage = 3
		perName = -1
		break
	default:
		return 0, fmt.Errorf("unsupported model %q", model)
	}

	tokens := 3 // every reply is primed with <|start|>assistant<|message|>
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

	return tokens, nil
}

func getCodec(model string) (tokenizer.Codec, error) {
	codecsMux.Lock()
	defer codecsMux.Unlock()

	codec, ok := codecs[model]
	if !ok {
		var err error
		if codec, err = tokenizer.ForModel(tokenizer.Model(model)); err != nil {
			return nil, err
		}
		codecs[model] = codec
	}

	return codec, nil
}

func PromptTokens(model string, prompt string) (int, error) {
	codec, err := getCodec(model)
	if err != nil {
		return 0, err
	}
	toks, _, err := codec.Encode(prompt)
	return len(toks), err
}
