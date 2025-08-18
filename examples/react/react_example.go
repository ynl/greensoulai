package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ynl/greensoulai/internal/agent"
	"github.com/ynl/greensoulai/internal/llm"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// 这个示例展示了如何使用greensoulai的ReAct模式
func main() {
	// 1. 创建必要的组件
	eventBus := events.NewEventBus(logger.NewConsoleLogger())
	agentLogger := logger.NewConsoleLogger()

	// 2. 创建LLM实例（这里使用模拟LLM，实际使用中应该是真实的LLM）
	mockLLM := createMockLLM()

	// 3. 配置Agent使用ReAct模式
	config := agent.AgentConfig{
		Role:      "Research Assistant",
		Goal:      "Help users conduct thorough research and analysis",
		Backstory: "I am a knowledgeable research assistant with expertise in analysis and investigation",
		LLM:       mockLLM,
		EventBus:  eventBus,
		Logger:    agentLogger,
		ExecutionConfig: agent.ExecutionConfig{
			MaxIterations:    15,
			MaxRPM:           60,
			Timeout:          30 * time.Minute,
			MaxExecutionTime: 10 * time.Minute,
			AllowDelegation:  false,
			VerboseLogging:   true,
			HumanInput:       false,
			UseSystemPrompt:  true,
			MaxTokens:        4096,
			Temperature:      0.7,
			CacheEnabled:     true,
			MaxRetryLimit:    3,
			Mode:             agent.ModeReAct, // 启用ReAct模式
			ReActConfig: &agent.ReActConfig{
				MaxIterations:          10,
				ThoughtTimeout:         30 * time.Second,
				EnableDebugOutput:      true,
				StrictFormatValidation: true,
				AllowFallbackToJSON:    true,
			},
		},
	}

	// 4. 创建Agent
	myAgent, err := agent.NewBaseAgent(config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create agent: %v", err))
	}

	// 5. 添加工具（可选）
	calculatorTool := agent.NewCalculatorTool()
	if err := myAgent.AddTool(calculatorTool); err != nil {
		panic(fmt.Sprintf("Failed to add tool: %v", err))
	}

	// 6. 创建任务
	task := agent.NewBaseTask(
		"研究人工智能在2024年的发展趋势，并分析其对未来5年技术发展的影响",
		"提供一份详细的分析报告，包含具体数据和趋势分析",
	)

	// 7. 使用ReAct模式执行任务
	ctx := context.Background()

	fmt.Println("🤖 启动ReAct模式执行...")
	fmt.Println("====================================================")

	output, trace, err := myAgent.ExecuteWithReAct(ctx, task)
	if err != nil {
		panic(fmt.Sprintf("Task execution failed: %v", err))
	}

	// 8. 显示结果
	fmt.Println("\n📊 执行结果:")
	fmt.Printf("最终答案: %s\n", output.Raw)
	fmt.Printf("执行时间: %v\n", output.ExecutionTime)
	fmt.Printf("模式: %s\n", output.Metadata["mode"])

	fmt.Println("\n🔍 ReAct轨迹:")
	for i, step := range trace.Steps {
		fmt.Printf("\n步骤 %d:\n", i+1)
		fmt.Printf("  思考: %s\n", step.Thought)
		if step.Action != "" {
			fmt.Printf("  动作: %s\n", step.Action)
			fmt.Printf("  输入: %v\n", step.ActionInput)
			fmt.Printf("  观察: %s\n", step.Observation)
		}
		if step.FinalAnswer != "" {
			fmt.Printf("  最终答案: %s\n", step.FinalAnswer)
		}
		if step.Error != "" {
			fmt.Printf("  错误: %s\n", step.Error)
		}
	}

	fmt.Printf("\n✅ 总步骤数: %d\n", trace.IterationCount)
	fmt.Printf("总耗时: %v\n", trace.TotalDuration)
	fmt.Printf("完成状态: %v\n", trace.IsCompleted)

	// 9. 演示模式切换
	fmt.Println("\n🔄 演示模式切换:")
	fmt.Printf("当前模式: %s\n", myAgent.GetCurrentMode().String())

	// 切换到JSON模式
	myAgent.SetReActMode(false)
	fmt.Printf("切换后模式: %s\n", myAgent.GetCurrentMode().String())

	// 切换回ReAct模式
	myAgent.SetReActMode(true)
	fmt.Printf("再次切换后模式: %s\n", myAgent.GetCurrentMode().String())

	// 10. 展示统计信息
	stats := myAgent.GetExecutionStats()
	fmt.Println("\n📈 Agent执行统计:")
	fmt.Printf("总执行次数: %d\n", stats.TotalExecutions)
	fmt.Printf("成功次数: %d\n", stats.SuccessfulExecutions)
	fmt.Printf("失败次数: %d\n", stats.FailedExecutions)
	fmt.Printf("平均执行时间: %v\n", stats.AverageExecutionTime)

	fmt.Println("\n🎉 ReAct模式演示完成!")
}

