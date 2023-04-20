package openai

import (
	"context"
	"fmt"
	"strings"

	"github.com/modernice/jotbot/generate"
	"github.com/modernice/jotbot/internal"
	"github.com/modernice/jotbot/internal/nodes"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/exp/slog"
)

// DefaultModel is a constant that represents the default OpenAI model used by
// the Service type.
const (
	DefaultModel     = openai.GPT3Dot5Turbo
	DefaultMaxTokens = 512
)

var _ generate.Service = (*Service)(nil)

var systemPrompt = `You are DocGPT, a code documentation writer.` +
	`You are given a a file name, the source code of that file, and an identifier. ` +
	`Using these, you will write the documentation for the type or function identified by the identifier. ` +
	`You will write the documentation in GoDoc format.`

var modelMaxTokens = map[string]int{
	"default":                 2049,
	openai.GPT432K0314:        32768,
	openai.GPT432K:            32768,
	openai.GPT40314:           8192,
	openai.GPT4:               8192,
	openai.GPT3Dot5Turbo0301:  4096,
	openai.GPT3Dot5Turbo:      4096,
	openai.GPT3TextDavinci003: 4097,
	openai.GPT3TextDavinci002: 4097,
}

// Service is a type that generates GoDoc documentation using OpenAI's GPT
// language model. It takes a file name and an identifier as input, and returns
// the documentation for the identified symbol in GoDoc format.
type Service struct {
	client       *openai.Client
	model        string
	maxTokens    int
	maxDocTokens int
	minifyTokens int
	minifySteps  []nodes.MinifyOptions
	forceMinify  bool
	log          *slog.Logger
}

// Option represents a function that sets configuration options for a Service
// type. Use WithLogger to set a logger for the service, WithClient to set an
// openai.Client, Model to set the OpenAI model to use, MaxTokens to set the
// maximum number of tokens to use for the generated documentation, and
// MinifyWith to set options for code minification.
type Option func(*Service)

// WithLogger returns an Option that sets the logger for a Service. This logger
// is used to log debug information during documentation generation.
func WithLogger(h slog.Handler) Option {
	return func(s *Service) {
		s.log = slog.New(h)
	}
}

// WithClient creates a new Option that sets the Service's OpenAI client.
func WithClient(c *openai.Client) Option {
	return func(s *Service) {
		s.client = c
	}
}

// Model represents an OpenAI language model used by the Service type. It
// contains a string identifying the model to be used, as well as maximum token
// limits for documentation generation. DefaultModel is a constant that
// represents the default OpenAI model used by the Service type.
func Model(model string) Option {
	return func(s *Service) {
		s.model = model
	}
}

// MaxTokens represents the default maximum number of tokens used by the Service
// type for generating documentation.
func MaxTokens(maxTokens int) Option {
	return func(s *Service) {
		s.maxDocTokens = maxTokens
	}
}

// MinifyWith minifies the source code by applying the MinifyOptions specified
// as argument, and returns the resulting code. If force is set to true,
// MinifyWith will always minify the code.
func MinifyWith(steps []nodes.MinifyOptions, force bool) Option {
	return func(s *Service) {
		s.minifySteps = steps
		s.forceMinify = force
	}
}

// New creates a new instance of the Service type, which provides a function for
// generating documentation for Go code using the OpenAI API. It takes an API
// key as a string and zero or more Option functions that configure the Service
// instance.
func New(apiKey string, opts ...Option) *Service {
	client := openai.NewClient(apiKey)
	return NewFrom(append([]Option{WithClient(client)}, opts...)...)
}

// NewFrom creates a new Service with the given options. It takes a variadic
// list of Options that can be used to set the Service's client, model,
// maxTokens, maxDocTokens, minifySteps, and forceMinify. If no model is
// specified, the Service will use DefaultModel. If no maxDocTokens is
// specified, the Service will use DefaultMaxTokens.
func NewFrom(opts ...Option) *Service {
	var svc Service
	for _, opt := range opts {
		opt(&svc)
	}

	if svc.model == "" {
		svc.model = DefaultModel
	}

	if svc.log == nil {
		svc.log = internal.NopLogger()
	}

	if svc.maxDocTokens == 0 {
		svc.maxDocTokens = DefaultMaxTokens
	}

	svc.maxTokens = modelMaxTokens[svc.model]
	if svc.maxTokens == 0 {
		svc.maxTokens = modelMaxTokens["default"]
	}
	svc.minifyTokens = svc.maxTokens - svc.maxDocTokens

	return &svc
}

