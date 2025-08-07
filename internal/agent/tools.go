package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// BaseTool 实现了Tool接口的基础结构
type BaseTool struct {
	name        string
	description string
	schema      ToolSchema
	handler     func(ctx context.Context, args map[string]interface{}) (interface{}, error)
	usageCount  int
	usageLimit  int
	mu          sync.RWMutex
}

// NewBaseTool 创建基础工具
func NewBaseTool(name, description string, handler func(ctx context.Context, args map[string]interface{}) (interface{}, error)) *BaseTool {
	return &BaseTool{
		name:        name,
		description: description,
		schema: ToolSchema{
			Name:        name,
			Description: description,
			Parameters:  make(map[string]interface{}),
			Required:    make([]string, 0),
		},
		handler:    handler,
		usageCount: 0,
		usageLimit: -1, // -1 表示无限制
	}
}

// GetName 返回工具名称
func (t *BaseTool) GetName() string {
	return t.name
}

// GetDescription 返回工具描述
func (t *BaseTool) GetDescription() string {
	return t.description
}

// GetSchema 返回工具模式
func (t *BaseTool) GetSchema() ToolSchema {
	return t.schema
}

// Execute 执行工具
func (t *BaseTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	t.mu.Lock()

	// 检查使用限制
	if t.usageLimit >= 0 && t.usageCount >= t.usageLimit {
		t.usageCount++ // 即使被阻止，也增加计数来标记尝试
		t.mu.Unlock()
		return nil, fmt.Errorf("tool usage limit exceeded: %d/%d", t.usageCount-1, t.usageLimit)
	}

	t.usageCount++
	t.mu.Unlock()

	// 执行工具逻辑
	return t.handler(ctx, args)
}

// ExecuteAsync 异步执行工具
func (t *BaseTool) ExecuteAsync(ctx context.Context, args map[string]interface{}) (<-chan ToolResult, error) {
	resultChan := make(chan ToolResult, 1)

	go func() {
		defer close(resultChan)

		start := time.Now()
		output, err := t.Execute(ctx, args)
		duration := time.Since(start)

		resultChan <- ToolResult{
			Output:   output,
			Error:    err,
			Duration: duration,
			Metadata: map[string]interface{}{
				"tool_name":      t.name,
				"execution_time": duration.Milliseconds(),
			},
		}
	}()

	return resultChan, nil
}

// GetUsageCount 获取使用次数
func (t *BaseTool) GetUsageCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.usageCount
}

// GetUsageLimit 获取使用限制
func (t *BaseTool) GetUsageLimit() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.usageLimit
}

// ResetUsage 重置使用统计
func (t *BaseTool) ResetUsage() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.usageCount = 0
}

// IsUsageLimitExceeded 检查是否超出使用限制
func (t *BaseTool) IsUsageLimitExceeded() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.usageLimit >= 0 && t.usageCount > t.usageLimit
}

// SetUsageLimit 设置使用限制
func (t *BaseTool) SetUsageLimit(limit int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.usageLimit = limit
}

// SetSchema 设置工具模式
func (t *BaseTool) SetSchema(schema ToolSchema) {
	t.schema = schema
}

// 预定义工具

// NewCalculatorTool 创建计算器工具
func NewCalculatorTool() Tool {
	return NewBaseTool(
		"calculator",
		"Perform basic mathematical calculations (add, subtract, multiply, divide)",
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			operation, ok := args["operation"].(string)
			if !ok {
				return nil, fmt.Errorf("operation is required and must be a string")
			}

			a, ok := args["a"].(float64)
			if !ok {
				return nil, fmt.Errorf("argument 'a' is required and must be a number")
			}

			b, ok := args["b"].(float64)
			if !ok {
				return nil, fmt.Errorf("argument 'b' is required and must be a number")
			}

			switch operation {
			case "add":
				return a + b, nil
			case "subtract":
				return a - b, nil
			case "multiply":
				return a * b, nil
			case "divide":
				if b == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				return a / b, nil
			default:
				return nil, fmt.Errorf("unsupported operation: %s", operation)
			}
		},
	)
}

// NewFileReaderTool 创建文件读取工具
func NewFileReaderTool() Tool {
	tool := NewBaseTool(
		"file_reader",
		"Read contents from a file",
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			filepath, ok := args["filepath"].(string)
			if !ok {
				return nil, fmt.Errorf("filepath is required and must be a string")
			}

			content, err := os.ReadFile(filepath)
			if err != nil {
				return nil, fmt.Errorf("failed to read file %s: %w", filepath, err)
			}

			return string(content), nil
		},
	)

	// 设置工具模式
	tool.SetSchema(ToolSchema{
		Name:        "file_reader",
		Description: "Read contents from a file",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"filepath": map[string]interface{}{
					"type":        "string",
					"description": "Path to the file to read",
				},
			},
		},
		Required: []string{"filepath"},
	})

	return tool
}

// NewFileWriterTool 创建文件写入工具
func NewFileWriterTool() Tool {
	tool := NewBaseTool(
		"file_writer",
		"Write content to a file",
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			filepath, ok := args["filepath"].(string)
			if !ok {
				return nil, fmt.Errorf("filepath is required and must be a string")
			}

			content, ok := args["content"].(string)
			if !ok {
				return nil, fmt.Errorf("content is required and must be a string")
			}

			err := os.WriteFile(filepath, []byte(content), 0644)
			if err != nil {
				return nil, fmt.Errorf("failed to write file %s: %w", filepath, err)
			}

			return fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), filepath), nil
		},
	)

	// 设置工具模式
	tool.SetSchema(ToolSchema{
		Name:        "file_writer",
		Description: "Write content to a file",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"filepath": map[string]interface{}{
					"type":        "string",
					"description": "Path to the file to write",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "Content to write to the file",
				},
			},
		},
		Required: []string{"filepath", "content"},
	})

	return tool
}

