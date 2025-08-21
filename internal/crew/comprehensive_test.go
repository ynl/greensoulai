package crew

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ynl/greensoulai/internal/agent"
	"github.com/ynl/greensoulai/internal/llm"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// ComprehensiveTestSuite 全面测试套件，验证与Python版本的一致性
func TestComprehensiveCrewFeatures(t *testing.T) {
	t.Run("Context Passing Mechanism", testContextPassing)
	t.Run("Callback Functions", testCallbackFunctions)
	t.Run("Agent Selection Logic", testAgentSelection)
	t.Run("Task Output Formats", testOutputFormats)
	t.Run("Error Handling", testErrorHandling)
	t.Run("Event System", testEventSystem)
	t.Run("Memory Integration", testMemoryIntegration)
	t.Run("Usage Metrics", testUsageMetrics)
}

// testContextPassing 测试上下文传递机制，对标Python版本的_get_context方法
func testContextPassing(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := events.NewEventBus(logger)

	// 创建Mock LLM
	mockLLM := &MockLLM{
		responses: []string{
			"Task 1 result: AI research completed",
			"Task 2 result: Analysis based on AI research completed",
			"Task 3 result: Final report combining research and analysis",
		},
	}

	// 创建agents
	researcher, err := createTestAgent("Senior Researcher", "Conduct research", mockLLM, eventBus, logger)
	if err != nil {
		t.Fatalf("Failed to create researcher: %v", err)
	}

	analyst, err := createTestAgent("Data Analyst", "Analyze data", mockLLM, eventBus, logger)
	if err != nil {
		t.Fatalf("Failed to create analyst: %v", err)
	}

	writer, err := createTestAgent("Report Writer", "Write reports", mockLLM, eventBus, logger)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}

	// 创建crew
	crewConfig := &CrewConfig{
		Name:    "ResearchCrew",
		Process: ProcessSequential,
		Verbose: true,
	}
	crew := NewBaseCrew(crewConfig, eventBus, logger)

	// 添加agents
	crew.AddAgent(researcher)
	crew.AddAgent(analyst)
	crew.AddAgent(writer)

	// 创建上下文感知任务
	task1 := NewContextAwareTask("t1", "Research AI trends", "AI research report")
	task2 := NewContextAwareTask("t2", "Analyze research findings", "Analysis report")
	task3 := NewContextAwareTask("t3", "Write final report", "Comprehensive report")

	crew.AddTask(task1)
	crew.AddTask(task2)
	crew.AddTask(task3)

	// 执行crew
	ctx := context.Background()
	inputs := map[string]interface{}{
		"topic": "Artificial Intelligence",
		"year":  "2024",
	}

	output, err := crew.Kickoff(ctx, inputs)
	if err != nil {
		t.Fatalf("Crew execution failed: %v", err)
	}

	// 验证结果
	if len(output.TasksOutput) != 3 {
		t.Errorf("Expected 3 task outputs, got %d", len(output.TasksOutput))
	}

	// 验证上下文传递
	task2Context := task2.GetContext()
	if task2Context["last_task_output"] == nil {
		t.Error("Task 2 should have received context from Task 1")
	}

	task3Context := task3.GetContext()

	// 验证Python版本对齐的上下文传递机制
	if task3Context["aggregated_context"] == nil {
		t.Error("Task 3 should have received aggregated context (Python-style)")
	}

	// 验证聚合上下文包含正确的分隔符格式
	if aggregatedCtx, ok := task3Context["aggregated_context"].(string); ok {
		if !strings.Contains(aggregatedCtx, "----------") {
			t.Error("Aggregated context should contain Python-style dividers")
		}
	}

	if len(task3Context) < 5 { // 应该有inputs + crew info + task outputs
		t.Errorf("Task 3 context seems incomplete, got %d keys", len(task3Context))
	}

	t.Logf("✅ Context passing test completed successfully")
}

