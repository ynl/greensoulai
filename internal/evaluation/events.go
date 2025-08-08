package evaluation

import (
	"time"

	"github.com/ynl/greensoulai/pkg/events"
)

// 评估系统事件类型常量
const (
	EventTypeEvaluationStarted          = "evaluation.started"
	EventTypeEvaluationCompleted        = "evaluation.completed"
	EventTypeEvaluationFailed           = "evaluation.failed"
	EventTypeTaskEvaluationStarted      = "evaluation.task.started"
	EventTypeTaskEvaluationCompleted    = "evaluation.task.completed"
	EventTypeTaskEvaluationFailed       = "evaluation.task.failed"
	EventTypeAgentEvaluationStarted     = "evaluation.agent.started"
	EventTypeAgentEvaluationCompleted   = "evaluation.agent.completed"
	EventTypeAgentEvaluationFailed      = "evaluation.agent.failed"
	EventTypeCrewEvaluationStarted      = "evaluation.crew.started"
	EventTypeCrewEvaluationCompleted    = "evaluation.crew.completed"
	EventTypeCrewEvaluationFailed       = "evaluation.crew.failed"
	EventTypeEvaluationSessionStarted   = "evaluation.session.started"
	EventTypeEvaluationSessionCompleted = "evaluation.session.completed"
	EventTypeEvaluationSessionFailed    = "evaluation.session.failed"
	EventTypeCrewTestResult             = "evaluation.crew.test.result" // 对应Python版本的CrewTestResultEvent
)

// EvaluationStartedEvent 评估开始事件
type EvaluationStartedEvent struct {
	events.BaseEvent
	EvaluationType string            `json:"evaluation_type"`  // 评估类型 (crew, task, agent)
	TargetID       string            `json:"target_id"`        // 被评估目标的ID
	TargetName     string            `json:"target_name"`      // 被评估目标的名称
	IterationID    string            `json:"iteration_id"`     // 迭代ID
	Config         *EvaluationConfig `json:"config,omitempty"` // 评估配置
}

func NewEvaluationStartedEvent(source interface{}, evaluationType, targetID, targetName, iterationID string, config *EvaluationConfig) *EvaluationStartedEvent {
	return &EvaluationStartedEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypeEvaluationStarted,
			Timestamp: time.Now(),
			Source:    source,
			Payload: map[string]interface{}{
				"evaluation_type": evaluationType,
				"target_id":       targetID,
				"target_name":     targetName,
				"iteration_id":    iterationID,
			},
		},
		EvaluationType: evaluationType,
		TargetID:       targetID,
		TargetName:     targetName,
		IterationID:    iterationID,
		Config:         config,
	}
}

// EvaluationCompletedEvent 评估完成事件
type EvaluationCompletedEvent struct {
	events.BaseEvent
	EvaluationType  string  `json:"evaluation_type"`
	TargetID        string  `json:"target_id"`
	TargetName      string  `json:"target_name"`
	IterationID     string  `json:"iteration_id"`
	Score           float64 `json:"score"`
	Grade           string  `json:"grade"`
	ExecutionTimeMs float64 `json:"execution_time_ms"`
	Success         bool    `json:"success"`
}

func NewEvaluationCompletedEvent(source interface{}, evaluationType, targetID, targetName, iterationID string, score float64, grade string, executionTimeMs float64, success bool) *EvaluationCompletedEvent {
	return &EvaluationCompletedEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypeEvaluationCompleted,
			Timestamp: time.Now(),
			Source:    source,
			Payload: map[string]interface{}{
				"evaluation_type":   evaluationType,
				"target_id":         targetID,
				"target_name":       targetName,
				"iteration_id":      iterationID,
				"score":             score,
				"grade":             grade,
				"execution_time_ms": executionTimeMs,
				"success":           success,
			},
		},
		EvaluationType:  evaluationType,
		TargetID:        targetID,
		TargetName:      targetName,
		IterationID:     iterationID,
		Score:           score,
		Grade:           grade,
		ExecutionTimeMs: executionTimeMs,
		Success:         success,
	}
}

// EvaluationFailedEvent 评估失败事件
type EvaluationFailedEvent struct {
	events.BaseEvent
	EvaluationType  string  `json:"evaluation_type"`
	TargetID        string  `json:"target_id"`
	TargetName      string  `json:"target_name"`
	IterationID     string  `json:"iteration_id"`
	Error           string  `json:"error"`
	Phase           string  `json:"phase"` // 失败阶段
	ExecutionTimeMs float64 `json:"execution_time_ms"`
}

