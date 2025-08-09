package source

import (
	"testing"

	"github.com/ynl/greensoulai/pkg/logger"
)

func TestNewStringKnowledgeSource(t *testing.T) {
	logger := logger.NewConsoleLogger()

	source := NewStringKnowledgeSource("test_source", "This is test content", logger)

	if source == nil {
		t.Fatal("NewStringKnowledgeSource returned nil")
	}

	if source.GetName() != "test_source" {
		t.Errorf("Expected name 'test_source', got '%s'", source.GetName())
	}

	if source.GetType() != "string" {
		t.Errorf("Expected type 'string', got '%s'", source.GetType())
	}

	if source.GetContent() != "This is test content" {
		t.Errorf("Expected content 'This is test content', got '%s'", source.GetContent())
	}
}

func TestStringKnowledgeSource_ValidateContent(t *testing.T) {
	logger := logger.NewConsoleLogger()

	tests := []struct {
		name        string
		sourceName  string
		content     string
		expectError bool
	}{
		{
			name:        "valid content",
			sourceName:  "test_source",
			content:     "This is valid test content with enough characters",
			expectError: false,
		},
		{
			name:        "empty content",
			sourceName:  "test_source",
			content:     "",
			expectError: true,
		},
		{
			name:        "content too short",
			sourceName:  "test_source",
			content:     "short",
			expectError: true,
		},
		{
			name:        "empty source name",
			sourceName:  "",
			content:     "This is valid test content",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := NewStringKnowledgeSource(tt.sourceName, tt.content, logger)
			err := source.ValidateContent()

			if tt.expectError && err == nil {
				t.Errorf("Expected error for test '%s', but got nil", tt.name)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for test '%s', but got: %v", tt.name, err)
			}
		})
	}
}

func TestStringKnowledgeSource_UpdateContent(t *testing.T) {
	logger := logger.NewConsoleLogger()

	source := NewStringKnowledgeSource("test_source", "original content", logger)

	// 测试成功更新
	err := source.UpdateContent("new updated content")
	if err != nil {
		t.Fatalf("UpdateContent failed: %v", err)
	}

	if source.GetContent() != "new updated content" {
		t.Errorf("Expected updated content 'new updated content', got '%s'", source.GetContent())
	}

	// 测试空内容更新
	err = source.UpdateContent("")
	if err == nil {
		t.Error("Expected error when updating with empty content, but got nil")
	}
}

func TestStringKnowledgeSource_GetStats(t *testing.T) {
	logger := logger.NewConsoleLogger()

	content := "This is a test content with multiple words.\nIt has two lines."
	source := NewStringKnowledgeSource("test_source", content, logger)

	stats := source.GetStats()

	// 验证基本统计信息
	if stats["name"] != "test_source" {
		t.Errorf("Expected name 'test_source', got '%v'", stats["name"])
	}

	if stats["type"] != "string" {
		t.Errorf("Expected type 'string', got '%v'", stats["type"])
	}

	if stats["content_length"] != len(content) {
		t.Errorf("Expected content_length %d, got '%v'", len(content), stats["content_length"])
	}

	// 验证单词统计
	expectedWordCount := 12 // 根据内容计算的单词数
	if wordCount, ok := stats["word_count"].(int); !ok || wordCount != expectedWordCount {
		t.Errorf("Expected word_count %d, got '%v'", expectedWordCount, stats["word_count"])
	}

	// 验证行统计
	expectedLineCount := 2 // 两行
	if lineCount, ok := stats["line_count"].(int); !ok || lineCount != expectedLineCount {
		t.Errorf("Expected line_count %d, got '%v'", expectedLineCount, stats["line_count"])
	}

	// 验证字符统计
	if charCount, ok := stats["char_count"].(int); !ok || charCount != len(content) {
		t.Errorf("Expected char_count %d, got '%v'", len(content), stats["char_count"])
	}
}

