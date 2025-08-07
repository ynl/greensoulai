package long_term

import (
	"context"
	"time"

	"github.com/ynl/greensoulai/internal/memory"
	"github.com/ynl/greensoulai/internal/memory/storage"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// LongTermMemory 长期记忆实现
// 遵循Python版本的设计：管理跨运行的数据，主要用于crew的整体执行和性能数据
// 与Python版本保持业务逻辑一致
type LongTermMemory struct {
	*memory.BaseMemory
	sqliteStorage *storage.LTMSQLiteStorage
}

// NewLongTermMemory 创建长期记忆实例，遵循Python版本的构造逻辑
// Python版本: def __init__(self, storage=None, path=None)
func NewLongTermMemory(memStorage memory.MemoryStorage, path string, eventBus events.EventBus, logger logger.Logger) *LongTermMemory {
	var storageInstance memory.MemoryStorage
	var sqliteStorage *storage.LTMSQLiteStorage

	if memStorage != nil {
		storageInstance = memStorage
	} else {
		// 创建SQLite存储实例，与Python版本默认行为一致
		if path != "" {
			sqliteStorage = storage.NewLTMSQLiteStorage(path, logger)
		} else {
			sqliteStorage = storage.NewLTMSQLiteStorage("", logger) // 使用默认路径
		}
		storageInstance = sqliteStorage
	}

	baseMemory := memory.NewBaseMemory(storageInstance, eventBus, logger)

	return &LongTermMemory{
		BaseMemory:    baseMemory,
		sqliteStorage: sqliteStorage,
	}
}

// Save 保存长期记忆项，遵循Python版本的接口设计
// Python版本: def save(self, item: LongTermMemoryItem) -> None
func (ltm *LongTermMemory) Save(ctx context.Context, item *LongTermMemoryItem) error {
	// 发射记忆保存开始事件（与Python版本的事件系统保持一致）
	if ltm.GetEventBus() != nil {
		startEvent := memory.NewMemorySaveStartedEvent(item.Task, "long_term_memory")
		ltm.GetEventBus().Emit(ctx, ltm, startEvent)
	}

	// 转换为通用MemoryItem格式以便存储
	memoryItem := memory.MemoryItem{
		ID:        item.ID,
		Value:     item.ToDict(), // 存储为字典格式，与Python版本一致
		Metadata:  make(map[string]interface{}),
		Agent:     item.Agent,
		CreatedAt: item.CreatedAt,
		Score:     item.Score,
	}

	// 复制原始metadata
	for k, v := range item.Metadata {
		memoryItem.Metadata[k] = v
	}

	// 添加长期记忆特有的元数据
	memoryItem.Metadata["memory_type"] = "long_term"
	memoryItem.Metadata["task"] = item.Task
	memoryItem.Metadata["expected_output"] = item.ExpectedOutput
	memoryItem.Metadata["datetime"] = item.DateTime

	if item.Quality != nil {
		memoryItem.Metadata["quality"] = *item.Quality
	}

	// 通过存储层保存
	var err error
	if ltm.sqliteStorage != nil {
		err = ltm.sqliteStorage.Save(ctx, memoryItem)
	} else {
		// 回退到BaseMemory的Save方法
		err = ltm.BaseMemory.Save(ctx, memoryItem.Value, memoryItem.Metadata, memoryItem.Agent)
	}

	if err != nil {
		// 发射失败事件
		if ltm.GetEventBus() != nil {
			failedEvent := memory.NewMemorySaveFailedEvent(item.Task, err.Error())
			ltm.GetEventBus().Emit(ctx, ltm, failedEvent)
		}
		return err
	}

	// 发射成功事件
	if ltm.GetEventBus() != nil {
		completedEvent := memory.NewMemorySaveCompletedEvent(item.Task, "long_term_memory")
		ltm.GetEventBus().Emit(ctx, ltm, completedEvent)
	}

	return nil
}

// Search 搜索长期记忆，遵循Python版本的接口设计
// Python版本: def search(self, task: str, latest_n: int = 3) -> List[Dict[str, Any]]
func (ltm *LongTermMemory) Search(ctx context.Context, task string, latestN int) ([]map[string]interface{}, error) {
	// 发射查询开始事件
	if ltm.GetEventBus() != nil {
		startEvent := memory.NewMemoryQueryStartedEvent(task, latestN)
		ltm.GetEventBus().Emit(ctx, ltm, startEvent)
	}

	var results []map[string]interface{}
	var err error

	if ltm.sqliteStorage != nil {
		// 使用SQLite存储的搜索功能
		memoryItems, searchErr := ltm.sqliteStorage.Search(ctx, task, latestN, 0.0)
		err = searchErr

		if err == nil {
			// 转换为字典列表格式，与Python版本返回格式一致
			results = make([]map[string]interface{}, len(memoryItems))
			for i, item := range memoryItems {
				if dictValue, ok := item.Value.(map[string]interface{}); ok {
					results[i] = dictValue
				} else {
					// 如果不是字典格式，构造一个基本格式
					results[i] = map[string]interface{}{
						"agent":           item.Agent,
						"task":            task,
						"expected_output": "",
						"datetime":        item.CreatedAt.Format(time.RFC3339),
						"quality":         nil,
						"metadata":        item.Metadata,
					}

					// 尝试从metadata中恢复字段
					if taskVal, exists := item.Metadata["task"]; exists {
						results[i]["task"] = taskVal
					}
					if expectedOutput, exists := item.Metadata["expected_output"]; exists {
						results[i]["expected_output"] = expectedOutput
					}
					if datetime, exists := item.Metadata["datetime"]; exists {
						results[i]["datetime"] = datetime
					}
					if quality, exists := item.Metadata["quality"]; exists {
						results[i]["quality"] = quality
					}
				}
			}
		}
	} else {
		// 回退到BaseMemory的搜索
		memoryItems, searchErr := ltm.BaseMemory.Search(ctx, task, latestN, 0.0)
		err = searchErr

		if err == nil {
			results = make([]map[string]interface{}, len(memoryItems))
			for i, item := range memoryItems {
				if dictValue, ok := item.Value.(map[string]interface{}); ok {
					results[i] = dictValue
				} else {
					results[i] = map[string]interface{}{
						"agent":           item.Agent,
						"task":            task,
						"expected_output": "",
						"datetime":        item.CreatedAt.Format(time.RFC3339),
						"quality":         nil,
						"metadata":        item.Metadata,
					}
				}
			}
		}
	}

	if err != nil {
		// 发射失败事件
		if ltm.GetEventBus() != nil {
			failedEvent := memory.NewMemoryQueryFailedEvent(task, err.Error())
			ltm.GetEventBus().Emit(ctx, ltm, failedEvent)
		}
		return nil, err
	}

	// 发射成功事件
	if ltm.GetEventBus() != nil {
		completedEvent := memory.NewMemoryQueryCompletedEvent(task, len(results))
		ltm.GetEventBus().Emit(ctx, ltm, completedEvent)
	}

	return results, nil
}

// SaveCompatible 兼容通用Memory接口的Save方法
func (ltm *LongTermMemory) SaveCompatible(ctx context.Context, value interface{}, metadata map[string]interface{}, agent string) error {
	return ltm.BaseMemory.Save(ctx, value, metadata, agent)
}

// SearchCompatible 兼容通用Memory接口的Search方法
func (ltm *LongTermMemory) SearchCompatible(ctx context.Context, query string, limit int, scoreThreshold float64) ([]memory.MemoryItem, error) {
	return ltm.BaseMemory.Search(ctx, query, limit, scoreThreshold)
}

// Reset 重置长期记忆，遵循Python版本的接口设计
// Python版本: def reset(self) -> None
func (ltm *LongTermMemory) Reset(ctx context.Context) error {
	if ltm.sqliteStorage != nil {
		return ltm.sqliteStorage.Clear(ctx)
	}
	return ltm.BaseMemory.Clear(ctx)
}

// GetEventBus 获取事件总线的辅助方法
func (ltm *LongTermMemory) GetEventBus() events.EventBus {
	// 由于BaseMemory的eventBus可能是私有的，我们通过这种方式访问
	// 这是一个临时解决方案，理想情况下BaseMemory应该提供公共访问方法
	if ltm.BaseMemory != nil {
		return ltm.BaseMemory.GetEventBus() // 假设BaseMemory有这个方法
	}
	return nil
}
