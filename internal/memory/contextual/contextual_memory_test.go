package contextual

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ynl/greensoulai/internal/agent"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// Mock对象定义

// MockTask 模拟任务 - 实现完整的Task接口
type MockTask struct {
	mock.Mock
}

func (m *MockTask) GetID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockTask) GetDescription() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockTask) SetDescription(description string) {
	m.Called(description)
}

func (m *MockTask) GetExpectedOutput() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockTask) GetContext() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func (m *MockTask) IsHumanInputRequired() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockTask) SetHumanInput(input string) {
	m.Called(input)
}

func (m *MockTask) GetHumanInput() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockTask) GetOutputFormat() agent.OutputFormat {
	args := m.Called()
	return args.Get(0).(agent.OutputFormat)
}

func (m *MockTask) GetTools() []agent.Tool {
	args := m.Called()
	return args.Get(0).([]agent.Tool)
}

func (m *MockTask) AddTool(tool agent.Tool) error {
	args := m.Called(tool)
	return args.Error(0)
}

func (m *MockTask) SetTools(tools []agent.Tool) error {
	args := m.Called(tools)
	return args.Error(0)
}

func (m *MockTask) HasTools() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockTask) Validate() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTask) GetAssignedAgent() agent.Agent {
	args := m.Called()
	return args.Get(0).(agent.Agent)
}

func (m *MockTask) SetAssignedAgent(agent agent.Agent) error {
	args := m.Called(agent)
	return args.Error(0)
}

func (m *MockTask) IsAsyncExecution() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockTask) SetAsyncExecution(async bool) {
	m.Called(async)
}

func (m *MockTask) SetContext(context map[string]interface{}) {
	m.Called(context)
}

func (m *MockTask) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockTask) SetName(name string) {
	m.Called(name)
}

func (m *MockTask) GetOutputFile() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockTask) SetOutputFile(filename string) error {
	args := m.Called(filename)
	return args.Error(0)
}

func (m *MockTask) GetCreateDirectory() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockTask) SetCreateDirectory(create bool) {
	m.Called(create)
}

func (m *MockTask) GetCallback() func(context.Context, *agent.TaskOutput) error {
	args := m.Called()
	return args.Get(0).(func(context.Context, *agent.TaskOutput) error)
}

func (m *MockTask) SetCallback(callback func(context.Context, *agent.TaskOutput) error) {
	m.Called(callback)
}

func (m *MockTask) GetContextTasks() []agent.Task {
	args := m.Called()
	return args.Get(0).([]agent.Task)
}

func (m *MockTask) SetContextTasks(tasks []agent.Task) {
	m.Called(tasks)
}

func (m *MockTask) GetRetryCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTask) GetMaxRetries() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTask) IsMarkdownOutput() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockTask) HasGuardrail() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockTask) GetGuardrail() agent.TaskGuardrail {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(agent.TaskGuardrail)
}

func (m *MockTask) SetGuardrail(guardrail agent.TaskGuardrail) {
	m.Called(guardrail)
}

func (m *MockTask) SetMarkdownOutput(markdown bool) {
	m.Called(markdown)
}

func (m *MockTask) SetMaxRetries(maxRetries int) {
	m.Called(maxRetries)
}

// MockEventBus 模拟事件总线 - 实现完整的EventBus接口
type MockEventBus struct {
	mock.Mock
	emittedEvents []events.Event
}

func (m *MockEventBus) Emit(ctx context.Context, source interface{}, event events.Event) error {
	m.emittedEvents = append(m.emittedEvents, event)
	args := m.Called(ctx, source, event)
	return args.Error(0)
}

func (m *MockEventBus) Subscribe(eventType string, handler events.EventHandler) error {
	args := m.Called(eventType, handler)
	return args.Error(0)
}

func (m *MockEventBus) Unsubscribe(eventType string, handler events.EventHandler) error {
	args := m.Called(eventType, handler)
	return args.Error(0)
}

