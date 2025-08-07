package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewOpenAILLM(t *testing.T) {
	llm := NewOpenAILLM("gpt-4")

	if llm.GetModel() != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got %s", llm.GetModel())
	}

	if llm.GetProvider() != "openai" {
		t.Errorf("Expected provider 'openai', got %s", llm.GetProvider())
	}

	if llm.GetBaseURL() != defaultOpenAIBaseURL {
		t.Errorf("Expected base URL %s, got %s", defaultOpenAIBaseURL, llm.GetBaseURL())
	}

	if !llm.SupportsFunctionCalling() {
		t.Error("Expected function calling support to be true")
	}

	// Check context window for known model
	expectedWindow := openAIContextWindows["gpt-4"]
	if llm.GetContextWindowSize() != expectedWindow {
		t.Errorf("Expected context window %d, got %d", expectedWindow, llm.GetContextWindowSize())
	}
}

func TestNewOpenAILLM_UnknownModel(t *testing.T) {
	llm := NewOpenAILLM("unknown-model")

	// Should default to 4096 for unknown models
	if llm.GetContextWindowSize() != 4096 {
		t.Errorf("Expected context window 4096 for unknown model, got %d", llm.GetContextWindowSize())
	}
}

func TestOpenAILLM_ConvertMessages(t *testing.T) {
	llm := NewOpenAILLM("gpt-4")

	messages := []Message{
		{Role: RoleSystem, Content: "You are a helpful assistant", Name: ""},
		{Role: RoleUser, Content: "Hello!", Name: "john"},
		{Role: RoleAssistant, Content: "Hi there!", Name: ""},
	}

	openAIMessages := llm.convertMessages(messages)

	if len(openAIMessages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(openAIMessages))
	}

	// Check first message
	if openAIMessages[0].Role != "system" {
		t.Errorf("Expected role 'system', got %s", openAIMessages[0].Role)
	}

	if openAIMessages[0].Content != "You are a helpful assistant" {
		t.Errorf("Expected content 'You are a helpful assistant', got %v", openAIMessages[0].Content)
	}

	// Check second message
	if openAIMessages[1].Role != "user" {
		t.Errorf("Expected role 'user', got %s", openAIMessages[1].Role)
	}

	if openAIMessages[1].Content != "Hello!" {
		t.Errorf("Expected content 'Hello!', got %v", openAIMessages[1].Content)
	}

	if openAIMessages[1].Name != "john" {
		t.Errorf("Expected name 'john', got %s", openAIMessages[1].Name)
	}
}

func TestOpenAILLM_BuildChatRequest(t *testing.T) {
	llm := NewOpenAILLM("gpt-4", WithAPIKey("test-key"))

	messages := []OpenAIMessage{
		{Role: "user", Content: "Hello"},
	}

	options := &CallOptions{
		Temperature:      func() *float64 { t := 0.7; return &t }(),
		MaxTokens:        func() *int { t := 1000; return &t }(),
		TopP:             func() *float64 { t := 0.9; return &t }(),
		FrequencyPenalty: func() *float64 { t := 0.5; return &t }(),
		PresencePenalty:  func() *float64 { t := 0.2; return &t }(),
		StopSequences:    []string{"STOP"},
		Stream:           true,
	}

	request := llm.buildChatRequest(messages, options)

	if request.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got %s", request.Model)
	}

	if len(request.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(request.Messages))
	}

	if *request.Temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %f", *request.Temperature)
	}

	if *request.MaxTokens != 1000 {
		t.Errorf("Expected max tokens 1000, got %d", *request.MaxTokens)
	}

	if *request.TopP != 0.9 {
		t.Errorf("Expected top_p 0.9, got %f", *request.TopP)
	}

	if *request.FrequencyPenalty != 0.5 {
		t.Errorf("Expected frequency penalty 0.5, got %f", *request.FrequencyPenalty)
	}

	if *request.PresencePenalty != 0.2 {
		t.Errorf("Expected presence penalty 0.2, got %f", *request.PresencePenalty)
	}

	if len(request.Stop) != 1 || request.Stop[0] != "STOP" {
		t.Errorf("Expected stop ['STOP'], got %v", request.Stop)
	}

	if !request.Stream {
		t.Error("Expected stream to be true")
	}
}

func TestOpenAILLM_BuildChatRequestWithTools(t *testing.T) {
	llm := NewOpenAILLM("gpt-4")

	messages := []OpenAIMessage{
		{Role: "user", Content: "What's the weather?"},
	}

	tools := []Tool{
		{
			Type: "function",
			Function: ToolSchema{
				Name:        "get_weather",
				Description: "Get current weather",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{
							"type":        "string",
							"description": "City name",
						},
					},
				},
			},
		},
	}

	options := &CallOptions{
		Tools:      tools,
		ToolChoice: "auto",
	}

	request := llm.buildChatRequest(messages, options)

	if len(request.Tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(request.Tools))
	}

	tool := request.Tools[0]
	if tool.Type != "function" {
		t.Errorf("Expected tool type 'function', got %s", tool.Type)
	}

	if tool.Function.Name != "get_weather" {
		t.Errorf("Expected tool name 'get_weather', got %s", tool.Function.Name)
	}

	if tool.Function.Description != "Get current weather" {
		t.Errorf("Expected tool description 'Get current weather', got %s", tool.Function.Description)
	}

	if request.ToolChoice != "auto" {
		t.Errorf("Expected tool choice 'auto', got %v", request.ToolChoice)
	}
}

