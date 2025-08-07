package planning

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ynl/greensoulai/pkg/logger"
)

func TestTaskSummaryGeneratorImpl(t *testing.T) {
	ctx := context.Background()
	testLogger := logger.NewConsoleLogger()
	generator := NewTaskSummaryGenerator(testLogger)

	t.Run("NewTaskSummaryGenerator", func(t *testing.T) {
		assert.NotNil(t, generator)
		assert.IsType(t, &TaskSummaryGeneratorImpl{}, generator)
	})

	t.Run("GenerateTaskSummary - Valid Task", func(t *testing.T) {
		taskInfo := &TaskInfo{
			ID:             "task-1",
			Description:    "Test task description",
			ExpectedOutput: "Test expected output",
			AgentRole:      "Test Agent",
			AgentGoal:      "Test goal",
			Tools:          []string{"tool1", "tool2"},
			Context:        []string{"context1"},
			Metadata:       map[string]interface{}{"key": "value"},
		}

		summary, err := generator.GenerateTaskSummary(ctx, taskInfo, 0)

		require.NoError(t, err)
		assert.NotNil(t, summary)
		assert.Equal(t, 1, summary.TaskNumber)
		assert.Equal(t, "Test task description", summary.Description)
		assert.Equal(t, "Test expected output", summary.ExpectedOutput)
		assert.Equal(t, "Test Agent", summary.AgentRole)
		assert.Equal(t, "Test goal", summary.AgentGoal)
		assert.Equal(t, []string{"tool1", "tool2"}, summary.TaskTools)
		assert.Equal(t, []string{"tool1", "tool2"}, summary.AgentTools)
		assert.Equal(t, map[string]interface{}{"key": "value"}, summary.AdditionalContext)
	})

	t.Run("GenerateTaskSummary - Nil TaskInfo", func(t *testing.T) {
		summary, err := generator.GenerateTaskSummary(ctx, nil, 0)

		assert.Error(t, err)
		assert.Nil(t, summary)
		assert.Contains(t, err.Error(), "taskInfo cannot be nil")
	})

	t.Run("GenerateTaskSummary - Empty Values", func(t *testing.T) {
		taskInfo := &TaskInfo{
			ID:             "task-empty",
			Description:    "Test description",
			ExpectedOutput: "Test output",
			// AgentRole and AgentGoal are empty
			// Tools is empty
		}

		summary, err := generator.GenerateTaskSummary(ctx, taskInfo, 0)

		require.NoError(t, err)
		assert.Equal(t, "None", summary.AgentRole)
		assert.Equal(t, "None", summary.AgentGoal)
		assert.Equal(t, []string{"agent has no tools"}, summary.TaskTools)
		assert.Equal(t, []string{"agent has no tools"}, summary.AgentTools)
	})

	t.Run("GenerateTaskSummary - With Knowledge", func(t *testing.T) {
		taskInfo := &TaskInfo{
			ID:             "task-knowledge",
			Description:    "Task with knowledge",
			ExpectedOutput: "Output with knowledge",
			Metadata: map[string]interface{}{
				"knowledge": []string{"knowledge1", "knowledge2"},
			},
		}

		summary, err := generator.GenerateTaskSummary(ctx, taskInfo, 1)

		require.NoError(t, err)
		assert.Equal(t, 2, summary.TaskNumber)
		assert.Equal(t, []string{"knowledge1", "knowledge2"}, summary.AgentKnowledge)
	})
}

