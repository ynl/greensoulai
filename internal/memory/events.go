package memory

import (
	"time"

	"github.com/ynl/greensoulai/pkg/events"
)

// MemorySaveStartedEvent 记忆保存开始事件
type MemorySaveStartedEvent struct {
	events.BaseEvent
	Agent string      `json:"agent"`
	Value interface{} `json:"value"`
}

// NewMemorySaveStartedEvent 创建记忆保存开始事件
func NewMemorySaveStartedEvent(agent string, value interface{}) *MemorySaveStartedEvent {
	return &MemorySaveStartedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "memory_save_started",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"agent": agent,
				"value": value,
			},
		},
		Agent: agent,
		Value: value,
	}
}

// MemorySaveCompletedEvent 记忆保存完成事件
type MemorySaveCompletedEvent struct {
	events.BaseEvent
	Agent    string `json:"agent"`
	MemoryID string `json:"memory_id"`
}

// NewMemorySaveCompletedEvent 创建记忆保存完成事件
func NewMemorySaveCompletedEvent(agent, memoryID string) *MemorySaveCompletedEvent {
	return &MemorySaveCompletedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "memory_save_completed",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"agent":     agent,
				"memory_id": memoryID,
			},
		},
		Agent:    agent,
		MemoryID: memoryID,
	}
}

// MemorySaveFailedEvent 记忆保存失败事件
type MemorySaveFailedEvent struct {
	events.BaseEvent
	Agent string `json:"agent"`
	Error string `json:"error"`
}

// NewMemorySaveFailedEvent 创建记忆保存失败事件
func NewMemorySaveFailedEvent(agent, errorMsg string) *MemorySaveFailedEvent {
	return &MemorySaveFailedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "memory_save_failed",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"agent": agent,
				"error": errorMsg,
			},
		},
		Agent: agent,
		Error: errorMsg,
	}
}

// MemoryQueryStartedEvent 记忆查询开始事件
type MemoryQueryStartedEvent struct {
	events.BaseEvent
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

// NewMemoryQueryStartedEvent 创建记忆查询开始事件
func NewMemoryQueryStartedEvent(query string, limit int) *MemoryQueryStartedEvent {
	return &MemoryQueryStartedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "memory_query_started",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"query": query,
				"limit": limit,
			},
		},
		Query: query,
		Limit: limit,
	}
}

// MemoryQueryCompletedEvent 记忆查询完成事件
type MemoryQueryCompletedEvent struct {
	events.BaseEvent
	Query        string `json:"query"`
	ResultsCount int    `json:"results_count"`
}

// NewMemoryQueryCompletedEvent 创建记忆查询完成事件
func NewMemoryQueryCompletedEvent(query string, resultsCount int) *MemoryQueryCompletedEvent {
	return &MemoryQueryCompletedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "memory_query_completed",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"query":         query,
				"results_count": resultsCount,
			},
		},
		Query:        query,
		ResultsCount: resultsCount,
	}
}

// MemoryQueryFailedEvent 记忆查询失败事件
type MemoryQueryFailedEvent struct {
	events.BaseEvent
	Query string `json:"query"`
	Error string `json:"error"`
}

// NewMemoryQueryFailedEvent 创建记忆查询失败事件
func NewMemoryQueryFailedEvent(query, errorMsg string) *MemoryQueryFailedEvent {
	return &MemoryQueryFailedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "memory_query_failed",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"query": query,
				"error": errorMsg,
			},
		},
		Query: query,
		Error: errorMsg,
	}
}

