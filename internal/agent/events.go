package agent

import (
	"time"

	"github.com/ynl/greensoulai/pkg/events"
)

// AgentExecutionStartedEvent 代表Agent开始执行任务的事件
type AgentExecutionStartedEvent struct {
	events.BaseEvent
	AgentID     string `json:"agent_id"`
	Agent       string `json:"agent"`
	TaskID      string `json:"task_id"`
	Task        string `json:"task"`
	ExecutionID int    `json:"execution_id"`
}

// AgentExecutionCompletedEvent 代表Agent完成任务执行的事件
type AgentExecutionCompletedEvent struct {
	events.BaseEvent
	AgentID     string        `json:"agent_id"`
	Agent       string        `json:"agent"`
	TaskID      string        `json:"task_id"`
	Task        string        `json:"task"`
	ExecutionID int           `json:"execution_id"`
	Duration    time.Duration `json:"duration"`
	Success     bool          `json:"success"`
	Output      *TaskOutput   `json:"output,omitempty"`
}

// AgentExecutionFailedEvent 代表Agent任务执行失败的事件
type AgentExecutionFailedEvent struct {
	events.BaseEvent
	AgentID     string        `json:"agent_id"`
	Agent       string        `json:"agent"`
	TaskID      string        `json:"task_id"`
	Task        string        `json:"task"`
	ExecutionID int           `json:"execution_id"`
	Duration    time.Duration `json:"duration"`
	Error       string        `json:"error"`
}

// AgentToolUsageStartedEvent 代表Agent开始使用工具的事件
type AgentToolUsageStartedEvent struct {
	events.BaseEvent
	AgentID  string                 `json:"agent_id"`
	Agent    string                 `json:"agent"`
	TaskID   string                 `json:"task_id"`
	ToolName string                 `json:"tool_name"`
	Args     map[string]interface{} `json:"args"`
}

// AgentToolUsageCompletedEvent 代表Agent完成工具使用的事件
type AgentToolUsageCompletedEvent struct {
	events.BaseEvent
	AgentID  string        `json:"agent_id"`
	Agent    string        `json:"agent"`
	TaskID   string        `json:"task_id"`
	ToolName string        `json:"tool_name"`
	Duration time.Duration `json:"duration"`
	Success  bool          `json:"success"`
	Output   interface{}   `json:"output,omitempty"`
	Error    string        `json:"error,omitempty"`
}

// AgentMemoryRetrievalStartedEvent 代表Agent开始检索记忆的事件
type AgentMemoryRetrievalStartedEvent struct {
	events.BaseEvent
	AgentID string `json:"agent_id"`
	Agent   string `json:"agent"`
	TaskID  string `json:"task_id"`
	Query   string `json:"query"`
}

// AgentMemoryRetrievalCompletedEvent 代表Agent完成记忆检索的事件
type AgentMemoryRetrievalCompletedEvent struct {
	events.BaseEvent
	AgentID     string        `json:"agent_id"`
	Agent       string        `json:"agent"`
	TaskID      string        `json:"task_id"`
	Query       string        `json:"query"`
	ResultCount int           `json:"result_count"`
	Duration    time.Duration `json:"duration"`
}

// AgentKnowledgeQueryStartedEvent 代表Agent开始查询知识的事件
type AgentKnowledgeQueryStartedEvent struct {
	events.BaseEvent
	AgentID string `json:"agent_id"`
	Agent   string `json:"agent"`
	TaskID  string `json:"task_id"`
	Source  string `json:"source"`
	Query   string `json:"query"`
}

// AgentKnowledgeQueryCompletedEvent 代表Agent完成知识查询的事件
type AgentKnowledgeQueryCompletedEvent struct {
	events.BaseEvent
	AgentID     string        `json:"agent_id"`
	Agent       string        `json:"agent"`
	TaskID      string        `json:"task_id"`
	Source      string        `json:"source"`
	Query       string        `json:"query"`
	ResultCount int           `json:"result_count"`
	Duration    time.Duration `json:"duration"`
}

// AgentHumanInputRequestedEvent 代表Agent请求人工输入的事件
type AgentHumanInputRequestedEvent struct {
	events.BaseEvent
	AgentID string   `json:"agent_id"`
	Agent   string   `json:"agent"`
	TaskID  string   `json:"task_id"`
	Prompt  string   `json:"prompt"`
	Options []string `json:"options,omitempty"`
}