func TestGenerateTasksSummary(t *testing.T) {
	ctx := context.Background()
	testLogger := logger.NewConsoleLogger()
	generator := NewTaskSummaryGenerator(testLogger)

	t.Run("Valid Multiple Tasks", func(t *testing.T) {
		tasks := []TaskInfo{
			{
				ID:             "task-1",
				Description:    "First task",
				ExpectedOutput: "First output",
				AgentRole:      "Agent 1",
				AgentGoal:      "Goal 1",
				Tools:          []string{"tool1"},
			},
			{
				ID:             "task-2",
				Description:    "Second task",
				ExpectedOutput: "Second output",
				AgentRole:      "Agent 2",
				AgentGoal:      "Goal 2",
				Tools:          []string{"tool2", "tool3"},
			},
		}

		summary, err := generator.GenerateTasksSummary(ctx, tasks)

		require.NoError(t, err)
		assert.NotEmpty(t, summary)

		// 验证包含所有任务信息
		assert.Contains(t, summary, "Task Number 1 - First task")
		assert.Contains(t, summary, "Task Number 2 - Second task")
		assert.Contains(t, summary, "First output")
		assert.Contains(t, summary, "Second output")
		assert.Contains(t, summary, "Agent 1")
		assert.Contains(t, summary, "Agent 2")
	})

	t.Run("Empty Tasks List", func(t *testing.T) {
		tasks := []TaskInfo{}

		summary, err := generator.GenerateTasksSummary(ctx, tasks)

		assert.Error(t, err)
		assert.Empty(t, summary)
		assert.Contains(t, err.Error(), "tasks list cannot be empty")
	})

	t.Run("Single Task", func(t *testing.T) {
		tasks := []TaskInfo{
			{
				ID:             "single-task",
				Description:    "Single task description",
				ExpectedOutput: "Single expected output",
			},
		}

		summary, err := generator.GenerateTasksSummary(ctx, tasks)

		require.NoError(t, err)
		assert.Contains(t, summary, "Task Number 1 - Single task description")
		assert.Contains(t, summary, "Single expected output")
		assert.Contains(t, summary, "None") // 默认Agent角色
	})
}

func TestFormatTaskSummary(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	generator := NewTaskSummaryGenerator(testLogger).(*TaskSummaryGeneratorImpl)

	t.Run("Format With Tools", func(t *testing.T) {
		summary := &TaskSummary{
			TaskNumber:     1,
			Description:    "Test task",
			ExpectedOutput: "Test output",
			AgentRole:      "Test Agent",
			AgentGoal:      "Test goal",
			TaskTools:      []string{"tool1", "tool2"},
			AgentTools:     []string{"agent_tool1"},
			AgentKnowledge: []string{},
		}

		formatted := generator.FormatTaskSummary(summary)

		assert.Contains(t, formatted, "Task Number 1 - Test task")
		assert.Contains(t, formatted, `"task_description": Test task`)
		assert.Contains(t, formatted, `"task_expected_output": Test output`)
		assert.Contains(t, formatted, `"agent": Test Agent`)
		assert.Contains(t, formatted, `"agent_goal": Test goal`)
		assert.Contains(t, formatted, `["tool1", "tool2"]`)
		assert.Contains(t, formatted, `["agent_tool1"]`)
	})

	t.Run("Format Without Tools", func(t *testing.T) {
		summary := &TaskSummary{
			TaskNumber:     2,
			Description:    "Task without tools",
			ExpectedOutput: "Output without tools",
			AgentRole:      "Agent",
			AgentGoal:      "Goal",
			TaskTools:      []string{"agent has no tools"},
			AgentTools:     []string{"agent has no tools"},
		}

		formatted := generator.FormatTaskSummary(summary)

		assert.Contains(t, formatted, "Task Number 2")
		assert.Contains(t, formatted, `"agent has no tools"`)
	})

	t.Run("Format With Knowledge", func(t *testing.T) {
		summary := &TaskSummary{
			TaskNumber:     3,
			Description:    "Task with knowledge",
			ExpectedOutput: "Output with knowledge",
			AgentRole:      "Agent",
			AgentGoal:      "Goal",
			TaskTools:      []string{},
			AgentTools:     []string{},
			AgentKnowledge: []string{"knowledge1"},
		}

		formatted := generator.FormatTaskSummary(summary)

		assert.Contains(t, formatted, `"agent_knowledge": "[\"knowledge1\"]"`)
	})
}

