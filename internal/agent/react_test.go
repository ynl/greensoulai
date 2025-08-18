package agent

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentMode_String(t *testing.T) {
	tests := []struct {
		mode     AgentMode
		expected string
	}{
		{ModeJSON, "json"},
		{ModeReAct, "react"},
		{ModeHybrid, "hybrid"},
		{AgentMode(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.mode.String())
		})
	}
}

func TestDefaultReActConfig(t *testing.T) {
	config := DefaultReActConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 10, config.MaxIterations)
	assert.Equal(t, 30*time.Second, config.ThoughtTimeout)
	assert.False(t, config.EnableDebugOutput)
	assert.True(t, config.StrictFormatValidation)
	assert.True(t, config.AllowFallbackToJSON)
	assert.Empty(t, config.CustomPromptTemplate)
}

func TestReActStep(t *testing.T) {
	t.Run("CreateBasicStep", func(t *testing.T) {
		step := &ReActStep{
			StepID:    "test-step-1",
			Thought:   "I need to search for information",
			Action:    "web_search",
			Timestamp: time.Now(),
		}

		assert.Equal(t, "test-step-1", step.StepID)
		assert.Equal(t, "I need to search for information", step.Thought)
		assert.Equal(t, "web_search", step.Action)
		assert.False(t, step.IsComplete)
		assert.Empty(t, step.FinalAnswer)
	})

	t.Run("CreateCompletedStep", func(t *testing.T) {
		step := &ReActStep{
			StepID:      "test-step-2",
			Thought:     "I now have all the information needed",
			FinalAnswer: "The answer is 42",
			IsComplete:  true,
			Timestamp:   time.Now(),
		}

		assert.True(t, step.IsComplete)
		assert.Equal(t, "The answer is 42", step.FinalAnswer)
	})
}

func TestNewReActTrace(t *testing.T) {
	trace := NewReActTrace()

	assert.NotNil(t, trace)
	assert.NotEmpty(t, trace.TraceID)
	assert.True(t, strings.HasPrefix(trace.TraceID, "trace_"))
	assert.Empty(t, trace.Steps)
	assert.False(t, trace.IsCompleted)
	assert.Equal(t, 0, trace.IterationCount)
	assert.NotZero(t, trace.StartTime)
}

func TestReActTrace_AddStep(t *testing.T) {
	trace := NewReActTrace()

	// 添加第一个步骤
	step1 := &ReActStep{
		StepID:  "step-1",
		Thought: "First thought",
		Action:  "search",
	}
	trace.AddStep(step1)

	assert.Len(t, trace.Steps, 1)
	assert.Equal(t, 1, trace.IterationCount)
	assert.False(t, trace.IsCompleted)

	// 添加完成步骤
	step2 := &ReActStep{
		StepID:      "step-2",
		Thought:     "Final thought",
		FinalAnswer: "Final result",
		IsComplete:  true,
	}
	trace.AddStep(step2)

	assert.Len(t, trace.Steps, 2)
	assert.Equal(t, 2, trace.IterationCount)
	assert.True(t, trace.IsCompleted)
	assert.Equal(t, "Final result", trace.FinalOutput)
	assert.NotZero(t, trace.TotalDuration)
}

func TestReActTrace_GetLastStep(t *testing.T) {
	trace := NewReActTrace()

	// 空轨迹
	assert.Nil(t, trace.GetLastStep())

	// 添加步骤
	step := &ReActStep{StepID: "test-step"}
	trace.AddStep(step)

	lastStep := trace.GetLastStep()
	assert.NotNil(t, lastStep)
	assert.Equal(t, "test-step", lastStep.StepID)
}

func TestReActTrace_HasCompletedStep(t *testing.T) {
	trace := NewReActTrace()

	// 无完成步骤
	assert.False(t, trace.HasCompletedStep())

	// 添加非完成步骤
	step1 := &ReActStep{StepID: "step-1", IsComplete: false}
	trace.AddStep(step1)
	assert.False(t, trace.HasCompletedStep())

	// 添加完成步骤
	step2 := &ReActStep{StepID: "step-2", IsComplete: true}
	trace.AddStep(step2)
	assert.True(t, trace.HasCompletedStep())
}

func TestNewStandardReActParser(t *testing.T) {
	parser := NewStandardReActParser()

	assert.NotNil(t, parser)
	assert.NotNil(t, parser.thoughtPattern)
	assert.NotNil(t, parser.actionPattern)
	assert.NotNil(t, parser.actionInputPattern)
	assert.NotNil(t, parser.observationPattern)
	assert.NotNil(t, parser.finalAnswerPattern)
}

