package evaluation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEvaluationStartedEvent(t *testing.T) {
	t.Run("Basic event creation", func(t *testing.T) {
		config := DefaultEvaluationConfig()
		event := NewEvaluationStartedEvent(
			"test_source",
			"task_evaluation",
			"task_001",
			"Test Task",
			"iter_1",
			config,
		)

		assert.Equal(t, EventTypeEvaluationStarted, event.Type)
		assert.Equal(t, "task_evaluation", event.EvaluationType)
		assert.Equal(t, "task_001", event.TargetID)
		assert.Equal(t, "Test Task", event.TargetName)
		assert.Equal(t, "iter_1", event.IterationID)
		assert.Equal(t, config, event.Config)
		assert.NotZero(t, event.Timestamp)
		assert.Equal(t, "test_source", event.Source)

		// Check payload
		payload := event.Payload
		assert.Equal(t, "task_evaluation", payload["evaluation_type"])
		assert.Equal(t, "task_001", payload["target_id"])
		assert.Equal(t, "Test Task", payload["target_name"])
		assert.Equal(t, "iter_1", payload["iteration_id"])
	})
}

func TestEvaluationCompletedEvent(t *testing.T) {
	t.Run("Successful evaluation completion", func(t *testing.T) {
		event := NewEvaluationCompletedEvent(
			"evaluator",
			"crew_evaluation",
			"crew_001",
			"Test Crew",
			"iter_2",
			8.5,
			"A",
			1500.0,
			true,
		)

		assert.Equal(t, EventTypeEvaluationCompleted, event.Type)
		assert.Equal(t, "crew_evaluation", event.EvaluationType)
		assert.Equal(t, "crew_001", event.TargetID)
		assert.Equal(t, "Test Crew", event.TargetName)
		assert.Equal(t, "iter_2", event.IterationID)
		assert.Equal(t, 8.5, event.Score)
		assert.Equal(t, "A", event.Grade)
		assert.Equal(t, 1500.0, event.ExecutionTimeMs)
		assert.True(t, event.Success)

		// Check payload
		payload := event.Payload
		assert.Equal(t, 8.5, payload["score"])
		assert.Equal(t, "A", payload["grade"])
		assert.Equal(t, 1500.0, payload["execution_time_ms"])
		assert.Equal(t, true, payload["success"])
	})
}

func TestEvaluationFailedEvent(t *testing.T) {
	t.Run("Evaluation failure", func(t *testing.T) {
		event := NewEvaluationFailedEvent(
			"evaluator",
			"agent_evaluation",
			"agent_001",
			"Test Agent",
			"iter_1",
			"LLM timeout",
			"llm_call",
			2000.0,
		)

		assert.Equal(t, EventTypeEvaluationFailed, event.Type)
		assert.Equal(t, "agent_evaluation", event.EvaluationType)
		assert.Equal(t, "agent_001", event.TargetID)
		assert.Equal(t, "Test Agent", event.TargetName)
		assert.Equal(t, "iter_1", event.IterationID)
		assert.Equal(t, "LLM timeout", event.Error)
		assert.Equal(t, "llm_call", event.Phase)
		assert.Equal(t, 2000.0, event.ExecutionTimeMs)

		// Check payload
		payload := event.Payload
		assert.Equal(t, "LLM timeout", payload["error"])
		assert.Equal(t, "llm_call", payload["phase"])
		assert.Equal(t, 2000.0, payload["execution_time_ms"])
	})
}

