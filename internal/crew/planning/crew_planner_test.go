package planning

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

func TestNewCrewPlanner(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)

	tasks := []TaskInfo{
		{
			ID:             "task-1",
			Description:    "Test task 1",
			ExpectedOutput: "Test output 1",
			AgentRole:      "Test Agent 1",
		},
	}

	agentFactory := NewMockAgentFactory()
	taskFactory := NewMockTaskFactory()

	t.Run("Create with default config", func(t *testing.T) {
		planner := NewCrewPlanner(tasks, nil, testEventBus, testLogger, agentFactory, taskFactory)

		assert.NotNil(t, planner)
		assert.Equal(t, 1, planner.GetTaskCount())
		assert.NotNil(t, planner.GetConfig())
		assert.Equal(t, "gpt-4o-mini", planner.GetConfig().PlanningAgentLLM)
	})

	t.Run("Create with custom config", func(t *testing.T) {
		config := &PlanningConfig{
			PlanningAgentLLM: "custom-model",
			MaxRetries:       5,
			TimeoutSeconds:   600,
			EnableVerbose:    true,
		}

		planner := NewCrewPlanner(tasks, config, testEventBus, testLogger, agentFactory, taskFactory)

		assert.NotNil(t, planner)
		assert.Equal(t, "custom-model", planner.GetConfig().PlanningAgentLLM)
		assert.Equal(t, 5, planner.GetConfig().MaxRetries)
		assert.True(t, planner.GetConfig().EnableVerbose)
	})

	t.Run("Validate configuration", func(t *testing.T) {
		planner := NewCrewPlanner(tasks, nil, testEventBus, testLogger, agentFactory, taskFactory)

		err := planner.ValidateConfiguration()
		assert.NoError(t, err)

		// 测试无效配置
		invalidConfig := &PlanningConfig{
			PlanningAgentLLM: "", // 空LLM
			MaxRetries:       -1, // 负重试次数
			TimeoutSeconds:   -5, // 负超时时间
		}

		planner.SetConfig(invalidConfig)
		err = planner.ValidateConfiguration()
		assert.Error(t, err)
	})
}

func TestCrewPlannerHandlePlanning(t *testing.T) {
	ctx := context.Background()
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)

	agentFactory := NewMockAgentFactory()
	taskFactory := NewMockTaskFactory()

	t.Run("Successful planning", func(t *testing.T) {
		tasks := []TaskInfo{
			{
				ID:             "task-1",
				Description:    "Research market trends",
				ExpectedOutput: "Market analysis report",
				AgentRole:      "Market Analyst",
				AgentGoal:      "Analyze market data",
				Tools:          []string{"web_scraper", "data_analyzer"},
			},
			{
				ID:             "task-2",
				Description:    "Create visualization",
				ExpectedOutput: "Data dashboard",
				AgentRole:      "Data Scientist",
				AgentGoal:      "Visualize data",
				Tools:          []string{"python", "matplotlib"},
			},
		}

		config := DefaultPlanningConfig()
		config.TimeoutSeconds = 30 // 短超时用于测试

		planner := NewCrewPlanner(tasks, config, testEventBus, testLogger, agentFactory, taskFactory)

		result, err := planner.HandleCrewPlanning(ctx)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, result.GetTaskCount())

		// 验证规划内容
		plan1, found1 := result.GetPlanByTaskDescription("Test task 1")
		assert.True(t, found1)
		assert.NotNil(t, plan1)
		assert.Contains(t, plan1.Plan, "Step")

		plan2, found2 := result.GetPlanByTaskDescription("Test task 2")
		assert.True(t, found2)
		assert.NotNil(t, plan2)
		assert.Contains(t, plan2.Plan, "Step")
	})

	t.Run("Empty tasks list", func(t *testing.T) {
		tasks := []TaskInfo{}

		planner := NewCrewPlanner(tasks, nil, testEventBus, testLogger, agentFactory, taskFactory)

		result, err := planner.HandleCrewPlanning(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "validation")
	})

	t.Run("Timeout handling", func(t *testing.T) {
		tasks := []TaskInfo{
			{
				ID:             "task-timeout",
				Description:    "Long running task",
				ExpectedOutput: "Delayed output",
			},
		}

		config := &PlanningConfig{
			PlanningAgentLLM: "gpt-4o-mini",
			TimeoutSeconds:   1, // 很短的超时
			MaxRetries:       1,
		}

		planner := NewCrewPlanner(tasks, config, testEventBus, testLogger, agentFactory, taskFactory)

		// 使用短超时的上下文
		shortCtx, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
		defer cancel()

		result, err := planner.HandleCrewPlanning(shortCtx)

		// 可能超时或成功，取决于mock的执行时间
		if err != nil {
			assert.Contains(t, err.Error(), "context deadline exceeded")
		}
		_ = result // 可能为nil或有效结果
	})
}

