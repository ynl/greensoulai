//go:build test
// +build test

package agent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ynl/greensoulai/internal/llm"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// ===== Mock LLM =====

// MockLLM 基础模拟LLM，适用于简单测试场景
type MockLLM struct {
	model      string
	response   *llm.Response
	shouldFail bool
	callCount  int
}

// NewMockLLM 创建基础模拟LLM
func NewMockLLM(response *llm.Response, shouldFail bool) *MockLLM {
	return &MockLLM{
		model:      "mock-model",
		response:   response,
		shouldFail: shouldFail,
		callCount:  0,
	}
}

func (m *MockLLM) Call(ctx context.Context, messages []llm.Message, options *llm.CallOptions) (*llm.Response, error) {
	m.callCount++
	if m.shouldFail {
		return nil, errors.New("mock LLM error")
	}
	return m.response, nil
}

func (m *MockLLM) Stream(ctx context.Context, messages []llm.Message, options *llm.CallOptions) (<-chan llm.StreamResponse, error) {
	ch := make(chan llm.StreamResponse, 1)
	defer close(ch)
	if m.shouldFail {
		ch <- llm.StreamResponse{Error: errors.New("mock stream error")}
		return ch, nil
	}
	ch <- llm.StreamResponse{Delta: m.response.Content}
	return ch, nil
}

func (m *MockLLM) CallStream(ctx context.Context, messages []llm.Message, options *llm.CallOptions) (<-chan llm.StreamResponse, error) {
	return m.Stream(ctx, messages, options)
}

func (m *MockLLM) GetProvider() string                  { return "mock" }
func (m *MockLLM) GetModel() string                     { return m.model }
func (m *MockLLM) GetAPIKey() string                    { return "mock-key" }
func (m *MockLLM) GetBaseURL() string                   { return "https://mock.api" }
func (m *MockLLM) GetTimeout() time.Duration            { return 30 * time.Second }
func (m *MockLLM) GetMaxRetries() int                   { return 3 }
func (m *MockLLM) GetHTTPClient() interface{}           { return nil }
func (m *MockLLM) GetLogger() logger.Logger             { return logger.NewTestLogger() }
func (m *MockLLM) GetEventBus() events.EventBus         { return nil }
func (m *MockLLM) GetCustomHeaders() map[string]string  { return nil }
func (m *MockLLM) SupportsFunctionCalling() bool        { return true }
func (m *MockLLM) GetContextWindowSize() int            { return 4096 }
func (m *MockLLM) SetEventBus(eventBus events.EventBus) {}
func (m *MockLLM) Close() error                         { return nil }

// GetCallCount 返回LLM被调用的次数
func (m *MockLLM) GetCallCount() int {
	return m.callCount
}

// ResetCallCount 重置调用计数
func (m *MockLLM) ResetCallCount() {
	m.callCount = 0
}

// ===== Extended Mock LLM =====

// ExtendedMockLLM 扩展的模拟LLM，支持多个响应、提示捕获等高级功能
type ExtendedMockLLM struct {
	responses     []llm.Response
	currentIndex  int
	capturePrompt bool
	onCall        func([]llm.Message)
	shouldFail    bool
	callCount     int
}

// NewExtendedMockLLM 创建扩展模拟LLM
func NewExtendedMockLLM(responses []llm.Response) *ExtendedMockLLM {
	return &ExtendedMockLLM{
		responses:    responses,
		currentIndex: 0,
		shouldFail:   false,
		callCount:    0,
	}
}

// WithCallHandler 设置调用处理器用于捕获提示
func (m *ExtendedMockLLM) WithCallHandler(handler func([]llm.Message)) *ExtendedMockLLM {
	m.onCall = handler
	return m
}

// WithFailure 设置是否失败
func (m *ExtendedMockLLM) WithFailure(shouldFail bool) *ExtendedMockLLM {
	m.shouldFail = shouldFail
	return m
}

func (m *ExtendedMockLLM) Call(ctx context.Context, messages []llm.Message, options *llm.CallOptions) (*llm.Response, error) {
	m.callCount++

	if m.onCall != nil {
		m.onCall(messages)
	}

	if m.shouldFail {
		return nil, errors.New("mock LLM error")
	}

	if m.currentIndex >= len(m.responses) {
		// 如果没有更多响应，返回最后一个
		if len(m.responses) > 0 {
			return &m.responses[len(m.responses)-1], nil
		}
		return &llm.Response{
			Content: "default response",
			Usage:   llm.Usage{TotalTokens: 10},
			Model:   "mock",
		}, nil
	}

	response := m.responses[m.currentIndex]
	m.currentIndex++
	return &response, nil
}

