package agent

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewBaseTool(t *testing.T) {
	handler := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		return "test result", nil
	}

	tool := NewBaseTool("test_tool", "Test tool description", handler)

	if tool.GetName() != "test_tool" {
		t.Errorf("expected name 'test_tool', got %s", tool.GetName())
	}

	if tool.GetDescription() != "Test tool description" {
		t.Errorf("expected description 'Test tool description', got %s", tool.GetDescription())
	}

	schema := tool.GetSchema()
	if schema.Name != "test_tool" {
		t.Errorf("expected schema name 'test_tool', got %s", schema.Name)
	}

	if tool.GetUsageCount() != 0 {
		t.Errorf("expected initial usage count 0, got %d", tool.GetUsageCount())
	}

	if tool.GetUsageLimit() != -1 {
		t.Errorf("expected default usage limit -1, got %d", tool.GetUsageLimit())
	}

	if tool.IsUsageLimitExceeded() {
		t.Error("expected usage limit not exceeded initially")
	}
}

func TestBaseTool_Execute(t *testing.T) {
	handler := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		name, exists := args["name"]
		if !exists {
			return nil, errors.New("name argument required")
		}
		return "Hello, " + name.(string), nil
	}

	tool := NewBaseTool("greeting", "Greeting tool", handler)
	ctx := context.Background()

	// 测试成功执行
	args := map[string]interface{}{
		"name": "World",
	}

	result, err := tool.Execute(ctx, args)
	if err != nil {
		t.Errorf("execution failed: %v", err)
	}

	expected := "Hello, World"
	if result != expected {
		t.Errorf("expected result %s, got %s", expected, result)
	}

	// 验证使用计数增加
	if tool.GetUsageCount() != 1 {
		t.Errorf("expected usage count 1, got %d", tool.GetUsageCount())
	}

	// 测试错误情况
	_, err = tool.Execute(ctx, map[string]interface{}{})
	if err == nil {
		t.Error("expected error for missing argument")
	}

	// 验证使用计数继续增加
	if tool.GetUsageCount() != 2 {
		t.Errorf("expected usage count 2, got %d", tool.GetUsageCount())
	}
}

func TestBaseTool_UsageLimit(t *testing.T) {
	handler := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		return "result", nil
	}

	baseTool := NewBaseTool("limited_tool", "Limited tool", handler)

	// 设置使用限制
	baseTool.SetUsageLimit(2)
	if baseTool.GetUsageLimit() != 2 {
		t.Errorf("expected usage limit 2, got %d", baseTool.GetUsageLimit())
	}

	ctx := context.Background()
	args := map[string]interface{}{}

	// 第一次执行
	_, err := baseTool.Execute(ctx, args)
	if err != nil {
		t.Errorf("first execution failed: %v", err)
	}
	if baseTool.IsUsageLimitExceeded() {
		t.Error("usage limit should not be exceeded after first use")
	}

	// 第二次执行
	_, err = baseTool.Execute(ctx, args)
	if err != nil {
		t.Errorf("second execution failed: %v", err)
	}
	if baseTool.IsUsageLimitExceeded() {
		t.Error("usage limit should not be exceeded after second use")
	}

	// 第三次执行应该失败
	_, err = baseTool.Execute(ctx, args)
	if err == nil {
		t.Error("expected error for third execution")
	}
	if !baseTool.IsUsageLimitExceeded() {
		t.Error("usage limit should be exceeded after third attempt")
	}
}

func TestBaseTool_ResetUsage(t *testing.T) {
	handler := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		return "result", nil
	}

	tool := NewBaseTool("reset_tool", "Reset tool", handler)
	ctx := context.Background()
	args := map[string]interface{}{}

	// 执行几次
	for i := 0; i < 3; i++ {
		_, err := tool.Execute(ctx, args)
		if err != nil {
			t.Errorf("execution %d failed: %v", i+1, err)
		}
	}

	if tool.GetUsageCount() != 3 {
		t.Errorf("expected usage count 3, got %d", tool.GetUsageCount())
	}

	// 重置使用统计
	tool.ResetUsage()
	if tool.GetUsageCount() != 0 {
		t.Errorf("expected usage count 0 after reset, got %d", tool.GetUsageCount())
	}
}

