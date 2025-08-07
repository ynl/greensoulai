package storage

import (
	"fmt"
	"sync"
	"time"

	"github.com/ynl/greensoulai/pkg/logger"
)

// EmbedderConfig 嵌入器配置（局部定义以避免循环导入）
type EmbedderConfig struct {
	Provider string                 `json:"provider"`
	Config   map[string]interface{} `json:"config"`
}

// KnowledgeResult 知识查询结果（局部定义以避免循环导入）
type KnowledgeResult struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Source    string                 `json:"source"`
	Score     float64                `json:"score"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}

// KnowledgeStorageImpl 知识存储的默认实现
type KnowledgeStorageImpl struct {
	embedder       *EmbedderConfig
	collectionName string
	logger         logger.Logger

	// 简化实现：内存存储
	documents []KnowledgeDocument
	mu        sync.RWMutex

	// 向量存储配置
	vectorDim   int
	initialized bool
}

// KnowledgeDocument 知识文档
type KnowledgeDocument struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Embedding []float64              `json:"embedding,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
	Source    string                 `json:"source"`
	CreatedAt time.Time              `json:"created_at"`
}

// NewKnowledgeStorage 创建知识存储实例
func NewKnowledgeStorage(embedder *EmbedderConfig, collectionName string, logger logger.Logger) *KnowledgeStorageImpl {
	return &KnowledgeStorageImpl{
		embedder:       embedder,
		collectionName: collectionName,
		logger:         logger,
		documents:      make([]KnowledgeDocument, 0),
		vectorDim:      384, // 默认向量维度
		initialized:    false,
	}
}

// InitializeKnowledgeStorage 初始化知识存储
func (ks *KnowledgeStorageImpl) InitializeKnowledgeStorage() error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	if ks.initialized {
		return nil
	}

	ks.logger.Info("initializing knowledge storage",
		logger.Field{Key: "collection", Value: ks.collectionName},
		logger.Field{Key: "vector_dim", Value: ks.vectorDim},
	)

	// 初始化向量数据库连接
	// TODO: 集成实际的向量数据库（如Chroma、Pinecone、Weaviate等）
	if ks.embedder != nil && ks.embedder.Provider != "default" {
		ks.logger.Info("using custom embedder",
			logger.Field{Key: "provider", Value: ks.embedder.Provider},
		)
	} else {
		ks.logger.Info("using default embedder")
	}

	ks.initialized = true

	ks.logger.Info("knowledge storage initialized successfully",
		logger.Field{Key: "collection", Value: ks.collectionName},
	)

	return nil
}

// Save 保存文档到知识存储
func (ks *KnowledgeStorageImpl) Save(documents []string, metadata ...interface{}) error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	if !ks.initialized {
		return fmt.Errorf("knowledge storage not initialized")
	}

	ks.logger.Debug("saving documents to knowledge storage",
		logger.Field{Key: "documents_count", Value: len(documents)},
		logger.Field{Key: "collection", Value: ks.collectionName},
	)

	// 处理元数据
	var docMetadata map[string]interface{}
	if len(metadata) > 0 {
		if meta, ok := metadata[0].(map[string]interface{}); ok {
			docMetadata = meta
		}
	}
	if docMetadata == nil {
		docMetadata = make(map[string]interface{})
	}

	// 保存每个文档
	for i, content := range documents {
		// 生成文档ID
		docID := fmt.Sprintf("%s_doc_%d_%d", ks.collectionName, time.Now().UnixNano(), i)

		// 生成嵌入向量（简化实现）
		embedding := ks.generateEmbedding(content)

		// 创建文档
		doc := KnowledgeDocument{
			ID:        docID,
			Content:   content,
			Embedding: embedding,
			Metadata:  docMetadata,
			Source:    ks.collectionName,
			CreatedAt: time.Now(),
		}

		// 如果元数据中有源信息，使用它
		if sourceName, ok := docMetadata["source"].(string); ok {
			doc.Source = sourceName
		}

		// 添加到存储
		ks.documents = append(ks.documents, doc)

		ks.logger.Debug("document saved",
			logger.Field{Key: "doc_id", Value: docID},
			logger.Field{Key: "content_length", Value: len(content)},
			logger.Field{Key: "source", Value: doc.Source},
		)
	}

	ks.logger.Info("documents saved to knowledge storage",
		logger.Field{Key: "total_documents", Value: len(ks.documents)},
		logger.Field{Key: "new_documents", Value: len(documents)},
		logger.Field{Key: "collection", Value: ks.collectionName},
	)

	return nil
}