func (m *ExtendedMockLLM) Stream(ctx context.Context, messages []llm.Message, options *llm.CallOptions) (<-chan llm.StreamResponse, error) {
	ch := make(chan llm.StreamResponse, 1)
	defer close(ch)

	response, err := m.Call(ctx, messages, options)
	if err != nil {
		ch <- llm.StreamResponse{Error: err}
		return ch, nil
	}

	ch <- llm.StreamResponse{Delta: response.Content}
	return ch, nil
}

func (m *ExtendedMockLLM) CallStream(ctx context.Context, messages []llm.Message, options *llm.CallOptions) (<-chan llm.StreamResponse, error) {
	return m.Stream(ctx, messages, options)
}

func (m *ExtendedMockLLM) GetProvider() string                  { return "mock" }
func (m *ExtendedMockLLM) GetModel() string                     { return "extended-mock" }
func (m *ExtendedMockLLM) GetAPIKey() string                    { return "mock-key" }
func (m *ExtendedMockLLM) GetBaseURL() string                   { return "https://mock.api" }
func (m *ExtendedMockLLM) GetTimeout() time.Duration            { return 30 * time.Second }
func (m *ExtendedMockLLM) GetMaxRetries() int                   { return 3 }
func (m *ExtendedMockLLM) GetHTTPClient() interface{}           { return nil }
func (m *ExtendedMockLLM) GetLogger() logger.Logger             { return logger.NewTestLogger() }
func (m *ExtendedMockLLM) GetEventBus() events.EventBus         { return nil }
func (m *ExtendedMockLLM) GetCustomHeaders() map[string]string  { return nil }
func (m *ExtendedMockLLM) SupportsFunctionCalling() bool        { return true }
func (m *ExtendedMockLLM) GetContextWindowSize() int            { return 4096 }
func (m *ExtendedMockLLM) SetEventBus(eventBus events.EventBus) {}
func (m *ExtendedMockLLM) Close() error                         { return nil }

// GetCallCount 返回LLM被调用的次数
func (m *ExtendedMockLLM) GetCallCount() int {
	return m.callCount
}

// ResetCallCount 重置调用计数
func (m *ExtendedMockLLM) ResetCallCount() {
	m.callCount = 0
}

// Reset 重置到初始状态
func (m *ExtendedMockLLM) Reset() {
	m.currentIndex = 0
	m.callCount = 0
}

// ===== Mock Tool =====

// MockTool 用于测试的模拟工具
type MockTool struct {
	name        string
	description string
	schema      ToolSchema
	executeFunc func(ctx context.Context, args map[string]interface{}) (interface{}, error)
	usageCount  int
	usageLimit  int
}

// NewMockTool 创建模拟工具
func NewMockTool(name, description string) *MockTool {
	return &MockTool{
		name:        name,
		description: description,
		schema: ToolSchema{
			Name:        name,
			Description: description,
			Parameters:  make(map[string]interface{}),
			Required:    make([]string, 0),
		},
		executeFunc: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return "mock result", nil
		},
		usageCount: 0,
		usageLimit: -1,
	}
}

// WithExecuteFunc 设置自定义执行函数
func (m *MockTool) WithExecuteFunc(fn func(ctx context.Context, args map[string]interface{}) (interface{}, error)) *MockTool {
	m.executeFunc = fn
	return m
}

// WithUsageLimit 设置使用限制
func (m *MockTool) WithUsageLimit(limit int) *MockTool {
	m.usageLimit = limit
	return m
}

func (m *MockTool) GetName() string {
	return m.name
}

func (m *MockTool) GetDescription() string {
	return m.description
}

func (m *MockTool) GetSchema() ToolSchema {
	return m.schema
}

func (m *MockTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	if m.usageLimit >= 0 && m.usageCount >= m.usageLimit {
		return nil, fmt.Errorf("tool usage limit exceeded: %d/%d", m.usageCount, m.usageLimit)
	}

	m.usageCount++
	return m.executeFunc(ctx, args)
}

func (m *MockTool) ExecuteAsync(ctx context.Context, args map[string]interface{}) (<-chan ToolResult, error) {
	resultChan := make(chan ToolResult, 1)
	go func() {
		defer close(resultChan)
		result, err := m.Execute(ctx, args)
		resultChan <- ToolResult{
			Output:   result,
			Error:    err,
			Duration: time.Millisecond * 10, // 模拟执行时间
		}
	}()
	return resultChan, nil
}

