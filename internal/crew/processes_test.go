package crew

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ynl/greensoulai/internal/agent"
	"github.com/ynl/greensoulai/internal/llm"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// MockLLM 用于测试的Mock LLM
type MockLLM struct {
	responses []string
	callCount int
}

func NewMockLLM(responses ...string) *MockLLM {
	return &MockLLM{
		responses: responses,
		callCount: 0,
	}
}

func (m *MockLLM) Call(ctx context.Context, messages []llm.Message, options *llm.CallOptions) (*llm.Response, error) {
	if m.callCount >= len(m.responses) {
		return &llm.Response{
			Content:      "Default mock response",
			Model:        "mock-model",
			FinishReason: "stop",
			Usage: llm.Usage{
				PromptTokens:     5,
				CompletionTokens: 5,
				TotalTokens:      10,
				Cost:             0.001,
			},
		}, nil
	}

	response := m.responses[m.callCount]
	m.callCount++

	return &llm.Response{
		Content:      response,
		Model:        "mock-model",
		FinishReason: "stop",
		Usage: llm.Usage{
			PromptTokens:     5,
			CompletionTokens: len(response),
			TotalTokens:      5 + len(response),
			Cost:             0.001,
		},
	}, nil
}

func (m *MockLLM) GetModel() string {
	return "mock-model"
}

func (m *MockLLM) SupportsFunctionCalling() bool {
	return false
}

func (m *MockLLM) GetContextWindowSize() int {
	return 4096
}

func (m *MockLLM) CallStream(ctx context.Context, messages []llm.Message, options *llm.CallOptions) (<-chan llm.StreamResponse, error) {
	responseChan := make(chan llm.StreamResponse, 1)

	go func() {
		defer close(responseChan)

		response, err := m.Call(ctx, messages, options)
		if err != nil {
			responseChan <- llm.StreamResponse{Error: err}
			return
		}

		responseChan <- llm.StreamResponse{
			Delta:        response.Content,
			Usage:        &response.Usage,
			FinishReason: response.FinishReason,
		}
	}()

	return responseChan, nil
}

func (m *MockLLM) SetEventBus(eventBus events.EventBus) {
	// Mock implementation - no-op
}

func (m *MockLLM) Close() error {
	return nil
}

// ContextAwareTask 可以记录和返回上下文的Task
type ContextAwareTask struct {
	*MockTask
	receivedContext map[string]interface{}
}

func NewContextAwareTask(id, description, expectedOutput string) *ContextAwareTask {
	return &ContextAwareTask{
		MockTask: &MockTask{
			id:             id,
			description:    description,
			expectedOutput: expectedOutput,
		},
		receivedContext: make(map[string]interface{}),
	}
}

func (c *ContextAwareTask) GetContext() map[string]interface{} {
	return c.receivedContext
}

// 模拟BaseTask的SetContext方法
func (c *ContextAwareTask) SetContext(context map[string]interface{}) {
	c.receivedContext = context
}