func TestTaskEvaluatedEvent(t *testing.T) {
	t.Run("Task evaluation completion", func(t *testing.T) {
		evaluation := &TaskEvaluationPydanticOutput{Quality: 7.8}

		event := NewTaskEvaluatedEvent(
			"task_evaluator",
			"task_123",
			"Analyze data trends",
			"Data Analyst",
			7.8,
			evaluation,
			1200.0,
			"eval_session_1",
		)

		assert.Equal(t, EventTypeTaskEvaluationCompleted, event.Type)
		assert.Equal(t, "task_123", event.TaskID)
		assert.Equal(t, "Analyze data trends", event.TaskDescription)
		assert.Equal(t, "Data Analyst", event.AgentRole)
		assert.Equal(t, 7.8, event.Score)
		assert.Equal(t, evaluation, event.Evaluation)
		assert.Equal(t, 1200.0, event.ExecutionTimeMs)
		assert.Equal(t, "eval_session_1", event.IterationID)

		// Check payload
		payload := event.Payload
		assert.Equal(t, "task_123", payload["task_id"])
		assert.Equal(t, "Analyze data trends", payload["task_description"])
		assert.Equal(t, "Data Analyst", payload["agent_role"])
		assert.Equal(t, 7.8, payload["score"])
		assert.Equal(t, 1200.0, payload["execution_time_ms"])
		assert.Equal(t, "eval_session_1", payload["iteration_id"])
	})
}

func TestAgentEvaluatedEvent(t *testing.T) {
	t.Run("Agent evaluation completion", func(t *testing.T) {
		result := &AgentEvaluationResult{
			AgentID: "agent_456",
			TaskID:  "task_789",
			Metrics: map[string]*EvaluationScore{
				"goal_alignment": {Score: 8.0, Feedback: "Good alignment"},
				"quality":        {Score: 7.5, Feedback: "High quality"},
			},
			Timestamp: time.Now(),
		}

		event := NewAgentEvaluatedEvent(
			"agent_evaluator",
			"agent_456",
			"Research Analyst",
			"task_789",
			result,
			1800.0,
			"eval_iter_3",
		)

		assert.Equal(t, EventTypeAgentEvaluationCompleted, event.Type)
		assert.Equal(t, "agent_456", event.AgentID)
		assert.Equal(t, "Research Analyst", event.AgentRole)
		assert.Equal(t, "task_789", event.TaskID)
		assert.Equal(t, result, event.Result)
		assert.Equal(t, result.GetAverageScore(), event.AverageScore)
		assert.Equal(t, len(result.Metrics), event.MetricsCount)
		assert.Equal(t, 1800.0, event.ExecutionTimeMs)
		assert.Equal(t, "eval_iter_3", event.IterationID)

		// Check payload
		payload := event.Payload
		assert.Equal(t, "agent_456", payload["agent_id"])
		assert.Equal(t, "Research Analyst", payload["agent_role"])
		assert.Equal(t, "task_789", payload["task_id"])
		assert.Equal(t, result.GetAverageScore(), payload["average_score"])
		assert.Equal(t, len(result.Metrics), payload["metrics_count"])
		assert.Equal(t, 1800.0, payload["execution_time_ms"])
		assert.Equal(t, "eval_iter_3", payload["iteration_id"])
	})
}

func TestCrewTestResultEvent(t *testing.T) {
	t.Run("Crew test result - corresponds to Python version", func(t *testing.T) {
		event := NewCrewTestResultEvent(
			"crew_evaluator",
			8.7,          // quality
			2500.0,       // execution duration
			"gpt-4",      // model
			"Sales Team", // crew name
			3,            // iteration
			"task_001",   // task ID
			"Sales Rep",  // agent role
		)

		assert.Equal(t, EventTypeCrewTestResult, event.Type)
		assert.Equal(t, 8.7, event.Quality)
		assert.Equal(t, 2500.0, event.ExecutionDuration)
		assert.Equal(t, "gpt-4", event.Model)
		assert.Equal(t, "Sales Team", event.CrewName)
		assert.Equal(t, 3, event.Iteration)
		assert.Equal(t, "task_001", event.TaskID)
		assert.Equal(t, "Sales Rep", event.AgentRole)

		// Check payload
		payload := event.Payload
		assert.Equal(t, 8.7, payload["quality"])
		assert.Equal(t, 2500.0, payload["execution_duration"])
		assert.Equal(t, "gpt-4", payload["model"])
		assert.Equal(t, "Sales Team", payload["crew_name"])
		assert.Equal(t, 3, payload["iteration"])
		assert.Equal(t, "task_001", payload["task_id"])
		assert.Equal(t, "Sales Rep", payload["agent_role"])
	})
}

