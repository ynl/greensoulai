package source

import (
	"fmt"

	"github.com/ynl/greensoulai/pkg/logger"
)

// StringKnowledgeSource 字符串知识源
// 从字符串内容创建知识源
type StringKnowledgeSource struct {
	*BaseKnowledgeSourceImpl
	content string
}

// NewStringKnowledgeSource 创建字符串知识源
func NewStringKnowledgeSource(name, content string, log logger.Logger) *StringKnowledgeSource {
	base := NewBaseKnowledgeSource(name, "string", log)

	source := &StringKnowledgeSource{
		BaseKnowledgeSourceImpl: base,
		content:                 content,
	}

	// 处理内容
	if err := source.processStringContent(); err != nil {
		log.Error("failed to process string content",
			logger.Field{Key: "source_name", Value: name},
			logger.Field{Key: "error", Value: err},
		)
	}

	return source
}

// processStringContent 处理字符串内容
func (sks *StringKnowledgeSource) processStringContent() error {
	if sks.content == "" {
		return fmt.Errorf("string content is empty")
	}

	// 使用基础实现处理内容
	err := sks.ProcessContent(sks.content)
	if err != nil {
		return fmt.Errorf("failed to process content: %w", err)
	}

	// 添加字符串特定的元数据
	sks.SetMetadata("original_content_preview", sks.getContentPreview())
	sks.SetMetadata("content_type", "plain_text")

	return nil
}

// getContentPreview 获取内容预览（前100个字符）
func (sks *StringKnowledgeSource) getContentPreview() string {
	if len(sks.content) <= 100 {
		return sks.content
	}
	return sks.content[:100] + "..."
}

// GetContent 获取原始内容
func (sks *StringKnowledgeSource) GetContent() string {
	return sks.content
}

// ValidateContent 验证字符串内容
func (sks *StringKnowledgeSource) ValidateContent() error {
	// 先执行基础验证
	if err := sks.BaseKnowledgeSourceImpl.ValidateContent(); err != nil {
		return err
	}

	// 字符串特定验证
	if sks.content == "" {
		return fmt.Errorf("string content is empty")
	}

	if len(sks.content) < 10 {
		return fmt.Errorf("string content too short (minimum 10 characters)")
	}

	sks.logger.Debug("string knowledge source validation passed",
		logger.Field{Key: "source_name", Value: sks.GetName()},
		logger.Field{Key: "content_length", Value: len(sks.content)},
	)

	return nil
}

// UpdateContent 更新字符串内容
func (sks *StringKnowledgeSource) UpdateContent(newContent string) error {
	if newContent == "" {
		return fmt.Errorf("new content cannot be empty")
	}

	sks.logger.Info("updating string knowledge source content",
		logger.Field{Key: "source_name", Value: sks.GetName()},
		logger.Field{Key: "old_length", Value: len(sks.content)},
		logger.Field{Key: "new_length", Value: len(newContent)},
	)

	sks.content = newContent

	// 重新处理内容
	err := sks.processStringContent()
	if err != nil {
		return fmt.Errorf("failed to process updated content: %w", err)
	}

	sks.logger.Info("string knowledge source content updated successfully",
		logger.Field{Key: "source_name", Value: sks.GetName()},
		logger.Field{Key: "new_chunks_count", Value: len(sks.GetChunks())},
	)

	return nil
}

// GetStats 获取字符串知识源统计信息
func (sks *StringKnowledgeSource) GetStats() map[string]interface{} {
	baseStats := sks.BaseKnowledgeSourceImpl.GetStats()

	// 添加字符串特定统计信息
	baseStats["content_length"] = len(sks.content)
	baseStats["content_preview"] = sks.getContentPreview()

	// 统计一些基本的文本特征
	baseStats["word_count"] = sks.countWords()
	baseStats["line_count"] = sks.countLines()
	baseStats["char_count"] = len(sks.content)

	return baseStats
}

// countWords 计算单词数量（简单实现）
func (sks *StringKnowledgeSource) countWords() int {
	if sks.content == "" {
		return 0
	}

	words := 0
	inWord := false

	for _, char := range sks.content {
		if char == ' ' || char == '\t' || char == '\n' || char == '\r' {
			if inWord {
				words++
				inWord = false
			}
		} else {
			inWord = true
		}
	}

	// 如果最后还在单词中，计数加一
	if inWord {
		words++
	}

	return words
}

// countLines 计算行数
func (sks *StringKnowledgeSource) countLines() int {
	if sks.content == "" {
		return 0
	}

	lines := 1
	for _, char := range sks.content {
		if char == '\n' {
			lines++
		}
	}

	return lines
}