func TestExtractAgentKnowledge(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	generator := NewTaskSummaryGenerator(testLogger).(*TaskSummaryGeneratorImpl)

	t.Run("Extract String Knowledge", func(t *testing.T) {
		taskInfo := &TaskInfo{
			Metadata: map[string]interface{}{
				"knowledge": "single knowledge item",
			},
		}

		knowledge := generator.extractAgentKnowledge(taskInfo)

		assert.Equal(t, []string{"single knowledge item"}, knowledge)
	})

	t.Run("Extract Array Knowledge", func(t *testing.T) {
		taskInfo := &TaskInfo{
			Metadata: map[string]interface{}{
				"knowledge": []string{"knowledge1", "knowledge2", "knowledge3"},
			},
		}

		knowledge := generator.extractAgentKnowledge(taskInfo)

		assert.Equal(t, []string{"knowledge1", "knowledge2", "knowledge3"}, knowledge)
	})

	t.Run("Extract Interface Array Knowledge", func(t *testing.T) {
		taskInfo := &TaskInfo{
			Metadata: map[string]interface{}{
				"knowledge": []interface{}{"knowledge1", "knowledge2"},
			},
		}

		knowledge := generator.extractAgentKnowledge(taskInfo)

		assert.Equal(t, []string{"knowledge1", "knowledge2"}, knowledge)
	})

	t.Run("Extract Knowledge Sources", func(t *testing.T) {
		taskInfo := &TaskInfo{
			Metadata: map[string]interface{}{
				"knowledge_sources": []string{"source1", "source2"},
			},
		}

		knowledge := generator.extractAgentKnowledge(taskInfo)

		assert.Equal(t, []string{"source1", "source2"}, knowledge)
	})

	t.Run("No Knowledge", func(t *testing.T) {
		taskInfo := &TaskInfo{
			Metadata: map[string]interface{}{
				"other_field": "value",
			},
		}

		knowledge := generator.extractAgentKnowledge(taskInfo)

		assert.Empty(t, knowledge)
	})

	t.Run("Empty Knowledge", func(t *testing.T) {
		taskInfo := &TaskInfo{
			Metadata: map[string]interface{}{
				"knowledge": "",
			},
		}

		knowledge := generator.extractAgentKnowledge(taskInfo)

		assert.Empty(t, knowledge)
	})

	t.Run("Duplicate Knowledge Removal", func(t *testing.T) {
		taskInfo := &TaskInfo{
			Metadata: map[string]interface{}{
				"knowledge":         []string{"item1", "item2", "item1"},
				"knowledge_sources": []string{"item2", "item3"},
			},
		}

		knowledge := generator.extractAgentKnowledge(taskInfo)

		// 应该去除重复项
		assert.Equal(t, 3, len(knowledge))
		assert.Contains(t, knowledge, "item1")
		assert.Contains(t, knowledge, "item2")
		assert.Contains(t, knowledge, "item3")
	})
}

func TestFormatToolsList(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	generator := NewTaskSummaryGenerator(testLogger).(*TaskSummaryGeneratorImpl)

	t.Run("Format Multiple Tools", func(t *testing.T) {
		tools := []string{"tool1", "tool2", "tool3"}
		formatted := generator.formatToolsList(tools)

		expected := `["tool1", "tool2", "tool3"]`
		assert.Equal(t, expected, formatted)
	})

	t.Run("Format Single Tool", func(t *testing.T) {
		tools := []string{"single_tool"}
		formatted := generator.formatToolsList(tools)

		expected := `["single_tool"]`
		assert.Equal(t, expected, formatted)
	})

	t.Run("Format Empty Tools", func(t *testing.T) {
		tools := []string{}
		formatted := generator.formatToolsList(tools)

		expected := `"agent has no tools"`
		assert.Equal(t, expected, formatted)
	})

	t.Run("Format No Tools Message", func(t *testing.T) {
		tools := []string{"agent has no tools"}
		formatted := generator.formatToolsList(tools)

		expected := `"agent has no tools"`
		assert.Equal(t, expected, formatted)
	})
}

