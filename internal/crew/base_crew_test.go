package crew

import (
	"context"
	"testing"
	"time"

	"github.com/ynl/greensoulai/internal/agent"
	"github.com/ynl/greensoulai/internal/llm"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// MockAgent 用于测试的Mock Agent
type MockAgent struct {
	id        string
	role      string
	goal      string
	backstory string
	llm       llm.LLM
}

func (m *MockAgent) Execute(ctx context.Context, task agent.Task) (*agent.TaskOutput, error) {
	return &agent.TaskOutput{
		Raw:            "Mock agent output for: " + task.GetDescription(),
		Agent:          m.role,
		Description:    task.GetDescription(),
		ExpectedOutput: task.GetExpectedOutput(),
		CreatedAt:      time.Now(),
		IsValid:        true,
		Metadata:       make(map[string]interface{}),
	}, nil
}

func (m *MockAgent) ExecuteAsync(ctx context.Context, task agent.Task) (<-chan agent.TaskResult, error) {
	resultChan := make(chan agent.TaskResult, 1)
	go func() {
		defer close(resultChan)
		output, err := m.Execute(ctx, task)
		resultChan <- agent.TaskResult{Output: output, Error: err}
	}()
	return resultChan, nil
}

func (m *MockAgent) ExecuteWithTimeout(ctx context.Context, task agent.Task, timeout time.Duration) (*agent.TaskOutput, error) {
	return m.Execute(ctx, task)
}

func (m *MockAgent) Initialize() error {
	return nil
}

func (m *MockAgent) GetID() string {
	return m.id
}

func (m *MockAgent) GetRole() string {
	return m.role
}

func (m *MockAgent) GetGoal() string {
	return m.goal
}

func (m *MockAgent) GetBackstory() string {
	return m.backstory
}

func (m *MockAgent) AddTool(tool agent.Tool) error {
	return nil
}

func (m *MockAgent) GetTools() []agent.Tool {
	return []agent.Tool{}
}

func (m *MockAgent) SetLLM(llmProvider llm.LLM) error {
	m.llm = llmProvider
	return nil
}

func (m *MockAgent) GetLLM() llm.LLM {
	return m.llm
}

func (m *MockAgent) SetMemory(memory agent.Memory) error {
	return nil
}

func (m *MockAgent) GetMemory() agent.Memory {
	return nil
}

func (m *MockAgent) SetKnowledgeSources(sources []agent.KnowledgeSource) error {
	return nil
}

func (m *MockAgent) GetKnowledgeSources() []agent.KnowledgeSource {
	return []agent.KnowledgeSource{}
}

func (m *MockAgent) SetHumanInputHandler(handler agent.HumanInputHandler) error {
	return nil
}

func (m *MockAgent) GetHumanInputHandler() agent.HumanInputHandler {
	return nil
}

func (m *MockAgent) SetExecutionConfig(config agent.ExecutionConfig) error {
	return nil
}

func (m *MockAgent) GetExecutionConfig() agent.ExecutionConfig {
	return agent.ExecutionConfig{}
}

func (m *MockAgent) GetExecutionStats() agent.ExecutionStats {
	return agent.ExecutionStats{}
}

func (m *MockAgent) ResetExecutionStats() {
}

func (m *MockAgent) ResetStats() error {
	return nil
}

func (m *MockAgent) Clone() agent.Agent {
	return &MockAgent{id: m.id, role: m.role, goal: m.goal, backstory: m.backstory}
}

func (m *MockAgent) GetEventBus() events.EventBus {
	return nil
}

func (m *MockAgent) SetEventBus(eventBus events.EventBus) error {
	return nil
}

func (m *MockAgent) GetLogger() logger.Logger {
	return nil
}

func (m *MockAgent) SetLogger(log logger.Logger) error {
	return nil
}

func (m *MockAgent) Close() error {
	return nil
}

// MockTask 用于测试的Mock Task
type MockTask struct {
	id             string
	description    string
	expectedOutput string
	humanInput     string
	tools          []agent.Tool
}

func (m *MockTask) GetID() string {
	return m.id
}

func (m *MockTask) GetDescription() string {
	return m.description
}

func (m *MockTask) GetExpectedOutput() string {
	return m.expectedOutput
}

func (m *MockTask) GetContext() map[string]interface{} {
	return make(map[string]interface{})
}

func (m *MockTask) IsHumanInputRequired() bool {
	return false
}

func (m *MockTask) SetHumanInput(input string) {
	m.humanInput = input
}

func (m *MockTask) GetHumanInput() string {
	return m.humanInput
}

func (m *MockTask) GetOutputFormat() agent.OutputFormat {
	return agent.OutputFormatRAW
}

func (m *MockTask) GetTools() []agent.Tool {
	return m.tools
}

func (m *MockTask) SetDescription(description string) {
	m.description = description
}

func (m *MockTask) AddTool(tool agent.Tool) error {
	m.tools = append(m.tools, tool)
	return nil
}

func (m *MockTask) SetTools(tools []agent.Tool) error {
	m.tools = tools
	return nil
}

func (m *MockTask) HasTools() bool {
	return len(m.tools) > 0
}

func (m *MockTask) Validate() error {
	return nil
}

// 实现新的Task接口方法，对标Python版本

func (m *MockTask) GetAssignedAgent() agent.Agent {
	return nil // Mock Task默认没有预分配Agent
}

func (m *MockTask) SetAssignedAgent(agent agent.Agent) error {
	return nil // Mock实现，不做实际存储
}

func (m *MockTask) IsAsyncExecution() bool {
	return false // Mock Task默认同步执行
}

func (m *MockTask) SetAsyncExecution(async bool) {
	// Mock实现，不做实际存储
}

func (m *MockTask) SetContext(context map[string]interface{}) {
	// Mock实现，不做实际存储
}

func (m *MockTask) GetName() string {
	return "mock-task"
}

func (m *MockTask) SetName(name string) {
	// Mock implementation
}

func (m *MockTask) GetOutputFile() string {
	return ""
}

func (m *MockTask) SetOutputFile(filename string) error {
	return nil
}

func (m *MockTask) GetCreateDirectory() bool {
	return false
}

func (m *MockTask) SetCreateDirectory(create bool) {
	// Mock implementation
}

func (m *MockTask) GetCallback() func(context.Context, *agent.TaskOutput) error {
	return nil
}

func (m *MockTask) SetCallback(callback func(context.Context, *agent.TaskOutput) error) {
	// Mock implementation
}

func (m *MockTask) GetContextTasks() []agent.Task {
	return nil
}

func (m *MockTask) SetContextTasks(tasks []agent.Task) {
	// Mock implementation
}

func (m *MockTask) GetRetryCount() int {
	return 0
}

func (m *MockTask) GetMaxRetries() int {
	return 3
}

func (m *MockTask) SetMaxRetries(maxRetries int) {
	// Mock implementation
}

func (m *MockTask) HasGuardrail() bool {
	return false
}

func (m *MockTask) SetGuardrail(guardrail agent.TaskGuardrail) {
	// Mock implementation
}

func (m *MockTask) GetGuardrail() agent.TaskGuardrail {
	return nil
}

func (m *MockTask) IsMarkdownOutput() bool {
	return false
}

func (m *MockTask) SetMarkdownOutput(markdown bool) {
	// Mock implementation
}

func TestNewBaseCrew(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := events.NewEventBus(logger)

	tests := []struct {
		name   string
		config *CrewConfig
	}{
		{
			name:   "with_nil_config",
			config: nil,
		},
		{
			name: "with_custom_config",
			config: &CrewConfig{
				Name:          "test_crew",
				Process:       ProcessSequential,
				Verbose:       true,
				MemoryEnabled: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crew := NewBaseCrew(tt.config, eventBus, logger)
			if crew == nil {
				t.Error("expected crew to be created")
			}

			if tt.config == nil {
				// 应该使用默认配置
				if crew.name != "crew" {
					t.Errorf("expected default name 'crew', got %s", crew.name)
				}
			} else {
				if crew.name != tt.config.Name {
					t.Errorf("expected name %s, got %s", tt.config.Name, crew.name)
				}
			}
		})
	}
}

func TestBaseCrew_AddAgent(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := events.NewEventBus(logger)
	crew := NewBaseCrew(nil, eventBus, logger)

	agent1 := &MockAgent{id: "agent1", role: "developer", goal: "write code", backstory: "experienced developer"}
	agent2 := &MockAgent{id: "agent2", role: "tester", goal: "test code", backstory: "quality assurance expert"}

	// 添加第一个agent
	err := crew.AddAgent(agent1)
	if err != nil {
		t.Errorf("failed to add agent: %v", err)
	}

	agents := crew.GetAgents()
	if len(agents) != 1 {
		t.Errorf("expected 1 agent, got %d", len(agents))
	}

	if agents[0].GetRole() != "developer" {
		t.Errorf("expected role 'developer', got %s", agents[0].GetRole())
	}

	// 添加第二个agent
	err = crew.AddAgent(agent2)
	if err != nil {
		t.Errorf("failed to add second agent: %v", err)
	}

	agents = crew.GetAgents()
	if len(agents) != 2 {
		t.Errorf("expected 2 agents, got %d", len(agents))
	}
}

func TestBaseCrew_AddTask(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := events.NewEventBus(logger)
	crew := NewBaseCrew(nil, eventBus, logger)

	task1 := &MockTask{id: "task1", description: "write unit tests", expectedOutput: "test files"}
	task2 := &MockTask{id: "task2", description: "review code", expectedOutput: "review comments"}

	// 添加第一个task
	err := crew.AddTask(task1)
	if err != nil {
		t.Errorf("failed to add task: %v", err)
	}

	tasks := crew.GetTasks()
	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}

	if tasks[0].GetDescription() != "write unit tests" {
		t.Errorf("expected description 'write unit tests', got %s", tasks[0].GetDescription())
	}

	// 添加第二个task
	err = crew.AddTask(task2)
	if err != nil {
		t.Errorf("failed to add second task: %v", err)
	}

	tasks = crew.GetTasks()
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestBaseCrew_Kickoff_Sequential(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := events.NewEventBus(logger)
	crew := NewBaseCrew(nil, eventBus, logger)

	// 添加agents和tasks
	agent1 := &MockAgent{id: "agent1", role: "developer", goal: "write code", backstory: "experienced developer"}
	task1 := &MockTask{id: "task1", description: "implement feature", expectedOutput: "working code"}

	crew.AddAgent(agent1)
	crew.AddTask(task1)
	crew.SetProcess(ProcessSequential)

	// 执行crew
	ctx := context.Background()
	inputs := map[string]interface{}{
		"feature": "user authentication",
	}

	result, err := crew.Kickoff(ctx, inputs)
	if err != nil {
		t.Errorf("crew execution failed: %v", err)
	}

	if result == nil {
		t.Error("expected result to be non-nil")
		return
	}

	if !result.Success {
		t.Error("expected execution to be successful")
	}

	if len(result.TasksOutput) != 1 {
		t.Errorf("expected 1 task output, got %d", len(result.TasksOutput))
	}

	if result.TasksOutput[0].Agent != "developer" {
		t.Errorf("expected agent 'developer', got %s", result.TasksOutput[0].Agent)
	}
}

func TestBaseCrew_KickoffAsync(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := events.NewEventBus(logger)
	crew := NewBaseCrew(nil, eventBus, logger)

	// 添加agents和tasks
	agent1 := &MockAgent{id: "agent1", role: "developer", goal: "write code", backstory: "experienced developer"}
	task1 := &MockTask{id: "task1", description: "implement feature", expectedOutput: "working code"}

	crew.AddAgent(agent1)
	crew.AddTask(task1)

	// 异步执行crew
	ctx := context.Background()
	inputs := map[string]interface{}{
		"feature": "user authentication",
	}

	resultChan, err := crew.KickoffAsync(ctx, inputs)
	if err != nil {
		t.Errorf("failed to start async execution: %v", err)
	}

	// 等待结果
	select {
	case result := <-resultChan:
		if result.Error != nil {
			t.Errorf("async execution failed: %v", result.Error)
		}
		if result.Output == nil {
			t.Error("expected output to be non-nil")
		}
		if !result.Output.Success {
			t.Error("expected execution to be successful")
		}
	case <-time.After(5 * time.Second):
		t.Error("async execution timed out")
	}
}

func TestBaseCrew_KickoffWithTimeout(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := events.NewEventBus(logger)
	crew := NewBaseCrew(nil, eventBus, logger)

	// 添加agents和tasks
	agent1 := &MockAgent{id: "agent1", role: "developer", goal: "write code", backstory: "experienced developer"}
	task1 := &MockTask{id: "task1", description: "implement feature", expectedOutput: "working code"}

	crew.AddAgent(agent1)
	crew.AddTask(task1)

	// 带超时执行crew
	ctx := context.Background()
	inputs := map[string]interface{}{
		"feature": "user authentication",
	}

	result, err := crew.KickoffWithTimeout(ctx, inputs, 10*time.Second)
	if err != nil {
		t.Errorf("crew execution with timeout failed: %v", err)
	}

	if result == nil {
		t.Error("expected result to be non-nil")
		return
	}

	if !result.Success {
		t.Error("expected execution to be successful")
	}
}

func TestBaseCrew_ValidationErrors(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := events.NewEventBus(logger)

	tests := []struct {
		name        string
		setupCrew   func() *BaseCrew
		expectError bool
	}{
		{
			name: "no_agents",
			setupCrew: func() *BaseCrew {
				crew := NewBaseCrew(nil, eventBus, logger)
				crew.AddTask(&MockTask{description: "test task", expectedOutput: "output"})
				return crew
			},
			expectError: true,
		},
		{
			name: "no_tasks",
			setupCrew: func() *BaseCrew {
				crew := NewBaseCrew(nil, eventBus, logger)
				crew.AddAgent(&MockAgent{role: "test agent", goal: "test goal"})
				return crew
			},
			expectError: true,
		},
		{
			name: "valid_configuration",
			setupCrew: func() *BaseCrew {
				crew := NewBaseCrew(nil, eventBus, logger)
				crew.AddAgent(&MockAgent{role: "test agent", goal: "test goal"})
				crew.AddTask(&MockTask{description: "test task", expectedOutput: "output"})
				return crew
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crew := tt.setupCrew()
			ctx := context.Background()
			inputs := map[string]interface{}{}

			_, err := crew.Kickoff(ctx, inputs)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

func TestBaseCrew_Clone(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := events.NewEventBus(logger)
	originalCrew := NewBaseCrew(nil, eventBus, logger)

	// 设置原始crew
	agent1 := &MockAgent{id: "agent1", role: "developer", goal: "write code", backstory: "experienced developer"}
	task1 := &MockTask{id: "task1", description: "implement feature", expectedOutput: "working code"}

	originalCrew.AddAgent(agent1)
	originalCrew.AddTask(task1)
	originalCrew.SetProcess(ProcessSequential)
	originalCrew.SetVerbose(true)

	// 克隆crew
	clonedCrew, err := originalCrew.Clone()
	if err != nil {
		t.Errorf("failed to clone crew: %v", err)
	}

	// 验证克隆的crew
	if clonedCrew.GetProcess() != ProcessSequential {
		t.Error("cloned crew should have same process")
	}

	clonedAgents := clonedCrew.GetAgents()
	if len(clonedAgents) != 1 {
		t.Errorf("expected 1 agent in cloned crew, got %d", len(clonedAgents))
	}

	clonedTasks := clonedCrew.GetTasks()
	if len(clonedTasks) != 1 {
		t.Errorf("expected 1 task in cloned crew, got %d", len(clonedTasks))
	}
}

func TestBaseCrew_Callbacks(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := events.NewEventBus(logger)
	crew := NewBaseCrew(nil, eventBus, logger)

	// 设置crew
	agent1 := &MockAgent{id: "agent1", role: "developer", goal: "write code", backstory: "experienced developer"}
	task1 := &MockTask{id: "task1", description: "implement feature", expectedOutput: "working code"}

	crew.AddAgent(agent1)
	crew.AddTask(task1)

	// 添加回调
	beforeCallbackCalled := false
	afterCallbackCalled := false

	beforeCallback := func(ctx context.Context, c Crew, output *CrewOutput) (*CrewOutput, error) {
		beforeCallbackCalled = true
		return output, nil
	}

	afterCallback := func(ctx context.Context, c Crew, output *CrewOutput) (*CrewOutput, error) {
		afterCallbackCalled = true
		return output, nil
	}

	crew.AddBeforeKickoffCallback(beforeCallback)
	crew.AddAfterKickoffCallback(afterCallback)

	// 执行crew
	ctx := context.Background()
	inputs := map[string]interface{}{}

	_, err := crew.Kickoff(ctx, inputs)
	if err != nil {
		t.Errorf("crew execution failed: %v", err)
	}

	if !beforeCallbackCalled {
		t.Error("before kickoff callback was not called")
	}

	if !afterCallbackCalled {
		t.Error("after kickoff callback was not called")
	}
}

func TestBaseCrew_UsageMetrics(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := events.NewEventBus(logger)
	crew := NewBaseCrew(nil, eventBus, logger)

	// 设置crew
	agent1 := &MockAgent{id: "agent1", role: "developer", goal: "write code", backstory: "experienced developer"}
	task1 := &MockTask{id: "task1", description: "implement feature", expectedOutput: "working code"}

	crew.AddAgent(agent1)
	crew.AddTask(task1)

	// 执行crew
	ctx := context.Background()
	inputs := map[string]interface{}{}

	result, err := crew.Kickoff(ctx, inputs)
	if err != nil {
		t.Errorf("crew execution failed: %v", err)
	}

	// 检查使用统计
	metrics := crew.GetUsageMetrics()
	if metrics == nil {
		t.Error("expected usage metrics to be available")
		return
	}

	if metrics.TotalTasks != 1 {
		t.Errorf("expected 1 total task, got %d", metrics.TotalTasks)
	}

	if metrics.SuccessfulTasks != 1 {
		t.Errorf("expected 1 successful task, got %d", metrics.SuccessfulTasks)
	}

	if metrics.FailedTasks != 0 {
		t.Errorf("expected 0 failed tasks, got %d", metrics.FailedTasks)
	}

	if result.TokenUsage == nil {
		t.Error("expected token usage in result")
	}
}
