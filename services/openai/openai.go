package openai

import (
	"context"
	"fmt"
	"strings"

	"github.com/modernice/opendocs/generate"
	"github.com/modernice/opendocs/internal"
	"github.com/modernice/opendocs/internal/nodes"
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

// Service represents a service that generates documentation for Go code using
// OpenAI's GPT-3 language model. It provides a GenerateDoc method that takes a
// generate.Context and returns the generated documentation as a string. The
// service can be configured with options such as WithLogger, WithClient, Model,
// and MaxTokens.
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

// Option is a type that represents a configuration option for the Service type.
// It is a function that takes a pointer to a Service and modifies its fields.
// The available options are WithLogger, WithClient, Model, and MaxTokens.
type Option func(*Service)

// WithLogger is an Option for the Service type that sets the logger for the
// OpenAI service. The logger is used to output debug information during the
// generation of documentation.
func WithLogger(h slog.Handler) Option {
	return func(s *Service) {
		s.log = slog.New(h)
	}
}

// WithClient sets the OpenAI client for the Service.
func WithClient(c *openai.Client) Option {
	return func(s *Service) {
		s.client = c
	}
}

// Model represents a service that generates documentation for a given code file
// and identifier. It uses the OpenAI GPT-3 language model to generate the
// documentation. The model can be set using the Model option, and the maximum
// number of tokens used by the model can be set using the MaxTokens option. The
// GenerateDoc method takes a generate.Context and returns the generated
// documentation as a string in GoDoc format.
func Model(model string) Option {
	return func(s *Service) {
		s.model = model
	}
}

// MaxTokens sets the maximum number of tokens to be used in generating the
// documentation. The default value is 512.
func MaxTokens(maxTokens int) Option {
	return func(s *Service) {
		s.maxDocTokens = maxTokens
	}
}

func MinifyWith(steps []nodes.MinifyOptions, force bool) Option {
	return func(s *Service) {
		s.minifySteps = steps
		s.forceMinify = force
	}
}

// New creates a new instance of Service with the given API key and options. It
// returns a pointer to the new Service. If no options are provided, it uses the
// default model and max tokens.
func New(apiKey string, opts ...Option) *Service {
	client := openai.NewClient(apiKey)
	return NewFrom(append([]Option{WithClient(client)}, opts...)...)
}

// NewFrom creates a new instance of Service with the given options. It returns
// a pointer to the new Service.
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

// GenerateDoc generates documentation for a given type or function identified
// by the input identifier. The documentation is written in GoDoc format and
// excludes input code and examples. The first sentence of the documentation
// begins with the identifier.
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
		Model:            svc.model,
		Temperature:      0.1,
		TopP:             0.3,
		MaxTokens:        512,
		PresencePenalty:  0.1,
		FrequencyPenalty: 0.1,
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
		"Write a concise documentation for %q in GoDoc format, with references to symbols wrapped within brackets. Provide only the documentation, excluding input code and examples. Do not link to any websites. Write in the style of the Go library documentations. Begin exactly with %q. This is the source code for %q:\n",
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