// GenerateDoc generates the documentation for a Go type or function using
// OpenAI's GPT-3 language model. It accepts a generate.Context and returns the
// resulting documentation as a string.
func (svc *Service) GenerateDoc(ctx generate.Context) (string, error) {
	file := ctx.File()
	longIdentifier := ctx.Identifier()

	code, err := ctx.Read(file)
	if err != nil {
		return "", err
	}

	identifier := normalizeIdentifier(longIdentifier)
	prompt := promptWithoutCode(file, identifier, longIdentifier)

	result, steps, err := MinifyOptions{
		MaxTokens: svc.minifyTokens,
		Model:     svc.model,
		Prepend:   prompt,
		Steps:     svc.minifySteps,
		Force:     svc.forceMinify,
	}.Minify(code)

	if err != nil {
		return "", fmt.Errorf("minify code: %w", err)
	}

	code = result.Minified
	prompt = prompt + string(code)

	if svc.forceMinify || len(steps) > 1 {
		svc.log.Debug(fmt.Sprintf("[OpenAI] Minified code to %d tokens in %d step(s)", len(result.Tokens), len(steps)))
	} else {
		svc.log.Debug(fmt.Sprintf("[OpenAI] Code has %d tokens. Not minified.", len(result.Tokens)))
	}

	svc.log.Debug("[OpenAI] Generating documentation ...", "file", file, "identifier", identifier, "model", svc.model)

	answer, err := svc.createCompletion(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("create completion: %w", err)
	}

	svc.log.Debug("[OpenAI] Documentation generated", "file", file, "identifier", identifier, "docs", answer)

	return answer, nil
}

func (svc *Service) createCompletion(ctx context.Context, prompt string) (string, error) {
	// TODO(bounoable): find optimal values for these parameters
	req := openai.CompletionRequest{
		Model:       svc.model,
		Temperature: 0.618,
		// TopP:             0.3,
		MaxTokens:        512,
		PresencePenalty:  0.2,
		FrequencyPenalty: 0.35,
		Prompt:           prompt,
	}

	generate := svc.useModel(req.Model)
	result, err := generate(ctx, req)
	if err != nil {
		return "", err
	}
	result.normalize()

	return result.text, nil
}

func (svc *Service) useModel(model string) func(context.Context, openai.CompletionRequest) (result, error) {
	if isChatModel(model) {
		return svc.createWithChat
	}
	return svc.createWithGPT
}

type result struct {
	finishReason string
	text         string
}

func (svc *Service) createWithGPT(ctx context.Context, req openai.CompletionRequest) (result, error) {
	resp, err := svc.client.CreateCompletion(ctx, req)
	if err != nil {
		return result{}, err
	}

	if len(resp.Choices) == 0 {
		return result{}, fmt.Errorf("openai: no choices returned")
	}

	svc.printUsage(resp.Usage)

	choice := resp.Choices[0]

	return result{
		finishReason: choice.FinishReason,
		text:         choice.Text,
	}, nil
}

func (svc *Service) createWithChat(ctx context.Context, req openai.CompletionRequest) (result, error) {
	resp, err := svc.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:            openai.GPT3Dot5Turbo,
		Temperature:      req.Temperature,
		MaxTokens:        req.MaxTokens,
		PresencePenalty:  req.PresencePenalty,
		FrequencyPenalty: req.FrequencyPenalty,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: req.Prompt.(string),
			},
		},
	})
	if err != nil {
		return result{}, err
	}

	svc.printUsage(resp.Usage)

	choice := resp.Choices[0]
	res := result{
		finishReason: choice.FinishReason,
		text:         choice.Message.Content,
	}

	if choice.Message.Role != openai.ChatMessageRoleAssistant {
		return res, fmt.Errorf("openai: unexpected message role in answer: %q", choice.Message.Role)
	}

	return res, nil
}

func (svc *Service) printUsage(usage openai.Usage) {
	svc.log.Debug("[OpenAI] Usage info", "prompt", usage.PromptTokens, "completion", usage.CompletionTokens, "total", usage.TotalTokens)
}

var chatModels = map[string]bool{
	openai.GPT4:              true,
	openai.GPT40314:          true,
	openai.GPT432K:           true,
	openai.GPT432K0314:       true,
	openai.GPT3Dot5Turbo:     true,
	openai.GPT3Dot5Turbo0301: true,
}

func isChatModel(model string) bool {
	return chatModels[model]
}

func normalizeIdentifier(identifier string) string {
	parts := strings.Split(identifier, ".")
	if len(parts) < 2 {
		return identifier
	}
	return parts[1]
}

func promptWithoutCode(file, identifier, longIdentifier string) string {
	return fmt.Sprintf(
		"Write a concise documentation for %q in GoDoc format, ensuring clarity and brevity. Use brackets to enclose symbol references, and start the first sentence with %q. Write in the style of Go library documentation, and avoid linking to external websites. Provide the documentation text only, excluding input code and examples. This is the source code of the file:\n\n%s",
		longIdentifier,
		fmt.Sprintf("%s ", identifier),
		file,
	)
}

func (r *result) normalize() {
	r.text = normalizeAnswer(r.text)
}

func normalizeAnswer(answer string) string {
	answer = strings.TrimSpace(answer)
	answer = strings.ReplaceAll(answer, "// ", "")
	return answer
}
