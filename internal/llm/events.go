package llm

import (
	"time"

	"github.com/ynl/greensoulai/pkg/events"
)

// LLM event types
const (
	EventTypeLLMCallStarted   = "llm_call_started"
	EventTypeLLMCallCompleted = "llm_call_completed"
	EventTypeLLMCallFailed    = "llm_call_failed"
	EventTypeLLMStreamStarted = "llm_stream_started"
	EventTypeLLMStreamChunk   = "llm_stream_chunk"
	EventTypeLLMStreamEnded   = "llm_stream_ended"
)

// LLMCallStartedEvent represents the start of an LLM call
type LLMCallStartedEvent struct {
	events.BaseEvent
	Provider string                 `json:"provider"`
	Model    string                 `json:"model"`
	Messages []Message              `json:"messages"`
	Options  *CallOptions           `json:"options"`
	Metadata map[string]interface{} `json:"metadata"`
}

// NewLLMCallStartedEvent creates a new LLM call started event
func NewLLMCallStartedEvent(provider, model string, messages []Message, options *CallOptions) *LLMCallStartedEvent {
	return &LLMCallStartedEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypeLLMCallStarted,
			Timestamp: time.Now(),
		},
		Provider: provider,
		Model:    model,
		Messages: messages,
		Options:  options,
		Metadata: make(map[string]interface{}),
	}
}

// LLMCallCompletedEvent represents the completion of an LLM call
type LLMCallCompletedEvent struct {
	events.BaseEvent
	Provider   string                 `json:"provider"`
	Model      string                 `json:"model"`
	Response   *Response              `json:"response"`
	Duration   time.Duration          `json:"duration"`
	TokensUsed int                    `json:"tokens_used"`
	Cost       float64                `json:"cost"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// NewLLMCallCompletedEvent creates a new LLM call completed event
func NewLLMCallCompletedEvent(provider, model string, response *Response, duration time.Duration) *LLMCallCompletedEvent {
	cost := calculateCost(provider, model, response.Usage)

	return &LLMCallCompletedEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypeLLMCallCompleted,
			Timestamp: time.Now(),
		},
		Provider:   provider,
		Model:      model,
		Response:   response,
		Duration:   duration,
		TokensUsed: response.Usage.TotalTokens,
		Cost:       cost,
		Metadata:   make(map[string]interface{}),
	}
}

// LLMCallFailedEvent represents a failed LLM call
type LLMCallFailedEvent struct {
	events.BaseEvent
	Provider string                 `json:"provider"`
	Model    string                 `json:"model"`
	Error    error                  `json:"error"`
	Duration time.Duration          `json:"duration"`
	Metadata map[string]interface{} `json:"metadata"`
}

// NewLLMCallFailedEvent creates a new LLM call failed event
func NewLLMCallFailedEvent(provider, model string, err error, duration time.Duration) *LLMCallFailedEvent {
	return &LLMCallFailedEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypeLLMCallFailed,
			Timestamp: time.Now(),
		},
		Provider: provider,
		Model:    model,
		Error:    err,
		Duration: duration,
		Metadata: make(map[string]interface{}),
	}
}

// LLMStreamStartedEvent represents the start of streaming
type LLMStreamStartedEvent struct {
	events.BaseEvent
	Provider string                 `json:"provider"`
	Model    string                 `json:"model"`
	Messages []Message              `json:"messages"`
	Options  *CallOptions           `json:"options"`
	Metadata map[string]interface{} `json:"metadata"`
}

// NewLLMStreamStartedEvent creates a new LLM stream started event
func NewLLMStreamStartedEvent(provider, model string, messages []Message, options *CallOptions) *LLMStreamStartedEvent {
	return &LLMStreamStartedEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypeLLMStreamStarted,
			Timestamp: time.Now(),
		},
		Provider: provider,
		Model:    model,
		Messages: messages,
		Options:  options,
		Metadata: make(map[string]interface{}),
	}
}

// LLMStreamChunkEvent represents a streaming chunk
type LLMStreamChunkEvent struct {
	events.BaseEvent
	Provider string                 `json:"provider"`
	Model    string                 `json:"model"`
	Chunk    string                 `json:"chunk"`
	Metadata map[string]interface{} `json:"metadata"`
}

// NewLLMStreamChunkEvent creates a new LLM stream chunk event
func NewLLMStreamChunkEvent(provider, model, chunk string) *LLMStreamChunkEvent {
	return &LLMStreamChunkEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypeLLMStreamChunk,
			Timestamp: time.Now(),
		},
		Provider: provider,
		Model:    model,
		Chunk:    chunk,
		Metadata: make(map[string]interface{}),
	}
}

// LLMStreamEndedEvent represents the end of streaming
type LLMStreamEndedEvent struct {
	events.BaseEvent
	Provider    string                 `json:"provider"`
	Model       string                 `json:"model"`
	Duration    time.Duration          `json:"duration"`
	TokensUsed  int                    `json:"tokens_used"`
	Cost        float64                `json:"cost"`
	ChunksCount int                    `json:"chunks_count"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// NewLLMStreamEndedEvent creates a new LLM stream ended event
func NewLLMStreamEndedEvent(provider, model string, duration time.Duration, tokensUsed int, chunksCount int) *LLMStreamEndedEvent {
	// For streaming, assume roughly half are prompt tokens and half are completion tokens
	promptTokens := tokensUsed / 2
	completionTokens := tokensUsed - promptTokens
	usage := Usage{
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      tokensUsed,
	}
	cost := calculateCost(provider, model, usage)

	return &LLMStreamEndedEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypeLLMStreamEnded,
			Timestamp: time.Now(),
		},
		Provider:    provider,
		Model:       model,
		Duration:    duration,
		TokensUsed:  tokensUsed,
		Cost:        cost,
		ChunksCount: chunksCount,
		Metadata:    make(map[string]interface{}),
	}
}

// calculateCost estimates the cost based on provider, model, and usage
func calculateCost(provider, model string, usage Usage) float64 {
	// This is a simplified cost calculation
	// In a real implementation, you would have a more comprehensive pricing table

	switch provider {
	case "openai":
		return calculateOpenAICost(model, usage)
	default:
		return 0.0
	}
}

// calculateOpenAICost calculates cost for OpenAI models
func calculateOpenAICost(model string, usage Usage) float64 {
	// OpenAI pricing (as of 2024, in USD per 1K tokens)
	// These are example prices and should be updated with current pricing

	var inputCostPer1K, outputCostPer1K float64

	switch model {
	case "gpt-4":
		inputCostPer1K = 0.03
		outputCostPer1K = 0.06
	case "gpt-4-turbo":
		inputCostPer1K = 0.01
		outputCostPer1K = 0.03
	case "gpt-4o":
		inputCostPer1K = 0.005
		outputCostPer1K = 0.015
	case "gpt-4o-mini":
		inputCostPer1K = 0.00015
		outputCostPer1K = 0.0006
	case "gpt-3.5-turbo":
		inputCostPer1K = 0.0015
		outputCostPer1K = 0.002
	default:
		// Default pricing
		inputCostPer1K = 0.002
		outputCostPer1K = 0.002
	}

	inputCost := float64(usage.PromptTokens) / 1000.0 * inputCostPer1K
	outputCost := float64(usage.CompletionTokens) / 1000.0 * outputCostPer1K

	return inputCost + outputCost
}
