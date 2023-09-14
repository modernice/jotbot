package openai

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/modernice/jotbot/generate"
	"github.com/modernice/jotbot/internal"
	"github.com/sashabaranov/go-openai"
	"github.com/tiktoken-go/tokenizer"
	"golang.org/x/exp/slog"
)

const (
	// DefaultModel is the default AI model used for generating text when no model
	// is explicitly provided.
	DefaultModel = openai.GPT3Dot5Turbo

	// DefaultMaxTokens is the default maximum number of tokens to be generated by
	// the Service instance when no specific value is provided.
	DefaultMaxTokens = 512
)

// Service is a struct that represents an OpenAI service, providing
// functionality to generate documents using the OpenAI API. It contains
// configurable options such as client, model, maxTokens, codec, and logger. It
// also provides methods to generate documents based on provided context and
// request configurations.
type Service struct {
	client    *openai.Client
	model     string
	maxTokens int
	codec     tokenizer.Codec
	log       *slog.Logger
}

// Option is a function type used to configure a Service instance. It takes a
// pointer to a Service and modifies its fields as needed. This allows for
// flexible and customizable configuration of the Service through the use of
// various Option functions.
type Option func(*Service)

// Model is an Option function that sets the model string for the Service
// struct. It returns an Option that modifies the model field of a given Service
// instance.
func Model(model string) Option {
	return func(s *Service) {
		s.model = model
	}
}

// Client sets the OpenAI client for the service to use when making API
// requests. It is an option function for configuring a new Service instance.
func Client(c *openai.Client) Option {
	return func(s *Service) {
		s.client = c
	}
}

// MaxTokens sets the maximum number of tokens to be generated by the Service
// instance. It is an option function for configuring a new Service instance.
func MaxTokens(max int) Option {
	return func(s *Service) {
		s.maxTokens = max
	}
}

// WithLogger sets the logger for the Service instance using the provided
// slog.Handler.
func WithLogger(h slog.Handler) Option {
	return func(s *Service) {
		s.log = slog.New(h)
	}
}

// New creates a new instance of the Service with the specified API key and
// options. It initializes the OpenAI client, sets the default model, and
// configures the tokenizer for the selected model. If no model is provided, it
// uses the default model.
func New(apiKey string, opts ...Option) (*Service, error) {
	svc := Service{maxTokens: DefaultMaxTokens}
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
	svc.log.Debug(fmt.Sprintf("[OpenAI] Using model %q", svc.model))

	codec, err := internal.OpenAITokenizer(svc.model)
	if err != nil {
		return nil, fmt.Errorf("get tokenizer for model %q: %w", svc.model, err)
	}
	svc.codec = codec

	if svc.log == nil {
		svc.log = internal.NopLogger()
	}

	return &svc, nil
}

// GenerateDoc generates a document using the specified generate.Context,
// invoking the OpenAI API with the appropriate model and options. The resulting
// document is returned as a string.
func (svc *Service) GenerateDoc(ctx generate.Context) (string, error) {
	svc.log.Debug(fmt.Sprintf("[OpenAI] Generating docs for %s (%s)", ctx.Input().Identifier, ctx.Input().Language))

	req := svc.makeBaseRequest(ctx)

	generate := svc.useModel(req.Model)

	// TODO(bounoable): Make timeout configurable
	timeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result, err := generate(timeout, req)
	if err != nil {
		return "", err
	}
	result.normalize()

	return result.text, nil
}

func (svc *Service) makeBaseRequest(ctx generate.Context) openai.CompletionRequest {
	req := openai.CompletionRequest{
		Model:            string(svc.model),
		Temperature:      0.618,
		TopP:             0.3,
		PresencePenalty:  0.2,
		FrequencyPenalty: 0.3,
		Prompt:           ctx.Prompt(),
	}

	return req
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
		Model:            req.Model,
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
		finishReason: string(choice.FinishReason),
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

func isChatModel(model string) bool {
	return strings.HasPrefix(model, "gpt-")
}

// MaxTokensForModel returns the maximum number of tokens allowed for the
// specified model. If the model is not recognized, it returns a default value.
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
	openai.GPT3Dot5Turbo16K:   16384,
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
