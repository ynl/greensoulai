package planning

import (
	"errors"
	"fmt"
)

// 规划系统相关错误定义
var (
	ErrEmptyPlanList         = errors.New("plan list cannot be empty")
	ErrInvalidTaskInfo       = errors.New("invalid task information")
	ErrPlanningAgentCreation = errors.New("failed to create planning agent")
	ErrPlanningTaskCreation  = errors.New("failed to create planning task")
	ErrPlanningExecution     = errors.New("failed to execute planning")
	ErrInvalidPlanOutput     = errors.New("invalid planning output format")
	ErrPlanningTimeout       = errors.New("planning execution timeout")
	ErrPlanningRetryExceeded = errors.New("planning retry attempts exceeded")
	ErrInvalidLLMResponse    = errors.New("invalid LLM response format")
)

// PlanValidationError 计划验证错误
type PlanValidationError struct {
	Field   string // 错误字段
	Index   int    // 错误索引
	Message string // 错误消息
}

// Error 实现error接口
func (e *PlanValidationError) Error() string {
	return fmt.Sprintf("validation error at index %d, field '%s': %s", e.Index, e.Field, e.Message)
}

// PlanningExecutionError 规划执行错误
type PlanningExecutionError struct {
	Phase      string // 执行阶段
	Cause      error  // 原因
	RetryCount int    // 重试次数
	TaskCount  int    // 任务数量
	AgentID    string // 规划代理ID
	ModelUsed  string // 使用的模型
}

// Error 实现error接口
func (e *PlanningExecutionError) Error() string {
	return fmt.Sprintf("planning execution failed at phase '%s' (retry %d/%d): %v",
		e.Phase, e.RetryCount, 3, e.Cause)
}

// Unwrap 返回原始错误
func (e *PlanningExecutionError) Unwrap() error {
	return e.Cause
}

// TaskSummaryError 任务摘要生成错误
type TaskSummaryError struct {
	TaskIndex int    // 任务索引
	TaskID    string // 任务ID
	Reason    string // 错误原因
}

// Error 实现error接口
func (e *TaskSummaryError) Error() string {
	return fmt.Sprintf("failed to create summary for task %d ('%s'): %s", e.TaskIndex, e.TaskID, e.Reason)
}

// AgentCreationError 代理创建错误
type AgentCreationError struct {
	AgentRole string // 代理角色
	Reason    string // 错误原因
}

// Error 实现error接口
func (e *AgentCreationError) Error() string {
	return fmt.Sprintf("failed to create agent with role '%s': %s", e.AgentRole, e.Reason)
}

// NewPlanValidationError 创建计划验证错误
func NewPlanValidationError(field string, index int, message string) *PlanValidationError {
	return &PlanValidationError{
		Field:   field,
		Index:   index,
		Message: message,
	}
}

// NewPlanningExecutionError 创建规划执行错误
func NewPlanningExecutionError(phase string, cause error, retryCount int, taskCount int, agentID string, modelUsed string) *PlanningExecutionError {
	return &PlanningExecutionError{
		Phase:      phase,
		Cause:      cause,
		RetryCount: retryCount,
		TaskCount:  taskCount,
		AgentID:    agentID,
		ModelUsed:  modelUsed,
	}
}

// NewTaskSummaryError 创建任务摘要错误
func NewTaskSummaryError(taskIndex int, taskID string, reason string) *TaskSummaryError {
	return &TaskSummaryError{
		TaskIndex: taskIndex,
		TaskID:    taskID,
		Reason:    reason,
	}
}

// NewAgentCreationError 创建代理创建错误
func NewAgentCreationError(agentRole string, reason string) *AgentCreationError {
	return &AgentCreationError{
		AgentRole: agentRole,
		Reason:    reason,
	}
}

