package anthropic

import (
	"fmt"

	"github.com/liushuangls/go-anthropic/v2"
	"github.com/modernice/jotbot/generate"
	"github.com/modernice/jotbot/internal"
	"golang.org/x/exp/slog"
)

const (
	// DefaultModel specifies the default language model used for generating
	// completions.
	DefaultModel = anthropic.ModelClaudeInstant1Dot2

	// DefaultMaxTokens specifies the default maximum number of tokens for a
	// request.
	DefaultMaxTokens = 512
)

// Service provides functionalities for generating text completions using Anthropic
// models, with options to customize the model, maximum tokens, and logging
// behavior.
type Service struct {
	client    *anthropic.Client
	model     string
	maxTokens int
	log       *slog.Logger
}

// Option modifies the configuration of a [Service].
type Option func(*Service)

// Model sets the model to be used by the [Service].
func Model(model string) Option {
	return func(s *Service) {
		s.model = model
	}
}

// Client sets the [*anthropic.Client] for the [Service].
func Client(c *anthropic.Client) Option {
	return func(s *Service) {
		s.client = c
	}
}

// MaxTokens sets the maximum number of tokens that can be used in a request.
func MaxTokens(max int) Option {
	return func(s *Service) {
		s.maxTokens = max
	}
}

// WithLogger configures the service to use a specified logger. Returns an
// [Option].
func WithLogger(h slog.Handler) Option {
	return func(s *Service) {
		s.log = slog.New(h)
	}
}

// New initializes and returns a new [*Service] configured with the provided API
// key and options.
func New(apiKey string, opts ...Option) (*Service, error) {
	svc := Service{maxTokens: DefaultMaxTokens}
	for _, opt := range opts {
		opt(&svc)
	}
	if svc.client == nil {
		svc.client = anthropic.NewClient(apiKey)
	}

	if svc.model == "" {
		svc.log.Debug(fmt.Sprintf("[Anthropic] No model provided. Using default model %q", DefaultModel))
		svc.model = DefaultModel
	}
	svc.log.Debug(fmt.Sprintf("[Anthropic] Using model %q", svc.model))

	if svc.log == nil {
		svc.log = internal.NopLogger()
	}

	return &svc, nil
}

// GenerateDoc creates documentation text based on the provided context and
// returns it as a string.
func (svc *Service) GenerateDoc(ctx generate.Context) (string, error) {
	svc.log.Debug(fmt.Sprintf("[Anthropic] Generating docs for %s (%s)", ctx.Input().Identifier, ctx.Input().Language))

	req := svc.makeBaseRequest(ctx)

	resp, err := svc.client.CreateMessages(ctx, req)
	if err != nil {
		return "", fmt.Errorf("create messages: %w", err)
	}

	if len(resp.Content) == 0 {
		return "", fmt.Errorf("anthropic: no content returned")
	}

	content := resp.Content[0]
	if content.Text == nil {
		return "", fmt.Errorf("anthropic: no content returned")
	}

	return *content.Text, nil
}

func (svc *Service) makeBaseRequest(ctx generate.Context) anthropic.MessagesRequest {
	return anthropic.MessagesRequest{
		Model: svc.model,
		Messages: []anthropic.Message{
			anthropic.NewUserTextMessage(ctx.Prompt()),
		},
		MaxTokens: svc.maxTokens,
	}
}
