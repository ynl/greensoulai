package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ynl/greensoulai/internal/llm"
	"github.com/ynl/greensoulai/pkg/logger"
)

// 简化的Kimi API测试
func main() {
	fmt.Println("🚀 OpenRouter Kimi API 简化测试")
	fmt.Println("=================================")

	// 从环境变量读取OpenRouter API密钥
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

	// 创建Kimi LLM实例
	fmt.Println("🤖 创建Kimi模型实例...")
	kimiLLM := llm.NewOpenAILLM("moonshotai/kimi-k2:free",
		llm.WithAPIKey(apiKey),
		llm.WithBaseURL("https://openrouter.ai/api/v1"),
		llm.WithTimeout(30*time.Second),
		llm.WithCustomHeader("HTTP-Referer", "https://github.com/ynl/greensoulai"),
		llm.WithCustomHeader("X-Title", "GreenSoulAI Kimi测试"),
		llm.WithLogger(logger.NewConsoleLogger()),
	)

	fmt.Printf("✅ 模型创建成功: %s\n", kimiLLM.GetModel())
	fmt.Printf("📏 上下文窗口: %d tokens\n", kimiLLM.GetContextWindowSize())
	fmt.Printf("🎯 支持函数调用: %v\n", kimiLLM.SupportsFunctionCalling())

	// 测试1: 基本中文对话
	fmt.Println("\n💬 测试1: 基本中文对话")
	testBasicChat(kimiLLM)

	// 测试2: 数学推理
	fmt.Println("\n🧮 测试2: 数学推理")
	testMathReasoning(kimiLLM)

	// 测试3: 文学理解
	fmt.Println("\n📚 测试3: 文学理解")
	testLiteratureUnderstanding(kimiLLM)

	fmt.Println("\n🎉 所有测试完成！")
}

func testBasicChat(llmInstance llm.LLM) {
	ctx := context.Background()

	messages := []llm.Message{
		{Role: llm.RoleSystem, Content: "你是一个友好的AI助手，请用中文回答。"},
		{Role: llm.RoleUser, Content: "你好，请简单介绍一下你自己。"},
	}

	response, err := llmInstance.Call(ctx, messages, &llm.CallOptions{
		Temperature: func() *float64 { t := 0.7; return &t }(),
		MaxTokens:   func() *int { t := 300; return &t }(),
	})

	if err != nil {
		fmt.Printf("❌ 基本对话失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 对话成功!\n")
	fmt.Printf("🤖 回复: %s\n", response.Content)
	fmt.Printf("📊 统计: %d tokens (提示: %d, 完成: %d)\n",
		response.Usage.TotalTokens,
		response.Usage.PromptTokens,
		response.Usage.CompletionTokens)
}

func testMathReasoning(llmInstance llm.LLM) {
	ctx := context.Background()

	messages := []llm.Message{
		{Role: llm.RoleUser, Content: "小明有10个苹果，给了小红3个，给了小李2个，然后妈妈又给了他5个。请问小明现在有几个苹果？请逐步计算。"},
	}

	response, err := llmInstance.Call(ctx, messages, &llm.CallOptions{
		Temperature: func() *float64 { t := 0.1; return &t }(), // 低温度确保准确性
		MaxTokens:   func() *int { t := 300; return &t }(),
	})

	if err != nil {
		fmt.Printf("❌ 数学推理失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 数学推理成功!\n")
	fmt.Printf("🔢 计算过程: %s\n", response.Content)
	fmt.Printf("📊 统计: %d tokens\n", response.Usage.TotalTokens)
}

func testLiteratureUnderstanding(llmInstance llm.LLM) {
	ctx := context.Background()

	messages := []llm.Message{
		{Role: llm.RoleUser, Content: "请解释这句古诗的含义：'山重水复疑无路，柳暗花明又一村'"},
	}

	response, err := llmInstance.Call(ctx, messages, &llm.CallOptions{
		Temperature: func() *float64 { t := 0.8; return &t }(),
		MaxTokens:   func() *int { t := 400; return &t }(),
	})

	if err != nil {
		fmt.Printf("❌ 文学理解失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 文学理解成功!\n")
	fmt.Printf("🎨 诗意解释: %s\n", response.Content)
	fmt.Printf("📊 统计: %d tokens\n", response.Usage.TotalTokens)
}