func NewEvaluationFailedEvent(source interface{}, evaluationType, targetID, targetName, iterationID, errorMsg, phase string, executionTimeMs float64) *EvaluationFailedEvent {
	return &EvaluationFailedEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypeEvaluationFailed,
			Timestamp: time.Now(),
			Source:    source,
			Payload: map[string]interface{}{
				"evaluation_type":   evaluationType,
				"target_id":         targetID,
				"target_name":       targetName,
				"iteration_id":      iterationID,
				"error":             errorMsg,
				"phase":             phase,
				"execution_time_ms": executionTimeMs,
			},
		},
		EvaluationType:  evaluationType,
		TargetID:        targetID,
		TargetName:      targetName,
		IterationID:     iterationID,
		Error:           errorMsg,
		Phase:           phase,
		ExecutionTimeMs: executionTimeMs,
	}
}

// TaskEvaluatedEvent 任务评估事件，对应Python版本的TaskEvaluationEvent
type TaskEvaluatedEvent struct {
	events.BaseEvent
	TaskID          string                        `json:"task_id"`
	TaskDescription string                        `json:"task_description"`
	AgentRole       string                        `json:"agent_role"`
	Score           float64                       `json:"score"`
	Evaluation      *TaskEvaluationPydanticOutput `json:"evaluation"`
	ExecutionTimeMs float64                       `json:"execution_time_ms"`
	IterationID     string                        `json:"iteration_id"`
}

func NewTaskEvaluatedEvent(source interface{}, taskID, taskDescription, agentRole string, score float64, evaluation *TaskEvaluationPydanticOutput, executionTimeMs float64, iterationID string) *TaskEvaluatedEvent {
	return &TaskEvaluatedEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypeTaskEvaluationCompleted,
			Timestamp: time.Now(),
			Source:    source,
			Payload: map[string]interface{}{
				"task_id":           taskID,
				"task_description":  taskDescription,
				"agent_role":        agentRole,
				"score":             score,
				"execution_time_ms": executionTimeMs,
				"iteration_id":      iterationID,
			},
		},
		TaskID:          taskID,
		TaskDescription: taskDescription,
		AgentRole:       agentRole,
		Score:           score,
		Evaluation:      evaluation,
		ExecutionTimeMs: executionTimeMs,
		IterationID:     iterationID,
	}
}

// AgentEvaluatedEvent Agent评估事件
type AgentEvaluatedEvent struct {
	events.BaseEvent
	AgentID         string                 `json:"agent_id"`
	AgentRole       string                 `json:"agent_role"`
	TaskID          string                 `json:"task_id,omitempty"`
	Result          *AgentEvaluationResult `json:"result"`
	AverageScore    float64                `json:"average_score"`
	MetricsCount    int                    `json:"metrics_count"`
	ExecutionTimeMs float64                `json:"execution_time_ms"`
	IterationID     string                 `json:"iteration_id"`
}

func NewAgentEvaluatedEvent(source interface{}, agentID, agentRole, taskID string, result *AgentEvaluationResult, executionTimeMs float64, iterationID string) *AgentEvaluatedEvent {
	return &AgentEvaluatedEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypeAgentEvaluationCompleted,
			Timestamp: time.Now(),
			Source:    source,
			Payload: map[string]interface{}{
				"agent_id":          agentID,
				"agent_role":        agentRole,
				"task_id":           taskID,
				"average_score":     result.GetAverageScore(),
				"metrics_count":     len(result.Metrics),
				"execution_time_ms": executionTimeMs,
				"iteration_id":      iterationID,
			},
		},
		AgentID:         agentID,
		AgentRole:       agentRole,
		TaskID:          taskID,
		Result:          result,
		AverageScore:    result.GetAverageScore(),
		MetricsCount:    len(result.Metrics),
		ExecutionTimeMs: executionTimeMs,
		IterationID:     iterationID,
	}
}

// CrewTestResultEvent Crew测试结果事件，对应Python版本的CrewTestResultEvent
type CrewTestResultEvent struct {
	events.BaseEvent
	Quality           float64 `json:"quality"`              // 质量评分
	ExecutionDuration float64 `json:"execution_duration"`   // 执行时间
	Model             string  `json:"model"`                // 使用的模型
	CrewName          string  `json:"crew_name"`            // Crew名称
	Iteration         int     `json:"iteration"`            // 迭代次数
	TaskID            string  `json:"task_id,omitempty"`    // 任务ID
	AgentRole         string  `json:"agent_role,omitempty"` // Agent角色
}

