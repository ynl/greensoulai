package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ynl/greensoulai/internal/agent"
	"github.com/ynl/greensoulai/internal/crew"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// 演示记忆如何传递给LLM的完整流程
func main() {
	fmt.Println("=== 记忆到LLM数据传递流程演示 ===\n")

	// 1. 初始化基础设施
	ctx := context.Background()
	logger := logger.NewConsoleLogger()
	eventBus := events.NewEventBus(logger)

	// 2. 创建记忆管理器
	memConfig := crew.DefaultMemoryManagerConfig()
	memConfig.StoragePath = "examples/memory/data"
	memoryManager := crew.NewMemoryManager(nil, memConfig, eventBus, logger)
	defer memoryManager.Close()

	fmt.Println("📊 步骤1：存储结构化记忆数据")
	fmt.Println("----------------------------------------")

	// 存储一些示例记忆数据，展示原始数据结构
	memories := []struct {
		Type     string
		Value    string
		Metadata map[string]interface{}
		Agent    string
	}{
		{
			Type:  "short_term",
			Value: "用户反馈显示界面复杂度是主要痛点",
			Metadata: map[string]interface{}{
				"context":    "基于500份用户调研的关键发现",
				"type":       "user_feedback",
				"priority":   "high",
				"source":     "Q3用户调研报告",
				"tags":       []string{"UI", "用户体验", "痛点分析"},
				"confidence": 0.95,
			},
			Agent: "user_research_analyst",
		},
		{
			Type:  "short_term",
			Value: "移动端用户抱怨按钮过小和层级太深",
			Metadata: map[string]interface{}{
				"context":  "移动端用户体验调研中的重要发现",
				"type":     "user_feedback",
				"priority": "high",
				"platform": "mobile",
			},
			Agent: "mobile_ux_analyst",
		},
		{
			Type:  "entity",
			Value: "产品经理张三：负责用户体验优化，具有5年产品设计经验",
			Metadata: map[string]interface{}{
				"context":   "关键团队成员信息",
				"type":      "team_member",
				"role":      "product_manager",
				"expertise": []string{"UX设计", "产品策略", "用户研究"},
			},
			Agent: "team_info_manager",
		},
	}

	// 保存记忆数据
	for i, mem := range memories {
		// 打印原始数据结构
		fmt.Printf("记忆项 %d (类型: %s):\n", i+1, mem.Type)

		// 展示完整的数据结构
		memoryData := map[string]interface{}{
			"value":    mem.Value,
			"metadata": mem.Metadata,
			"agent":    mem.Agent,
			"type":     mem.Type,
		}

		jsonData, _ := json.MarshalIndent(memoryData, "  ", "  ")
		fmt.Printf("  原始结构: %s\n", string(jsonData))

		// 保存到记忆系统
		err := memoryManager.SaveMemory(ctx, mem.Type, mem.Value, mem.Metadata, mem.Agent)
		if err != nil {
			log.Printf("保存记忆失败: %v", err)
		} else {
			fmt.Printf("  ✓ 已保存到%s记忆\n", mem.Type)
		}
		fmt.Println()
	}

	fmt.Println("🔍 步骤2：模拟任务查询和记忆检索")
	fmt.Println("----------------------------------------")

	// 创建一个任务来触发记忆检索
	task := &MockTask{
		id:          "task_001",
		description: "分析用户反馈数据，找出产品改进点",
	}

	fmt.Printf("任务描述: %s\n", task.GetDescription())
	fmt.Println()

	fmt.Println("🧠 步骤3：ContextualMemory智能上下文构建")
	fmt.Println("----------------------------------------")

	// 使用ContextualMemory构建智能上下文
	contextResult, err := memoryManager.BuildTaskContext(ctx, task, "重点关注用户体验改进")
	if err != nil {
		log.Fatalf("上下文构建失败: %v", err)
	}

	fmt.Println("构建的上下文信息：")
	if contextResult != "" {
		fmt.Printf("```\n%s\n```\n", contextResult)
	} else {
		fmt.Println("(由于使用测试embedder，可能无法检索到相似内容)")
		// 为了演示，我们手动构建一个示例上下文
		contextResult = buildDemoContext()
		fmt.Println("\n为演示目的，手动构建示例上下文：")
		fmt.Printf("```\n%s\n```\n", contextResult)
	}

	fmt.Println("\n📝 步骤4：Agent构建完整Prompt")
	fmt.Println("----------------------------------------")

	// 模拟Agent构建完整prompt的过程
	fullPrompt := buildFullPromptDemo(task, contextResult)

	fmt.Println("发送给LLM的完整Prompt：")
	fmt.Printf("```\n%s\n```\n", fullPrompt)

	fmt.Println("\n🤖 步骤5：LLM消息结构")
	fmt.Println("----------------------------------------")

	// 展示LLM实际接收的消息结构
	messages := buildLLMMessages(fullPrompt)
	fmt.Println("LLM接收的Messages结构：")
	for i, msg := range messages {
		fmt.Printf("Message %d:\n", i+1)
		fmt.Printf("  Role: %s\n", msg.Role)
		fmt.Printf("  Content: %s\n", truncateString(msg.Content, 200))
		if len(msg.Content) > 200 {
			fmt.Printf("  ... (总长度: %d 字符)\n", len(msg.Content))
		}
		fmt.Println()
	}

	fmt.Println("🎯 步骤6：数据传递总结")
	fmt.Println("----------------------------------------")

	fmt.Printf("数据流转统计：\n")
	fmt.Printf("- 原始记忆项数量: %d\n", len(memories))
	fmt.Printf("- 上下文长度: %d 字符\n", len(contextResult))
	fmt.Printf("- 完整Prompt长度: %d 字符\n", len(fullPrompt))
	fmt.Printf("- LLM消息数量: %d\n", len(messages))

	fmt.Println("\n数据转换过程：")
	fmt.Println("  结构化存储 → 相似性检索 → 格式化文本 → Prompt集成 → LLM调用")
	fmt.Println("  MemoryItem  → Search API  → Contextual  → Agent     → Provider")
	fmt.Println("  (JSON)      → ([]Items)   → (String)    → (Prompt)  → (Messages)")

	fmt.Println("\n=== 演示完成 ===")
	fmt.Println("💡 关键要点：")
	fmt.Println("1. 记忆以结构化JSON格式存储，包含丰富的元数据")
	fmt.Println("2. ContextualMemory智能检索和格式化相关记忆")
	fmt.Println("3. Agent将格式化的记忆上下文无缝集成到Prompt中")
	fmt.Println("4. LLM接收到包含记忆信息的完整上下文，提升回答质量")
}