// testCallbackFunctions 测试回调函数功能
func testCallbackFunctions(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := events.NewEventBus(logger)
	mockLLM := &MockLLM{responses: []string{"Callback test result"}}

	// 创建带回调的crew
	var callbackExecuted bool
	var callbackTaskDesc string
	var callbackOutput string

	taskCallback := func(ctx context.Context, task agent.Task, output *agent.TaskOutput) error {
		callbackExecuted = true
		callbackTaskDesc = task.GetDescription()
		callbackOutput = output.Raw
		return nil
	}

	crewConfig := &CrewConfig{
		Name:         "CallbackCrew",
		Process:      ProcessSequential,
		TaskCallback: taskCallback,
	}
	crew := NewBaseCrew(crewConfig, eventBus, logger)

	// 添加agent和task
	testAgent, _ := createTestAgent("Test Agent", "Test callbacks", mockLLM, eventBus, logger)
	crew.AddAgent(testAgent)

	testTask := agent.NewBaseTask("Test callback functionality", "Callback test output")
	crew.AddTask(testTask)

	// 执行
	ctx := context.Background()
	_, err := crew.Kickoff(ctx, nil)
	if err != nil {
		t.Fatalf("Crew execution failed: %v", err)
	}

	// 验证回调执行
	if !callbackExecuted {
		t.Error("Task callback was not executed")
	}
	if callbackTaskDesc != "Test callback functionality" {
		t.Error("Callback received wrong task description")
	}
	if callbackOutput != "Callback test result" {
		t.Error("Callback received wrong output")
	}

	t.Logf("✅ Callback functions test completed successfully")
}

// testAgentSelection 测试Agent选择逻辑，对标Python版本的_get_agent_to_use方法
func testAgentSelection(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := events.NewEventBus(logger)
	mockLLM := &MockLLM{responses: []string{"Agent selection test"}}

	tests := []struct {
		name         string
		agentCount   int
		taskCount    int
		process      Process
		expectedFunc func(agentRoles []string) bool
	}{
		{
			name:       "Sequential_MoreAgentsThanTasks",
			agentCount: 3,
			taskCount:  2,
			process:    ProcessSequential,
			expectedFunc: func(roles []string) bool {
				// 应该使用前两个agents
				return roles[0] == "Agent0" && roles[1] == "Agent1"
			},
		},
		{
			name:       "Sequential_MoreTasksThanAgents",
			agentCount: 2,
			taskCount:  4,
			process:    ProcessSequential,
			expectedFunc: func(roles []string) bool {
				// 应该循环使用agents: Agent0, Agent1, Agent0, Agent1
				return roles[0] == "Agent0" && roles[1] == "Agent1" &&
					roles[2] == "Agent0" && roles[3] == "Agent1"
			},
		},
		{
			name:       "Hierarchical_UsesManagerAgent",
			agentCount: 3,
			taskCount:  2,
			process:    ProcessHierarchical,
			expectedFunc: func(roles []string) bool {
				// 应该都使用Manager Agent
				return strings.Contains(roles[0], "Manager") && strings.Contains(roles[1], "Manager")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var crewConfig *CrewConfig
			if tt.process == ProcessHierarchical {
				crewConfig = &CrewConfig{
					Name:       "AgentSelectionCrew",
					Process:    tt.process,
					ManagerLLM: mockLLM,
				}
			} else {
				crewConfig = &CrewConfig{
					Name:    "AgentSelectionCrew",
					Process: tt.process,
				}
			}

			crew := NewBaseCrew(crewConfig, eventBus, logger)

			// 创建agents
			for i := 0; i < tt.agentCount; i++ {
				agentName := fmt.Sprintf("Agent%d", i)
				testAgent, _ := createTestAgent(agentName, "Test agent selection", mockLLM, eventBus, logger)
				crew.AddAgent(testAgent)
			}

			// 创建tasks
			for i := 0; i < tt.taskCount; i++ {
				taskDesc := fmt.Sprintf("Task %d", i+1)
				testTask := agent.NewBaseTask(taskDesc, "Test output")
				crew.AddTask(testTask)
			}

			// 执行crew
			ctx := context.Background()
			output, err := crew.Kickoff(ctx, nil)
			if err != nil {
				t.Fatalf("Crew execution failed: %v", err)
			}

			// 收集执行agent的角色
			agentRoles := make([]string, len(output.TasksOutput))
			for i, taskOutput := range output.TasksOutput {
				agentRoles[i] = taskOutput.Agent
			}

			// 验证agent选择逻辑
			if !tt.expectedFunc(agentRoles) {
				t.Errorf("Agent selection logic failed. Got roles: %v", agentRoles)
			}
		})
	}

	t.Logf("✅ Agent selection test completed successfully")
}

