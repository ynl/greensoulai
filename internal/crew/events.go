package crew

import (
	"time"

	"github.com/ynl/greensoulai/pkg/events"
)

// CrewKickoffStartedEvent Crew启动事件
type CrewKickoffStartedEvent struct {
	events.BaseEvent
	CrewID      string `json:"crew_id"`
	CrewName    string `json:"crew_name"`
	ExecutionID int    `json:"execution_id"`
	Process     string `json:"process"`
}

// NewCrewKickoffStartedEvent 创建Crew启动事件
func NewCrewKickoffStartedEvent(crewID, crewName string, executionID int, process string) *CrewKickoffStartedEvent {
	return &CrewKickoffStartedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "crew_kickoff_started",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"crew_id":      crewID,
				"crew_name":    crewName,
				"execution_id": executionID,
				"process":      process,
			},
		},
		CrewID:      crewID,
		CrewName:    crewName,
		ExecutionID: executionID,
		Process:     process,
	}
}

// CrewKickoffCompletedEvent Crew完成事件
type CrewKickoffCompletedEvent struct {
	events.BaseEvent
	CrewID      string        `json:"crew_id"`
	CrewName    string        `json:"crew_name"`
	ExecutionID int           `json:"execution_id"`
	Duration    time.Duration `json:"duration"`
	Success     bool          `json:"success"`
}

// NewCrewKickoffCompletedEvent 创建Crew完成事件
func NewCrewKickoffCompletedEvent(crewID, crewName string, executionID int, duration time.Duration, success bool) *CrewKickoffCompletedEvent {
	return &CrewKickoffCompletedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "crew_kickoff_completed",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"crew_id":      crewID,
				"crew_name":    crewName,
				"execution_id": executionID,
				"duration_ms":  duration.Milliseconds(),
				"success":      success,
			},
		},
		CrewID:      crewID,
		CrewName:    crewName,
		ExecutionID: executionID,
		Duration:    duration,
		Success:     success,
	}
}

// TaskExecutionStartedEvent 任务开始执行事件
type TaskExecutionStartedEvent struct {
	events.BaseEvent
	TaskIndex       int    `json:"task_index"`
	TaskDescription string `json:"task_description"`
	AgentRole       string `json:"agent_role"`
}

// NewTaskExecutionStartedEvent 创建任务开始执行事件
func NewTaskExecutionStartedEvent(taskIndex int, taskDescription, agentRole string) *TaskExecutionStartedEvent {
	return &TaskExecutionStartedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "task_execution_started",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"task_index":       taskIndex,
				"task_description": taskDescription,
				"agent_role":       agentRole,
			},
		},
		TaskIndex:       taskIndex,
		TaskDescription: taskDescription,
		AgentRole:       agentRole,
	}
}

// TaskExecutionCompletedEvent 任务完成执行事件
type TaskExecutionCompletedEvent struct {
	events.BaseEvent
	TaskIndex       int           `json:"task_index"`
	TaskDescription string        `json:"task_description"`
	AgentRole       string        `json:"agent_role"`
	Duration        time.Duration `json:"duration"`
	Success         bool          `json:"success"`
}

// NewTaskExecutionCompletedEvent 创建任务完成执行事件
func NewTaskExecutionCompletedEvent(taskIndex int, taskDescription, agentRole string, duration time.Duration, success bool) *TaskExecutionCompletedEvent {
	return &TaskExecutionCompletedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "task_execution_completed",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"task_index":       taskIndex,
				"task_description": taskDescription,
				"agent_role":       agentRole,
				"duration_ms":      duration.Milliseconds(),
				"success":          success,
			},
		},
		TaskIndex:       taskIndex,
		TaskDescription: taskDescription,
		AgentRole:       agentRole,
		Duration:        duration,
		Success:         success,
	}
}

// TaskExecutionFailedEvent 任务执行失败事件
type TaskExecutionFailedEvent struct {
	events.BaseEvent
	TaskIndex       int           `json:"task_index"`
	TaskDescription string        `json:"task_description"`
	AgentRole       string        `json:"agent_role"`
	Error           string        `json:"error"`
	Duration        time.Duration `json:"duration"`
}

// NewTaskExecutionFailedEvent 创建任务执行失败事件
func NewTaskExecutionFailedEvent(taskIndex int, taskDescription, agentRole, errorMsg string, duration time.Duration) *TaskExecutionFailedEvent {
	return &TaskExecutionFailedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "task_execution_failed",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"task_index":       taskIndex,
				"task_description": taskDescription,
				"agent_role":       agentRole,
				"error":            errorMsg,
				"duration_ms":      duration.Milliseconds(),
			},
		},
		TaskIndex:       taskIndex,
		TaskDescription: taskDescription,
		AgentRole:       agentRole,
		Error:           errorMsg,
		Duration:        duration,
	}
}

