package knowledge

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// MockKnowledgeStorage 用于测试的模拟存储
type MockKnowledgeStorage struct {
	documents   []MockDocument
	initialized bool
}

type MockDocument struct {
	content  string
	metadata map[string]interface{}
}

func (m *MockKnowledgeStorage) InitializeKnowledgeStorage() error {
	m.initialized = true
	return nil
}

func (m *MockKnowledgeStorage) Save(documents []string, metadata ...interface{}) error {
	for _, doc := range documents {
		mockDoc := MockDocument{
			content:  doc,
			metadata: make(map[string]interface{}),
		}
		if len(metadata) > 0 {
			if meta, ok := metadata[0].(map[string]interface{}); ok {
				mockDoc.metadata = meta
			}
		}
		m.documents = append(m.documents, mockDoc)
	}
	return nil
}

func (m *MockKnowledgeStorage) Search(query []string, limit int, scoreThreshold float64) ([]KnowledgeResult, error) {
	var results []KnowledgeResult

	for i, doc := range m.documents {
		if i >= limit {
			break
		}

		// 简单的文本匹配作为模拟搜索
		score := 0.8 // 固定高分数用于测试
		if len(query) > 0 && query[0] != "" {
			// 简单检查查询词是否在文档中
			found := false
			for _, q := range query {
				if len(doc.content) > 0 && len(q) > 0 {
					found = true
					break
				}
			}
			if !found {
				score = 0.2
			}
		}

		if score >= scoreThreshold {
			results = append(results, KnowledgeResult{
				ID:        string(rune('0' + i)),
				Content:   doc.content,
				Source:    "test_source",
				Score:     score,
				Metadata:  doc.metadata,
				CreatedAt: time.Now(),
			})
		}
	}

	return results, nil
}

func (m *MockKnowledgeStorage) Reset() error {
	m.documents = nil
	return nil
}

func (m *MockKnowledgeStorage) Close() error {
	m.initialized = false
	return nil
}

// MockKnowledgeSource 用于测试的模拟知识源
type MockKnowledgeSource struct {
	name     string
	typ      string
	chunks   []string
	metadata map[string]interface{}
	storage  KnowledgeStorage
}

func NewMockKnowledgeSource(name, typ string) *MockKnowledgeSource {
	return &MockKnowledgeSource{
		name:     name,
		typ:      typ,
		chunks:   []string{"test chunk 1", "test chunk 2"},
		metadata: make(map[string]interface{}),
	}
}

func (m *MockKnowledgeSource) GetName() string {
	return m.name
}

func (m *MockKnowledgeSource) GetType() string {
	return m.typ
}

func (m *MockKnowledgeSource) ValidateContent() error {
	if m.name == "" {
		return fmt.Errorf("source name is empty")
	}
	if len(m.chunks) == 0 {
		return fmt.Errorf("no chunks available")
	}
	return nil
}

func (m *MockKnowledgeSource) Add() error {
	if m.storage == nil {
		return fmt.Errorf("storage not set")
	}
	return m.storage.Save(m.chunks, m.metadata)
}

func (m *MockKnowledgeSource) GetEmbeddings() ([][]float64, error) {
	return [][]float64{{0.1, 0.2, 0.3}}, nil
}

func (m *MockKnowledgeSource) SetStorage(storage KnowledgeStorage) {
	m.storage = storage
}

func (m *MockKnowledgeSource) GetChunks() []string {
	return m.chunks
}

func (m *MockKnowledgeSource) GetMetadata() map[string]interface{} {
	return m.metadata
}

func TestNewKnowledge(t *testing.T) {
	logger := logger.NewConsoleLogger()
	eventBus := events.NewEventBus(logger)
	mockStorage := &MockKnowledgeStorage{}

	knowledge := NewKnowledge(
		"test_collection",
		[]BaseKnowledgeSource{},
		nil,
		mockStorage,
		eventBus,
		logger,
	)

	if knowledge == nil {
		t.Fatal("NewKnowledge returned nil")
	}

	if knowledge.collectionName != "test_collection" {
		t.Errorf("Expected collection name 'test_collection', got '%s'", knowledge.collectionName)
	}

	if len(knowledge.sources) != 0 {
		t.Errorf("Expected 0 sources, got %d", len(knowledge.sources))
	}
}

