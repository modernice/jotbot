package openai

import (
	"context"
	"fmt"
	"strings"

	"github.com/modernice/opendocs/go/generate"
	"github.com/modernice/opendocs/go/internal"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/exp/slog"
)

var _ generate.Service = (*Service)(nil)

var systemPrompt = `You are DocGPT, a code documentation writer.` +
	`You are given a file list, a file name, the source code of that file, and an identifier. ` +
	`Using these, you will write the documentation for the type or function identified by the identifier. ` +
	`You will write the documentation in GoDoc format.`

type Service struct {
	client *openai.Client
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

func New(apiKey string, opts ...Option) *Service {
	client := openai.NewClient(apiKey)
	return NewFrom(append([]Option{WithClient(client)}, opts...)...)
}

func NewFrom(opts ...Option) *Service {
	var svc Service
	for _, opt := range opts {
		opt(&svc)
	}
	if svc.log == nil {
		svc.log = internal.NopLogger()
	}
	return &svc
}

func (g *Service) GenerateDoc(ctx generate.Context) (string, error) {
	files := ctx.Files()
	file := ctx.File()
	identifier := ctx.Identifier()
	code, err := ctx.Read(file)
	if err != nil {
		return "", err
	}

	answer, err := g.createCompletion(ctx, files, file, identifier, code)
	if err != nil {
		return "", fmt.Errorf("create completion: %w", err)
	}

	return answer, nil
}

func (g *Service) createCompletion(
	ctx context.Context,
	files []string,
	file,
	longIdentifier string,
	code []byte,
) (string, error) {
	// filesPrompt := filesPrompt(files)

	identifier := normalizeIdentifier(longIdentifier)
	msg := prompt(file, identifier, longIdentifier, code)

	g.log.Debug("[OpenAI] Creating chat completion ...", "file", file, "identifier", identifier)

	resp, err := g.client.CreateCompletion(ctx, openai.CompletionRequest{
		Model:            openai.GPT3TextDavinci003,
		TopP:             0.2,
		MaxTokens:        512,
		PresencePenalty:  0.1,
		FrequencyPenalty: 0.1,
		Prompt:           msg,
	})

	// resp, err := g.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
	// 	Model:            openai.GPT3Dot5Turbo,
	// 	Temperature:      0.618,
	// 	MaxTokens:        512,
	// 	PresencePenalty:  0.1,
	// 	FrequencyPenalty: 0.2,
	// 	Messages: []openai.ChatCompletionMessage{
	// 		{
	// 			Role:    openai.ChatMessageRoleSystem,
	// 			Content: systemPrompt,
	// 		},
	// 		{
	// 			Role:    openai.ChatMessageRoleUser,
	// 			Content: filesPrompt,
	// 		},
	// 		{
	// 			Role:    openai.ChatMessageRoleUser,
	// 			Content: msg,
	// 		},
	// 	},
	// })
	if err != nil {
		return "", fmt.Errorf("create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("openai: no choices returned")
	}

	answer := resp.Choices[0].Text
	answer = normalizeAnswer(answer)

	g.log.Debug("[OpenAI] Documentation generated", "file", file, "identifier", identifier, "docs", answer)

	return answer, nil

	// choice := resp.Choices[0]
	// if choice.FinishReason != "stop" {
	// 	return choice, fmt.Errorf("openai: unexpected finish reason: %q", choice.FinishReason)
	// }

	// if choice.Message.Role != openai.ChatMessageRoleAssistant {
	// 	return choice, fmt.Errorf("openai: unexpected message role in answer: %q", choice.Message.Role)
	// }

	// g.log.Debug("[OpenAI] Documentation generated", "file", file, "identifier", identifier, "docs", choice.Message.Content)

	// return choice, nil
}

func normalizeIdentifier(identifier string) string {
	parts := strings.Split(identifier, ".")
	if len(parts) < 2 {
		return identifier
	}
	return parts[1]
}

// func filesPrompt(files []string) string {
// 	var sb strings.Builder
// 	sb.WriteString("Files:")

// 	for _, f := range files {
// 		sb.WriteString("\n- ")
// 		sb.WriteString(f)
// 	}

// 	return sb.String()
// }

func prompt(file, identifier, longIdentifier string, code []byte) string {
	// "Write a short documentation for %q in idiomatic GoDoc format, with references to symbols wrapped within brackets. Only output the documentation, not the input code. Do not include examples. Do not describe any other symbols besides %q. Keep it as short as possible while not being too unspecific. Start the first sentence with %q. This is the source code of %q:\n%s",

	return fmt.Sprintf(
		"Create a concise documentation for %q in GoDoc format, with references to symbols wrapped within brackets. Provide only the documentation, excluding the input code and examples. Begin the first sentence with %q. Maintain brevity without sacrificing specificity. Here is the source code for %q:\n%s",
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
