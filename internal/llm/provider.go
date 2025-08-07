package llm

import (
	"fmt"
	"sync"
	"time"
)

// ProviderRegistry manages LLM providers
type ProviderRegistry struct {
	providers map[string]Provider
	mu        sync.RWMutex
}

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry() *ProviderRegistry {
	registry := &ProviderRegistry{
		providers: make(map[string]Provider),
	}

	// Register built-in providers
	registry.RegisterProvider(&OpenAIProvider{})

	return registry
}

// RegisterProvider registers a new LLM provider
func (r *ProviderRegistry) RegisterProvider(provider Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[provider.Name()] = provider
}

// GetProvider retrieves a provider by name
func (r *ProviderRegistry) GetProvider(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider '%s' not found", name)
	}

	return provider, nil
}

// ListProviders returns all registered provider names
func (r *ProviderRegistry) ListProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}

	return names
}

// CreateLLM creates an LLM instance using the specified provider
func (r *ProviderRegistry) CreateLLM(config *Config) (LLM, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	provider, err := r.GetProvider(config.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Convert config to map for provider
	configMap := map[string]interface{}{
		"model":       config.Model,
		"api_key":     config.APIKey,
		"base_url":    config.BaseURL,
		"timeout":     config.Timeout,
		"max_retries": config.MaxRetries,
		"temperature": config.Temperature,
		"max_tokens":  config.MaxTokens,
		"metadata":    config.Metadata,
	}

	return provider.CreateLLM(configMap)
}

// Global registry instance
var globalRegistry = NewProviderRegistry()

// RegisterProvider registers a provider with the global registry
func RegisterProvider(provider Provider) {
	globalRegistry.RegisterProvider(provider)
}

// GetProvider retrieves a provider from the global registry
func GetProvider(name string) (Provider, error) {
	return globalRegistry.GetProvider(name)
}

// ListProviders returns all provider names from the global registry
func ListProviders() []string {
	return globalRegistry.ListProviders()
}

// CreateLLM creates an LLM using the global registry
func CreateLLM(config *Config) (LLM, error) {
	return globalRegistry.CreateLLM(config)
}

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct{}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// CreateLLM creates a new OpenAI LLM instance
func (p *OpenAIProvider) CreateLLM(config map[string]interface{}) (LLM, error) {
	model, ok := config["model"].(string)
	if !ok || model == "" {
		return nil, fmt.Errorf("model is required for OpenAI provider")
	}

	var options []BaseLLMOption

	if apiKey, ok := config["api_key"].(string); ok && apiKey != "" {
		options = append(options, WithAPIKey(apiKey))
	}

	if baseURL, ok := config["base_url"].(string); ok && baseURL != "" {
		options = append(options, WithBaseURL(baseURL))
	}

	if timeout, ok := config["timeout"]; ok {
		switch v := timeout.(type) {
		case int:
			options = append(options, WithTimeout(time.Duration(v)*time.Second))
		case int64:
			options = append(options, WithTimeout(time.Duration(v)*time.Second))
		case float64:
			options = append(options, WithTimeout(time.Duration(v)*time.Second))
		case time.Duration:
			options = append(options, WithTimeout(v))
		}
	}

	if maxRetries, ok := config["max_retries"].(int); ok {
		options = append(options, WithMaxRetries(maxRetries))
	}

	return NewOpenAILLM(model, options...), nil
}

// SupportedModels returns the list of supported OpenAI models
func (p *OpenAIProvider) SupportedModels() []string {
	models := make([]string, 0, len(openAIContextWindows))
	for model := range openAIContextWindows {
		models = append(models, model)
	}
	return models
}
