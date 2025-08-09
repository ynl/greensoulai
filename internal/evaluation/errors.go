package evaluation

import (
	"fmt"
)

// EvaluationExecutionError 评估执行过程中发生的错误
type EvaluationExecutionError struct {
	Phase      string // 错误发生的阶段
	TargetID   string // 被评估目标的ID
	TargetType string // 被评估目标的类型 (crew, task, agent)
	Iteration  int    // 当前迭代次数
	Retry      int    // 当前重试次数
	Err        error  // 原始错误
}

func NewEvaluationExecutionError(phase, targetID, targetType string, iteration, retry int, err error) *EvaluationExecutionError {
	return &EvaluationExecutionError{
		Phase:      phase,
		TargetID:   targetID,
		TargetType: targetType,
		Iteration:  iteration,
		Retry:      retry,
		Err:        err,
	}
}

func (e *EvaluationExecutionError) Error() string {
	return fmt.Sprintf("evaluation execution failed for %s '%s' at phase '%s' (iteration %d, retry %d): %v",
		e.TargetType, e.TargetID, e.Phase, e.Iteration, e.Retry, e.Err)
}

func (e *EvaluationExecutionError) Unwrap() error {
	return e.Err
}

// EvaluationConfigError 评估配置错误
type EvaluationConfigError struct {
	Field   string // 配置字段
	Value   string // 配置值
	Message string // 错误消息
}

func NewEvaluationConfigError(field, value, message string) *EvaluationConfigError {
	return &EvaluationConfigError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

func (e *EvaluationConfigError) Error() string {
	return fmt.Sprintf("evaluation config error for field '%s' with value '%s': %s", e.Field, e.Value, e.Message)
}

// EvaluatorCreationError 评估器创建失败的错误
type EvaluatorCreationError struct {
	EvaluatorType string // 评估器类型
	Category      string // 评估类别
	Message       string // 错误消息
	Err           error  // 原始错误
}

func NewEvaluatorCreationError(evaluatorType, category, message string, err error) *EvaluatorCreationError {
	return &EvaluatorCreationError{
		EvaluatorType: evaluatorType,
		Category:      category,
		Message:       message,
		Err:           err,
	}
}

func (e *EvaluatorCreationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("failed to create %s evaluator for category '%s': %s: %v",
			e.EvaluatorType, e.Category, e.Message, e.Err)
	}
	return fmt.Sprintf("failed to create %s evaluator for category '%s': %s",
		e.EvaluatorType, e.Category, e.Message)
}

func (e *EvaluatorCreationError) Unwrap() error {
	return e.Err
}

// TaskOutputError 任务输出相关错误
type TaskOutputError struct {
	TaskID  string // 任务ID
	Phase   string // 错误阶段 (parsing, validation, extraction)
	Message string // 错误消息
	Err     error  // 原始错误
}

func NewTaskOutputError(taskID, phase, message string, err error) *TaskOutputError {
	return &TaskOutputError{
		TaskID:  taskID,
		Phase:   phase,
		Message: message,
		Err:     err,
	}
}

func (e *TaskOutputError) Error() string {
	if e.TaskID != "" {
		return fmt.Sprintf("task output error for task '%s' at phase '%s': %s: %v",
			e.TaskID, e.Phase, e.Message, e.Err)
	}
	return fmt.Sprintf("task output error at phase '%s': %s: %v", e.Phase, e.Message, e.Err)
}

func (e *TaskOutputError) Unwrap() error {
	return e.Err
}

// ScoreValidationError 评分验证错误
type ScoreValidationError struct {
	Score    float64 // 评分值
	MinScore float64 // 最小值
	MaxScore float64 // 最大值
	Context  string  // 上下文信息
}

func NewScoreValidationError(score, minScore, maxScore float64, context string) *ScoreValidationError {
	return &ScoreValidationError{
		Score:    score,
		MinScore: minScore,
		MaxScore: maxScore,
		Context:  context,
	}
}

func (e *ScoreValidationError) Error() string {
	return fmt.Sprintf("score validation error in %s: score %.2f is not within valid range [%.2f, %.2f]",
		e.Context, e.Score, e.MinScore, e.MaxScore)
}

// LLMResponseError LLM响应错误
type LLMResponseError struct {
	Model    string // LLM模型名称
	Request  string // 请求内容
	Response string // 响应内容
	Phase    string // 错误阶段 (call, parse, validate)
	Err      error  // 原始错误
}

func NewLLMResponseError(model, request, response, phase string, err error) *LLMResponseError {
	return &LLMResponseError{
		Model:    model,
		Request:  request,
		Response: response,
		Phase:    phase,
		Err:      err,
	}
}

func (e *LLMResponseError) Error() string {
	return fmt.Sprintf("LLM response error from model '%s' at phase '%s': %v", e.Model, e.Phase, e.Err)
}

func (e *LLMResponseError) Unwrap() error {
	return e.Err
}

// EvaluationSessionError 评估会话错误
type EvaluationSessionError struct {
	SessionID string // 会话ID
	Operation string // 操作类型 (start, end, save, load)
	Message   string // 错误消息
	Err       error  // 原始错误
}

func NewEvaluationSessionError(sessionID, operation, message string, err error) *EvaluationSessionError {
	return &EvaluationSessionError{
		SessionID: sessionID,
		Operation: operation,
		Message:   message,
		Err:       err,
	}
}

func (e *EvaluationSessionError) Error() string {
	return fmt.Sprintf("evaluation session error for session '%s' during '%s': %s: %v",
		e.SessionID, e.Operation, e.Message, e.Err)
}

func (e *EvaluationSessionError) Unwrap() error {
	return e.Err
}

