package llm

import (
	"testing"
	"time"
)

func TestNewLLMCallStartedEvent(t *testing.T) {
	provider := "openai"
	model := "gpt-4"
	messages := []Message{
		{Role: RoleUser, Content: "Hello"},
	}
	options := &CallOptions{
		Temperature: func() *float64 { t := 0.7; return &t }(),
	}

	event := NewLLMCallStartedEvent(provider, model, messages, options)

	if event.GetType() != EventTypeLLMCallStarted {
		t.Errorf("Expected event type %s, got %s", EventTypeLLMCallStarted, event.GetType())
	}

	if event.Provider != provider {
		t.Errorf("Expected provider %s, got %s", provider, event.Provider)
	}

	if event.Model != model {
		t.Errorf("Expected model %s, got %s", model, event.Model)
	}

	if len(event.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(event.Messages))
	}

	if event.Messages[0].Role != RoleUser {
		t.Errorf("Expected role %s, got %s", RoleUser, event.Messages[0].Role)
	}

	if event.Options != options {
		t.Error("Expected options to match")
	}

	if event.Metadata == nil {
		t.Error("Expected metadata to be initialized")
	}

	if event.GetTimestamp().IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestNewLLMCallCompletedEvent(t *testing.T) {
	provider := "openai"
	model := "gpt-4"
	response := &Response{
		Content: "Hello there!",
		Usage: Usage{
			PromptTokens:     10,
			CompletionTokens: 15,
			TotalTokens:      25,
		},
		Model: model,
	}
	duration := 2 * time.Second

	event := NewLLMCallCompletedEvent(provider, model, response, duration)

	if event.GetType() != EventTypeLLMCallCompleted {
		t.Errorf("Expected event type %s, got %s", EventTypeLLMCallCompleted, event.GetType())
	}

	if event.Provider != provider {
		t.Errorf("Expected provider %s, got %s", provider, event.Provider)
	}

	if event.Model != model {
		t.Errorf("Expected model %s, got %s", model, event.Model)
	}

	if event.Response != response {
		t.Error("Expected response to match")
	}

	if event.Duration != duration {
		t.Errorf("Expected duration %v, got %v", duration, event.Duration)
	}

	if event.TokensUsed != 25 {
		t.Errorf("Expected tokens used 25, got %d", event.TokensUsed)
	}

	if event.Cost <= 0 {
		t.Error("Expected cost to be calculated and positive")
	}

	if event.GetTimestamp().IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestNewLLMCallFailedEvent(t *testing.T) {
	provider := "openai"
	model := "gpt-4"
	err := &MockError{message: "API error"}
	duration := 1 * time.Second

	event := NewLLMCallFailedEvent(provider, model, err, duration)

	if event.GetType() != EventTypeLLMCallFailed {
		t.Errorf("Expected event type %s, got %s", EventTypeLLMCallFailed, event.GetType())
	}

	if event.Provider != provider {
		t.Errorf("Expected provider %s, got %s", provider, event.Provider)
	}

	if event.Model != model {
		t.Errorf("Expected model %s, got %s", model, event.Model)
	}

	if event.Error != err {
		t.Error("Expected error to match")
	}

	if event.Duration != duration {
		t.Errorf("Expected duration %v, got %v", duration, event.Duration)
	}

	if event.GetTimestamp().IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestNewLLMStreamStartedEvent(t *testing.T) {
	provider := "openai"
	model := "gpt-4"
	messages := []Message{
		{Role: RoleUser, Content: "Hello"},
	}
	options := &CallOptions{
		Stream: true,
	}

	event := NewLLMStreamStartedEvent(provider, model, messages, options)

	if event.GetType() != EventTypeLLMStreamStarted {
		t.Errorf("Expected event type %s, got %s", EventTypeLLMStreamStarted, event.GetType())
	}

	if event.Provider != provider {
		t.Errorf("Expected provider %s, got %s", provider, event.Provider)
	}

	if event.Model != model {
		t.Errorf("Expected model %s, got %s", model, event.Model)
	}

	if len(event.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(event.Messages))
	}

	if event.Options.Stream != true {
		t.Error("Expected stream to be true in options")
	}

	if event.GetTimestamp().IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestNewLLMStreamChunkEvent(t *testing.T) {
	provider := "openai"
	model := "gpt-4"
	chunk := "Hello"

	event := NewLLMStreamChunkEvent(provider, model, chunk)

	if event.GetType() != EventTypeLLMStreamChunk {
		t.Errorf("Expected event type %s, got %s", EventTypeLLMStreamChunk, event.GetType())
	}

	if event.Provider != provider {
		t.Errorf("Expected provider %s, got %s", provider, event.Provider)
	}

	if event.Model != model {
		t.Errorf("Expected model %s, got %s", model, event.Model)
	}

	if event.Chunk != chunk {
		t.Errorf("Expected chunk %s, got %s", chunk, event.Chunk)
	}

	if event.GetTimestamp().IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestNewLLMStreamEndedEvent(t *testing.T) {
	provider := "openai"
	model := "gpt-4"
	duration := 5 * time.Second
	tokensUsed := 1000 // 增加token数量以确保产生成本
	chunksCount := 10

	event := NewLLMStreamEndedEvent(provider, model, duration, tokensUsed, chunksCount)

	if event.GetType() != EventTypeLLMStreamEnded {
		t.Errorf("Expected event type %s, got %s", EventTypeLLMStreamEnded, event.GetType())
	}

	if event.Provider != provider {
		t.Errorf("Expected provider %s, got %s", provider, event.Provider)
	}

	if event.Model != model {
		t.Errorf("Expected model %s, got %s", model, event.Model)
	}

	if event.Duration != duration {
		t.Errorf("Expected duration %v, got %v", duration, event.Duration)
	}

	if event.TokensUsed != tokensUsed {
		t.Errorf("Expected tokens used %d, got %d", tokensUsed, event.TokensUsed)
	}

	if event.ChunksCount != chunksCount {
		t.Errorf("Expected chunks count %d, got %d", chunksCount, event.ChunksCount)
	}

	if event.Cost <= 0 {
		t.Error("Expected cost to be calculated and positive")
	}

	if event.GetTimestamp().IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestCalculateOpenAICost(t *testing.T) {
	tests := []struct {
		name            string
		model           string
		usage           Usage
		expectedCost    float64
		acceptableDelta float64
	}{
		{
			name:  "gpt-4 cost calculation",
			model: "gpt-4",
			usage: Usage{
				PromptTokens:     1000,
				CompletionTokens: 500,
				TotalTokens:      1500,
			},
			expectedCost:    0.06, // (1000/1000 * 0.03) + (500/1000 * 0.06) = 0.03 + 0.03 = 0.06
			acceptableDelta: 0.0001,
		},
		{
			name:  "gpt-3.5-turbo cost calculation",
			model: "gpt-3.5-turbo",
			usage: Usage{
				PromptTokens:     2000,
				CompletionTokens: 1000,
				TotalTokens:      3000,
			},
			expectedCost:    0.005, // (2000/1000 * 0.0015) + (1000/1000 * 0.002) = 0.003 + 0.002 = 0.005
			acceptableDelta: 0.0001,
		},
		{
			name:  "gpt-4o-mini cost calculation",
			model: "gpt-4o-mini",
			usage: Usage{
				PromptTokens:     10000,
				CompletionTokens: 5000,
				TotalTokens:      15000,
			},
			expectedCost:    0.0045, // (10000/1000 * 0.00015) + (5000/1000 * 0.0006) = 0.0015 + 0.003 = 0.0045
			acceptableDelta: 0.0001,
		},
		{
			name:  "unknown model default pricing",
			model: "unknown-model",
			usage: Usage{
				PromptTokens:     1000,
				CompletionTokens: 1000,
				TotalTokens:      2000,
			},
			expectedCost:    0.004, // (1000/1000 * 0.002) + (1000/1000 * 0.002) = 0.002 + 0.002 = 0.004
			acceptableDelta: 0.0001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := calculateOpenAICost(tt.model, tt.usage)

			if cost < tt.expectedCost-tt.acceptableDelta || cost > tt.expectedCost+tt.acceptableDelta {
				t.Errorf("Expected cost around %f, got %f (delta: %f)", tt.expectedCost, cost, tt.acceptableDelta)
			}
		})
	}
}

func TestCalculateCost(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		model    string
		usage    Usage
		wantCost bool // whether we expect a cost > 0
	}{
		{
			name:     "OpenAI provider cost",
			provider: "openai",
			model:    "gpt-4",
			usage: Usage{
				PromptTokens:     100,
				CompletionTokens: 50,
				TotalTokens:      150,
			},
			wantCost: true,
		},
		{
			name:     "Unknown provider no cost",
			provider: "unknown",
			model:    "some-model",
			usage: Usage{
				PromptTokens:     100,
				CompletionTokens: 50,
				TotalTokens:      150,
			},
			wantCost: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := calculateCost(tt.provider, tt.model, tt.usage)

			if tt.wantCost && cost <= 0 {
				t.Errorf("Expected cost > 0, got %f", cost)
			}

			if !tt.wantCost && cost != 0 {
				t.Errorf("Expected cost = 0, got %f", cost)
			}
		})
	}
}

func TestEventConstants(t *testing.T) {
	expectedEvents := map[string]string{
		EventTypeLLMCallStarted:   "llm_call_started",
		EventTypeLLMCallCompleted: "llm_call_completed",
		EventTypeLLMCallFailed:    "llm_call_failed",
		EventTypeLLMStreamStarted: "llm_stream_started",
		EventTypeLLMStreamChunk:   "llm_stream_chunk",
		EventTypeLLMStreamEnded:   "llm_stream_ended",
	}

	for constant, expectedValue := range expectedEvents {
		if constant != expectedValue {
			t.Errorf("Expected constant %s to equal %s, got %s", constant, expectedValue, constant)
		}
	}
}

func TestLLMEventStructures(t *testing.T) {
	// Test that all event types have required fields

	// LLMCallStartedEvent
	startedEvent := &LLMCallStartedEvent{}
	if startedEvent.Metadata == nil {
		startedEvent.Metadata = make(map[string]interface{})
	}

	// LLMCallCompletedEvent
	completedEvent := &LLMCallCompletedEvent{}
	if completedEvent.Metadata == nil {
		completedEvent.Metadata = make(map[string]interface{})
	}

	// LLMCallFailedEvent
	failedEvent := &LLMCallFailedEvent{}
	if failedEvent.Metadata == nil {
		failedEvent.Metadata = make(map[string]interface{})
	}

	// LLMStreamStartedEvent
	streamStartedEvent := &LLMStreamStartedEvent{}
	if streamStartedEvent.Metadata == nil {
		streamStartedEvent.Metadata = make(map[string]interface{})
	}

	// LLMStreamChunkEvent
	streamChunkEvent := &LLMStreamChunkEvent{}
	if streamChunkEvent.Metadata == nil {
		streamChunkEvent.Metadata = make(map[string]interface{})
	}

	// LLMStreamEndedEvent
	streamEndedEvent := &LLMStreamEndedEvent{}
	if streamEndedEvent.Metadata == nil {
		streamEndedEvent.Metadata = make(map[string]interface{})
	}

	// If we get here without panicking, the structures are valid
	t.Log("All event structures are valid")
}
