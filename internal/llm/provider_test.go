package llm

import (
	"context"
	"testing"
	"time"

	"github.com/ynl/greensoulai/pkg/events"
)

func TestNewProviderRegistry(t *testing.T) {
	registry := NewProviderRegistry()

	if registry == nil {
		t.Fatal("Expected registry to be created")
	}

	if registry.providers == nil {
		t.Fatal("Expected providers map to be initialized")
	}

	// Check if OpenAI provider is registered by default
	provider, err := registry.GetProvider("openai")
	if err != nil {
		t.Errorf("Expected OpenAI provider to be registered by default, got error: %v", err)
	}

	if provider == nil {
		t.Error("Expected OpenAI provider to be non-nil")
	}
}

func TestProviderRegistry_RegisterProvider(t *testing.T) {
	registry := NewProviderRegistry()

	// Create a mock provider
	mockProvider := &MockProvider{
		name:            "mock",
		supportedModels: []string{"mock-model-1", "mock-model-2"},
	}

	registry.RegisterProvider(mockProvider)

	// Try to get the registered provider
	provider, err := registry.GetProvider("mock")
	if err != nil {
		t.Errorf("Expected to retrieve registered provider, got error: %v", err)
	}

	if provider != mockProvider {
		t.Error("Expected to retrieve the same provider instance")
	}
}

func TestProviderRegistry_GetProvider(t *testing.T) {
	registry := NewProviderRegistry()

	// Test getting existing provider
	provider, err := registry.GetProvider("openai")
	if err != nil {
		t.Errorf("Expected to get OpenAI provider, got error: %v", err)
	}

	if provider == nil {
		t.Error("Expected provider to be non-nil")
	}

	// Test getting non-existing provider
	_, err = registry.GetProvider("non-existing")
	if err == nil {
		t.Error("Expected error for non-existing provider")
	}

	expectedErrMsg := "provider 'non-existing' not found"
	if err.Error() != expectedErrMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

func TestProviderRegistry_ListProviders(t *testing.T) {
	registry := NewProviderRegistry()

	providers := registry.ListProviders()

	// Should have at least OpenAI provider
	if len(providers) == 0 {
		t.Error("Expected at least one provider")
	}

	// Check if OpenAI is in the list
	found := false
	for _, name := range providers {
		if name == "openai" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected 'openai' to be in the providers list")
	}

	// Add another provider and check
	mockProvider := &MockProvider{name: "mock"}
	registry.RegisterProvider(mockProvider)

	providers = registry.ListProviders()
	if len(providers) < 2 {
		t.Error("Expected at least two providers after adding mock")
	}
}

func TestProviderRegistry_CreateLLM(t *testing.T) {
	registry := NewProviderRegistry()

	config := &Config{
		Provider:    "openai",
		Model:       "gpt-4",
		APIKey:      "test-key",
		BaseURL:     "https://api.openai.com/v1",
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		Temperature: func() *float64 { t := 0.7; return &t }(),
		MaxTokens:   func() *int { t := 1000; return &t }(),
		Metadata:    map[string]interface{}{"test": "value"},
	}

	llm, err := registry.CreateLLM(config)
	if err != nil {
		t.Errorf("Expected to create LLM, got error: %v", err)
	}

	if llm == nil {
		t.Fatal("Expected LLM to be created")
	}

	if llm.GetModel() != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got %s", llm.GetModel())
	}
}

func TestProviderRegistry_CreateLLM_NilConfig(t *testing.T) {
	registry := NewProviderRegistry()

	_, err := registry.CreateLLM(nil)
	if err == nil {
		t.Error("Expected error for nil config")
	}

	expectedErrMsg := "config cannot be nil"
	if err.Error() != expectedErrMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

func TestProviderRegistry_CreateLLM_InvalidProvider(t *testing.T) {
	registry := NewProviderRegistry()

	config := &Config{
		Provider: "invalid",
		Model:    "test-model",
	}

	_, err := registry.CreateLLM(config)
	if err == nil {
		t.Error("Expected error for invalid provider")
	}
}

func TestGlobalRegistryFunctions(t *testing.T) {
	// Test global functions work
	providers := ListProviders()
	if len(providers) == 0 {
		t.Error("Expected at least one provider in global registry")
	}

	provider, err := GetProvider("openai")
	if err != nil {
		t.Errorf("Expected to get OpenAI provider from global registry, got error: %v", err)
	}

	if provider == nil {
		t.Error("Expected provider to be non-nil")
	}

	// Test registering with global registry
	mockProvider := &MockProvider{name: "global-mock"}
	RegisterProvider(mockProvider)

	retrievedProvider, err := GetProvider("global-mock")
	if err != nil {
		t.Errorf("Expected to get registered provider from global registry, got error: %v", err)
	}

	if retrievedProvider != mockProvider {
		t.Error("Expected to get the same provider instance from global registry")
	}
}

func TestCreateLLM_GlobalRegistry(t *testing.T) {
	config := &Config{
		Provider: "openai",
		Model:    "gpt-3.5-turbo",
		APIKey:   "test-key",
	}

	llm, err := CreateLLM(config)
	if err != nil {
		t.Errorf("Expected to create LLM with global registry, got error: %v", err)
	}

	if llm == nil {
		t.Fatal("Expected LLM to be created")
	}

	if llm.GetModel() != "gpt-3.5-turbo" {
		t.Errorf("Expected model 'gpt-3.5-turbo', got %s", llm.GetModel())
	}
}

func TestOpenAIProvider_Name(t *testing.T) {
	provider := &OpenAIProvider{}

	if provider.Name() != "openai" {
		t.Errorf("Expected provider name 'openai', got %s", provider.Name())
	}
}

func TestOpenAIProvider_CreateLLM(t *testing.T) {
	provider := &OpenAIProvider{}

	// Test successful creation
	config := map[string]interface{}{
		"model":       "gpt-4",
		"api_key":     "test-key",
		"base_url":    "https://api.openai.com/v1",
		"timeout":     30,
		"max_retries": 3,
	}

	llm, err := provider.CreateLLM(config)
	if err != nil {
		t.Errorf("Expected to create LLM, got error: %v", err)
	}

	if llm == nil {
		t.Fatal("Expected LLM to be created")
	}

	openAILLM, ok := llm.(*OpenAILLM)
	if !ok {
		t.Fatal("Expected LLM to be OpenAILLM instance")
	}

	if openAILLM.GetModel() != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got %s", openAILLM.GetModel())
	}

	if openAILLM.GetAPIKey() != "test-key" {
		t.Errorf("Expected API key 'test-key', got %s", openAILLM.GetAPIKey())
	}

	if openAILLM.GetBaseURL() != "https://api.openai.com/v1" {
		t.Errorf("Expected base URL 'https://api.openai.com/v1', got %s", openAILLM.GetBaseURL())
	}

	if openAILLM.GetTimeout() != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", openAILLM.GetTimeout())
	}

	if openAILLM.GetMaxRetries() != 3 {
		t.Errorf("Expected max retries 3, got %d", openAILLM.GetMaxRetries())
	}
}