func TestSequentialProcess(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := events.NewEventBus(logger)

	tests := []struct {
		name          string
		agentsCount   int
		tasksCount    int
		expectSuccess bool
		expectError   bool
	}{
		{
			name:          "single_agent_single_task",
			agentsCount:   1,
			tasksCount:    1,
			expectSuccess: true,
			expectError:   false,
		},
		{
			name:          "single_agent_multiple_tasks",
			agentsCount:   1,
			tasksCount:    3,
			expectSuccess: true,
			expectError:   false,
		},
		{
			name:          "multiple_agents_multiple_tasks",
			agentsCount:   2,
			tasksCount:    3,
			expectSuccess: true,
			expectError:   false,
		},
		{
			name:          "more_agents_than_tasks",
			agentsCount:   3,
			tasksCount:    2,
			expectSuccess: true,
			expectError:   false,
		},
		{
			name:          "no_agents",
			agentsCount:   0,
			tasksCount:    1,
			expectSuccess: false,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crew := NewBaseCrew(nil, eventBus, logger)
			crew.SetProcess(ProcessSequential)

			// 添加agents
			for i := 0; i < tt.agentsCount; i++ {
				agent := &MockAgent{
					id:        string(rune('A' + i)),
					role:      "Agent" + string(rune('A'+i)),
					goal:      "Complete tasks efficiently",
					backstory: "Experienced AI assistant",
				}
				crew.AddAgent(agent)
			}

			// 添加tasks
			for i := 0; i < tt.tasksCount; i++ {
				task := NewContextAwareTask(
					string(rune('1'+i)),
					"Task "+string(rune('1'+i)),
					"Expected output "+string(rune('1'+i)),
				)
				crew.AddTask(task)
			}

			// 执行crew
			ctx := context.Background()
			inputs := map[string]interface{}{
				"initial_input": "test data",
			}

			result, err := crew.Kickoff(ctx, inputs)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("expected result to be non-nil")
				return
			}

			if tt.expectSuccess && !result.Success {
				t.Error("expected successful execution")
			}

			if tt.expectSuccess {
				expectedTasks := tt.tasksCount
				if len(result.TasksOutput) != expectedTasks {
					t.Errorf("expected %d task outputs, got %d", expectedTasks, len(result.TasksOutput))
				}

				// 验证任务按顺序执行
				for i, taskOutput := range result.TasksOutput {
					expectedAgent := "Agent" + string(rune('A'+(i%tt.agentsCount)))
					if taskOutput.Agent != expectedAgent {
						t.Errorf("task %d: expected agent %s, got %s", i, expectedAgent, taskOutput.Agent)
					}
				}
			}
		})
	}
}

func TestSequentialContextPassing(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := events.NewEventBus(logger)
	crew := NewBaseCrew(nil, eventBus, logger)
	crew.SetProcess(ProcessSequential)

	// 添加一个agent
	agent := &MockAgent{
		id:        "A",
		role:      "TestAgent",
		goal:      "Process tasks",
		backstory: "Test agent for context verification",
	}
	crew.AddAgent(agent)

	// 添加多个上下文感知的tasks
	task1 := NewContextAwareTask("1", "First task", "First output")
	task2 := NewContextAwareTask("2", "Second task", "Second output")
	task3 := NewContextAwareTask("3", "Third task", "Third output")

	crew.AddTask(task1)
	crew.AddTask(task2)
	crew.AddTask(task3)

	// 执行crew
	ctx := context.Background()
	inputs := map[string]interface{}{
		"initial_data": "start_value",
		"workflow_id":  "test_workflow",
	}

	result, err := crew.Kickoff(ctx, inputs)
	if err != nil {
		t.Fatalf("crew execution failed: %v", err)
	}

	if !result.Success {
		t.Fatal("expected successful execution")
	}

	// 验证第一个任务接收到初始输入
	task1Context := task1.GetContext()
	if task1Context["initial_data"] != "start_value" {
		t.Error("first task should receive initial inputs")
	}

	// 验证第二个任务接收到第一个任务的输出上下文
	task2Context := task2.GetContext()
	if task2Context["crew_name"] == "" {
		t.Error("second task should receive crew context")
	}
	if task2Context["completed_tasks"] != 1 {
		t.Error("second task should know about completed tasks")
	}

	// 验证第三个任务接收到之前所有任务的上下文
	task3Context := task3.GetContext()
	if task3Context["completed_tasks"] != 2 {
		t.Error("third task should know about 2 completed tasks")
	}
	if task3Context["last_task_output"] == nil {
		t.Error("third task should receive last task output")
	}
}

