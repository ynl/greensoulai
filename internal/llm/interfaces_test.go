package llm

import (
	"testing"
)

func TestMessage_Creation(t *testing.T) {
	tests := []struct {
		name     string
		role     Role
		content  interface{}
		name_    string
		expected Message
	}{
		{
			name:    "system message",
			role:    RoleSystem,
			content: "You are a helpful assistant",
			name_:   "",
			expected: Message{
				Role:    RoleSystem,
				Content: "You are a helpful assistant",
				Name:    "",
			},
		},
		{
			name:    "user message with name",
			role:    RoleUser,
			content: "Hello, how are you?",
			name_:   "john",
			expected: Message{
				Role:    RoleUser,
				Content: "Hello, how are you?",
				Name:    "john",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := Message{
				Role:    tt.role,
				Content: tt.content,
				Name:    tt.name_,
			}

			if msg.Role != tt.expected.Role {
				t.Errorf("Expected role %v, got %v", tt.expected.Role, msg.Role)
			}

			if msg.Content != tt.expected.Content {
				t.Errorf("Expected content %v, got %v", tt.expected.Content, msg.Content)
			}

			if msg.Name != tt.expected.Name {
				t.Errorf("Expected name %v, got %v", tt.expected.Name, msg.Name)
			}
		})
	}
}

func TestRole_Constants(t *testing.T) {
	expectedRoles := map[Role]string{
		RoleSystem:    "system",
		RoleUser:      "user",
		RoleAssistant: "assistant",
		RoleTool:      "tool",
	}

	for role, expectedStr := range expectedRoles {
		if string(role) != expectedStr {
			t.Errorf("Expected role %s to be %s, got %s", role, expectedStr, string(role))
		}
	}
}

func TestUsage_Structure(t *testing.T) {
	usage := Usage{
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
		Cost:             0.01,
	}

	if usage.PromptTokens != 100 {
		t.Errorf("Expected PromptTokens 100, got %d", usage.PromptTokens)
	}

	if usage.CompletionTokens != 50 {
		t.Errorf("Expected CompletionTokens 50, got %d", usage.CompletionTokens)
	}

	if usage.TotalTokens != 150 {
		t.Errorf("Expected TotalTokens 150, got %d", usage.TotalTokens)
	}

	if usage.Cost != 0.01 {
		t.Errorf("Expected Cost 0.01, got %f", usage.Cost)
	}
}

func TestResponse_Structure(t *testing.T) {
	response := Response{
		Content: "Hello, world!",
		Usage: Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
			Cost:             0.001,
		},
		Model:        "gpt-4",
		FinishReason: "stop",
		Metadata:     map[string]interface{}{"test": "value"},
	}

	if response.Content != "Hello, world!" {
		t.Errorf("Expected content 'Hello, world!', got %s", response.Content)
	}

	if response.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got %s", response.Model)
	}

	if response.FinishReason != "stop" {
		t.Errorf("Expected finish reason 'stop', got %s", response.FinishReason)
	}

	if response.Metadata["test"] != "value" {
		t.Errorf("Expected metadata test=value, got %v", response.Metadata["test"])
	}
}

func TestCallOptions_Functional(t *testing.T) {
	opts := DefaultCallOptions()

	// Test WithTemperature
	WithTemperature(0.7)(opts)
	if opts.Temperature == nil || *opts.Temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %v", opts.Temperature)
	}

	// Test WithMaxTokens
	WithMaxTokens(1000)(opts)
	if opts.MaxTokens == nil || *opts.MaxTokens != 1000 {
		t.Errorf("Expected max tokens 1000, got %v", opts.MaxTokens)
	}

	// Test WithTopP
	WithTopP(0.9)(opts)
	if opts.TopP == nil || *opts.TopP != 0.9 {
		t.Errorf("Expected top_p 0.9, got %v", opts.TopP)
	}

	// Test WithStopSequences
	stops := []string{"STOP", "END"}
	WithStopSequences(stops)(opts)
	if len(opts.StopSequences) != 2 || opts.StopSequences[0] != "STOP" || opts.StopSequences[1] != "END" {
		t.Errorf("Expected stop sequences %v, got %v", stops, opts.StopSequences)
	}

	// Test WithStream
	WithStream(true)(opts)
	if !opts.Stream {
		t.Errorf("Expected stream true, got %v", opts.Stream)
	}

	// Test WithMetadata
	metadata := map[string]interface{}{"key": "value"}
	WithMetadata(metadata)(opts)
	if opts.Metadata["key"] != "value" {
		t.Errorf("Expected metadata key=value, got %v", opts.Metadata["key"])
	}
}

