package planning

import (
	"context"
	"fmt"
	"time"

	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// PlanExecutorImpl PlanExecutor的实现
// 负责执行规划逻辑，包括重试机制和超时控制
type PlanExecutorImpl struct {
	agentFactory AgentFactory    // 代理工厂
	taskFactory  TaskFactory     // 任务工厂
	eventBus     events.EventBus // 事件总线
	logger       logger.Logger   // 日志器
	maxRetries   int             // 最大重试次数
	timeoutSec   int             // 超时时间（秒）
}

// NewPlanExecutor 创建计划执行器
func NewPlanExecutor(
	agentFactory AgentFactory,
	taskFactory TaskFactory,
	eventBus events.EventBus,
	logger logger.Logger,
) PlanExecutor {
	return &PlanExecutorImpl{
		agentFactory: agentFactory,
		taskFactory:  taskFactory,
		eventBus:     eventBus,
		logger:       logger,
		maxRetries:   3,   // 默认最大重试次数
		timeoutSec:   300, // 默认5分钟超时
	}
}

// ExecutePlanning 执行规划逻辑
func (pe *PlanExecutorImpl) ExecutePlanning(ctx context.Context, request *PlanningRequest) (*PlanningResult, error) {
	startTime := time.Now()

	// 验证请求
	if err := pe.validateRequest(request); err != nil {
		return nil, fmt.Errorf("invalid planning request: %w", err)
	}

	// 创建CrewPlanner实例
	config := pe.buildPlanningConfig(request)
	planner := NewCrewPlanner(
		request.Tasks,
		config,
		pe.eventBus,
		pe.logger,
		pe.agentFactory,
		pe.taskFactory,
	)

	// 执行规划
	output, err := planner.HandleCrewPlanning(ctx)
	if err != nil {
		executionTime := float64(time.Since(startTime).Nanoseconds()) / 1e6
		return &PlanningResult{
			Success:       false,
			ErrorMessage:  err.Error(),
			ExecutionTime: executionTime,
			RetryCount:    0,
		}, err
	}

	executionTime := float64(time.Since(startTime).Nanoseconds()) / 1e6

	// 构建成功结果
	result := &PlanningResult{
		Output:        *output,
		ExecutionTime: executionTime,
		Success:       true,
		RetryCount:    0,
		ModelUsed:     config.PlanningAgentLLM,
	}

	pe.logger.Info("Planning executed successfully",
		logger.Field{Key: "task_count", Value: len(request.Tasks)},
		logger.Field{Key: "execution_time_ms", Value: executionTime},
		logger.Field{Key: "plans_generated", Value: output.GetTaskCount()},
	)

	return result, nil
}

// ExecutePlanningWithRetry 带重试的规划执行
func (pe *PlanExecutorImpl) ExecutePlanningWithRetry(ctx context.Context, request *PlanningRequest) (*PlanningResult, error) {
	var lastErr error
	var lastResult *PlanningResult

	// 确定重试次数
	maxRetries := pe.maxRetries
	if request.MaxRetries > 0 {
		maxRetries = request.MaxRetries
	}

	for retry := 0; retry <= maxRetries; retry++ {
		// 创建带超时的上下文
		timeoutSec := pe.timeoutSec
		if request.TimeoutSec > 0 {
			timeoutSec = request.TimeoutSec
		}

		retryCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)

		// 发射重试事件（非首次执行时）
		if retry > 0 && pe.eventBus != nil {
			event := NewPlanningRetryEvent(retry, maxRetries, lastErr.Error(), "execution_failed")
			pe.eventBus.Emit(retryCtx, pe, event)

			// 添加重试延迟（指数退避）
			delay := time.Duration(retry*retry) * time.Second
			if delay > 30*time.Second {
				delay = 30 * time.Second // 最大30秒延迟
			}
			time.Sleep(delay)
		}

		pe.logger.Debug("Executing planning",
			logger.Field{Key: "retry", Value: retry},
			logger.Field{Key: "max_retries", Value: maxRetries},
			logger.Field{Key: "timeout_sec", Value: timeoutSec},
		)

		result, err := pe.ExecutePlanning(retryCtx, request)
		cancel()

		if err == nil {
			// 成功，更新重试计数并返回
			result.RetryCount = retry
			return result, nil
		}

		lastErr = err
		lastResult = result

		// 检查是否应该重试
		if !pe.shouldRetry(err, retry, maxRetries) {
			break
		}

		pe.logger.Warn("Planning attempt failed, will retry",
			logger.Field{Key: "retry", Value: retry},
			logger.Field{Key: "max_retries", Value: maxRetries},
			logger.Field{Key: "error", Value: err},
		)
	}

	// 所有重试都失败了
	if lastResult != nil {
		lastResult.RetryCount = maxRetries
		lastResult.ErrorMessage = fmt.Sprintf("failed after %d retries: %v", maxRetries, lastErr)
	} else {
		lastResult = &PlanningResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed after %d retries: %v", maxRetries, lastErr),
			RetryCount:   maxRetries,
		}
	}

	pe.logger.Error("Planning failed after all retries",
		logger.Field{Key: "max_retries", Value: maxRetries},
		logger.Field{Key: "final_error", Value: lastErr},
	)

	return lastResult, lastErr
}

