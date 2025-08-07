package external

import (
	"context"
	"fmt"
	"time"

	"github.com/ynl/greensoulai/internal/memory"
	"github.com/ynl/greensoulai/internal/memory/storage"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// ExternalMemory 外部记忆实现
// 用于管理来自外部源的记忆数据
type ExternalMemory struct {
	*memory.BaseMemory
	sources      []ExternalSource
	syncInterval time.Duration
	lastSyncTime time.Time
	autoSync     bool
}

// ExternalSource 外部源接口
type ExternalSource interface {
	// 获取源名称
	GetName() string

	// 获取源类型
	GetType() string

	// 连接到外部源
	Connect(ctx context.Context) error

	// 从外部源获取数据
	Fetch(ctx context.Context, query string, limit int) ([]ExternalMemoryItem, error)

	// 同步数据到本地记忆
	Sync(ctx context.Context) error

	// 断开连接
	Disconnect() error

	// 检查源是否可用
	IsAvailable() bool
}

// ExternalMemoryItem 外部记忆项
type ExternalMemoryItem struct {
	memory.MemoryItem
	SourceName   string    `json:"source_name"`
	SourceType   string    `json:"source_type"`
	SourceID     string    `json:"source_id"`
	SyncTime     time.Time `json:"sync_time"`
	LastModified time.Time `json:"last_modified"`
	Version      string    `json:"version,omitempty"`
}

// ExternalSourceType 外部源类型
type ExternalSourceType string

const (
	SourceTypeDatabase   ExternalSourceType = "database"
	SourceTypeAPI        ExternalSourceType = "api"
	SourceTypeFile       ExternalSourceType = "file"
	SourceTypeWebservice ExternalSourceType = "webservice"
	SourceTypeCloud      ExternalSourceType = "cloud"
	SourceTypeKnowledge  ExternalSourceType = "knowledge"
)

// NewExternalMemory 创建外部记忆实例
func NewExternalMemory(crew interface{}, embedderConfig *memory.EmbedderConfig, memStorage memory.MemoryStorage, path string, eventBus events.EventBus, logger logger.Logger) *ExternalMemory {
	var storageInstance memory.MemoryStorage

	if memStorage != nil {
		storageInstance = memStorage
	} else {
		// 使用RAG存储
		storageInstance = storage.NewRAGStorage("external", embedderConfig, crew, path, logger)
	}

	baseMemory := memory.NewBaseMemory(storageInstance, eventBus, logger)

	return &ExternalMemory{
		BaseMemory:   baseMemory,
		sources:      make([]ExternalSource, 0),
		syncInterval: 1 * time.Hour, // 默认每小时同步一次
		autoSync:     false,
	}
}

// Save 保存外部记忆项
func (em *ExternalMemory) Save(ctx context.Context, value interface{}, metadata map[string]interface{}, agent string) error {
	// 为外部记忆添加特定的元数据
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	metadata["memory_type"] = "external"
	metadata["sync_time"] = time.Now().Format(time.RFC3339)

	return em.BaseMemory.Save(ctx, value, metadata, agent)
}

// AddSource 添加外部源
func (em *ExternalMemory) AddSource(source ExternalSource) error {
	// 检查源是否已存在
	for _, existingSource := range em.sources {
		if existingSource.GetName() == source.GetName() {
			return fmt.Errorf("source already exists: %s", source.GetName())
		}
	}

	em.sources = append(em.sources, source)
	return nil
}

// RemoveSource 移除外部源
func (em *ExternalMemory) RemoveSource(sourceName string) error {
	for i, source := range em.sources {
		if source.GetName() == sourceName {
			// 断开连接
			source.Disconnect()

			// 从切片中移除
			em.sources = append(em.sources[:i], em.sources[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("source not found: %s", sourceName)
}

// GetSources 获取所有外部源
func (em *ExternalMemory) GetSources() []ExternalSource {
	return em.sources
}

// SyncAll 同步所有外部源
func (em *ExternalMemory) SyncAll(ctx context.Context) error {
	var errors []string

	for _, source := range em.sources {
		if !source.IsAvailable() {
			continue
		}

		err := source.Sync(ctx)
		if err != nil {
			errors = append(errors, fmt.Sprintf("sync failed for %s: %v", source.GetName(), err))
		}
	}

	em.lastSyncTime = time.Now()

	if len(errors) > 0 {
		return fmt.Errorf("sync errors: %v", errors)
	}

	return nil
}

// SyncSource 同步指定外部源
func (em *ExternalMemory) SyncSource(ctx context.Context, sourceName string) error {
	for _, source := range em.sources {
		if source.GetName() == sourceName {
			if !source.IsAvailable() {
				return fmt.Errorf("source not available: %s", sourceName)
			}

			return source.Sync(ctx)
		}
	}
	return fmt.Errorf("source not found: %s", sourceName)
}

// FetchFromSource 从指定源获取数据
func (em *ExternalMemory) FetchFromSource(ctx context.Context, sourceName, query string, limit int) ([]ExternalMemoryItem, error) {
	for _, source := range em.sources {
		if source.GetName() == sourceName {
			if !source.IsAvailable() {
				return nil, fmt.Errorf("source not available: %s", sourceName)
			}

			return source.Fetch(ctx, query, limit)
		}
	}
	return nil, fmt.Errorf("source not found: %s", sourceName)
}

// SearchBySource 根据源搜索记忆
func (em *ExternalMemory) SearchBySource(ctx context.Context, sourceName, query string, limit int, scoreThreshold float64) ([]memory.MemoryItem, error) {
	// 执行基础搜索
	results, err := em.BaseMemory.Search(ctx, query, limit*2, scoreThreshold)
	if err != nil {
		return nil, err
	}

	// 过滤出指定源的记忆
	var filteredResults []memory.MemoryItem
	for _, item := range results {
		if item.Metadata != nil {
			if itemSourceName, ok := item.Metadata["source_name"].(string); ok && itemSourceName == sourceName {
				filteredResults = append(filteredResults, item)
				if len(filteredResults) >= limit {
					break
				}
			}
		}
	}

	return filteredResults, nil
}

// SearchBySourceType 根据源类型搜索记忆
func (em *ExternalMemory) SearchBySourceType(ctx context.Context, sourceType ExternalSourceType, query string, limit int, scoreThreshold float64) ([]memory.MemoryItem, error) {
	// 执行基础搜索
	results, err := em.BaseMemory.Search(ctx, query, limit*2, scoreThreshold)
	if err != nil {
		return nil, err
	}

	// 过滤出指定源类型的记忆
	var filteredResults []memory.MemoryItem
	for _, item := range results {
		if item.Metadata != nil {
			if itemSourceType, ok := item.Metadata["source_type"].(string); ok && itemSourceType == string(sourceType) {
				filteredResults = append(filteredResults, item)
				if len(filteredResults) >= limit {
					break
				}
			}
		}
	}

	return filteredResults, nil
}

// EnableAutoSync 启用自动同步
func (em *ExternalMemory) EnableAutoSync(ctx context.Context, interval time.Duration) {
	em.autoSync = true
	em.syncInterval = interval

	// 启动自动同步协程
	go em.autoSyncLoop(ctx)
}

// DisableAutoSync 禁用自动同步
func (em *ExternalMemory) DisableAutoSync() {
	em.autoSync = false
}

// autoSyncLoop 自动同步循环
func (em *ExternalMemory) autoSyncLoop(ctx context.Context) {
	ticker := time.NewTicker(em.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if em.autoSync {
				err := em.SyncAll(ctx)
				if err != nil {
					// 记录错误但继续运行
					// em.logger.Error("auto sync failed", logger.Field{Key: "error", Value: err})
				}
			}
		}
	}
}

// GetSyncStatus 获取同步状态
func (em *ExternalMemory) GetSyncStatus() map[string]interface{} {
	status := map[string]interface{}{
		"auto_sync":         em.autoSync,
		"sync_interval":     em.syncInterval.String(),
		"last_sync_time":    em.lastSyncTime.Format(time.RFC3339),
		"sources_count":     len(em.sources),
		"available_sources": 0,
	}

	availableCount := 0
	sourceStatus := make(map[string]bool)
	for _, source := range em.sources {
		isAvailable := source.IsAvailable()
		sourceStatus[source.GetName()] = isAvailable
		if isAvailable {
			availableCount++
		}
	}

	status["available_sources"] = availableCount
	status["source_status"] = sourceStatus

	return status
}

// Close 关闭外部记忆
func (em *ExternalMemory) Close() error {
	// 禁用自动同步
	em.DisableAutoSync()

	// 断开所有外部源连接
	for _, source := range em.sources {
		source.Disconnect()
	}

	// 关闭基础记忆
	return em.BaseMemory.Close()
}
