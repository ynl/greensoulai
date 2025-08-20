package crew

import (
	"context"
	"fmt"
	"time"

	"github.com/ynl/greensoulai/internal/agent"
	"github.com/ynl/greensoulai/internal/memory"
	"github.com/ynl/greensoulai/internal/memory/contextual"
	"github.com/ynl/greensoulai/internal/memory/entity"
	"github.com/ynl/greensoulai/internal/memory/external"
	"github.com/ynl/greensoulai/internal/memory/long_term"
	"github.com/ynl/greensoulai/internal/memory/short_term"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// MemoryManager 记忆管理器，集成ContextualMemory和其他记忆系统
type MemoryManager struct {
	// 基础记忆系统
	shortTermMemory *short_term.ShortTermMemory
	longTermMemory  *long_term.LongTermMemory
	entityMemory    *entity.EntityMemory
	externalMemory  *external.ExternalMemory

	// 上下文记忆系统（核心）
	contextualMemory *contextual.ContextualMemory

	// 配置
	config MemoryManagerConfig

	// 基础设施
	eventBus events.EventBus
	logger   logger.Logger
	crew     interface{} // crew实例引用
}

// MemoryManagerConfig 记忆管理器配置
type MemoryManagerConfig struct {
	// 基础配置
	Enabled          bool   `json:"enabled"`
	StoragePath      string `json:"storage_path"`
	EmbedderProvider string `json:"embedder_provider"`

	// 记忆系统开关
	EnableShortTerm  bool `json:"enable_short_term"`
	EnableLongTerm   bool `json:"enable_long_term"`
	EnableEntity     bool `json:"enable_entity"`
	EnableExternal   bool `json:"enable_external"`
	EnableContextual bool `json:"enable_contextual"`

	// 嵌入器配置
	EmbedderConfig *memory.EmbedderConfig `json:"embedder_config"`

	// 上下文记忆配置
	ContextualConfig *contextual.ContextualMemoryConfig `json:"contextual_config"`
}

// DefaultMemoryManagerConfig 默认记忆管理器配置
func DefaultMemoryManagerConfig() MemoryManagerConfig {
	return MemoryManagerConfig{
		Enabled:          true,
		StoragePath:      "data",
		EmbedderProvider: "default",
		EnableShortTerm:  true,
		EnableLongTerm:   true,
		EnableEntity:     true,
		EnableExternal:   false, // 外部记忆默认关闭
		EnableContextual: true,
		EmbedderConfig: &memory.EmbedderConfig{
			Provider: "default",
			Config:   map[string]interface{}{},
		},
		ContextualConfig: nil, // 使用默认配置
	}
}

// NewMemoryManager 创建记忆管理器
func NewMemoryManager(
	crew interface{},
	config MemoryManagerConfig,
	eventBus events.EventBus,
	logger logger.Logger,
) *MemoryManager {
	mm := &MemoryManager{
		config:   config,
		eventBus: eventBus,
		logger:   logger,
		crew:     crew,
	}

	// 初始化记忆系统
	mm.initializeMemorySystems()

	return mm
}

// initializeMemorySystems 初始化各种记忆系统
func (mm *MemoryManager) initializeMemorySystems() {
	// 初始化短期记忆
	if mm.config.EnableShortTerm {
		mm.shortTermMemory = short_term.NewShortTermMemory(
			mm.crew,
			mm.config.EmbedderConfig,
			nil, // 使用默认存储
			mm.config.StoragePath,
			mm.eventBus,
			mm.logger,
		)
		mm.logger.Debug("short-term memory initialized")
	}

	// 初始化长期记忆
	if mm.config.EnableLongTerm {
		mm.longTermMemory = long_term.NewLongTermMemory(
			nil, // 使用默认存储
			mm.config.StoragePath,
			mm.eventBus,
			mm.logger,
		)
		mm.logger.Debug("long-term memory initialized")
	}

	// 初始化实体记忆
	if mm.config.EnableEntity {
		mm.entityMemory = entity.NewEntityMemory(
			mm.crew,
			mm.config.EmbedderConfig,
			nil, // 使用默认存储
			mm.config.StoragePath,
			mm.eventBus,
			mm.logger,
		)
		mm.logger.Debug("entity memory initialized")
	}

	// 初始化外部记忆
	if mm.config.EnableExternal {
		mm.externalMemory = external.NewExternalMemory(
			mm.crew,
			mm.config.EmbedderConfig,
			nil, // 使用默认存储
			mm.config.StoragePath,
			mm.eventBus,
			mm.logger,
		)
		mm.logger.Debug("external memory initialized")
	}

	// 初始化上下文记忆（核心）
	if mm.config.EnableContextual {
		mm.contextualMemory = contextual.NewContextualMemory(
			mm.shortTermMemory,
			mm.longTermMemory,
			mm.entityMemory,
			mm.externalMemory,
			mm.eventBus,
			mm.logger,
			mm.config.ContextualConfig,
		)
		mm.logger.Info("contextual memory system initialized")
	}
}

// BuildTaskContext 为任务构建上下文信息（主要接口）
func (mm *MemoryManager) BuildTaskContext(ctx context.Context, task agent.Task, additionalContext string) (string, error) {
	if !mm.config.Enabled || mm.contextualMemory == nil {
		mm.logger.Debug("memory system disabled or contextual memory not available")
		return "", nil
	}

	return mm.contextualMemory.BuildContextForTask(ctx, task, additionalContext)
}

// SaveMemory 保存记忆到指定的记忆系统
func (mm *MemoryManager) SaveMemory(ctx context.Context, memoryType string, value interface{}, metadata map[string]interface{}, agentName string) error {
	if !mm.config.Enabled {
		return nil
	}

	switch memoryType {
	case "short_term":
		if mm.shortTermMemory != nil {
			return mm.shortTermMemory.Save(ctx, value, metadata, agentName)
		}
	case "long_term":
		if mm.longTermMemory != nil {
			// 需要转换为LongTermMemoryItem
			qualityPtr := new(float64)
			*qualityPtr = 0.0

			item := &long_term.LongTermMemoryItem{
				Agent:          agentName,
				Task:           fmt.Sprintf("%v", value),
				ExpectedOutput: "",
				DateTime:       fmt.Sprintf("%d", time.Now().Unix()),
				Quality:        qualityPtr,
				Metadata:       metadata,
			}
			return mm.longTermMemory.Save(ctx, item)
		}
	case "entity":
		if mm.entityMemory != nil {
			return mm.entityMemory.Save(ctx, value, metadata, agentName)
		}
	case "external":
		if mm.externalMemory != nil {
			return mm.externalMemory.Save(ctx, value, metadata, agentName)
		}
	default:
		return fmt.Errorf("unknown memory type: %s", memoryType)
	}

	return fmt.Errorf("memory system not available for type: %s", memoryType)
}

// SearchMemory 搜索指定记忆系统
func (mm *MemoryManager) SearchMemory(ctx context.Context, memoryType string, query string, limit int, scoreThreshold float64) ([]memory.MemoryItem, error) {
	if !mm.config.Enabled {
		return nil, nil
	}

	switch memoryType {
	case "short_term":
		if mm.shortTermMemory != nil {
			return mm.shortTermMemory.Search(ctx, query, limit, scoreThreshold)
		}
	case "entity":
		if mm.entityMemory != nil {
			return mm.entityMemory.Search(ctx, query, limit, scoreThreshold)
		}
	case "external":
		if mm.externalMemory != nil {
			return mm.externalMemory.Search(ctx, query, limit, scoreThreshold)
		}
	default:
		return nil, fmt.Errorf("unknown memory type: %s", memoryType)
	}

	return nil, fmt.Errorf("memory system not available for type: %s", memoryType)
}

// ClearMemory 清理指定记忆系统
func (mm *MemoryManager) ClearMemory(ctx context.Context, memoryType string) error {
	if !mm.config.Enabled {
		return nil
	}

	switch memoryType {
	case "short_term":
		if mm.shortTermMemory != nil {
			return mm.shortTermMemory.Clear(ctx)
		}
	case "entity":
		if mm.entityMemory != nil {
			return mm.entityMemory.Clear(ctx)
		}
	case "external":
		if mm.externalMemory != nil {
			return mm.externalMemory.Clear(ctx)
		}
	case "all":
		// 清理所有记忆系统
		var errors []string
		if mm.shortTermMemory != nil {
			if err := mm.shortTermMemory.Clear(ctx); err != nil {
				errors = append(errors, fmt.Sprintf("short_term: %v", err))
			}
		}
		if mm.entityMemory != nil {
			if err := mm.entityMemory.Clear(ctx); err != nil {
				errors = append(errors, fmt.Sprintf("entity: %v", err))
			}
		}
		if mm.externalMemory != nil {
			if err := mm.externalMemory.Clear(ctx); err != nil {
				errors = append(errors, fmt.Sprintf("external: %v", err))
			}
		}
		if len(errors) > 0 {
			return fmt.Errorf("errors clearing memories: %v", errors)
		}
		return nil
	default:
		return fmt.Errorf("unknown memory type: %s", memoryType)
	}

	return fmt.Errorf("memory system not available for type: %s", memoryType)
}

// GetContextualMemory 获取上下文记忆实例（用于高级使用）
func (mm *MemoryManager) GetContextualMemory() *contextual.ContextualMemory {
	return mm.contextualMemory
}

// GetMemoryInstances 获取所有记忆实例（用于调试和测试）
func (mm *MemoryManager) GetMemoryInstances() (
	*short_term.ShortTermMemory,
	*long_term.LongTermMemory,
	*entity.EntityMemory,
	*external.ExternalMemory,
) {
	return mm.shortTermMemory, mm.longTermMemory, mm.entityMemory, mm.externalMemory
}

// IsEnabled 检查记忆系统是否启用
func (mm *MemoryManager) IsEnabled() bool {
	return mm.config.Enabled
}

// GetConfig 获取配置
func (mm *MemoryManager) GetConfig() MemoryManagerConfig {
	return mm.config
}

// UpdateContextualConfig 更新上下文记忆配置
func (mm *MemoryManager) UpdateContextualConfig(config contextual.ContextualMemoryConfig) error {
	if mm.contextualMemory == nil {
		return fmt.Errorf("contextual memory not initialized")
	}

	mm.contextualMemory.UpdateConfig(config)
	mm.config.ContextualConfig = &config
	mm.logger.Info("contextual memory config updated")

	return nil
}

// Close 关闭记忆管理器
func (mm *MemoryManager) Close() error {
	var errors []string

	if mm.shortTermMemory != nil {
		if err := mm.shortTermMemory.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("short_term: %v", err))
		}
	}

	if mm.longTermMemory != nil {
		// long_term.LongTermMemory 可能没有Close方法，需要检查
		// if err := mm.longTermMemory.Close(); err != nil {
		//     errors = append(errors, fmt.Sprintf("long_term: %v", err))
		// }
	}

	if mm.entityMemory != nil {
		if err := mm.entityMemory.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("entity: %v", err))
		}
	}

	if mm.externalMemory != nil {
		if err := mm.externalMemory.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("external: %v", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing memory systems: %v", errors)
	}

	mm.logger.Info("memory manager closed successfully")
	return nil
}