func TestKnowledge_AddSource(t *testing.T) {
	logger := logger.NewConsoleLogger()
	eventBus := events.NewEventBus(logger)
	mockStorage := &MockKnowledgeStorage{}

	knowledge := NewKnowledge(
		"test_collection",
		[]BaseKnowledgeSource{},
		nil,
		mockStorage,
		eventBus,
		logger,
	)

	// 创建测试知识源
	mockSource := NewMockKnowledgeSource("test_source", "mock")

	// 添加知识源
	err := knowledge.AddSource(mockSource)
	if err != nil {
		t.Fatalf("AddSource failed: %v", err)
	}

	// 验证源已添加
	sources := knowledge.GetSources()
	if len(sources) != 1 {
		t.Errorf("Expected 1 source, got %d", len(sources))
	}

	if sources[0].GetName() != "test_source" {
		t.Errorf("Expected source name 'test_source', got '%s'", sources[0].GetName())
	}
}

func TestKnowledge_AddSourceDuplicate(t *testing.T) {
	logger := logger.NewConsoleLogger()
	eventBus := events.NewEventBus(logger)
	mockStorage := &MockKnowledgeStorage{}

	knowledge := NewKnowledge(
		"test_collection",
		[]BaseKnowledgeSource{},
		nil,
		mockStorage,
		eventBus,
		logger,
	)

	// 创建测试知识源
	mockSource1 := NewMockKnowledgeSource("test_source", "mock")
	mockSource2 := NewMockKnowledgeSource("test_source", "mock")

	// 添加第一个源
	err := knowledge.AddSource(mockSource1)
	if err != nil {
		t.Fatalf("First AddSource failed: %v", err)
	}

	// 尝试添加重复的源
	err = knowledge.AddSource(mockSource2)
	if err == nil {
		t.Error("Expected error when adding duplicate source, but got nil")
	}

	// 验证只有一个源
	sources := knowledge.GetSources()
	if len(sources) != 1 {
		t.Errorf("Expected 1 source after duplicate add, got %d", len(sources))
	}
}

func TestKnowledge_RemoveSource(t *testing.T) {
	logger := logger.NewConsoleLogger()
	eventBus := events.NewEventBus(logger)
	mockStorage := &MockKnowledgeStorage{}

	knowledge := NewKnowledge(
		"test_collection",
		[]BaseKnowledgeSource{},
		nil,
		mockStorage,
		eventBus,
		logger,
	)

	// 添加知识源
	mockSource := NewMockKnowledgeSource("test_source", "mock")
	knowledge.AddSource(mockSource)

	// 移除知识源
	err := knowledge.RemoveSource("test_source")
	if err != nil {
		t.Fatalf("RemoveSource failed: %v", err)
	}

	// 验证源已移除
	sources := knowledge.GetSources()
	if len(sources) != 0 {
		t.Errorf("Expected 0 sources after removal, got %d", len(sources))
	}
}

func TestKnowledge_RemoveSourceNotFound(t *testing.T) {
	logger := logger.NewConsoleLogger()
	eventBus := events.NewEventBus(logger)
	mockStorage := &MockKnowledgeStorage{}

	knowledge := NewKnowledge(
		"test_collection",
		[]BaseKnowledgeSource{},
		nil,
		mockStorage,
		eventBus,
		logger,
	)

	// 尝试移除不存在的源
	err := knowledge.RemoveSource("nonexistent")
	if err == nil {
		t.Error("Expected error when removing nonexistent source, but got nil")
	}
}

func TestKnowledge_Query(t *testing.T) {
	logger := logger.NewConsoleLogger()
	eventBus := events.NewEventBus(logger)
	mockStorage := &MockKnowledgeStorage{
		documents: []MockDocument{
			{content: "Go is a programming language", metadata: map[string]interface{}{"type": "info"}},
			{content: "Python is also a programming language", metadata: map[string]interface{}{"type": "info"}},
		},
		initialized: true,
	}

	knowledge := NewKnowledge(
		"test_collection",
		[]BaseKnowledgeSource{},
		nil,
		mockStorage,
		eventBus,
		logger,
	)

	ctx := context.Background()
	results, err := knowledge.Query(ctx, []string{"programming"}, 10, 0.5)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one result, got none")
	}

	// 验证结果结构
	if len(results) > 0 {
		result := results[0]
		if result.Content == "" {
			t.Error("Result content is empty")
		}
		if result.Score < 0.5 {
			t.Errorf("Result score %f is below threshold", result.Score)
		}
	}
}