func TestOpenAIProvider_CreateLLM_NoModel(t *testing.T) {
	provider := &OpenAIProvider{}

	config := map[string]interface{}{
		"api_key": "test-key",
	}

	_, err := provider.CreateLLM(config)
	if err == nil {
		t.Error("Expected error for missing model")
	}

	expectedErrMsg := "model is required for OpenAI provider"
	if err.Error() != expectedErrMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, err.Error())
	}
}

func TestOpenAIProvider_CreateLLM_EmptyModel(t *testing.T) {
	provider := &OpenAIProvider{}

	config := map[string]interface{}{
		"model":   "",
		"api_key": "test-key",
	}

	_, err := provider.CreateLLM(config)
	if err == nil {
		t.Error("Expected error for empty model")
	}
}

func TestOpenAIProvider_CreateLLM_DifferentTimeoutTypes(t *testing.T) {
	provider := &OpenAIProvider{}

	testCases := []struct {
		name         string
		timeout      interface{}
		expectedTime time.Duration
	}{
		{
			name:         "int timeout",
			timeout:      30,
			expectedTime: 30 * time.Second,
		},
		{
			name:         "int64 timeout",
			timeout:      int64(45),
			expectedTime: 45 * time.Second,
		},
		{
			name:         "float64 timeout",
			timeout:      60.0,
			expectedTime: 60 * time.Second,
		},
		{
			name:         "duration timeout",
			timeout:      90 * time.Second,
			expectedTime: 90 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := map[string]interface{}{
				"model":   "gpt-4",
				"api_key": "test-key",
				"timeout": tc.timeout,
			}

			llm, err := provider.CreateLLM(config)
			if err != nil {
				t.Errorf("Expected to create LLM, got error: %v", err)
			}

			openAILLM, ok := llm.(*OpenAILLM)
			if !ok {
				t.Fatal("Expected LLM to be OpenAILLM instance")
			}

			if openAILLM.GetTimeout() != tc.expectedTime {
				t.Errorf("Expected timeout %v, got %v", tc.expectedTime, openAILLM.GetTimeout())
			}
		})
	}
}

func TestOpenAIProvider_SupportedModels(t *testing.T) {
	provider := &OpenAIProvider{}

	models := provider.SupportedModels()

	if len(models) == 0 {
		t.Error("Expected at least one supported model")
	}

	// Check if known models are in the list
	expectedModels := []string{"gpt-4", "gpt-3.5-turbo"}
	for _, expectedModel := range expectedModels {
		found := false
		for _, model := range models {
			if model == expectedModel {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected model '%s' to be in supported models", expectedModel)
		}
	}
}

// MockProvider for testing
type MockProvider struct {
	name            string
	supportedModels []string
	createLLMFunc   func(map[string]interface{}) (LLM, error)
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) CreateLLM(config map[string]interface{}) (LLM, error) {
	if m.createLLMFunc != nil {
		return m.createLLMFunc(config)
	}

	// Default implementation for testing
	model, ok := config["model"].(string)
	if !ok || model == "" {
		return nil, &MockError{message: "model is required"}
	}

	return &MockLLM{model: model}, nil
}

func (m *MockProvider) SupportedModels() []string {
	if m.supportedModels != nil {
		return m.supportedModels
	}
	return []string{"mock-model"}
}

// MockLLM for testing
type MockLLM struct {
	model string
}

func (m *MockLLM) Call(ctx context.Context, messages []Message, options *CallOptions) (*Response, error) {
	return &Response{
		Content: "Mock response",
		Model:   m.model,
	}, nil
}

func (m *MockLLM) CallStream(ctx context.Context, messages []Message, options *CallOptions) (<-chan StreamResponse, error) {
	ch := make(chan StreamResponse, 1)
	go func() {
		defer close(ch)
		ch <- StreamResponse{Delta: "Mock stream response"}
	}()
	return ch, nil
}

func (m *MockLLM) GetModel() string {
	return m.model
}

func (m *MockLLM) SupportsFunctionCalling() bool {
	return false
}

func (m *MockLLM) GetContextWindowSize() int {
	return 4096
}

func (m *MockLLM) SetEventBus(eventBus events.EventBus) {
	// Mock implementation
}

func (m *MockLLM) Close() error {
	return nil
}

// MockError for testing
type MockError struct {
	message string
}

func (e *MockError) Error() string {
	return e.message
}
