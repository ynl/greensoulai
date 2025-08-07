package source

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ynl/greensoulai/pkg/logger"
)

// BaseFileKnowledgeSource 文件知识源基类
// 为从文件加载内容的知识源提供基础功能
type BaseFileKnowledgeSource struct {
	*BaseKnowledgeSourceImpl
	filePaths     []string
	content       map[string]string // 文件路径 -> 内容
	safeFilePaths []string          // 验证过的文件路径
}

// NewBaseFileKnowledgeSource 创建文件知识源基类
func NewBaseFileKnowledgeSource(name, sourceType string, filePaths []string, logger logger.Logger) *BaseFileKnowledgeSource {
	base := NewBaseKnowledgeSource(name, sourceType, logger)

	source := &BaseFileKnowledgeSource{
		BaseKnowledgeSourceImpl: base,
		filePaths:               filePaths,
		content:                 make(map[string]string),
		safeFilePaths:           make([]string, 0),
	}

	return source
}

// ValidateContent 验证文件内容
func (bfs *BaseFileKnowledgeSource) ValidateContent() error {
	// 处理文件路径
	if err := bfs.processFilePaths(); err != nil {
		return fmt.Errorf("failed to process file paths: %w", err)
	}

	// 验证至少有一个有效文件
	if len(bfs.safeFilePaths) == 0 {
		return fmt.Errorf("no valid file paths found")
	}

	// 加载内容
	if err := bfs.loadContent(); err != nil {
		return fmt.Errorf("failed to load content: %w", err)
	}

	// 验证内容不为空
	if len(bfs.content) == 0 {
		return fmt.Errorf("no content loaded from files")
	}

	// 处理所有内容
	if err := bfs.processAllContent(); err != nil {
		return fmt.Errorf("failed to process content: %w", err)
	}

	// 执行基础验证
	return bfs.BaseKnowledgeSourceImpl.ValidateContent()
}

// processFilePaths 处理文件路径
func (bfs *BaseFileKnowledgeSource) processFilePaths() error {
	if len(bfs.filePaths) == 0 {
		return fmt.Errorf("no file paths provided")
	}

	bfs.safeFilePaths = make([]string, 0)

	for _, path := range bfs.filePaths {
		// 清理路径
		cleanPath := filepath.Clean(path)

		// 检查文件是否存在
		if _, err := os.Stat(cleanPath); err != nil {
			bfs.logger.Warn("file not found, skipping",
				logger.Field{Key: "file_path", Value: cleanPath},
				logger.Field{Key: "error", Value: err},
			)
			continue
		}

		// 检查是否为常规文件
		fileInfo, err := os.Stat(cleanPath)
		if err != nil {
			continue
		}

		if fileInfo.IsDir() {
			bfs.logger.Warn("path is directory, skipping",
				logger.Field{Key: "file_path", Value: cleanPath},
			)
			continue
		}

		bfs.safeFilePaths = append(bfs.safeFilePaths, cleanPath)
	}

	bfs.logger.Info("file paths processed",
		logger.Field{Key: "source_name", Value: bfs.GetName()},
		logger.Field{Key: "total_paths", Value: len(bfs.filePaths)},
		logger.Field{Key: "valid_paths", Value: len(bfs.safeFilePaths)},
	)

	return nil
}

// loadContent 加载文件内容（抽象方法，由子类实现）
func (bfs *BaseFileKnowledgeSource) loadContent() error {
	// 默认实现：读取文本文件
	for _, filePath := range bfs.safeFilePaths {
		content, err := bfs.readTextFile(filePath)
		if err != nil {
			bfs.logger.Error("failed to read file",
				logger.Field{Key: "file_path", Value: filePath},
				logger.Field{Key: "error", Value: err},
			)
			continue
		}

		bfs.content[filePath] = content

		bfs.logger.Debug("file content loaded",
			logger.Field{Key: "file_path", Value: filePath},
			logger.Field{Key: "content_length", Value: len(content)},
		)
	}

	return nil
}