func TestKnowledge_AddSources(t *testing.T) {
	logger := logger.NewConsoleLogger()
	eventBus := events.NewEventBus(logger)
	mockStorage := &MockKnowledgeStorage{}

	knowledge := NewKnowledge(
		"test_collection",
		[]BaseKnowledgeSource{},
		nil,
		mockStorage,
		eventBus,
		logger,
	)

	// 添加多个知识源
	mockSource1 := NewMockKnowledgeSource("source1", "mock")
	mockSource2 := NewMockKnowledgeSource("source2", "mock")

	knowledge.AddSource(mockSource1)
	knowledge.AddSource(mockSource2)

	// 将所有源添加到存储
	err := knowledge.AddSources()
	if err != nil {
		t.Fatalf("AddSources failed: %v", err)
	}

	// 验证文档已保存到存储
	if len(mockStorage.documents) == 0 {
		t.Error("Expected documents to be saved to storage")
	}
}

func TestKnowledge_Reset(t *testing.T) {
	logger := logger.NewConsoleLogger()
	eventBus := events.NewEventBus(logger)
	mockStorage := &MockKnowledgeStorage{
		documents: []MockDocument{
			{content: "test", metadata: map[string]interface{}{}},
		},
		initialized: true,
	}

	knowledge := NewKnowledge(
		"test_collection",
		[]BaseKnowledgeSource{},
		nil,
		mockStorage,
		eventBus,
		logger,
	)

	// 重置存储
	err := knowledge.Reset()
	if err != nil {
		t.Fatalf("Reset failed: %v", err)
	}

	// 验证存储已清空
	if len(mockStorage.documents) != 0 {
		t.Error("Expected storage to be empty after reset")
	}
}

func TestKnowledge_Close(t *testing.T) {
	logger := logger.NewConsoleLogger()
	eventBus := events.NewEventBus(logger)
	mockStorage := &MockKnowledgeStorage{initialized: true}

	knowledge := NewKnowledge(
		"test_collection",
		[]BaseKnowledgeSource{},
		nil,
		mockStorage,
		eventBus,
		logger,
	)

	// 关闭知识系统
	err := knowledge.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// 验证存储已关闭
	if mockStorage.initialized {
		t.Error("Expected storage to be closed")
	}
}

func TestKnowledge_GetStats(t *testing.T) {
	logger := logger.NewConsoleLogger()
	eventBus := events.NewEventBus(logger)
	mockStorage := &MockKnowledgeStorage{}

	knowledge := NewKnowledge(
		"test_collection",
		[]BaseKnowledgeSource{},
		nil,
		mockStorage,
		eventBus,
		logger,
	)

	// 添加一些知识源
	mockSource := NewMockKnowledgeSource("test_source", "mock")
	knowledge.AddSource(mockSource)

	stats := knowledge.GetStats()

	if stats["collection_name"] != "test_collection" {
		t.Errorf("Expected collection name 'test_collection', got '%v'", stats["collection_name"])
	}

	if stats["sources_count"] != 1 {
		t.Errorf("Expected sources count 1, got '%v'", stats["sources_count"])
	}

	// 验证统计信息的结构
	if _, ok := stats["source_types"]; !ok {
		t.Error("Expected source_types in stats")
	}

	if _, ok := stats["total_chunks"]; !ok {
		t.Error("Expected total_chunks in stats")
	}
}

// 性能基准测试
func BenchmarkKnowledge_Query(b *testing.B) {
	logger := logger.NewConsoleLogger()
	eventBus := events.NewEventBus(logger)

	// 创建包含大量文档的存储
	documents := make([]MockDocument, 1000)
	for i := 0; i < 1000; i++ {
		documents[i] = MockDocument{
			content:  "This is test document content for benchmarking purposes",
			metadata: map[string]interface{}{"index": i},
		}
	}

	mockStorage := &MockKnowledgeStorage{
		documents:   documents,
		initialized: true,
	}

	knowledge := NewKnowledge(
		"benchmark_collection",
		[]BaseKnowledgeSource{},
		nil,
		mockStorage,
		eventBus,
		logger,
	)

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := knowledge.Query(ctx, []string{"test"}, 10, 0.5)
		if err != nil {
			b.Fatalf("Query failed: %v", err)
		}
	}
}
