package llm

import (
	"context"
	"time"

	"github.com/ynl/greensoulai/pkg/events"
)

// Message represents a conversation message
type Message struct {
	Role    Role        `json:"role"`
	Content interface{} `json:"content"`
	Name    string      `json:"name,omitempty"`
}

// Role represents the role of a message sender
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Usage represents token usage information
type Usage struct {
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	Cost             float64 `json:"cost,omitempty"`
}

// Response represents an LLM response
type Response struct {
	Content      string                 `json:"content"`
	Usage        Usage                  `json:"usage"`
	Model        string                 `json:"model"`
	FinishReason string                 `json:"finish_reason,omitempty"`
	ToolCalls    []ToolCall             `json:"tool_calls,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// StreamResponse represents a streaming LLM response
type StreamResponse struct {
	Delta        string     `json:"delta"`
	Usage        *Usage     `json:"usage,omitempty"`
	FinishReason string     `json:"finish_reason,omitempty"`
	Error        error      `json:"error,omitempty"`
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall represents a function/tool call
type ToolCall struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Function ToolCallFunction       `json:"function"`
	Args     map[string]interface{} `json:"args,omitempty"`
}

// ToolCallFunction represents the function part of a tool call
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Tool represents a tool schema for function calling
type Tool struct {
	Type     string     `json:"type"`
	Function ToolSchema `json:"function"`
}

// ToolSchema defines the schema for a tool
type ToolSchema struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// CallOptions contains options for LLM calls
type CallOptions struct {
	Temperature      *float64               `json:"temperature,omitempty"`
	MaxTokens        *int                   `json:"max_tokens,omitempty"`
	TopP             *float64               `json:"top_p,omitempty"`
	FrequencyPenalty *float64               `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64               `json:"presence_penalty,omitempty"`
	StopSequences    []string               `json:"stop,omitempty"`
	Tools            []Tool                 `json:"tools,omitempty"`
	ToolChoice       interface{}            `json:"tool_choice,omitempty"`
	Stream           bool                   `json:"stream,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// LLM defines the interface for language model implementations
type LLM interface {
	// Call sends a synchronous request to the LLM
	Call(ctx context.Context, messages []Message, options *CallOptions) (*Response, error)

	// CallStream sends a streaming request to the LLM
	CallStream(ctx context.Context, messages []Message, options *CallOptions) (<-chan StreamResponse, error)

	// GetModel returns the model identifier
	GetModel() string

	// SupportsFunctionCalling returns true if the model supports function calling
	SupportsFunctionCalling() bool

	// GetContextWindowSize returns the maximum context window size
	GetContextWindowSize() int

	// SetEventBus sets the event bus for emitting events
	SetEventBus(eventBus events.EventBus)

	// Close cleans up resources
	Close() error
}

// CallOption is a functional option for configuring LLM calls
type CallOption func(*CallOptions)

// WithTemperature sets the temperature for the LLM call
func WithTemperature(temperature float64) CallOption {
	return func(opts *CallOptions) {
		opts.Temperature = &temperature
	}
}

// WithMaxTokens sets the maximum number of tokens for the LLM call
func WithMaxTokens(maxTokens int) CallOption {
	return func(opts *CallOptions) {
		opts.MaxTokens = &maxTokens
	}
}

// WithTopP sets the top-p value for the LLM call
func WithTopP(topP float64) CallOption {
	return func(opts *CallOptions) {
		opts.TopP = &topP
	}
}

// WithStopSequences sets the stop sequences for the LLM call
func WithStopSequences(stops []string) CallOption {
	return func(opts *CallOptions) {
		opts.StopSequences = stops
	}
}

// WithTools sets the tools available for the LLM call
func WithTools(tools []Tool) CallOption {
	return func(opts *CallOptions) {
		opts.Tools = tools
	}
}

// WithStream enables streaming for the LLM call
func WithStream(stream bool) CallOption {
	return func(opts *CallOptions) {
		opts.Stream = stream
	}
}

// WithMetadata sets metadata for the LLM call
func WithMetadata(metadata map[string]interface{}) CallOption {
	return func(opts *CallOptions) {
		opts.Metadata = metadata
	}
}

// Provider represents an LLM provider
type Provider interface {
	// Name returns the provider name
	Name() string

	// CreateLLM creates a new LLM instance
	CreateLLM(config map[string]interface{}) (LLM, error)

	// SupportedModels returns the list of supported models
	SupportedModels() []string
}

// Config represents LLM configuration
type Config struct {
	Provider    string                 `json:"provider"`
	Model       string                 `json:"model"`
	APIKey      string                 `json:"api_key,omitempty"`
	BaseURL     string                 `json:"base_url,omitempty"`
	Timeout     time.Duration          `json:"timeout,omitempty"`
	MaxRetries  int                    `json:"max_retries,omitempty"`
	Temperature *float64               `json:"temperature,omitempty"`
	MaxTokens   *int                   `json:"max_tokens,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// DefaultCallOptions returns default call options
func DefaultCallOptions() *CallOptions {
	return &CallOptions{
		Temperature: nil,
		MaxTokens:   nil,
		TopP:        nil,
		Stream:      false,
		Metadata:    make(map[string]interface{}),
	}
}

// ApplyOptions applies functional options to CallOptions
func (opts *CallOptions) ApplyOptions(options ...CallOption) {
	for _, option := range options {
		option(opts)
	}
}
