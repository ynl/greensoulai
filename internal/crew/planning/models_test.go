package planning

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanPerTask(t *testing.T) {
	t.Run("Valid PlanPerTask", func(t *testing.T) {
		plan := PlanPerTask{
			Task: "Test task description",
			Plan: "Step 1: Do something\nStep 2: Do something else",
		}

		assert.Equal(t, "Test task description", plan.Task)
		assert.Equal(t, "Step 1: Do something\nStep 2: Do something else", plan.Plan)
	})

	t.Run("ToJSON", func(t *testing.T) {
		plan := PlanPerTask{
			Task: "Test task",
			Plan: "Test plan",
		}

		jsonStr, err := plan.ToJSON()
		require.NoError(t, err)
		assert.Contains(t, jsonStr, "Test task")
		assert.Contains(t, jsonStr, "Test plan")

		// 验证是否为有效JSON
		var parsed map[string]interface{}
		err = json.Unmarshal([]byte(jsonStr), &parsed)
		require.NoError(t, err)
	})
}

func TestPlannerTaskPydanticOutput(t *testing.T) {
	t.Run("Valid Output", func(t *testing.T) {
		output := PlannerTaskPydanticOutput{
			ListOfPlansPerTask: []PlanPerTask{
				{Task: "Task 1", Plan: "Plan 1"},
				{Task: "Task 2", Plan: "Plan 2"},
			},
		}

		assert.Equal(t, 2, len(output.ListOfPlansPerTask))
		assert.Equal(t, "Task 1", output.ListOfPlansPerTask[0].Task)
		assert.Equal(t, "Plan 1", output.ListOfPlansPerTask[0].Plan)
	})

	t.Run("ToJSON and FromJSON", func(t *testing.T) {
		original := PlannerTaskPydanticOutput{
			ListOfPlansPerTask: []PlanPerTask{
				{Task: "Task 1", Plan: "Plan 1"},
				{Task: "Task 2", Plan: "Plan 2"},
			},
		}

		// 转换为JSON
		jsonStr, err := original.ToJSON()
		require.NoError(t, err)

		// 从JSON恢复
		var restored PlannerTaskPydanticOutput
		err = restored.FromJSON(jsonStr)
		require.NoError(t, err)

		// 验证内容一致
		assert.Equal(t, len(original.ListOfPlansPerTask), len(restored.ListOfPlansPerTask))
		assert.Equal(t, original.ListOfPlansPerTask[0].Task, restored.ListOfPlansPerTask[0].Task)
		assert.Equal(t, original.ListOfPlansPerTask[1].Plan, restored.ListOfPlansPerTask[1].Plan)
	})

	t.Run("Validate - Valid Output", func(t *testing.T) {
		output := PlannerTaskPydanticOutput{
			ListOfPlansPerTask: []PlanPerTask{
				{Task: "Task 1", Plan: "Plan 1"},
				{Task: "Task 2", Plan: "Plan 2"},
			},
		}

		err := output.Validate()
		assert.NoError(t, err)
	})

	t.Run("Validate - Empty Plan List", func(t *testing.T) {
		output := PlannerTaskPydanticOutput{
			ListOfPlansPerTask: []PlanPerTask{},
		}

		err := output.Validate()
		assert.Error(t, err)
		assert.Equal(t, ErrEmptyPlanList, err)
	})

	t.Run("Validate - Empty Task", func(t *testing.T) {
		output := PlannerTaskPydanticOutput{
			ListOfPlansPerTask: []PlanPerTask{
				{Task: "", Plan: "Plan 1"},
			},
		}

		err := output.Validate()
		assert.Error(t, err)
		assert.IsType(t, &PlanValidationError{}, err)
	})

	t.Run("Validate - Empty Plan", func(t *testing.T) {
		output := PlannerTaskPydanticOutput{
			ListOfPlansPerTask: []PlanPerTask{
				{Task: "Task 1", Plan: ""},
			},
		}

		err := output.Validate()
		assert.Error(t, err)
		assert.IsType(t, &PlanValidationError{}, err)
	})

	t.Run("GetTaskCount", func(t *testing.T) {
		output := PlannerTaskPydanticOutput{
			ListOfPlansPerTask: []PlanPerTask{
				{Task: "Task 1", Plan: "Plan 1"},
				{Task: "Task 2", Plan: "Plan 2"},
				{Task: "Task 3", Plan: "Plan 3"},
			},
		}

		assert.Equal(t, 3, output.GetTaskCount())
	})

	t.Run("GetPlanByTaskDescription", func(t *testing.T) {
		output := PlannerTaskPydanticOutput{
			ListOfPlansPerTask: []PlanPerTask{
				{Task: "Task 1", Plan: "Plan 1"},
				{Task: "Task 2", Plan: "Plan 2"},
			},
		}

		// 找到存在的任务
		plan, found := output.GetPlanByTaskDescription("Task 1")
		assert.True(t, found)
		assert.Equal(t, "Plan 1", plan.Plan)

		// 找不到的任务
		plan, found = output.GetPlanByTaskDescription("Task 3")
		assert.False(t, found)
		assert.Nil(t, plan)
	})

	t.Run("AddPlan", func(t *testing.T) {
		output := PlannerTaskPydanticOutput{}

		output.AddPlan(PlanPerTask{Task: "Task 1", Plan: "Plan 1"})
		assert.Equal(t, 1, output.GetTaskCount())

		output.AddPlan(PlanPerTask{Task: "Task 2", Plan: "Plan 2"})
		assert.Equal(t, 2, output.GetTaskCount())
	})

	t.Run("String", func(t *testing.T) {
		output := PlannerTaskPydanticOutput{
			ListOfPlansPerTask: []PlanPerTask{
				{Task: "Task 1", Plan: "Plan 1"},
				{Task: "Task 2", Plan: "Plan 2"},
			},
		}

		str := output.String()
		assert.Contains(t, str, "Planning Output:")
		assert.Contains(t, str, "Task 1: Task 1")
		assert.Contains(t, str, "Plan: Plan 1")
		assert.Contains(t, str, "Task 2: Task 2")
		assert.Contains(t, str, "Plan: Plan 2")
	})
}

