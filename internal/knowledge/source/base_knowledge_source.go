package source

import (
	"fmt"

	"github.com/ynl/greensoulai/internal/knowledge"
	"github.com/ynl/greensoulai/pkg/logger"
)

// BaseKnowledgeSourceImpl 基础知识源实现
type BaseKnowledgeSourceImpl struct {
	name         string
	sourceType   string
	chunkSize    int
	chunkOverlap int
	chunks       []string
	embeddings   [][]float64
	storage      knowledge.KnowledgeStorage
	metadata     map[string]interface{}
	logger       logger.Logger
}

// NewBaseKnowledgeSource 创建基础知识源
func NewBaseKnowledgeSource(name, sourceType string, logger logger.Logger) *BaseKnowledgeSourceImpl {
	return &BaseKnowledgeSourceImpl{
		name:         name,
		sourceType:   sourceType,
		chunkSize:    4000, // 默认块大小
		chunkOverlap: 200,  // 默认重叠大小
		chunks:       make([]string, 0),
		embeddings:   make([][]float64, 0),
		metadata:     make(map[string]interface{}),
		logger:       logger,
	}
}

// GetName 获取源名称
func (bs *BaseKnowledgeSourceImpl) GetName() string {
	return bs.name
}

// GetType 获取源类型
func (bs *BaseKnowledgeSourceImpl) GetType() string {
	return bs.sourceType
}

// ValidateContent 验证内容（基础实现）
func (bs *BaseKnowledgeSourceImpl) ValidateContent() error {
	if bs.name == "" {
		return fmt.Errorf("knowledge source name cannot be empty")
	}

	if bs.sourceType == "" {
		return fmt.Errorf("knowledge source type cannot be empty")
	}

	if len(bs.chunks) == 0 {
		return fmt.Errorf("knowledge source has no content chunks")
	}

	bs.logger.Debug("knowledge source validation passed",
		logger.Field{Key: "source_name", Value: bs.name},
		logger.Field{Key: "source_type", Value: bs.sourceType},
		logger.Field{Key: "chunks_count", Value: len(bs.chunks)},
	)

	return nil
}

// Add 添加到知识存储
func (bs *BaseKnowledgeSourceImpl) Add() error {
	if bs.storage == nil {
		return fmt.Errorf("knowledge storage not set")
	}

	if len(bs.chunks) == 0 {
		return fmt.Errorf("no content chunks to add")
	}

	bs.logger.Info("adding knowledge source to storage",
		logger.Field{Key: "source_name", Value: bs.name},
		logger.Field{Key: "chunks_count", Value: len(bs.chunks)},
	)

	// 添加源信息到元数据
	bs.metadata["source_name"] = bs.name
	bs.metadata["source_type"] = bs.sourceType
	bs.metadata["chunk_size"] = bs.chunkSize
	bs.metadata["chunk_overlap"] = bs.chunkOverlap

	// 保存到存储
	err := bs.storage.Save(bs.chunks, bs.metadata)
	if err != nil {
		bs.logger.Error("failed to save knowledge source to storage",
			logger.Field{Key: "source_name", Value: bs.name},
			logger.Field{Key: "error", Value: err},
		)
		return fmt.Errorf("failed to save to storage: %w", err)
	}

	bs.logger.Info("knowledge source added to storage successfully",
		logger.Field{Key: "source_name", Value: bs.name},
	)

	return nil
}

// GetEmbeddings 获取嵌入向量
func (bs *BaseKnowledgeSourceImpl) GetEmbeddings() ([][]float64, error) {
	if len(bs.embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings available")
	}
	return bs.embeddings, nil
}

// SetStorage 设置存储
func (bs *BaseKnowledgeSourceImpl) SetStorage(storage knowledge.KnowledgeStorage) {
	bs.storage = storage
}

// GetChunks 获取块
func (bs *BaseKnowledgeSourceImpl) GetChunks() []string {
	return bs.chunks
}

// GetMetadata 获取元数据
func (bs *BaseKnowledgeSourceImpl) GetMetadata() map[string]interface{} {
	return bs.metadata
}

// SetChunkSize 设置块大小
func (bs *BaseKnowledgeSourceImpl) SetChunkSize(size int) {
	if size > 0 {
		bs.chunkSize = size
	}
}

// SetChunkOverlap 设置块重叠
func (bs *BaseKnowledgeSourceImpl) SetChunkOverlap(overlap int) {
	if overlap >= 0 {
		bs.chunkOverlap = overlap
	}
}

// SetMetadata 设置元数据
func (bs *BaseKnowledgeSourceImpl) SetMetadata(key string, value interface{}) {
	bs.metadata[key] = value
}

// ChunkText 将文本分块
func (bs *BaseKnowledgeSourceImpl) ChunkText(text string) []string {
	if text == "" {
		return []string{}
	}

	// 简单的分块实现
	var chunks []string
	textLen := len(text)

	if textLen <= bs.chunkSize {
		// 文本长度小于块大小，直接返回
		chunks = append(chunks, text)
	} else {
		// 按块大小分割
		start := 0
		for start < textLen {
			end := start + bs.chunkSize
			if end > textLen {
				end = textLen
			}

			chunk := text[start:end]
			chunks = append(chunks, chunk)

			// 计算下一个块的起始位置（考虑重叠）
			start = end - bs.chunkOverlap
			if start <= 0 || start >= textLen {
				break
			}
		}
	}

	bs.logger.Debug("text chunked",
		logger.Field{Key: "source_name", Value: bs.name},
		logger.Field{Key: "original_length", Value: textLen},
		logger.Field{Key: "chunks_count", Value: len(chunks)},
		logger.Field{Key: "chunk_size", Value: bs.chunkSize},
		logger.Field{Key: "overlap", Value: bs.chunkOverlap},
	)

	return chunks
}

// ProcessContent 处理内容（设置块）
func (bs *BaseKnowledgeSourceImpl) ProcessContent(content string) error {
	if content == "" {
		return fmt.Errorf("content is empty")
	}

	// 将内容分块
	bs.chunks = bs.ChunkText(content)

	// 添加内容相关的元数据
	bs.metadata["content_length"] = len(content)
	bs.metadata["processed_at"] = "now" // 简化时间处理

	bs.logger.Info("content processed",
		logger.Field{Key: "source_name", Value: bs.name},
		logger.Field{Key: "content_length", Value: len(content)},
		logger.Field{Key: "chunks_generated", Value: len(bs.chunks)},
	)

	return nil
}

// GetStats 获取知识源统计信息
func (bs *BaseKnowledgeSourceImpl) GetStats() map[string]interface{} {
	totalLength := 0
	for _, chunk := range bs.chunks {
		totalLength += len(chunk)
	}

	avgChunkSize := 0
	if len(bs.chunks) > 0 {
		avgChunkSize = totalLength / len(bs.chunks)
	}

	return map[string]interface{}{
		"name":                 bs.name,
		"type":                 bs.sourceType,
		"chunks_count":         len(bs.chunks),
		"total_content_length": totalLength,
		"avg_chunk_size":       avgChunkSize,
		"chunk_size":           bs.chunkSize,
		"chunk_overlap":        bs.chunkOverlap,
		"embeddings_count":     len(bs.embeddings),
		"metadata":             bs.metadata,
	}
}