func TestOpenAILLM_ConvertResponse(t *testing.T) {
	llm := NewOpenAILLM("gpt-4")

	openAIResponse := &OpenAIChatResponse{
		ID:      "chatcmpl-123",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "gpt-4",
		Usage: OpenAIUsage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
		Choices: []OpenAIChoice{
			{
				Index: 0,
				Message: OpenAIMessage{
					Role:    "assistant",
					Content: "Hello, how can I help you?",
				},
				FinishReason: "stop",
			},
		},
	}

	response := llm.convertResponse(openAIResponse)

	if response.Content != "Hello, how can I help you?" {
		t.Errorf("Expected content 'Hello, how can I help you?', got %s", response.Content)
	}

	if response.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got %s", response.Model)
	}

	if response.FinishReason != "stop" {
		t.Errorf("Expected finish reason 'stop', got %s", response.FinishReason)
	}

	if response.Usage.PromptTokens != 10 {
		t.Errorf("Expected prompt tokens 10, got %d", response.Usage.PromptTokens)
	}

	if response.Usage.CompletionTokens != 20 {
		t.Errorf("Expected completion tokens 20, got %d", response.Usage.CompletionTokens)
	}

	if response.Usage.TotalTokens != 30 {
		t.Errorf("Expected total tokens 30, got %d", response.Usage.TotalTokens)
	}

	if response.Metadata["id"] != "chatcmpl-123" {
		t.Errorf("Expected metadata id 'chatcmpl-123', got %v", response.Metadata["id"])
	}
}

func TestOpenAILLM_ConvertResponseWithToolCalls(t *testing.T) {
	llm := NewOpenAILLM("gpt-4")

	openAIResponse := &OpenAIChatResponse{
		ID:      "chatcmpl-123",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "gpt-4",
		Usage: OpenAIUsage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
		Choices: []OpenAIChoice{
			{
				Index: 0,
				Message: OpenAIMessage{
					Role:    "assistant",
					Content: "",
					ToolCalls: []OpenAIToolCall{
						{
							ID:   "call_123",
							Type: "function",
							Function: OpenAIToolCallFunc{
								Name:      "get_weather",
								Arguments: `{"location": "New York"}`,
							},
						},
					},
				},
				FinishReason: "tool_calls",
			},
		},
	}

	response := llm.convertResponse(openAIResponse)

	if response.FinishReason != "tool_calls" {
		t.Errorf("Expected finish reason 'tool_calls', got %s", response.FinishReason)
	}

	if len(response.ToolCalls) != 1 {
		t.Errorf("Expected 1 tool call, got %d", len(response.ToolCalls))
	}

	toolCall := response.ToolCalls[0]
	if toolCall.ID != "call_123" {
		t.Errorf("Expected tool call ID 'call_123', got %s", toolCall.ID)
	}

	if toolCall.Type != "function" {
		t.Errorf("Expected tool call type 'function', got %s", toolCall.Type)
	}

	if toolCall.Function.Name != "get_weather" {
		t.Errorf("Expected function name 'get_weather', got %s", toolCall.Function.Name)
	}

	if toolCall.Function.Arguments != `{"location": "New York"}` {
		t.Errorf("Expected arguments '{\"location\": \"New York\"}', got %s", toolCall.Function.Arguments)
	}
}

func TestOpenAILLM_Call_InvalidInput(t *testing.T) {
	llm := NewOpenAILLM("gpt-4", WithAPIKey("test-key"))
	ctx := context.Background()

	// Test empty messages
	_, err := llm.Call(ctx, []Message{}, nil)
	if err == nil {
		t.Error("Expected error for empty messages")
	}

	// Test invalid options
	invalidOptions := &CallOptions{
		Temperature: func() *float64 { t := -1.0; return &t }(),
	}

	messages := []Message{
		{Role: RoleUser, Content: "Hello"},
	}

	_, err = llm.Call(ctx, messages, invalidOptions)
	if err == nil {
		t.Error("Expected error for invalid options")
	}
}

// Mock HTTP server for testing API calls
func createMockOpenAIServer(t *testing.T, responseBody string, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.URL.Path != openAIChatEndpoint {
			t.Errorf("Expected path %s, got %s", openAIChatEndpoint, r.URL.Path)
		}

		// Verify headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			t.Errorf("Expected Authorization header with Bearer token, got %s", r.Header.Get("Authorization"))
		}

		// Read and verify request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
		}

		var request OpenAIChatRequest
		if err := json.Unmarshal(body, &request); err != nil {
			t.Errorf("Failed to unmarshal request: %v", err)
		}

		// Set status and write response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		w.Write([]byte(responseBody))
	}))
}

