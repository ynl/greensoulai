package agent

import (
	"context"
	"testing"
	"time"

	"github.com/ynl/greensoulai/internal/llm"
	"github.com/ynl/greensoulai/pkg/logger"
)

// MockLLM和其他测试辅助对象在 mock_test.go 中集中管理

// 测试Agent创建
func TestNewBaseAgent(t *testing.T) {
	tests := []struct {
		name        string
		config      AgentConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: AgentConfig{
				Role:      "Test Agent",
				Goal:      "Test goal",
				Backstory: "Test backstory",
			},
			expectError: false,
		},
		{
			name: "missing role",
			config: AgentConfig{
				Goal:      "Test goal",
				Backstory: "Test backstory",
			},
			expectError: true,
		},
		{
			name: "missing goal",
			config: AgentConfig{
				Role:      "Test Agent",
				Backstory: "Test backstory",
			},
			expectError: true,
		},
		{
			name: "missing backstory",
			config: AgentConfig{
				Role: "Test Agent",
				Goal: "Test goal",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := NewBaseAgent(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if agent == nil {
				t.Errorf("expected agent, got nil")
				return
			}

			// 验证基础属性
			if agent.GetRole() != tt.config.Role {
				t.Errorf("expected role %s, got %s", tt.config.Role, agent.GetRole())
			}
			if agent.GetGoal() != tt.config.Goal {
				t.Errorf("expected goal %s, got %s", tt.config.Goal, agent.GetGoal())
			}
			if agent.GetBackstory() != tt.config.Backstory {
				t.Errorf("expected backstory %s, got %s", tt.config.Backstory, agent.GetBackstory())
			}
		})
	}
}

// 测试Agent初始化
func TestBaseAgent_Initialize(t *testing.T) {
	mockLLM := NewMockLLM(&llm.Response{
		Content: "Test response",
		Usage: llm.Usage{
			TotalTokens: 10,
		},
	}, false)

	config := AgentConfig{
		Role:      "Test Agent",
		Goal:      "Test goal",
		Backstory: "Test backstory",
		LLM:       mockLLM,
	}

	agent, err := NewBaseAgent(config)
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	// 测试初始化
	err = agent.Initialize()
	if err != nil {
		t.Errorf("initialization failed: %v", err)
	}

	// 测试重复初始化
	err = agent.Initialize()
	if err != nil {
		t.Errorf("repeated initialization should not fail: %v", err)
	}
}

// 测试Agent初始化失败（缺少LLM）
func TestBaseAgent_Initialize_MissingLLM(t *testing.T) {
	config := AgentConfig{
		Role:      "Test Agent",
		Goal:      "Test goal",
		Backstory: "Test backstory",
		// 没有LLM
	}

	agent, err := NewBaseAgent(config)
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	err = agent.Initialize()
	if err == nil {
		t.Errorf("expected initialization to fail without LLM")
	}
}

// 测试简单任务执行
func TestBaseAgent_Execute_Simple(t *testing.T) {
	// 使用辅助函数创建标准Mock响应
	mockResponse := createStandardMockResponse("This is a test response")
	mockLLM := NewMockLLM(mockResponse, false)

	// 使用辅助函数创建测试Agent
	agent, err := createTestAgent(mockLLM)
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	task := NewBaseTask("Test task", "Expected output")
	ctx := context.Background()

	output, err := agent.Execute(ctx, task)
	if err != nil {
		t.Errorf("execution failed: %v", err)
		return
	}

	if output == nil {
		t.Errorf("expected output, got nil")
		return
	}

	if output.Raw != mockResponse.Content {
		t.Errorf("expected content %s, got %s", mockResponse.Content, output.Raw)
	}

	if output.Agent != agent.GetRole() {
		t.Errorf("expected agent %s, got %s", agent.GetRole(), output.Agent)
	}

	if output.TokensUsed != mockResponse.Usage.TotalTokens {
		t.Errorf("expected tokens %d, got %d", mockResponse.Usage.TotalTokens, output.TokensUsed)
	}

	// 验证LLM被调用
	if mockLLM.GetCallCount() != 1 {
		t.Errorf("expected 1 LLM call, got %d", mockLLM.GetCallCount())
	}
}

