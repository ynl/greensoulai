package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ynl/greensoulai/internal/agent"
	"github.com/ynl/greensoulai/internal/llm"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// 快速开始示例 - 5分钟上手 GreenSoulAI
func main() {
	fmt.Println("🚀 GreenSoulAI 快速开始示例")
	fmt.Println("=============================")

	// 1. 从环境变量读取 OpenRouter API 密钥
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ 错误：未设置 OPENROUTER_API_KEY 环境变量")
		fmt.Println()
		fmt.Println("请先设置您的OpenRouter API密钥：")
		fmt.Println("export OPENROUTER_API_KEY='your-openrouter-api-key-here'")
		fmt.Println()
		fmt.Println("获取免费API密钥：https://openrouter.ai/")
		return
	}
	fmt.Println("✅ 成功读取 OpenRouter API 密钥")

	// 2. 初始化基础组件
	logger := logger.NewConsoleLogger()
	eventBus := events.NewEventBus(logger)

	// 3. 创建 OpenRouter Kimi LLM 实例
	llmInstance := llm.NewOpenAILLM("moonshotai/kimi-k2:free",
		llm.WithAPIKey(apiKey),
		llm.WithBaseURL("https://openrouter.ai/api/v1"),
		llm.WithTimeout(30*time.Second),
		llm.WithMaxRetries(3),
		llm.WithCustomHeader("HTTP-Referer", "https://github.com/ynl/greensoulai"),
		llm.WithCustomHeader("X-Title", "GreenSoulAI 快速开始"),
	)

	defer llmInstance.Close()
	fmt.Printf("✅ LLM初始化成功: %s\n", llmInstance.GetModel())

	// 4. 创建 AI 助手 Agent
	assistantConfig := agent.AgentConfig{
		Role:      "智能助手",
		Goal:      "高效准确地帮助用户完成各种任务",
		Backstory: "你是一个知识渊博的AI助手，能够帮助用户解决各种问题。你总是用中文回答，态度友善，回答准确。",
		LLM:       llmInstance,
		EventBus:  eventBus,
		Logger:    logger,
	}

	assistant, err := agent.NewBaseAgent(assistantConfig)
	if err != nil {
		fmt.Printf("❌ 创建Agent失败: %v\n", err)
		return
	}

	// 5. 为 Agent 添加工具
	calculator := agent.NewCalculatorTool()
	if err := assistant.AddTool(calculator); err != nil {
		fmt.Printf("❌ 添加工具失败: %v\n", err)
		return
	}

	// 6. 初始化 Agent
	if err := assistant.Initialize(); err != nil {
		fmt.Printf("❌ 初始化Agent失败: %v\n", err)
		return
	}

	fmt.Printf("✅ AI助手初始化完成，配备 %d 个工具\n", len(assistant.GetTools()))

	// 7. 创建和执行任务
	runDemoTasks(assistant)

	fmt.Println("\n🎉 快速开始示例完成！")
	fmt.Println("✨ 接下来可以尝试：")
	fmt.Println("   - 运行完整示例: go run ai_research_assistant.go")
	fmt.Println("   - 查看更多示例: ls ../")
	fmt.Println("   - 阅读文档: cat README.md")
}

// 运行演示任务
func runDemoTasks(assistant agent.Agent) {
	ctx := context.Background()

	// 任务1: 简单对话
	fmt.Println("\n📝 任务1: 简单对话")
	task1 := agent.NewBaseTask(
		"请用一句话解释什么是人工智能",
		"一句话的AI定义",
	)

	output1, err := runTaskWithTimeout(ctx, assistant, task1, 30*time.Second)
	if err != nil {
		fmt.Printf("❌ 任务1失败: %v\n", err)
	} else {
		fmt.Printf("🤖 回答: %s\n", output1.Raw)
		fmt.Printf("📊 Token使用: %d\n", output1.TokensUsed)
	}

	// 任务2: 使用工具计算
	fmt.Println("\n🔢 任务2: 数学计算")
	task2 := agent.NewBaseTask(
		"计算 (25 + 15) × 3 - 8 的结果",
		"数学计算结果",
	)

	output2, err := runTaskWithTimeout(ctx, assistant, task2, 30*time.Second)
	if err != nil {
		fmt.Printf("❌ 任务2失败: %v\n", err)
	} else {
		fmt.Printf("🤖 回答: %s\n", output2.Raw)
		fmt.Printf("📊 Token使用: %d\n", output2.TokensUsed)
	}

	// 任务3: 复杂推理
	fmt.Println("\n🧠 任务3: 复杂推理")
	task3 := agent.NewBaseTask(
		"分析：如果一个公司今年营收是1000万，同比增长了25%，那么去年的营收是多少？请详细解释计算过程。",
		"详细的计算过程和结果",
	)

	output3, err := runTaskWithTimeout(ctx, assistant, task3, 45*time.Second)
	if err != nil {
		fmt.Printf("❌ 任务3失败: %v\n", err)
	} else {
		fmt.Printf("🤖 分析:\n%s\n", output3.Raw)
		fmt.Printf("📊 Token使用: %d\n", output3.TokensUsed)
	}

	// 显示总统计
	tools := assistant.GetTools()
	fmt.Printf("\n📈 会话统计:\n")
	for _, tool := range tools {
		fmt.Printf("   - %s: 使用了 %d 次\n", tool.GetName(), tool.GetUsageCount())
	}
}

// 带超时的任务执行
func runTaskWithTimeout(ctx context.Context, agent agent.Agent, task agent.Task, timeout time.Duration) (*agent.TaskOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	output, err := agent.Execute(ctx, task)
	duration := time.Since(start)

	if err == nil {
		fmt.Printf("⏱️  执行时间: %v\n", duration)
	}

	return output, err
}
