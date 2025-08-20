package contextual

import (
	"time"

	"github.com/ynl/greensoulai/pkg/events"
)

// 上下文记忆事件类型常量
const (
	EventTypeContextBuildStarted   = "context_build_started"
	EventTypeContextBuildCompleted = "context_build_completed"
	EventTypeContextBuildFailed    = "context_build_failed"
)

// ContextBuildStartedEvent 上下文构建开始事件
type ContextBuildStartedEvent struct {
	events.BaseEvent
	TaskID    string    `json:"task_id"`
	Query     string    `json:"query"`
	Timestamp time.Time `json:"timestamp"`
}

// NewContextBuildStartedEvent 创建上下文构建开始事件
func NewContextBuildStartedEvent(taskID, query string) *ContextBuildStartedEvent {
	return &ContextBuildStartedEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypeContextBuildStarted,
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"task_id": taskID,
				"query":   query,
			},
		},
		TaskID:    taskID,
		Query:     query,
		Timestamp: time.Now(),
	}
}

// GetTimestamp 获取事件时间戳
func (e *ContextBuildStartedEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

// ContextBuildCompletedEvent 上下文构建完成事件
type ContextBuildCompletedEvent struct {
	events.BaseEvent
	TaskID        string        `json:"task_id"`
	Query         string        `json:"query"`
	PartsCount    int           `json:"parts_count"`    // 上下文部分数量
	ContextLength int           `json:"context_length"` // 最终上下文长度
	Duration      time.Duration `json:"duration"`       // 构建耗时
	Timestamp     time.Time     `json:"timestamp"`
}

// NewContextBuildCompletedEvent 创建上下文构建完成事件
func NewContextBuildCompletedEvent(taskID, query string, partsCount, contextLength int) *ContextBuildCompletedEvent {
	return &ContextBuildCompletedEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypeContextBuildCompleted,
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"task_id":        taskID,
				"query":          query,
				"parts_count":    partsCount,
				"context_length": contextLength,
			},
		},
		TaskID:        taskID,
		Query:         query,
		PartsCount:    partsCount,
		ContextLength: contextLength,
		Timestamp:     time.Now(),
	}
}

// GetTimestamp 获取事件时间戳
func (e *ContextBuildCompletedEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

// ContextBuildFailedEvent 上下文构建失败事件
type ContextBuildFailedEvent struct {
	events.BaseEvent
	TaskID    string    `json:"task_id"`
	Query     string    `json:"query"`
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
}

// NewContextBuildFailedEvent 创建上下文构建失败事件
func NewContextBuildFailedEvent(taskID, query, errorMsg string) *ContextBuildFailedEvent {
	return &ContextBuildFailedEvent{
		BaseEvent: events.BaseEvent{
			Type:      EventTypeContextBuildFailed,
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"task_id": taskID,
				"query":   query,
				"error":   errorMsg,
			},
		},
		TaskID:    taskID,
		Query:     query,
		Error:     errorMsg,
		Timestamp: time.Now(),
	}
}

// GetTimestamp 获取事件时间戳
func (e *ContextBuildFailedEvent) GetTimestamp() time.Time {
	return e.Timestamp
}
