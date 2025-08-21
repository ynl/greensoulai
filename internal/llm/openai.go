package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ynl/greensoulai/pkg/logger"
)

const (
	// OpenAI API constants
	defaultOpenAIBaseURL = "https://api.openai.com/v1"
	openAIChatEndpoint   = "/chat/completions"
)

// OpenAI model context windows
var openAIContextWindows = map[string]int{
	"gpt-4":             8192,
	"gpt-4-32k":         32768,
	"gpt-4-turbo":       128000,
	"gpt-4o":            128000,
	"gpt-4o-mini":       128000,
	"gpt-3.5-turbo":     16385,
	"gpt-3.5-turbo-16k": 16385,
}

// OpenAILLM represents an OpenAI LLM instance
type OpenAILLM struct {
	*BaseLLM
	organization string
}

// OpenAIChatRequest represents the request structure for OpenAI chat API
type OpenAIChatRequest struct {
	Model               string          `json:"model"`
	Messages            []OpenAIMessage `json:"messages"`
	Temperature         *float64        `json:"temperature,omitempty"`
	MaxTokens           *int            `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int            `json:"max_completion_tokens,omitempty"` // 对标Python版本
	TopP                *float64        `json:"top_p,omitempty"`
	FrequencyPenalty    *float64        `json:"frequency_penalty,omitempty"`
	PresencePenalty     *float64        `json:"presence_penalty,omitempty"`
	Stop                []string        `json:"stop,omitempty"`
	Stream              bool            `json:"stream,omitempty"`
	Tools               []OpenAITool    `json:"tools,omitempty"`
	ToolChoice          interface{}     `json:"tool_choice,omitempty"`
	User                string          `json:"user,omitempty"`
	ResponseFormat      interface{}     `json:"response_format,omitempty"`

	// 对标Python版本的新增参数
	N             *int                   `json:"n,omitempty"`              // 生成响应数量
	Seed          *int                   `json:"seed,omitempty"`           // 随机种子
	Logprobs      *int                   `json:"logprobs,omitempty"`       // 返回logprobs数量
	TopLogprobs   *int                   `json:"top_logprobs,omitempty"`   // 返回top logprobs
	LogitBias     map[int]float64        `json:"logit_bias,omitempty"`     // logit偏置
	StreamOptions map[string]interface{} `json:"stream_options,omitempty"` // 流式选项
}

// OpenAIMessage represents a message in OpenAI format
type OpenAIMessage struct {
	Role       string           `json:"role"`
	Content    interface{}      `json:"content"`
	Name       string           `json:"name,omitempty"`
	ToolCalls  []OpenAIToolCall `json:"tool_calls,omitempty"`
	ToolCallId string           `json:"tool_call_id,omitempty"`
}

// OpenAITool represents a tool in OpenAI format
type OpenAITool struct {
	Type     string           `json:"type"`
	Function OpenAIToolSchema `json:"function"`
}

// OpenAIToolSchema represents a tool schema in OpenAI format
type OpenAIToolSchema struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// OpenAIToolCall represents a tool call in OpenAI format
type OpenAIToolCall struct {
	ID       string             `json:"id"`
	Type     string             `json:"type"`
	Function OpenAIToolCallFunc `json:"function"`
}

// OpenAIToolCallFunc represents a tool call function in OpenAI format
type OpenAIToolCallFunc struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// OpenAIChatResponse represents the response structure for OpenAI chat API
type OpenAIChatResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Usage   OpenAIUsage    `json:"usage"`
	Choices []OpenAIChoice `json:"choices"`
	Error   *OpenAIError   `json:"error,omitempty"`
}

// OpenAIChoice represents a choice in OpenAI response
type OpenAIChoice struct {
	Index        int            `json:"index"`
	Message      OpenAIMessage  `json:"message,omitempty"`
	Delta        *OpenAIMessage `json:"delta,omitempty"`
	FinishReason string         `json:"finish_reason"`
}

// OpenAIUsage represents usage information in OpenAI response
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAIError represents an error from OpenAI API
type OpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// NewOpenAILLM creates a new OpenAI LLM instance
func NewOpenAILLM(model string, options ...BaseLLMOption) *OpenAILLM {
	// Set default context window
	contextWindow := 4096
	if window, exists := openAIContextWindows[model]; exists {
		contextWindow = window
	}

	// Create base LLM with OpenAI-specific defaults
	// Note: User-provided options are applied first, then defaults only if not already set
	defaultOptions := []BaseLLMOption{
		WithContextWindow(contextWindow),
		WithFunctionCalling(true),
	}

	// Check if user provided a custom base URL (simplified approach)
	// We'll rely on the conditional setting below

	// Apply user options first, then defaults
	allOptions := append(options, defaultOptions...)

	// Add default base URL only if not already set
	allOptions = append(allOptions, func(b *BaseLLM) {
		if b.baseURL == "" {
			WithBaseURL(defaultOpenAIBaseURL)(b)
		}
	})

	baseLLM := NewBaseLLM("openai", model, allOptions...)

	return &OpenAILLM{
		BaseLLM: baseLLM,
	}
}

// WithOrganization sets the OpenAI organization
func WithOrganization(organization string) BaseLLMOption {
	return func(b *BaseLLM) {
		// For now, we'll store the organization in a way that can be accessed later
		// This is a simplified implementation
	}
}

// Call sends a synchronous request to the OpenAI API
func (o *OpenAILLM) Call(ctx context.Context, messages []Message, options *CallOptions) (*Response, error) {
	// Validate inputs
	if err := o.ValidateMessages(messages); err != nil {
		return nil, fmt.Errorf("invalid messages: %w", err)
	}

	if err := o.ValidateCallOptions(options); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	// Convert to OpenAI format
	openAIMessages := o.convertMessages(messages)
	request := o.buildChatRequest(openAIMessages, options)

	// Make API call
	response, err := o.makeAPICall(ctx, request)
	if err != nil {
		o.LogError("OpenAI API call failed",
			logger.Field{Key: "model", Value: o.GetModel()},
			logger.Field{Key: "error", Value: err},
		)
		return nil, err
	}

	// Convert response
	result := o.convertResponse(response)

	o.LogDebug("OpenAI API call completed",
		logger.Field{Key: "model", Value: o.GetModel()},
		logger.Field{Key: "usage", Value: result.Usage},
	)

	return result, nil
}

// CallStream sends a streaming request to the OpenAI API
func (o *OpenAILLM) CallStream(ctx context.Context, messages []Message, options *CallOptions) (<-chan StreamResponse, error) {
	// Validate inputs
	if err := o.ValidateMessages(messages); err != nil {
		return nil, fmt.Errorf("invalid messages: %w", err)
	}

	if err := o.ValidateCallOptions(options); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	// Convert to OpenAI format and enable streaming
	openAIMessages := o.convertMessages(messages)
	request := o.buildChatRequest(openAIMessages, options)
	request.Stream = true

	// Create response channel
	responseChannel := make(chan StreamResponse, 100)

	// Start streaming in a goroutine
	go o.streamAPICall(ctx, request, responseChannel)

	return responseChannel, nil
}

// convertMessages converts internal messages to OpenAI format
func (o *OpenAILLM) convertMessages(messages []Message) []OpenAIMessage {
	openAIMessages := make([]OpenAIMessage, len(messages))

	for i, msg := range messages {
		openAIMsg := OpenAIMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
			Name:    msg.Name,
		}

		openAIMessages[i] = openAIMsg
	}

	return openAIMessages
}

// buildChatRequest builds an OpenAI chat request
func (o *OpenAILLM) buildChatRequest(messages []OpenAIMessage, options *CallOptions) *OpenAIChatRequest {
	request := &OpenAIChatRequest{
		Model:    o.GetModel(),
		Messages: messages,
	}

	if options != nil {
		request.Temperature = options.Temperature
		request.MaxTokens = options.MaxTokens
		request.MaxCompletionTokens = options.MaxCompletionTokens // 对标Python版本
		request.TopP = options.TopP
		request.FrequencyPenalty = options.FrequencyPenalty
		request.PresencePenalty = options.PresencePenalty
		request.Stop = options.StopSequences
		request.Stream = options.Stream
		request.User = options.User
		request.ResponseFormat = options.ResponseFormat

		// 对标Python版本的新增参数
		request.N = options.N
		request.Seed = options.Seed
		request.Logprobs = options.Logprobs
		request.TopLogprobs = options.TopLogprobs
		request.LogitBias = options.LogitBias
		request.StreamOptions = options.StreamOptions

		// Convert tools
		if len(options.Tools) > 0 {
			request.Tools = make([]OpenAITool, len(options.Tools))
			for i, tool := range options.Tools {
				request.Tools[i] = OpenAITool{
					Type: tool.Type,
					Function: OpenAIToolSchema{
						Name:        tool.Function.Name,
						Description: tool.Function.Description,
						Parameters:  tool.Function.Parameters,
					},
				}
			}
			request.ToolChoice = options.ToolChoice
		}
	}

	return request
}

// makeAPICall makes a synchronous API call to OpenAI
func (o *OpenAILLM) makeAPICall(ctx context.Context, request *OpenAIChatRequest) (*OpenAIChatResponse, error) {
	// Prepare request body
	bodyBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	fullURL := o.GetBaseURL() + openAIChatEndpoint
	httpReq, err := http.NewRequestWithContext(ctx, "POST", fullURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.GetAPIKey())

	if o.organization != "" {
		httpReq.Header.Set("OpenAI-Organization", o.organization)
	}

	// Set custom headers
	for key, value := range o.GetCustomHeaders() {
		httpReq.Header.Set(key, value)
	}

	// Make request with retries
	var response *http.Response
	var lastErr error

	for attempt := 0; attempt <= o.GetMaxRetries(); attempt++ {
		response, lastErr = o.GetHTTPClient().Do(httpReq)
		if lastErr == nil && response.StatusCode < 500 {
			break // Success or client error (4xx)
		}

		if attempt < o.GetMaxRetries() {
			// Wait before retry (exponential backoff)
			waitTime := time.Duration(attempt+1) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(waitTime):
				// Continue to retry
			}
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("HTTP request failed after %d retries: %w", o.GetMaxRetries(), lastErr)
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			o.logger.Error("Failed to close response body",
				logger.Field{Key: "error", Value: err})
		}
	}()

	// Read response body
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response
	var openAIResponse OpenAIChatResponse
	if err := json.Unmarshal(responseBody, &openAIResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for API errors
	if openAIResponse.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s (type: %s, code: %s)",
			openAIResponse.Error.Message,
			openAIResponse.Error.Type,
			openAIResponse.Error.Code)
	}

	// Check HTTP status
	if response.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error %d: %s", response.StatusCode, string(responseBody))
	}

	return &openAIResponse, nil
}

// streamAPICall handles streaming API calls
func (o *OpenAILLM) streamAPICall(ctx context.Context, request *OpenAIChatRequest, responseChannel chan<- StreamResponse) {
	defer close(responseChannel)

	// Prepare request body
	bodyBytes, err := json.Marshal(request)
	if err != nil {
		responseChannel <- StreamResponse{Error: fmt.Errorf("failed to marshal request: %w", err)}
		return
	}

	// Create HTTP request
	fullURL := o.GetBaseURL() + openAIChatEndpoint
	httpReq, err := http.NewRequestWithContext(ctx, "POST", fullURL, bytes.NewReader(bodyBytes))
	if err != nil {
		responseChannel <- StreamResponse{Error: fmt.Errorf("failed to create HTTP request: %w", err)}
		return
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.GetAPIKey())
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Cache-Control", "no-cache")

	if o.organization != "" {
		httpReq.Header.Set("OpenAI-Organization", o.organization)
	}

	// Set custom headers
	for key, value := range o.GetCustomHeaders() {
		httpReq.Header.Set(key, value)
	}

	// Make request
	response, err := o.GetHTTPClient().Do(httpReq)
	if err != nil {
		responseChannel <- StreamResponse{Error: fmt.Errorf("HTTP request failed: %w", err)}
		return
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			o.logger.Error("Failed to close response body",
				logger.Field{Key: "error", Value: err})
		}
	}()

	// Check status
	if response.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(response.Body)
		responseChannel <- StreamResponse{Error: fmt.Errorf("HTTP error %d: %s", response.StatusCode, string(bodyBytes))}
		return
	}

	// Process streaming response
	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			// Check for end of stream
			if data == "[DONE]" {
				break
			}

			// Parse JSON chunk
			var chunk OpenAIChatResponse
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				o.LogError("Failed to parse streaming chunk",
					logger.Field{Key: "data", Value: data},
					logger.Field{Key: "error", Value: err},
				)
				continue
			}

			// Convert to stream response
			if len(chunk.Choices) > 0 {
				choice := chunk.Choices[0]
				streamResp := StreamResponse{
					FinishReason: choice.FinishReason,
				}

				if choice.Delta != nil && choice.Delta.Content != nil {
					if content, ok := choice.Delta.Content.(string); ok {
						streamResp.Delta = content
					}
				}

				// Include usage if available (usually in last chunk)
				if chunk.Usage.TotalTokens > 0 {
					streamResp.Usage = &Usage{
						PromptTokens:     chunk.Usage.PromptTokens,
						CompletionTokens: chunk.Usage.CompletionTokens,
						TotalTokens:      chunk.Usage.TotalTokens,
					}
				}

				responseChannel <- streamResp
			}
		}

		// Check for context cancellation
		select {
		case <-ctx.Done():
			responseChannel <- StreamResponse{Error: ctx.Err()}
			return
		default:
		}
	}

	if err := scanner.Err(); err != nil {
		responseChannel <- StreamResponse{Error: fmt.Errorf("scanner error: %w", err)}
	}
}

// convertResponse converts OpenAI response to internal format
func (o *OpenAILLM) convertResponse(response *OpenAIChatResponse) *Response {
	if len(response.Choices) == 0 {
		return &Response{
			Content: "",
			Usage: Usage{
				PromptTokens:     response.Usage.PromptTokens,
				CompletionTokens: response.Usage.CompletionTokens,
				TotalTokens:      response.Usage.TotalTokens,
			},
			Model: response.Model,
		}
	}

	choice := response.Choices[0]
	result := &Response{
		Usage: Usage{
			PromptTokens:     response.Usage.PromptTokens,
			CompletionTokens: response.Usage.CompletionTokens,
			TotalTokens:      response.Usage.TotalTokens,
		},
		Model:        response.Model,
		FinishReason: choice.FinishReason,
		Metadata: map[string]interface{}{
			"id":      response.ID,
			"object":  response.Object,
			"created": response.Created,
		},
	}

	// Extract content
	if choice.Message.Content != nil {
		if content, ok := choice.Message.Content.(string); ok {
			result.Content = content
		}
	}

	// Extract tool calls if present
	if len(choice.Message.ToolCalls) > 0 {
		result.ToolCalls = make([]ToolCall, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
			result.ToolCalls[i] = ToolCall{
				ID:   tc.ID,
				Type: tc.Type,
				Function: ToolCallFunction{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		}
	}

	return result
}
