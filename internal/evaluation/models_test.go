package evaluation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluationScore(t *testing.T) {
	t.Run("Valid score", func(t *testing.T) {
		score := &EvaluationScore{
			Score:    8.5,
			Feedback: "Excellent work with minor improvements needed",
			Category: "quality",
			Criteria: "completion, accuracy, creativity",
		}

		assert.Equal(t, 8.5, score.Score)
		assert.Equal(t, "Excellent work with minor improvements needed", score.Feedback)
		assert.True(t, score.IsPass())
		assert.Equal(t, "Score: 8.5/10 - Excellent work with minor improvements needed", score.String())
	})

	t.Run("Failing score", func(t *testing.T) {
		score := &EvaluationScore{
			Score:    4.0,
			Feedback: "Needs significant improvement",
		}

		assert.False(t, score.IsPass())
	})

	t.Run("Edge case - exactly passing", func(t *testing.T) {
		score := &EvaluationScore{Score: 6.0, Feedback: "Just passing"}
		assert.True(t, score.IsPass())
	})

	t.Run("Edge case - just failing", func(t *testing.T) {
		score := &EvaluationScore{Score: 5.9, Feedback: "Just failing"}
		assert.False(t, score.IsPass())
	})
}

func TestTaskEvaluationPydanticOutput(t *testing.T) {
	t.Run("JSON serialization", func(t *testing.T) {
		output := &TaskEvaluationPydanticOutput{Quality: 7.5}

		jsonStr, err := output.ToJSON()
		require.NoError(t, err)
		assert.Contains(t, jsonStr, "\"quality\":7.5")

		// Test deserialization
		newOutput := &TaskEvaluationPydanticOutput{}
		err = newOutput.FromJSON(jsonStr)
		require.NoError(t, err)
		assert.Equal(t, 7.5, newOutput.Quality)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		output := &TaskEvaluationPydanticOutput{}
		err := output.FromJSON("invalid json")
		assert.Error(t, err)
	})
}

func TestTaskEvaluation(t *testing.T) {
	t.Run("Overall score calculation", func(t *testing.T) {
		evaluation := &TaskEvaluation{
			CompletionScore:  8.0,
			QualityScore:     7.5,
			PerformanceScore: 9.0,
		}

		expectedScore := (8.0 + 7.5 + 9.0) / 3.0
		assert.Equal(t, expectedScore, evaluation.GetOverallScore())
	})

	t.Run("Overall score with explicit score", func(t *testing.T) {
		evaluation := &TaskEvaluation{
			Score:            7.8,
			CompletionScore:  8.0,
			QualityScore:     7.5,
			PerformanceScore: 9.0,
		}

		assert.Equal(t, 7.8, evaluation.GetOverallScore())
	})

	t.Run("Grade calculation", func(t *testing.T) {
		testCases := []struct {
			score    float64
			expected string
		}{
			{9.5, "A+"},
			{8.5, "A"},
			{7.5, "B"},
			{6.5, "C"},
			{5.5, "D"},
			{4.5, "F"},
		}

		for _, tc := range testCases {
			evaluation := &TaskEvaluation{Score: tc.score}
			assert.Equal(t, tc.expected, evaluation.GetGrade())
		}
	})
}

func TestEntityExtraction(t *testing.T) {
	t.Run("Valid entity with relationships", func(t *testing.T) {
		entity := EntityExtraction{
			Name:        "Customer",
			Type:        "Business Entity",
			Description: "A customer entity in the system",
			Confidence:  0.95,
			Relationships: []EntityRelationship{
				{
					RelationType: "owns",
					TargetEntity: "Account",
					Confidence:   0.90,
				},
			},
		}

		assert.Equal(t, "Customer", entity.Name)
		assert.Equal(t, "Business Entity", entity.Type)
		assert.Equal(t, 0.95, entity.Confidence)
		assert.Len(t, entity.Relationships, 1)
		assert.Equal(t, "owns", entity.Relationships[0].RelationType)
	})
}

