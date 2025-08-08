package knowledge

import (
	"time"

	"github.com/ynl/greensoulai/pkg/events"
)

// KnowledgeQueryStartedEvent 知识查询开始事件
type KnowledgeQueryStartedEvent struct {
	events.BaseEvent
	Collection   string   `json:"collection"`
	Query        []string `json:"query"`
	ResultsLimit int      `json:"results_limit"`
}

// NewKnowledgeQueryStartedEvent 创建知识查询开始事件
func NewKnowledgeQueryStartedEvent(collection string, query []string, resultsLimit int) *KnowledgeQueryStartedEvent {
	return &KnowledgeQueryStartedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "knowledge_query_started",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"collection":    collection,
				"query":         query,
				"results_limit": resultsLimit,
			},
		},
		Collection:   collection,
		Query:        query,
		ResultsLimit: resultsLimit,
	}
}

// KnowledgeQueryCompletedEvent 知识查询完成事件
type KnowledgeQueryCompletedEvent struct {
	events.BaseEvent
	Collection   string   `json:"collection"`
	Query        []string `json:"query"`
	ResultsCount int      `json:"results_count"`
}

// NewKnowledgeQueryCompletedEvent 创建知识查询完成事件
func NewKnowledgeQueryCompletedEvent(collection string, query []string, resultsCount int) *KnowledgeQueryCompletedEvent {
	return &KnowledgeQueryCompletedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "knowledge_query_completed",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"collection":    collection,
				"query":         query,
				"results_count": resultsCount,
			},
		},
		Collection:   collection,
		Query:        query,
		ResultsCount: resultsCount,
	}
}

// KnowledgeQueryFailedEvent 知识查询失败事件
type KnowledgeQueryFailedEvent struct {
	events.BaseEvent
	Collection string   `json:"collection"`
	Query      []string `json:"query"`
	Error      string   `json:"error"`
}

// NewKnowledgeQueryFailedEvent 创建知识查询失败事件
func NewKnowledgeQueryFailedEvent(collection string, query []string, errorMsg string) *KnowledgeQueryFailedEvent {
	return &KnowledgeQueryFailedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "knowledge_query_failed",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"collection": collection,
				"query":      query,
				"error":      errorMsg,
			},
		},
		Collection: collection,
		Query:      query,
		Error:      errorMsg,
	}
}

// KnowledgeSourceAddedEvent 知识源添加事件
type KnowledgeSourceAddedEvent struct {
	events.BaseEvent
	Collection string `json:"collection"`
	SourceName string `json:"source_name"`
	SourceType string `json:"source_type"`
	ChunkCount int    `json:"chunk_count"`
}

// NewKnowledgeSourceAddedEvent 创建知识源添加事件
func NewKnowledgeSourceAddedEvent(collection, sourceName, sourceType string, chunkCount int) *KnowledgeSourceAddedEvent {
	return &KnowledgeSourceAddedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "knowledge_source_added",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"collection":  collection,
				"source_name": sourceName,
				"source_type": sourceType,
				"chunk_count": chunkCount,
			},
		},
		Collection: collection,
		SourceName: sourceName,
		SourceType: sourceType,
		ChunkCount: chunkCount,
	}
}

// KnowledgeSourceFailedEvent 知识源处理失败事件
type KnowledgeSourceFailedEvent struct {
	events.BaseEvent
	Collection string `json:"collection"`
	SourceName string `json:"source_name"`
	SourceType string `json:"source_type"`
	Error      string `json:"error"`
}

// NewKnowledgeSourceFailedEvent 创建知识源处理失败事件
func NewKnowledgeSourceFailedEvent(collection, sourceName, sourceType, errorMsg string) *KnowledgeSourceFailedEvent {
	return &KnowledgeSourceFailedEvent{
		BaseEvent: events.BaseEvent{
			Type:      "knowledge_source_failed",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"collection":  collection,
				"source_name": sourceName,
				"source_type": sourceType,
				"error":       errorMsg,
			},
		},
		Collection: collection,
		SourceName: sourceName,
		SourceType: sourceType,
		Error:      errorMsg,
	}
}

// KnowledgeStorageResetEvent 知识存储重置事件
type KnowledgeStorageResetEvent struct {
	events.BaseEvent
	Collection string `json:"collection"`
}

// NewKnowledgeStorageResetEvent 创建知识存储重置事件
func NewKnowledgeStorageResetEvent(collection string) *KnowledgeStorageResetEvent {
	return &KnowledgeStorageResetEvent{
		BaseEvent: events.BaseEvent{
			Type:      "knowledge_storage_reset",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"collection": collection,
			},
		},
		Collection: collection,
	}
}