func TestCallOptions_ApplyOptions(t *testing.T) {
	opts := DefaultCallOptions()

	opts.ApplyOptions(
		WithTemperature(0.5),
		WithMaxTokens(500),
		WithStream(true),
	)

	if opts.Temperature == nil || *opts.Temperature != 0.5 {
		t.Errorf("Expected temperature 0.5, got %v", opts.Temperature)
	}

	if opts.MaxTokens == nil || *opts.MaxTokens != 500 {
		t.Errorf("Expected max tokens 500, got %v", opts.MaxTokens)
	}

	if !opts.Stream {
		t.Errorf("Expected stream true, got %v", opts.Stream)
	}
}

func TestToolSchema_Structure(t *testing.T) {
	schema := ToolSchema{
		Name:        "get_weather",
		Description: "Get current weather",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"location": map[string]interface{}{
					"type":        "string",
					"description": "The city name",
				},
			},
		},
	}

	if schema.Name != "get_weather" {
		t.Errorf("Expected name 'get_weather', got %s", schema.Name)
	}

	if schema.Description != "Get current weather" {
		t.Errorf("Expected description 'Get current weather', got %s", schema.Description)
	}

	if schema.Parameters["type"] != "object" {
		t.Errorf("Expected type 'object', got %v", schema.Parameters["type"])
	}
}

func TestConfig_Structure(t *testing.T) {
	config := Config{
		Provider:    "openai",
		Model:       "gpt-4",
		APIKey:      "test-key",
		BaseURL:     "https://api.openai.com/v1",
		MaxRetries:  3,
		Temperature: func() *float64 { t := 0.7; return &t }(),
		MaxTokens:   func() *int { t := 1000; return &t }(),
		Metadata:    map[string]interface{}{"test": "value"},
	}

	if config.Provider != "openai" {
		t.Errorf("Expected provider 'openai', got %s", config.Provider)
	}

	if config.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got %s", config.Model)
	}

	if config.APIKey != "test-key" {
		t.Errorf("Expected API key 'test-key', got %s", config.APIKey)
	}

	if *config.Temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %f", *config.Temperature)
	}

	if *config.MaxTokens != 1000 {
		t.Errorf("Expected max tokens 1000, got %d", *config.MaxTokens)
	}
}