func TestTaskEvaluationStartedEvent(t *testing.T) {
	t.Run("Task evaluation started", func(t *testing.T) {
		event := NewTaskEvaluationStartedEvent(
			"task_evaluator",
			"task_999",
			"Process customer feedback",
			"Customer Service Rep",
			"eval_start_1",
		)

		assert.Equal(t, EventTypeTaskEvaluationStarted, event.Type)
		assert.Equal(t, "task_999", event.TaskID)
		assert.Equal(t, "Process customer feedback", event.TaskDescription)
		assert.Equal(t, "Customer Service Rep", event.AgentRole)
		assert.Equal(t, "eval_start_1", event.IterationID)

		// Check payload
		payload := event.Payload
		assert.Equal(t, "task_999", payload["task_id"])
		assert.Equal(t, "Process customer feedback", payload["task_description"])
		assert.Equal(t, "Customer Service Rep", payload["agent_role"])
		assert.Equal(t, "eval_start_1", payload["iteration_id"])
	})
}

func TestAgentEvaluationStartedEvent(t *testing.T) {
	t.Run("Agent evaluation started", func(t *testing.T) {
		event := NewAgentEvaluationStartedEvent(
			"agent_evaluator",
			"agent_777",
			"Marketing Manager",
			"task_888",
			5,
			"eval_session_5",
		)

		assert.Equal(t, EventTypeAgentEvaluationStarted, event.Type)
		assert.Equal(t, "agent_777", event.AgentID)
		assert.Equal(t, "Marketing Manager", event.AgentRole)
		assert.Equal(t, "task_888", event.TaskID)
		assert.Equal(t, 5, event.Iteration)
		assert.Equal(t, "eval_session_5", event.IterationID)

		// Check payload
		payload := event.Payload
		assert.Equal(t, "agent_777", payload["agent_id"])
		assert.Equal(t, "Marketing Manager", payload["agent_role"])
		assert.Equal(t, "task_888", payload["task_id"])
		assert.Equal(t, 5, payload["iteration"])
		assert.Equal(t, "eval_session_5", payload["iteration_id"])
	})
}

func TestAgentEvaluationCompletedEvent(t *testing.T) {
	t.Run("Agent evaluation completed", func(t *testing.T) {
		score := &EvaluationScore{
			Score:    9.2,
			Feedback: "Exceptional performance",
			Category: "efficiency",
		}

		event := NewAgentEvaluationCompletedEvent(
			"agent_evaluator",
			"agent_555",
			"Product Manager",
			"task_666",
			2,
			"eval_complete_2",
			MetricCategoryEfficiency,
			score,
			950.0,
		)

		assert.Equal(t, EventTypeAgentEvaluationCompleted, event.Type)
		assert.Equal(t, "agent_555", event.AgentID)
		assert.Equal(t, "Product Manager", event.AgentRole)
		assert.Equal(t, "task_666", event.TaskID)
		assert.Equal(t, 2, event.Iteration)
		assert.Equal(t, "eval_complete_2", event.IterationID)
		assert.Equal(t, MetricCategoryEfficiency, event.MetricCategory)
		assert.Equal(t, score, event.Score)
		assert.Equal(t, 950.0, event.ExecutionTimeMs)

		// Check payload
		payload := event.Payload
		assert.Equal(t, "agent_555", payload["agent_id"])
		assert.Equal(t, "Product Manager", payload["agent_role"])
		assert.Equal(t, "task_666", payload["task_id"])
		assert.Equal(t, 2, payload["iteration"])
		assert.Equal(t, "eval_complete_2", payload["iteration_id"])
		assert.Equal(t, "efficiency", payload["metric_category"])
		assert.Equal(t, 950.0, payload["execution_time_ms"])
	})
}