func TestStringKnowledgeSource_countWords(t *testing.T) {
	logger := logger.NewConsoleLogger()

	tests := []struct {
		name          string
		content       string
		expectedWords int
	}{
		{
			name:          "empty content",
			content:       "",
			expectedWords: 0,
		},
		{
			name:          "single word",
			content:       "hello",
			expectedWords: 1,
		},
		{
			name:          "multiple words",
			content:       "hello world test",
			expectedWords: 3,
		},
		{
			name:          "words with tabs and newlines",
			content:       "hello\tworld\ntest",
			expectedWords: 3,
		},
		{
			name:          "words with multiple spaces",
			content:       "hello   world    test",
			expectedWords: 3,
		},
		{
			name:          "content with punctuation",
			content:       "hello, world! test?",
			expectedWords: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := NewStringKnowledgeSource("test_source", tt.content, logger)
			wordCount := source.countWords()

			if wordCount != tt.expectedWords {
				t.Errorf("Expected word count %d, got %d for content: '%s'",
					tt.expectedWords, wordCount, tt.content)
			}
		})
	}
}

func TestStringKnowledgeSource_countLines(t *testing.T) {
	logger := logger.NewConsoleLogger()

	tests := []struct {
		name          string
		content       string
		expectedLines int
	}{
		{
			name:          "empty content",
			content:       "",
			expectedLines: 0,
		},
		{
			name:          "single line",
			content:       "hello world",
			expectedLines: 1,
		},
		{
			name:          "two lines",
			content:       "hello\nworld",
			expectedLines: 2,
		},
		{
			name:          "multiple lines",
			content:       "line1\nline2\nline3",
			expectedLines: 3,
		},
		{
			name:          "lines with empty lines",
			content:       "line1\n\nline3",
			expectedLines: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := NewStringKnowledgeSource("test_source", tt.content, logger)
			lineCount := source.countLines()

			if lineCount != tt.expectedLines {
				t.Errorf("Expected line count %d, got %d for content: '%s'",
					tt.expectedLines, lineCount, tt.content)
			}
		})
	}
}

func TestStringKnowledgeSource_getContentPreview(t *testing.T) {
	logger := logger.NewConsoleLogger()

	tests := []struct {
		name            string
		content         string
		expectedPreview string
	}{
		{
			name:            "short content",
			content:         "short",
			expectedPreview: "short",
		},
		{
			name:            "exactly 100 chars",
			content:         "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
			expectedPreview: "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
		},
		{
			name:            "more than 100 chars",
			content:         "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890EXTRA",
			expectedPreview: "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := NewStringKnowledgeSource("test_source", tt.content, logger)
			preview := source.getContentPreview()

			if preview != tt.expectedPreview {
				t.Errorf("Expected preview '%s', got '%s'", tt.expectedPreview, preview)
			}
		})
	}
}

// 测试知识源的分块功能
func TestStringKnowledgeSource_ChunkText(t *testing.T) {
	logger := logger.NewConsoleLogger()

	// 创建一个长内容用于测试分块（减少循环次数避免超时）
	longContent := ""
	for i := 0; i < 100; i++ {
		longContent += "This is a test sentence for chunking. "
	}

	source := NewStringKnowledgeSource("test_source", longContent, logger)

	chunks := source.GetChunks()
	if len(chunks) == 0 {
		t.Error("Expected chunks to be generated, but got none")
	}

	// 验证分块逻辑
	if len(chunks) > 1 {
		// 检查块大小
		for i, chunk := range chunks {
			if i < len(chunks)-1 { // 除了最后一个块
				if len(chunk) > source.chunkSize {
					t.Errorf("Chunk %d size %d exceeds chunk_size %d", i, len(chunk), source.chunkSize)
				}
			}
		}
	}

	// 验证所有块的总长度近似等于原始内容长度
	totalChunkLength := 0
	for _, chunk := range chunks {
		totalChunkLength += len(chunk)
	}

	// 考虑重叠部分，总长度可能会稍大于原始长度
	if totalChunkLength < len(longContent) {
		t.Errorf("Total chunk length %d is less than original content length %d",
			totalChunkLength, len(longContent))
	}
}

// 性能基准测试
func BenchmarkStringKnowledgeSource_countWords(b *testing.B) {
	logger := logger.NewConsoleLogger()

	// 创建大量文本内容
	content := ""
	for i := 0; i < 10000; i++ {
		content += "word "
	}

	source := NewStringKnowledgeSource("benchmark_source", content, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		source.countWords()
	}
}

func BenchmarkStringKnowledgeSource_ChunkText(b *testing.B) {
	logger := logger.NewConsoleLogger()

	// 创建大量文本内容
	content := ""
	for i := 0; i < 50000; i++ {
		content += "This is a test sentence for chunking performance benchmarking. "
	}

	source := NewStringKnowledgeSource("benchmark_source", content, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		source.ChunkText(content)
	}
}