func NewCrewTestResultEvent(source interface{}, quality, executionDuration float64, model, crewName string, iteration int, taskID, agentRole string) *CrewTestResultEvent {
	return &CrewTestResultEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypeCrewTestResult,
			Timestamp: time.Now(),
			Source:    source,
			Payload: map[string]interface{}{
				"quality":            quality,
				"execution_duration": executionDuration,
				"model":              model,
				"crew_name":          crewName,
				"iteration":          iteration,
				"task_id":            taskID,
				"agent_role":         agentRole,
			},
		},
		Quality:           quality,
		ExecutionDuration: executionDuration,
		Model:             model,
		CrewName:          crewName,
		Iteration:         iteration,
		TaskID:            taskID,
		AgentRole:         agentRole,
	}
}

// TaskEvaluationStartedEvent 任务评估开始事件
type TaskEvaluationStartedEvent struct {
	events.BaseEvent
	TaskID          string `json:"task_id"`
	TaskDescription string `json:"task_description"`
	AgentRole       string `json:"agent_role"`
	IterationID     string `json:"iteration_id"`
}

func NewTaskEvaluationStartedEvent(source interface{}, taskID, taskDescription, agentRole, iterationID string) *TaskEvaluationStartedEvent {
	return &TaskEvaluationStartedEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypeTaskEvaluationStarted,
			Timestamp: time.Now(),
			Source:    source,
			Payload: map[string]interface{}{
				"task_id":          taskID,
				"task_description": taskDescription,
				"agent_role":       agentRole,
				"iteration_id":     iterationID,
			},
		},
		TaskID:          taskID,
		TaskDescription: taskDescription,
		AgentRole:       agentRole,
		IterationID:     iterationID,
	}
}

// AgentEvaluationStartedEvent Agent评估开始事件
type AgentEvaluationStartedEvent struct {
	events.BaseEvent
	AgentID     string `json:"agent_id"`
	AgentRole   string `json:"agent_role"`
	TaskID      string `json:"task_id,omitempty"`
	Iteration   int    `json:"iteration"`
	IterationID string `json:"iteration_id"`
}

func NewAgentEvaluationStartedEvent(source interface{}, agentID, agentRole, taskID string, iteration int, iterationID string) *AgentEvaluationStartedEvent {
	return &AgentEvaluationStartedEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypeAgentEvaluationStarted,
			Timestamp: time.Now(),
			Source:    source,
			Payload: map[string]interface{}{
				"agent_id":     agentID,
				"agent_role":   agentRole,
				"task_id":      taskID,
				"iteration":    iteration,
				"iteration_id": iterationID,
			},
		},
		AgentID:     agentID,
		AgentRole:   agentRole,
		TaskID:      taskID,
		Iteration:   iteration,
		IterationID: iterationID,
	}
}

// AgentEvaluationCompletedEvent Agent评估完成事件
type AgentEvaluationCompletedEvent struct {
	events.BaseEvent
	AgentID         string           `json:"agent_id"`
	AgentRole       string           `json:"agent_role"`
	TaskID          string           `json:"task_id,omitempty"`
	Iteration       int              `json:"iteration"`
	IterationID     string           `json:"iteration_id"`
	MetricCategory  MetricCategory   `json:"metric_category,omitempty"`
	Score           *EvaluationScore `json:"score,omitempty"`
	ExecutionTimeMs float64          `json:"execution_time_ms"`
}

func NewAgentEvaluationCompletedEvent(source interface{}, agentID, agentRole, taskID string, iteration int, iterationID string, metricCategory MetricCategory, score *EvaluationScore, executionTimeMs float64) *AgentEvaluationCompletedEvent {
	return &AgentEvaluationCompletedEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypeAgentEvaluationCompleted,
			Timestamp: time.Now(),
			Source:    source,
			Payload: map[string]interface{}{
				"agent_id":          agentID,
				"agent_role":        agentRole,
				"task_id":           taskID,
				"iteration":         iteration,
				"iteration_id":      iterationID,
				"metric_category":   string(metricCategory),
				"execution_time_ms": executionTimeMs,
			},
		},
		AgentID:         agentID,
		AgentRole:       agentRole,
		TaskID:          taskID,
		Iteration:       iteration,
		IterationID:     iterationID,
		MetricCategory:  metricCategory,
		Score:           score,
		ExecutionTimeMs: executionTimeMs,
	}
}