func TestTaskSummary(t *testing.T) {
	t.Run("Valid TaskSummary", func(t *testing.T) {
		summary := TaskSummary{
			TaskNumber:        1,
			Description:       "Test task",
			ExpectedOutput:    "Test output",
			AgentRole:         "Test agent",
			AgentGoal:         "Test goal",
			TaskTools:         []string{"tool1", "tool2"},
			AgentTools:        []string{"agent_tool1"},
			AgentKnowledge:    []string{"knowledge1"},
			AdditionalContext: map[string]interface{}{"key": "value"},
		}

		assert.Equal(t, 1, summary.TaskNumber)
		assert.Equal(t, "Test task", summary.Description)
		assert.Equal(t, 2, len(summary.TaskTools))
		assert.Equal(t, "value", summary.AdditionalContext["key"])
	})
}

func TestTaskInfo(t *testing.T) {
	t.Run("Valid TaskInfo", func(t *testing.T) {
		taskInfo := TaskInfo{
			ID:             "task-1",
			Description:    "Test task description",
			ExpectedOutput: "Test expected output",
			AgentRole:      "Test Agent",
			AgentGoal:      "Test goal",
			Tools:          []string{"tool1", "tool2"},
			Context:        []string{"context1", "context2"},
			Metadata:       map[string]interface{}{"key": "value"},
		}

		assert.Equal(t, "task-1", taskInfo.ID)
		assert.Equal(t, "Test task description", taskInfo.Description)
		assert.Equal(t, 2, len(taskInfo.Tools))
		assert.Equal(t, 2, len(taskInfo.Context))
		assert.Equal(t, "value", taskInfo.Metadata["key"])
	})
}

func TestPlanningRequest(t *testing.T) {
	t.Run("Valid PlanningRequest", func(t *testing.T) {
		request := PlanningRequest{
			Tasks: []TaskInfo{
				{Description: "Task 1", ExpectedOutput: "Output 1"},
				{Description: "Task 2", ExpectedOutput: "Output 2"},
			},
			PlanningLLM:   "gpt-4o-mini",
			MaxRetries:    3,
			TimeoutSec:    300,
			CustomPrompts: map[string]string{"key": "value"},
		}

		assert.Equal(t, 2, len(request.Tasks))
		assert.Equal(t, "gpt-4o-mini", request.PlanningLLM)
		assert.Equal(t, 3, request.MaxRetries)
		assert.Equal(t, 300, request.TimeoutSec)
	})
}

func TestPlanningResult(t *testing.T) {
	t.Run("Successful PlanningResult", func(t *testing.T) {
		result := PlanningResult{
			Output: PlannerTaskPydanticOutput{
				ListOfPlansPerTask: []PlanPerTask{
					{Task: "Task 1", Plan: "Plan 1"},
				},
			},
			ExecutionTime: 1500.5,
			Success:       true,
			ModelUsed:     "gpt-4o-mini",
			RetryCount:    0,
		}

		assert.True(t, result.Success)
		assert.Equal(t, 1500.5, result.ExecutionTime)
		assert.Equal(t, "gpt-4o-mini", result.ModelUsed)
		assert.Equal(t, 0, result.RetryCount)
		assert.Equal(t, 1, result.Output.GetTaskCount())
	})

	t.Run("Failed PlanningResult", func(t *testing.T) {
		result := PlanningResult{
			Success:       false,
			ErrorMessage:  "Planning failed",
			ExecutionTime: 500.0,
			RetryCount:    2,
		}

		assert.False(t, result.Success)
		assert.Equal(t, "Planning failed", result.ErrorMessage)
		assert.Equal(t, 2, result.RetryCount)
	})
}