func TestTrainingTaskEvaluation(t *testing.T) {
	t.Run("Complete training evaluation", func(t *testing.T) {
		evaluation := &TrainingTaskEvaluation{
			TaskID:          "training_001",
			AgentRole:       "Data Analyst",
			TaskDescription: "Analyze sales data",
			ExpectedOutput:  "Summary report with insights",
			ActualOutput:    "Detailed analysis with recommendations",
			Score:           8.5,
			Feedback:        "Good analysis with actionable insights",
			Improvements:    []string{"Add more visualizations", "Include trend analysis"},
			ExecutionTimeMs: 2500.0,
			Iteration:       3,
			ModelUsed:       "gpt-4",
			TokensUsed:      1500,
		}

		assert.Equal(t, "training_001", evaluation.TaskID)
		assert.Equal(t, "Data Analyst", evaluation.AgentRole)
		assert.Equal(t, 8.5, evaluation.Score)
		assert.Len(t, evaluation.Improvements, 2)
		assert.Equal(t, 3, evaluation.Iteration)
		assert.Equal(t, "gpt-4", evaluation.ModelUsed)
	})
}

func TestAgentEvaluationResult(t *testing.T) {
	t.Run("Average score calculation", func(t *testing.T) {
		result := &AgentEvaluationResult{
			AgentID: "agent_001",
			TaskID:  "task_001",
			Metrics: map[string]*EvaluationScore{
				"goal_alignment": {Score: 8.0, Feedback: "Good alignment"},
				"quality":        {Score: 7.5, Feedback: "High quality"},
				"efficiency":     {Score: 9.0, Feedback: "Very efficient"},
			},
		}

		expectedAvg := (8.0 + 7.5 + 9.0) / 3.0
		assert.Equal(t, expectedAvg, result.GetAverageScore())
	})

	t.Run("Empty metrics", func(t *testing.T) {
		result := &AgentEvaluationResult{
			AgentID: "agent_001",
			Metrics: make(map[string]*EvaluationScore),
		}

		assert.Equal(t, 0.0, result.GetAverageScore())
	})
}

func TestCrewEvaluationResult(t *testing.T) {
	t.Run("Statistics calculation", func(t *testing.T) {
		result := &CrewEvaluationResult{
			CrewName: "Test Crew",
			TasksScores: map[int][]float64{
				1: {8.0, 7.5, 6.0},
				2: {9.0, 8.5, 7.0},
			},
			ExecutionTimes: map[int][]float64{
				1: {1000, 1500, 2000},
				2: {800, 1200, 1800},
			},
		}

		result.CalculateStats()

		assert.Equal(t, 6, result.TotalTasks)
		assert.Equal(t, 6, result.PassedTasks) // All 6 tasks >= 6.0 (8.0, 7.5, 6.0, 9.0, 8.5, 7.0)
		assert.Equal(t, 7.666666666666667, result.AverageScore)
		assert.Equal(t, 1383.3333333333333, result.AverageTime)
		assert.Equal(t, 100.0, result.SuccessRate) // 6/6 * 100 = 100%
	})

	t.Run("Performance grade", func(t *testing.T) {
		testCases := []struct {
			avgScore float64
			expected string
		}{
			{9.5, "Excellent"},
			{8.5, "Very Good"},
			{7.5, "Good"},
			{6.5, "Satisfactory"},
			{5.5, "Needs Improvement"},
			{4.0, "Poor"},
		}

		for _, tc := range testCases {
			result := &CrewEvaluationResult{AverageScore: tc.avgScore}
			assert.Equal(t, tc.expected, result.GetPerformanceGrade())
		}
	})

	t.Run("Empty results", func(t *testing.T) {
		result := &CrewEvaluationResult{
			TasksScores:    make(map[int][]float64),
			ExecutionTimes: make(map[int][]float64),
		}

		result.CalculateStats()

		assert.Equal(t, 0, result.TotalTasks)
		assert.Equal(t, 0, result.PassedTasks)
		assert.Equal(t, 0.0, result.AverageScore)
		assert.Equal(t, 0.0, result.AverageTime)
		assert.Equal(t, 0.0, result.SuccessRate)
	})
}

