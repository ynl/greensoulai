package planning

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlanValidationError(t *testing.T) {
	t.Run("Create and test PlanValidationError", func(t *testing.T) {
		err := NewPlanValidationError("task", 1, "task cannot be empty")

		assert.Equal(t, "task", err.Field)
		assert.Equal(t, 1, err.Index)
		assert.Equal(t, "task cannot be empty", err.Message)

		expectedMsg := "validation error at index 1, field 'task': task cannot be empty"
		assert.Equal(t, expectedMsg, err.Error())
	})

	t.Run("PlanValidationError implements error interface", func(t *testing.T) {
		var err error = &PlanValidationError{
			Field:   "plan",
			Index:   0,
			Message: "plan is too short",
		}

		assert.Contains(t, err.Error(), "validation error")
		assert.Contains(t, err.Error(), "plan")
		assert.Contains(t, err.Error(), "too short")
	})
}

func TestPlanningExecutionError(t *testing.T) {
	t.Run("Create and test PlanningExecutionError", func(t *testing.T) {
		originalErr := errors.New("original error")
		err := NewPlanningExecutionError("execution", originalErr, 2, 3, "agent-123", "gpt-4o-mini")

		assert.Equal(t, "execution", err.Phase)
		assert.Equal(t, originalErr, err.Cause)
		assert.Equal(t, 2, err.RetryCount)
		assert.Equal(t, 3, err.TaskCount)
		assert.Equal(t, "agent-123", err.AgentID)
		assert.Equal(t, "gpt-4o-mini", err.ModelUsed)

		assert.Contains(t, err.Error(), "planning execution failed")
		assert.Contains(t, err.Error(), "execution")
		assert.Contains(t, err.Error(), "retry 2/3")
	})

	t.Run("Unwrap method", func(t *testing.T) {
		originalErr := errors.New("root cause")
		err := NewPlanningExecutionError("test", originalErr, 1, 2, "", "")

		unwrapped := err.Unwrap()
		assert.Equal(t, originalErr, unwrapped)

		// Test with errors.Is
		assert.True(t, errors.Is(err, originalErr))
	})

	t.Run("Error formatting", func(t *testing.T) {
		originalErr := errors.New("network timeout")
		err := &PlanningExecutionError{
			Phase:      "agent_creation",
			Cause:      originalErr,
			RetryCount: 1,
			TaskCount:  5,
			AgentID:    "agent-456",
			ModelUsed:  "custom-model",
		}

		errorMsg := err.Error()
		assert.Contains(t, errorMsg, "agent_creation")
		assert.Contains(t, errorMsg, "retry 1/3")
		assert.Contains(t, errorMsg, "network timeout")
	})
}

func TestTaskSummaryError(t *testing.T) {
	t.Run("Create and test TaskSummaryError", func(t *testing.T) {
		err := NewTaskSummaryError(2, "task-123", "missing required field")

		assert.Equal(t, 2, err.TaskIndex)
		assert.Equal(t, "task-123", err.TaskID)
		assert.Equal(t, "missing required field", err.Reason)

		expectedMsg := "failed to create summary for task 2 ('task-123'): missing required field"
		assert.Equal(t, expectedMsg, err.Error())
	})

	t.Run("Error with empty task ID", func(t *testing.T) {
		err := &TaskSummaryError{
			TaskIndex: 0,
			TaskID:    "",
			Reason:    "validation failed",
		}

		errorMsg := err.Error()
		assert.Contains(t, errorMsg, "task 0")
		assert.Contains(t, errorMsg, "('')")
		assert.Contains(t, errorMsg, "validation failed")
	})
}

func TestAgentCreationError(t *testing.T) {
	t.Run("Create and test AgentCreationError", func(t *testing.T) {
		err := NewAgentCreationError("Task Execution Planner", "LLM initialization failed")

		assert.Equal(t, "Task Execution Planner", err.AgentRole)
		assert.Equal(t, "LLM initialization failed", err.Reason)

		expectedMsg := "failed to create agent with role 'Task Execution Planner': LLM initialization failed"
		assert.Equal(t, expectedMsg, err.Error())
	})

	t.Run("Error with special characters in role", func(t *testing.T) {
		err := &AgentCreationError{
			AgentRole: "Agent with \"quotes\" and 'apostrophes'",
			Reason:    "configuration error",
		}

		errorMsg := err.Error()
		assert.Contains(t, errorMsg, "Agent with \"quotes\" and 'apostrophes'")
		assert.Contains(t, errorMsg, "configuration error")
	})
}

