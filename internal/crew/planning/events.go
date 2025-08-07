package planning

import (
	"time"

	"github.com/ynl/greensoulai/pkg/events"
)

// 规划系统事件类型常量
const (
	EventTypePlanningStarted      = "planning.started"
	EventTypePlanningCompleted    = "planning.completed"
	EventTypePlanningFailed       = "planning.failed"
	EventTypeTaskSummaryGenerated = "planning.task_summary.generated"
	EventTypePlanningAgentCreated = "planning.agent.created"
	EventTypePlanningTaskCreated  = "planning.task.created"
	EventTypePlanValidated        = "planning.plan.validated"
	EventTypePlanningRetry        = "planning.retry"
	EventTypePlanningTimeout      = "planning.timeout"
)

// PlanningStartedEvent 规划开始事件
type PlanningStartedEvent struct {
	events.BaseEvent
	TaskCount       int                    `json:"task_count"`
	PlanningLLM     string                 `json:"planning_llm"`
	Configuration   *PlanningConfig        `json:"configuration,omitempty"`
	Context         map[string]interface{} `json:"context,omitempty"`
	CrewID          string                 `json:"crew_id,omitempty"`
	PlanningAgentID string                 `json:"planning_agent_id,omitempty"`
}

// PlanningCompletedEvent 规划完成事件
type PlanningCompletedEvent struct {
	events.BaseEvent
	TaskCount       int                        `json:"task_count"`
	PlansGenerated  int                        `json:"plans_generated"`
	ExecutionTime   float64                    `json:"execution_time_ms"`
	TokensUsed      int                        `json:"tokens_used,omitempty"`
	ModelUsed       string                     `json:"model_used,omitempty"`
	Success         bool                       `json:"success"`
	RetryCount      int                        `json:"retry_count"`
	Result          *PlannerTaskPydanticOutput `json:"result,omitempty"`
	CrewID          string                     `json:"crew_id,omitempty"`
	PlanningAgentID string                     `json:"planning_agent_id,omitempty"`
}

// PlanningFailedEvent 规划失败事件
type PlanningFailedEvent struct {
	events.BaseEvent
	TaskCount       int     `json:"task_count"`
	ErrorMessage    string  `json:"error_message"`
	ErrorType       string  `json:"error_type"`
	Phase           string  `json:"phase"`
	ExecutionTime   float64 `json:"execution_time_ms"`
	RetryCount      int     `json:"retry_count"`
	MaxRetries      int     `json:"max_retries"`
	CrewID          string  `json:"crew_id,omitempty"`
	PlanningAgentID string  `json:"planning_agent_id,omitempty"`
}

// TaskSummaryGeneratedEvent 任务摘要生成事件
type TaskSummaryGeneratedEvent struct {
	events.BaseEvent
	TaskIndex      int          `json:"task_index"`
	TaskID         string       `json:"task_id"`
	TaskSummary    *TaskSummary `json:"task_summary"`
	GeneratedText  string       `json:"generated_text,omitempty"`
	ProcessingTime float64      `json:"processing_time_ms"`
}

// PlanningAgentCreatedEvent 规划代理创建事件
type PlanningAgentCreatedEvent struct {
	events.BaseEvent
	AgentID      string  `json:"agent_id"`
	AgentRole    string  `json:"agent_role"`
	AgentGoal    string  `json:"agent_goal,omitempty"`
	LLMUsed      string  `json:"llm_used"`
	CreationTime float64 `json:"creation_time_ms"`
}

// PlanningTaskCreatedEvent 规划任务创建事件
type PlanningTaskCreatedEvent struct {
	events.BaseEvent
	TaskID       string  `json:"task_id"`
	TaskType     string  `json:"task_type"`
	AgentID      string  `json:"agent_id"`
	TasksSummary string  `json:"tasks_summary,omitempty"`
	CreationTime float64 `json:"creation_time_ms"`
}

// PlanValidatedEvent 计划验证事件
type PlanValidatedEvent struct {
	events.BaseEvent
	PlanCount      int     `json:"plan_count"`
	ValidationTime float64 `json:"validation_time_ms"`
	Success        bool    `json:"success"`
	ErrorMessage   string  `json:"error_message,omitempty"`
}

// PlanningRetryEvent 规划重试事件
type PlanningRetryEvent struct {
	events.BaseEvent
	RetryCount   int    `json:"retry_count"`
	MaxRetries   int    `json:"max_retries"`
	LastError    string `json:"last_error"`
	RetryReason  string `json:"retry_reason"`
	DelaySeconds int    `json:"delay_seconds"`
}