// createMockLLM 创建用于演示的模拟LLM
func createMockLLM() llm.LLM {
	// 这里创建一个模拟LLM，实际使用中应该是真实的LLM实现
	responses := []llm.Response{
		{
			Content: `Thought: 我需要分析人工智能在2024年的发展趋势，这是一个复杂的研究任务
Action: calculator  
Action Input: {"operation": "add", "a": 2024, "b": 5}`,
			Model: "demo-model",
			Usage: llm.Usage{TotalTokens: 50},
		},
		{
			Content: `Thought: 我已经计算了未来5年的时间范围，现在我需要基于我的知识提供详细分析
Final Answer: 基于对人工智能发展的分析，2024年呈现以下关键趋势：

1. **大语言模型的成熟化**: GPT-4和其他先进模型在各行业中得到广泛应用
2. **多模态AI的兴起**: 文本、图像、音频、视频的统一处理能力显著提升  
3. **AI基础设施的完善**: 云端AI服务和边缘计算能力大幅增强
4. **监管框架的建立**: 各国开始制定AI治理法规和伦理标准

未来5年(2025-2029)影响预测：
- AI将成为企业数字化转型的核心驱动力
- 预计AI市场规模将从当前的1500亿美元增长到5000亿美元
- 自动化程度将显著提升，同时催生新的工作岗位
- AI安全和可解释性将成为技术发展重点

这一分析基于当前技术发展轨迹和市场趋势，为决策提供参考。`,
			Model: "demo-model",
			Usage: llm.Usage{TotalTokens: 200},
		},
	}

	return &mockLLMImpl{responses: responses, currentIndex: 0}
}

// mockLLMImpl 模拟LLM实现
type mockLLMImpl struct {
	responses    []llm.Response
	currentIndex int
}

func (m *mockLLMImpl) Call(ctx context.Context, messages []llm.Message, options *llm.CallOptions) (*llm.Response, error) {
	if m.currentIndex >= len(m.responses) {
		m.currentIndex = len(m.responses) - 1
	}

	response := m.responses[m.currentIndex]
	m.currentIndex++

	// 模拟处理时间
	time.Sleep(100 * time.Millisecond)

	return &response, nil
}

func (m *mockLLMImpl) Stream(ctx context.Context, messages []llm.Message, options *llm.CallOptions) (<-chan llm.StreamResponse, error) {
	// 简单的流实现
	ch := make(chan llm.StreamResponse, 1)
	go func() {
		defer close(ch)
		response, _ := m.Call(ctx, messages, options)
		ch <- llm.StreamResponse{Delta: response.Content}
	}()
	return ch, nil
}

func (m *mockLLMImpl) CallStream(ctx context.Context, messages []llm.Message, options *llm.CallOptions) (<-chan llm.StreamResponse, error) {
	return m.Stream(ctx, messages, options)
}

// LLM接口的其他必需方法
func (m *mockLLMImpl) GetModel() string                     { return "demo-model" }
func (m *mockLLMImpl) SupportsFunctionCalling() bool        { return false }
func (m *mockLLMImpl) GetContextWindowSize() int            { return 4096 }
func (m *mockLLMImpl) SetEventBus(eventBus events.EventBus) {}
func (m *mockLLMImpl) Close() error                         { return nil }

// 确保mockLLMImpl实现了LLM接口
var _ llm.LLM = (*mockLLMImpl)(nil)