// Sequential Process Events

// SequentialProcessStartedEvent Sequential流程开始事件
type SequentialProcessStartedEvent struct {
	events.BaseEvent
	CrewName    string `json:"crew_name"`
	TasksCount  int    `json:"tasks_count"`
	AgentsCount int    `json:"agents_count"`
}

// NewSequentialProcessStartedEvent 创建Sequential流程开始事件
func NewSequentialProcessStartedEvent(crewName string, tasksCount, agentsCount int) *SequentialProcessStartedEvent {
	return &SequentialProcessStartedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "sequential_process_started",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"crew_name":    crewName,
				"tasks_count":  tasksCount,
				"agents_count": agentsCount,
			},
		},
		CrewName:    crewName,
		TasksCount:  tasksCount,
		AgentsCount: agentsCount,
	}
}

// SequentialProcessCompletedEvent Sequential流程完成事件
type SequentialProcessCompletedEvent struct {
	events.BaseEvent
	CrewName            string `json:"crew_name"`
	CompletedTasksCount int    `json:"completed_tasks_count"`
}

// NewSequentialProcessCompletedEvent 创建Sequential流程完成事件
func NewSequentialProcessCompletedEvent(crewName string, completedTasksCount int) *SequentialProcessCompletedEvent {
	return &SequentialProcessCompletedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "sequential_process_completed",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"crew_name":             crewName,
				"completed_tasks_count": completedTasksCount,
			},
		},
		CrewName:            crewName,
		CompletedTasksCount: completedTasksCount,
	}
}

// SequentialProcessFailedEvent Sequential流程失败事件
type SequentialProcessFailedEvent struct {
	events.BaseEvent
	CrewName string `json:"crew_name"`
	Error    string `json:"error"`
}

// NewSequentialProcessFailedEvent 创建Sequential流程失败事件
func NewSequentialProcessFailedEvent(crewName, errorMsg string) *SequentialProcessFailedEvent {
	return &SequentialProcessFailedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "sequential_process_failed",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"crew_name": crewName,
				"error":     errorMsg,
			},
		},
		CrewName: crewName,
		Error:    errorMsg,
	}
}

// Hierarchical Process Events

// HierarchicalProcessStartedEvent Hierarchical流程开始事件
type HierarchicalProcessStartedEvent struct {
	events.BaseEvent
	CrewName    string `json:"crew_name"`
	TasksCount  int    `json:"tasks_count"`
	AgentsCount int    `json:"agents_count"`
}

// NewHierarchicalProcessStartedEvent 创建Hierarchical流程开始事件
func NewHierarchicalProcessStartedEvent(crewName string, tasksCount, agentsCount int) *HierarchicalProcessStartedEvent {
	return &HierarchicalProcessStartedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "hierarchical_process_started",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"crew_name":    crewName,
				"tasks_count":  tasksCount,
				"agents_count": agentsCount,
			},
		},
		CrewName:    crewName,
		TasksCount:  tasksCount,
		AgentsCount: agentsCount,
	}
}

// HierarchicalProcessCompletedEvent Hierarchical流程完成事件
type HierarchicalProcessCompletedEvent struct {
	events.BaseEvent
	CrewName            string `json:"crew_name"`
	CompletedTasksCount int    `json:"completed_tasks_count"`
}

// NewHierarchicalProcessCompletedEvent 创建Hierarchical流程完成事件
func NewHierarchicalProcessCompletedEvent(crewName string, completedTasksCount int) *HierarchicalProcessCompletedEvent {
	return &HierarchicalProcessCompletedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "hierarchical_process_completed",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"crew_name":             crewName,
				"completed_tasks_count": completedTasksCount,
			},
		},
		CrewName:            crewName,
		CompletedTasksCount: completedTasksCount,
	}
}

// HierarchicalProcessFailedEvent Hierarchical流程失败事件
type HierarchicalProcessFailedEvent struct {
	events.BaseEvent
	CrewName string `json:"crew_name"`
	Error    string `json:"error"`
}

// NewHierarchicalProcessFailedEvent 创建Hierarchical流程失败事件
func NewHierarchicalProcessFailedEvent(crewName, errorMsg string) *HierarchicalProcessFailedEvent {
	return &HierarchicalProcessFailedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "hierarchical_process_failed",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"crew_name": crewName,
				"error":     errorMsg,
			},
		},
		CrewName: crewName,
		Error:    errorMsg,
	}
}