// MetricCategoryError 评估指标类别错误
type MetricCategoryError struct {
	Category MetricCategory // 指标类别
	Message  string         // 错误消息
}

func NewMetricCategoryError(category MetricCategory, message string) *MetricCategoryError {
	return &MetricCategoryError{
		Category: category,
		Message:  message,
	}
}

func (e *MetricCategoryError) Error() string {
	return fmt.Sprintf("metric category error for '%s': %s", e.Category, e.Message)
}

// AgentEvaluationError Agent评估错误
type AgentEvaluationError struct {
	AgentID   string // Agent ID
	AgentRole string // Agent角色
	TaskID    string // 任务ID（可选）
	Phase     string // 错误阶段
	Message   string // 错误消息
	Err       error  // 原始错误
}

func NewAgentEvaluationError(agentID, agentRole, taskID, phase, message string, err error) *AgentEvaluationError {
	return &AgentEvaluationError{
		AgentID:   agentID,
		AgentRole: agentRole,
		TaskID:    taskID,
		Phase:     phase,
		Message:   message,
		Err:       err,
	}
}

func (e *AgentEvaluationError) Error() string {
	if e.TaskID != "" {
		return fmt.Sprintf("agent evaluation error for agent '%s' (role: %s) on task '%s' at phase '%s': %s: %v",
			e.AgentID, e.AgentRole, e.TaskID, e.Phase, e.Message, e.Err)
	}
	return fmt.Sprintf("agent evaluation error for agent '%s' (role: %s) at phase '%s': %s: %v",
		e.AgentID, e.AgentRole, e.Phase, e.Message, e.Err)
}

func (e *AgentEvaluationError) Unwrap() error {
	return e.Err
}

// 预定义错误变量
var (
	// 配置相关错误
	ErrInvalidEvaluationConfig = fmt.Errorf("invalid evaluation configuration")
	ErrMissingEvaluatorLLM     = fmt.Errorf("evaluator LLM is required but not provided")
	ErrInvalidPassingScore     = fmt.Errorf("passing score must be between 0 and 10")
	ErrInvalidMetricCategory   = fmt.Errorf("invalid metric category")

	// 评估器相关错误
	ErrEvaluatorNotFound      = fmt.Errorf("evaluator not found")
	ErrEvaluatorAlreadyExists = fmt.Errorf("evaluator already exists")
	ErrInvalidEvaluatorType   = fmt.Errorf("invalid evaluator type")

	// 任务相关错误
	ErrTaskNotFound         = fmt.Errorf("task not found")
	ErrTaskOutputEmpty      = fmt.Errorf("task output is empty")
	ErrTaskOutputInvalid    = fmt.Errorf("task output is invalid")
	ErrTaskEvaluationFailed = fmt.Errorf("task evaluation failed")

	// Agent相关错误
	ErrAgentNotFound         = fmt.Errorf("agent not found")
	ErrAgentEvaluationFailed = fmt.Errorf("agent evaluation failed")
	ErrInvalidExecutionTrace = fmt.Errorf("invalid execution trace")

	// Crew相关错误
	ErrCrewNotFound         = fmt.Errorf("crew not found")
	ErrCrewEvaluationFailed = fmt.Errorf("crew evaluation failed")
	ErrNoTasksToEvaluate    = fmt.Errorf("no tasks to evaluate")

	// 会话相关错误
	ErrSessionNotStarted     = fmt.Errorf("evaluation session not started")
	ErrSessionAlreadyStarted = fmt.Errorf("evaluation session already started")
	ErrSessionAlreadyEnded   = fmt.Errorf("evaluation session already ended")
	ErrInvalidSessionID      = fmt.Errorf("invalid session ID")

	// LLM相关错误
	ErrLLMNotAvailable    = fmt.Errorf("LLM not available")
	ErrLLMResponseEmpty   = fmt.Errorf("LLM response is empty")
	ErrLLMResponseInvalid = fmt.Errorf("LLM response is invalid")
	ErrLLMCallTimeout     = fmt.Errorf("LLM call timeout")

	// 数据相关错误
	ErrInvalidScore              = fmt.Errorf("invalid score value")
	ErrMissingRequiredField      = fmt.Errorf("missing required field")
	ErrInvalidDataFormat         = fmt.Errorf("invalid data format")
	ErrDataSerializationFailed   = fmt.Errorf("data serialization failed")
	ErrDataDeserializationFailed = fmt.Errorf("data deserialization failed")
)

// IsEvaluationError 检查是否为评估相关的错误
func IsEvaluationError(err error) bool {
	if err == nil {
		return false
	}

	switch err.(type) {
	case *EvaluationExecutionError, *EvaluationConfigError, *EvaluatorCreationError,
		*TaskOutputError, *ScoreValidationError, *LLMResponseError,
		*EvaluationSessionError, *MetricCategoryError, *AgentEvaluationError:
		return true
	default:
		return false
	}
}

// GetErrorType 获取错误类型字符串
func GetErrorType(err error) string {
	if err == nil {
		return "no_error"
	}

	switch err.(type) {
	case *EvaluationExecutionError:
		return "evaluation_execution_error"
	case *EvaluationConfigError:
		return "evaluation_config_error"
	case *EvaluatorCreationError:
		return "evaluator_creation_error"
	case *TaskOutputError:
		return "task_output_error"
	case *ScoreValidationError:
		return "score_validation_error"
	case *LLMResponseError:
		return "llm_response_error"
	case *EvaluationSessionError:
		return "evaluation_session_error"
	case *MetricCategoryError:
		return "metric_category_error"
	case *AgentEvaluationError:
		return "agent_evaluation_error"
	default:
		return "unknown_error"
	}
}
