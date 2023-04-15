package openai

import (
	"context"
	"fmt"
	"strings"

	"github.com/modernice/opendocs/go/generate"
	"github.com/sashabaranov/go-openai"
)

var _ generate.Service = (*Service)(nil)

var systemPrompt = `You are DocGPT, a code documentation writer.` +
	`You are given a file list, a file name, the source code of that file, and an identifier. ` +
	`Using these, you will write the documentation for the type or function identified by the identifier. ` +
	`You will write the documentation in GoDoc format, without the orginal source code.`

type Service struct {
	client *openai.Client
}

func New(client *openai.Client) *Service {
	return &Service{
		client: client,
	}
}

func (g *Service) GenerateDoc(ctx generate.Context) (string, error) {
	files := ctx.Files()
	file := ctx.File()
	identifier := ctx.Identifier()
	code, err := ctx.Read(file)
	if err != nil {
		return "", err
	}

	choice, err := g.createCompletion(ctx, files, file, identifier, code)
	if err != nil {
		return "", fmt.Errorf("create completion: %w", err)
	}

	return choice.Message.Content, nil
}

func (g *Service) createCompletion(ctx context.Context, files []string, file, identifier string, code []byte) (openai.ChatCompletionChoice, error) {
	var zero openai.ChatCompletionChoice

	filesPrompt := filesPrompt(files)

	msg, err := prompt(file, identifier, code)
	if err != nil {
		return zero, fmt.Errorf("build prompt: %w", err)
	}

	resp, err := g.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		// Temperature:      0,
		// TopP:            0,
		MaxTokens: 512,
		// PresencePenalty:  0,
		// FrequencyPenalty: 0,
		// LogitBias: map[string]int{},
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: filesPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: msg,
			},
		},
	})
	if err != nil {
		return zero, fmt.Errorf("create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return zero, fmt.Errorf("openai: no choices returned")
	}

	choice := resp.Choices[0]
	if choice.FinishReason != "stop" {
		return choice, fmt.Errorf("openai: unexpected finish reason: %q", choice.FinishReason)
	}

	if choice.Message.Role != openai.ChatMessageRoleAssistant {
		return choice, fmt.Errorf("openai: unexpected message role in answer: %q", choice.Message.Role)
	}

	return choice, nil
}

func filesPrompt(files []string) string {
	var sb strings.Builder
	sb.WriteString("Files:")

	for _, f := range files {
		sb.WriteString("\n- ")
		sb.WriteString(f)
	}

	return sb.String()
}

func prompt(file, identifier string, code []byte) (string, error) {
	return "Write the documentation for the \"" + identifier + "\" type or function in GoDoc format. " +
		"This is the source code of the \"" + file + "\" file:\n\n" + string(code), nil
}