// AgentEvaluationFailedEvent Agent评估失败事件
type AgentEvaluationFailedEvent struct {
	events.BaseEvent
	AgentID         string  `json:"agent_id"`
	AgentRole       string  `json:"agent_role"`
	TaskID          string  `json:"task_id,omitempty"`
	Iteration       int     `json:"iteration"`
	IterationID     string  `json:"iteration_id"`
	Error           string  `json:"error"`
	ExecutionTimeMs float64 `json:"execution_time_ms"`
}

func NewAgentEvaluationFailedEvent(source interface{}, agentID, agentRole, taskID string, iteration int, iterationID, errorMsg string, executionTimeMs float64) *AgentEvaluationFailedEvent {
	return &AgentEvaluationFailedEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypeAgentEvaluationFailed,
			Timestamp: time.Now(),
			Source:    source,
			Payload: map[string]interface{}{
				"agent_id":          agentID,
				"agent_role":        agentRole,
				"task_id":           taskID,
				"iteration":         iteration,
				"iteration_id":      iterationID,
				"error":             errorMsg,
				"execution_time_ms": executionTimeMs,
			},
		},
		AgentID:         agentID,
		AgentRole:       agentRole,
		TaskID:          taskID,
		Iteration:       iteration,
		IterationID:     iterationID,
		Error:           errorMsg,
		ExecutionTimeMs: executionTimeMs,
	}
}

// EvaluationSessionStartedEvent 评估会话开始事件
type EvaluationSessionStartedEvent struct {
	events.BaseEvent
	SessionID      string                 `json:"session_id"`
	EvaluationType string                 `json:"evaluation_type"`
	Config         *EvaluationConfig      `json:"config,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

func NewEvaluationSessionStartedEvent(source interface{}, sessionID, evaluationType string, config *EvaluationConfig, metadata map[string]interface{}) *EvaluationSessionStartedEvent {
	return &EvaluationSessionStartedEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypeEvaluationSessionStarted,
			Timestamp: time.Now(),
			Source:    source,
			Payload: map[string]interface{}{
				"session_id":      sessionID,
				"evaluation_type": evaluationType,
			},
		},
		SessionID:      sessionID,
		EvaluationType: evaluationType,
		Config:         config,
		Metadata:       metadata,
	}
}

// EvaluationSessionCompletedEvent 评估会话完成事件
type EvaluationSessionCompletedEvent struct {
	events.BaseEvent
	SessionID             string  `json:"session_id"`
	EvaluationType        string  `json:"evaluation_type"`
	TotalEvaluations      int     `json:"total_evaluations"`
	SuccessfulEvaluations int     `json:"successful_evaluations"`
	FailedEvaluations     int     `json:"failed_evaluations"`
	AverageScore          float64 `json:"average_score"`
	TotalExecutionTimeMs  float64 `json:"total_execution_time_ms"`
	SuccessRate           float64 `json:"success_rate"`
}

func NewEvaluationSessionCompletedEvent(source interface{}, sessionID, evaluationType string, totalEvals, successfulEvals, failedEvals int, avgScore, totalExecTime float64) *EvaluationSessionCompletedEvent {
	successRate := 0.0
	if totalEvals > 0 {
		successRate = float64(successfulEvals) / float64(totalEvals) * 100.0
	}

	return &EvaluationSessionCompletedEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypeEvaluationSessionCompleted,
			Timestamp: time.Now(),
			Source:    source,
			Payload: map[string]interface{}{
				"session_id":              sessionID,
				"evaluation_type":         evaluationType,
				"total_evaluations":       totalEvals,
				"successful_evaluations":  successfulEvals,
				"failed_evaluations":      failedEvals,
				"average_score":           avgScore,
				"total_execution_time_ms": totalExecTime,
				"success_rate":            successRate,
			},
		},
		SessionID:             sessionID,
		EvaluationType:        evaluationType,
		TotalEvaluations:      totalEvals,
		SuccessfulEvaluations: successfulEvals,
		FailedEvaluations:     failedEvals,
		AverageScore:          avgScore,
		TotalExecutionTimeMs:  totalExecTime,
		SuccessRate:           successRate,
	}
}