func TestHierarchicalProcess(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := events.NewEventBus(logger)

	tests := []struct {
		name              string
		setupManagerAgent bool
		setupManagerLLM   bool
		expectSuccess     bool
		expectError       bool
	}{
		{
			name:              "with_manager_agent",
			setupManagerAgent: true,
			setupManagerLLM:   false,
			expectSuccess:     true,
			expectError:       false,
		},
		{
			name:              "with_manager_llm",
			setupManagerAgent: false,
			setupManagerLLM:   true,
			expectSuccess:     true,
			expectError:       false,
		},
		{
			name:              "no_manager",
			setupManagerAgent: false,
			setupManagerLLM:   false,
			expectSuccess:     false,
			expectError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultCrewConfig()
			config.Process = ProcessHierarchical
			crew := NewBaseCrew(config, eventBus, logger)

			// 添加工作agents
			agent1 := &MockAgent{
				id:        "worker1",
				role:      "Developer",
				goal:      "Write code",
				backstory: "Experienced developer",
			}
			agent2 := &MockAgent{
				id:        "worker2",
				role:      "Tester",
				goal:      "Test code",
				backstory: "Quality assurance expert",
			}
			crew.AddAgent(agent1)
			crew.AddAgent(agent2)

			// 添加任务
			task1 := &MockTask{
				id:             "t1",
				description:    "Implement feature",
				expectedOutput: "Working code",
			}
			task2 := &MockTask{
				id:             "t2",
				description:    "Test feature",
				expectedOutput: "Test results",
			}
			crew.AddTask(task1)
			crew.AddTask(task2)

			// 设置管理器
			if tt.setupManagerAgent {
				mockLLM := NewMockLLM("Manager coordination response")
				managerAgent := &MockAgent{
					id:        "manager",
					role:      "Project Manager",
					goal:      "Coordinate team",
					backstory: "Experienced manager",
					llm:       mockLLM, // 设置LLM
				}
				crew.managerAgent = managerAgent
			}

			if tt.setupManagerLLM {
				mockLLM := NewMockLLM(
					"I'll coordinate the team to complete these tasks efficiently.",
					"Task completed successfully.",
				)
				crew.managerLLM = mockLLM
			}

			// 执行crew
			ctx := context.Background()
			inputs := map[string]interface{}{
				"feature_spec": "user authentication",
			}

			result, err := crew.Kickoff(ctx, inputs)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("expected result to be non-nil")
				return
			}

			if tt.expectSuccess && !result.Success {
				t.Error("expected successful execution")
			}

			if tt.expectSuccess {
				if len(result.TasksOutput) != 2 {
					t.Errorf("expected 2 task outputs, got %d", len(result.TasksOutput))
				}

				// 在Hierarchical模式中，所有任务都应该由管理器执行
				for i, taskOutput := range result.TasksOutput {
					expectedAgent := "Project Manager"
					if tt.setupManagerLLM {
						expectedAgent = "Crew Manager" // 默认管理器名称
					}
					if taskOutput.Agent != expectedAgent {
						t.Errorf("task %d: expected manager agent, got %s", i, taskOutput.Agent)
					}
				}
			}
		})
	}
}

func TestHierarchicalManagerCreation(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := events.NewEventBus(logger)

	t.Run("create_default_manager_with_llm", func(t *testing.T) {
		config := DefaultCrewConfig()
		config.Process = ProcessHierarchical
		crew := NewBaseCrew(config, eventBus, logger)

		// 添加工作agents
		worker := &MockAgent{id: "worker", role: "Worker", goal: "Work", backstory: "Worker"}
		crew.AddAgent(worker)

		// 添加任务
		task := &MockTask{id: "t1", description: "Work task", expectedOutput: "Work result"}
		crew.AddTask(task)

		// 设置管理器LLM
		mockLLM := NewMockLLM("Manager response")
		crew.managerLLM = mockLLM

		// 执行创建管理器Agent的逻辑
		err := crew.createManagerAgent()
		if err != nil {
			t.Errorf("failed to create manager agent: %v", err)
		}

		if crew.managerAgent == nil {
			t.Error("expected manager agent to be created")
		}

		if crew.managerAgent.GetRole() != "Crew Manager" {
			t.Errorf("expected manager role 'Crew Manager', got %s", crew.managerAgent.GetRole())
		}
	})

	t.Run("validate_existing_manager", func(t *testing.T) {
		config := DefaultCrewConfig()
		config.Process = ProcessHierarchical
		crew := NewBaseCrew(config, eventBus, logger)

		// 创建管理器Agent
		mockLLM := NewMockLLM("Manager response")
		managerAgent := &MockAgent{
			id:        "manager",
			role:      "Custom Manager",
			goal:      "Manage team",
			backstory: "Experienced manager",
			llm:       mockLLM, // 直接设置LLM
		}
		crew.managerAgent = managerAgent

		// 验证管理器Agent
		err := crew.createManagerAgent()
		if err != nil {
			t.Errorf("failed to validate manager agent: %v", err)
		}

		// 验证管理器配置是否正确
		if crew.managerAgent.GetRole() != "Custom Manager" {
			t.Error("manager agent role should be preserved")
		}
	})
}