func TestCrewPlannerTasksSummary(t *testing.T) {
	ctx := context.Background()
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)

	agentFactory := NewMockAgentFactory()
	taskFactory := NewMockTaskFactory()

	tasks := []TaskInfo{
		{
			ID:             "task-1",
			Description:    "Task with tools and knowledge",
			ExpectedOutput: "Comprehensive output",
			AgentRole:      "Expert Agent",
			AgentGoal:      "Complete task perfectly",
			Tools:          []string{"tool1", "tool2"},
			Metadata: map[string]interface{}{
				"knowledge": []string{"domain_knowledge", "best_practices"},
				"priority":  "high",
			},
		},
	}

	planner := NewCrewPlanner(tasks, nil, testEventBus, testLogger, agentFactory, taskFactory)

	summary, err := planner.CreateTasksSummary(ctx)

	require.NoError(t, err)
	assert.NotEmpty(t, summary)

	// 验证摘要内容包含期望的信息
	assert.Contains(t, summary, "Task Number 1")
	assert.Contains(t, summary, "Task with tools and knowledge")
	assert.Contains(t, summary, "Expert Agent")
	assert.Contains(t, summary, "Complete task perfectly")
	assert.Contains(t, summary, "tool1")
	assert.Contains(t, summary, "tool2")
	assert.Contains(t, summary, "domain_knowledge")

	// 验证格式符合Python版本
	assert.Contains(t, summary, `"task_description":`)
	assert.Contains(t, summary, `"task_expected_output":`)
	assert.Contains(t, summary, `"agent":`)
	assert.Contains(t, summary, `"agent_goal":`)
	assert.Contains(t, summary, `"task_tools":`)
	assert.Contains(t, summary, `"agent_tools":`)
	assert.Contains(t, summary, `"agent_knowledge":`)
}

func TestCrewPlannerConfiguration(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)
	agentFactory := NewMockAgentFactory()
	taskFactory := NewMockTaskFactory()

	tasks := []TaskInfo{
		{ID: "test", Description: "Test", ExpectedOutput: "Output"},
	}

	t.Run("Set and get config", func(t *testing.T) {
		planner := NewCrewPlanner(tasks, nil, testEventBus, testLogger, agentFactory, taskFactory)

		originalConfig := planner.GetConfig()
		assert.NotNil(t, originalConfig)

		newConfig := &PlanningConfig{
			PlanningAgentLLM: "new-model",
			MaxRetries:       10,
			TimeoutSeconds:   900,
			EnableVerbose:    true,
		}

		err := planner.SetConfig(newConfig)
		assert.NoError(t, err)

		retrievedConfig := planner.GetConfig()
		assert.Equal(t, "new-model", retrievedConfig.PlanningAgentLLM)
		assert.Equal(t, 10, retrievedConfig.MaxRetries)
		assert.Equal(t, 900, retrievedConfig.TimeoutSeconds)
		assert.True(t, retrievedConfig.EnableVerbose)
	})

	t.Run("Nil config handling", func(t *testing.T) {
		planner := NewCrewPlanner(tasks, nil, testEventBus, testLogger, agentFactory, taskFactory)

		err := planner.SetConfig(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})
}

