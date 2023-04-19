package openai

import (
	"fmt"

	"github.com/dave/dst/decorator"
	"github.com/modernice/opendocs/internal/nodes"
	"github.com/tiktoken-go/tokenizer"
)

// DefaultMinification is a variable that contains an array of MinifyOptions.
// These options are used as the default steps for the Minify function. The
// Minify function takes a byte slice of code and a maximum number of tokens as
// input, and returns a Minification struct, a slice of Minification structs,
// and an error. The Minification struct contains the input code, the minified
// code, the number of tokens in the minified code, and the MinifyOptions used
// to minify the code. The MinifyOptions struct contains the maximum number of
// tokens, a model string, a prepend string, and a slice of MinifyOptions.
var DefaultMinification = [...]nodes.MinifyOptions{
	nodes.MinifyUnexported,
	{
		FuncBody: true,
		Exported: true,
	},
	nodes.MinifyExported,
	nodes.MinifyAll,
}

// SourceTooLarge is a type that represents an error when the source code is too
// large to be minified to a certain number of tokens. It contains the maximum
// number of tokens allowed and the number of tokens in the minified code. The
// Error method returns a string that describes the error.
type SourceTooLarge struct {
	MaxTokens      int
	MinifiedTokens int
}

// Error *SourceTooLarge.Error represents an error that occurs when the source
// code is too large to be minified to a specified number of tokens. The error
// message includes the maximum number of tokens and the number of tokens in the
// minified code.
func (err *SourceTooLarge) Error() string {
	return fmt.Sprintf("source code is too large to be minified to %d tokens. minified code has %d tokens", err.MaxTokens, err.MinifiedTokens)
}

// Minification is a package that provides functions for minifying source code.
// The Minify function takes a byte slice of source code and a maximum number of
// tokens, and returns a Minification struct containing the minified code, the
// number of tokens in the minified code, and the MinifyOptions used to generate
// the minified code. The MinifyOptions struct allows for customization of the
// minification process, including setting a maximum number of tokens,
// specifying a model for tokenization, and providing text to prepend to the
// source code. If the source code cannot be minified to the specified number of
// tokens, an error is returned.
type Minification struct {
	Input    []byte
	Minified []byte

	// Tokens is the number of tokens in the minified code.
	Tokens  []uint
	Options nodes.MinifyOptions
}

// MinifyOptions represents the options for minifying code. It contains the
// maximum number of tokens allowed in the minified code, the name of the model
// to use for tokenization, a string to prepend to the code before minification,
// and a slice of MinifyOptions to apply in order.
type MinifyOptions struct {
	MaxTokens int
	Model     string
	Prepend   string
	Steps     []nodes.MinifyOptions
}

<<<<<<< Updated upstream
// Minify is a function that takes a byte slice of code and a maximum number of
// tokens, and returns a Minification struct, a slice of Minification structs,
// and an error. The Minification struct contains the input code, the minified
// code, the number of tokens in the minified code, and the MinifyOptions used.
// The MinifyOptions struct contains the maximum number of tokens, the model to
// use for tokenization, a string to prepend to the code, and a slice of
// nodes.MinifyOptions to use for minification. If the code has fewer tokens
// than the maximum number of tokens, Minify returns a Minification struct with
// the input code as the minified code. If the code has more tokens than the
// maximum number of tokens, Minify applies each nodes.MinifyOptions in the
// slice of MinifyOptions until the minified code has fewer tokens than the
// maximum number of tokens. If no MinifyOptions in the slice result in a
// minified code with fewer tokens than the maximum number of tokens, Minify
// returns an error of type *SourceTooLarge.
=======
// Minify minifies a byte slice of source code to a specified number of tokens.
// It takes a byte slice of code and a maximum number of tokens as input, and
// returns a Minification struct, a slice of Minification structs, and an error.
// The Minification struct contains the input code, the minified code, the
// number of tokens in the minified code, and the
// [MinifyOptions](#MinifyOptions) used to minify the code. The MinifyOptions
// struct allows for customization of the minification process, including
// setting a maximum number of tokens, specifying a model for tokenization, and
// providing text to prepend to the source code when counting tokens. If the
// source code cannot be minified to the specified number of tokens, an error of
// type *SourceTooLarge is returned.
>>>>>>> Stashed changes
func Minify(code []byte, maxTokens int) (Minification, []Minification, error) {
	return MinifyOptions{MaxTokens: maxTokens}.Minify(code)
}

// MinifyOptions.Minify minifies the given code using the options specified in
// the MinifyOptions receiver. It returns a Minification struct containing the
// input code, the minified code, the number of tokens in the minified code, and
// the MinifyOptions used. If the minified code exceeds the maximum number of
// tokens specified in the receiver, it returns a slice of Minification structs
// representing each step of the minification process and an error of type
// *SourceTooLarge.
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