// SetMaxRetries 设置最大重试次数
func (pe *PlanExecutorImpl) SetMaxRetries(maxRetries int) {
	if maxRetries < 0 {
		maxRetries = 0
	}
	if maxRetries > 10 {
		maxRetries = 10 // 限制最大重试次数
	}
	pe.maxRetries = maxRetries
}

// SetTimeout 设置超时时间
func (pe *PlanExecutorImpl) SetTimeout(timeoutSeconds int) {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 60 // 最小1分钟
	}
	if timeoutSeconds > 3600 {
		timeoutSeconds = 3600 // 最大1小时
	}
	pe.timeoutSec = timeoutSeconds
}

// validateRequest 验证规划请求
func (pe *PlanExecutorImpl) validateRequest(request *PlanningRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if len(request.Tasks) == 0 {
		return fmt.Errorf("tasks list cannot be empty")
	}

	if len(request.Tasks) > 100 {
		return fmt.Errorf("too many tasks (maximum 100 allowed)")
	}

	// 验证每个任务
	for i, task := range request.Tasks {
		if task.Description == "" {
			return fmt.Errorf("task %d description cannot be empty", i)
		}
		if task.ExpectedOutput == "" {
			return fmt.Errorf("task %d expected output cannot be empty", i)
		}
	}

	// 验证超时和重试参数
	if request.TimeoutSec < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}
	if request.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	return nil
}

// buildPlanningConfig 构建规划配置
func (pe *PlanExecutorImpl) buildPlanningConfig(request *PlanningRequest) *PlanningConfig {
	config := DefaultPlanningConfig()

	// 应用请求中的配置
	if request.PlanningLLM != "" {
		config.PlanningAgentLLM = request.PlanningLLM
	}

	if request.MaxRetries > 0 {
		config.MaxRetries = request.MaxRetries
	}

	if request.TimeoutSec > 0 {
		config.TimeoutSeconds = request.TimeoutSec
	}

	// 合并自定义提示词
	if request.CustomPrompts != nil {
		for key, value := range request.CustomPrompts {
			config.CustomPrompts[key] = value
		}
	}

	// 合并上下文到附加配置
	if request.Context != nil {
		config.AdditionalConfig["context"] = request.Context
	}

	return config
}

// shouldRetry 判断是否应该重试
func (pe *PlanExecutorImpl) shouldRetry(err error, currentRetry, maxRetries int) bool {
	if currentRetry >= maxRetries {
		return false
	}

	// 检查错误类型，某些错误不应该重试
	switch e := err.(type) {
	case *PlanValidationError:
		// 验证错误通常不应该重试
		return false
	case *AgentCreationError:
		// 代理创建错误可能是配置问题，可以重试
		return true
	case *PlanningExecutionError:
		// 根据执行阶段决定是否重试
		switch e.Phase {
		case "validation":
			return false // 验证失败不重试
		case "agent_creation", "tasks_summary", "planner_task", "execution":
			return true // 这些阶段可以重试
		default:
			return true
		}
	default:
		// 其他错误默认可以重试
		return true
	}
}

// GetExecutorStatistics 获取执行器统计信息
func (pe *PlanExecutorImpl) GetExecutorStatistics() map[string]interface{} {
	return map[string]interface{}{
		"max_retries":     pe.maxRetries,
		"timeout_sec":     pe.timeoutSec,
		"factory_ready":   pe.agentFactory != nil && pe.taskFactory != nil,
		"event_bus_ready": pe.eventBus != nil,
		"logger_ready":    pe.logger != nil,
	}
}

// ValidateFactories 验证工厂依赖
func (pe *PlanExecutorImpl) ValidateFactories() error {
	if pe.agentFactory == nil {
		return fmt.Errorf("agent factory is nil")
	}

	if pe.taskFactory == nil {
		return fmt.Errorf("task factory is nil")
	}

	return nil
}

// SetEventBus 设置事件总线
func (pe *PlanExecutorImpl) SetEventBus(eventBus events.EventBus) {
	pe.eventBus = eventBus
}

// SetLogger 设置日志器
func (pe *PlanExecutorImpl) SetLogger(logger logger.Logger) {
	pe.logger = logger
}
