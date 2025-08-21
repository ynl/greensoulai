package storage

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ynl/greensoulai/internal/memory"
	"github.com/ynl/greensoulai/pkg/logger"
)

// RAGStorage RAG存储实现（向量存储+检索）
type RAGStorage struct {
	storageType    string
	embedderConfig *memory.EmbedderConfig
	crew           interface{}
	path           string
	logger         logger.Logger

	// 内存存储（简化实现）
	items []memory.MemoryItem
	mu    sync.RWMutex

	// 向量存储相关
	vectorDim  int
	indexBuilt bool
}

// NewRAGStorage 创建RAG存储实例
func NewRAGStorage(storageType string, embedderConfig *memory.EmbedderConfig, crew interface{}, path string, logger logger.Logger) *RAGStorage {
	storage := &RAGStorage{
		storageType:    storageType,
		embedderConfig: embedderConfig,
		crew:           crew,
		path:           path,
		logger:         logger,
		items:          make([]memory.MemoryItem, 0),
		vectorDim:      384, // 默认向量维度
		indexBuilt:     false,
	}

	// 初始化存储
	storage.initialize()

	return storage
}

// initialize 初始化存储
func (rs *RAGStorage) initialize() error {
	rs.logger.Info("initializing RAG storage",
		logger.Field{Key: "type", Value: rs.storageType},
		logger.Field{Key: "path", Value: rs.path},
	)

	// 如果指定了路径，创建目录
	if rs.path != "" {
		dir := filepath.Dir(rs.path)
		// 这里应该创建目录，简化实现先跳过
		rs.logger.Debug("storage path configured", logger.Field{Key: "dir", Value: dir})
	}

	return nil
}

// Save 保存记忆项到存储
func (rs *RAGStorage) Save(ctx context.Context, item memory.MemoryItem) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	// 生成向量表示（简化实现，实际应使用嵌入模型）
	item.Score = 0.0 // 初始分数

	// 添加到内存存储
	rs.items = append(rs.items, item)

	rs.logger.Debug("memory item saved to RAG storage",
		logger.Field{Key: "id", Value: item.ID},
		logger.Field{Key: "type", Value: rs.storageType},
	)

	// 标记需要重建索引
	rs.indexBuilt = false

	return nil
}

// Search 搜索记忆项
func (rs *RAGStorage) Search(ctx context.Context, query string, limit int, scoreThreshold float64) ([]memory.MemoryItem, error) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	// 简化的搜索实现（实际应使用向量搜索）
	var results []memory.MemoryItem
	queryLower := strings.ToLower(query)

	for _, item := range rs.items {
		// 简单的文本匹配
		score := rs.calculateRelevanceScore(item, queryLower)

		if score >= scoreThreshold {
			item.Score = score
			results = append(results, item)
		}
	}

	// 按分数排序（简化实现）
	rs.sortByScore(results)

	// 限制结果数量
	if len(results) > limit {
		results = results[:limit]
	}

	rs.logger.Debug("RAG storage search completed",
		logger.Field{Key: "query", Value: query},
		logger.Field{Key: "results_count", Value: len(results)},
		logger.Field{Key: "type", Value: rs.storageType},
	)

	return results, nil
}

// Delete 删除记忆项
func (rs *RAGStorage) Delete(ctx context.Context, id string) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	for i, item := range rs.items {
		if item.ID == id {
			// 从切片中删除
			rs.items = append(rs.items[:i], rs.items[i+1:]...)
			rs.logger.Debug("memory item deleted from RAG storage",
				logger.Field{Key: "id", Value: id},
			)
			return nil
		}
	}

	return fmt.Errorf("memory item not found: %s", id)
}

// Clear 清除所有记忆项
func (rs *RAGStorage) Clear(ctx context.Context) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	count := len(rs.items)
	rs.items = make([]memory.MemoryItem, 0)
	rs.indexBuilt = false

	rs.logger.Info("RAG storage cleared",
		logger.Field{Key: "deleted_count", Value: count},
		logger.Field{Key: "type", Value: rs.storageType},
	)

	return nil
}

// Close 关闭存储
func (rs *RAGStorage) Close() error {
	rs.logger.Info("closing RAG storage",
		logger.Field{Key: "type", Value: rs.storageType},
	)
	return nil
}

// calculateRelevanceScore 计算相关性分数（简化实现）
func (rs *RAGStorage) calculateRelevanceScore(item memory.MemoryItem, queryLower string) float64 {
	score := 0.0

	// 检查值字段
	if item.Value != nil {
		valueStr := strings.ToLower(fmt.Sprintf("%v", item.Value))
		if strings.Contains(valueStr, queryLower) {
			score += 0.8
		}

		// 检查单词匹配
		queryWords := strings.Fields(queryLower)
		valueWords := strings.Fields(valueStr)

		matchCount := 0
		for _, qWord := range queryWords {
			for _, vWord := range valueWords {
				if qWord == vWord {
					matchCount++
					break
				}
			}
		}

		if len(queryWords) > 0 {
			score += 0.5 * float64(matchCount) / float64(len(queryWords))
		}
	}

	// 检查元数据
	if item.Metadata != nil {
		for key, value := range item.Metadata {
			keyStr := strings.ToLower(key)
			valueStr := strings.ToLower(fmt.Sprintf("%v", value))

			if strings.Contains(keyStr, queryLower) || strings.Contains(valueStr, queryLower) {
				score += 0.3
			}
		}
	}

	// 检查agent匹配
	if item.Agent != "" {
		agentLower := strings.ToLower(item.Agent)
		if strings.Contains(agentLower, queryLower) {
			score += 0.4
		}
	}

	// 时间衰减（越新的记忆分数越高）
	timeSince := time.Since(item.CreatedAt)
	if timeSince < 24*time.Hour {
		score += 0.2
	} else if timeSince < 7*24*time.Hour {
		score += 0.1
	}

	// 确保分数在0-1范围内
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// sortByScore 按分数排序（高效实现）
func (rs *RAGStorage) sortByScore(items []memory.MemoryItem) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].Score > items[j].Score
	})
}

// GetStats 获取存储统计信息
func (rs *RAGStorage) GetStats() map[string]interface{} {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	return map[string]interface{}{
		"storage_type": rs.storageType,
		"items_count":  len(rs.items),
		"vector_dim":   rs.vectorDim,
		"index_built":  rs.indexBuilt,
		"path":         rs.path,
	}
}
