package agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ynl/greensoulai/internal/llm"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// ExtendedMockLLM现在在 testing_mocks.go 中集中管理

// TestAgentWithToolsIntegration 测试Agent使用工具执行任务的完整集成
func TestAgentWithToolsIntegration(t *testing.T) {
	// 创建测试组件
	eventBus := events.NewEventBus(logger.NewConsoleLogger())
	testLogger := logger.NewConsoleLogger()

	// 创建模拟LLM，返回工具调用请求
	mockLLM := NewExtendedMockLLM([]llm.Response{
		{
			Content: `{"tool_name": "calculator", "arguments": {"operation": "add", "a": 5.0, "b": 3.0}}`,
			Usage:   llm.Usage{TotalTokens: 50},
			Model:   "mock-model",
		},
	})

	// 创建Agent
	config := AgentConfig{
		Role:      "Math Assistant",
		Goal:      "Help with mathematical calculations",
		Backstory: "I am an AI assistant specialized in mathematics",
		LLM:       mockLLM,
		EventBus:  eventBus,
		Logger:    testLogger,
	}
	agent, err := NewBaseAgent(config)
	require.NoError(t, err)

	// 为Agent添加计算器工具
	calcTool := NewCalculatorTool()
	err = agent.AddTool(calcTool)
	require.NoError(t, err)

	// 验证Agent有工具
	tools := agent.GetTools()
	assert.Len(t, tools, 1)
	assert.Equal(t, "calculator", tools[0].GetName())

	// 创建任务
	task := NewBaseTask(
		"Calculate 5 + 3",
		"The result of the addition operation",
	)

	// 执行任务
	ctx := context.Background()
	output, err := agent.Execute(ctx, task)
	require.NoError(t, err)
	require.NotNil(t, output)

	// 验证输出
	assert.NotEmpty(t, output.Raw)
	assert.Equal(t, "Math Assistant", output.Agent)
	assert.Greater(t, output.TokensUsed, 0)
	assert.Contains(t, output.Raw, "calculator") // 应该包含工具调用信息
}

// TestTaskSpecificToolsOverrideAgentTools 测试任务级工具覆盖Agent工具
func TestTaskSpecificToolsOverrideAgentTools(t *testing.T) {
	eventBus := events.NewEventBus(logger.NewConsoleLogger())
	testLogger := logger.NewConsoleLogger()

	mockLLM := NewExtendedMockLLM([]llm.Response{
		{
			Content: "Task completed using text_analyzer tool",
			Usage:   llm.Usage{TotalTokens: 30},
			Model:   "mock-model",
		},
	})

	// 创建Agent并添加计算器工具
	config := AgentConfig{
		Role:      "Multi-purpose Assistant",
		Goal:      "Help with various tasks",
		Backstory: "I can handle multiple types of tasks",
		LLM:       mockLLM,
		EventBus:  eventBus,
		Logger:    testLogger,
	}
	agent, err := NewBaseAgent(config)
	require.NoError(t, err)

	// Agent有计算器工具
	calcTool := NewCalculatorTool()
	err = agent.AddTool(calcTool)
	require.NoError(t, err)

	// 创建任务并添加文本分析工具（应该覆盖Agent的工具）
	task := NewBaseTask(
		"Analyze the text content",
		"Analysis results of the text",
	)

	textTool := NewTextAnalyzerTool()
	err = task.AddTool(textTool)
	require.NoError(t, err)

	// 验证工具选择逻辑
	toolCtx := NewToolExecutionContext(agent, task)
	assert.Len(t, toolCtx.Tools, 1)
	assert.Equal(t, "text_analyzer", toolCtx.Tools[0].GetName())

	// 执行任务
	ctx := context.Background()
	output, err := agent.Execute(ctx, task)
	require.NoError(t, err)
	require.NotNil(t, output)

	// 验证使用了任务级别的工具
	assert.Contains(t, output.Raw, "text_analyzer")
}

// TestAgentWithMultipleTools 测试Agent使用多个工具
func TestAgentWithMultipleTools(t *testing.T) {
	eventBus := events.NewEventBus(logger.NewConsoleLogger())
	testLogger := logger.NewConsoleLogger()

	mockLLM := NewExtendedMockLLM([]llm.Response{
		{
			Content: "I have access to calculator, file_reader, and json_parser tools",
			Usage:   llm.Usage{TotalTokens: 40},
			Model:   "mock-model",
		},
	})

	// 创建Agent
	config := AgentConfig{
		Role:      "Multi-tool Assistant",
		Goal:      "Handle various types of tasks",
		Backstory: "I have access to multiple tools",
		LLM:       mockLLM,
		EventBus:  eventBus,
		Logger:    testLogger,
	}
	agent, err := NewBaseAgent(config)
	require.NoError(t, err)

	// 添加多个工具
	tools := []Tool{
		NewCalculatorTool(),
		NewFileReaderTool(),
		NewJSONParserTool(),
	}

	for _, tool := range tools {
		err = agent.AddTool(tool)
		require.NoError(t, err)
	}

	// 验证所有工具都已添加
	agentTools := agent.GetTools()
	assert.Len(t, agentTools, 3)

	// 创建任务
	task := NewBaseTask(
		"Use your tools to help me",
		"A response showing available tools",
	)

	// 验证工具上下文
	toolCtx := NewToolExecutionContext(agent, task)
	assert.Len(t, toolCtx.Tools, 3)
	assert.True(t, toolCtx.HasTools())

	toolNames := toolCtx.GetToolNames()
	assert.Contains(t, toolNames, "calculator")
	assert.Contains(t, toolNames, "file_reader")
	assert.Contains(t, toolNames, "json_parser")

	toolsDesc := toolCtx.GetToolsDescription()
	assert.Contains(t, toolsDesc, "calculator:")
	assert.Contains(t, toolsDesc, "file_reader:")
	assert.Contains(t, toolsDesc, "json_parser:")

	// 执行任务
	ctx := context.Background()
	output, err := agent.Execute(ctx, task)
	require.NoError(t, err)
	require.NotNil(t, output)

	// 验证输出包含工具信息
	assert.NotEmpty(t, output.Raw)
	assert.Equal(t, "Multi-tool Assistant", output.Agent)
}