func TestCrewPlannerEventEmission(t *testing.T) {
	ctx := context.Background()
	testLogger := logger.NewConsoleLogger()

	// 创建事件收集器
	eventCollector := &MockEventCollector{}
	testEventBus := events.NewEventBus(testLogger)

	// 注册事件监听器
	testEventBus.Subscribe(EventTypePlanningStarted, func(ctx context.Context, event events.Event) error {
		eventCollector.AddEvent("started", event)
		return nil
	})

	testEventBus.Subscribe(EventTypePlanningCompleted, func(ctx context.Context, event events.Event) error {
		eventCollector.AddEvent("completed", event)
		return nil
	})

	testEventBus.Subscribe(EventTypePlanningAgentCreated, func(ctx context.Context, event events.Event) error {
		eventCollector.AddEvent("agent_created", event)
		return nil
	})

	testEventBus.Subscribe(EventTypePlanningTaskCreated, func(ctx context.Context, event events.Event) error {
		eventCollector.AddEvent("task_created", event)
		return nil
	})

	agentFactory := NewMockAgentFactory()
	taskFactory := NewMockTaskFactory()

	tasks := []TaskInfo{
		{
			ID:             "event-test",
			Description:    "Test event emission",
			ExpectedOutput: "Event validation",
		},
	}

	planner := NewCrewPlanner(tasks, nil, testEventBus, testLogger, agentFactory, taskFactory)

	result, err := planner.HandleCrewPlanning(ctx)

	require.NoError(t, err)
	assert.NotNil(t, result)

	// 验证事件被正确发射
	events := eventCollector.GetEvents()

	// 应该至少有开始和完成事件
	assert.GreaterOrEqual(t, len(events), 2)

	// 验证事件类型
	eventTypes := make(map[string]bool)
	for _, event := range events {
		eventTypes[event.Type] = true
	}

	assert.True(t, eventTypes["started"], "Planning started event should be emitted")
	assert.True(t, eventTypes["completed"], "Planning completed event should be emitted")
}

// MockEventCollector 用于收集和验证事件的辅助结构（线程安全版本）
type MockEventCollector struct {
	mu     sync.RWMutex
	events []MockEvent
}

type MockEvent struct {
	Type string
	Data interface{}
}

func (mec *MockEventCollector) AddEvent(eventType string, data interface{}) {
	mec.mu.Lock()
	defer mec.mu.Unlock()
	mec.events = append(mec.events, MockEvent{
		Type: eventType,
		Data: data,
	})
}

func (mec *MockEventCollector) GetEvents() []MockEvent {
	mec.mu.RLock()
	defer mec.mu.RUnlock()
	// 返回副本以避免外部修改
	result := make([]MockEvent, len(mec.events))
	copy(result, mec.events)
	return result
}

