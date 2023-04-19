package openai

import (
	"fmt"

	"github.com/dave/dst/decorator"
	"github.com/modernice/opendocs/internal/nodes"
	"github.com/tiktoken-go/tokenizer"
)

// DefaultMinification is an array of MinifyOptions used as the default steps
// for the Minify function. The Minify function takes a byte slice of code and a
// maximum number of tokens as input, and returns a Minification struct, a slice
// of Minification structs, and an error. The Minification struct contains the
// input code, the minified code, the number of tokens in the minified code, and
// the MinifyOptions used to minify the code. The MinifyOptions struct contains
// the maximum number of tokens, a model string, a prepend string, and a slice
// of MinifyOptions.
var DefaultMinification = [...]nodes.MinifyOptions{
	nodes.MinifyUnexported,
	{
		FuncBody: true,
		Exported: true,
	},
	nodes.MinifyExported,
	nodes.MinifyAll,
}

// SourceTooLarge represents an error returned by the Minify function if the
// input source code is too large to be minified to the specified maximum number
// of tokens. The error message contains the maximum number of tokens and the
// number of tokens in the minified code.
type SourceTooLarge struct {
	MaxTokens      int
	MinifiedTokens int
}

// Error *SourceTooLarge is an error type that represents when the source code
// is too large to be minified to the specified maximum number of tokens. The
// error message includes the maximum number of tokens and the number of tokens
// in the minified code.
func (err *SourceTooLarge) Error() string {
	return fmt.Sprintf("source code is too large to be minified to %d tokens. minified code has %d tokens", err.MaxTokens, err.MinifiedTokens)
}

// Minification provides functions and types for minifying source code. The
// Minify function takes a byte slice of code and a maximum number of tokens as
// input, and returns a Minification struct, a slice of Minification structs,
// and an error. The Minification struct contains the input code, the minified
// code, the number of tokens in the minified code, and the MinifyOptions used
// to minify the code. The MinifyOptions struct contains the maximum number of
// tokens, a model string, a prepend string, and a slice of MinifyOptions.
// DefaultMinification is a variable that contains an array of MinifyOptions
// used as the default steps for the Minify function.
type Minification struct {
	Input    []byte
	Minified []byte

	// Tokens is the number of tokens in the minified code.
	Tokens  []uint
	Options nodes.MinifyOptions
}

// MinifyOptions is a struct that contains options for the Minify function.
// These options include the maximum number of tokens, a model string, a prepend
// string, a force flag, and a slice of MinifyOptions.
type MinifyOptions struct {
	MaxTokens int
	Model     string
	Prepend   string
	Force     bool
	Steps     []nodes.MinifyOptions
}

// Minify is a function that takes a byte slice of code and a maximum number of
// tokens as input, and returns a Minification struct, a slice of Minification
// structs, and an error. The Minification struct contains the input code, the
// minified code, the number of tokens in the minified code, and the
// MinifyOptions used to minify the code. The MinifyOptions struct contains the
// maximum number of tokens, a model string, a prepend string, and a slice of
// MinifyOptions. DefaultMinification is a variable that contains an array of
// MinifyOptions. These options are used as the default steps for the Minify
// function.
func Minify(code []byte, maxTokens int) (Minification, []Minification, error) {
	return MinifyOptions{MaxTokens: maxTokens}.Minify(code)
}

// MinifyOptions.Minify minifies a given byte slice of code using the
// MinifyOptions struct. It returns a Minification struct, a slice of
// Minification structs, and an error. The Minification struct contains the
// input code, the minified code, the number of tokens in the minified code, and
// the MinifyOptions used to minify the code. The MinifyOptions struct contains
// the maximum number of tokens, a model string, a prepend string, and a slice
// of MinifyOptions. If the minified code has more tokens than the maximum
// number specified in the MinifyOptions struct, it returns an error of type
// SourceTooLarge.
func (opts MinifyOptions) Minify(code []byte) (Minification, []Minification, error) {
	if len(opts.Steps) == 0 {
		opts.Steps = DefaultMinification[:]
	}

	if opts.Model == "" {
		opts.Model = string(DefaultModel)
	}

	var msteps []Minification

	node, err := decorator.Parse(code)
	if err != nil {
		return Minification{}, nil, fmt.Errorf("parse code: %w", err)
	}

	codec, err := tokenizer.ForModel(tokenizer.Model(opts.Model))
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
	if !opts.Force && total <= opts.MaxTokens {
		min := Minification{
			Input:    code,
			Minified: code,
			Tokens:   ids,
			Options:  nodes.MinifyNone,
		}
		return min, []Minification{min}, nil
	}

	for _, s := range opts.Steps {
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

		if !opts.Force && len(ids) <= opts.MaxTokens {
			return min, msteps, nil
		}
	}

	var min Minification
	if len(opts.Steps) > 0 {
		min = msteps[len(opts.Steps)-1]
	}

	err = nil
	if !opts.Force {
		err = &SourceTooLarge{
			MaxTokens:      opts.MaxTokens,
			MinifiedTokens: len(min.Tokens),
		}
	}

	return min, msteps, err
}
