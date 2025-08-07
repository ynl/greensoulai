package memory

import (
	"context"
	"time"

	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// Memory 基础记忆接口
type Memory interface {
	// 保存数据到记忆中
	Save(ctx context.Context, value interface{}, metadata map[string]interface{}, agent string) error

	// 从记忆中搜索相关信息
	Search(ctx context.Context, query string, limit int, scoreThreshold float64) ([]MemoryItem, error)

	// 清理记忆
	Clear(ctx context.Context) error

	// 设置crew引用
	SetCrew(crew interface{}) Memory

	// 关闭记忆系统
	Close() error
}

// MemoryItem 记忆项
type MemoryItem struct {
	ID        string                 `json:"id"`
	Value     interface{}            `json:"value"`
	Metadata  map[string]interface{} `json:"metadata"`
	Agent     string                 `json:"agent,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	Score     float64                `json:"score,omitempty"`
}

// BaseMemory 基础记忆实现
type BaseMemory struct {
	storage  MemoryStorage
	crew     interface{}
	eventBus events.EventBus
	logger   logger.Logger
	embedder EmbedderConfig
}

// EmbedderConfig 嵌入器配置
type EmbedderConfig struct {
	Provider string                 `json:"provider"`
	Config   map[string]interface{} `json:"config"`
}

// MemoryStorage 记忆存储接口
type MemoryStorage interface {
	// 保存记忆项
	Save(ctx context.Context, item MemoryItem) error

	// 搜索记忆项
	Search(ctx context.Context, query string, limit int, scoreThreshold float64) ([]MemoryItem, error)

	// 删除记忆项
	Delete(ctx context.Context, id string) error

	// 清除所有记忆项
	Clear(ctx context.Context) error

	// 关闭存储
	Close() error
}

// NewBaseMemory 创建基础记忆实例
func NewBaseMemory(storage MemoryStorage, eventBus events.EventBus, logger logger.Logger) *BaseMemory {
	return &BaseMemory{
		storage:  storage,
		eventBus: eventBus,
		logger:   logger,
		embedder: EmbedderConfig{
			Provider: "default",
			Config:   make(map[string]interface{}),
		},
	}
}

// Save 实现Memory接口
func (m *BaseMemory) Save(ctx context.Context, value interface{}, metadata map[string]interface{}, agent string) error {
	// 发射记忆保存开始事件
	startEvent := NewMemorySaveStartedEvent(agent, value)
	m.eventBus.Emit(ctx, m, startEvent)

	// 创建记忆项
	item := MemoryItem{
		ID:        generateMemoryID(),
		Value:     value,
		Metadata:  metadata,
		Agent:     agent,
		CreatedAt: time.Now(),
	}

	// 保存到存储
	err := m.storage.Save(ctx, item)

	if err != nil {
		// 发射失败事件
		failedEvent := NewMemorySaveFailedEvent(agent, err.Error())
		m.eventBus.Emit(ctx, m, failedEvent)

		m.logger.Error("memory save failed",
			logger.Field{Key: "agent", Value: agent},
			logger.Field{Key: "error", Value: err},
		)
		return err
	}

	// 发射成功事件
	completedEvent := NewMemorySaveCompletedEvent(agent, item.ID)
	m.eventBus.Emit(ctx, m, completedEvent)

	m.logger.Debug("memory saved",
		logger.Field{Key: "agent", Value: agent},
		logger.Field{Key: "memory_id", Value: item.ID},
	)

	return nil
}

// Search 实现Memory接口
func (m *BaseMemory) Search(ctx context.Context, query string, limit int, scoreThreshold float64) ([]MemoryItem, error) {
	// 发射记忆查询开始事件
	startEvent := NewMemoryQueryStartedEvent(query, limit)
	m.eventBus.Emit(ctx, m, startEvent)

	// 从存储搜索
	results, err := m.storage.Search(ctx, query, limit, scoreThreshold)

	if err != nil {
		// 发射失败事件
		failedEvent := NewMemoryQueryFailedEvent(query, err.Error())
		m.eventBus.Emit(ctx, m, failedEvent)

		m.logger.Error("memory search failed",
			logger.Field{Key: "query", Value: query},
			logger.Field{Key: "error", Value: err},
		)
		return nil, err
	}

	// 发射成功事件
	completedEvent := NewMemoryQueryCompletedEvent(query, len(results))
	m.eventBus.Emit(ctx, m, completedEvent)

	m.logger.Debug("memory search completed",
		logger.Field{Key: "query", Value: query},
		logger.Field{Key: "results_count", Value: len(results)},
	)

	return results, nil
}

// Clear 实现Memory接口
func (m *BaseMemory) Clear(ctx context.Context) error {
	err := m.storage.Clear(ctx)
	if err != nil {
		m.logger.Error("memory clear failed", logger.Field{Key: "error", Value: err})
		return err
	}

	m.logger.Info("memory cleared successfully")
	return nil
}

// SetCrew 实现Memory接口
func (m *BaseMemory) SetCrew(crew interface{}) Memory {
	m.crew = crew
	return m
}

// Close 实现Memory接口
func (m *BaseMemory) Close() error {
	if m.storage != nil {
		return m.storage.Close()
	}
	return nil
}

// GetEventBus 获取事件总线（辅助方法）
func (m *BaseMemory) GetEventBus() events.EventBus {
	return m.eventBus
}

// GetLogger 获取日志器（辅助方法）
func (m *BaseMemory) GetLogger() logger.Logger {
	return m.logger
}

// generateMemoryID 生成记忆ID
func generateMemoryID() string {
	return time.Now().Format("20060102150405") + "_" + generateRandomString(8)
}

// generateRandomString 生成随机字符串
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