// testOutputFormats 测试输出格式处理
func testOutputFormats(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := events.NewEventBus(logger)

	// 创建返回JSON格式的Mock LLM
	jsonResponse := `{"result": "success", "score": 95, "details": "Analysis completed successfully"}`
	mockLLM := &MockLLM{responses: []string{jsonResponse}}

	crewConfig := &CrewConfig{
		Name:    "OutputFormatCrew",
		Process: ProcessSequential,
	}
	crew := NewBaseCrew(crewConfig, eventBus, logger)

	testAgent, _ := createTestAgent("Format Test Agent", "Test output formats", mockLLM, eventBus, logger)
	crew.AddAgent(testAgent)

	// 创建期望JSON输出的任务
	jsonTask := agent.NewTaskWithOptions(
		"Generate JSON analysis",
		"Analysis results in JSON format",
		agent.WithOutputFormat(agent.OutputFormatJSON),
	)
	crew.AddTask(jsonTask)

	// 执行
	ctx := context.Background()
	output, err := crew.Kickoff(ctx, nil)
	if err != nil {
		t.Fatalf("Crew execution failed: %v", err)
	}

	// 验证输出格式
	taskOutput := output.TasksOutput[0]
	if taskOutput.JSON == nil {
		t.Error("Expected JSON output but got nil")
	}

	if taskOutput.OutputFormat != agent.OutputFormatJSON {
		t.Errorf("Expected JSON output format, got %v", taskOutput.OutputFormat)
	}

	// 验证JSON内容
	if result, ok := taskOutput.JSON["result"].(string); !ok || result != "success" {
		t.Error("JSON output format or content is incorrect")
	}

	t.Logf("✅ Output formats test completed successfully")
}

// testErrorHandling 测试错误处理机制
func testErrorHandling(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := events.NewEventBus(logger)

	tests := []struct {
		name          string
		setupCrew     func() *BaseCrew
		expectedError string
	}{
		{
			name: "NoAgents",
			setupCrew: func() *BaseCrew {
				crew := NewBaseCrew(nil, eventBus, logger)
				testTask := agent.NewBaseTask("Test task", "Test output")
				crew.AddTask(testTask)
				return crew
			},
			expectedError: "crew must have at least one agent",
		},
		{
			name: "NoTasks",
			setupCrew: func() *BaseCrew {
				crew := NewBaseCrew(nil, eventBus, logger)
				mockLLM := &MockLLM{responses: []string{"test"}}
				testAgent, _ := createTestAgent("Test Agent", "Test", mockLLM, eventBus, logger)
				crew.AddAgent(testAgent)
				return crew
			},
			expectedError: "crew must have at least one task",
		},
		{
			name: "HierarchicalWithoutManager",
			setupCrew: func() *BaseCrew {
				crewConfig := &CrewConfig{
					Name:    "ErrorCrew",
					Process: ProcessHierarchical,
					// 没有ManagerLLM或ManagerAgent
				}
				crew := NewBaseCrew(crewConfig, eventBus, logger)
				mockLLM := &MockLLM{responses: []string{"test"}}
				testAgent, _ := createTestAgent("Test Agent", "Test", mockLLM, eventBus, logger)
				crew.AddAgent(testAgent)
				testTask := agent.NewBaseTask("Test task", "Test output")
				crew.AddTask(testTask)
				return crew
			},
			expectedError: "hierarchical process requires either manager agent or manager LLM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crew := tt.setupCrew()

			ctx := context.Background()
			_, err := crew.Kickoff(ctx, nil)

			if err == nil {
				t.Errorf("Expected error containing '%s', but got no error", tt.expectedError)
			} else if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, err.Error())
			}
		})
	}

	t.Logf("✅ Error handling test completed successfully")
}

// testEventSystem 测试事件系统完整性
func testEventSystem(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := events.NewEventBus(logger)
	mockLLM := &MockLLM{responses: []string{"Event test result"}}

	// 事件收集器
	var capturedEvents []events.Event
	var eventMutex sync.Mutex

	// 订阅所有相关事件
	eventTypes := []string{
		"crew_kickoff_started",
		"sequential_process_started",
		"task_execution_started",
		"task_execution_completed",
		"sequential_process_completed",
		"crew_kickoff_completed",
	}

	for _, eventType := range eventTypes {
		eventBus.Subscribe(eventType, func(ctx context.Context, event events.Event) error {
			eventMutex.Lock()
			capturedEvents = append(capturedEvents, event)
			eventMutex.Unlock()
			return nil
		})
	}

	// 创建和执行crew
	crewConfig := &CrewConfig{Name: "EventCrew", Process: ProcessSequential}
	crew := NewBaseCrew(crewConfig, eventBus, logger)

	testAgent, _ := createTestAgent("Event Test Agent", "Test events", mockLLM, eventBus, logger)
	crew.AddAgent(testAgent)

	testTask := agent.NewBaseTask("Test event emission", "Event test output")
	crew.AddTask(testTask)

	ctx := context.Background()
	_, err := crew.Kickoff(ctx, nil)
	if err != nil {
		t.Fatalf("Crew execution failed: %v", err)
	}

	// 等待事件处理
	time.Sleep(50 * time.Millisecond)

	// 验证事件发射
	eventMutex.Lock()
	eventCount := len(capturedEvents)
	eventMutex.Unlock()

	if eventCount < 4 { // 至少应该有kickoff_started, task_started, task_completed, kickoff_completed
		t.Errorf("Expected at least 4 events, got %d", eventCount)
	}

	// 验证事件类型
	eventTypesSeen := make(map[string]bool)
	eventMutex.Lock()
	for _, event := range capturedEvents {
		eventTypesSeen[event.GetType()] = true
	}
	eventMutex.Unlock()

	requiredEvents := []string{"crew_kickoff_started", "task_execution_started", "task_execution_completed"}
	for _, requiredEvent := range requiredEvents {
		if !eventTypesSeen[requiredEvent] {
			t.Errorf("Required event '%s' was not emitted", requiredEvent)
		}
	}

	t.Logf("✅ Event system test completed successfully with %d events", eventCount)
}

