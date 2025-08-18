package agent

import (
	"context"
	"fmt"

	"github.com/ynl/greensoulai/pkg/logger"
)

// 确保BaseAgent实现ReActAgent接口
var _ ReActAgent = (*BaseAgent)(nil)

// SetReActMode 设置是否启用ReAct模式
func (a *BaseAgent) SetReActMode(enabled bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if enabled {
		a.executionConfig.Mode = ModeReAct
		// 如果没有ReAct配置，使用默认配置
		if a.executionConfig.ReActConfig == nil {
			a.executionConfig.ReActConfig = DefaultReActConfig()
		}
	} else {
		a.executionConfig.Mode = ModeJSON
	}
}

// GetReActMode 获取当前是否启用ReAct模式
func (a *BaseAgent) GetReActMode() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.executionConfig.Mode == ModeReAct
}

// SetReActConfig 设置ReAct配置
func (a *BaseAgent) SetReActConfig(config *ReActConfig) error {
	if config == nil {
		return fmt.Errorf("ReAct config cannot be nil")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.executionConfig.ReActConfig = config
	return nil
}

// GetReActConfig 获取ReAct配置
func (a *BaseAgent) GetReActConfig() *ReActConfig {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.executionConfig.ReActConfig == nil {
		return DefaultReActConfig()
	}
	return a.executionConfig.ReActConfig
}

// GetReActTrace 获取最近的ReAct推理轨迹
func (a *BaseAgent) GetReActTrace() *ReActTrace {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.lastReActTrace
}

// ExecuteWithReAct 使用ReAct模式执行任务
func (a *BaseAgent) ExecuteWithReAct(ctx context.Context, task Task) (*TaskOutput, *ReActTrace, error) {
	if !a.isInitialized {
		if err := a.Initialize(); err != nil {
			return nil, nil, fmt.Errorf("failed to initialize agent: %w", err)
		}
	}

	// 验证必要组件
	if a.reactExecutor == nil {
		return nil, nil, fmt.Errorf("ReAct executor not available")
	}

	// 记录执行开始
	a.mu.Lock()
	a.timesExecuted++
	a.mu.Unlock()

	// 发送开始执行事件
	if a.eventBus != nil {
		startEvent := NewAgentExecutionStartedEvent(a.id, a.role, task.GetID(), task.GetDescription(), a.timesExecuted)
		if err := a.eventBus.Emit(ctx, a, startEvent); err != nil {
			a.logger.Warn("Failed to publish start event", logger.Field{Key: "error", Value: err})
		}
	}

	// 使用ReAct执行器执行任务
	trace, err := a.reactExecutor.ExecuteReAct(ctx, a, task)
	if err != nil {
		// 记录失败
		a.mu.Lock()
		a.stats.FailedExecutions++
		a.mu.Unlock()

		if a.eventBus != nil {
			errorEvent := NewAgentExecutionFailedEvent(a.id, a.role, task.GetID(), task.GetDescription(), a.timesExecuted, 0, err)
			if pubErr := a.eventBus.Emit(ctx, a, errorEvent); pubErr != nil {
				a.logger.Warn("Failed to publish error event", logger.Field{Key: "error", Value: pubErr})
			}
		}

		return nil, trace, fmt.Errorf("ReAct execution failed: %w", err)
	}

	// 保存轨迹
	a.mu.Lock()
	a.lastReActTrace = trace
	a.stats.SuccessfulExecutions++
	a.mu.Unlock()

	// 构建任务输出
	output := &TaskOutput{
		Raw:            trace.FinalOutput,
		Agent:          a.role,
		Task:           task.GetID(),
		Description:    task.GetDescription(),
		Summary:        a.generateSummaryFromTrace(trace),
		ExpectedOutput: task.GetExpectedOutput(),
		OutputFormat:   OutputFormatRAW,
		ExecutionTime:  trace.TotalDuration,
		CreatedAt:      trace.EndTime,
		TokensUsed:     0, // TODO: 从LLM响应中获取
		Cost:           0, // TODO: 计算成本
		Model:          a.getLLMModelName(),
		IsValid:        trace.IsCompleted && len(trace.FinalOutput) > 0,
		ToolsUsed:      a.extractToolsFromTrace(trace),
		Metadata: map[string]interface{}{
			"mode":            "react",
			"trace_id":        trace.TraceID,
			"iteration_count": trace.IterationCount,
			"total_duration":  trace.TotalDuration,
			"steps_count":     len(trace.Steps),
		},
	}

	// 验证输出（如果任务有guardrail）
	if task.HasGuardrail() {
		if guardrail := task.GetGuardrail(); guardrail != nil {
			if result, err := guardrail.Validate(ctx, output); err != nil {
				output.ValidationError = err.Error()
				output.IsValid = false
			} else if !result.Valid {
				output.ValidationError = result.Error
				output.IsValid = false
			}
		}
	}

	// 执行回调
	if err := a.executeCallbacks(ctx, output); err != nil {
		a.logger.Warn("Callback execution failed", logger.Field{Key: "error", Value: err})
	}

	// 发送完成事件
	if a.eventBus != nil {
		completedEvent := NewAgentExecutionCompletedEvent(a.id, a.role, task.GetID(), task.GetDescription(), a.timesExecuted, trace.TotalDuration, true, output)
		if err := a.eventBus.Emit(ctx, a, completedEvent); err != nil {
			a.logger.Warn("Failed to publish completion event", logger.Field{Key: "error", Value: err})
		}
	}

	return output, trace, nil
}

// 扩展原有的Execute方法以支持ReAct模式
func (a *BaseAgent) executeWithMode(ctx context.Context, task Task) (*TaskOutput, error) {
	// 检查执行模式
	if a.executionConfig.Mode == ModeReAct {
		output, _, err := a.ExecuteWithReAct(ctx, task)
		return output, err
	}

	// 原有的JSON模式执行逻辑保持不变
	return a.executeOriginal(ctx, task)
}

// executeOriginal 保留原有的执行逻辑
func (a *BaseAgent) executeOriginal(ctx context.Context, task Task) (*TaskOutput, error) {
	// 原有的Execute方法逻辑
	// 这里应该是原来的Execute方法的实现
	// 为了不破坏现有系统，我们保持原有逻辑不变

	if !a.isInitialized {
		if err := a.Initialize(); err != nil {
			return nil, fmt.Errorf("failed to initialize agent: %w", err)
		}
	}

	// ... 原有的执行逻辑 ...
	// 这里应该调用原有的实现，但为了演示，我们提供一个基本实现

	// 1. 创建工具执行上下文
	toolCtx := NewToolExecutionContext(a, task)

	// 2. 构建任务提示
	prompt, err := a.buildTaskPromptWithTools(ctx, task, toolCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to build task prompt: %w", err)
	}

	// 3. 准备LLM消息
	messages := a.buildMessages(prompt)

	// 4. 调用LLM
	response, err := a.llmProvider.Call(ctx, messages, a.buildLLMCallOptionsWithTools(toolCtx))
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	// 5. 构建输出
	output := a.buildTaskOutput(task, response)

	// 6. 执行回调
	if err := a.executeCallbacks(ctx, output); err != nil {
		a.logger.Warn("Callback execution failed", logger.Field{Key: "error", Value: err})
	}

	return output, nil
}

// 辅助方法

// generateSummaryFromTrace 从ReAct轨迹生成摘要
func (a *BaseAgent) generateSummaryFromTrace(trace *ReActTrace) string {
	if trace == nil || len(trace.Steps) == 0 {
		return "No steps executed"
	}

	stepCount := len(trace.Steps)
	duration := trace.TotalDuration

	summary := fmt.Sprintf("Completed %d ReAct steps in %v", stepCount, duration)

	if trace.IsCompleted {
		summary += " - Task completed successfully"
	} else {
		summary += " - Task incomplete"
	}

	return summary
}

// extractToolsFromTrace 从ReAct轨迹中提取使用的工具
func (a *BaseAgent) extractToolsFromTrace(trace *ReActTrace) []string {
	if trace == nil {
		return []string{}
	}

	toolsUsed := make(map[string]bool)
	for _, step := range trace.Steps {
		if step.Action != "" {
			toolsUsed[step.Action] = true
		}
	}

	tools := make([]string, 0, len(toolsUsed))
	for tool := range toolsUsed {
		tools = append(tools, tool)
	}

	return tools
}

// getLLMModelName 获取LLM模型名称
func (a *BaseAgent) getLLMModelName() string {
	if a.llmProvider != nil {
		return a.llmProvider.GetModel()
	}
	return "unknown"
}

// SwitchMode 切换Agent执行模式（便利方法）
func (a *BaseAgent) SwitchMode(mode AgentMode) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	switch mode {
	case ModeJSON:
		a.executionConfig.Mode = ModeJSON
	case ModeReAct:
		a.executionConfig.Mode = ModeReAct
		if a.executionConfig.ReActConfig == nil {
			a.executionConfig.ReActConfig = DefaultReActConfig()
		}
	case ModeHybrid:
		a.executionConfig.Mode = ModeHybrid
		if a.executionConfig.ReActConfig == nil {
			a.executionConfig.ReActConfig = DefaultReActConfig()
		}
	default:
		return fmt.Errorf("unsupported mode: %v", mode)
	}

	return nil
}

// GetCurrentMode 获取当前执行模式
func (a *BaseAgent) GetCurrentMode() AgentMode {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.executionConfig.Mode
}

// IsReActEnabled 检查是否启用了ReAct模式
func (a *BaseAgent) IsReActEnabled() bool {
	return a.GetCurrentMode() == ModeReAct
}

// GetReActStats 获取ReAct执行统计
func (a *BaseAgent) GetReActStats() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	stats := make(map[string]interface{})

	if a.lastReActTrace != nil {
		stats["last_trace_id"] = a.lastReActTrace.TraceID
		stats["last_steps_count"] = len(a.lastReActTrace.Steps)
		stats["last_duration"] = a.lastReActTrace.TotalDuration
		stats["last_completed"] = a.lastReActTrace.IsCompleted
	}

	return stats
}