func TestMetricCategory(t *testing.T) {
	t.Run("Valid categories", func(t *testing.T) {
		validCategories := []MetricCategory{
			MetricCategoryGoalAlignment,
			MetricCategorySemanticQuality,
			MetricCategoryTaskCompletion,
			MetricCategoryEfficiency,
			MetricCategoryAccuracy,
			MetricCategoryCreativity,
			MetricCategoryCoherence,
			MetricCategoryRelevance,
		}

		for _, category := range validCategories {
			assert.True(t, category.IsValid())
			assert.NotEmpty(t, category.String())
		}
	})

	t.Run("Invalid category", func(t *testing.T) {
		invalidCategory := MetricCategory("invalid_category")
		assert.False(t, invalidCategory.IsValid())
	})
}

func TestEvaluationConfig(t *testing.T) {
	t.Run("Default config", func(t *testing.T) {
		config := DefaultEvaluationConfig()

		assert.Equal(t, "gpt-4o-mini", config.EvaluatorLLM)
		assert.False(t, config.EnableVerbose)
		assert.Equal(t, 3, config.MaxRetries)
		assert.Equal(t, 60, config.TimeoutSeconds)
		assert.Equal(t, 6.0, config.PassingScore)
		assert.Contains(t, config.Categories, MetricCategoryGoalAlignment)
		assert.Contains(t, config.Categories, MetricCategorySemanticQuality)
		assert.Contains(t, config.Categories, MetricCategoryTaskCompletion)
		assert.NotNil(t, config.CustomCriteria)
		assert.NotNil(t, config.Metadata)
	})

	t.Run("Config validation - valid", func(t *testing.T) {
		config := &EvaluationConfig{
			EvaluatorLLM:   "gpt-4",
			MaxRetries:     5,
			TimeoutSeconds: 120,
			PassingScore:   7.0,
			Categories: []MetricCategory{
				MetricCategoryGoalAlignment,
				MetricCategoryAccuracy,
			},
		}

		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("Config validation - invalid LLM", func(t *testing.T) {
		config := &EvaluationConfig{EvaluatorLLM: ""}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "evaluator LLM cannot be empty")
	})

	t.Run("Config validation - negative retries", func(t *testing.T) {
		config := &EvaluationConfig{
			EvaluatorLLM: "gpt-4",
			MaxRetries:   -1,
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max retries cannot be negative")
	})

	t.Run("Config validation - invalid passing score", func(t *testing.T) {
		config := &EvaluationConfig{
			EvaluatorLLM: "gpt-4",
			PassingScore: 15.0,
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "passing score must be between 0 and 10")
	})

	t.Run("Config validation - invalid category", func(t *testing.T) {
		config := &EvaluationConfig{
			EvaluatorLLM: "gpt-4",
			PassingScore: 6.0,
			Categories:   []MetricCategory{MetricCategory("invalid")},
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid metric category")
	})
}

func TestEvaluationModels_EdgeCases(t *testing.T) {
	t.Run("TaskEvaluation with zero scores", func(t *testing.T) {
		evaluation := &TaskEvaluation{}
		assert.Equal(t, 0.0, evaluation.GetOverallScore())
		assert.Equal(t, "F", evaluation.GetGrade())
	})

	t.Run("CrewEvaluationResult with single task", func(t *testing.T) {
		result := &CrewEvaluationResult{
			TasksScores: map[int][]float64{
				1: {8.5},
			},
			ExecutionTimes: map[int][]float64{
				1: {1500},
			},
		}

		result.CalculateStats()

		assert.Equal(t, 1, result.TotalTasks)
		assert.Equal(t, 1, result.PassedTasks)
		assert.Equal(t, 8.5, result.AverageScore)
		assert.Equal(t, 1500.0, result.AverageTime)
		assert.Equal(t, 100.0, result.SuccessRate)
	})

	t.Run("EvaluationScore with extreme values", func(t *testing.T) {
		// Test minimum
		minScore := &EvaluationScore{Score: 0.0, Feedback: "Minimum"}
		assert.False(t, minScore.IsPass())
		assert.Equal(t, "Score: 0.0/10 - Minimum", minScore.String())

		// Test maximum
		maxScore := &EvaluationScore{Score: 10.0, Feedback: "Perfect"}
		assert.True(t, maxScore.IsPass())
		assert.Equal(t, "Score: 10.0/10 - Perfect", maxScore.String())
	})
}