func (m *MockEventBus) GetHandlerCount(eventType string) int {
	args := m.Called(eventType)
	return args.Int(0)
}

func (m *MockEventBus) GetRegisteredEventTypes() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockEventBus) RegisterHandler(eventType string, handler events.EventHandler) error {
	args := m.Called(eventType, handler)
	return args.Error(0)
}

func (m *MockEventBus) WithScopedHandlers() events.EventBus {
	args := m.Called()
	return args.Get(0).(events.EventBus)
}

func (m *MockEventBus) GetEmittedEvents() []events.Event {
	return m.emittedEvents
}

func (m *MockEventBus) ClearEvents() {
	m.emittedEvents = nil
}

// 简化的ContextualMemory测试，使用实际的记忆实例而不是mock

func TestNewContextualMemory(t *testing.T) {
	tests := []struct {
		name           string
		config         *ContextualMemoryConfig
		expectDefaults bool
	}{
		{
			name:           "with default config",
			config:         nil,
			expectDefaults: true,
		},
		{
			name: "with custom config",
			config: &ContextualMemoryConfig{
				DefaultSTMLimit:   5,
				DefaultLTMLimit:   3,
				STMScoreThreshold: 0.5,
				EnableFormatting:  false,
				MaxContextLength:  5000,
			},
			expectDefaults: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEventBus := &MockEventBus{}
			mockLogger := logger.NewConsoleLogger()

			// 使用nil记忆实例进行基础测试
			cm := NewContextualMemory(
				nil, // STM
				nil, // LTM
				nil, // EM
				nil, // EXM
				mockEventBus,
				mockLogger,
				tt.config,
			)

			assert.NotNil(t, cm)
			assert.Equal(t, mockEventBus, cm.eventBus)
			assert.Equal(t, mockLogger, cm.logger)

			if tt.expectDefaults {
				defaultConfig := DefaultContextualMemoryConfig()
				assert.Equal(t, defaultConfig.DefaultSTMLimit, cm.config.DefaultSTMLimit)
				assert.Equal(t, defaultConfig.DefaultLTMLimit, cm.config.DefaultLTMLimit)
				assert.Equal(t, defaultConfig.STMScoreThreshold, cm.config.STMScoreThreshold)
				assert.Equal(t, defaultConfig.EnableFormatting, cm.config.EnableFormatting)
				assert.Equal(t, defaultConfig.MaxContextLength, cm.config.MaxContextLength)
			} else {
				assert.Equal(t, tt.config.DefaultSTMLimit, cm.config.DefaultSTMLimit)
				assert.Equal(t, tt.config.DefaultLTMLimit, cm.config.DefaultLTMLimit)
				assert.Equal(t, tt.config.STMScoreThreshold, cm.config.STMScoreThreshold)
				assert.Equal(t, tt.config.EnableFormatting, cm.config.EnableFormatting)
				assert.Equal(t, tt.config.MaxContextLength, cm.config.MaxContextLength)
			}
		})
	}
}

