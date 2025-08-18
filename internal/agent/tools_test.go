package agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBaseTool 测试基础工具功能
func TestBaseTool(t *testing.T) {
	// 创建一个简单的工具
	tool := NewBaseTool(
		"test_tool",
		"A test tool",
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return "test result", nil
		},
	)

	// 测试基本属性
	assert.Equal(t, "test_tool", tool.GetName())
	assert.Equal(t, "A test tool", tool.GetDescription())
	assert.Equal(t, 0, tool.GetUsageCount())
	assert.Equal(t, -1, tool.GetUsageLimit())
	assert.False(t, tool.IsUsageLimitExceeded())

	// 测试执行
	result, err := tool.Execute(context.Background(), map[string]interface{}{})
	require.NoError(t, err)
	assert.Equal(t, "test result", result)
	assert.Equal(t, 1, tool.GetUsageCount())
}

// TestToolWithUsageLimit 测试工具使用限制
func TestToolWithUsageLimit(t *testing.T) {
	tool := NewBaseTool(
		"limited_tool",
		"A tool with usage limit",
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return "limited result", nil
		},
	)

	// 设置使用限制
	tool.SetUsageLimit(2)
	assert.Equal(t, 2, tool.GetUsageLimit())

	// 第一次执行
	result, err := tool.Execute(context.Background(), map[string]interface{}{})
	require.NoError(t, err)
	assert.Equal(t, "limited result", result)
	assert.Equal(t, 1, tool.GetUsageCount())
	assert.False(t, tool.IsUsageLimitExceeded())

	// 第二次执行
	result, err = tool.Execute(context.Background(), map[string]interface{}{})
	require.NoError(t, err)
	assert.Equal(t, "limited result", result)
	assert.Equal(t, 2, tool.GetUsageCount())
	assert.False(t, tool.IsUsageLimitExceeded())

	// 第三次执行应该失败
	_, err = tool.Execute(context.Background(), map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tool usage limit exceeded")
	assert.Equal(t, 3, tool.GetUsageCount())
	assert.True(t, tool.IsUsageLimitExceeded())

	// 重置使用统计
	tool.ResetUsage()
	assert.Equal(t, 0, tool.GetUsageCount())
	assert.False(t, tool.IsUsageLimitExceeded())
}

// TestToolAsyncExecution 测试工具异步执行
func TestToolAsyncExecution(t *testing.T) {
	tool := NewBaseTool(
		"async_tool",
		"An async test tool",
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return "async result", nil
		},
	)

	// 异步执行
	resultChan, err := tool.ExecuteAsync(context.Background(), map[string]interface{}{})
	require.NoError(t, err)

	// 等待结果
	result := <-resultChan
	require.NoError(t, result.Error)
	assert.Equal(t, "async result", result.Output)
	// 执行时间应该大于等于0（在高性能机器上可能为0）
	assert.GreaterOrEqual(t, result.Duration.Nanoseconds(), int64(0))
}