// PlanningTimeoutEvent 规划超时事件
type PlanningTimeoutEvent struct {
	events.BaseEvent
	TimeoutSeconds int    `json:"timeout_seconds"`
	ElapsedSeconds int    `json:"elapsed_seconds"`
	Phase          string `json:"phase"`
	TaskCount      int    `json:"task_count"`
}

// NewPlanningStartedEvent 创建规划开始事件
func NewPlanningStartedEvent(taskCount int, planningLLM string, config *PlanningConfig) *PlanningStartedEvent {
	return &PlanningStartedEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypePlanningStarted,
			Timestamp: time.Now(),
			Source:    "crew.planning",
		},
		TaskCount:     taskCount,
		PlanningLLM:   planningLLM,
		Configuration: config,
	}
}

// NewPlanningCompletedEvent 创建规划完成事件
func NewPlanningCompletedEvent(taskCount int, plansGenerated int, executionTime float64, success bool) *PlanningCompletedEvent {
	return &PlanningCompletedEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypePlanningCompleted,
			Timestamp: time.Now(),
			Source:    "crew.planning",
		},
		TaskCount:      taskCount,
		PlansGenerated: plansGenerated,
		ExecutionTime:  executionTime,
		Success:        success,
	}
}

// NewPlanningFailedEvent 创建规划失败事件
func NewPlanningFailedEvent(taskCount int, errorMessage string, errorType string, phase string, executionTime float64) *PlanningFailedEvent {
	return &PlanningFailedEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypePlanningFailed,
			Timestamp: time.Now(),
			Source:    "crew.planning",
		},
		TaskCount:     taskCount,
		ErrorMessage:  errorMessage,
		ErrorType:     errorType,
		Phase:         phase,
		ExecutionTime: executionTime,
	}
}

// NewTaskSummaryGeneratedEvent 创建任务摘要生成事件
func NewTaskSummaryGeneratedEvent(taskIndex int, taskID string, taskSummary *TaskSummary, processingTime float64) *TaskSummaryGeneratedEvent {
	return &TaskSummaryGeneratedEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypeTaskSummaryGenerated,
			Timestamp: time.Now(),
			Source:    "crew.planning.task_summary",
		},
		TaskIndex:      taskIndex,
		TaskID:         taskID,
		TaskSummary:    taskSummary,
		ProcessingTime: processingTime,
	}
}

// NewPlanningAgentCreatedEvent 创建规划代理创建事件
func NewPlanningAgentCreatedEvent(agentID string, agentRole string, llmUsed string, creationTime float64) *PlanningAgentCreatedEvent {
	return &PlanningAgentCreatedEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypePlanningAgentCreated,
			Timestamp: time.Now(),
			Source:    "crew.planning.agent",
		},
		AgentID:      agentID,
		AgentRole:    agentRole,
		LLMUsed:      llmUsed,
		CreationTime: creationTime,
	}
}

// NewPlanningTaskCreatedEvent 创建规划任务创建事件
func NewPlanningTaskCreatedEvent(taskID string, taskType string, agentID string, creationTime float64) *PlanningTaskCreatedEvent {
	return &PlanningTaskCreatedEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypePlanningTaskCreated,
			Timestamp: time.Now(),
			Source:    "crew.planning.task",
		},
		TaskID:       taskID,
		TaskType:     taskType,
		AgentID:      agentID,
		CreationTime: creationTime,
	}
}

// NewPlanValidatedEvent 创建计划验证事件
func NewPlanValidatedEvent(planCount int, validationTime float64, success bool, errorMessage string) *PlanValidatedEvent {
	return &PlanValidatedEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypePlanValidated,
			Timestamp: time.Now(),
			Source:    "crew.planning.validator",
		},
		PlanCount:      planCount,
		ValidationTime: validationTime,
		Success:        success,
		ErrorMessage:   errorMessage,
	}
}

// NewPlanningRetryEvent 创建规划重试事件
func NewPlanningRetryEvent(retryCount int, maxRetries int, lastError string, retryReason string) *PlanningRetryEvent {
	return &PlanningRetryEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypePlanningRetry,
			Timestamp: time.Now(),
			Source:    "crew.planning.retry",
		},
		RetryCount:  retryCount,
		MaxRetries:  maxRetries,
		LastError:   lastError,
		RetryReason: retryReason,
	}
}

// NewPlanningTimeoutEvent 创建规划超时事件
func NewPlanningTimeoutEvent(timeoutSeconds int, elapsedSeconds int, phase string, taskCount int) *PlanningTimeoutEvent {
	return &PlanningTimeoutEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypePlanningTimeout,
			Timestamp: time.Now(),
			Source:    "crew.planning.timeout",
		},
		TimeoutSeconds: timeoutSeconds,
		ElapsedSeconds: elapsedSeconds,
		Phase:          phase,
		TaskCount:      taskCount,
	}
}