// 测试执行失败
func TestBaseAgent_Execute_LLMError(t *testing.T) {
	mockLLM := NewMockLLM(nil, true) // 设置为失败
	testLogger := logger.NewTestLogger()

	config := AgentConfig{
		Role:      "Test Agent",
		Goal:      "Test goal",
		Backstory: "Test backstory",
		LLM:       mockLLM,
		Logger:    testLogger,
	}

	agent, err := NewBaseAgent(config)
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	task := NewBaseTask("Test task", "Expected output")
	ctx := context.Background()

	output, err := agent.Execute(ctx, task)
	if err == nil {
		t.Errorf("expected execution to fail")
	}

	if output != nil {
		t.Errorf("expected no output on failure, got %v", output)
	}
}

// 测试异步执行
func TestBaseAgent_ExecuteAsync(t *testing.T) {
	mockResponse := &llm.Response{
		Content: "Async response",
		Usage: llm.Usage{
			TotalTokens: 15,
		},
	}

	mockLLM := NewMockLLM(mockResponse, false)
	testLogger := logger.NewTestLogger()

	config := AgentConfig{
		Role:      "Async Agent",
		Goal:      "Test async",
		Backstory: "Test backstory",
		LLM:       mockLLM,
		Logger:    testLogger,
	}

	agent, err := NewBaseAgent(config)
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	task := NewBaseTask("Async task", "Expected async output")
	ctx := context.Background()

	resultChan, err := agent.ExecuteAsync(ctx, task)
	if err != nil {
		t.Errorf("async execution setup failed: %v", err)
		return
	}

	// 等待结果
	select {
	case result := <-resultChan:
		if result.Error != nil {
			t.Errorf("async execution failed: %v", result.Error)
			return
		}

		if result.Output == nil {
			t.Errorf("expected output, got nil")
			return
		}

		if result.Output.Raw != mockResponse.Content {
			t.Errorf("expected content %s, got %s", mockResponse.Content, result.Output.Raw)
		}

	case <-time.After(5 * time.Second):
		t.Errorf("async execution timeout")
	}
}

// 测试超时执行
func TestBaseAgent_ExecuteWithTimeout(t *testing.T) {
	mockResponse := &llm.Response{
		Content: "Timeout test response",
		Usage: llm.Usage{
			TotalTokens: 12,
		},
	}

	mockLLM := NewMockLLM(mockResponse, false)
	testLogger := logger.NewTestLogger()

	config := AgentConfig{
		Role:      "Timeout Agent",
		Goal:      "Test timeout",
		Backstory: "Test backstory",
		LLM:       mockLLM,
		Logger:    testLogger,
	}

	agent, err := NewBaseAgent(config)
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	task := NewBaseTask("Timeout task", "Expected output")
	ctx := context.Background()

	// 测试正常超时（足够时间）
	output, err := agent.ExecuteWithTimeout(ctx, task, 10*time.Second)
	if err != nil {
		t.Errorf("timeout execution failed: %v", err)
		return
	}

	if output == nil {
		t.Errorf("expected output, got nil")
		return
	}

	if output.Raw != mockResponse.Content {
		t.Errorf("expected content %s, got %s", mockResponse.Content, output.Raw)
	}
}

// 测试工具管理
func TestBaseAgent_ToolManagement(t *testing.T) {
	mockLLM := NewMockLLM(&llm.Response{Content: "Tool test"}, false)

	config := AgentConfig{
		Role:      "Tool Agent",
		Goal:      "Test tools",
		Backstory: "Test backstory",
		LLM:       mockLLM,
	}

	agent, err := NewBaseAgent(config)
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	// 初始工具列表应该为空
	tools := agent.GetTools()
	if len(tools) != 0 {
		t.Errorf("expected 0 tools initially, got %d", len(tools))
	}

	// 添加工具
	calculator := NewCalculatorTool()
	err = agent.AddTool(calculator)
	if err != nil {
		t.Errorf("failed to add tool: %v", err)
	}

	// 验证工具已添加
	tools = agent.GetTools()
	if len(tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(tools))
	}

	if tools[0].GetName() != calculator.GetName() {
		t.Errorf("expected tool %s, got %s", calculator.GetName(), tools[0].GetName())
	}
}

