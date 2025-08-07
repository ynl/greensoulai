# LLM模块使用指南

## 概述

GreenSoulAI的LLM模块提供了一个统一、强大且易于使用的接口来集成各种语言模型提供商。模块采用现代Go设计模式，提供了生产级的功能和性能。

## 核心特性

### 🎯 **统一接口**
- 统一的LLM接口，支持多种提供商
- 一致的API设计，降低学习成本
- 易于扩展新的LLM提供商

### ⚡ **高性能**
- 异步和流式响应支持
- 智能连接池管理
- 自动重试和错误恢复

### 🛠️ **丰富功能**
- 完整的函数调用支持
- 详细的使用统计和成本追踪
- 事件系统集成
- 灵活的配置选项

### 🔒 **生产就绪**
- 强类型安全
- 完整的错误处理
- 全面的单元测试覆盖
- 详细的日志记录

## 快速开始

### 1. 基础使用

```go
package main

import (
    "context"
    "github.com/ynl/greensoulai/internal/llm"
)

func main() {
    // 创建配置
    config := &llm.Config{
        Provider: "openai",
        Model:    "gpt-4o-mini",
        APIKey:   "your-api-key",
    }

    // 创建LLM实例
    llmInstance, err := llm.CreateLLM(config)
    if err != nil {
        panic(err)
    }
    defer llmInstance.Close()

    // 发送消息
    messages := []llm.Message{
        {Role: llm.RoleUser, Content: "Hello, world!"},
    }

    response, err := llmInstance.Call(context.Background(), messages, nil)
    if err != nil {
        panic(err)
    }

    fmt.Println("回复:", response.Content)
}
```

### 2. 流式响应

```go
// 创建流式请求
stream, err := llmInstance.CallStream(ctx, messages, nil)
if err != nil {
    panic(err)
}

// 处理流式数据
for chunk := range stream {
    if chunk.Error != nil {
        fmt.Println("错误:", chunk.Error)
        break
    }
    fmt.Print(chunk.Delta)
}
```

### 3. 函数调用

```go
// 定义工具
tools := []llm.Tool{
    {
        Type: "function",
        Function: llm.ToolSchema{
            Name:        "get_weather",
            Description: "获取天气信息",
            Parameters: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "city": map[string]interface{}{
                        "type": "string",
                        "description": "城市名称",
                    },
                },
            },
        },
    },
}

// 发送带工具的请求
response, err := llmInstance.Call(ctx, messages, &llm.CallOptions{
    Tools:      tools,
    ToolChoice: "auto",
})

// 检查工具调用
if len(response.ToolCalls) > 0 {
    toolCall := response.ToolCalls[0]
    fmt.Printf("调用工具: %s\n", toolCall.Function.Name)
    fmt.Printf("参数: %s\n", toolCall.Function.Arguments)
}
```

## 配置选项

### CallOptions

```go
options := &llm.CallOptions{
    Temperature:      &[]float64{0.7}[0],  // 创造性控制
    MaxTokens:        &[]int{1000}[0],     // 最大输出长度
    TopP:             &[]float64{0.9}[0],  // 核心采样
    FrequencyPenalty: &[]float64{0.5}[0],  // 频率惩罚
    PresencePenalty:  &[]float64{0.2}[0],  // 存在惩罚
    StopSequences:    []string{"STOP"},    // 停止序列
    Stream:           true,                // 流式响应
}
```

### 函数式选项

```go
// 使用函数式选项
response, err := llmInstance.Call(ctx, messages, 
    llm.DefaultCallOptions().ApplyOptions(
        llm.WithTemperature(0.8),
        llm.WithMaxTokens(500),
        llm.WithStream(true),
    ),
)
```

## 支持的模型

### OpenAI
- gpt-4, gpt-4-turbo, gpt-4o, gpt-4o-mini
- gpt-3.5-turbo, gpt-3.5-turbo-16k
- 自动上下文窗口检测

### OpenRouter
- 支持200+种模型，包括免费模型
- 统一API访问多种提供商
- 完全兼容OpenAI API格式
- 内置成本优化和模型路由

### 未来支持
- Anthropic Claude (直接API)
- Google Gemini (直接API)
- 本地模型支持

## 事件系统

LLM模块集成了完整的事件系统，可以监控整个调用生命周期：

```go
// 事件类型
- llm_call_started      // 调用开始
- llm_call_completed    // 调用完成
- llm_call_failed       // 调用失败
- llm_stream_started    // 流式开始
- llm_stream_chunk      // 流式数据块
- llm_stream_ended      // 流式结束
```

## 错误处理

模块提供了完整的错误处理机制：