func TestBaseTool_ExecuteAsync(t *testing.T) {
	handler := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		// 模拟一些处理时间
		time.Sleep(10 * time.Millisecond)
		return "async result", nil
	}

	tool := NewBaseTool("async_tool", "Async tool", handler)
	ctx := context.Background()
	args := map[string]interface{}{}

	resultChan, err := tool.ExecuteAsync(ctx, args)
	if err != nil {
		t.Errorf("async execution setup failed: %v", err)
		return
	}

	select {
	case result := <-resultChan:
		if result.Error != nil {
			t.Errorf("async execution failed: %v", result.Error)
			return
		}

		if result.Output != "async result" {
			t.Errorf("expected output 'async result', got %s", result.Output)
		}

		if result.Duration <= 0 {
			t.Error("expected positive duration")
		}

		if result.Metadata == nil {
			t.Error("expected metadata")
		}

	case <-time.After(1 * time.Second):
		t.Error("async execution timeout")
	}
}

func TestCalculatorTool(t *testing.T) {
	calculator := NewCalculatorTool()
	ctx := context.Background()

	tests := []struct {
		name        string
		args        map[string]interface{}
		expected    float64
		expectError bool
	}{
		{
			name: "addition",
			args: map[string]interface{}{
				"operation": "add",
				"a":         float64(5),
				"b":         float64(3),
			},
			expected: 8,
		},
		{
			name: "subtraction",
			args: map[string]interface{}{
				"operation": "subtract",
				"a":         float64(10),
				"b":         float64(4),
			},
			expected: 6,
		},
		{
			name: "multiplication",
			args: map[string]interface{}{
				"operation": "multiply",
				"a":         float64(6),
				"b":         float64(7),
			},
			expected: 42,
		},
		{
			name: "division",
			args: map[string]interface{}{
				"operation": "divide",
				"a":         float64(15),
				"b":         float64(3),
			},
			expected: 5,
		},
		{
			name: "division by zero",
			args: map[string]interface{}{
				"operation": "divide",
				"a":         float64(10),
				"b":         float64(0),
			},
			expectError: true,
		},
		{
			name: "unsupported operation",
			args: map[string]interface{}{
				"operation": "power",
				"a":         float64(2),
				"b":         float64(3),
			},
			expectError: true,
		},
		{
			name: "missing operation",
			args: map[string]interface{}{
				"a": float64(5),
				"b": float64(3),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calculator.Execute(ctx, tt.args)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("expected result, got nil")
				return
			}

			if value, ok := result.(float64); !ok {
				t.Errorf("expected float64 result, got %T", result)
			} else if value != tt.expected {
				t.Errorf("expected result %f, got %f", tt.expected, value)
			}
		})
	}
}

func TestFileReaderTool(t *testing.T) {
	// 创建临时文件
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello, World!\nThis is a test file."

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	fileReader := NewFileReaderTool()
	ctx := context.Background()

	// 测试成功读取文件
	args := map[string]interface{}{
		"filepath": testFile,
	}

	result, err := fileReader.Execute(ctx, args)
	if err != nil {
		t.Errorf("file read failed: %v", err)
		return
	}

	if content, ok := result.(string); !ok {
		t.Errorf("expected string result, got %T", result)
	} else if content != testContent {
		t.Errorf("expected content %s, got %s", testContent, content)
	}

	// 测试文件不存在
	args = map[string]interface{}{
		"filepath": "/nonexistent/file.txt",
	}

	_, err = fileReader.Execute(ctx, args)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestFileWriterTool(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "write_test.txt")
	testContent := "This is test content for file writing."

	fileWriter := NewFileWriterTool()
	ctx := context.Background()

	// 测试文件写入
	args := map[string]interface{}{
		"filepath": testFile,
		"content":  testContent,
	}

	result, err := fileWriter.Execute(ctx, args)
	if err != nil {
		t.Errorf("file write failed: %v", err)
		return
	}

	if result == nil {
		t.Error("expected result, got nil")
		return
	}

	// 验证文件是否被创建并包含正确内容
	writtenContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("failed to read written file: %v", err)
		return
	}

	if string(writtenContent) != testContent {
		t.Errorf("expected file content %s, got %s", testContent, string(writtenContent))
	}
}