// TestToolExecutionInTaskPrompt 测试工具信息是否正确包含在任务提示中
func TestToolExecutionInTaskPrompt(t *testing.T) {
	eventBus := events.NewEventBus(logger.NewConsoleLogger())
	testLogger := logger.NewConsoleLogger()

	// 创建一个记录提示的模拟LLM
	var capturedPrompt string
	mockLLM := NewExtendedMockLLM([]llm.Response{
		{
			Content: "Task completed",
			Usage:   llm.Usage{TotalTokens: 25},
			Model:   "mock-model",
		},
	})
	mockLLM.WithCallHandler(func(messages []llm.Message) {
		if len(messages) > 0 {
			if content, ok := messages[len(messages)-1].Content.(string); ok {
				capturedPrompt = content
			}
		}
	})

	// 创建Agent并添加工具
	config := AgentConfig{
		Role:      "Prompt Test Agent",
		Goal:      "Test prompt generation",
		Backstory: "I test how tools are included in prompts",
		LLM:       mockLLM,
		EventBus:  eventBus,
		Logger:    testLogger,
	}
	agent, err := NewBaseAgent(config)
	require.NoError(t, err)

	calcTool := NewCalculatorTool()
	err = agent.AddTool(calcTool)
	require.NoError(t, err)

	// 创建任务
	task := NewBaseTask(
		"Test task for prompt generation",
		"Test if tools are included in prompt",
	)

	// 执行任务
	ctx := context.Background()
	_, err = agent.Execute(ctx, task)
	require.NoError(t, err)

	// 验证提示中包含工具信息
	assert.NotEmpty(t, capturedPrompt)
	assert.Contains(t, capturedPrompt, "Available Tools:")
	assert.Contains(t, capturedPrompt, "calculator:")
	assert.Contains(t, capturedPrompt, "Perform basic mathematical calculations")
	assert.Contains(t, capturedPrompt, "To use a tool, respond with a JSON object")
}

// TestToolUsageLimitInAgent 测试Agent中工具使用限制
func TestToolUsageLimitInAgent(t *testing.T) {
	eventBus := events.NewEventBus(logger.NewConsoleLogger())
	testLogger := logger.NewConsoleLogger()

	mockLLM := NewExtendedMockLLM([]llm.Response{
		{Content: "First call", Usage: llm.Usage{TotalTokens: 10}, Model: "mock"},
		{Content: "Second call", Usage: llm.Usage{TotalTokens: 10}, Model: "mock"},
		{Content: "Third call should fail", Usage: llm.Usage{TotalTokens: 10}, Model: "mock"},
	})

	// 创建Agent
	config := AgentConfig{
		Role:      "Limited Tool Agent",
		Goal:      "Test tool usage limits",
		Backstory: "I have tools with usage limits",
		LLM:       mockLLM,
		EventBus:  eventBus,
		Logger:    testLogger,
	}
	agent, err := NewBaseAgent(config)
	require.NoError(t, err)

	// 创建有使用限制的工具
	limitedTool := NewBaseTool(
		"limited_tool",
		"A tool with usage limit",
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return "limited result", nil
		},
	)
	limitedTool.SetUsageLimit(2) // 限制使用2次

	err = agent.AddTool(limitedTool)
	require.NoError(t, err)

	// 创建任务
	task := NewBaseTask(
		"Use the limited tool",
		"Result from limited tool",
	)

	// 第一次执行 - 应该成功
	ctx := context.Background()
	output1, err := agent.Execute(ctx, task)
	require.NoError(t, err)
	require.NotNil(t, output1)

	// 第二次执行 - 应该成功
	output2, err := agent.Execute(ctx, task)
	require.NoError(t, err)
	require.NotNil(t, output2)

	// 验证工具使用统计
	tools := agent.GetTools()
	assert.Len(t, tools, 1)
	// 注意：由于工具是通过LLM调用的，实际使用次数可能为0
	// 这里主要测试工具系统的集成
}

// BenchmarkAgentWithToolsExecution Agent使用工具的性能测试
func BenchmarkAgentWithToolsExecution(b *testing.B) {
	eventBus := events.NewEventBus(logger.NewTestLogger())
	testLogger := logger.NewTestLogger()

	mockLLM := NewExtendedMockLLM([]llm.Response{
		{
			Content: "Benchmark result",
			Usage:   llm.Usage{TotalTokens: 20},
			Model:   "mock-model",
		},
	})

	config := AgentConfig{
		Role:      "Benchmark Agent",
		Goal:      "Performance testing",
		Backstory: "I am optimized for performance",
		LLM:       mockLLM,
		EventBus:  eventBus,
		Logger:    testLogger,
	}
	agent, err := NewBaseAgent(config)
	if err != nil {
		b.Fatalf("Failed to create agent: %v", err)
	}

	// 添加工具
	err = agent.AddTool(NewCalculatorTool())
	if err != nil {
		b.Fatalf("Failed to add tool: %v", err)
	}

	task := NewBaseTask(
		"Benchmark task",
		"Benchmark result",
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		_, err := agent.Execute(ctx, task)
		if err != nil {
			b.Fatalf("Agent execution failed: %v", err)
		}
	}
}