func TestAgentEvaluationFailedEvent(t *testing.T) {
	t.Run("Agent evaluation failed", func(t *testing.T) {
		event := NewAgentEvaluationFailedEvent(
			"agent_evaluator",
			"agent_333",
			"Operations Manager",
			"task_444",
			1,
			"eval_fail_1",
			"Agent timeout during evaluation",
			1500.0,
		)

		assert.Equal(t, EventTypeAgentEvaluationFailed, event.Type)
		assert.Equal(t, "agent_333", event.AgentID)
		assert.Equal(t, "Operations Manager", event.AgentRole)
		assert.Equal(t, "task_444", event.TaskID)
		assert.Equal(t, 1, event.Iteration)
		assert.Equal(t, "eval_fail_1", event.IterationID)
		assert.Equal(t, "Agent timeout during evaluation", event.Error)
		assert.Equal(t, 1500.0, event.ExecutionTimeMs)

		// Check payload
		payload := event.Payload
		assert.Equal(t, "agent_333", payload["agent_id"])
		assert.Equal(t, "Operations Manager", payload["agent_role"])
		assert.Equal(t, "task_444", payload["task_id"])
		assert.Equal(t, 1, payload["iteration"])
		assert.Equal(t, "eval_fail_1", payload["iteration_id"])
		assert.Equal(t, "Agent timeout during evaluation", payload["error"])
		assert.Equal(t, 1500.0, payload["execution_time_ms"])
	})
}

func TestEvaluationSessionStartedEvent(t *testing.T) {
	t.Run("Evaluation session started", func(t *testing.T) {
		config := DefaultEvaluationConfig()
		metadata := map[string]interface{}{
			"session_type": "comprehensive",
			"total_tasks":  10,
		}

		event := NewEvaluationSessionStartedEvent(
			"evaluation_manager",
			"session_abc123",
			"full_crew_evaluation",
			config,
			metadata,
		)

		assert.Equal(t, EventTypeEvaluationSessionStarted, event.Type)
		assert.Equal(t, "session_abc123", event.SessionID)
		assert.Equal(t, "full_crew_evaluation", event.EvaluationType)
		assert.Equal(t, config, event.Config)
		assert.Equal(t, metadata, event.Metadata)

		// Check payload
		payload := event.Payload
		assert.Equal(t, "session_abc123", payload["session_id"])
		assert.Equal(t, "full_crew_evaluation", payload["evaluation_type"])
	})
}

func TestEvaluationSessionCompletedEvent(t *testing.T) {
	t.Run("Evaluation session completed", func(t *testing.T) {
		event := NewEvaluationSessionCompletedEvent(
			"evaluation_manager",
			"session_xyz789",
			"batch_evaluation",
			15,      // total evaluations
			12,      // successful evaluations
			3,       // failed evaluations
			7.6,     // average score
			45000.0, // total execution time
		)

		assert.Equal(t, EventTypeEvaluationSessionCompleted, event.Type)
		assert.Equal(t, "session_xyz789", event.SessionID)
		assert.Equal(t, "batch_evaluation", event.EvaluationType)
		assert.Equal(t, 15, event.TotalEvaluations)
		assert.Equal(t, 12, event.SuccessfulEvaluations)
		assert.Equal(t, 3, event.FailedEvaluations)
		assert.Equal(t, 7.6, event.AverageScore)
		assert.Equal(t, 45000.0, event.TotalExecutionTimeMs)
		assert.Equal(t, 80.0, event.SuccessRate) // 12/15 * 100

		// Check payload
		payload := event.Payload
		assert.Equal(t, "session_xyz789", payload["session_id"])
		assert.Equal(t, "batch_evaluation", payload["evaluation_type"])
		assert.Equal(t, 15, payload["total_evaluations"])
		assert.Equal(t, 12, payload["successful_evaluations"])
		assert.Equal(t, 3, payload["failed_evaluations"])
		assert.Equal(t, 7.6, payload["average_score"])
		assert.Equal(t, 45000.0, payload["total_execution_time_ms"])
		assert.Equal(t, 80.0, payload["success_rate"])
	})

	t.Run("Zero evaluations edge case", func(t *testing.T) {
		event := NewEvaluationSessionCompletedEvent(
			"evaluation_manager",
			"empty_session",
			"empty_evaluation",
			0, 0, 0, 0.0, 0.0,
		)

		assert.Equal(t, 0.0, event.SuccessRate)
		assert.Equal(t, 0.0, event.Payload["success_rate"])
	})
}