func (m *MockTool) GetUsageCount() int {
	return m.usageCount
}

func (m *MockTool) GetUsageLimit() int {
	return m.usageLimit
}

func (m *MockTool) ResetUsage() {
	m.usageCount = 0
}

func (m *MockTool) IsUsageLimitExceeded() bool {
	return m.usageLimit >= 0 && m.usageCount > m.usageLimit
}

// 确保MockTool实现了Tool接口
var _ Tool = (*MockTool)(nil)

// ===== Mock Agent =====

// MockAgent 用于测试的模拟Agent
type MockAgent struct {
	id        string
	role      string
	goal      string
	backstory string
	tools     []Tool
}

// NewMockAgent 创建模拟Agent
func NewMockAgent(role, goal, backstory string) *MockAgent {
	return &MockAgent{
		id:        "mock-agent-id",
		role:      role,
		goal:      goal,
		backstory: backstory,
		tools:     make([]Tool, 0),
	}
}

func (m *MockAgent) GetID() string        { return m.id }
func (m *MockAgent) GetRole() string      { return m.role }
func (m *MockAgent) GetGoal() string      { return m.goal }
func (m *MockAgent) GetBackstory() string { return m.backstory }
func (m *MockAgent) GetTools() []Tool     { return m.tools }

func (m *MockAgent) AddTool(tool Tool) error {
	m.tools = append(m.tools, tool)
	return nil
}

// 实现Agent接口的其他必需方法（简化版）
func (m *MockAgent) Execute(ctx context.Context, task Task) (*TaskOutput, error) {
	return &TaskOutput{
		Raw:   "mock result",
		Agent: m.role,
	}, nil
}

func (m *MockAgent) ExecuteAsync(ctx context.Context, task Task) (<-chan TaskResult, error) {
	resultChan := make(chan TaskResult, 1)
	go func() {
		defer close(resultChan)
		output, err := m.Execute(ctx, task)
		resultChan <- TaskResult{Output: output, Error: err}
	}()
	return resultChan, nil
}

// 其他必需的方法（空实现）
func (m *MockAgent) ExecuteWithTimeout(ctx context.Context, task Task, timeout time.Duration) (*TaskOutput, error) {
	return m.Execute(ctx, task)
}
func (m *MockAgent) SetLLM(llm llm.LLM) error                                            { return nil }
func (m *MockAgent) GetLLM() llm.LLM                                                     { return nil }
func (m *MockAgent) SetMemory(memory Memory) error                                       { return nil }
func (m *MockAgent) GetMemory() Memory                                                   { return nil }
func (m *MockAgent) SetKnowledgeSources(sources []KnowledgeSource) error                 { return nil }
func (m *MockAgent) GetKnowledgeSources() []KnowledgeSource                              { return nil }
func (m *MockAgent) SetExecutionConfig(config ExecutionConfig) error                     { return nil }
func (m *MockAgent) GetExecutionConfig() ExecutionConfig                                 { return ExecutionConfig{} }
func (m *MockAgent) SetHumanInputHandler(handler HumanInputHandler) error                { return nil }
func (m *MockAgent) GetHumanInputHandler() HumanInputHandler                             { return nil }
func (m *MockAgent) SetEventBus(eventBus events.EventBus) error                          { return nil }
func (m *MockAgent) GetEventBus() events.EventBus                                        { return nil }
func (m *MockAgent) SetLogger(logger logger.Logger) error                                { return nil }
func (m *MockAgent) GetLogger() logger.Logger                                            { return nil }
func (m *MockAgent) Start(ctx context.Context) error                                     { return nil }
func (m *MockAgent) Stop(ctx context.Context) error                                      { return nil }
func (m *MockAgent) IsRunning() bool                                                     { return true }
func (m *MockAgent) AddCallback(callback func(context.Context, *TaskOutput) error) error { return nil }
func (m *MockAgent) SetStepCallback(callback func(context.Context, *AgentStep) error) error {
	return nil
}
func (m *MockAgent) GetStepCallback() func(context.Context, *AgentStep) error { return nil }
func (m *MockAgent) Initialize() error                                        { return nil }
func (m *MockAgent) Close() error                                             { return nil }
func (m *MockAgent) Clone() Agent {
	clone := *m
	clone.tools = make([]Tool, len(m.tools))
	copy(clone.tools, m.tools)
	return &clone
}
func (m *MockAgent) GetExecutionStats() ExecutionStats            { return ExecutionStats{} }
func (m *MockAgent) ResetStats() error                            { return nil }
func (m *MockAgent) SetReasoningHandler(handler ReasoningHandler) {}
func (m *MockAgent) GetReasoningHandler() ReasoningHandler        { return nil }