func TestJSONParserTool(t *testing.T) {
	jsonParser := NewJSONParserTool()
	ctx := context.Background()

	tests := []struct {
		name        string
		jsonString  string
		expectError bool
	}{
		{
			name:       "valid object",
			jsonString: `{"name": "John", "age": 30, "city": "New York"}`,
		},
		{
			name:       "valid array",
			jsonString: `["apple", "banana", "cherry"]`,
		},
		{
			name:       "valid number",
			jsonString: `42`,
		},
		{
			name:       "valid string",
			jsonString: `"Hello, World!"`,
		},
		{
			name:        "invalid JSON",
			jsonString:  `{"name": "John", "age": 30,}`,
			expectError: true,
		},
		{
			name:        "empty string",
			jsonString:  ``,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := map[string]interface{}{
				"json_string": tt.jsonString,
			}

			result, err := jsonParser.Execute(ctx, args)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("expected result, got nil")
			}
		})
	}
}

func TestJSONFormatterTool(t *testing.T) {
	jsonFormatter := NewJSONFormatterTool()
	ctx := context.Background()

	testData := map[string]interface{}{
		"name":    "John",
		"age":     30,
		"city":    "New York",
		"hobbies": []string{"reading", "swimming", "coding"},
	}

	// 测试紧凑格式
	args := map[string]interface{}{
		"data":   testData,
		"indent": false,
	}

	result, err := jsonFormatter.Execute(ctx, args)
	if err != nil {
		t.Errorf("JSON formatting failed: %v", err)
		return
	}

	jsonString, ok := result.(string)
	if !ok {
		t.Errorf("expected string result, got %T", result)
		return
	}

	// 验证结果是有效的JSON
	var parsed interface{}
	err = json.Unmarshal([]byte(jsonString), &parsed)
	if err != nil {
		t.Errorf("formatted JSON is invalid: %v", err)
	}

	// 测试缩进格式
	args["indent"] = true
	result, err = jsonFormatter.Execute(ctx, args)
	if err != nil {
		t.Errorf("indented JSON formatting failed: %v", err)
		return
	}

	indentedJSON, ok := result.(string)
	if !ok {
		t.Errorf("expected string result, got %T", result)
		return
	}

	// 缩进版本应该更长
	if len(indentedJSON) <= len(jsonString) {
		t.Error("indented JSON should be longer than compact JSON")
	}
}

func TestTextAnalyzerTool(t *testing.T) {
	textAnalyzer := NewTextAnalyzerTool()
	ctx := context.Background()

	testText := "Hello world! This is a test text with multiple words and sentences. " +
		"It contains punctuation, numbers like 123, and various characters."

	args := map[string]interface{}{
		"text": testText,
	}

	result, err := textAnalyzer.Execute(ctx, args)
	if err != nil {
		t.Errorf("text analysis failed: %v", err)
		return
	}

	analysis, ok := result.(map[string]interface{})
	if !ok {
		t.Errorf("expected map result, got %T", result)
		return
	}

	// 验证分析结果包含必要字段
	requiredFields := []string{"character_count", "word_count", "line_count", "average_word_length"}
	for _, field := range requiredFields {
		if _, exists := analysis[field]; !exists {
			t.Errorf("expected field %s in analysis result", field)
		}
	}

	// 验证字符计数
	if charCount, ok := analysis["character_count"].(int); !ok {
		t.Error("character_count should be int")
	} else if charCount != len(testText) {
		t.Errorf("expected character count %d, got %d", len(testText), charCount)
	}

	// 验证单词计数
	expectedWords := len(strings.Fields(testText))
	if wordCount, ok := analysis["word_count"].(int); !ok {
		t.Error("word_count should be int")
	} else if wordCount != expectedWords {
		t.Errorf("expected word count %d, got %d", expectedWords, wordCount)
	}

	// 验证行数计数
	expectedLines := len(strings.Split(testText, "\n"))
	if lineCount, ok := analysis["line_count"].(int); !ok {
		t.Error("line_count should be int")
	} else if lineCount != expectedLines {
		t.Errorf("expected line count %d, got %d", expectedLines, lineCount)
	}
}