func TestStandardReActParser_Parse(t *testing.T) {
	parser := NewStandardReActParser()
	ctx := context.Background()

	t.Run("ParseThoughtOnly", func(t *testing.T) {
		output := "Thought: I need to search for information about the topic"

		step, err := parser.Parse(ctx, output)
		require.NoError(t, err)
		assert.Equal(t, "I need to search for information about the topic", step.Thought)
		assert.Empty(t, step.Action)
		assert.False(t, step.IsComplete)
	})

	t.Run("ParseCompleteAction", func(t *testing.T) {
		output := `Thought: I need to search for information
Action: web_search
Action Input: {"query": "AI technology trends", "limit": 5}`

		step, err := parser.Parse(ctx, output)
		require.NoError(t, err)
		assert.Equal(t, "I need to search for information", step.Thought)
		assert.Equal(t, "web_search", step.Action)
		assert.NotNil(t, step.ActionInput)
		assert.Equal(t, "AI technology trends", step.ActionInput["query"])
		assert.Equal(t, float64(5), step.ActionInput["limit"])
		assert.False(t, step.IsComplete)
	})

	t.Run("ParseFinalAnswer", func(t *testing.T) {
		output := `Thought: I now have all the information I need
Final Answer: AI technology is rapidly evolving with significant advances in LLMs and automation.`

		step, err := parser.Parse(ctx, output)
		require.NoError(t, err)
		assert.Equal(t, "I now have all the information I need", step.Thought)
		assert.Equal(t, "AI technology is rapidly evolving with significant advances in LLMs and automation.", step.FinalAnswer)
		assert.True(t, step.IsComplete)
	})

	t.Run("ParseWithObservation", func(t *testing.T) {
		output := `Thought: Let me analyze the results
Action: analyze_data
Action Input: {"data": "sample"}
Observation: The analysis shows positive trends`

		step, err := parser.Parse(ctx, output)
		require.NoError(t, err)
		assert.Equal(t, "Let me analyze the results", step.Thought)
		assert.Equal(t, "analyze_data", step.Action)
		assert.Equal(t, "The analysis shows positive trends", step.Observation)
	})

	t.Run("ParseInvalidActionInput", func(t *testing.T) {
		output := `Thought: Testing invalid JSON
Action: test_action
Action Input: {invalid json}`

		_, err := parser.Parse(ctx, output)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse action input JSON")
	})

	t.Run("ParseCaseInsensitive", func(t *testing.T) {
		output := `THOUGHT: This should work with any case
ACTION: test_action
ACTION INPUT: {"test": true}
FINAL ANSWER: Case insensitive parsing works`

		step, err := parser.Parse(ctx, output)
		require.NoError(t, err)
		assert.Equal(t, "This should work with any case", step.Thought)
		assert.Equal(t, "Case insensitive parsing works", step.FinalAnswer)
		assert.True(t, step.IsComplete)
	})
}

func TestStandardReActParser_Validate(t *testing.T) {
	parser := NewStandardReActParser()

	t.Run("ValidateNilStep", func(t *testing.T) {
		err := parser.Validate(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "step cannot be nil")
	})

	t.Run("ValidateCompletedStep", func(t *testing.T) {
		// 有效的完成步骤
		step := &ReActStep{
			Thought:     "Final thought",
			FinalAnswer: "Final answer",
			IsComplete:  true,
		}
		err := parser.Validate(step)
		assert.NoError(t, err)

		// 无效的完成步骤（缺少FinalAnswer）
		step.FinalAnswer = ""
		err = parser.Validate(step)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "complete step must have final_answer")
	})

	t.Run("ValidateIncompleteStep", func(t *testing.T) {
		// 无效步骤（缺少Thought）
		step := &ReActStep{
			Action: "test_action",
		}
		err := parser.Validate(step)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "step must have thought")

		// 有效步骤（仅有Thought，无Action）
		step.Thought = "I'm thinking"
		step.Action = "" // 确保Action为空
		err = parser.Validate(step)
		assert.NoError(t, err)

		// 无效步骤（有Action但无ActionInput）
		step.Action = "test_action"
		step.ActionInput = nil
		err = parser.Validate(step)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "action step must have action_input")

		// 有效步骤（有Action和ActionInput）
		step.ActionInput = map[string]interface{}{"test": true}
		err = parser.Validate(step)
		assert.NoError(t, err)
	})
}