func TestPredefinedErrors(t *testing.T) {
	t.Run("Test all predefined errors", func(t *testing.T) {
		errors := []error{
			ErrEmptyPlanList,
			ErrInvalidTaskInfo,
			ErrPlanningAgentCreation,
			ErrPlanningTaskCreation,
			ErrPlanningExecution,
			ErrInvalidPlanOutput,
			ErrPlanningTimeout,
			ErrPlanningRetryExceeded,
			ErrInvalidLLMResponse,
		}

		for _, err := range errors {
			assert.NotNil(t, err)
			assert.NotEmpty(t, err.Error())
			assert.NotEqual(t, "", err.Error())
		}
	})

	t.Run("Error uniqueness", func(t *testing.T) {
		// 确保每个预定义错误都有唯一的消息
		errorMessages := map[string]bool{
			ErrEmptyPlanList.Error():         true,
			ErrInvalidTaskInfo.Error():       true,
			ErrPlanningAgentCreation.Error(): true,
			ErrPlanningTaskCreation.Error():  true,
			ErrPlanningExecution.Error():     true,
			ErrInvalidPlanOutput.Error():     true,
			ErrPlanningTimeout.Error():       true,
			ErrPlanningRetryExceeded.Error(): true,
			ErrInvalidLLMResponse.Error():    true,
		}

		// 验证所有错误消息都是唯一的
		assert.Equal(t, 9, len(errorMessages))
	})
}

func TestErrorChaining(t *testing.T) {
	t.Run("Nested error handling", func(t *testing.T) {
		rootErr := errors.New("database connection failed")
		taskErr := NewTaskSummaryError(1, "task-1", rootErr.Error())
		planningErr := NewPlanningExecutionError("tasks_summary", taskErr, 1, 5, "agent-1", "gpt-4")

		// 测试错误信息包含所有层级
		errorMsg := planningErr.Error()
		assert.Contains(t, errorMsg, "tasks_summary")
		assert.Contains(t, errorMsg, "task-1")
		assert.Contains(t, errorMsg, "database connection failed")

		// 测试 Unwrap
		unwrapped := planningErr.Unwrap()
		assert.Equal(t, taskErr, unwrapped)
	})

	t.Run("Error type assertions", func(t *testing.T) {
		originalErr := errors.New("test error")
		planningErr := NewPlanningExecutionError("test", originalErr, 0, 1, "", "")

		// 类型断言
		var execErr *PlanningExecutionError
		assert.True(t, errors.As(planningErr, &execErr))
		assert.Equal(t, "test", execErr.Phase)

		// Is 检查
		assert.True(t, errors.Is(planningErr, originalErr))
	})
}

func TestErrorConstructors(t *testing.T) {
	t.Run("All constructors create valid errors", func(t *testing.T) {
		planValidationErr := NewPlanValidationError("field", 0, "message")
		assert.NotNil(t, planValidationErr)
		assert.IsType(t, &PlanValidationError{}, planValidationErr)

		planningExecErr := NewPlanningExecutionError("phase", errors.New("cause"), 1, 2, "agent", "model")
		assert.NotNil(t, planningExecErr)
		assert.IsType(t, &PlanningExecutionError{}, planningExecErr)

		taskSummaryErr := NewTaskSummaryError(1, "task", "reason")
		assert.NotNil(t, taskSummaryErr)
		assert.IsType(t, &TaskSummaryError{}, taskSummaryErr)

		agentCreationErr := NewAgentCreationError("role", "reason")
		assert.NotNil(t, agentCreationErr)
		assert.IsType(t, &AgentCreationError{}, agentCreationErr)
	})

	t.Run("Constructor parameter validation", func(t *testing.T) {
		// 测试空参数
		planValidationErr := NewPlanValidationError("", -1, "")
		assert.Equal(t, "", planValidationErr.Field)
		assert.Equal(t, -1, planValidationErr.Index)
		assert.Equal(t, "", planValidationErr.Message)

		// 错误消息应该仍然格式正确
		errorMsg := planValidationErr.Error()
		assert.Contains(t, errorMsg, "validation error")
	})
}

func TestErrorBehavior(t *testing.T) {
	t.Run("Error comparison", func(t *testing.T) {
		err1 := NewPlanValidationError("task", 0, "empty")
		err2 := NewPlanValidationError("task", 0, "empty")
		err3 := NewPlanValidationError("plan", 0, "empty")

		// 相同内容的错误应该相等
		assert.Equal(t, err1.Error(), err2.Error())
		// 不同内容的错误应该不相等
		assert.NotEqual(t, err1.Error(), err3.Error())
	})

	t.Run("Error string representation", func(t *testing.T) {
		err := &PlanningExecutionError{
			Phase:      "execution",
			Cause:      errors.New("timeout"),
			RetryCount: 2,
			TaskCount:  3,
			AgentID:    "agent-123",
			ModelUsed:  "gpt-4o-mini",
		}

		str := err.Error()
		// 验证格式化字符串包含所有重要信息
		assert.Contains(t, str, "execution")
		assert.Contains(t, str, "retry 2/3")
		assert.Contains(t, str, "timeout")
	})

	t.Run("Nil error handling", func(t *testing.T) {
		err := NewPlanningExecutionError("test", nil, 0, 1, "", "")

		// 即使原因为nil，错误消息也应该有意义
		errorMsg := err.Error()
		assert.Contains(t, errorMsg, "planning execution failed")
		assert.Contains(t, errorMsg, "test")

		// Unwrap应该返回nil
		assert.Nil(t, err.Unwrap())
	})
}
