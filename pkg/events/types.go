package events

import (
	"context"
	"time"
)

// EventType 事件类型常量
const (
	// Agent Events
	EventTypeAgentStarted             = "agent_execution_started"
	EventTypeAgentCompleted           = "agent_execution_completed"
	EventTypeAgentError               = "agent_execution_error"
	EventTypeAgentEvaluationStarted   = "agent_evaluation_started"
	EventTypeAgentEvaluationCompleted = "agent_evaluation_completed"
	EventTypeAgentEvaluationFailed    = "agent_evaluation_failed"

	// Task Events
	EventTypeTaskStarted    = "task_started"
	EventTypeTaskCompleted  = "task_completed"
	EventTypeTaskFailed     = "task_failed"
	EventTypeTaskEvaluation = "task_evaluation"

	// Crew Events
	EventTypeCrewStarted        = "crew_kickoff_started"
	EventTypeCrewCompleted      = "crew_kickoff_completed"
	EventTypeCrewFailed         = "crew_kickoff_failed"
	EventTypeCrewTrainStarted   = "crew_train_started"
	EventTypeCrewTrainCompleted = "crew_train_completed"
	EventTypeCrewTrainFailed    = "crew_train_failed"
	EventTypeCrewTestStarted    = "crew_test_started"
	EventTypeCrewTestCompleted  = "crew_test_completed"
	EventTypeCrewTestFailed     = "crew_test_failed"

	// LLM Events
	EventTypeLLMCallStarted   = "llm_call_started"
	EventTypeLLMCallCompleted = "llm_call_completed"
	EventTypeLLMCallFailed    = "llm_call_failed"
	EventTypeLLMStreamChunk   = "llm_stream_chunk"

	// Tool Events
	EventTypeToolUsageStarted       = "tool_usage_started"
	EventTypeToolUsageFinished      = "tool_usage_finished"
	EventTypeToolUsageError         = "tool_usage_error"
	EventTypeToolExecutionError     = "tool_execution_error"
	EventTypeToolSelectionError     = "tool_selection_error"
	EventTypeToolValidateInputError = "tool_validate_input_error"

	// Memory Events
	EventTypeMemorySaveStarted        = "memory_save_started"
	EventTypeMemorySaveCompleted      = "memory_save_completed"
	EventTypeMemorySaveFailed         = "memory_save_failed"
	EventTypeMemoryQueryStarted       = "memory_query_started"
	EventTypeMemoryQueryCompleted     = "memory_query_completed"
	EventTypeMemoryQueryFailed        = "memory_query_failed"
	EventTypeMemoryRetrievalStarted   = "memory_retrieval_started"
	EventTypeMemoryRetrievalCompleted = "memory_retrieval_completed"

	// Flow Events
	EventTypeFlowCreated             = "flow_created"
	EventTypeFlowStarted             = "flow_started"
	EventTypeFlowFinished            = "flow_finished"
	EventTypeFlowPlot                = "flow_plot"
	EventTypeMethodExecutionStarted  = "method_execution_started"
	EventTypeMethodExecutionFinished = "method_execution_finished"
	EventTypeMethodExecutionFailed   = "method_execution_failed"

	// LLM Guardrail Events
	EventTypeLLMGuardrailStarted   = "llm_guardrail_started"
	EventTypeLLMGuardrailCompleted = "llm_guardrail_completed"
)

// Event 事件接口
type Event interface {
	GetType() string
	GetTimestamp() time.Time
	GetSource() interface{}
	GetPayload() map[string]interface{}
	GetSourceFingerprint() string
	GetSourceType() string
	GetFingerprintMetadata() map[string]interface{}
}

// EventHandler 事件处理器
type EventHandler func(ctx context.Context, event Event) error

// EventBus 事件总线接口
type EventBus interface {
	Emit(ctx context.Context, source interface{}, event Event) error
	Subscribe(eventType string, handler EventHandler) error
	Unsubscribe(eventType string, handler EventHandler) error
	GetHandlerCount(eventType string) int
	GetRegisteredEventTypes() []string
	// 新增方法以匹配crewAI功能
	RegisterHandler(eventType string, handler EventHandler) error
	WithScopedHandlers() EventBus
}

// BaseEvent 基础事件结构 - 更新以匹配crewAI
type BaseEvent struct {
	Type                string                 `json:"type"`
	Timestamp           time.Time              `json:"timestamp"`
	Source              interface{}            `json:"source"`
	Payload             map[string]interface{} `json:"payload"`
	SourceFingerprint   string                 `json:"source_fingerprint,omitempty"`
	SourceType          string                 `json:"source_type,omitempty"`
	FingerprintMetadata map[string]interface{} `json:"fingerprint_metadata,omitempty"`
}