func TestValidateTaskInfo(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	generator := NewTaskSummaryGenerator(testLogger).(*TaskSummaryGeneratorImpl)

	t.Run("Valid Task Info", func(t *testing.T) {
		taskInfo := &TaskInfo{
			Description:    "Valid description",
			ExpectedOutput: "Valid expected output",
		}

		err := generator.ValidateTaskInfo(taskInfo)
		assert.NoError(t, err)
	})

	t.Run("Nil Task Info", func(t *testing.T) {
		err := generator.ValidateTaskInfo(nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task info cannot be nil")
	})

	t.Run("Empty Description", func(t *testing.T) {
		taskInfo := &TaskInfo{
			Description:    "",
			ExpectedOutput: "Valid expected output",
		}

		err := generator.ValidateTaskInfo(taskInfo)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task description cannot be empty")
	})

	t.Run("Empty Expected Output", func(t *testing.T) {
		taskInfo := &TaskInfo{
			Description:    "Valid description",
			ExpectedOutput: "",
		}

		err := generator.ValidateTaskInfo(taskInfo)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task expected output cannot be empty")
	})

	t.Run("Whitespace Only", func(t *testing.T) {
		taskInfo := &TaskInfo{
			Description:    "   \n\t  ",
			ExpectedOutput: "Valid expected output",
		}

		err := generator.ValidateTaskInfo(taskInfo)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task description cannot be empty")
	})
}

func TestGetFormattingSummary(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	generator := NewTaskSummaryGenerator(testLogger).(*TaskSummaryGeneratorImpl)

	t.Run("Complete Summary Statistics", func(t *testing.T) {
		tasks := []TaskInfo{
			{
				Description:    "First task description",
				ExpectedOutput: "First output",
				AgentRole:      "Agent 1",
				Tools:          []string{"tool1"},
				Metadata:       map[string]interface{}{"knowledge": "knowledge1"},
			},
			{
				Description:    "Second task description that is longer",
				ExpectedOutput: "Second output",
				AgentRole:      "Agent 2",
				Tools:          []string{"tool2", "tool3"},
				Metadata:       map[string]interface{}{"other": "value"},
			},
			{
				Description:    "Third task",
				ExpectedOutput: "Third output",
				// No agent, tools, or knowledge
			},
		}

		summary := generator.GetFormattingSummary(tasks)

		assert.Equal(t, 3, summary["total_tasks"])
		assert.Equal(t, 2, summary["tasks_with_agent"])
		assert.Equal(t, 2, summary["tasks_with_tools"])
		assert.Equal(t, 1, summary["tasks_with_knowledge"])

		// 验证平均描述长度
		totalLength := len("First task description") + len("Second task description that is longer") + len("Third task")
		expectedAvg := totalLength / 3
		assert.Equal(t, expectedAvg, summary["average_desc_length"])
	})

	t.Run("Empty Tasks Summary", func(t *testing.T) {
		tasks := []TaskInfo{}

		summary := generator.GetFormattingSummary(tasks)

		assert.Equal(t, 0, summary["total_tasks"])
		assert.Equal(t, 0, summary["average_desc_length"])
	})
}

func TestRemoveDuplicatesAndEmpty(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	generator := NewTaskSummaryGenerator(testLogger).(*TaskSummaryGeneratorImpl)

	t.Run("Remove Duplicates", func(t *testing.T) {
		items := []string{"item1", "item2", "item1", "item3", "item2"}
		result := generator.removeDuplicatesAndEmpty(items)

		assert.Equal(t, 3, len(result))
		assert.Contains(t, result, "item1")
		assert.Contains(t, result, "item2")
		assert.Contains(t, result, "item3")
	})

	t.Run("Remove Empty Strings", func(t *testing.T) {
		items := []string{"item1", "", "item2", "   ", "\t\n", "item3"}
		result := generator.removeDuplicatesAndEmpty(items)

		assert.Equal(t, 3, len(result))
		assert.Contains(t, result, "item1")
		assert.Contains(t, result, "item2")
		assert.Contains(t, result, "item3")
	})

	t.Run("All Empty", func(t *testing.T) {
		items := []string{"", "   ", "\t", "\n"}
		result := generator.removeDuplicatesAndEmpty(items)

		assert.Empty(t, result)
	})

	t.Run("No Changes Needed", func(t *testing.T) {
		items := []string{"item1", "item2", "item3"}
		result := generator.removeDuplicatesAndEmpty(items)

		assert.Equal(t, items, result)
	})
}

func TestIntegrationScenarios(t *testing.T) {
	ctx := context.Background()
	testLogger := logger.NewConsoleLogger()
	generator := NewTaskSummaryGenerator(testLogger)

	t.Run("Complex Real-World Scenario", func(t *testing.T) {
		tasks := []TaskInfo{
			{
				ID:             "research-task",
				Description:    "Research market trends for Q1 2024",
				ExpectedOutput: "Comprehensive market analysis report",
				AgentRole:      "Market Research Analyst",
				AgentGoal:      "Provide accurate market insights",
				Tools:          []string{"web_scraper", "data_analyzer", "report_generator"},
				Context:        []string{"Previous Q4 report", "Industry guidelines"},
				Metadata: map[string]interface{}{
					"knowledge":         []string{"market_data", "competitor_analysis"},
					"knowledge_sources": []string{"Bloomberg", "Reuters"},
					"priority":          "high",
				},
			},
			{
				ID:             "analysis-task",
				Description:    "Analyze research data and create visualizations",
				ExpectedOutput: "Data visualization dashboard",
				AgentRole:      "Data Scientist",
				AgentGoal:      "Transform data into actionable insights",
				Tools:          []string{"python", "matplotlib", "tableau"},
				Metadata: map[string]interface{}{
					"dependencies":       []string{"research-task"},
					"visualization_type": "interactive_dashboard",
				},
			},
		}

		summary, err := generator.GenerateTasksSummary(ctx, tasks)

		require.NoError(t, err)
		assert.NotEmpty(t, summary)

		// 验证包含所有关键信息
		assert.Contains(t, summary, "Market Research Analyst")
		assert.Contains(t, summary, "Data Scientist")
		assert.Contains(t, summary, "web_scraper")
		assert.Contains(t, summary, "python")
		assert.Contains(t, summary, `"agent_knowledge": "[\"market_data\", \"competitor_analysis\", \"Bloomberg\", \"Reuters\"]"`)

		// 验证格式符合Python版本的预期
		lines := strings.Split(summary, "\n")
		assert.True(t, len(lines) > 10) // 应该有多行内容

		// 验证任务编号正确
		assert.Contains(t, summary, "Task Number 1 - Research market trends")
		assert.Contains(t, summary, "Task Number 2 - Analyze research data")
	})

	t.Run("Edge Case with Special Characters", func(t *testing.T) {
		tasks := []TaskInfo{
			{
				Description:    "Task with \"quotes\" and 'apostrophes'",
				ExpectedOutput: "Output with special chars: !@#$%^&*()",
				AgentRole:      "Agent with émojis 🤖",
				Tools:          []string{"tool_with_underscores", "tool-with-dashes"},
				Metadata: map[string]interface{}{
					"knowledge": []string{"知识 with unicode", "knowledge with spaces"},
				},
			},
		}

		summary, err := generator.GenerateTasksSummary(ctx, tasks)

		require.NoError(t, err)

		// 验证特殊字符被正确处理
		assert.Contains(t, summary, "quotes")
		assert.Contains(t, summary, "apostrophes")
		assert.Contains(t, summary, "🤖")
		assert.Contains(t, summary, "知识")
	})
}
