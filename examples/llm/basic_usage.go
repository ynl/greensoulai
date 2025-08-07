package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ynl/greensoulai/internal/llm"
)

func main() {
	fmt.Println("🚀 GreenSoulAI LLM 模块演示")
	fmt.Println("=========================")

	// 1. 创建LLM配置
	config := &llm.Config{
		Provider:    "openai",
		Model:       "gpt-4o-mini",
		APIKey:      "your-api-key-here", // 在实际使用中设置真实的API密钥
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		Temperature: func() *float64 { t := 0.7; return &t }(),
		MaxTokens:   func() *int { t := 1000; return &t }(),
	}

	// 2. 创建LLM实例
	llmInstance, err := llm.CreateLLM(config)
	if err != nil {
		fmt.Printf("❌ 创建LLM失败: %v\n", err)
		return
	}
	defer llmInstance.Close()

	fmt.Printf("✅ 成功创建 %s 模型实例\n", llmInstance.GetModel())
	fmt.Printf("🎯 支持函数调用: %v\n", llmInstance.SupportsFunctionCalling())
	fmt.Printf("📏 上下文窗口: %d tokens\n", llmInstance.GetContextWindowSize())

	// 3. 基础对话示例
	fmt.Println("\n📝 基础对话演示:")
	messages := []llm.Message{
		{Role: llm.RoleSystem, Content: "你是一个有用的AI助手。"},
		{Role: llm.RoleUser, Content: "请用一句话解释什么是Go语言。"},
	}

	ctx := context.Background()
	response, err := llmInstance.Call(ctx, messages, &llm.CallOptions{
		Temperature: func() *float64 { t := 0.5; return &t }(),
		MaxTokens:   func() *int { t := 100; return &t }(),
	})

	if err != nil {
		fmt.Printf("❌ 调用失败: %v\n", err)
	} else {
		fmt.Printf("🤖 回复: %s\n", response.Content)
		fmt.Printf("📊 使用统计: %d tokens (提示: %d, 完成: %d)\n",
			response.Usage.TotalTokens,
			response.Usage.PromptTokens,
			response.Usage.CompletionTokens)
	}

	// 4. 流式响应演示
	fmt.Println("\n🌊 流式响应演示:")
	streamMessages := []llm.Message{
		{Role: llm.RoleUser, Content: "请简单介绍一下人工智能的发展历程。"},
	}

	stream, err := llmInstance.CallStream(ctx, streamMessages, &llm.CallOptions{
		Temperature: func() *float64 { t := 0.7; return &t }(),
		MaxTokens:   func() *int { t := 200; return &t }(),
	})

	if err != nil {
		fmt.Printf("❌ 流式调用失败: %v\n", err)
	} else {
		fmt.Print("🤖 流式回复: ")
		for chunk := range stream {
			if chunk.Error != nil {
				fmt.Printf("\n❌ 流式错误: %v\n", chunk.Error)
				break
			}
			fmt.Print(chunk.Delta)
		}
		fmt.Println()
	}

	// 5. 工具调用演示（函数调用）
	fmt.Println("\n🔧 工具调用演示:")
	tools := []llm.Tool{
		{
			Type: "function",
			Function: llm.ToolSchema{
				Name:        "get_weather",
				Description: "获取指定城市的天气信息",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"city": map[string]interface{}{
							"type":        "string",
							"description": "城市名称",
						},
					},
					"required": []string{"city"},
				},
			},
		},
	}

	toolMessages := []llm.Message{
		{Role: llm.RoleUser, Content: "北京今天天气怎么样？"},
	}

	toolResponse, err := llmInstance.Call(ctx, toolMessages, &llm.CallOptions{
		Tools:      tools,
		ToolChoice: "auto",
	})

	if err != nil {
		fmt.Printf("❌ 工具调用失败: %v\n", err)
	} else if len(toolResponse.ToolCalls) > 0 {
		fmt.Printf("🔧 工具调用: %s(%s)\n",
			toolResponse.ToolCalls[0].Function.Name,
			toolResponse.ToolCalls[0].Function.Arguments)
	} else {
		fmt.Printf("🤖 普通回复: %s\n", toolResponse.Content)
	}

	// 6. 展示不同的配置选项
	fmt.Println("\n⚙️  配置选项演示:")
	fmt.Println("   - 高创造性 (Temperature: 0.9)")
	fmt.Println("   - 短回复 (MaxTokens: 50)")
	fmt.Println("   - 停止序列: [\"\\n\"]")

	optionsResponse, err := llmInstance.Call(ctx, []llm.Message{
		{Role: llm.RoleUser, Content: "用一句话介绍人工智能"},
	}, &llm.CallOptions{
		Temperature:   func() *float64 { t := 0.9; return &t }(),
		MaxTokens:     func() *int { t := 50; return &t }(),
		StopSequences: []string{"\n"},
	})

	if err != nil {
		fmt.Printf("❌ 配置选项演示失败: %v\n", err)
	} else {
		fmt.Printf("🤖 高创造性回复: %s\n", optionsResponse.Content)
		fmt.Printf("📊 使用Token: %d\n", optionsResponse.Usage.TotalTokens)
	}

	configResponse, err := llmInstance.Call(ctx, []llm.Message{
		{Role: llm.RoleUser, Content: "用一个词描述Go语言"},
	}, llm.DefaultCallOptions())

	// 应用选项
	configResponse2, _ := llmInstance.Call(ctx, []llm.Message{
		{Role: llm.RoleUser, Content: "用一个词描述Go语言"},
	}, llm.DefaultCallOptions())

	if configResponse2 != nil {
		configResponse2.Usage = llm.Usage{} // 避免空指针
	}

	if err == nil && configResponse != nil {
		fmt.Printf("🎛️  配置演示完成\n")
	}

	fmt.Println("\n🎉 LLM模块演示完成！")
	fmt.Println("✨ 特性包括:")
	fmt.Println("   - 统一的LLM接口抽象")
	fmt.Println("   - 多提供商支持")
	fmt.Println("   - 流式响应")
	fmt.Println("   - 函数调用/工具集成")
	fmt.Println("   - 详细的使用统计")
	fmt.Println("   - 强大的错误处理")
	fmt.Println("   - 事件系统集成")
}