// NewJSONParserTool 创建JSON解析工具
func NewJSONParserTool() Tool {
	return NewBaseTool(
		"json_parser",
		"Parse JSON string into structured data",
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			jsonStr, ok := args["json_string"].(string)
			if !ok {
				return nil, fmt.Errorf("json_string is required and must be a string")
			}

			var result interface{}
			err := json.Unmarshal([]byte(jsonStr), &result)
			if err != nil {
				return nil, fmt.Errorf("failed to parse JSON: %w", err)
			}

			return result, nil
		},
	)
}

// NewJSONFormatterTool 创建JSON格式化工具
func NewJSONFormatterTool() Tool {
	return NewBaseTool(
		"json_formatter",
		"Format data as JSON string",
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			data, ok := args["data"]
			if !ok {
				return nil, fmt.Errorf("data is required")
			}

			indent, _ := args["indent"].(bool)

			var jsonBytes []byte
			var err error

			if indent {
				jsonBytes, err = json.MarshalIndent(data, "", "  ")
			} else {
				jsonBytes, err = json.Marshal(data)
			}

			if err != nil {
				return nil, fmt.Errorf("failed to format JSON: %w", err)
			}

			return string(jsonBytes), nil
		},
	)
}

// NewTextAnalyzerTool 创建文本分析工具
func NewTextAnalyzerTool() Tool {
	return NewBaseTool(
		"text_analyzer",
		"Analyze text and provide statistics",
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			text, ok := args["text"].(string)
			if !ok {
				return nil, fmt.Errorf("text is required and must be a string")
			}

			words := len(strings.Fields(text))
			chars := len(text)
			lines := len(strings.Split(text, "\n"))

			result := map[string]interface{}{
				"character_count": chars,
				"word_count":      words,
				"line_count":      lines,
				"average_word_length": func() float64 {
					if words == 0 {
						return 0
					}
					return float64(chars) / float64(words)
				}(),
			}

			return result, nil
		},
	)
}

// ToolCollection 工具集合
type ToolCollection struct {
	tools map[string]Tool
	mu    sync.RWMutex
}

// NewToolCollection 创建工具集合
func NewToolCollection() *ToolCollection {
	return &ToolCollection{
		tools: make(map[string]Tool),
	}
}

// Add 添加工具
func (tc *ToolCollection) Add(tool Tool) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	name := tool.GetName()
	if _, exists := tc.tools[name]; exists {
		return fmt.Errorf("tool with name %s already exists", name)
	}

	tc.tools[name] = tool
	return nil
}

// Get 获取工具
func (tc *ToolCollection) Get(name string) (Tool, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	tool, exists := tc.tools[name]
	return tool, exists
}

// Remove 移除工具
func (tc *ToolCollection) Remove(name string) bool {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if _, exists := tc.tools[name]; exists {
		delete(tc.tools, name)
		return true
	}
	return false
}

// List 列出所有工具名称
func (tc *ToolCollection) List() []string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	names := make([]string, 0, len(tc.tools))
	for name := range tc.tools {
		names = append(names, name)
	}
	return names
}

// All 返回所有工具
func (tc *ToolCollection) All() []Tool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	tools := make([]Tool, 0, len(tc.tools))
	for _, tool := range tc.tools {
		tools = append(tools, tool)
	}
	return tools
}

// Size 返回工具数量
func (tc *ToolCollection) Size() int {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return len(tc.tools)
}

// Clear 清空所有工具
func (tc *ToolCollection) Clear() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.tools = make(map[string]Tool)
}

// Has 检查是否包含指定工具
func (tc *ToolCollection) Has(name string) bool {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	_, exists := tc.tools[name]
	return exists
}

// Execute 执行指定工具
func (tc *ToolCollection) Execute(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
	tool, exists := tc.Get(name)
	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}

	return tool.Execute(ctx, args)
}

// LoadBasicTools 加载基础工具集
func (tc *ToolCollection) LoadBasicTools() error {
	basicTools := []Tool{
		NewCalculatorTool(),
		NewFileReaderTool(),
		NewFileWriterTool(),
		NewJSONParserTool(),
		NewJSONFormatterTool(),
		NewTextAnalyzerTool(),
	}

	for _, tool := range basicTools {
		if err := tc.Add(tool); err != nil {
			return fmt.Errorf("failed to add tool %s: %w", tool.GetName(), err)
		}
	}

	return nil
}

// ToolRegistry 工具注册表（全局）
type ToolRegistry struct {
	*ToolCollection
}

var globalToolRegistry = &ToolRegistry{
	ToolCollection: NewToolCollection(),
}

// GetGlobalToolRegistry 获取全局工具注册表
func GetGlobalToolRegistry() *ToolRegistry {
	return globalToolRegistry
}

// RegisterTool 注册工具到全局注册表
func RegisterTool(tool Tool) error {
	return globalToolRegistry.Add(tool)
}

// GetRegisteredTool 从全局注册表获取工具
func GetRegisteredTool(name string) (Tool, bool) {
	return globalToolRegistry.Get(name)
}

// ListRegisteredTools 列出全局注册的工具
func ListRegisteredTools() []string {
	return globalToolRegistry.List()
}

// LoadBasicToolsGlobally 在全局注册表中加载基础工具
func LoadBasicToolsGlobally() error {
	return globalToolRegistry.LoadBasicTools()
}

// 确保BaseTool实现了Tool接口
var _ Tool = (*BaseTool)(nil)

func init() {
	// 自动加载基础工具到全局注册表
	_ = LoadBasicToolsGlobally()
}
