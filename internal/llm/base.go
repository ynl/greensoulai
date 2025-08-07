package llm

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// BaseLLM provides a base implementation for LLM instances
type BaseLLM struct {
	provider         string
	model            string
	apiKey           string
	baseURL          string
	timeout          time.Duration
	maxRetries       int
	client           *http.Client
	logger           logger.Logger
	eventBus         events.EventBus
	contextWindow    int
	supportsFuncCall bool
	customHeaders    map[string]string
}

// BaseLLMOption is a functional option for configuring BaseLLM
type BaseLLMOption func(*BaseLLM)

// NewBaseLLM creates a new BaseLLM instance
func NewBaseLLM(provider, model string, options ...BaseLLMOption) *BaseLLM {
	b := &BaseLLM{
		provider:         provider,
		model:            model,
		timeout:          30 * time.Second,
		maxRetries:       3,
		contextWindow:    4096,
		supportsFuncCall: false,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		logger:        logger.NewConsoleLogger(),
		customHeaders: make(map[string]string),
	}

	// Apply options
	for _, option := range options {
		option(b)
	}

	return b
}

// WithAPIKey sets the API key for the LLM
func WithAPIKey(apiKey string) BaseLLMOption {
	return func(b *BaseLLM) {
		b.apiKey = apiKey
	}
}

// WithBaseURL sets the base URL for the LLM API
func WithBaseURL(baseURL string) BaseLLMOption {
	return func(b *BaseLLM) {
		b.baseURL = baseURL
	}
}

// WithTimeout sets the request timeout
func WithTimeout(timeout time.Duration) BaseLLMOption {
	return func(b *BaseLLM) {
		b.timeout = timeout
		if b.client != nil {
			b.client.Timeout = timeout
		}
	}
}

// WithMaxRetries sets the maximum number of retries
func WithMaxRetries(maxRetries int) BaseLLMOption {
	return func(b *BaseLLM) {
		b.maxRetries = maxRetries
	}
}

// WithHTTPClient sets the HTTP client
func WithHTTPClient(client *http.Client) BaseLLMOption {
	return func(b *BaseLLM) {
		b.client = client
	}
}

// WithLogger sets the logger
func WithLogger(logger logger.Logger) BaseLLMOption {
	return func(b *BaseLLM) {
		b.logger = logger
	}
}

// WithContextWindow sets the context window size
func WithContextWindow(size int) BaseLLMOption {
	return func(b *BaseLLM) {
		b.contextWindow = size
	}
}

// WithFunctionCalling enables/disables function calling support
func WithFunctionCalling(supports bool) BaseLLMOption {
	return func(b *BaseLLM) {
		b.supportsFuncCall = supports
	}
}

// WithCustomHeaders sets custom HTTP headers
func WithCustomHeaders(headers map[string]string) BaseLLMOption {
	return func(b *BaseLLM) {
		if b.customHeaders == nil {
			b.customHeaders = make(map[string]string)
		}
		for k, v := range headers {
			b.customHeaders[k] = v
		}
	}
}

// WithCustomHeader sets a single custom HTTP header
func WithCustomHeader(key, value string) BaseLLMOption {
	return func(b *BaseLLM) {
		if b.customHeaders == nil {
			b.customHeaders = make(map[string]string)
		}
		b.customHeaders[key] = value
	}
}

// GetModel returns the model identifier
func (b *BaseLLM) GetModel() string {
	return b.model
}

// SupportsFunctionCalling returns true if the model supports function calling
func (b *BaseLLM) SupportsFunctionCalling() bool {
	return b.supportsFuncCall
}

// GetContextWindowSize returns the maximum context window size
func (b *BaseLLM) GetContextWindowSize() int {
	return b.contextWindow
}

// SetEventBus sets the event bus for emitting events
func (b *BaseLLM) SetEventBus(eventBus events.EventBus) {
	b.eventBus = eventBus
}

// Close cleans up resources
func (b *BaseLLM) Close() error {
	// Clean up any resources here
	return nil
}

// GetProvider returns the provider name
func (b *BaseLLM) GetProvider() string {
	return b.provider
}

// GetAPIKey returns the API key (for internal use)
func (b *BaseLLM) GetAPIKey() string {
	return b.apiKey
}

// GetBaseURL returns the base URL
func (b *BaseLLM) GetBaseURL() string {
	return b.baseURL
}

// GetTimeout returns the request timeout
func (b *BaseLLM) GetTimeout() time.Duration {
	return b.timeout
}

// GetMaxRetries returns the maximum number of retries
func (b *BaseLLM) GetMaxRetries() int {
	return b.maxRetries
}

// GetHTTPClient returns the HTTP client
func (b *BaseLLM) GetHTTPClient() *http.Client {
	return b.client
}

// GetLogger returns the logger
func (b *BaseLLM) GetLogger() logger.Logger {
	return b.logger
}

// GetEventBus returns the event bus
func (b *BaseLLM) GetEventBus() events.EventBus {
	return b.eventBus
}

// GetCustomHeaders returns the custom headers
func (b *BaseLLM) GetCustomHeaders() map[string]string {
	return b.customHeaders
}

// EmitEvent emits an event if event bus is available
func (b *BaseLLM) EmitEvent(ctx context.Context, event events.Event) {
	if b.eventBus != nil {
		err := b.eventBus.Emit(ctx, nil, event)
		if err != nil && b.logger != nil {
			b.logger.Error("Failed to emit event",
				logger.Field{Key: "event_type", Value: event.GetType()},
				logger.Field{Key: "error", Value: err},
			)
		}
	}
}

// LogInfo logs an info message
func (b *BaseLLM) LogInfo(message string, fields ...logger.Field) {
	if b.logger != nil {
		b.logger.Info(message, fields...)
	}
}

// LogError logs an error message
func (b *BaseLLM) LogError(message string, fields ...logger.Field) {
	if b.logger != nil {
		b.logger.Error(message, fields...)
	}
}

// LogDebug logs a debug message
func (b *BaseLLM) LogDebug(message string, fields ...logger.Field) {
	if b.logger != nil {
		b.logger.Debug(message, fields...)
	}
}

// ValidateMessages validates the input messages
func (b *BaseLLM) ValidateMessages(messages []Message) error {
	if len(messages) == 0 {
		return fmt.Errorf("messages cannot be empty")
	}

	for i, msg := range messages {
		if msg.Role == "" {
			return fmt.Errorf("message %d: role cannot be empty", i)
		}

		if msg.Content == nil {
			return fmt.Errorf("message %d: content cannot be nil", i)
		}

		// Validate role
		switch msg.Role {
		case RoleSystem, RoleUser, RoleAssistant, RoleTool:
			// Valid roles
		default:
			return fmt.Errorf("message %d: invalid role '%s'", i, msg.Role)
		}
	}

	return nil
}

// ValidateCallOptions validates call options
func (b *BaseLLM) ValidateCallOptions(options *CallOptions) error {
	if options == nil {
		return nil
	}

	if options.Temperature != nil {
		temp := *options.Temperature
		if temp < 0 || temp > 2 {
			return fmt.Errorf("temperature must be between 0 and 2, got %f", temp)
		}
	}

	if options.TopP != nil {
		topP := *options.TopP
		if topP < 0 || topP > 1 {
			return fmt.Errorf("top_p must be between 0 and 1, got %f", topP)
		}
	}

	if options.MaxTokens != nil {
		maxTokens := *options.MaxTokens
		if maxTokens < 1 {
			return fmt.Errorf("max_tokens must be positive, got %d", maxTokens)
		}
	}

	return nil
}
