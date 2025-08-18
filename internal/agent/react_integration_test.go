package agent

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ynl/greensoulai/internal/llm"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// TestReActBasicIntegration 测试ReAct基础集成功能
func TestReActBasicIntegration(t *testing.T) {
	// 创建测试agent
	eventBus := events.NewEventBus(logger.NewConsoleLogger())
	testLogger := logger.NewConsoleLogger()

	mockLLM := NewExtendedMockLLM([]llm.Response{
		{
			Content: "Thought: This is a simple test\nFinal Answer: Integration test completed successfully",
			Model:   "test-model",
			Usage:   llm.Usage{TotalTokens: 30},
		},
	})

	config := AgentConfig{
		Role:            "Test Agent",
		Goal:            "Test integration",
		Backstory:       "I am testing ReAct integration",
		LLM:             mockLLM,
		EventBus:        eventBus,
		Logger:          testLogger,
		ExecutionConfig: DefaultExecutionConfig(),
	}

	agent, err := NewBaseAgent(config)
	require.NoError(t, err)

	t.Run("ModeSwitch", func(t *testing.T) {
		// 测试模式切换
		assert.False(t, agent.GetReActMode())

		agent.SetReActMode(true)
		assert.True(t, agent.GetReActMode())
		assert.Equal(t, ModeReAct, agent.GetCurrentMode())

		agent.SetReActMode(false)
		assert.False(t, agent.GetReActMode())
		assert.Equal(t, ModeJSON, agent.GetCurrentMode())
	})

	t.Run("ConfigTest", func(t *testing.T) {
		config := agent.GetReActConfig()
		assert.NotNil(t, config)
		assert.Equal(t, 10, config.MaxIterations)

		newConfig := &ReActConfig{
			MaxIterations:  5,
			ThoughtTimeout: 10 * time.Second,
		}

		err := agent.SetReActConfig(newConfig)
		assert.NoError(t, err)

		retrievedConfig := agent.GetReActConfig()
		assert.Equal(t, 5, retrievedConfig.MaxIterations)
	})

	t.Run("ReActExecution", func(t *testing.T) {
		agent.SetReActMode(true)

		task := NewBaseTask("Integration test", "Test ReAct execution")
		ctx := context.Background()

		output, trace, err := agent.ExecuteWithReAct(ctx, task)

		require.NoError(t, err)
		assert.NotNil(t, output)
		assert.NotNil(t, trace)
		assert.True(t, trace.IsCompleted)
		assert.Equal(t, "Integration test completed successfully", trace.FinalOutput)
		assert.Contains(t, output.Metadata, "mode")
		assert.Equal(t, "react", output.Metadata["mode"])
	})
}

// TestReActParser 测试ReAct解析器
func TestReActParser(t *testing.T) {
	parser := NewStandardReActParser()
	ctx := context.Background()

	t.Run("ParseBasicThought", func(t *testing.T) {
		output := "Thought: I need to think about this"
		step, err := parser.Parse(ctx, output)
		require.NoError(t, err)
		assert.Equal(t, "I need to think about this", step.Thought)
		assert.False(t, step.IsComplete)
	})

	t.Run("ParseFinalAnswer", func(t *testing.T) {
		output := "Thought: I have the answer\nFinal Answer: The answer is 42"
		step, err := parser.Parse(ctx, output)
		require.NoError(t, err)
		assert.Equal(t, "I have the answer", step.Thought)
		assert.Equal(t, "The answer is 42", step.FinalAnswer)
		assert.True(t, step.IsComplete)
	})

	t.Run("ParseWithAction", func(t *testing.T) {
		output := `Thought: I need to use a tool
Action: calculator
Action Input: {"operation": "add", "a": 1, "b": 2}`

		step, err := parser.Parse(ctx, output)
		require.NoError(t, err)
		assert.Equal(t, "I need to use a tool", step.Thought)
		assert.Equal(t, "calculator", step.Action)
		assert.Equal(t, "add", step.ActionInput["operation"])
		assert.Equal(t, float64(1), step.ActionInput["a"])
		assert.Equal(t, float64(2), step.ActionInput["b"])
	})

	t.Run("ValidateStep", func(t *testing.T) {
		validStep := &ReActStep{
			Thought:     "Valid thought",
			FinalAnswer: "Valid answer",
			IsComplete:  true,
		}
		assert.NoError(t, parser.Validate(validStep))

		invalidStep := &ReActStep{
			Action:      "test_action",
			ActionInput: nil,
		}
		assert.Error(t, parser.Validate(invalidStep))
	})

	t.Run("FormatStep", func(t *testing.T) {
		step := &ReActStep{
			Thought:     "Test thought",
			Action:      "test_action",
			ActionInput: map[string]interface{}{"test": "value"},
		}

		formatted := parser.Format(step)
		assert.Contains(t, formatted, "Thought: Test thought")
		assert.Contains(t, formatted, "Action: test_action")
		assert.Contains(t, formatted, "Action Input:")
	})
}

// TestReActTrace 测试ReAct轨迹
func TestReActTrace(t *testing.T) {
	trace := NewReActTrace()

	assert.NotEmpty(t, trace.TraceID)
	assert.Empty(t, trace.Steps)
	assert.False(t, trace.IsCompleted)
	assert.Equal(t, 0, trace.IterationCount)

	step1 := &ReActStep{
		StepID:  "step1",
		Thought: "First step",
	}
	trace.AddStep(step1)

	assert.Len(t, trace.Steps, 1)
	assert.Equal(t, 1, trace.IterationCount)
	assert.False(t, trace.IsCompleted)

	finalStep := &ReActStep{
		StepID:      "step2",
		FinalAnswer: "Done",
		IsComplete:  true,
	}
	trace.AddStep(finalStep)

	assert.Len(t, trace.Steps, 2)
	assert.Equal(t, 2, trace.IterationCount)
	assert.True(t, trace.IsCompleted)
	assert.Equal(t, "Done", trace.FinalOutput)

	lastStep := trace.GetLastStep()
	assert.Equal(t, "step2", lastStep.StepID)

	assert.True(t, trace.HasCompletedStep())
}
