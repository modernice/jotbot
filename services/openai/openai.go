package openai

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/tiktoken-go/tokenizer"
	"golang.org/x/exp/slog"
)

const (
	DefaultModel     = openai.GPT3Dot5Turbo
	DefaultMaxTokens = 512
)

type Service struct {
	client    *openai.Client
	model     tokenizer.Model
	maxTokens int
	codec     tokenizer.Codec
	log       *slog.Logger
}

type Option func(*Service)

func Model(model tokenizer.Model) Option {
	return func(s *Service) {
		s.model = model
	}
}

func Client(c *openai.Client) Option {
	return func(s *Service) {
		s.client = c
	}
}

func New(apiKey string, opts ...Option) (*Service, error) {
	var svc Service
	for _, opt := range opts {
		opt(&svc)
	}
	if svc.client == nil {
		svc.client = openai.NewClient(apiKey)
	}

	if svc.model == "" {
		svc.log.Debug(fmt.Sprintf("[OpenAI] No model provided. Using default model %q", DefaultModel))
		svc.model = DefaultModel
	}

	codec, err := tokenizer.ForModel(svc.model)
	if err != nil {
		return nil, fmt.Errorf("get tokenizer for model %q: %w", svc.model, err)
	}
	svc.codec = codec

	return &svc, nil
}

func (svc *Service) GenerateDoc(ctx context.Context, prompt string) (string, error) {
	// TODO(bounoable): find optimal values for these parameters
	req := openai.CompletionRequest{
		Model:            string(svc.model),
		Temperature:      0.618,
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

func (svc *Service) createWithGPT(ctx context.Context, req openai.CompletionRequest) (result, error) {
	maxTokens, err := svc.maxGPTTokens(req.Prompt.(string))
	if err != nil {
		return result{}, fmt.Errorf("max tokens: %w", err)
	}
	req.MaxTokens = maxTokens

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
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: req.Prompt.(string),
		},
	}

	maxTokens, err := svc.maxChatTokens(messages)
	if err != nil {
		return result{}, fmt.Errorf("max tokens: %w", err)
	}

	resp, err := svc.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:            openai.GPT3Dot5Turbo,
		Temperature:      req.Temperature,
		MaxTokens:        maxTokens,
		PresencePenalty:  req.PresencePenalty,
		FrequencyPenalty: req.FrequencyPenalty,
		Messages:         messages,
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

func (svc *Service) maxGPTTokens(prompt string) (int, error) {
	promptTokens, err := PromptTokens(svc.model, prompt)
	if err != nil {
		return 0, fmt.Errorf("compute tokens for prompt: %w", err)
	}

	maxTokensForModel, ok := modelMaxTokens[string(svc.model)]
	if !ok {
		maxTokensForModel = modelMaxTokens["default"]
	}

	remaining := maxTokensForModel - promptTokens

	maxTokens := int(math.Min(float64(svc.maxTokens), float64(maxTokensForModel)))
	maxTokens = int(math.Min(float64(maxTokens), float64(remaining)))
	if maxTokens < 0 {
		maxTokens = 0
	}

	return maxTokens, nil
}

func (svc *Service) maxChatTokens(messages []openai.ChatCompletionMessage) (int, error) {
	promptTokens, err := ChatTokens(svc.model, messages)
	if err != nil {
		return 0, fmt.Errorf("compute tokens for chat messages: %w", err)
	}

	maxTokensForModel, ok := modelMaxTokens[string(svc.model)]
	if !ok {
		maxTokensForModel = modelMaxTokens["default"]
	}

	remaining := maxTokensForModel - promptTokens

	maxTokens := int(math.Min(float64(svc.maxTokens), float64(maxTokensForModel)))
	maxTokens = int(math.Min(float64(maxTokens), float64(remaining)))
	if maxTokens < 0 {
		maxTokens = 0
	}

	return maxTokens, nil
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

func MaxTokensForModel(model string) int {
	if t, ok := modelMaxTokens[model]; ok {
		return t
	}
	return modelMaxTokens["default"]
}

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

type result struct {
	finishReason string
	text         string
}

func (r *result) normalize() {
	r.text = strings.TrimSpace(r.text)
}