// TestPrebuiltTools 测试预构建的工具
func TestCalculatorTool(t *testing.T) {
	calc := NewCalculatorTool()

	// 测试加法
	result, err := calc.Execute(context.Background(), map[string]interface{}{
		"operation": "add",
		"a":         5.0,
		"b":         3.0,
	})
	require.NoError(t, err)
	assert.Equal(t, 8.0, result)

	// 测试除法
	result, err = calc.Execute(context.Background(), map[string]interface{}{
		"operation": "divide",
		"a":         10.0,
		"b":         2.0,
	})
	require.NoError(t, err)
	assert.Equal(t, 5.0, result)

	// 测试除零错误
	_, err = calc.Execute(context.Background(), map[string]interface{}{
		"operation": "divide",
		"a":         10.0,
		"b":         0.0,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "division by zero")
}

// TestToolCollection 测试工具集合
func TestToolCollection(t *testing.T) {
	collection := NewToolCollection()

	// 测试空集合
	assert.Equal(t, 0, collection.Size())
	assert.False(t, collection.Has("test_tool"))

	// 添加工具
	tool1 := NewCalculatorTool()
	err := collection.Add(tool1)
	require.NoError(t, err)
	assert.Equal(t, 1, collection.Size())
	assert.True(t, collection.Has("calculator"))

	// 测试重复添加
	err = collection.Add(tool1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// 获取工具
	retrieved, exists := collection.Get("calculator")
	assert.True(t, exists)
	assert.Equal(t, tool1.GetName(), retrieved.GetName())

	// 执行工具
	result, err := collection.Execute(context.Background(), "calculator", map[string]interface{}{
		"operation": "add",
		"a":         1.0,
		"b":         2.0,
	})
	require.NoError(t, err)
	assert.Equal(t, 3.0, result)

	// 测试不存在的工具
	_, err = collection.Execute(context.Background(), "nonexistent", map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// 移除工具
	removed := collection.Remove("calculator")
	assert.True(t, removed)
	assert.Equal(t, 0, collection.Size())
	assert.False(t, collection.Has("calculator"))
}

// TestToolUtilityFunctions 测试工具实用函数
func TestParseToolsFunction(t *testing.T) {
	// 创建测试工具
	tools := []Tool{
		NewCalculatorTool(),
		NewFileReaderTool(),
		nil, // 应该被过滤掉
	}

	// 解析工具
	parsed := parseTools(tools)
	assert.Len(t, parsed, 2)
	assert.Equal(t, "calculator", parsed[0].GetName())
	assert.Equal(t, "file_reader", parsed[1].GetName())
}

func TestGetToolNamesFunction(t *testing.T) {
	tools := []Tool{
		NewCalculatorTool(),
		NewFileReaderTool(),
		NewJSONParserTool(),
	}

	names := getToolNames(tools)
	expected := "calculator, file_reader, json_parser"
	assert.Equal(t, expected, names)
}

func TestRenderToolDescriptionFunction(t *testing.T) {
	tools := []Tool{
		NewCalculatorTool(),
		NewFileReaderTool(),
	}

	description := renderTextDescriptionAndArgs(tools)
	assert.Contains(t, description, "calculator:")
	assert.Contains(t, description, "file_reader:")
	assert.Contains(t, description, "Perform basic mathematical calculations")
	assert.Contains(t, description, "Read contents from a file")
}

// TestLoadBasicTools 测试加载基础工具
func TestLoadBasicTools(t *testing.T) {
	collection := NewToolCollection()

	err := collection.LoadBasicTools()
	require.NoError(t, err)

	// 验证基础工具已加载
	expectedTools := []string{
		"calculator",
		"file_reader",
		"file_writer",
		"json_parser",
		"json_formatter",
		"text_analyzer",
	}

	assert.Equal(t, len(expectedTools), collection.Size())

	for _, toolName := range expectedTools {
		assert.True(t, collection.Has(toolName), "Tool %s should exist", toolName)
	}
}

// BenchmarkToolExecution 工具执行性能测试
func BenchmarkToolExecution(b *testing.B) {
	tool := NewCalculatorTool()
	args := map[string]interface{}{
		"operation": "add",
		"a":         1.0,
		"b":         2.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tool.Execute(context.Background(), args)
		if err != nil {
			b.Fatalf("Tool execution failed: %v", err)
		}
	}
}

func BenchmarkToolCollectionExecution(b *testing.B) {
	collection := NewToolCollection()
	err := collection.LoadBasicTools()
	if err != nil {
		b.Fatalf("Failed to load basic tools: %v", err)
	}

	args := map[string]interface{}{
		"operation": "multiply",
		"a":         5.0,
		"b":         6.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := collection.Execute(context.Background(), "calculator", args)
		if err != nil {
			b.Fatalf("Tool execution failed: %v", err)
		}
	}
}