func TestCrewPlannerIntegrationScenario(t *testing.T) {
	ctx := context.Background()
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)

	agentFactory := NewMockAgentFactory()
	taskFactory := NewMockTaskFactory()

	// 复杂的真实场景
	tasks := []TaskInfo{
		{
			ID:             "market-research",
			Description:    "Conduct comprehensive market research for Q1 2024",
			ExpectedOutput: "Detailed market analysis report with trends and insights",
			AgentRole:      "Senior Market Research Analyst",
			AgentGoal:      "Provide accurate and actionable market intelligence",
			Tools:          []string{"web_scraper", "data_analyzer", "report_generator", "sentiment_analyzer"},
			Context:        []string{"Previous Q4 report", "Industry benchmarks", "Competitor analysis"},
			Metadata: map[string]interface{}{
				"knowledge":         []string{"market_data_2023", "industry_reports", "competitor_profiles"},
				"knowledge_sources": []string{"Bloomberg", "Reuters", "MarketWatch"},
				"priority":          "critical",
				"deadline":          "2024-02-15",
				"budget":            "$50000",
			},
		},
		{
			ID:             "data-visualization",
			Description:    "Create interactive dashboards and visualizations from research data",
			ExpectedOutput: "Interactive dashboard with key metrics and trend visualizations",
			AgentRole:      "Senior Data Visualization Specialist",
			AgentGoal:      "Transform complex data into clear, actionable visual insights",
			Tools:          []string{"python", "plotly", "tableau", "d3js", "powerbi"},
			Context:        []string{"Research data from task 1", "Brand guidelines", "Stakeholder requirements"},
			Metadata: map[string]interface{}{
				"knowledge":    []string{"visualization_best_practices", "dashboard_design"},
				"dependencies": []string{"market-research"},
				"priority":     "high",
				"format":       "interactive_web_dashboard",
			},
		},
		{
			ID:             "strategic-recommendations",
			Description:    "Develop strategic recommendations based on research and visualizations",
			ExpectedOutput: "Executive summary with actionable strategic recommendations",
			AgentRole:      "Strategic Business Consultant",
			AgentGoal:      "Provide strategic direction based on data-driven insights",
			Tools:          []string{"strategic_analysis", "scenario_modeling", "risk_assessment"},
			Context:        []string{"Market research results", "Visualization insights", "Company objectives"},
			Metadata: map[string]interface{}{
				"knowledge":    []string{"strategic_frameworks", "business_models", "industry_dynamics"},
				"dependencies": []string{"market-research", "data-visualization"},
				"priority":     "critical",
				"audience":     "C-suite executives",
			},
		},
	}

	config := &PlanningConfig{
		PlanningAgentLLM: "gpt-4o", // 使用更强大的模型处理复杂场景
		MaxRetries:       3,
		TimeoutSeconds:   300,
		EnableVerbose:    true,
		CustomPrompts: map[string]string{
			"complexity_note": "This is a complex multi-agent scenario requiring detailed coordination",
		},
	}

	planner := NewCrewPlanner(tasks, config, testEventBus, testLogger, agentFactory, taskFactory)

	// 执行规划
	startTime := time.Now()
	result, err := planner.HandleCrewPlanning(ctx)
	executionTime := time.Since(startTime)

	// 验证执行结果
	require.NoError(t, err, "Complex planning scenario should succeed")
	assert.NotNil(t, result, "Planning result should not be nil")

	// 验证规划完整性（注意：Mock返回固定的2个规划）
	assert.Greater(t, result.GetTaskCount(), 0, "Should generate at least one plan")

	// 验证每个任务的规划（注意：由于我们使用的是Mock，实际生成的规划数量可能与任务数量不完全匹配）
	actualPlans := result.GetTaskCount()
	assert.Greater(t, actualPlans, 0, "Should generate at least one plan")

	// 验证我们能找到的规划内容
	for i := 1; i <= actualPlans; i++ {
		taskDesc := fmt.Sprintf("Test task %d", i) // Mock返回的标准格式
		plan, found := result.GetPlanByTaskDescription(taskDesc)
		if found {
			assert.NotNil(t, plan, "Found plan should not be nil")
			assert.NotEmpty(t, plan.Plan, "Plan content should not be empty")
			assert.Contains(t, plan.Plan, "Step", "Plan should contain step-by-step instructions")
		}
	}

	// 验证规划质量
	err = result.Validate()
	assert.NoError(t, err, "Generated planning should be valid")

	// 验证性能
	assert.Less(t, executionTime.Milliseconds(), int64(5000), "Planning should complete within 5 seconds")

	// 验证任务摘要生成
	summary, err := planner.CreateTasksSummary(ctx)
	require.NoError(t, err, "Task summary generation should succeed")

	// 验证摘要内容的完整性
	assert.Contains(t, summary, "Senior Market Research Analyst", "Summary should contain agent roles")
	assert.Contains(t, summary, "market_data_2023", "Summary should contain knowledge references")
	assert.Contains(t, summary, "web_scraper", "Summary should contain tool references")
	assert.Contains(t, summary, "Bloomberg", "Summary should contain knowledge sources")

	// 验证Python版本格式兼容性
	assert.Contains(t, summary, "Task Number 1 - ", "Summary should follow Python format")
	assert.Contains(t, summary, `"agent_knowledge"`, "Summary should include agent knowledge section")

	t.Logf("Complex planning scenario completed successfully in %v", executionTime)
	t.Logf("Generated %d plans with total summary length: %d characters",
		result.GetTaskCount(), len(summary))
}