func TestStandardReActParser_Format(t *testing.T) {
	parser := NewStandardReActParser()

	t.Run("FormatBasicStep", func(t *testing.T) {
		step := &ReActStep{
			Thought: "I need to search",
		}

		formatted := parser.Format(step)
		assert.Equal(t, "Thought: I need to search", formatted)
	})

	t.Run("FormatActionStep", func(t *testing.T) {
		step := &ReActStep{
			Thought:     "I need to search",
			Action:      "web_search",
			ActionInput: map[string]interface{}{"query": "test"},
		}

		formatted := parser.Format(step)
		expected := `Thought: I need to search
Action: web_search
Action Input: {"query":"test"}`
		assert.Equal(t, expected, formatted)
	})

	t.Run("FormatCompleteStep", func(t *testing.T) {
		step := &ReActStep{
			Thought:     "I have the answer",
			FinalAnswer: "The answer is 42",
			IsComplete:  true,
		}

		formatted := parser.Format(step)
		expected := `Thought: I have the answer
Final Answer: The answer is 42`
		assert.Equal(t, expected, formatted)
	})

	t.Run("FormatWithObservation", func(t *testing.T) {
		step := &ReActStep{
			Thought:     "Analyzing results",
			Action:      "analyze",
			ActionInput: map[string]interface{}{"data": "test"},
			Observation: "Analysis complete",
		}

		formatted := parser.Format(step)
		expected := `Thought: Analyzing results
Action: analyze
Action Input: {"data":"test"}
Observation: Analysis complete`
		assert.Equal(t, expected, formatted)
	})

	t.Run("FormatInvalidActionInput", func(t *testing.T) {
		// 测试无法序列化的ActionInput
		step := &ReActStep{
			Thought: "Testing",
			Action:  "test",
			ActionInput: map[string]interface{}{
				"invalid": make(chan int), // channels无法序列化为JSON
			},
		}

		formatted := parser.Format(step)
		// 应该忽略无法序列化的ActionInput
		expected := `Thought: Testing
Action: test`
		assert.Equal(t, expected, formatted)
	})
}

// 集成测试
func TestReActParser_Integration(t *testing.T) {
	parser := NewStandardReActParser()
	ctx := context.Background()

	t.Run("FullReActFlow", func(t *testing.T) {
		// 模拟完整的ReAct对话流程
		outputs := []string{
			"Thought: I need to search for information about renewable energy",
			`Thought: Let me search for recent developments
Action: web_search
Action Input: {"query": "renewable energy 2024", "limit": 3}`,
			`Thought: Now I have good information, let me analyze the trends
Action: analyze_trends
Action Input: {"data": "solar, wind, battery technology"}`,
			`Thought: Based on my analysis, I can now provide a comprehensive answer
Final Answer: Renewable energy in 2024 shows significant growth in solar and wind sectors, with major improvements in battery storage technology.`,
		}

		trace := NewReActTrace()

		for i, output := range outputs {
			step, err := parser.Parse(ctx, output)
			require.NoError(t, err, "Failed to parse output %d", i)

			err = parser.Validate(step)
			require.NoError(t, err, "Step %d validation failed", i)

			trace.AddStep(step)

			if step.IsComplete {
				break
			}
		}

		assert.True(t, trace.IsCompleted)
		assert.Len(t, trace.Steps, 4)
		assert.Contains(t, trace.FinalOutput, "Renewable energy in 2024")

		// 验证轨迹的完整性
		lastStep := trace.GetLastStep()
		assert.True(t, lastStep.IsComplete)
		assert.NotEmpty(t, lastStep.FinalAnswer)
	})
}

// 性能测试
func BenchmarkReActParser_Parse(b *testing.B) {
	parser := NewStandardReActParser()
	ctx := context.Background()
	output := `Thought: I need to search for information
Action: web_search
Action Input: {"query": "test query", "limit": 5}
Observation: Found 5 results`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(ctx, output)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReActParser_Format(b *testing.B) {
	parser := NewStandardReActParser()
	step := &ReActStep{
		Thought:     "Testing performance",
		Action:      "test_action",
		ActionInput: map[string]interface{}{"test": true, "count": 42},
		Observation: "Performance test completed",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parser.Format(step)
	}
}