// AgentHumanInputReceivedEvent 代表Agent收到人工输入的事件
type AgentHumanInputReceivedEvent struct {
	events.BaseEvent
	AgentID  string        `json:"agent_id"`
	Agent    string        `json:"agent"`
	TaskID   string        `json:"task_id"`
	Input    string        `json:"input"`
	Duration time.Duration `json:"duration"`
}

// NewAgentExecutionStartedEvent 创建Agent执行开始事件
func NewAgentExecutionStartedEvent(agentID, agent, taskID, task string, executionID int) *AgentExecutionStartedEvent {
	return &AgentExecutionStartedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "agent_execution_started",
			Timestamp: time.Now(),
			Source:    agent,
			Payload: map[string]interface{}{
				"agent_id":     agentID,
				"agent":        agent,
				"task_id":      taskID,
				"task":         task,
				"execution_id": executionID,
			},
		},
		AgentID:     agentID,
		Agent:       agent,
		TaskID:      taskID,
		Task:        task,
		ExecutionID: executionID,
	}
}

// NewAgentExecutionCompletedEvent 创建Agent执行完成事件
func NewAgentExecutionCompletedEvent(agentID, agent, taskID, task string, executionID int, duration time.Duration, success bool, output *TaskOutput) *AgentExecutionCompletedEvent {
	payload := map[string]interface{}{
		"agent_id":     agentID,
		"agent":        agent,
		"task_id":      taskID,
		"task":         task,
		"execution_id": executionID,
		"duration_ms":  duration.Milliseconds(),
		"success":      success,
	}

	if output != nil {
		payload["tokens_used"] = output.TokensUsed
		payload["cost"] = output.Cost
		payload["model"] = output.Model
		payload["tools_used"] = output.ToolsUsed
	}

	return &AgentExecutionCompletedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "agent_execution_completed",
			Timestamp: time.Now(),
			Source:    agent,
			Payload:   payload,
		},
		AgentID:     agentID,
		Agent:       agent,
		TaskID:      taskID,
		Task:        task,
		ExecutionID: executionID,
		Duration:    duration,
		Success:     success,
		Output:      output,
	}
}

// NewAgentExecutionFailedEvent 创建Agent执行失败事件
func NewAgentExecutionFailedEvent(agentID, agent, taskID, task string, executionID int, duration time.Duration, err error) *AgentExecutionFailedEvent {
	return &AgentExecutionFailedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "agent_execution_failed",
			Timestamp: time.Now(),
			Source:    agent,
			Payload: map[string]interface{}{
				"agent_id":     agentID,
				"agent":        agent,
				"task_id":      taskID,
				"task":         task,
				"execution_id": executionID,
				"duration_ms":  duration.Milliseconds(),
				"error":        err.Error(),
			},
		},
		AgentID:     agentID,
		Agent:       agent,
		TaskID:      taskID,
		Task:        task,
		ExecutionID: executionID,
		Duration:    duration,
		Error:       err.Error(),
	}
}

// NewAgentToolUsageStartedEvent 创建Agent工具使用开始事件
func NewAgentToolUsageStartedEvent(agentID, agent, taskID, toolName string, args map[string]interface{}) *AgentToolUsageStartedEvent {
	return &AgentToolUsageStartedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "agent_tool_usage_started",
			Timestamp: time.Now(),
			Source:    agent,
			Payload: map[string]interface{}{
				"agent_id":  agentID,
				"agent":     agent,
				"task_id":   taskID,
				"tool_name": toolName,
				"args":      args,
			},
		},
		AgentID:  agentID,
		Agent:    agent,
		TaskID:   taskID,
		ToolName: toolName,
		Args:     args,
	}
}

// NewAgentToolUsageCompletedEvent 创建Agent工具使用完成事件
func NewAgentToolUsageCompletedEvent(agentID, agent, taskID, toolName string, duration time.Duration, success bool, output interface{}, err error) *AgentToolUsageCompletedEvent {
	payload := map[string]interface{}{
		"agent_id":    agentID,
		"agent":       agent,
		"task_id":     taskID,
		"tool_name":   toolName,
		"duration_ms": duration.Milliseconds(),
		"success":     success,
	}

	event := &AgentToolUsageCompletedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "agent_tool_usage_completed",
			Timestamp: time.Now(),
			Source:    agent,
			Payload:   payload,
		},
		AgentID:  agentID,
		Agent:    agent,
		TaskID:   taskID,
		ToolName: toolName,
		Duration: duration,
		Success:  success,
	}

	if success && output != nil {
		event.Output = output
		payload["output"] = output
	}

	if !success && err != nil {
		event.Error = err.Error()
		payload["error"] = err.Error()
	}

	return event
}