func TestPlanningConfig(t *testing.T) {
	t.Run("DefaultPlanningConfig", func(t *testing.T) {
		config := DefaultPlanningConfig()

		assert.NotNil(t, config)
		assert.Equal(t, "gpt-4o-mini", config.PlanningAgentLLM)
		assert.Equal(t, 3, config.MaxRetries)
		assert.Equal(t, 300, config.TimeoutSeconds)
		assert.False(t, config.EnableVerbose)
		assert.NotNil(t, config.CustomPrompts)
		assert.NotNil(t, config.AdditionalConfig)
	})

	t.Run("Valid PlanningConfig", func(t *testing.T) {
		config := &PlanningConfig{
			PlanningAgentLLM: "custom-llm",
			MaxRetries:       5,
			TimeoutSeconds:   600,
			EnableVerbose:    true,
			CustomPrompts:    map[string]string{"test": "prompt"},
			AdditionalConfig: map[string]interface{}{"setting": "value"},
		}

		assert.Equal(t, "custom-llm", config.PlanningAgentLLM)
		assert.Equal(t, 5, config.MaxRetries)
		assert.Equal(t, 600, config.TimeoutSeconds)
		assert.True(t, config.EnableVerbose)
		assert.Equal(t, "prompt", config.CustomPrompts["test"])
		assert.Equal(t, "value", config.AdditionalConfig["setting"])
	})
}

func TestJSONSerialization(t *testing.T) {
	t.Run("Complex PlannerTaskPydanticOutput JSON", func(t *testing.T) {
		original := PlannerTaskPydanticOutput{
			ListOfPlansPerTask: []PlanPerTask{
				{
					Task: "Complex task with special characters: \"quotes\" and 'apostrophes'",
					Plan: "Step 1: Handle special characters\nStep 2: Validate output\nStep 3: Complete task",
				},
				{
					Task: "Another task with numbers 123 and symbols !@#",
					Plan: "1. Initialize\n2. Process data\n3. Return results",
				},
			},
		}

		// 序列化
		jsonStr, err := original.ToJSON()
		require.NoError(t, err)
		assert.True(t, json.Valid([]byte(jsonStr)))

		// 反序列化
		var restored PlannerTaskPydanticOutput
		err = restored.FromJSON(jsonStr)
		require.NoError(t, err)

		// 验证数据完整性
		assert.Equal(t, len(original.ListOfPlansPerTask), len(restored.ListOfPlansPerTask))
		for i := range original.ListOfPlansPerTask {
			assert.Equal(t, original.ListOfPlansPerTask[i].Task, restored.ListOfPlansPerTask[i].Task)
			assert.Equal(t, original.ListOfPlansPerTask[i].Plan, restored.ListOfPlansPerTask[i].Plan)
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		var output PlannerTaskPydanticOutput
		err := output.FromJSON(`{"invalid": json}`)
		assert.Error(t, err)
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("Empty strings", func(t *testing.T) {
		plan := PlanPerTask{Task: "", Plan: ""}
		jsonStr, err := plan.ToJSON()
		require.NoError(t, err)

		var restored PlanPerTask
		err = json.Unmarshal([]byte(jsonStr), &restored)
		require.NoError(t, err)
		assert.Equal(t, "", restored.Task)
		assert.Equal(t, "", restored.Plan)
	})

	t.Run("Very long strings", func(t *testing.T) {
		longString := string(make([]byte, 10000))
		for range longString {
			longString = "a" + longString[1:]
		}

		plan := PlanPerTask{Task: longString, Plan: longString}
		jsonStr, err := plan.ToJSON()
		require.NoError(t, err)

		var restored PlanPerTask
		err = json.Unmarshal([]byte(jsonStr), &restored)
		require.NoError(t, err)
		assert.Equal(t, len(longString), len(restored.Task))
		assert.Equal(t, len(longString), len(restored.Plan))
	})

	t.Run("Unicode characters", func(t *testing.T) {
		plan := PlanPerTask{
			Task: "任务描述 with 中文 and émojis 🚀",
			Plan: "步骤 1: 处理 Unicode\n步骤 2: 验证输出 ✅",
		}

		jsonStr, err := plan.ToJSON()
		require.NoError(t, err)

		var restored PlanPerTask
		err = json.Unmarshal([]byte(jsonStr), &restored)
		require.NoError(t, err)
		assert.Equal(t, plan.Task, restored.Task)
		assert.Equal(t, plan.Plan, restored.Plan)
	})
}