// MockTask 简化的任务实现（仅用于演示）
type MockTask struct {
	id          string
	description string
}

func (t *MockTask) GetID() string                     { return t.id }
func (t *MockTask) GetDescription() string            { return t.description }
func (t *MockTask) SetDescription(description string) { t.description = description }
func (t *MockTask) GetExpectedOutput() string {
	return "生成详细的分析报告，包含具体的改进建议"
}
func (t *MockTask) GetContext() map[string]interface{}                                  { return nil }
func (t *MockTask) IsHumanInputRequired() bool                                          { return false }
func (t *MockTask) SetHumanInput(input string)                                          {}
func (t *MockTask) GetHumanInput() string                                               { return "" }
func (t *MockTask) GetOutputFormat() agent.OutputFormat                                 { return agent.OutputFormatRAW }
func (t *MockTask) GetTools() []agent.Tool                                              { return nil }
func (t *MockTask) AddTool(tool agent.Tool) error                                       { return nil }
func (t *MockTask) SetTools(tools []agent.Tool) error                                   { return nil }
func (t *MockTask) HasTools() bool                                                      { return false }
func (t *MockTask) Validate() error                                                     { return nil }
func (t *MockTask) GetAssignedAgent() agent.Agent                                       { return nil }
func (t *MockTask) SetAssignedAgent(agent agent.Agent) error                            { return nil }
func (t *MockTask) IsAsyncExecution() bool                                              { return false }
func (t *MockTask) SetAsyncExecution(async bool)                                        {}
func (t *MockTask) SetContext(context map[string]interface{})                           {}
func (t *MockTask) GetName() string                                                     { return "" }
func (t *MockTask) SetName(name string)                                                 {}
func (t *MockTask) GetOutputFile() string                                               { return "" }
func (t *MockTask) SetOutputFile(filename string) error                                 { return nil }
func (t *MockTask) GetCreateDirectory() bool                                            { return false }
func (t *MockTask) SetCreateDirectory(create bool)                                      {}
func (t *MockTask) GetCallback() func(context.Context, *agent.TaskOutput) error         { return nil }
func (t *MockTask) SetCallback(callback func(context.Context, *agent.TaskOutput) error) {}
func (t *MockTask) GetContextTasks() []agent.Task                                       { return nil }
func (t *MockTask) SetContextTasks(tasks []agent.Task)                                  {}
func (t *MockTask) GetRetryCount() int                                                  { return 0 }
func (t *MockTask) GetMaxRetries() int                                                  { return 0 }
func (t *MockTask) SetMaxRetries(maxRetries int)                                        {}
func (t *MockTask) IsMarkdownOutput() bool                                              { return false }
func (t *MockTask) SetMarkdownOutput(markdown bool)                                     {}
func (t *MockTask) HasGuardrail() bool                                                  { return false }
func (t *MockTask) GetGuardrail() agent.TaskGuardrail                                   { return nil }
func (t *MockTask) SetGuardrail(guardrail agent.TaskGuardrail)                          {}