// readTextFile 读取文本文件
func (bfs *BaseFileKnowledgeSource) readTextFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// processAllContent 处理所有内容
func (bfs *BaseFileKnowledgeSource) processAllContent() error {
	var allChunks []string

	for filePath, fileContent := range bfs.content {
		if fileContent == "" {
			continue
		}

		// 分块处理每个文件的内容
		chunks := bfs.ChunkText(fileContent)

		// 为每个块添加文件信息
		for i, chunk := range chunks {
			// 添加文件信息到块的开头（可选）
			enhancedChunk := fmt.Sprintf("[File: %s, Chunk: %d]\n%s",
				filepath.Base(filePath), i+1, chunk)
			allChunks = append(allChunks, enhancedChunk)
		}

		bfs.logger.Debug("file content processed",
			logger.Field{Key: "file_path", Value: filePath},
			logger.Field{Key: "content_length", Value: len(fileContent)},
			logger.Field{Key: "chunks_generated", Value: len(chunks)},
		)
	}

	// 设置所有块
	bfs.chunks = allChunks

	// 添加文件相关元数据
	bfs.SetMetadata("file_count", len(bfs.safeFilePaths))
	bfs.SetMetadata("file_paths", bfs.safeFilePaths)
	bfs.SetMetadata("total_files_size", bfs.calculateTotalSize())

	bfs.logger.Info("all file content processed",
		logger.Field{Key: "source_name", Value: bfs.GetName()},
		logger.Field{Key: "files_count", Value: len(bfs.safeFilePaths)},
		logger.Field{Key: "total_chunks", Value: len(allChunks)},
	)

	return nil
}

// calculateTotalSize 计算所有文件内容的总大小
func (bfs *BaseFileKnowledgeSource) calculateTotalSize() int {
	totalSize := 0
	for _, content := range bfs.content {
		totalSize += len(content)
	}
	return totalSize
}

// GetFilePaths 获取文件路径列表
func (bfs *BaseFileKnowledgeSource) GetFilePaths() []string {
	return bfs.safeFilePaths
}

// GetFileContent 获取指定文件的内容
func (bfs *BaseFileKnowledgeSource) GetFileContent(filePath string) (string, error) {
	content, exists := bfs.content[filePath]
	if !exists {
		return "", fmt.Errorf("file content not found: %s", filePath)
	}
	return content, nil
}

// GetAllFileContent 获取所有文件内容的映射
func (bfs *BaseFileKnowledgeSource) GetAllFileContent() map[string]string {
	// 返回副本以避免外部修改
	contentCopy := make(map[string]string)
	for path, content := range bfs.content {
		contentCopy[path] = content
	}
	return contentCopy
}

// GetSupportedExtensions 获取支持的文件扩展名（由子类覆盖）
func (bfs *BaseFileKnowledgeSource) GetSupportedExtensions() []string {
	return []string{".txt", ".md", ".log"}
}

// IsFileSupported 检查文件是否支持
func (bfs *BaseFileKnowledgeSource) IsFileSupported(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	supportedExts := bfs.GetSupportedExtensions()

	for _, supportedExt := range supportedExts {
		if ext == supportedExt {
			return true
		}
	}

	return false
}

// GetStats 获取文件知识源统计信息
func (bfs *BaseFileKnowledgeSource) GetStats() map[string]interface{} {
	baseStats := bfs.BaseKnowledgeSourceImpl.GetStats()

	// 添加文件特定统计信息
	baseStats["files_count"] = len(bfs.safeFilePaths)
	baseStats["files_total_size"] = bfs.calculateTotalSize()

	// 按扩展名统计文件
	extStats := make(map[string]int)
	for _, filePath := range bfs.safeFilePaths {
		ext := strings.ToLower(filepath.Ext(filePath))
		if ext == "" {
			ext = "no_extension"
		}
		extStats[ext]++
	}
	baseStats["extension_stats"] = extStats

	// 文件大小统计
	fileSizes := make(map[string]int)
	for filePath, content := range bfs.content {
		fileSizes[filepath.Base(filePath)] = len(content)
	}
	baseStats["file_sizes"] = fileSizes

	return baseStats
}