// 测试执行统计
func TestBaseAgent_ExecutionStats(t *testing.T) {
	mockResponse := &llm.Response{
		Content: "Stats test",
		Usage: llm.Usage{
			TotalTokens: 20,
			Cost:        0.02,
		},
	}

	mockLLM := NewMockLLM(mockResponse, false)
	testLogger := logger.NewTestLogger()

	config := AgentConfig{
		Role:      "Stats Agent",
		Goal:      "Test stats",
		Backstory: "Test backstory",
		LLM:       mockLLM,
		Logger:    testLogger,
	}

	agent, err := NewBaseAgent(config)
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	// 初始统计
	stats := agent.GetExecutionStats()
	if stats.TotalExecutions != 0 {
		t.Errorf("expected 0 executions initially, got %d", stats.TotalExecutions)
	}

	// 执行任务
	task := NewBaseTask("Stats task", "Expected output")
	ctx := context.Background()

	_, err = agent.Execute(ctx, task)
	if err != nil {
		t.Errorf("execution failed: %v", err)
		return
	}

	// 验证统计更新
	stats = agent.GetExecutionStats()
	if stats.TotalExecutions != 1 {
		t.Errorf("expected 1 execution, got %d", stats.TotalExecutions)
	}

	if stats.SuccessfulExecutions != 1 {
		t.Errorf("expected 1 successful execution, got %d", stats.SuccessfulExecutions)
	}

	if stats.TokensUsed != mockResponse.Usage.TotalTokens {
		t.Errorf("expected tokens %d, got %d", mockResponse.Usage.TotalTokens, stats.TokensUsed)
	}

	if stats.TotalCost != mockResponse.Usage.Cost {
		t.Errorf("expected cost %f, got %f", mockResponse.Usage.Cost, stats.TotalCost)
	}
}

// 测试人工输入
func TestBaseAgent_HumanInput(t *testing.T) {
	mockResponse := &llm.Response{
		Content: "Human input response",
		Usage: llm.Usage{
			TotalTokens: 15,
		},
	}

	mockLLM := NewMockLLM(mockResponse, false)
	testLogger := logger.NewTestLogger()

	// 创建模拟输入处理器
	mockInputHandler := NewMockInputHandler([]string{"test input"}, testLogger)

	config := AgentConfig{
		Role:              "Human Input Agent",
		Goal:              "Test human input",
		Backstory:         "Test backstory",
		LLM:               mockLLM,
		Logger:            testLogger,
		HumanInputHandler: mockInputHandler,
	}

	agent, err := NewBaseAgent(config)
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	// 创建需要人工输入的任务
	task := NewTaskWithOptions(
		"Human input task",
		"Expected output",
		WithHumanInput(true),
	)

	ctx := context.Background()

	output, err := agent.Execute(ctx, task)
	if err != nil {
		t.Errorf("execution with human input failed: %v", err)
		return
	}

	if output == nil {
		t.Errorf("expected output, got nil")
		return
	}

	// 验证人工输入被设置
	if task.GetHumanInput() != "test input" {
		t.Errorf("expected human input 'test input', got '%s'", task.GetHumanInput())
	}
}