// 新增：测试Python对标的新增参数选项
func TestCallOptions_PythonAlignmentOptions(t *testing.T) {
	tests := []struct {
		name         string
		applyOptions func(*CallOptions)
		verify       func(*testing.T, *CallOptions)
	}{
		{
			name: "max_completion_tokens",
			applyOptions: func(opts *CallOptions) {
				opts.ApplyOptions(WithMaxCompletionTokens(500))
			},
			verify: func(t *testing.T, opts *CallOptions) {
				if opts.MaxCompletionTokens == nil || *opts.MaxCompletionTokens != 500 {
					t.Errorf("Expected MaxCompletionTokens 500, got %v", opts.MaxCompletionTokens)
				}
			},
		},
		{
			name: "frequency_penalty",
			applyOptions: func(opts *CallOptions) {
				opts.ApplyOptions(WithFrequencyPenalty(0.5))
			},
			verify: func(t *testing.T, opts *CallOptions) {
				if opts.FrequencyPenalty == nil || *opts.FrequencyPenalty != 0.5 {
					t.Errorf("Expected FrequencyPenalty 0.5, got %v", opts.FrequencyPenalty)
				}
			},
		},
		{
			name: "presence_penalty",
			applyOptions: func(opts *CallOptions) {
				opts.ApplyOptions(WithPresencePenalty(0.3))
			},
			verify: func(t *testing.T, opts *CallOptions) {
				if opts.PresencePenalty == nil || *opts.PresencePenalty != 0.3 {
					t.Errorf("Expected PresencePenalty 0.3, got %v", opts.PresencePenalty)
				}
			},
		},
		{
			name: "n_responses",
			applyOptions: func(opts *CallOptions) {
				opts.ApplyOptions(WithN(3))
			},
			verify: func(t *testing.T, opts *CallOptions) {
				if opts.N == nil || *opts.N != 3 {
					t.Errorf("Expected N 3, got %v", opts.N)
				}
			},
		},
		{
			name: "seed",
			applyOptions: func(opts *CallOptions) {
				opts.ApplyOptions(WithSeed(12345))
			},
			verify: func(t *testing.T, opts *CallOptions) {
				if opts.Seed == nil || *opts.Seed != 12345 {
					t.Errorf("Expected Seed 12345, got %v", opts.Seed)
				}
			},
		},
		{
			name: "response_format",
			applyOptions: func(opts *CallOptions) {
				format := map[string]interface{}{"type": "json_object"}
				opts.ApplyOptions(WithResponseFormat(format))
			},
			verify: func(t *testing.T, opts *CallOptions) {
				expected := map[string]interface{}{"type": "json_object"}
				if opts.ResponseFormat == nil {
					t.Error("Expected ResponseFormat to be set")
					return
				}
				formatMap, ok := opts.ResponseFormat.(map[string]interface{})
				if !ok {
					t.Errorf("Expected ResponseFormat to be map[string]interface{}, got %T", opts.ResponseFormat)
					return
				}
				if formatMap["type"] != expected["type"] {
					t.Errorf("Expected ResponseFormat %v, got %v", expected, formatMap)
				}
			},
		},
		{
			name: "logit_bias",
			applyOptions: func(opts *CallOptions) {
				bias := map[int]float64{50256: -100}
				opts.ApplyOptions(WithLogitBias(bias))
			},
			verify: func(t *testing.T, opts *CallOptions) {
				if opts.LogitBias == nil {
					t.Error("Expected LogitBias to be set")
					return
				}
				if len(opts.LogitBias) != 1 || opts.LogitBias[50256] != -100 {
					t.Errorf("Expected LogitBias map[50256:-100], got %v", opts.LogitBias)
				}
			},
		},
		{
			name: "user",
			applyOptions: func(opts *CallOptions) {
				opts.ApplyOptions(WithUser("test_user_123"))
			},
			verify: func(t *testing.T, opts *CallOptions) {
				if opts.User != "test_user_123" {
					t.Errorf("Expected User 'test_user_123', got '%s'", opts.User)
				}
			},
		},
		{
			name: "callbacks",
			applyOptions: func(opts *CallOptions) {
				callbacks := []interface{}{"callback1", "callback2"}
				opts.ApplyOptions(WithCallbacks(callbacks))
			},
			verify: func(t *testing.T, opts *CallOptions) {
				if len(opts.Callbacks) != 2 {
					t.Errorf("Expected 2 callbacks, got %d", len(opts.Callbacks))
					return
				}
				if opts.Callbacks[0] != "callback1" || opts.Callbacks[1] != "callback2" {
					t.Errorf("Expected callbacks [callback1, callback2], got %v", opts.Callbacks)
				}
			},
		},
		{
			name: "available_functions",
			applyOptions: func(opts *CallOptions) {
				functions := map[string]interface{}{
					"func1": "handler1",
					"func2": "handler2",
				}
				opts.ApplyOptions(WithAvailableFunctions(functions))
			},
			verify: func(t *testing.T, opts *CallOptions) {
				if len(opts.AvailableFunctions) != 2 {
					t.Errorf("Expected 2 available functions, got %d", len(opts.AvailableFunctions))
					return
				}
				if opts.AvailableFunctions["func1"] != "handler1" {
					t.Errorf("Expected func1 -> handler1, got %v", opts.AvailableFunctions["func1"])
				}
			},
		},
		{
			name: "stream_options",
			applyOptions: func(opts *CallOptions) {
				streamOpts := map[string]interface{}{"include_usage": true}
				opts.ApplyOptions(WithStreamOptions(streamOpts))
			},
			verify: func(t *testing.T, opts *CallOptions) {
				if len(opts.StreamOptions) != 1 {
					t.Errorf("Expected 1 stream option, got %d", len(opts.StreamOptions))
					return
				}
				if opts.StreamOptions["include_usage"] != true {
					t.Errorf("Expected include_usage -> true, got %v", opts.StreamOptions["include_usage"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultCallOptions()
			tt.applyOptions(opts)
			tt.verify(t, opts)
		})
	}
}

// 测试组合多个Python对标参数
func TestCallOptions_CombinedPythonOptions(t *testing.T) {
	opts := DefaultCallOptions()

	// 应用多个Python对标参数
	opts.ApplyOptions(
		WithMaxCompletionTokens(1000),
		WithFrequencyPenalty(0.1),
		WithPresencePenalty(0.2),
		WithN(2),
		WithSeed(42),
		WithUser("test_user"),
		WithLogitBias(map[int]float64{100: -50}),
	)

	// 验证所有参数都被正确设置
	if opts.MaxCompletionTokens == nil || *opts.MaxCompletionTokens != 1000 {
		t.Errorf("Expected MaxCompletionTokens 1000, got %v", opts.MaxCompletionTokens)
	}
	if opts.FrequencyPenalty == nil || *opts.FrequencyPenalty != 0.1 {
		t.Errorf("Expected FrequencyPenalty 0.1, got %v", opts.FrequencyPenalty)
	}
	if opts.PresencePenalty == nil || *opts.PresencePenalty != 0.2 {
		t.Errorf("Expected PresencePenalty 0.2, got %v", opts.PresencePenalty)
	}
	if opts.N == nil || *opts.N != 2 {
		t.Errorf("Expected N 2, got %v", opts.N)
	}
	if opts.Seed == nil || *opts.Seed != 42 {
		t.Errorf("Expected Seed 42, got %v", opts.Seed)
	}
	if opts.User != "test_user" {
		t.Errorf("Expected User 'test_user', got '%s'", opts.User)
	}
	if opts.LogitBias[100] != -50 {
		t.Errorf("Expected LogitBias[100] = -50, got %v", opts.LogitBias[100])
	}
}