func TestToolCollection(t *testing.T) {
	collection := NewToolCollection()

	// 初始状态
	if collection.Size() != 0 {
		t.Errorf("expected initial size 0, got %d", collection.Size())
	}

	// 添加工具
	calculator := NewCalculatorTool()
	fileReader := NewFileReaderTool()

	err := collection.Add(calculator)
	if err != nil {
		t.Errorf("failed to add calculator: %v", err)
	}

	err = collection.Add(fileReader)
	if err != nil {
		t.Errorf("failed to add file reader: %v", err)
	}

	if collection.Size() != 2 {
		t.Errorf("expected size 2, got %d", collection.Size())
	}

	// 测试重复添加
	err = collection.Add(calculator)
	if err == nil {
		t.Error("expected error for duplicate tool")
	}

	// 获取工具
	retrieved, found := collection.Get("calculator")
	if !found {
		t.Error("expected to find calculator")
	}
	if retrieved.GetName() != "calculator" {
		t.Errorf("expected calculator, got %s", retrieved.GetName())
	}

	// 检查存在性
	if !collection.Has("file_reader") {
		t.Error("expected to have file_reader")
	}

	if collection.Has("nonexistent") {
		t.Error("expected to not have nonexistent tool")
	}

	// 列出工具
	toolNames := collection.List()
	if len(toolNames) != 2 {
		t.Errorf("expected 2 tool names, got %d", len(toolNames))
	}

	// 获取所有工具
	allTools := collection.All()
	if len(allTools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(allTools))
	}

	// 执行工具
	args := map[string]interface{}{
		"operation": "add",
		"a":         float64(5),
		"b":         float64(3),
	}

	result, err := collection.Execute(context.Background(), "calculator", args)
	if err != nil {
		t.Errorf("tool execution failed: %v", err)
	}

	if result != float64(8) {
		t.Errorf("expected result 8, got %v", result)
	}

	// 移除工具
	removed := collection.Remove("file_reader")
	if !removed {
		t.Error("expected to remove file_reader")
	}

	if collection.Size() != 1 {
		t.Errorf("expected size 1 after removal, got %d", collection.Size())
	}

	// 清空集合
	collection.Clear()
	if collection.Size() != 0 {
		t.Errorf("expected size 0 after clear, got %d", collection.Size())
	}
}

func TestToolCollection_LoadBasicTools(t *testing.T) {
	collection := NewToolCollection()

	err := collection.LoadBasicTools()
	if err != nil {
		t.Errorf("failed to load basic tools: %v", err)
	}

	expectedTools := []string{
		"calculator",
		"file_reader",
		"file_writer",
		"json_parser",
		"json_formatter",
		"text_analyzer",
	}

	if collection.Size() != len(expectedTools) {
		t.Errorf("expected %d tools, got %d", len(expectedTools), collection.Size())
	}

	for _, toolName := range expectedTools {
		if !collection.Has(toolName) {
			t.Errorf("expected to have tool %s", toolName)
		}
	}
}

func TestGlobalToolRegistry(t *testing.T) {
	// 注意：这个测试可能会影响全局状态，应该小心处理
	registry := GetGlobalToolRegistry()

	// 全局注册表应该已经包含基础工具
	if registry.Size() == 0 {
		t.Error("expected global registry to have tools")
	}

	// 测试获取已注册的工具
	calculator, found := GetRegisteredTool("calculator")
	if !found {
		t.Error("expected to find calculator in global registry")
	}

	if calculator.GetName() != "calculator" {
		t.Errorf("expected calculator, got %s", calculator.GetName())
	}

	// 列出注册的工具
	registeredTools := ListRegisteredTools()
	if len(registeredTools) == 0 {
		t.Error("expected some registered tools")
	}
}

func BenchmarkBaseTool_Execute(b *testing.B) {
	handler := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		return "benchmark result", nil
	}

	tool := NewBaseTool("benchmark_tool", "Benchmark tool", handler)
	ctx := context.Background()
	args := map[string]interface{}{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tool.Execute(ctx, args)
		if err != nil {
			b.Fatalf("execution failed: %v", err)
		}
	}
}

func BenchmarkCalculatorTool(b *testing.B) {
	calculator := NewCalculatorTool()
	ctx := context.Background()
	args := map[string]interface{}{
		"operation": "add",
		"a":         float64(5),
		"b":         float64(3),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := calculator.Execute(ctx, args)
		if err != nil {
			b.Fatalf("execution failed: %v", err)
		}
	}
}