// 测试Agent克隆
func TestBaseAgent_Clone(t *testing.T) {
	mockLLM := NewMockLLM(&llm.Response{Content: "Clone test"}, false)
	calculator := NewCalculatorTool()

	config := AgentConfig{
		Role:      "Original Agent",
		Goal:      "Original goal",
		Backstory: "Original backstory",
		LLM:       mockLLM,
		Tools:     []Tool{calculator},
	}

	original, err := NewBaseAgent(config)
	if err != nil {
		t.Fatalf("failed to create original agent: %v", err)
	}

	// 克隆Agent
	cloned := original.Clone()

	// 验证基础属性相同
	if cloned.GetRole() != original.GetRole() {
		t.Errorf("cloned role mismatch: expected %s, got %s", original.GetRole(), cloned.GetRole())
	}

	if cloned.GetGoal() != original.GetGoal() {
		t.Errorf("cloned goal mismatch: expected %s, got %s", original.GetGoal(), cloned.GetGoal())
	}

	if cloned.GetBackstory() != original.GetBackstory() {
		t.Errorf("cloned backstory mismatch: expected %s, got %s", original.GetBackstory(), cloned.GetBackstory())
	}

	// 验证ID不同（应该是新的）
	if cloned.GetID() == original.GetID() {
		t.Errorf("cloned agent should have different ID")
	}

	// 验证工具被复制
	if len(cloned.GetTools()) != len(original.GetTools()) {
		t.Errorf("cloned tools count mismatch: expected %d, got %d",
			len(original.GetTools()), len(cloned.GetTools()))
	}
}

// 测试Agent关闭
func TestBaseAgent_Close(t *testing.T) {
	mockLLM := NewMockLLM(&llm.Response{Content: "Close test"}, false)

	config := AgentConfig{
		Role:      "Close Agent",
		Goal:      "Test close",
		Backstory: "Test backstory",
		LLM:       mockLLM,
	}

	agent, err := NewBaseAgent(config)
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	// 初始化Agent
	err = agent.Initialize()
	if err != nil {
		t.Fatalf("failed to initialize agent: %v", err)
	}

	// 关闭Agent
	err = agent.Close()
	if err != nil {
		t.Errorf("close failed: %v", err)
	}
}

// 测试统计重置
func TestBaseAgent_ResetStats(t *testing.T) {
	mockResponse := &llm.Response{
		Content: "Reset stats test",
		Usage: llm.Usage{
			TotalTokens: 25,
			Cost:        0.03,
		},
	}

	mockLLM := NewMockLLM(mockResponse, false)

	config := AgentConfig{
		Role:      "Reset Stats Agent",
		Goal:      "Test reset",
		Backstory: "Test backstory",
		LLM:       mockLLM,
	}

	agent, err := NewBaseAgent(config)
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	// 执行任务以生成统计
	task := NewBaseTask("Reset task", "Expected output")
	ctx := context.Background()

	_, err = agent.Execute(ctx, task)
	if err != nil {
		t.Errorf("execution failed: %v", err)
		return
	}

	// 验证有统计数据
	stats := agent.GetExecutionStats()
	if stats.TotalExecutions == 0 {
		t.Errorf("expected some executions before reset")
	}

	// 重置统计
	err = agent.ResetStats()
	if err != nil {
		t.Errorf("reset stats failed: %v", err)
	}

	// 验证统计已重置
	stats = agent.GetExecutionStats()
	if stats.TotalExecutions != 0 {
		t.Errorf("expected 0 executions after reset, got %d", stats.TotalExecutions)
	}

	if stats.TokensUsed != 0 {
		t.Errorf("expected 0 tokens after reset, got %d", stats.TokensUsed)
	}

	if stats.TotalCost != 0 {
		t.Errorf("expected 0 cost after reset, got %f", stats.TotalCost)
	}
}

// Benchmark测试
func BenchmarkBaseAgent_Execute(b *testing.B) {
	mockResponse := &llm.Response{
		Content: "Benchmark response",
		Usage: llm.Usage{
			TotalTokens: 10,
		},
	}

	mockLLM := NewMockLLM(mockResponse, false)
	testLogger := logger.NewTestLogger()

	config := AgentConfig{
		Role:      "Benchmark Agent",
		Goal:      "Benchmark goal",
		Backstory: "Benchmark backstory",
		LLM:       mockLLM,
		Logger:    testLogger,
	}

	agent, err := NewBaseAgent(config)
	if err != nil {
		b.Fatalf("failed to create agent: %v", err)
	}

	task := NewBaseTask("Benchmark task", "Expected output")
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := agent.Execute(ctx, task)
		if err != nil {
			b.Fatalf("execution failed: %v", err)
		}
	}
}

