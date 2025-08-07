package planning

import (
	"context"
	"fmt"
	"time"
)

// 这个文件包含用于测试的模拟依赖项，避免循环依赖

// MockAgent 模拟Agent接口，用于测试
type MockAgent struct {
	id        string
	role      string
	goal      string
	backstory string
}

// NewMockAgent 创建模拟Agent
func NewMockAgent(id, role, goal, backstory string) *MockAgent {
	return &MockAgent{
		id:        id,
		role:      role,
		goal:      goal,
		backstory: backstory,
	}
}

// ID 返回Agent ID
func (m *MockAgent) ID() string {
	return m.id
}

// Role 返回Agent角色
func (m *MockAgent) Role() string {
	return m.role
}

// Goal 返回Agent目标
func (m *MockAgent) Goal() string {
	return m.goal
}

// Backstory 返回Agent背景故事
func (m *MockAgent) Backstory() string {
	return m.backstory
}

// MockTask 模拟Task接口，用于测试
type MockTask struct {
	id             string
	description    string
	expectedOutput string
	agent          Agent
	outputJSON     bool
}

// NewMockTask 创建模拟Task
func NewMockTask(id, description, expectedOutput string, agent Agent) *MockTask {
	return &MockTask{
		id:             id,
		description:    description,
		expectedOutput: expectedOutput,
		agent:          agent,
	}
}

// ID 返回Task ID
func (m *MockTask) ID() string {
	return m.id
}

// Description 返回Task描述
func (m *MockTask) Description() string {
	return m.description
}

// ExpectedOutput 返回期望输出
func (m *MockTask) ExpectedOutput() string {
	return m.expectedOutput
}

// Agent 返回关联的Agent
func (m *MockTask) Agent() Agent {
	return m.agent
}

// ExecuteSync 同步执行任务（模拟实现）
func (m *MockTask) ExecuteSync(ctx context.Context) (*MockTaskOutput, error) {
	// 模拟执行延迟
	time.Sleep(10 * time.Millisecond)

	// 构造模拟的规划输出，确保符合质量验证标准
	mockOutput := &MockTaskOutput{
		Description: "Planning completed successfully",
		Raw: `{
			"list_of_plans_per_task": [
				{
					"task": "Test task 1",
					"plan": "Step 1: Initialize the task environment using the provided tools\nStep 2: Execute the main task workflow with proper error handling\nStep 3: Generate the expected output and validate results\nStep 4: Complete task documentation and cleanup"
				},
				{
					"task": "Test task 2", 
					"plan": "Step 1: Prepare the required tools and resources for task execution\nStep 2: Process the input data using advanced analytical tools\nStep 3: Generate comprehensive output according to specifications\nStep 4: Finalize results and ensure quality standards are met"
				}
			]
		}`,
		JSONDict: map[string]interface{}{
			"list_of_plans_per_task": []interface{}{
				map[string]interface{}{
					"task": "Test task 1",
					"plan": "Step 1: Initialize the task environment using the provided tools\nStep 2: Execute the main task workflow with proper error handling\nStep 3: Generate the expected output and validate results\nStep 4: Complete task documentation and cleanup",
				},
				map[string]interface{}{
					"task": "Test task 2",
					"plan": "Step 1: Prepare the required tools and resources for task execution\nStep 2: Process the input data using advanced analytical tools\nStep 3: Generate comprehensive output according to specifications\nStep 4: Finalize results and ensure quality standards are met",
				},
			},
		},
	}

	return mockOutput, nil
}

// SetOutputJSON 设置输出JSON格式
func (m *MockTask) SetOutputJSON(outputJSON bool) {
	m.outputJSON = outputJSON
}

// MockTaskOutput 模拟TaskOutput
type MockTaskOutput struct {
	Description string
	Raw         string
	JSONDict    map[string]interface{}
}

// MockAgentFactory 模拟Agent工厂
type MockAgentFactory struct{}

// NewMockAgentFactory 创建模拟Agent工厂
func NewMockAgentFactory() *MockAgentFactory {
	return &MockAgentFactory{}
}