// Search 搜索知识存储
func (ks *KnowledgeStorageImpl) Search(query []string, limit int, scoreThreshold float64) ([]KnowledgeResult, error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	if !ks.initialized {
		return nil, fmt.Errorf("knowledge storage not initialized")
	}

	// 合并查询字符串
	queryText := ""
	if len(query) > 0 {
		queryText = query[0] // 简化实现，只使用第一个查询
	}

	ks.logger.Debug("searching knowledge storage",
		logger.Field{Key: "query", Value: queryText},
		logger.Field{Key: "limit", Value: limit},
		logger.Field{Key: "threshold", Value: scoreThreshold},
		logger.Field{Key: "collection", Value: ks.collectionName},
	)

	// 生成查询向量
	queryEmbedding := ks.generateEmbedding(queryText)

	// 计算相似度并排序
	var results []ScoredDocument
	for _, doc := range ks.documents {
		score := ks.cosineSimilarity(queryEmbedding, doc.Embedding)

		if score >= scoreThreshold {
			results = append(results, ScoredDocument{
				Document: doc,
				Score:    score,
			})
		}
	}

	// 按分数排序
	ks.sortResultsByScore(results)

	// 限制结果数量
	if len(results) > limit {
		results = results[:limit]
	}

	// 转换为知识结果格式
	knowledgeResults := make([]KnowledgeResult, len(results))
	for i, result := range results {
		knowledgeResults[i] = KnowledgeResult{
			ID:        result.Document.ID,
			Content:   result.Document.Content,
			Source:    result.Document.Source,
			Score:     result.Score,
			Metadata:  result.Document.Metadata,
			CreatedAt: result.Document.CreatedAt,
		}
	}

	ks.logger.Debug("knowledge search completed",
		logger.Field{Key: "results_count", Value: len(knowledgeResults)},
		logger.Field{Key: "query", Value: queryText},
		logger.Field{Key: "collection", Value: ks.collectionName},
	)

	return knowledgeResults, nil
}

// Reset 重置知识存储
func (ks *KnowledgeStorageImpl) Reset() error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	count := len(ks.documents)
	ks.documents = make([]KnowledgeDocument, 0)

	ks.logger.Info("knowledge storage reset",
		logger.Field{Key: "deleted_count", Value: count},
		logger.Field{Key: "collection", Value: ks.collectionName},
	)

	return nil
}

// Close 关闭知识存储
func (ks *KnowledgeStorageImpl) Close() error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	ks.logger.Info("closing knowledge storage",
		logger.Field{Key: "collection", Value: ks.collectionName},
		logger.Field{Key: "documents_count", Value: len(ks.documents)},
	)

	// 清理资源
	ks.documents = nil
	ks.initialized = false

	return nil
}

// ScoredDocument 带分数的文档
type ScoredDocument struct {
	Document KnowledgeDocument
	Score    float64
}

// generateEmbedding 生成嵌入向量（简化实现）
// 在实际实现中，这里应该调用真正的嵌入模型
func (ks *KnowledgeStorageImpl) generateEmbedding(text string) []float64 {
	// 简化实现：基于文本生成伪向量
	// 实际应该使用如OpenAI embeddings、Sentence Transformers等

	embedding := make([]float64, ks.vectorDim)

	// 使用简单的哈希方法生成向量
	hash := 0
	for _, char := range text {
		hash = hash*31 + int(char)
	}

	// 填充向量
	for i := 0; i < ks.vectorDim; i++ {
		// 使用哈希值和索引生成伪随机值
		value := float64((hash*i+i*i)%1000) / 1000.0
		if value > 0.5 {
			value = value - 0.5
		} else {
			value = 0.5 - value
		}
		embedding[i] = value
	}

	// 归一化向量
	ks.normalizeVector(embedding)

	return embedding
}

// cosineSimilarity 计算余弦相似度
func (ks *KnowledgeStorageImpl) cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float64

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (ks.sqrt(normA) * ks.sqrt(normB))
}

// normalizeVector 归一化向量
func (ks *KnowledgeStorageImpl) normalizeVector(vector []float64) {
	var norm float64
	for _, val := range vector {
		norm += val * val
	}

	norm = ks.sqrt(norm)
	if norm == 0 {
		return
	}

	for i := range vector {
		vector[i] /= norm
	}
}

// sqrt 计算平方根（简化实现）
func (ks *KnowledgeStorageImpl) sqrt(x float64) float64 {
	if x == 0 {
		return 0
	}

	// 牛顿法求平方根
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

// sortResultsByScore 按分数排序结果
func (ks *KnowledgeStorageImpl) sortResultsByScore(results []ScoredDocument) {
	// 简单的冒泡排序（实际应使用更高效的排序算法）
	n := len(results)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if results[j].Score < results[j+1].Score {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}
}

// GetStats 获取存储统计信息
func (ks *KnowledgeStorageImpl) GetStats() map[string]interface{} {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	// 统计不同源的文档数量
	sourceStats := make(map[string]int)
	for _, doc := range ks.documents {
		sourceStats[doc.Source]++
	}

	return map[string]interface{}{
		"collection_name":  ks.collectionName,
		"documents_count":  len(ks.documents),
		"vector_dimension": ks.vectorDim,
		"initialized":      ks.initialized,
		"source_stats":     sourceStats,
		"embedder":         ks.embedder,
	}
}
