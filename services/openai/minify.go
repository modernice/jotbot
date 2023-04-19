package openai

import (
	"fmt"

	"github.com/dave/dst/decorator"
	"github.com/modernice/opendocs/internal/nodes"
	"github.com/tiktoken-go/tokenizer"
)

var DefaultMinification = [...]nodes.MinifyOptions{
	nodes.MinifyUnexported,
	{
		FuncBody: true,
		Exported: true,
	},
	nodes.MinifyExported,
	nodes.MinifyAll,
}

type SourceTooLarge struct {
	MaxTokens      int
	MinifiedTokens int
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

type MinifyOptions struct {
	MaxTokens int
	Prepend   string
	Steps     []nodes.MinifyOptions
}

func Minify(code []byte, maxTokens int) (Minification, []Minification, error) {
	return MinifyOptions{MaxTokens: maxTokens}.Minify(code)
}

func (opts MinifyOptions) Minify(code []byte) (Minification, []Minification, error) {
	if len(opts.Steps) == 0 {
		opts.Steps = DefaultMinification[:]
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

	var prependLen int
	if opts.Prepend != "" {
		ids, _, err := codec.Encode(string(opts.Prepend))
		if err != nil {
			return Minification{}, nil, fmt.Errorf("tiktoken: encode prepended text: %w", err)
		}
		prependLen = len(ids)
	}

	ids, _, err := codec.Encode(string(code))
	if err != nil {
		return Minification{}, nil, fmt.Errorf("tiktoken: encode code: %w", err)
	}

	total := prependLen + len(ids)

	if total <= opts.MaxTokens {
		min := Minification{
			Input:    code,
			Minified: code,
			Tokens:   ids,
			Options:  nodes.MinifyNone,
		}
		return min, []Minification{min}, nil
	}

	for _, s := range DefaultMinification {
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

		if len(ids) <= opts.MaxTokens {
			return min, msteps, nil
		}
	}

	var min Minification
	if len(opts.Steps) > 0 {
		min = msteps[len(opts.Steps)-1]
	}

	return min, msteps, &SourceTooLarge{
		MaxTokens:      opts.MaxTokens,
		MinifiedTokens: len(min.Tokens),
	}
}