// CreateAgent 创建Agent（模拟实现）
func (f *MockAgentFactory) CreateAgent(ctx context.Context, config *MockAgentConfig) (Agent, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// 模拟一些验证逻辑
	if config.Role == "" {
		return nil, fmt.Errorf("role cannot be empty")
	}

	if config.LLM == "" {
		return nil, fmt.Errorf("LLM cannot be empty")
	}

	// 创建模拟Agent
	agent := NewMockAgent(
		fmt.Sprintf("agent-%d", time.Now().UnixNano()),
		config.Role,
		config.Goal,
		config.Backstory,
	)

	return agent, nil
}

// MockAgentConfig 模拟Agent配置
type MockAgentConfig struct {
	Role        string
	Goal        string
	Backstory   string
	LLM         string
	Verbose     bool
	MaxIter     int
	Temperature float64
}

// MockTaskFactory 模拟Task工厂
type MockTaskFactory struct{}

// NewMockTaskFactory 创建模拟Task工厂
func NewMockTaskFactory() *MockTaskFactory {
	return &MockTaskFactory{}
}

// CreateTask 创建Task（模拟实现）
func (f *MockTaskFactory) CreateTask(ctx context.Context, config *MockTaskConfig) (Task, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.Description == "" {
		return nil, fmt.Errorf("description cannot be empty")
	}

	if config.ExpectedOutput == "" {
		return nil, fmt.Errorf("expected output cannot be empty")
	}

	if config.Agent == nil {
		return nil, fmt.Errorf("agent cannot be nil")
	}

	task := NewMockTask(
		fmt.Sprintf("task-%d", time.Now().UnixNano()),
		config.Description,
		config.ExpectedOutput,
		config.Agent,
	)

	task.SetOutputJSON(config.OutputJSON)

	return task, nil
}

// MockTaskConfig 模拟Task配置
type MockTaskConfig struct {
	Description    string
	ExpectedOutput string
	Agent          Agent
	OutputJSON     bool
	Verbose        bool
}

// 接口定义（用于类型安全）

// Agent 简化的Agent接口
type Agent interface {
	ID() string
	Role() string
	Goal() string
	Backstory() string
}

// Task 简化的Task接口
type Task interface {
	ID() string
	Description() string
	ExpectedOutput() string
	Agent() Agent
	ExecuteSync(ctx context.Context) (*MockTaskOutput, error)
	SetOutputJSON(outputJSON bool)
}

// AgentFactory 简化的Agent工厂接口
type AgentFactory interface {
	CreateAgent(ctx context.Context, config *MockAgentConfig) (Agent, error)
}

// TaskFactory 简化的Task工厂接口
type TaskFactory interface {
	CreateTask(ctx context.Context, config *MockTaskConfig) (Task, error)
}

// 将这些类型适配到我们的planning系统需要的接口格式

// Config 适配器，将MockAgentConfig转换为通用Config
type Config struct {
	Role        string
	Goal        string
	Backstory   string
	LLM         string
	Verbose     bool
	MaxIter     int
	Temperature float64
}

// TaskOutput 适配器
type TaskOutput struct {
	Description string
	Raw         string
	JSONDict    map[string]interface{}
}

// 适配器函数

// AdaptAgentConfig 将Config转换为MockAgentConfig
func AdaptAgentConfig(config *Config) *MockAgentConfig {
	return &MockAgentConfig{
		Role:        config.Role,
		Goal:        config.Goal,
		Backstory:   config.Backstory,
		LLM:         config.LLM,
		Verbose:     config.Verbose,
		MaxIter:     config.MaxIter,
		Temperature: config.Temperature,
	}
}

// AdaptTaskConfig 将Config转换为MockTaskConfig
func AdaptTaskConfig(description, expectedOutput string, agent Agent, outputJSON, verbose bool) *MockTaskConfig {
	return &MockTaskConfig{
		Description:    description,
		ExpectedOutput: expectedOutput,
		Agent:          agent,
		OutputJSON:     outputJSON,
		Verbose:        verbose,
	}
}

// AdaptTaskOutput 将MockTaskOutput转换为TaskOutput
func AdaptTaskOutput(mockOutput *MockTaskOutput) *TaskOutput {
	return &TaskOutput{
		Description: mockOutput.Description,
		Raw:         mockOutput.Raw,
		JSONDict:    mockOutput.JSONDict,
	}
}
