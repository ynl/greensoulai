package source

import (
	"fmt"
	"path/filepath"

	"github.com/ynl/greensoulai/pkg/logger"
)

// TextFileKnowledgeSource 文本文件知识源
// 从文本文件加载内容创建知识源
type TextFileKnowledgeSource struct {
	*BaseFileKnowledgeSource
}

// NewTextFileKnowledgeSource 创建文本文件知识源
func NewTextFileKnowledgeSource(name string, filePaths []string, logger logger.Logger) *TextFileKnowledgeSource {
	base := NewBaseFileKnowledgeSource(name, "text_file", filePaths, logger)

	source := &TextFileKnowledgeSource{
		BaseFileKnowledgeSource: base,
	}

	return source
}

// NewTextFileKnowledgeSourceFromSingleFile 从单个文件创建文本文件知识源
func NewTextFileKnowledgeSourceFromSingleFile(filePath string, logger logger.Logger) *TextFileKnowledgeSource {
	name := fmt.Sprintf("text_file_%s", filepath.Base(filePath))
	return NewTextFileKnowledgeSource(name, []string{filePath}, logger)
}

// GetSupportedExtensions 获取支持的文件扩展名
func (tfs *TextFileKnowledgeSource) GetSupportedExtensions() []string {
	return []string{
		".txt",  // 纯文本文件
		".md",   // Markdown文件
		".mdx",  // MDX文件
		".log",  // 日志文件
		".csv",  // CSV文件（作为文本处理）
		".json", // JSON文件（作为文本处理）
		".yaml", // YAML文件
		".yml",  // YAML文件
		".xml",  // XML文件
		".html", // HTML文件（作为文本处理）
		".rst",  // reStructuredText文件
		".tex",  // LaTeX文件
		".py",   // Python源文件（作为文本处理）
		".go",   // Go源文件（作为文本处理）
		".js",   // JavaScript源文件（作为文本处理）
		".ts",   // TypeScript源文件（作为文本处理）
		".java", // Java源文件（作为文本处理）
		".cpp",  // C++源文件（作为文本处理）
		".c",    // C源文件（作为文本处理）
		".h",    // C/C++头文件（作为文本处理）
		"",      // 无扩展名的文件
	}
}

// ValidateContent 验证文本文件内容
func (tfs *TextFileKnowledgeSource) ValidateContent() error {
	// 首先验证文件路径和基础内容
	if err := tfs.BaseFileKnowledgeSource.ValidateContent(); err != nil {
		return err
	}

	// 检查文件扩展名是否支持
	unsupportedFiles := make([]string, 0)
	for _, filePath := range tfs.GetFilePaths() {
		if !tfs.IsFileSupported(filePath) {
			unsupportedFiles = append(unsupportedFiles, filePath)
		}
	}

	if len(unsupportedFiles) > 0 {
		tfs.logger.Warn("some files have unsupported extensions",
			logger.Field{Key: "source_name", Value: tfs.GetName()},
			logger.Field{Key: "unsupported_files", Value: unsupportedFiles},
		)
	}

	// 验证文本内容质量
	if err := tfs.validateTextQuality(); err != nil {
		return fmt.Errorf("text quality validation failed: %w", err)
	}

	tfs.logger.Debug("text file knowledge source validation passed",
		logger.Field{Key: "source_name", Value: tfs.GetName()},
		logger.Field{Key: "files_count", Value: len(tfs.GetFilePaths())},
	)

	return nil
}

// validateTextQuality 验证文本质量
func (tfs *TextFileKnowledgeSource) validateTextQuality() error {
	totalContentLength := 0
	emptyFiles := 0

	for filePath, content := range tfs.GetAllFileContent() {
		if content == "" {
			emptyFiles++
			tfs.logger.Warn("empty file found",
				logger.Field{Key: "file_path", Value: filePath},
			)
			continue
		}

		totalContentLength += len(content)

		// 检查内容是否过短
		if len(content) < 50 {
			tfs.logger.Warn("file content very short",
				logger.Field{Key: "file_path", Value: filePath},
				logger.Field{Key: "content_length", Value: len(content)},
			)
		}
	}

	// 验证总体内容
	if totalContentLength == 0 {
		return fmt.Errorf("no valid text content found in any file")
	}

	if emptyFiles == len(tfs.GetFilePaths()) {
		return fmt.Errorf("all files are empty")
	}

	tfs.logger.Info("text quality validation completed",
		logger.Field{Key: "source_name", Value: tfs.GetName()},
		logger.Field{Key: "total_content_length", Value: totalContentLength},
		logger.Field{Key: "empty_files_count", Value: emptyFiles},
	)

	return nil
}

// PreprocessText 预处理文本（可以被子类覆盖）
func (tfs *TextFileKnowledgeSource) PreprocessText(text string) string {
	// 基本的文本清理
	// 移除多余的空行
	lines := make([]string, 0)
	currentLine := ""

	for _, char := range text {
		if char == '\n' {
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = ""
			}
		} else {
			currentLine += string(char)
		}
	}

	// 添加最后一行（如果有）
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	// 重新组合文本
	result := ""
	for i, line := range lines {
		result += line
		if i < len(lines)-1 {
			result += "\n"
		}
	}

	return result
}

// GetTextStats 获取文本统计信息
func (tfs *TextFileKnowledgeSource) GetTextStats() map[string]interface{} {
	stats := make(map[string]interface{})

	totalChars := 0
	totalLines := 0
	totalWords := 0

	for _, content := range tfs.GetAllFileContent() {
		totalChars += len(content)

		// 计算行数
		for _, char := range content {
			if char == '\n' {
				totalLines++
			}
		}

		// 计算单词数（简单实现）
		inWord := false
		for _, char := range content {
			if char == ' ' || char == '\t' || char == '\n' || char == '\r' {
				if inWord {
					totalWords++
					inWord = false
				}
			} else {
				inWord = true
			}
		}
		if inWord {
			totalWords++
		}
	}

	stats["total_characters"] = totalChars
	stats["total_lines"] = totalLines
	stats["total_words"] = totalWords

	if len(tfs.GetFilePaths()) > 0 {
		stats["avg_chars_per_file"] = totalChars / len(tfs.GetFilePaths())
		stats["avg_lines_per_file"] = totalLines / len(tfs.GetFilePaths())
		stats["avg_words_per_file"] = totalWords / len(tfs.GetFilePaths())
	}

	return stats
}

// GetStats 获取文本文件知识源统计信息
func (tfs *TextFileKnowledgeSource) GetStats() map[string]interface{} {
	baseStats := tfs.BaseFileKnowledgeSource.GetStats()

	// 添加文本特定统计信息
	textStats := tfs.GetTextStats()
	for key, value := range textStats {
		baseStats[key] = value
	}

	// 添加支持的扩展名信息
	baseStats["supported_extensions"] = tfs.GetSupportedExtensions()

	return baseStats
}