func (e *BaseEvent) GetType() string {
	return e.Type
}

func (e *BaseEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

func (e *BaseEvent) GetSource() interface{} {
	return e.Source
}

func (e *BaseEvent) GetPayload() map[string]interface{} {
	return e.Payload
}

func (e *BaseEvent) GetSourceFingerprint() string {
	return e.SourceFingerprint
}

func (e *BaseEvent) GetSourceType() string {
	return e.SourceType
}

func (e *BaseEvent) GetFingerprintMetadata() map[string]interface{} {
	return e.FingerprintMetadata
}

// 具体事件类型 - 更新以匹配crewAI
type AgentExecutionStartedEvent struct {
	BaseEvent
	Agent      interface{}   `json:"agent"`
	Task       interface{}   `json:"task"`
	Tools      []interface{} `json:"tools,omitempty"`
	TaskPrompt string        `json:"task_prompt"`
}

type AgentExecutionCompletedEvent struct {
	BaseEvent
	Agent  interface{} `json:"agent"`
	Task   interface{} `json:"task"`
	Output string      `json:"output"`
}

type AgentExecutionErrorEvent struct {
	BaseEvent
	Agent interface{} `json:"agent"`
	Task  interface{} `json:"task"`
	Error string      `json:"error"`
}

type TaskStartedEvent struct {
	BaseEvent
	Task  interface{} `json:"task"`
	Agent interface{} `json:"agent"`
}

type TaskCompletedEvent struct {
	BaseEvent
	Task     interface{}   `json:"task"`
	Agent    interface{}   `json:"agent"`
	Output   string        `json:"output"`
	Duration time.Duration `json:"duration"`
}

type TaskFailedEvent struct {
	BaseEvent
	Task  interface{} `json:"task"`
	Agent interface{} `json:"agent"`
	Error string      `json:"error"`
}

type LLMCallStartedEvent struct {
	BaseEvent
	Model        string `json:"model"`
	MessageCount int    `json:"message_count"`
}

type LLMCallCompletedEvent struct {
	BaseEvent
	Model      string        `json:"model"`
	Duration   time.Duration `json:"duration"`
	Success    bool          `json:"success"`
	TokensUsed int           `json:"tokens_used,omitempty"`
}

type LLMCallFailedEvent struct {
	BaseEvent
	Model string `json:"model"`
	Error string `json:"error"`
}

type ToolUsageStartedEvent struct {
	BaseEvent
	ToolName string                 `json:"tool_name"`
	Args     map[string]interface{} `json:"args"`
}

type ToolUsageFinishedEvent struct {
	BaseEvent
	ToolName string        `json:"tool_name"`
	Duration time.Duration `json:"duration"`
	Success  bool          `json:"success"`
	Output   interface{}   `json:"output,omitempty"`
}

type ToolUsageErrorEvent struct {
	BaseEvent
	ToolName string `json:"tool_name"`
	Error    string `json:"error"`
}

type CrewKickoffStartedEvent struct {
	BaseEvent
	Crew interface{} `json:"crew"`
}

type CrewKickoffCompletedEvent struct {
	BaseEvent
	Crew     interface{}   `json:"crew"`
	Duration time.Duration `json:"duration"`
	Success  bool          `json:"success"`
}

type CrewKickoffFailedEvent struct {
	BaseEvent
	Crew  interface{} `json:"crew"`
	Error string      `json:"error"`
}

type MemorySaveStartedEvent struct {
	BaseEvent
	MemoryType string `json:"memory_type"`
	Key        string `json:"key"`
}

type MemorySaveCompletedEvent struct {
	BaseEvent
	MemoryType string `json:"memory_type"`
	Key        string `json:"key"`
	Success    bool   `json:"success"`
}

type MemorySaveFailedEvent struct {
	BaseEvent
	MemoryType string `json:"memory_type"`
	Key        string `json:"key"`
	Error      string `json:"error"`
}

type MemoryQueryStartedEvent struct {
	BaseEvent
	MemoryType string `json:"memory_type"`
	Query      string `json:"query"`
}

type MemoryQueryCompletedEvent struct {
	BaseEvent
	MemoryType string        `json:"memory_type"`
	Query      string        `json:"query"`
	Results    []interface{} `json:"results"`
	Duration   time.Duration `json:"duration"`
}

type MemoryQueryFailedEvent struct {
	BaseEvent
	MemoryType string `json:"memory_type"`
	Query      string `json:"query"`
	Error      string `json:"error"`
}