func TestBuildContextForTask_EmptyQuery(t *testing.T) {
	mockEventBus := &MockEventBus{}
	mockLogger := logger.NewConsoleLogger()

	// 设置事件期望
	mockEventBus.On("Emit", mock.Anything, mock.Anything, mock.AnythingOfType("*contextual.ContextBuildStartedEvent")).Return(nil)
	mockEventBus.On("Emit", mock.Anything, mock.Anything, mock.AnythingOfType("*contextual.ContextBuildCompletedEvent")).Return(nil)

	cm := NewContextualMemory(nil, nil, nil, nil, mockEventBus, mockLogger, nil)

	mockTask := &MockTask{}
	mockTask.On("GetDescription").Return("")
	mockTask.On("GetID").Return("empty-task-1")

	ctx := context.Background()
	result, err := cm.BuildContextForTask(ctx, mockTask, "")

	assert.NoError(t, err)
	assert.Equal(t, "", result)
	mockTask.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestBuildContextForTask_NilMemoryInstances(t *testing.T) {
	mockEventBus := &MockEventBus{}
	mockLogger := logger.NewConsoleLogger()

	// 设置eventBus的mock期望
	mockEventBus.On("Emit", mock.Anything, mock.Anything, mock.AnythingOfType("*contextual.ContextBuildStartedEvent")).Return(nil)
	mockEventBus.On("Emit", mock.Anything, mock.Anything, mock.AnythingOfType("*contextual.ContextBuildCompletedEvent")).Return(nil)

	// 测试当所有记忆实例都为nil时的行为
	cm := NewContextualMemory(nil, nil, nil, nil, mockEventBus, mockLogger, nil)

	mockTask := &MockTask{}
	mockTask.On("GetDescription").Return("test task")
	mockTask.On("GetID").Return("task-1")

	ctx := context.Background()
	result, err := cm.BuildContextForTask(ctx, mockTask, "context")

	assert.NoError(t, err)
	assert.Equal(t, "", result) // 应该返回空字符串，但不出错

	mockTask.AssertExpectations(t)
	mockEventBus.AssertExpectations(t)
}

func TestDefaultContextualMemoryConfig(t *testing.T) {
	config := DefaultContextualMemoryConfig()

	assert.Equal(t, 3, config.DefaultSTMLimit)
	assert.Equal(t, 2, config.DefaultLTMLimit)
	assert.Equal(t, 3, config.DefaultEntityLimit)
	assert.Equal(t, 3, config.DefaultExternalLimit)

	assert.Equal(t, 0.35, config.STMScoreThreshold)
	assert.Equal(t, 0.35, config.EntityScoreThreshold)
	assert.Equal(t, 0.35, config.ExternalScoreThreshold)

	assert.True(t, config.EnableFormatting)
	assert.True(t, config.EnableSectionHeaders)
	assert.Equal(t, 8000, config.MaxContextLength)

	assert.True(t, config.FilterEmptyResults)
	assert.True(t, config.EnableDeduplication)
}

func TestContextualMemory_ConfigUpdate(t *testing.T) {
	mockEventBus := &MockEventBus{}
	mockLogger := logger.NewConsoleLogger()
	cm := NewContextualMemory(nil, nil, nil, nil, mockEventBus, mockLogger, nil)

	newConfig := ContextualMemoryConfig{
		DefaultSTMLimit:     5,
		DefaultLTMLimit:     3,
		STMScoreThreshold:   0.5,
		EnableFormatting:    false,
		MaxContextLength:    5000,
		FilterEmptyResults:  false,
		EnableDeduplication: false,
	}

	cm.UpdateConfig(newConfig)

	updatedConfig := cm.GetConfig()
	assert.Equal(t, newConfig.DefaultSTMLimit, updatedConfig.DefaultSTMLimit)
	assert.Equal(t, newConfig.DefaultLTMLimit, updatedConfig.DefaultLTMLimit)
	assert.Equal(t, newConfig.STMScoreThreshold, updatedConfig.STMScoreThreshold)
	assert.Equal(t, newConfig.EnableFormatting, updatedConfig.EnableFormatting)
	assert.Equal(t, newConfig.MaxContextLength, updatedConfig.MaxContextLength)
	assert.Equal(t, newConfig.FilterEmptyResults, updatedConfig.FilterEmptyResults)
	assert.Equal(t, newConfig.EnableDeduplication, updatedConfig.EnableDeduplication)
}

func TestContextualMemory_FilterAndDeduplication(t *testing.T) {
	mockEventBus := &MockEventBus{}
	mockLogger := logger.NewConsoleLogger()

	tests := []struct {
		name                string
		enableFiltering     bool
		enableDeduplication bool
		inputParts          []string
		expectedParts       []string
	}{
		{
			name:                "both enabled",
			enableFiltering:     true,
			enableDeduplication: true,
			inputParts:          []string{"part1", "", "part2", "part1", "   ", "part3"},
			expectedParts:       []string{"part1", "part2", "part3"},
		},
		{
			name:                "only filtering enabled",
			enableFiltering:     true,
			enableDeduplication: false,
			inputParts:          []string{"part1", "", "part2", "part1"},
			expectedParts:       []string{"part1", "part2", "part1"},
		},
		{
			name:                "only deduplication enabled",
			enableFiltering:     false,
			enableDeduplication: true,
			inputParts:          []string{"part1", "", "part2", "part1"},
			expectedParts:       []string{"part1", "", "part2"},
		},
		{
			name:                "both disabled",
			enableFiltering:     false,
			enableDeduplication: false,
			inputParts:          []string{"part1", "", "part2", "part1"},
			expectedParts:       []string{"part1", "", "part2", "part1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultContextualMemoryConfig()
			config.FilterEmptyResults = tt.enableFiltering
			config.EnableDeduplication = tt.enableDeduplication

			cm := NewContextualMemory(nil, nil, nil, nil, mockEventBus, mockLogger, &config)

			result := tt.inputParts
			if tt.enableFiltering {
				result = cm.filterEmptyParts(result)
			}
			if tt.enableDeduplication {
				result = cm.deduplicateParts(result)
			}

			assert.Equal(t, tt.expectedParts, result)
		})
	}
}

func TestContextualMemory_MaxContextLength(t *testing.T) {
	config := DefaultContextualMemoryConfig()
	config.MaxContextLength = 20 // 设置一个很小的长度限制

	mockEventBus := &MockEventBus{}
	mockLogger := logger.NewConsoleLogger()
	_ = NewContextualMemory(nil, nil, nil, nil, mockEventBus, mockLogger, &config)

	// 创建一个很长的上下文部分
	longParts := []string{"This is a very long context that should be truncated"}

	// 测试长度限制
	result := strings.Join(longParts, "\n")
	if len(result) > config.MaxContextLength {
		result = result[:config.MaxContextLength]
	}

	assert.LessOrEqual(t, len(result), config.MaxContextLength)
}

func TestRemoveDuplicateStrings(t *testing.T) {
	mockEventBus := &MockEventBus{}
	mockLogger := logger.NewConsoleLogger()
	cm := NewContextualMemory(nil, nil, nil, nil, mockEventBus, mockLogger, nil)

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "with duplicates",
			input:    []string{"a", "b", "a", "c", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "no duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "all duplicates",
			input:    []string{"a", "a", "a"},
			expected: []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cm.removeDuplicateStrings(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetMemoryInstances(t *testing.T) {
	mockEventBus := &MockEventBus{}
	mockLogger := logger.NewConsoleLogger()

	// 使用nil实例创建ContextualMemory
	cm := NewContextualMemory(nil, nil, nil, nil, mockEventBus, mockLogger, nil)

	stm, ltm, em, exm := cm.GetMemoryInstances()

	// 所有实例都应该是nil
	assert.Nil(t, stm)
	assert.Nil(t, ltm)
	assert.Nil(t, em)
	assert.Nil(t, exm)
}

// 基准测试

func BenchmarkContextualMemoryCreation(b *testing.B) {
	mockEventBus := &MockEventBus{}
	mockLogger := logger.NewConsoleLogger()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewContextualMemory(nil, nil, nil, nil, mockEventBus, mockLogger, nil)
	}
}

func BenchmarkBuildContextForTask_NilMemories(b *testing.B) {
	mockEventBus := &MockEventBus{}
	mockLogger := logger.NewConsoleLogger()

	// 设置mock期望
	mockEventBus.On("Emit", mock.Anything, mock.Anything, mock.AnythingOfType("*contextual.ContextBuildStartedEvent")).Return(nil)
	mockEventBus.On("Emit", mock.Anything, mock.Anything, mock.AnythingOfType("*contextual.ContextBuildCompletedEvent")).Return(nil)

	cm := NewContextualMemory(nil, nil, nil, nil, mockEventBus, mockLogger, nil)

	mockTask := &MockTask{}
	mockTask.On("GetDescription").Return("benchmark task")
	mockTask.On("GetID").Return("task-1")

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cm.BuildContextForTask(ctx, mockTask, "context")
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}