func TestEventTypes(t *testing.T) {
	t.Run("Event type constants", func(t *testing.T) {
		expectedTypes := map[string]string{
			EventTypeEvaluationStarted:          "evaluation.started",
			EventTypeEvaluationCompleted:        "evaluation.completed",
			EventTypeEvaluationFailed:           "evaluation.failed",
			EventTypeTaskEvaluationStarted:      "evaluation.task.started",
			EventTypeTaskEvaluationCompleted:    "evaluation.task.completed",
			EventTypeTaskEvaluationFailed:       "evaluation.task.failed",
			EventTypeAgentEvaluationStarted:     "evaluation.agent.started",
			EventTypeAgentEvaluationCompleted:   "evaluation.agent.completed",
			EventTypeAgentEvaluationFailed:      "evaluation.agent.failed",
			EventTypeCrewEvaluationStarted:      "evaluation.crew.started",
			EventTypeCrewEvaluationCompleted:    "evaluation.crew.completed",
			EventTypeCrewEvaluationFailed:       "evaluation.crew.failed",
			EventTypeEvaluationSessionStarted:   "evaluation.session.started",
			EventTypeEvaluationSessionCompleted: "evaluation.session.completed",
			EventTypeEvaluationSessionFailed:    "evaluation.session.failed",
			EventTypeCrewTestResult:             "evaluation.crew.test.result",
		}

		for constant, expected := range expectedTypes {
			assert.Equal(t, expected, constant)
		}
	})
}

func TestEventTimestamps(t *testing.T) {
	t.Run("Event timestamps are recent", func(t *testing.T) {
		before := time.Now()

		event := NewEvaluationStartedEvent(
			"test",
			"task_evaluation",
			"task_001",
			"Test Task",
			"iter_1",
			nil,
		)

		after := time.Now()

		assert.True(t, event.Timestamp.After(before.Add(-time.Second)))
		assert.True(t, event.Timestamp.Before(after.Add(time.Second)))
	})
}

func TestEventPayloadConsistency(t *testing.T) {
	t.Run("Event payload matches struct fields", func(t *testing.T) {
		// Test that event payloads contain the expected structured data
		event := NewEvaluationCompletedEvent(
			"evaluator",
			"test_type",
			"test_id",
			"test_name",
			"test_iter",
			8.5,
			"B+",
			1000.0,
			true,
		)

		// Payload should match struct fields
		assert.Equal(t, event.EvaluationType, event.Payload["evaluation_type"])
		assert.Equal(t, event.TargetID, event.Payload["target_id"])
		assert.Equal(t, event.TargetName, event.Payload["target_name"])
		assert.Equal(t, event.IterationID, event.Payload["iteration_id"])
		assert.Equal(t, event.Score, event.Payload["score"])
		assert.Equal(t, event.Grade, event.Payload["grade"])
		assert.Equal(t, event.ExecutionTimeMs, event.Payload["execution_time_ms"])
		assert.Equal(t, event.Success, event.Payload["success"])
	})
}
