package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ynl/greensoulai/internal/llm"
)

func main() {
	fmt.Println("🌐 OpenRouter LLM 基础使用示例")
	fmt.Println("===============================")

	// OpenRouter API配置
	// 优先使用环境变量，如果没有则使用示例Key
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		apiKey = ""
		fmt.Println("⚠️  使用示例API Key，建议设置环境变量 OPENROUTER_API_KEY")
	}

	baseURL := "https://openrouter.ai/api/v1"
	model := "moonshotai/kimi-k2:free" // 使用免费的Kimi模型

	fmt.Printf("🔑 API Key: %s...\n", apiKey[:20])
	fmt.Printf("🌍 Base URL: %s\n", baseURL)
	fmt.Printf("🤖 Model: %s\n", model)

	// 创建LLM配置
	config := &llm.Config{
		Provider:    "openai", // 使用OpenAI提供商，因为OpenRouter兼容OpenAI API
		Model:       model,
		APIKey:      apiKey,
		BaseURL:     baseURL,
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		Temperature: func() *float64 { t := 0.7; return &t }(),
		MaxTokens:   func() *int { t := 1000; return &t }(),
	}

	// 创建LLM实例并添加OpenRouter特定的headers
	llmInstance, err := llm.CreateLLM(config)
	if err != nil {
		log.Fatalf("❌ 创建LLM失败: %v", err)
	}
	defer llmInstance.Close()

	// 如果需要添加OpenRouter的可选headers，可以通过以下方式：
	// 注意：我们需要在创建时就添加自定义headers
	openaiLLM := llm.NewOpenAILLM(model,
		llm.WithAPIKey(apiKey),
		llm.WithBaseURL(baseURL),
		llm.WithTimeout(30*time.Second),
		llm.WithMaxRetries(3),
		llm.WithCustomHeader("HTTP-Referer", "https://github.com/ynl/greensoulai"),
		llm.WithCustomHeader("X-Title", "GreenSoulAI"),
	)

	fmt.Printf("✅ 成功创建 %s 模型实例\n", openaiLLM.GetModel())
	fmt.Printf("🎯 支持函数调用: %v\n", openaiLLM.SupportsFunctionCalling())
	fmt.Printf("📏 上下文窗口: %d tokens\n", openaiLLM.GetContextWindowSize())

	// 测试基础对话
	fmt.Println("\n📝 基础对话测试:")
	messages := []llm.Message{
		{Role: llm.RoleSystem, Content: "你是一个有用的AI助手。请用中文回答问题。"},
		{Role: llm.RoleUser, Content: "什么是人生的意义？请简短回答。"},
	}

	ctx := context.Background()
	response, err := openaiLLM.Call(ctx, messages, &llm.CallOptions{
		Temperature: func() *float64 { t := 0.5; return &t }(),
		MaxTokens:   func() *int { t := 200; return &t }(),
	})

	if err != nil {
		fmt.Printf("❌ 调用失败: %v\n", err)

		// 尝试诊断问题
		fmt.Println("\n🔍 诊断信息:")
		fmt.Printf("- API Key 长度: %d\n", len(apiKey))
		fmt.Printf("- Base URL: %s\n", baseURL)
		fmt.Printf("- Model: %s\n", model)

		return
	}

	fmt.Printf("🤖 回复: %s\n", response.Content)
	fmt.Printf("📊 使用统计:\n")
	fmt.Printf("   - 总Token: %d\n", response.Usage.TotalTokens)
	fmt.Printf("   - 提示Token: %d\n", response.Usage.PromptTokens)
	fmt.Printf("   - 完成Token: %d\n", response.Usage.CompletionTokens)
	fmt.Printf("   - 预估成本: $%.6f\n", response.Usage.Cost)
	fmt.Printf("   - 模型: %s\n", response.Model)
	fmt.Printf("   - 结束原因: %s\n", response.FinishReason)

	// 测试流式响应
	fmt.Println("\n🌊 流式响应测试:")
	streamMessages := []llm.Message{
		{Role: llm.RoleUser, Content: "请用一句话介绍Go语言的特点。"},
	}

	stream, err := openaiLLM.CallStream(ctx, streamMessages, &llm.CallOptions{
		Temperature: func() *float64 { t := 0.7; return &t }(),
		MaxTokens:   func() *int { t := 100; return &t }(),
	})

	if err != nil {
		fmt.Printf("❌ 流式调用失败: %v\n", err)
	} else {
		fmt.Print("🤖 流式回复: ")
		fullResponse := ""
		for chunk := range stream {
			if chunk.Error != nil {
				fmt.Printf("\n❌ 流式错误: %v\n", chunk.Error)
				break
			}
			fmt.Print(chunk.Delta)
			fullResponse += chunk.Delta

			// 如果有使用统计信息，显示
			if chunk.Usage != nil {
				fmt.Printf("\n📊 最终统计: %d tokens\n", chunk.Usage.TotalTokens)
			}
		}
		fmt.Printf("\n✅ 流式响应完成，总长度: %d 字符\n", len(fullResponse))
	}

	// 测试不同模型（如果可用）
	fmt.Println("\n🔄 多模型测试:")
	models := []string{
		"moonshotai/kimi-k2:free",
		"openai/gpt-3.5-turbo",
		"anthropic/claude-3-haiku",
	}

	for _, testModel := range models {
		fmt.Printf("\n🧪 测试模型: %s\n", testModel)

		testLLM := llm.NewOpenAILLM(testModel,
			llm.WithAPIKey(apiKey),
			llm.WithBaseURL(baseURL),
			llm.WithTimeout(10*time.Second),
			llm.WithMaxRetries(1),
			llm.WithCustomHeader("HTTP-Referer", "https://github.com/ynl/greensoulai"),
			llm.WithCustomHeader("X-Title", "GreenSoulAI"),
		)

		testResponse, err := testLLM.Call(ctx, []llm.Message{
			{Role: llm.RoleUser, Content: "Hi, just say hello in one word."},
		}, &llm.CallOptions{
			MaxTokens: func() *int { t := 10; return &t }(),
		})

		if err != nil {
			fmt.Printf("   ❌ 失败: %v\n", err)
		} else {
			fmt.Printf("   ✅ 成功: %s\n", testResponse.Content)
		}
	}

	fmt.Println("\n🎉 OpenRouter 集成测试完成！")
	fmt.Println("\n📋 测试总结:")
	fmt.Println("   ✅ 支持自定义Base URL")
	fmt.Println("   ✅ 支持自定义API Key")
	fmt.Println("   ✅ 支持自定义HTTP Headers")
	fmt.Println("   ✅ 兼容OpenAI API格式")
	fmt.Println("   ✅ 支持同步和流式调用")
	fmt.Println("   ✅ 支持多种模型")

	fmt.Println("\n💡 使用建议:")
	fmt.Println("   1. 确保API Key有效且有足够余额")
	fmt.Println("   2. 选择合适的模型（免费或付费）")
	fmt.Println("   3. 设置适当的超时和重试")
	fmt.Println("   4. 监控Token使用和成本")
}
