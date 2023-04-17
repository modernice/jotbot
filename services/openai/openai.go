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

const DefaultModel = openai.GPT3TextDavinci003

var _ generate.Service = (*Service)(nil)

var systemPrompt = `You are DocGPT, a code documentation writer.` +
	`You are given a a file name, the source code of that file, and an identifier. ` +
	`Using these, you will write the documentation for the type or function identified by the identifier. ` +
	`You will write the documentation in GoDoc format.`

type Service struct {
	client *openai.Client
	model  string
	log    *slog.Logger
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
	return &svc
}

func (svc *Service) GenerateDoc(ctx generate.Context) (string, error) {
	files := ctx.Files()
	file := ctx.File()
	identifier := ctx.Identifier()
	code, err := ctx.Read(file)
	if err != nil {
		return "", err
	}

	answer, err := svc.createCompletion(ctx, files, file, identifier, code)
	if err != nil {
		return "", fmt.Errorf("create completion: %w", err)
	}

	return answer, nil
}

func (svc *Service) createCompletion(
	ctx context.Context,
	files []string,
	file,
	longIdentifier string,
	code []byte,
) (string, error) {
	identifier := normalizeIdentifier(longIdentifier)
	msg := prompt(file, identifier, longIdentifier, code)

	// TODO(bounoable): find optimal values for these parameters
	req := openai.CompletionRequest{
		Model:            svc.model,
		Temperature:      0.1,
		TopP:             0.3,
		MaxTokens:        512,
		PresencePenalty:  0.1,
		FrequencyPenalty: 0.1,
		Prompt:           msg,
	}

	svc.log.Debug("[OpenAI] Creating completion ...", "file", file, "identifier", identifier, "model", req.Model)

	generate := svc.useModel(req.Model)
	answer, err := generate(ctx, req)
	if err != nil {
		return "", err
	}
	answer = normalizeAnswer(answer)

	svc.log.Debug("[OpenAI] Documentation generated", "file", file, "identifier", identifier, "docs", answer)

	return answer, nil
}

func (svc *Service) useModel(model string) func(context.Context, openai.CompletionRequest) (string, error) {
	if isChatModel(model) {
		return svc.createChatCompletion
	}

	return func(ctx context.Context, req openai.CompletionRequest) (string, error) {
		resp, err := svc.client.CreateCompletion(ctx, req)
		if err != nil {
			return "", err
		}

		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("openai: no choices returned")
		}

		choice := resp.Choices[0]
		if choice.FinishReason != "stop" {
			return choice.Text, fmt.Errorf("openai: unexpected finish reason: %q", choice.FinishReason)
		}

		return choice.Text, nil
	}
}

func (svc *Service) createChatCompletion(ctx context.Context, req openai.CompletionRequest) (string, error) {
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
		return "", err
	}

	choice := resp.Choices[0]
	if choice.FinishReason != "stop" {
		return choice.Message.Content, fmt.Errorf("openai: unexpected finish reason: %q", choice.FinishReason)
	}

	if choice.Message.Role != openai.ChatMessageRoleAssistant {
		return choice.Message.Content, fmt.Errorf("openai: unexpected message role in answer: %q", choice.Message.Role)
	}

	return choice.Message.Content, nil
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

func prompt(file, identifier, longIdentifier string, code []byte) string {
	return fmt.Sprintf(
		"Write the documentation for %q in GoDoc format, with references to symbols wrapped within brackets. Provide only the documentation, excluding the input code and examples. Begin the first sentence with %q. Maintain brevity without sacrificing specificity. Write in the style of the Go library documentations. Do not link to any websites. Here is the source code for %q:\n%s",
		longIdentifier,
		fmt.Sprintf("%s ", identifier),
		file,
		string(code),
	)
}

func normalizeAnswer(answer string) string {
	answer = strings.TrimSpace(answer)
	answer = strings.ReplaceAll(answer, "// ", "")
	return answer
}