// NewAgentMemoryRetrievalStartedEvent 创建Agent记忆检索开始事件
func NewAgentMemoryRetrievalStartedEvent(agentID, agent, taskID, query string) *AgentMemoryRetrievalStartedEvent {
	return &AgentMemoryRetrievalStartedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "agent_memory_retrieval_started",
			Timestamp: time.Now(),
			Source:    agent,
			Payload: map[string]interface{}{
				"agent_id": agentID,
				"agent":    agent,
				"task_id":  taskID,
				"query":    query,
			},
		},
		AgentID: agentID,
		Agent:   agent,
		TaskID:  taskID,
		Query:   query,
	}
}

// NewAgentMemoryRetrievalCompletedEvent 创建Agent记忆检索完成事件
func NewAgentMemoryRetrievalCompletedEvent(agentID, agent, taskID, query string, resultCount int, duration time.Duration) *AgentMemoryRetrievalCompletedEvent {
	return &AgentMemoryRetrievalCompletedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "agent_memory_retrieval_completed",
			Timestamp: time.Now(),
			Source:    agent,
			Payload: map[string]interface{}{
				"agent_id":     agentID,
				"agent":        agent,
				"task_id":      taskID,
				"query":        query,
				"result_count": resultCount,
				"duration_ms":  duration.Milliseconds(),
			},
		},
		AgentID:     agentID,
		Agent:       agent,
		TaskID:      taskID,
		Query:       query,
		ResultCount: resultCount,
		Duration:    duration,
	}
}

// NewAgentKnowledgeQueryStartedEvent 创建Agent知识查询开始事件
func NewAgentKnowledgeQueryStartedEvent(agentID, agent, taskID, source, query string) *AgentKnowledgeQueryStartedEvent {
	return &AgentKnowledgeQueryStartedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "agent_knowledge_query_started",
			Timestamp: time.Now(),
			Source:    agent,
			Payload: map[string]interface{}{
				"agent_id": agentID,
				"agent":    agent,
				"task_id":  taskID,
				"source":   source,
				"query":    query,
			},
		},
		AgentID: agentID,
		Agent:   agent,
		TaskID:  taskID,
		Source:  source,
		Query:   query,
	}
}

// NewAgentKnowledgeQueryCompletedEvent 创建Agent知识查询完成事件
func NewAgentKnowledgeQueryCompletedEvent(agentID, agent, taskID, source, query string, resultCount int, duration time.Duration) *AgentKnowledgeQueryCompletedEvent {
	return &AgentKnowledgeQueryCompletedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "agent_knowledge_query_completed",
			Timestamp: time.Now(),
			Source:    agent,
			Payload: map[string]interface{}{
				"agent_id":     agentID,
				"agent":        agent,
				"task_id":      taskID,
				"source":       source,
				"query":        query,
				"result_count": resultCount,
				"duration_ms":  duration.Milliseconds(),
			},
		},
		AgentID:     agentID,
		Agent:       agent,
		TaskID:      taskID,
		Source:      source,
		Query:       query,
		ResultCount: resultCount,
		Duration:    duration,
	}
}

// NewAgentHumanInputRequestedEvent 创建Agent请求人工输入事件
func NewAgentHumanInputRequestedEvent(agentID, agent, taskID, prompt string, options []string) *AgentHumanInputRequestedEvent {
	return &AgentHumanInputRequestedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "agent_human_input_requested",
			Timestamp: time.Now(),
			Source:    agent,
			Payload: map[string]interface{}{
				"agent_id": agentID,
				"agent":    agent,
				"task_id":  taskID,
				"prompt":   prompt,
				"options":  options,
			},
		},
		AgentID: agentID,
		Agent:   agent,
		TaskID:  taskID,
		Prompt:  prompt,
		Options: options,
	}
}

// NewAgentHumanInputReceivedEvent 创建Agent收到人工输入事件
func NewAgentHumanInputReceivedEvent(agentID, agent, taskID, input string, duration time.Duration) *AgentHumanInputReceivedEvent {
	return &AgentHumanInputReceivedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "agent_human_input_received",
			Timestamp: time.Now(),
			Source:    agent,
			Payload: map[string]interface{}{
				"agent_id":     agentID,
				"agent":        agent,
				"task_id":      taskID,
				"input_length": len(input),
				"duration_ms":  duration.Milliseconds(),
			},
		},
		AgentID:  agentID,
		Agent:    agent,
		TaskID:   taskID,
		Input:    input,
		Duration: duration,
	}
}
