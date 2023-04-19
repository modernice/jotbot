package openai

import (
	"fmt"

	"github.com/dave/dst/decorator"
	"github.com/modernice/opendocs/internal/nodes"
	"github.com/tiktoken-go/tokenizer"
)

type SourceTooLarge struct {
	MaxTokens      uint
	MinifiedTokens uint
}

func (err *SourceTooLarge) Error() string {
	return fmt.Sprintf("source code is too large to be minified to %d tokens. minified code has %d tokens", err.MaxTokens, err.MinifiedTokens)
}

type Minification struct {
	Input    []byte
	Minified []byte

	// Tokens is the number of tokens in the minified code.
	Tokens  []uint
	Options nodes.MinifyOptions
}

var minificationSteps = [...]nodes.MinifyOptions{
	nodes.MinifyUnexported,
	nodes.MinifyExported,
	nodes.MinifyAll,
}

func Minify(code []byte, maxTokens uint, steps ...nodes.MinifyOptions) (Minification, []Minification, error) {
	if len(steps) == 0 {
		steps = minificationSteps[:]
	}

	var msteps []Minification

	node, err := decorator.Parse(code)
	if err != nil {
		return Minification{}, nil, fmt.Errorf("parse code: %w", err)
	}

	codec, err := tokenizer.Get(tokenizer.Cl100kBase)
	if err != nil {
		return Minification{}, nil, fmt.Errorf("get tokenizer: %w", err)
	}

	ids, _, err := codec.Encode(string(code))
	if err != nil {
		return Minification{}, nil, fmt.Errorf("tiktoken: encode code: %w", err)
	}

	if uint(len(ids)) <= maxTokens {
		min := Minification{
			Input:    code,
			Minified: code,
			Tokens:   ids,
			Options:  nodes.MinifyNone,
		}
		return min, []Minification{min}, nil
	}

	for _, s := range minificationSteps {
		input, err := nodes.Format(node)
		if err != nil {
			return Minification{}, nil, fmt.Errorf("format code: %w", err)
		}

		node = nodes.Minify(node, s)

		minified, err := nodes.Format(node)
		if err != nil {
			return Minification{}, nil, fmt.Errorf("format minified code: %w", err)
		}

		ids, _, err := codec.Encode(string(minified))
		if err != nil {
			return Minification{}, nil, fmt.Errorf("tiktoken: encode minified code: %w", err)
		}

		min := Minification{
			Input:    input,
			Minified: minified,
			Tokens:   ids,
			Options:  s,
		}

		msteps = append(msteps, min)

		if uint(len(ids)) <= maxTokens {
			return min, msteps, nil
		}
	}

	var min Minification
	if len(steps) > 0 {
		min = msteps[len(steps)-1]
	}

	return min, msteps, &SourceTooLarge{
		MaxTokens:      maxTokens,
		MinifiedTokens: uint(len(min.Tokens)),
	}
}