func TestOpenAILLM_Call_Success(t *testing.T) {
	successResponse := `{
		"id": "chatcmpl-123",
		"object": "chat.completion",
		"created": 1234567890,
		"model": "gpt-4",
		"choices": [
			{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": "Hello! How can I help you today?"
				},
				"finish_reason": "stop"
			}
		],
		"usage": {
			"prompt_tokens": 10,
			"completion_tokens": 20,
			"total_tokens": 30
		}
	}`

	server := createMockOpenAIServer(t, successResponse, 200)
	defer server.Close()

	llm := NewOpenAILLM("gpt-4",
		WithAPIKey("test-key"),
		WithBaseURL(server.URL),
	)

	ctx := context.Background()
	messages := []Message{
		{Role: RoleUser, Content: "Hello"},
	}

	response, err := llm.Call(ctx, messages, nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Content != "Hello! How can I help you today?" {
		t.Errorf("Expected content 'Hello! How can I help you today?', got %s", response.Content)
	}

	if response.Usage.TotalTokens != 30 {
		t.Errorf("Expected total tokens 30, got %d", response.Usage.TotalTokens)
	}
}

func TestOpenAILLM_Call_APIError(t *testing.T) {
	errorResponse := `{
		"error": {
			"message": "Invalid API key",
			"type": "invalid_request_error",
			"code": "invalid_api_key"
		}
	}`

	server := createMockOpenAIServer(t, errorResponse, 401)
	defer server.Close()

	llm := NewOpenAILLM("gpt-4",
		WithAPIKey("invalid-key"),
		WithBaseURL(server.URL),
	)

	ctx := context.Background()
	messages := []Message{
		{Role: RoleUser, Content: "Hello"},
	}

	_, err := llm.Call(ctx, messages, nil)
	if err == nil {
		t.Error("Expected error for invalid API key")
	}

	if !strings.Contains(err.Error(), "Invalid API key") {
		t.Errorf("Expected error message to contain 'Invalid API key', got %s", err.Error())
	}
}

func TestOpenAILLM_CallStream_InvalidInput(t *testing.T) {
	llm := NewOpenAILLM("gpt-4", WithAPIKey("test-key"))
	ctx := context.Background()

	// Test empty messages
	_, err := llm.CallStream(ctx, []Message{}, nil)
	if err == nil {
		t.Error("Expected error for empty messages")
	}

	// Test invalid options
	invalidOptions := &CallOptions{
		Temperature: func() *float64 { t := 3.0; return &t }(),
	}

	messages := []Message{
		{Role: RoleUser, Content: "Hello"},
	}

	_, err = llm.CallStream(ctx, messages, invalidOptions)
	if err == nil {
		t.Error("Expected error for invalid options")
	}
}

func createMockStreamingServer(t *testing.T, chunks []string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Error("Streaming not supported")
			return
		}

		for _, chunk := range chunks {
			w.Write([]byte("data: " + chunk + "\n\n"))
			flusher.Flush()
		}

		w.Write([]byte("data: [DONE]\n\n"))
		flusher.Flush()
	}))
}

func TestOpenAILLM_CallStream_Success(t *testing.T) {
	chunks := []string{
		`{"id":"chatcmpl-123","object":"chat.completion.chunk","created":1234567890,"model":"gpt-4","choices":[{"index":0,"delta":{"role":"assistant","content":"Hello"},"finish_reason":null}]}`,
		`{"id":"chatcmpl-123","object":"chat.completion.chunk","created":1234567890,"model":"gpt-4","choices":[{"index":0,"delta":{"content":" there"},"finish_reason":null}]}`,
		`{"id":"chatcmpl-123","object":"chat.completion.chunk","created":1234567890,"model":"gpt-4","choices":[{"index":0,"delta":{"content":"!"},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":20,"total_tokens":30}}`,
	}

	server := createMockStreamingServer(t, chunks)
	defer server.Close()

	llm := NewOpenAILLM("gpt-4",
		WithAPIKey("test-key"),
		WithBaseURL(server.URL),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	messages := []Message{
		{Role: RoleUser, Content: "Hello"},
	}

	respChan, err := llm.CallStream(ctx, messages, nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	var responses []StreamResponse
	for response := range respChan {
		if response.Error != nil {
			t.Errorf("Unexpected error in stream: %v", response.Error)
			continue
		}
		responses = append(responses, response)
	}

	if len(responses) == 0 {
		t.Error("Expected at least one response")
	}

	// Check if we received the expected content
	var fullContent string
	var finalUsage *Usage
	var finalFinishReason string

	for _, response := range responses {
		fullContent += response.Delta
		if response.Usage != nil {
			finalUsage = response.Usage
		}
		if response.FinishReason != "" {
			finalFinishReason = response.FinishReason
		}
	}

	if fullContent != "Hello there!" {
		t.Errorf("Expected full content 'Hello there!', got %s", fullContent)
	}

	if finalFinishReason != "stop" {
		t.Errorf("Expected finish reason 'stop', got %s", finalFinishReason)
	}

	if finalUsage == nil {
		t.Error("Expected usage information in final response")
	} else if finalUsage.TotalTokens != 30 {
		t.Errorf("Expected total tokens 30, got %d", finalUsage.TotalTokens)
	}
}