// buildDemoContext 构建演示用的上下文（模拟ContextualMemory的输出）
func buildDemoContext() string {
	return `Recent Insights:
- 基于500份用户调研的关键发现：界面复杂度是主要痛点
- 移动端用户体验调研中的重要发现：按钮过小和层级太深

Entities:
- 关键团队成员信息：产品经理张三负责用户体验优化，具有5年产品设计经验`
}

// buildFullPromptDemo 模拟Agent构建完整Prompt的过程
func buildFullPromptDemo(task *MockTask, memoryContext string) string {
	prompt := task.GetDescription()

	// 添加期望输出
	if expectedOutput := task.GetExpectedOutput(); expectedOutput != "" {
		prompt += fmt.Sprintf("\n\nExpected Output: %s", expectedOutput)
	}

	// 添加记忆上下文
	if memoryContext != "" {
		prompt += fmt.Sprintf("\n\nRelevant Memory:\n%s", memoryContext)
	}

	// 添加工具信息（演示用）
	prompt += "\n\nAvailable Tools:\n- data_analyzer: 分析数据并生成洞察\n- report_generator: 生成结构化报告"

	// 添加人工输入
	prompt += "\n\nHuman Input: 重点关注用户体验改进"

	// 添加工具使用指导
	prompt += "\n\nTo use a tool, respond with a JSON object in the following format:"
	prompt += "\n{\"tool_name\": \"<tool_name>\", \"arguments\": {\"arg1\": \"value1\", \"arg2\": \"value2\"}}"
	prompt += "\nIf no tool is needed, provide your response directly."

	return prompt
}

// LLMMessage LLM消息结构
type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// buildLLMMessages 构建LLM消息结构
func buildLLMMessages(prompt string) []LLMMessage {
	return []LLMMessage{
		{
			Role: "system",
			Content: "你是一个专业的产品分析师，擅长分析用户反馈并提出改进建议。你有访问记忆系统的能力，" +
				"可以利用历史数据和洞察来提供更准确的分析。请基于提供的信息进行深入分析。",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}
}

// truncateString 截断字符串用于显示
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