func TestAgentSelection(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := events.NewEventBus(logger)

	t.Run("sequential_agent_assignment", func(t *testing.T) {
		crew := NewBaseCrew(nil, eventBus, logger)
		crew.SetProcess(ProcessSequential)

		// 添加多个agents
		agents := []*MockAgent{
			{id: "1", role: "Agent1", goal: "Goal1", backstory: "Backstory1"},
			{id: "2", role: "Agent2", goal: "Goal2", backstory: "Backstory2"},
			{id: "3", role: "Agent3", goal: "Goal3", backstory: "Backstory3"},
		}
		for _, agent := range agents {
			crew.AddAgent(agent)
		}

		// 测试多个任务的agent分配
		testCases := []struct {
			taskIndex     int
			expectedAgent string
		}{
			{0, "Agent1"}, // 第一个任务分配给第一个agent
			{1, "Agent2"}, // 第二个任务分配给第二个agent
			{2, "Agent3"}, // 第三个任务分配给第三个agent
			{3, "Agent1"}, // 第四个任务循环分配给第一个agent
			{4, "Agent2"}, // 第五个任务循环分配给第二个agent
		}

		for _, tc := range testCases {
			task := &MockTask{
				id:             string(rune('A' + tc.taskIndex)),
				description:    "Test task",
				expectedOutput: "Test output",
			}

			selectedAgent, err := crew.selectAgentForTask(task, tc.taskIndex)
			if err != nil {
				t.Errorf("task index %d: unexpected error: %v", tc.taskIndex, err)
				continue
			}

			if selectedAgent.GetRole() != tc.expectedAgent {
				t.Errorf("task index %d: expected agent %s, got %s",
					tc.taskIndex, tc.expectedAgent, selectedAgent.GetRole())
			}
		}
	})

	t.Run("hierarchical_manager_assignment", func(t *testing.T) {
		crew := NewBaseCrew(nil, eventBus, logger)
		crew.SetProcess(ProcessHierarchical)

		// 添加工作agents
		worker := &MockAgent{id: "worker", role: "Worker", goal: "Work", backstory: "Worker"}
		crew.AddAgent(worker)

		// 添加管理器agent
		manager := &MockAgent{id: "manager", role: "Manager", goal: "Manage", backstory: "Manager"}
		crew.managerAgent = manager

		// 测试任务分配给管理器
		task := &MockTask{id: "t1", description: "Test task", expectedOutput: "Test output"}

		selectedAgent, err := crew.selectAgentForTask(task, 0)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if selectedAgent.GetRole() != "Manager" {
			t.Errorf("expected Manager, got %s", selectedAgent.GetRole())
		}
	})
}

func TestTaskContextPassing(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := events.NewEventBus(logger)
	crew := NewBaseCrew(nil, eventBus, logger)

	// 创建模拟任务输出
	taskOutputs := []*agent.TaskOutput{
		{
			Raw:         "First task output",
			Agent:       "Agent1",
			Description: "First task",
			CreatedAt:   time.Now().Add(-2 * time.Minute),
		},
		{
			Raw:         "Second task output",
			Agent:       "Agent2",
			Description: "Second task",
			CreatedAt:   time.Now().Add(-1 * time.Minute),
		},
	}

	lastOutput := &agent.TaskOutput{
		Raw:         "Latest task output",
		Agent:       "Agent3",
		Description: "Latest task",
		JSON:        map[string]interface{}{"result": "success"},
		CreatedAt:   time.Now(),
	}

	inputs := map[string]interface{}{
		"user_id":      "user123",
		"request_type": "analysis",
	}

	// 测试上下文准备
	context := crew.prepareTaskContext(inputs, taskOutputs, lastOutput)

	// 验证初始输入
	if context["user_id"] != "user123" {
		t.Error("context should include initial inputs")
	}
	if context["request_type"] != "analysis" {
		t.Error("context should include all initial inputs")
	}

	// 验证任务历史
	if context["previous_tasks_output"] == nil {
		t.Error("context should include previous task outputs")
	}

	// 验证最后任务输出
	if context["last_task_output"] != "Latest task output" {
		t.Error("context should include last task output")
	}
	if context["last_task_json"] == nil {
		t.Error("context should include last task JSON if available")
	}

	// 验证crew信息
	if context["crew_name"] != "crew" {
		t.Error("context should include crew name")
	}
	if context["completed_tasks"] != 2 {
		t.Errorf("context should show 2 completed tasks, got %v", context["completed_tasks"])
	}
}