// testMemoryIntegration 测试内存集成（基础测试）
func testMemoryIntegration(t *testing.T) {
	// 这是一个占位符测试，因为当前的实现可能还没有完整的内存系统
	// 但我们可以测试相关的接口是否存在

	logger := logger.NewTestLogger()
	eventBus := events.NewEventBus(logger)
	mockLLM := &MockLLM{responses: []string{"Memory test"}}

	crewConfig := &CrewConfig{
		Name:          "MemoryCrew",
		Process:       ProcessSequential,
		MemoryEnabled: true, // 启用内存
	}
	crew := NewBaseCrew(crewConfig, eventBus, logger)

	// 验证内存配置
	if !crew.memoryEnabled {
		t.Error("Memory should be enabled")
	}

	testAgent, _ := createTestAgent("Memory Test Agent", "Test memory", mockLLM, eventBus, logger)
	crew.AddAgent(testAgent)

	testTask := agent.NewBaseTask("Test memory integration", "Memory test output")
	crew.AddTask(testTask)

	ctx := context.Background()
	_, err := crew.Kickoff(ctx, nil)
	if err != nil {
		t.Fatalf("Crew execution failed: %v", err)
	}

	t.Logf("✅ Memory integration test completed successfully")
}

// testUsageMetrics 测试使用指标统计
func testUsageMetrics(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := events.NewEventBus(logger)
	mockLLM := &MockLLM{responses: []string{"Usage metrics test"}}

	crewConfig := &CrewConfig{Name: "MetricsCrew", Process: ProcessSequential}
	crew := NewBaseCrew(crewConfig, eventBus, logger)

	testAgent, _ := createTestAgent("Metrics Test Agent", "Test metrics", mockLLM, eventBus, logger)
	crew.AddAgent(testAgent)

	testTask := agent.NewBaseTask("Test usage metrics", "Metrics test output")
	crew.AddTask(testTask)

	// 记录开始时间
	startTime := time.Now()

	ctx := context.Background()
	output, err := crew.Kickoff(ctx, nil)
	if err != nil {
		t.Fatalf("Crew execution failed: %v", err)
	}

	// 验证基础指标
	if output.TokenUsage == nil {
		t.Error("Token usage metrics should not be nil")
	}

	// 验证执行时间统计
	if crew.totalExecutionTime == 0 {
		t.Error("Total execution time should be recorded")
	}

	// 验证执行计数
	if crew.executionCount == 0 {
		t.Error("Execution count should be incremented")
	}

	// 验证时间合理性
	if time.Since(startTime) < crew.totalExecutionTime {
		// 这个检查确保执行时间被正确记录
		t.Error("Recorded execution time seems unreasonable")
	}

	t.Logf("✅ Usage metrics test completed successfully")
}

// 辅助函数：创建测试Agent
func createTestAgent(role, goal string, mockLLM llm.LLM, eventBus events.EventBus, logger logger.Logger) (agent.Agent, error) {
	config := agent.AgentConfig{
		Role:      role,
		Goal:      goal,
		Backstory: fmt.Sprintf("You are a %s focused on %s", role, goal),
		LLM:       mockLLM,
		EventBus:  eventBus,
		Logger:    logger,
	}

	ag, err := agent.NewBaseAgent(config)
	if err != nil {
		return nil, err
	}

	return ag, ag.Initialize()
}
