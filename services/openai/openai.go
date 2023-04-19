package openai

import (
	"context"
	"fmt"
	"strings"

	"github.com/modernice/opendocs/generate"
	"github.com/modernice/opendocs/internal"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/exp/slog"
)

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

type Service struct {
	client       *openai.Client
	model        string
	maxTokens    int
	maxDocTokens int
	minifyTokens int
	log          *slog.Logger
}

type Option func(*Service)

func WithLogger(h slog.Handler) Option {
	return func(s *Service) {
		s.log = slog.New(h)
	}
}

func WithClient(c *openai.Client) Option {
	return func(s *Service) {
		s.client = c
	}
}

func Model(model string) Option {
	return func(s *Service) {
		s.model = model
	}
}

func MaxTokens(maxTokens int) Option {
	return func(s *Service) {
		s.maxDocTokens = maxTokens
	}
}

func New(apiKey string, opts ...Option) *Service {
	client := openai.NewClient(apiKey)
	return NewFrom(append([]Option{WithClient(client)}, opts...)...)
}

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
	}.Minify(code)

	if err != nil {
		return "", fmt.Errorf("minify code: %w", err)
	}

	code = result.Minified
	prompt = prompt + string(code)

	if len(steps) > 1 {
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
		"Write the documentation for %q in GoDoc format, with references to symbols wrapped within brackets. Provide only the documentation, excluding the input code and examples. Begin the first sentence with %q. Maintain brevity without sacrificing specificity. Write in the style of the Go library documentations. Do not link to any websites. Here is the source code for %q:\n",
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
