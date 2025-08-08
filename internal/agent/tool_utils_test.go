package agent

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTool现在在 testing_mocks.go 中集中管理

// MockAgent现在在 testing_mocks.go 中集中管理

func TestParseTools(t *testing.T) {
	tests := []struct {
		name     string
		input    []Tool
		expected int
	}{
		{
			name:     "empty tools",
			input:    []Tool{},
			expected: 0,
		},
		{
			name:     "nil tools",
			input:    nil,
			expected: 0,
		},
		{
			name: "valid tools",
			input: []Tool{
				NewMockTool("tool1", "description1"),
				NewMockTool("tool2", "description2"),
			},
			expected: 2,
		},
		{
			name: "tools with nil",
			input: []Tool{
				NewMockTool("tool1", "description1"),
				nil,
				NewMockTool("tool2", "description2"),
			},
			expected: 2,
		},
		{
			name: "tools with empty name",
			input: []Tool{
				NewMockTool("tool1", "description1"),
				NewMockTool("", "description2"), // 空名称，应该被过滤
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTools(tt.input)
			assert.Len(t, result, tt.expected)
		})
	}
}

func TestGetToolNames(t *testing.T) {
	tests := []struct {
		name     string
		input    []Tool
		expected string
	}{
		{
			name:     "empty tools",
			input:    []Tool{},
			expected: "",
		},
		{
			name: "single tool",
			input: []Tool{
				NewMockTool("calculator", "A calculator tool"),
			},
			expected: "calculator",
		},
		{
			name: "multiple tools",
			input: []Tool{
				NewMockTool("calculator", "A calculator tool"),
				NewMockTool("file_reader", "A file reader tool"),
				NewMockTool("web_search", "A web search tool"),
			},
			expected: "calculator, file_reader, web_search",
		},
		{
			name: "tools with nil and empty name",
			input: []Tool{
				NewMockTool("calculator", "A calculator tool"),
				nil,
				NewMockTool("", "Empty name tool"),
				NewMockTool("file_reader", "A file reader tool"),
			},
			expected: "calculator, file_reader",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getToolNames(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRenderTextDescriptionAndArgs(t *testing.T) {
	tests := []struct {
		name     string
		input    []Tool
		contains []string
	}{
		{
			name:     "empty tools",
			input:    []Tool{},
			contains: []string{"No tools available."},
		},
		{
			name: "single tool without parameters",
			input: []Tool{
				NewMockTool("calculator", "A simple calculator"),
			},
			contains: []string{"calculator: A simple calculator"},
		},
		{
			name: "multiple tools",
			input: []Tool{
				NewMockTool("calculator", "A simple calculator"),
				NewMockTool("file_reader", "Read files"),
			},
			contains: []string{
				"calculator: A simple calculator",
				"file_reader: Read files",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderTextDescriptionAndArgs(tt.input)
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func TestRenderTextDescriptionWithParameters(t *testing.T) {
	// 创建带参数的工具
	toolWithParams := NewMockTool("file_reader", "Read contents from a file")
	toolWithParams.schema = ToolSchema{
		Name:        "file_reader",
		Description: "Read contents from a file",
		Parameters: map[string]interface{}{
			"properties": map[string]interface{}{
				"filepath": map[string]interface{}{
					"type":        "string",
					"description": "Path to the file to read",
				},
				"encoding": map[string]interface{}{
					"type":        "string",
					"description": "File encoding (optional)",
				},
			},
		},
		Required: []string{"filepath"},
	}

	result := renderTextDescriptionAndArgs([]Tool{toolWithParams})

	// 验证包含基本信息
	assert.Contains(t, result, "file_reader: Read contents from a file")
	assert.Contains(t, result, "Parameters:")
	assert.Contains(t, result, "filepath: Path to the file to read")
	assert.Contains(t, result, "encoding: File encoding (optional)")
	assert.Contains(t, result, "Required: filepath")
}

func TestSelectToolsForTask(t *testing.T) {

	// 创建测试工具
	agentTool1 := NewMockTool("agent_tool1", "Agent tool 1")
	agentTool2 := NewMockTool("agent_tool2", "Agent tool 2")
	taskTool1 := NewMockTool("task_tool1", "Task tool 1")

	tests := []struct {
		name       string
		setupAgent func(Agent)
		setupTask  func(Task)
		expected   []string
	}{
		{
			name: "task has tools - should use task tools",
			setupAgent: func(a Agent) {
				a.AddTool(agentTool1)
				a.AddTool(agentTool2)
			},
			setupTask: func(t Task) {
				t.AddTool(taskTool1)
			},
			expected: []string{"task_tool1"},
		},
		{
			name: "task has no tools - should use agent tools",
			setupAgent: func(a Agent) {
				a.AddTool(agentTool1)
				a.AddTool(agentTool2)
			},
			setupTask: func(t Task) {
				// 不添加任何工具
			},
			expected: []string{"agent_tool1", "agent_tool2"},
		},
		{
			name: "neither has tools - should return empty",
			setupAgent: func(a Agent) {
				// 不添加任何工具
			},
			setupTask: func(t Task) {
				// 不添加任何工具
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重新创建干净的Agent和Task
			agent := NewMockAgent("test-agent", "test goal", "test backstory")
			task := NewBaseTask("test task", "expected output")

			// 设置测试环境
			tt.setupAgent(agent)
			tt.setupTask(task)

			// 执行测试
			result := selectToolsForTask(task, agent)

			// 验证结果
			assert.Len(t, result, len(tt.expected))

			resultNames := make([]string, len(result))
			for i, tool := range result {
				resultNames[i] = tool.GetName()
			}

			for _, expectedName := range tt.expected {
				assert.Contains(t, resultNames, expectedName)
			}
		})
	}
}

func TestPrepareToolsForAgent(t *testing.T) {
	agent := NewMockAgent("test-agent", "test goal", "test backstory")
	task := NewBaseTask("test task", "expected output")

	tools := []Tool{
		NewMockTool("tool1", "Tool 1"),
		NewMockTool("tool2", "Tool 2"),
		nil,                                // 应该被过滤掉
		NewMockTool("", "Empty name tool"), // 应该被过滤掉
	}

	result := prepareToolsForAgent(agent, task, tools)

	// 应该只保留有效的工具
	assert.Len(t, result, 2)
	assert.Equal(t, "tool1", result[0].GetName())
	assert.Equal(t, "tool2", result[1].GetName())
}

func TestValidateToolSchema(t *testing.T) {
	tests := []struct {
		name      string
		tool      Tool
		expectErr bool
	}{
		{
			name:      "nil tool",
			tool:      nil,
			expectErr: true,
		},
		{
			name:      "tool with empty name",
			tool:      NewMockTool("", "description"),
			expectErr: true,
		},
		{
			name:      "tool with empty description",
			tool:      NewMockTool("name", ""),
			expectErr: true,
		},
		{
			name:      "valid tool",
			tool:      NewMockTool("name", "description"),
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateToolSchema(tt.tool)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFindToolByName(t *testing.T) {
	tools := []Tool{
		NewMockTool("calculator", "Calculator tool"),
		NewMockTool("file_reader", "File reader tool"),
		NewMockTool("web_search", "Web search tool"),
	}

	tests := []struct {
		name       string
		searchName string
		expectFind bool
		expectName string
	}{
		{
			name:       "find existing tool",
			searchName: "calculator",
			expectFind: true,
			expectName: "calculator",
		},
		{
			name:       "find non-existing tool",
			searchName: "non_existing",
			expectFind: false,
			expectName: "",
		},
		{
			name:       "find with exact match",
			searchName: "file_reader",
			expectFind: true,
			expectName: "file_reader",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool, found := findToolByName(tools, tt.searchName)
			assert.Equal(t, tt.expectFind, found)
			if found {
				assert.Equal(t, tt.expectName, tool.GetName())
			}
		})
	}
}

func TestFilterValidTools(t *testing.T) {
	tools := []Tool{
		NewMockTool("valid1", "Valid tool 1"),
		NewMockTool("valid2", "Valid tool 2"),
		NewMockTool("", "Invalid tool with empty name"),
		NewMockTool("valid3", ""), // Invalid tool with empty description
		nil,                       // Nil tool
	}

	result := filterValidTools(tools)

	// 应该只保留2个有效工具
	assert.Len(t, result, 2)
	assert.Equal(t, "valid1", result[0].GetName())
	assert.Equal(t, "valid2", result[1].GetName())
}

func TestToolExecutionContext(t *testing.T) {
	agent := NewMockAgent("test-agent", "test goal", "test backstory")
	task := NewBaseTask("test task", "expected output")

	// 为Agent添加工具
	tool1 := NewMockTool("calculator", "Calculator tool")
	tool2 := NewMockTool("file_reader", "File reader tool")
	agent.AddTool(tool1)
	agent.AddTool(tool2)

	// 创建工具执行上下文
	ctx := NewToolExecutionContext(agent, task)

	// 验证上下文
	assert.NotNil(t, ctx)
	assert.Equal(t, agent, ctx.Agent)
	assert.Equal(t, task, ctx.Task)
	assert.True(t, ctx.HasTools())
	assert.Len(t, ctx.Tools, 2)

	// 验证工具名称
	toolNames := ctx.GetToolNames()
	assert.Contains(t, toolNames, "calculator")
	assert.Contains(t, toolNames, "file_reader")

	// 验证工具描述
	toolsDesc := ctx.GetToolsDescription()
	assert.Contains(t, toolsDesc, "calculator: Calculator tool")
	assert.Contains(t, toolsDesc, "file_reader: File reader tool")
}

func TestToolExecutionContextExecuteTool(t *testing.T) {
	agent := NewMockAgent("test-agent", "test goal", "test backstory")
	task := NewBaseTask("test task", "expected output")

	// 创建自定义执行函数的工具
	mockTool := NewMockTool("test_tool", "Test tool").WithExecuteFunc(
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return "test result", nil
		},
	)

	agent.AddTool(mockTool)

	ctx := NewToolExecutionContext(agent, task)

	// 测试执行存在的工具
	result, err := ctx.ExecuteTool(context.Background(), "test_tool", map[string]interface{}{})
	require.NoError(t, err)
	assert.Equal(t, "test result", result)

	// 测试执行不存在的工具
	_, err = ctx.ExecuteTool(context.Background(), "non_existing_tool", map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tool 'non_existing_tool' not found")
}

func TestCreateToolsDescription(t *testing.T) {
	tests := []struct {
		name     string
		tools    []Tool
		expected string
	}{
		{
			name:     "no tools",
			tools:    []Tool{},
			expected: "No tools available for this task.",
		},
		{
			name: "single tool",
			tools: []Tool{
				NewMockTool("calculator", "Simple calculator"),
			},
			expected: "Available tools:\n- calculator: Simple calculator\n",
		},
		{
			name: "multiple tools",
			tools: []Tool{
				NewMockTool("calculator", "Simple calculator"),
				NewMockTool("file_reader", "Read files"),
			},
			expected: "Available tools:\n- calculator: Simple calculator\n- file_reader: Read files\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createToolsDescription(tt.tools)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// 基准测试
func BenchmarkParseTools(b *testing.B) {
	tools := make([]Tool, 100)
	for i := 0; i < 100; i++ {
		tools[i] = NewMockTool(fmt.Sprintf("tool%d", i), fmt.Sprintf("Description %d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseTools(tools)
	}
}

func BenchmarkGetToolNames(b *testing.B) {
	tools := make([]Tool, 100)
	for i := 0; i < 100; i++ {
		tools[i] = NewMockTool(fmt.Sprintf("tool%d", i), fmt.Sprintf("Description %d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getToolNames(tools)
	}
}

func BenchmarkRenderTextDescriptionAndArgs(b *testing.B) {
	tools := make([]Tool, 50)
	for i := 0; i < 50; i++ {
		tools[i] = NewMockTool(fmt.Sprintf("tool%d", i), fmt.Sprintf("Description %d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderTextDescriptionAndArgs(tools)
	}
}