// 确保MockAgent实现了Agent接口
var _ Agent = (*MockAgent)(nil)

// ===== Mock Reasoning Handler =====

// MockReasoningHandler 用于测试的模拟推理处理器
type MockReasoningHandler struct{}

func (m *MockReasoningHandler) HandleReasoning(ctx context.Context, task Task, agent Agent) (*ReasoningOutput, error) {
	return &ReasoningOutput{
		Plan: ReasoningPlan{
			Plan:  "Mock reasoning plan",
			Ready: true,
		},
		Success:    true,
		Duration:   time.Millisecond * 100,
		Iterations: 1,
		FinalReady: true,
		Metadata:   make(map[string]interface{}),
		CreatedAt:  time.Now(),
	}, nil
}

func (m *MockReasoningHandler) CreatePlan(ctx context.Context, task Task, agent Agent) (*ReasoningPlan, error) {
	return &ReasoningPlan{
		Plan:  "Mock plan",
		Ready: true,
	}, nil
}

func (m *MockReasoningHandler) RefinePlan(ctx context.Context, plan *ReasoningPlan, feedback string) (*ReasoningPlan, error) {
	return plan, nil
}

func (m *MockReasoningHandler) IsReady(plan *ReasoningPlan) bool {
	return plan != nil && plan.Ready
}

func (m *MockReasoningHandler) GetPlanSteps(plan *ReasoningPlan) []ReasoningStep {
	return []ReasoningStep{}
}

// ===== Test Helper Functions =====

// createTestAgent 创建用于测试的标准Agent配置
func createTestAgent(mockLLM llm.LLM) (*BaseAgent, error) {
	config := AgentConfig{
		Role:      "Test Agent",
		Goal:      "Test goal",
		Backstory: "Test backstory",
		LLM:       mockLLM,
		Logger:    logger.NewTestLogger(),
		EventBus:  events.NewEventBus(logger.NewTestLogger()),
	}
	return NewBaseAgent(config)
}

// createStandardMockResponse 创建标准的Mock响应
func createStandardMockResponse(content string) *llm.Response {
	return &llm.Response{
		Content:      content,
		Model:        "mock-model",
		FinishReason: "stop",
		Usage: llm.Usage{
			PromptTokens:     5,
			CompletionTokens: 5,
			TotalTokens:      10,
			Cost:             0.01,
		},
	}
}

// createErrorMockResponse 创建错误Mock响应  
func createErrorMockResponse(content string) *llm.Response {
	return &llm.Response{
		Content:      content,
		Model:        "mock-model",
		FinishReason: "error",
		Usage: llm.Usage{
			TotalTokens: 5,
		},
	}
}

// CreateTestLogger 创建测试用的日志记录器
func CreateTestLogger() logger.Logger {
	return logger.NewTestLogger()
}

// CreateTestEventBus 创建测试用的事件总线
func CreateTestEventBus() events.EventBus {
	return events.NewEventBus(CreateTestLogger())
}

// CreateTestAgentConfig 创建测试用的Agent配置
func CreateTestAgentConfig(role, goal, backstory string, mockLLM llm.LLM) AgentConfig {
	return AgentConfig{
		Role:      role,
		Goal:      goal,
		Backstory: backstory,
		LLM:       mockLLM,
		Logger:    CreateTestLogger(),
		EventBus:  CreateTestEventBus(),
	}
}

// CreateStandardMockResponse 创建标准的模拟响应
func CreateStandardMockResponse(content string) *llm.Response {
	return &llm.Response{
		Content:      content,
		Model:        "mock-model",
		FinishReason: "stop",
		Usage: llm.Usage{
			PromptTokens:     10,
			CompletionTokens: 10,
			TotalTokens:      20,
			Cost:             0.01,
		},
	}
}

// CreateErrorMockResponse 创建错误的模拟响应
func CreateErrorMockResponse(content string) *llm.Response {
	return &llm.Response{
		Content:      content,
		Model:        "mock-model",
		FinishReason: "error",
		Usage: llm.Usage{
			TotalTokens: 5,
		},
	}
}
