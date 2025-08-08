package short_term

import (
	"context"
	"fmt"

	"github.com/ynl/greensoulai/internal/memory"
	"github.com/ynl/greensoulai/internal/memory/storage"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// ShortTermMemory 短期记忆实现
// 用于管理与即时任务和交互相关的临时数据
type ShortTermMemory struct {
	*memory.BaseMemory
	memoryProvider string
}

// ShortTermMemoryItem 短期记忆项
type ShortTermMemoryItem struct {
	memory.MemoryItem
	SessionID string `json:"session_id,omitempty"`
	TaskID    string `json:"task_id,omitempty"`
}

// NewShortTermMemory 创建短期记忆实例
func NewShortTermMemory(crew interface{}, embedderConfig *memory.EmbedderConfig, memStorage memory.MemoryStorage, path string, eventBus events.EventBus, logger logger.Logger) *ShortTermMemory {
	var memoryProvider string
	var storageInstance memory.MemoryStorage

	// 根据配置选择存储提供者
	if embedderConfig != nil {
		memoryProvider = embedderConfig.Provider
	}

	if memoryProvider == "mem0" {
		// 如果使用mem0存储
		logger.Info("using mem0 storage for short term memory")
		storageInstance = storage.NewMem0Storage("short_term", crew, embedderConfig.Config, logger)
	} else {
		// 默认使用RAG存储
		if memStorage != nil {
			storageInstance = memStorage
		} else {
			logger.Info("using default RAG storage for short term memory")
			storageInstance = storage.NewRAGStorage("short_term", embedderConfig, crew, path, logger)
		}
	}

	baseMemory := memory.NewBaseMemory(storageInstance, eventBus, logger)

	return &ShortTermMemory{
		BaseMemory:     baseMemory,
		memoryProvider: memoryProvider,
	}
}

// Save 保存短期记忆项
func (stm *ShortTermMemory) Save(ctx context.Context, value interface{}, metadata map[string]interface{}, agent string) error {
	// 为短期记忆添加特定的元数据
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	metadata["memory_type"] = "short_term"
	metadata["provider"] = stm.memoryProvider

	// 如果有会话ID，添加到元数据
	if sessionID := ctx.Value("session_id"); sessionID != nil {
		metadata["session_id"] = sessionID
	}

	// 如果有任务ID，添加到元数据
	if taskID := ctx.Value("task_id"); taskID != nil {
		metadata["task_id"] = taskID
	}

	return stm.BaseMemory.Save(ctx, value, metadata, agent)
}

// SaveTaskMemory 保存任务相关的短期记忆
func (stm *ShortTermMemory) SaveTaskMemory(ctx context.Context, taskID string, value interface{}, metadata map[string]interface{}, agent string) error {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	metadata["task_id"] = taskID
	metadata["memory_type"] = "task_memory"

	return stm.Save(ctx, value, metadata, agent)
}

// SaveInteractionMemory 保存交互记忆
func (stm *ShortTermMemory) SaveInteractionMemory(ctx context.Context, sessionID string, interaction interface{}, metadata map[string]interface{}, agent string) error {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	metadata["session_id"] = sessionID
	metadata["memory_type"] = "interaction_memory"

	return stm.Save(ctx, interaction, metadata, agent)
}

// SearchByTask 根据任务ID搜索记忆
func (stm *ShortTermMemory) SearchByTask(ctx context.Context, taskID string, query string, limit int, scoreThreshold float64) ([]memory.MemoryItem, error) {
	// 首先执行基础搜索
	results, err := stm.BaseMemory.Search(ctx, query, limit*2, scoreThreshold) // 获取更多结果进行过滤
	if err != nil {
		return nil, err
	}

	// 过滤出与指定任务相关的记忆
	var filteredResults []memory.MemoryItem
	for _, item := range results {
		if item.Metadata != nil {
			if itemTaskID, ok := item.Metadata["task_id"].(string); ok && itemTaskID == taskID {
				filteredResults = append(filteredResults, item)
				if len(filteredResults) >= limit {
					break
				}
			}
		}
	}

	return filteredResults, nil
}

// SearchBySession 根据会话ID搜索记忆
func (stm *ShortTermMemory) SearchBySession(ctx context.Context, sessionID string, query string, limit int, scoreThreshold float64) ([]memory.MemoryItem, error) {
	// 首先执行基础搜索
	results, err := stm.BaseMemory.Search(ctx, query, limit*2, scoreThreshold)
	if err != nil {
		return nil, err
	}

	// 过滤出与指定会话相关的记忆
	var filteredResults []memory.MemoryItem
	for _, item := range results {
		if item.Metadata != nil {
			if itemSessionID, ok := item.Metadata["session_id"].(string); ok && itemSessionID == sessionID {
				filteredResults = append(filteredResults, item)
				if len(filteredResults) >= limit {
					break
				}
			}
		}
	}

	return filteredResults, nil
}

// GetRecentMemories 获取最近的记忆项
func (stm *ShortTermMemory) GetRecentMemories(ctx context.Context, agent string, limit int) ([]memory.MemoryItem, error) {
	// 使用通用查询获取记忆项
	query := fmt.Sprintf("agent:%s", agent)
	return stm.BaseMemory.Search(ctx, query, limit, 0.1) // 使用较低的阈值获取更多结果
}

// ClearSession 清除指定会话的记忆
func (stm *ShortTermMemory) ClearSession(ctx context.Context, sessionID string) error {
	// TODO: 实现根据会话ID删除记忆的功能
	// 这需要存储层支持按元数据删除
	return fmt.Errorf("session-specific clear not implemented yet")
}

// ClearTask 清除指定任务的记忆
func (stm *ShortTermMemory) ClearTask(ctx context.Context, taskID string) error {
	// TODO: 实现根据任务ID删除记忆的功能
	// 这需要存储层支持按元数据删除
	return fmt.Errorf("task-specific clear not implemented yet")
}

