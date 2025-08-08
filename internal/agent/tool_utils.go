package agent

import (
	"context"
	"fmt"
	"strings"
)

// parseTools 将原始工具转换为结构化工具
// 对标 crewAI Python版本的 parse_tools 函数
func parseTools(rawTools []Tool) []Tool {
	if len(rawTools) == 0 {
		return []Tool{}
	}

	parsedTools := make([]Tool, 0, len(rawTools))

	for _, tool := range rawTools {
		if tool == nil {
			continue
		}

		// 确保工具有有效的名称和描述
		if tool.GetName() == "" {
			continue
		}

		parsedTools = append(parsedTools, tool)
	}

	return parsedTools
}

// getToolNames 获取工具名称列表
// 对标 crewAI Python版本的 get_tool_names 函数
func getToolNames(tools []Tool) string {
	if len(tools) == 0 {
		return ""
	}

	names := make([]string, 0, len(tools))
	for _, tool := range tools {
		if tool != nil && tool.GetName() != "" {
			names = append(names, tool.GetName())
		}
	}

	return strings.Join(names, ", ")
}

// renderTextDescriptionAndArgs 渲染工具描述和参数
// 对标 crewAI Python版本的 render_text_description_and_args 函数
func renderTextDescriptionAndArgs(tools []Tool) string {
	if len(tools) == 0 {
		return "No tools available."
	}

	var builder strings.Builder

	for i, tool := range tools {
		if tool == nil {
			continue
		}

		// 工具名称和描述
		builder.WriteString(fmt.Sprintf("%s: %s", tool.GetName(), tool.GetDescription()))

		// 添加参数信息（如果有）
		schema := tool.GetSchema()
		if len(schema.Parameters) > 0 {
			builder.WriteString("\n  Parameters:")

			// 处理参数
			if properties, ok := schema.Parameters["properties"].(map[string]interface{}); ok {
				for paramName, paramInfo := range properties {
					if paramMap, ok := paramInfo.(map[string]interface{}); ok {
						if desc, exists := paramMap["description"]; exists {
							builder.WriteString(fmt.Sprintf("\n    - %s: %v", paramName, desc))
						} else {
							builder.WriteString(fmt.Sprintf("\n    - %s", paramName))
						}
					}
				}
			}

			// 添加必需参数信息
			if len(schema.Required) > 0 {
				builder.WriteString(fmt.Sprintf("\n  Required: %s", strings.Join(schema.Required, ", ")))
			}
		}

		// 添加分隔符（除了最后一个工具）
		if i < len(tools)-1 {
			builder.WriteString("\n\n")
		}
	}

	return builder.String()
}

// selectToolsForTask 为任务选择工具
// 实现工具优先级逻辑：task.tools > agent.tools > []
func selectToolsForTask(task Task, agent Agent) []Tool {
	// 1. 优先使用任务级别的工具
	if taskTools := task.GetTools(); len(taskTools) > 0 {
		return taskTools
	}

	// 2. 如果任务没有工具，使用Agent的工具
	if agentTools := agent.GetTools(); len(agentTools) > 0 {
		return agentTools
	}

	// 3. 默认返回空工具列表
	return []Tool{}
}

// prepareToolsForAgent 为Agent准备工具
// 对标 crewAI Python版本的 _prepare_tools 方法
func prepareToolsForAgent(agent Agent, task Task, tools []Tool) []Tool {
	if len(tools) == 0 {
		return []Tool{}
	}

	// 解析和验证工具
	parsedTools := parseTools(tools)

	// 可以在这里添加更多的工具准备逻辑，比如：
	// - 工具权限验证
	// - 工具使用限制检查
	// - 工具依赖关系处理

	return parsedTools
}

// createToolsDescription 创建工具描述字符串
// 用于传递给LLM的提示信息
func createToolsDescription(tools []Tool) string {
	if len(tools) == 0 {
		return "No tools available for this task."
	}

	var builder strings.Builder
	builder.WriteString("Available tools:\n")

	for _, tool := range tools {
		if tool == nil {
			continue
		}

		builder.WriteString(fmt.Sprintf("- %s: %s\n", tool.GetName(), tool.GetDescription()))
	}

	return builder.String()
}

// validateToolSchema 验证工具模式
func validateToolSchema(tool Tool) error {
	if tool == nil {
		return fmt.Errorf("tool is nil")
	}

	if tool.GetName() == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if tool.GetDescription() == "" {
		return fmt.Errorf("tool description cannot be empty for tool: %s", tool.GetName())
	}

	return nil
}

// findToolByName 根据名称查找工具
func findToolByName(tools []Tool, name string) (Tool, bool) {
	for _, tool := range tools {
		if tool != nil && tool.GetName() == name {
			return tool, true
		}
	}
	return nil, false
}

// filterValidTools 过滤有效的工具
func filterValidTools(tools []Tool) []Tool {
	validTools := make([]Tool, 0, len(tools))

	for _, tool := range tools {
		if err := validateToolSchema(tool); err == nil {
			validTools = append(validTools, tool)
		}
	}

	return validTools
}

// ToolExecutionContext 工具执行上下文
type ToolExecutionContext struct {
	Agent   Agent
	Task    Task
	Tools   []Tool
	Context map[string]interface{}
}

// NewToolExecutionContext 创建工具执行上下文
func NewToolExecutionContext(agent Agent, task Task) *ToolExecutionContext {
	tools := selectToolsForTask(task, agent)
	preparedTools := prepareToolsForAgent(agent, task, tools)

	return &ToolExecutionContext{
		Agent:   agent,
		Task:    task,
		Tools:   preparedTools,
		Context: make(map[string]interface{}),
	}
}

// GetToolNames 获取工具名称列表
func (ctx *ToolExecutionContext) GetToolNames() string {
	return getToolNames(ctx.Tools)
}

// GetToolsDescription 获取工具描述
func (ctx *ToolExecutionContext) GetToolsDescription() string {
	return renderTextDescriptionAndArgs(ctx.Tools)
}

// HasTools 检查是否有可用工具
func (ctx *ToolExecutionContext) HasTools() bool {
	return len(ctx.Tools) > 0
}

// ExecuteTool 执行指定工具
func (ctx *ToolExecutionContext) ExecuteTool(execCtx context.Context, toolName string, args map[string]interface{}) (interface{}, error) {
	tool, found := findToolByName(ctx.Tools, toolName)
	if !found {
		return nil, fmt.Errorf("tool '%s' not found. Available tools: %s", toolName, ctx.GetToolNames())
	}

	// 这里可以添加执行前的验证和日志记录
	return tool.Execute(execCtx, args)
}
