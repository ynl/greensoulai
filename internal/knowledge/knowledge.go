package knowledge

import (
	"context"
	"fmt"
	"time"

	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// Knowledge 知识系统主要接口
// 用于管理知识源的集合和向量存储的配置
type Knowledge interface {
	// 添加知识源
	AddSource(source BaseKnowledgeSource) error

	// 移除知识源
	RemoveSource(sourceName string) error

	// 获取所有知识源
	GetSources() []BaseKnowledgeSource

	// 查询知识
	Query(ctx context.Context, query []string, resultsLimit int, scoreThreshold float64) ([]KnowledgeResult, error)

	// 添加所有源到存储
	AddSources() error

	// 重置知识存储
	Reset() error

	// 关闭知识系统
	Close() error
}

// KnowledgeResult 知识查询结果
type KnowledgeResult struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Source    string                 `json:"source"`
	Score     float64                `json:"score"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}

// KnowledgeStorage 知识存储接口
type KnowledgeStorage interface {
	// 初始化知识存储
	InitializeKnowledgeStorage() error

	// 保存文档
	Save(documents []string, metadata ...interface{}) error

	// 搜索文档
	Search(query []string, limit int, scoreThreshold float64) ([]KnowledgeResult, error)

	// 重置存储
	Reset() error

	// 关闭存储
	Close() error
}

// BaseKnowledgeSource 知识源基础接口
type BaseKnowledgeSource interface {
	// 获取源名称
	GetName() string

	// 获取源类型
	GetType() string

	// 验证内容
	ValidateContent() error

	// 添加到知识存储
	Add() error

	// 获取嵌入向量
	GetEmbeddings() ([][]float64, error)

	// 设置存储
	SetStorage(storage KnowledgeStorage)

	// 获取块
	GetChunks() []string

	// 获取元数据
	GetMetadata() map[string]interface{}
}

// KnowledgeImpl 知识系统实现
type KnowledgeImpl struct {
	collectionName string
	sources        []BaseKnowledgeSource
	storage        KnowledgeStorage
	embedder       EmbedderConfig
	eventBus       events.EventBus
	logger         logger.Logger
}

// EmbedderConfig 嵌入器配置
type EmbedderConfig struct {
	Provider string                 `json:"provider"`
	Config   map[string]interface{} `json:"config"`
}

// NewKnowledge 创建知识系统实例
func NewKnowledge(
	collectionName string,
	sources []BaseKnowledgeSource,
	embedder *EmbedderConfig,
	storage KnowledgeStorage,
	eventBus events.EventBus,
	log logger.Logger,
) *KnowledgeImpl {
	var knowledgeStorage KnowledgeStorage

	if storage != nil {
		knowledgeStorage = storage
	} else {
		// TODO: 创建默认存储实现
		// 由于循环依赖问题，需要通过工厂模式或依赖注入来解决
		log.Warn("no knowledge storage provided, some features may not work")
		// 临时使用nil，实际使用时需要提供具体存储实现
		knowledgeStorage = nil
	}

	k := &KnowledgeImpl{
		collectionName: collectionName,
		sources:        sources,
		storage:        knowledgeStorage,
		eventBus:       eventBus,
		logger:         log,
	}

	if embedder != nil {
		k.embedder = *embedder
	} else {
		// 默认嵌入器配置
		k.embedder = EmbedderConfig{
			Provider: "default",
			Config:   make(map[string]interface{}),
		}
	}

	// 初始化存储
	if k.storage != nil {
		if err := k.storage.InitializeKnowledgeStorage(); err != nil {
			log.Error("failed to initialize knowledge storage",
				logger.Field{Key: "error", Value: err},
				logger.Field{Key: "collection", Value: collectionName},
			)
		}
	}

	return k
}

// AddSource 添加知识源
func (k *KnowledgeImpl) AddSource(source BaseKnowledgeSource) error {
	// 检查源是否已存在
	for _, existingSource := range k.sources {
		if existingSource.GetName() == source.GetName() {
			return fmt.Errorf("knowledge source already exists: %s", source.GetName())
		}
	}

	// 设置存储引用
	source.SetStorage(k.storage)

	// 添加到源列表
	k.sources = append(k.sources, source)

	k.logger.Info("knowledge source added",
		logger.Field{Key: "source_name", Value: source.GetName()},
		logger.Field{Key: "source_type", Value: source.GetType()},
		logger.Field{Key: "collection", Value: k.collectionName},
	)

	return nil
}

// RemoveSource 移除知识源
func (k *KnowledgeImpl) RemoveSource(sourceName string) error {
	for i, source := range k.sources {
		if source.GetName() == sourceName {
			// 从切片中移除
			k.sources = append(k.sources[:i], k.sources[i+1:]...)

			k.logger.Info("knowledge source removed",
				logger.Field{Key: "source_name", Value: sourceName},
				logger.Field{Key: "collection", Value: k.collectionName},
			)

			return nil
		}
	}

	return fmt.Errorf("knowledge source not found: %s", sourceName)
}

// GetSources 获取所有知识源
func (k *KnowledgeImpl) GetSources() []BaseKnowledgeSource {
	return k.sources
}

// Query 查询知识
func (k *KnowledgeImpl) Query(ctx context.Context, query []string, resultsLimit int, scoreThreshold float64) ([]KnowledgeResult, error) {
	if k.storage == nil {
		return nil, fmt.Errorf("storage is not initialized")
	}

	// 发射查询开始事件
	startEvent := NewKnowledgeQueryStartedEvent(k.collectionName, query, resultsLimit)
	k.eventBus.Emit(ctx, k, startEvent)

	// 执行搜索
	results, err := k.storage.Search(query, resultsLimit, scoreThreshold)
	if err != nil {
		// 发射失败事件
		failedEvent := NewKnowledgeQueryFailedEvent(k.collectionName, query, err.Error())
		k.eventBus.Emit(ctx, k, failedEvent)

		k.logger.Error("knowledge query failed",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "query", Value: query},
			logger.Field{Key: "collection", Value: k.collectionName},
		)

		return nil, err
	}

	// 发射成功事件
	completedEvent := NewKnowledgeQueryCompletedEvent(k.collectionName, query, len(results))
	k.eventBus.Emit(ctx, k, completedEvent)

	k.logger.Debug("knowledge query completed",
		logger.Field{Key: "query", Value: query},
		logger.Field{Key: "results_count", Value: len(results)},
		logger.Field{Key: "collection", Value: k.collectionName},
	)

	return results, nil
}

// AddSources 添加所有源到存储
func (k *KnowledgeImpl) AddSources() error {
	k.logger.Info("adding all knowledge sources to storage",
		logger.Field{Key: "sources_count", Value: len(k.sources)},
		logger.Field{Key: "collection", Value: k.collectionName},
	)

	var errors []string

	for _, source := range k.sources {
		k.logger.Debug("processing knowledge source",
			logger.Field{Key: "source_name", Value: source.GetName()},
			logger.Field{Key: "source_type", Value: source.GetType()},
		)

		// 验证内容
		if err := source.ValidateContent(); err != nil {
			errorMsg := fmt.Sprintf("validation failed for source %s: %v", source.GetName(), err)
			errors = append(errors, errorMsg)
			k.logger.Error("knowledge source validation failed",
				logger.Field{Key: "source_name", Value: source.GetName()},
				logger.Field{Key: "error", Value: err},
			)
			continue
		}

		// 添加到存储
		if err := source.Add(); err != nil {
			errorMsg := fmt.Sprintf("failed to add source %s: %v", source.GetName(), err)
			errors = append(errors, errorMsg)
			k.logger.Error("failed to add knowledge source to storage",
				logger.Field{Key: "source_name", Value: source.GetName()},
				logger.Field{Key: "error", Value: err},
			)
			continue
		}

		k.logger.Info("knowledge source added to storage",
			logger.Field{Key: "source_name", Value: source.GetName()},
			logger.Field{Key: "chunks_count", Value: len(source.GetChunks())},
		)
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors occurred while adding sources: %v", errors)
	}

	k.logger.Info("all knowledge sources added successfully",
		logger.Field{Key: "collection", Value: k.collectionName},
	)

	return nil
}

// Reset 重置知识存储
func (k *KnowledgeImpl) Reset() error {
	if k.storage == nil {
		return fmt.Errorf("storage is not initialized")
	}

	err := k.storage.Reset()
	if err != nil {
		k.logger.Error("failed to reset knowledge storage",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "collection", Value: k.collectionName},
		)
		return err
	}

	k.logger.Info("knowledge storage reset",
		logger.Field{Key: "collection", Value: k.collectionName},
	)

	return nil
}

// Close 关闭知识系统
func (k *KnowledgeImpl) Close() error {
	k.logger.Info("closing knowledge system",
		logger.Field{Key: "collection", Value: k.collectionName},
	)

	if k.storage != nil {
		return k.storage.Close()
	}

	return nil
}

// GetStats 获取知识系统统计信息
func (k *KnowledgeImpl) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"collection_name": k.collectionName,
		"sources_count":   len(k.sources),
		"embedder":        k.embedder,
	}

	// 统计各类型源的数量
	sourceTypes := make(map[string]int)
	totalChunks := 0

	for _, source := range k.sources {
		sourceType := source.GetType()
		sourceTypes[sourceType]++
		totalChunks += len(source.GetChunks())
	}

	stats["source_types"] = sourceTypes
	stats["total_chunks"] = totalChunks

	return stats
}
