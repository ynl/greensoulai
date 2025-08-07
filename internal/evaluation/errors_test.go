package evaluation

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvaluationExecutionError(t *testing.T) {
	t.Run("Basic error creation", func(t *testing.T) {
		originalErr := errors.New("original error")
		err := NewEvaluationExecutionError("task_execution", "task_001", "task", 2, 1, originalErr)

		assert.Equal(t, "task_execution", err.Phase)
		assert.Equal(t, "task_001", err.TargetID)
		assert.Equal(t, "task", err.TargetType)
		assert.Equal(t, 2, err.Iteration)
		assert.Equal(t, 1, err.Retry)
		assert.Equal(t, originalErr, err.Err)
	})

	t.Run("Error message formatting", func(t *testing.T) {
		originalErr := errors.New("connection failed")
		err := NewEvaluationExecutionError("llm_call", "agent_001", "agent", 1, 0, originalErr)

		expected := "evaluation execution failed for agent 'agent_001' at phase 'llm_call' (iteration 1, retry 0): connection failed"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("Error unwrapping", func(t *testing.T) {
		originalErr := errors.New("root cause")
		err := NewEvaluationExecutionError("parsing", "task_001", "task", 1, 0, originalErr)

		assert.Equal(t, originalErr, errors.Unwrap(err))
	})
}

func TestEvaluationConfigError(t *testing.T) {
	t.Run("Basic config error", func(t *testing.T) {
		err := NewEvaluationConfigError("timeout", "invalid", "timeout must be positive")

		assert.Equal(t, "timeout", err.Field)
		assert.Equal(t, "invalid", err.Value)
		assert.Equal(t, "timeout must be positive", err.Message)
	})

	t.Run("Error message formatting", func(t *testing.T) {
		err := NewEvaluationConfigError("llm_model", "invalid-model", "unsupported model")

		expected := "evaluation config error for field 'llm_model' with value 'invalid-model': unsupported model"
		assert.Equal(t, expected, err.Error())
	})
}

func TestEvaluatorCreationError(t *testing.T) {
	t.Run("Error with original cause", func(t *testing.T) {
		originalErr := errors.New("LLM initialization failed")
		err := NewEvaluatorCreationError("crew_evaluator", "goal_alignment", "failed to initialize", originalErr)

		assert.Equal(t, "crew_evaluator", err.EvaluatorType)
		assert.Equal(t, "goal_alignment", err.Category)
		assert.Equal(t, "failed to initialize", err.Message)
		assert.Equal(t, originalErr, err.Err)
	})

	t.Run("Error message with original cause", func(t *testing.T) {
		originalErr := errors.New("network error")
		err := NewEvaluatorCreationError("task_evaluator", "quality", "connection failed", originalErr)

		expected := "failed to create task_evaluator evaluator for category 'quality': connection failed: network error"
		assert.Equal(t, expected, err.Error())
		assert.Equal(t, originalErr, errors.Unwrap(err))
	})

	t.Run("Error message without original cause", func(t *testing.T) {
		err := NewEvaluatorCreationError("agent_evaluator", "efficiency", "missing configuration", nil)

		expected := "failed to create agent_evaluator evaluator for category 'efficiency': missing configuration"
		assert.Equal(t, expected, err.Error())
		assert.Nil(t, errors.Unwrap(err))
	})
}

func TestTaskOutputError(t *testing.T) {
	t.Run("Error with task ID", func(t *testing.T) {
		originalErr := errors.New("parsing failed")
		err := NewTaskOutputError("task_123", "json_parsing", "invalid JSON format", originalErr)

		assert.Equal(t, "task_123", err.TaskID)
		assert.Equal(t, "json_parsing", err.Phase)
		assert.Equal(t, "invalid JSON format", err.Message)
		assert.Equal(t, originalErr, err.Err)
	})

	t.Run("Error message with task ID", func(t *testing.T) {
		originalErr := errors.New("syntax error")
		err := NewTaskOutputError("task_456", "validation", "schema mismatch", originalErr)

		expected := "task output error for task 'task_456' at phase 'validation': schema mismatch: syntax error"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("Error message without task ID", func(t *testing.T) {
		originalErr := errors.New("empty content")
		err := NewTaskOutputError("", "content_check", "no content found", originalErr)

		expected := "task output error at phase 'content_check': no content found: empty content"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("Error unwrapping", func(t *testing.T) {
		originalErr := errors.New("root error")
		err := NewTaskOutputError("task_001", "processing", "failed", originalErr)

		assert.Equal(t, originalErr, errors.Unwrap(err))
	})
}

func TestScoreValidationError(t *testing.T) {
	t.Run("Score validation error", func(t *testing.T) {
		err := NewScoreValidationError(15.5, 0.0, 10.0, "task evaluation")

		assert.Equal(t, 15.5, err.Score)
		assert.Equal(t, 0.0, err.MinScore)
		assert.Equal(t, 10.0, err.MaxScore)
		assert.Equal(t, "task evaluation", err.Context)
	})

	t.Run("Error message formatting", func(t *testing.T) {
		err := NewScoreValidationError(-2.0, 0.0, 10.0, "agent metrics")

		expected := "score validation error in agent metrics: score -2.00 is not within valid range [0.00, 10.00]"
		assert.Equal(t, expected, err.Error())
	})
}

func TestLLMResponseError(t *testing.T) {
	t.Run("LLM response error", func(t *testing.T) {
		originalErr := errors.New("timeout")
		err := NewLLMResponseError("gpt-4", "evaluate task", "partial response", "api_call", originalErr)

		assert.Equal(t, "gpt-4", err.Model)
		assert.Equal(t, "evaluate task", err.Request)
		assert.Equal(t, "partial response", err.Response)
		assert.Equal(t, "api_call", err.Phase)
		assert.Equal(t, originalErr, err.Err)
	})

	t.Run("Error message formatting", func(t *testing.T) {
		originalErr := errors.New("rate limited")
		err := NewLLMResponseError("gpt-3.5", "test prompt", "", "rate_check", originalErr)

		expected := "LLM response error from model 'gpt-3.5' at phase 'rate_check': rate limited"
		assert.Equal(t, expected, err.Error())
		assert.Equal(t, originalErr, errors.Unwrap(err))
	})
}

func TestEvaluationSessionError(t *testing.T) {
	t.Run("Session error", func(t *testing.T) {
		originalErr := errors.New("file not found")
		err := NewEvaluationSessionError("session_001", "save", "failed to save session data", originalErr)

		assert.Equal(t, "session_001", err.SessionID)
		assert.Equal(t, "save", err.Operation)
		assert.Equal(t, "failed to save session data", err.Message)
		assert.Equal(t, originalErr, err.Err)
	})

	t.Run("Error message formatting", func(t *testing.T) {
		originalErr := errors.New("permission denied")
		err := NewEvaluationSessionError("session_123", "load", "access denied", originalErr)

		expected := "evaluation session error for session 'session_123' during 'load': access denied: permission denied"
		assert.Equal(t, expected, err.Error())
		assert.Equal(t, originalErr, errors.Unwrap(err))
	})
}

func TestMetricCategoryError(t *testing.T) {
	t.Run("Metric category error", func(t *testing.T) {
		err := NewMetricCategoryError(MetricCategoryGoalAlignment, "invalid configuration")

		assert.Equal(t, MetricCategoryGoalAlignment, err.Category)
		assert.Equal(t, "invalid configuration", err.Message)
	})

	t.Run("Error message formatting", func(t *testing.T) {
		err := NewMetricCategoryError(MetricCategorySemanticQuality, "evaluator not available")

		expected := "metric category error for 'semantic_quality': evaluator not available"
		assert.Equal(t, expected, err.Error())
	})
}

func TestAgentEvaluationError(t *testing.T) {
	t.Run("Agent evaluation error with task", func(t *testing.T) {
		originalErr := errors.New("execution failed")
		err := NewAgentEvaluationError("agent_001", "Analyst", "task_456", "execution", "agent crashed", originalErr)

		assert.Equal(t, "agent_001", err.AgentID)
		assert.Equal(t, "Analyst", err.AgentRole)
		assert.Equal(t, "task_456", err.TaskID)
		assert.Equal(t, "execution", err.Phase)
		assert.Equal(t, "agent crashed", err.Message)
		assert.Equal(t, originalErr, err.Err)
	})

	t.Run("Error message with task", func(t *testing.T) {
		originalErr := errors.New("memory error")
		err := NewAgentEvaluationError("agent_002", "Writer", "task_789", "memory_access", "insufficient memory", originalErr)

		expected := "agent evaluation error for agent 'agent_002' (role: Writer) on task 'task_789' at phase 'memory_access': insufficient memory: memory error"
		assert.Equal(t, expected, err.Error())
		assert.Equal(t, originalErr, errors.Unwrap(err))
	})

	t.Run("Error message without task", func(t *testing.T) {
		originalErr := errors.New("config error")
		err := NewAgentEvaluationError("agent_003", "Reviewer", "", "initialization", "config missing", originalErr)

		expected := "agent evaluation error for agent 'agent_003' (role: Reviewer) at phase 'initialization': config missing: config error"
		assert.Equal(t, expected, err.Error())
	})
}

func TestPredefinedErrors(t *testing.T) {
	t.Run("Predefined errors exist", func(t *testing.T) {
		// Test that all predefined errors are defined
		predefinedErrors := []error{
			ErrInvalidEvaluationConfig,
			ErrMissingEvaluatorLLM,
			ErrInvalidPassingScore,
			ErrInvalidMetricCategory,
			ErrEvaluatorNotFound,
			ErrEvaluatorAlreadyExists,
			ErrInvalidEvaluatorType,
			ErrTaskNotFound,
			ErrTaskOutputEmpty,
			ErrTaskOutputInvalid,
			ErrTaskEvaluationFailed,
			ErrAgentNotFound,
			ErrAgentEvaluationFailed,
			ErrInvalidExecutionTrace,
			ErrCrewNotFound,
			ErrCrewEvaluationFailed,
			ErrNoTasksToEvaluate,
			ErrSessionNotStarted,
			ErrSessionAlreadyStarted,
			ErrSessionAlreadyEnded,
			ErrInvalidSessionID,
			ErrLLMNotAvailable,
			ErrLLMResponseEmpty,
			ErrLLMResponseInvalid,
			ErrLLMCallTimeout,
			ErrInvalidScore,
			ErrMissingRequiredField,
			ErrInvalidDataFormat,
			ErrDataSerializationFailed,
			ErrDataDeserializationFailed,
		}

		for _, err := range predefinedErrors {
			assert.NotNil(t, err)
			assert.NotEmpty(t, err.Error())
		}
	})
}

func TestIsEvaluationError(t *testing.T) {
	t.Run("Evaluation errors", func(t *testing.T) {
		evaluationErrors := []error{
			NewEvaluationExecutionError("test", "id", "type", 1, 0, errors.New("test")),
			NewEvaluationConfigError("field", "value", "message"),
			NewEvaluatorCreationError("type", "category", "message", nil),
			NewTaskOutputError("id", "phase", "message", nil),
			NewScoreValidationError(5.0, 0.0, 10.0, "context"),
			NewLLMResponseError("model", "request", "response", "phase", nil),
			NewEvaluationSessionError("session", "operation", "message", nil),
			NewMetricCategoryError(MetricCategoryGoalAlignment, "message"),
			NewAgentEvaluationError("agent", "role", "task", "phase", "message", nil),
		}

		for _, err := range evaluationErrors {
			assert.True(t, IsEvaluationError(err))
		}
	})

	t.Run("Non-evaluation errors", func(t *testing.T) {
		nonEvaluationErrors := []error{
			errors.New("standard error"),
			nil,
		}

		for _, err := range nonEvaluationErrors {
			assert.False(t, IsEvaluationError(err))
		}
	})
}

func TestGetErrorType(t *testing.T) {
	t.Run("Evaluation error types", func(t *testing.T) {
		testCases := []struct {
			err      error
			expected string
		}{
			{NewEvaluationExecutionError("test", "id", "type", 1, 0, errors.New("test")), "evaluation_execution_error"},
			{NewEvaluationConfigError("field", "value", "message"), "evaluation_config_error"},
			{NewEvaluatorCreationError("type", "category", "message", nil), "evaluator_creation_error"},
			{NewTaskOutputError("id", "phase", "message", nil), "task_output_error"},
			{NewScoreValidationError(5.0, 0.0, 10.0, "context"), "score_validation_error"},
			{NewLLMResponseError("model", "request", "response", "phase", nil), "llm_response_error"},
			{NewEvaluationSessionError("session", "operation", "message", nil), "evaluation_session_error"},
			{NewMetricCategoryError(MetricCategoryGoalAlignment, "message"), "metric_category_error"},
			{NewAgentEvaluationError("agent", "role", "task", "phase", "message", nil), "agent_evaluation_error"},
			{errors.New("standard error"), "unknown_error"},
			{nil, "no_error"},
		}

		for _, tc := range testCases {
			assert.Equal(t, tc.expected, GetErrorType(tc.err))
		}
	})
}

func TestErrorChaining(t *testing.T) {
	t.Run("Error chain with multiple levels", func(t *testing.T) {
		// Create a chain of errors
		rootErr := errors.New("root cause: network timeout")
		llmErr := NewLLMResponseError("gpt-4", "test prompt", "", "api_call", rootErr)
		execErr := NewEvaluationExecutionError("llm_evaluation", "task_001", "task", 1, 2, llmErr)

		// Test error messages contain context
		assert.Contains(t, execErr.Error(), "task_001")
		assert.Contains(t, execErr.Error(), "llm_evaluation")
		assert.Contains(t, execErr.Error(), "iteration 1, retry 2")

		// Test error unwrapping chain
		assert.Equal(t, llmErr, errors.Unwrap(execErr))
		assert.Equal(t, rootErr, errors.Unwrap(errors.Unwrap(execErr)))
	})
}

func TestErrorScenarios(t *testing.T) {
	t.Run("Task evaluation failure scenario", func(t *testing.T) {
		// Simulate a complete task evaluation failure
		networkErr := errors.New("connection refused")
		llmErr := NewLLMResponseError("gpt-4", "Evaluate this task...", "", "network_call", networkErr)
		taskErr := NewTaskOutputError("task_001", "llm_evaluation", "failed to get evaluation", llmErr)
		execErr := NewEvaluationExecutionError("task_evaluation", "task_001", "task", 1, 2, taskErr)

		// Verify error context propagation
		errorMsg := execErr.Error()
		assert.Contains(t, errorMsg, "task_001")
		assert.Contains(t, errorMsg, "task_evaluation")
		assert.Contains(t, errorMsg, "iteration 1, retry 2")

		// Verify error type detection
		assert.True(t, IsEvaluationError(execErr))
		assert.Equal(t, "evaluation_execution_error", GetErrorType(execErr))
	})

	t.Run("Configuration validation failure scenario", func(t *testing.T) {
		// Test various configuration validation errors
		configErrors := []error{
			NewEvaluationConfigError("llm_model", "", "LLM model cannot be empty"),
			NewEvaluationConfigError("timeout", "-5", "timeout cannot be negative"),
			NewEvaluationConfigError("passing_score", "15", "passing score must be between 0 and 10"),
		}

		for _, err := range configErrors {
			assert.True(t, IsEvaluationError(err))
			assert.Equal(t, "evaluation_config_error", GetErrorType(err))
		}
	})
}
