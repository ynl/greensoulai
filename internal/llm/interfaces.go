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
	Temperature         *float64    `json:"temperature,omitempty"`
	MaxTokens           *int        `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int        `json:"max_completion_tokens,omitempty"` // 对标Python版本
	TopP                *float64    `json:"top_p,omitempty"`
	FrequencyPenalty    *float64    `json:"frequency_penalty,omitempty"`
	PresencePenalty     *float64    `json:"presence_penalty,omitempty"`
	StopSequences       []string    `json:"stop,omitempty"`
	Tools               []Tool      `json:"tools,omitempty"`
	ToolChoice          interface{} `json:"tool_choice,omitempty"`
	Stream              bool        `json:"stream,omitempty"`

	// 对标Python版本的新增参数
	N              *int            `json:"n,omitempty"`               // 生成响应数量
	Seed           *int            `json:"seed,omitempty"`            // 随机种子
	Logprobs       *int            `json:"logprobs,omitempty"`        // 返回logprobs数量
	TopLogprobs    *int            `json:"top_logprobs,omitempty"`    // 返回top logprobs
	ResponseFormat interface{}     `json:"response_format,omitempty"` // 响应格式
	LogitBias      map[int]float64 `json:"logit_bias,omitempty"`      // logit偏置
	User           string          `json:"user,omitempty"`            // 用户ID
	Timeout        *time.Duration  `json:"timeout,omitempty"`         // 请求超时

	// 回调和事件相关，对标Python版本
	Callbacks          []interface{}          `json:"callbacks,omitempty"`           // 回调函数
	AvailableFunctions map[string]interface{} `json:"available_functions,omitempty"` // 可用函数
	FromTask           interface{}            `json:"from_task,omitempty"`           // 任务来源
	FromAgent          interface{}            `json:"from_agent,omitempty"`          // Agent来源

	// 流式响应选项
	StreamOptions map[string]interface{} `json:"stream_options,omitempty"` // 流式选项

	Metadata map[string]interface{} `json:"metadata,omitempty"`
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

// WithMaxCompletionTokens sets the maximum completion tokens, 对标Python版本
func WithMaxCompletionTokens(maxCompletionTokens int) CallOption {
	return func(opts *CallOptions) {
		opts.MaxCompletionTokens = &maxCompletionTokens
	}
}

// WithFrequencyPenalty sets the frequency penalty
func WithFrequencyPenalty(penalty float64) CallOption {
	return func(opts *CallOptions) {
		opts.FrequencyPenalty = &penalty
	}
}

// WithPresencePenalty sets the presence penalty
func WithPresencePenalty(penalty float64) CallOption {
	return func(opts *CallOptions) {
		opts.PresencePenalty = &penalty
	}
}

// WithN sets the number of responses to generate, 对标Python版本
func WithN(n int) CallOption {
	return func(opts *CallOptions) {
		opts.N = &n
	}
}

// WithSeed sets the random seed for reproducible outputs, 对标Python版本
func WithSeed(seed int) CallOption {
	return func(opts *CallOptions) {
		opts.Seed = &seed
	}
}

// WithResponseFormat sets the response format, 对标Python版本
func WithResponseFormat(format interface{}) CallOption {
	return func(opts *CallOptions) {
		opts.ResponseFormat = format
	}
}

// WithLogitBias sets logit bias for specific tokens, 对标Python版本
func WithLogitBias(bias map[int]float64) CallOption {
	return func(opts *CallOptions) {
		opts.LogitBias = bias
	}
}

// WithUser sets the user ID for tracking purposes, 对标Python版本
func WithUser(user string) CallOption {
	return func(opts *CallOptions) {
		opts.User = user
	}
}

// Note: WithTimeout is defined in base.go as a BaseLLMOption

// WithCallbacks sets callback functions, 对标Python版本
func WithCallbacks(callbacks []interface{}) CallOption {
	return func(opts *CallOptions) {
		opts.Callbacks = callbacks
	}
}

// WithAvailableFunctions sets available functions for tool calls, 对标Python版本
func WithAvailableFunctions(functions map[string]interface{}) CallOption {
	return func(opts *CallOptions) {
		opts.AvailableFunctions = functions
	}
}

// WithFromTask sets the originating task, 对标Python版本
func WithFromTask(task interface{}) CallOption {
	return func(opts *CallOptions) {
		opts.FromTask = task
	}
}

// WithFromAgent sets the originating agent, 对标Python版本
func WithFromAgent(agent interface{}) CallOption {
	return func(opts *CallOptions) {
		opts.FromAgent = agent
	}
}

// WithStreamOptions sets streaming options, 对标Python版本
func WithStreamOptions(options map[string]interface{}) CallOption {
	return func(opts *CallOptions) {
		opts.StreamOptions = options
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