```go
// 自动重试配置
config := &llm.Config{
    MaxRetries: 3,
    Timeout:    30 * time.Second,
}

// 错误类型检查
if err != nil {
    switch {
    case strings.Contains(err.Error(), "context"):
        // 上下文超时
    case strings.Contains(err.Error(), "API"):
        // API错误
    default:
        // 其他错误
    }
}
```

## 成本管理

内置成本追踪功能：

```go
response, err := llmInstance.Call(ctx, messages, options)
if err == nil {
    fmt.Printf("使用Token: %d\n", response.Usage.TotalTokens)
    fmt.Printf("预估成本: $%.4f\n", response.Usage.Cost)
}
```

## 最佳实践

### 1. 连接管理
```go
// 复用LLM实例，避免频繁创建
var globalLLM llm.LLM

func init() {
    var err error
    globalLLM, err = llm.CreateLLM(config)
    if err != nil {
        panic(err)
    }
}

// 程序结束时清理
defer globalLLM.Close()
```

### 2. 上下文管理
```go
// 设置合理的超时
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

response, err := llmInstance.Call(ctx, messages, options)
```

### 3. 错误处理
```go
// 实现重试逻辑
for i := 0; i < 3; i++ {
    response, err := llmInstance.Call(ctx, messages, options)
    if err == nil {
        break
    }
    
    if i == 2 {
        return err // 最后一次尝试失败
    }
    
    time.Sleep(time.Duration(i+1) * time.Second)
}
```

## 性能调优

### 1. 批量处理
- 合并多个小请求
- 使用流式响应减少延迟
- 合理设置Token限制

### 2. 缓存策略  
- 实现响应缓存
- 使用相似度检测避免重复请求

### 3. 监控指标
- 监控Token使用量
- 跟踪响应时间
- 监控错误率

## 扩展开发

### 添加新的LLM提供商

```go
// 1. 实现Provider接口
type CustomProvider struct{}

func (p *CustomProvider) Name() string {
    return "custom"
}

func (p *CustomProvider) CreateLLM(config map[string]interface{}) (llm.LLM, error) {
    // 实现创建逻辑
}

func (p *CustomProvider) SupportedModels() []string {
    return []string{"custom-model-1", "custom-model-2"}
}

// 2. 注册提供商
llm.RegisterProvider(&CustomProvider{})
```

### 实现自定义LLM

```go
// 实现LLM接口
type CustomLLM struct {
    *llm.BaseLLM
}

func (c *CustomLLM) Call(ctx context.Context, messages []llm.Message, options *llm.CallOptions) (*llm.Response, error) {
    // 实现调用逻辑
}

func (c *CustomLLM) CallStream(ctx context.Context, messages []llm.Message, options *llm.CallOptions) (<-chan llm.StreamResponse, error) {
    // 实现流式调用逻辑
}
```

## 运行示例

```bash
# 运行基础示例
cd examples/llm/basic
export OPENAI_API_KEY="your-api-key-here"
go run main.go

# OpenRouter基础示例
cd examples/llm/openrouter/basic
export OPENROUTER_API_KEY="sk-or-v1-your-key-here"
go run main.go

# OpenRouter高级功能示例
cd examples/llm/openrouter/advanced
go run main.go
```

## OpenRouter集成

OpenRouter是一个统一的LLM API网关，支持200+种模型，包括免费选项。我们的LLM模块完全支持OpenRouter：

### 基础配置

```go
// 使用OpenRouter
llmInstance := llm.NewOpenAILLM("moonshotai/kimi-k2:free",
    llm.WithAPIKey("sk-or-v1-your-openrouter-key"),
    llm.WithBaseURL("https://openrouter.ai/api/v1"),
    llm.WithCustomHeader("HTTP-Referer", "https://your-site.com"),
    llm.WithCustomHeader("X-Title", "Your App Name"),
)
```

### 推荐的免费模型
- `moonshotai/kimi-k2:free` - 中文优化，免费
- `google/gemini-flash-1.5` - 快速响应
- `meta-llama/llama-3.1-8b-instruct:free` - 开源模型

### 自定义Headers
OpenRouter支持可选的排名headers：
- `HTTP-Referer`: 您的网站URL（用于openrouter.ai排名）
- `X-Title`: 您的应用名称（用于openrouter.ai排名）

## 测试

```bash
# 运行所有测试
go test ./internal/llm/... -v

# 运行特定测试
go test ./internal/llm -run TestOpenAILLM -v

# 生成测试覆盖率报告
go test ./internal/llm/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

**注意**: 在生产环境中使用时，请确保：
1. 设置正确的API密钥和配置
2. 实施适当的速率限制
3. 监控成本和使用量
4. 处理所有可能的错误情况
5. 定期更新依赖和安全补丁