func TestEventEmission(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := events.NewEventBus(logger)

	// 创建事件收集器 - 使用互斥锁保护并发访问
	var capturedEvents []events.Event
	var mutex sync.RWMutex

	eventCapture := func(ctx context.Context, event events.Event) error {
		mutex.Lock()
		capturedEvents = append(capturedEvents, event)
		mutex.Unlock()
		return nil
	}

	eventBus.Subscribe("sequential_process_started", eventCapture)
	eventBus.Subscribe("sequential_process_completed", eventCapture)
	eventBus.Subscribe("task_execution_started", eventCapture)
	eventBus.Subscribe("task_execution_completed", eventCapture)

	crew := NewBaseCrew(nil, eventBus, logger)
	crew.SetProcess(ProcessSequential)

	// 添加agent和task
	agent := &MockAgent{id: "a1", role: "TestAgent", goal: "Test", backstory: "Test"}
	task := &MockTask{id: "t1", description: "Test task", expectedOutput: "Test output"}

	crew.AddAgent(agent)
	crew.AddTask(task)

	// 执行crew
	ctx := context.Background()
	inputs := map[string]interface{}{"test": "data"}

	_, err := crew.Kickoff(ctx, inputs)
	if err != nil {
		t.Fatalf("crew execution failed: %v", err)
	}

	// 等待一小段时间让事件处理器完成
	time.Sleep(10 * time.Millisecond)

	// 验证事件发射
	expectedEventTypes := []string{
		"sequential_process_started",
		"task_execution_started",
		"task_execution_completed",
		"sequential_process_completed",
	}

	mutex.RLock()
	actualEvents := make([]events.Event, len(capturedEvents))
	copy(actualEvents, capturedEvents)
	mutex.RUnlock()

	if len(actualEvents) != len(expectedEventTypes) {
		t.Errorf("expected %d events, got %d", len(expectedEventTypes), len(actualEvents))
		for i, event := range actualEvents {
			t.Logf("event %d: %s", i, event.GetType())
		}
	}

	for i, expectedType := range expectedEventTypes {
		if i < len(actualEvents) && actualEvents[i].GetType() != expectedType {
			t.Errorf("event %d: expected type %s, got %s", i, expectedType, actualEvents[i].GetType())
		}
	}
}

func TestErrorHandling(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := events.NewEventBus(logger)

	t.Run("task_execution_error", func(t *testing.T) {
		crew := NewBaseCrew(nil, eventBus, logger)

		// 创建会失败的agent
		failingAgent := &FailingMockAgent{
			MockAgent:  MockAgent{id: "fail", role: "FailAgent", goal: "Fail", backstory: "Fails"},
			shouldFail: true,
		}
		task := &MockTask{id: "t1", description: "Test task", expectedOutput: "Test output"}

		crew.AddAgent(failingAgent)
		crew.AddTask(task)

		// 执行crew应该失败
		ctx := context.Background()
		_, err := crew.Kickoff(ctx, map[string]interface{}{})

		if err == nil {
			t.Error("expected error when task execution fails")
		}
	})

	t.Run("no_available_agents", func(t *testing.T) {
		crew := NewBaseCrew(nil, eventBus, logger)
		task := &MockTask{id: "t1", description: "Test task", expectedOutput: "Test output"}

		// 只添加任务，不添加agents
		crew.AddTask(task)

		ctx := context.Background()
		_, err := crew.Kickoff(ctx, map[string]interface{}{})

		if err == nil {
			t.Error("expected error when no agents available")
		}
	})
}

// FailingMockAgent 用于测试错误处理的Agent
type FailingMockAgent struct {
	MockAgent
	shouldFail bool
}

func (f *FailingMockAgent) Execute(ctx context.Context, task agent.Task) (*agent.TaskOutput, error) {
	if f.shouldFail {
		return nil, errors.New("simulated task execution failure")
	}
	return f.MockAgent.Execute(ctx, task)
}