// 新增：测试推理功能相关的方法
func TestBaseAgent_ReasoningFeatures(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "set_and_get_reasoning_handler",
			test: func(t *testing.T) {
				mockResponse := &llm.Response{
					Content: "Test response",
					Usage: llm.Usage{
						TotalTokens: 10,
					},
				}
				mockLLM := NewMockLLM(mockResponse, false)
				testLogger := logger.NewTestLogger()

				config := AgentConfig{
					Role:      "Test Agent",
					Goal:      "Test goal",
					Backstory: "Test backstory",
					LLM:       mockLLM,
					Logger:    testLogger,
				}

				agent, err := NewBaseAgent(config)
				if err != nil {
					t.Fatalf("failed to create agent: %v", err)
				}

				// 创建mock推理处理器
				mockHandler := &MockReasoningHandler{}

				// 设置推理处理器
				agent.SetReasoningHandler(mockHandler)

				// 验证可以获取推理处理器
				handler := agent.GetReasoningHandler()
				if handler != mockHandler {
					t.Error("Expected reasoning handler to be set correctly")
				}
			},
		},
		{
			name: "set_and_get_step_callback",
			test: func(t *testing.T) {
				mockResponse := &llm.Response{
					Content: "Test response",
					Usage: llm.Usage{
						TotalTokens: 10,
					},
				}
				mockLLM := NewMockLLM(mockResponse, false)
				testLogger := logger.NewTestLogger()

				config := AgentConfig{
					Role:      "Test Agent",
					Goal:      "Test goal",
					Backstory: "Test backstory",
					LLM:       mockLLM,
					Logger:    testLogger,
				}

				agent, err := NewBaseAgent(config)
				if err != nil {
					t.Fatalf("failed to create agent: %v", err)
				}

				// 创建步骤回调函数
				callbackCalled := false
				stepCallback := func(ctx context.Context, step *AgentStep) error {
					callbackCalled = true
					return nil
				}

				// 设置步骤回调
				agent.SetStepCallback(stepCallback)

				// 验证可以获取步骤回调
				callback := agent.GetStepCallback()
				if callback == nil {
					t.Error("Expected step callback to be set")
					return
				}

				// 测试回调函数工作正常
				testStep := &AgentStep{
					StepID:      "test_step",
					StepType:    "test",
					Description: "Test step",
				}
				err = callback(context.Background(), testStep)
				if err != nil {
					t.Errorf("Callback failed: %v", err)
				}
				if !callbackCalled {
					t.Error("Expected callback to be called")
				}
			},
		},
		{
			name: "reasoning_handler_nil_check",
			test: func(t *testing.T) {
				mockResponse := &llm.Response{
					Content: "Test response",
					Usage: llm.Usage{
						TotalTokens: 10,
					},
				}
				mockLLM := NewMockLLM(mockResponse, false)
				testLogger := logger.NewTestLogger()

				config := AgentConfig{
					Role:      "Test Agent",
					Goal:      "Test goal",
					Backstory: "Test backstory",
					LLM:       mockLLM,
					Logger:    testLogger,
				}

				agent, err := NewBaseAgent(config)
				if err != nil {
					t.Fatalf("failed to create agent: %v", err)
				}

				// 默认推理处理器应该为nil
				handler := agent.GetReasoningHandler()
				if handler != nil {
					t.Error("Expected reasoning handler to be nil by default")
				}
			},
		},
		{
			name: "step_callback_nil_check",
			test: func(t *testing.T) {
				mockResponse := &llm.Response{
					Content: "Test response",
					Usage: llm.Usage{
						TotalTokens: 10,
					},
				}
				mockLLM := NewMockLLM(mockResponse, false)
				testLogger := logger.NewTestLogger()

				config := AgentConfig{
					Role:      "Test Agent",
					Goal:      "Test goal",
					Backstory: "Test backstory",
					LLM:       mockLLM,
					Logger:    testLogger,
				}

				agent, err := NewBaseAgent(config)
				if err != nil {
					t.Fatalf("failed to create agent: %v", err)
				}

				// 默认步骤回调应该为nil
				callback := agent.GetStepCallback()
				if callback != nil {
					t.Error("Expected step callback to be nil by default")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// MockReasoningHandler现在在 testing_mocks.go 中定义
