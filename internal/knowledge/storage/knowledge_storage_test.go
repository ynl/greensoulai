package storage

import (
	"math"
	"testing"

	"github.com/ynl/greensoulai/pkg/logger"
)

func TestNewKnowledgeStorage(t *testing.T) {
	logger := logger.NewConsoleLogger()
	embedder := &EmbedderConfig{
		Provider: "test_provider",
		Config:   map[string]interface{}{"key": "value"},
	}

	storage := NewKnowledgeStorage(embedder, "test_collection", logger)

	if storage == nil {
		t.Fatal("NewKnowledgeStorage returned nil")
	}

	if storage.collectionName != "test_collection" {
		t.Errorf("Expected collection name 'test_collection', got '%s'", storage.collectionName)
	}

	if storage.embedder.Provider != "test_provider" {
		t.Errorf("Expected embedder provider 'test_provider', got '%s'", storage.embedder.Provider)
	}

	if storage.vectorDim != 384 {
		t.Errorf("Expected default vector dimension 384, got %d", storage.vectorDim)
	}

	if storage.initialized {
		t.Error("Storage should not be initialized immediately after creation")
	}
}

func TestKnowledgeStorage_InitializeKnowledgeStorage(t *testing.T) {
	logger := logger.NewConsoleLogger()
	storage := NewKnowledgeStorage(nil, "test_collection", logger)

	err := storage.InitializeKnowledgeStorage()
	if err != nil {
		t.Fatalf("InitializeKnowledgeStorage failed: %v", err)
	}

	if !storage.initialized {
		t.Error("Storage should be initialized after calling InitializeKnowledgeStorage")
	}

	// 测试重复初始化
	err = storage.InitializeKnowledgeStorage()
	if err != nil {
		t.Errorf("Repeat initialization should not fail: %v", err)
	}
}

func TestKnowledgeStorage_Save(t *testing.T) {
	logger := logger.NewConsoleLogger()
	storage := NewKnowledgeStorage(nil, "test_collection", logger)
	storage.InitializeKnowledgeStorage()

	documents := []string{
		"This is the first document",
		"This is the second document",
		"This is the third document",
	}

	metadata := map[string]interface{}{
		"source": "test_source",
		"type":   "test",
	}

	err := storage.Save(documents, metadata)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// 验证文档已保存
	if len(storage.documents) != len(documents) {
		t.Errorf("Expected %d documents, got %d", len(documents), len(storage.documents))
	}

	// 验证文档内容
	for i, doc := range storage.documents {
		if doc.Content != documents[i] {
			t.Errorf("Document %d: expected content '%s', got '%s'",
				i, documents[i], doc.Content)
		}

		if doc.Source != "test_source" {
			t.Errorf("Document %d: expected source 'test_source', got '%s'",
				i, doc.Source)
		}

		if len(doc.Embedding) != storage.vectorDim {
			t.Errorf("Document %d: expected embedding dimension %d, got %d",
				i, storage.vectorDim, len(doc.Embedding))
		}
	}
}

func TestKnowledgeStorage_SaveNotInitialized(t *testing.T) {
	logger := logger.NewConsoleLogger()
	storage := NewKnowledgeStorage(nil, "test_collection", logger)
	// 故意不初始化

	documents := []string{"test document"}

	err := storage.Save(documents)
	if err == nil {
		t.Error("Expected error when saving to uninitialized storage, but got nil")
	}
}

func TestKnowledgeStorage_Search(t *testing.T) {
	logger := logger.NewConsoleLogger()
	storage := NewKnowledgeStorage(nil, "test_collection", logger)
	storage.InitializeKnowledgeStorage()

	// 先保存一些文档
	documents := []string{
		"Go programming language tutorial",
		"Python programming basics",
		"JavaScript web development",
		"Machine learning with Python",
	}

	metadata := map[string]interface{}{
		"source": "programming_docs",
		"type":   "tutorial",
	}

	storage.Save(documents, metadata)

	// 测试搜索
	query := []string{"programming"}
	results, err := storage.Search(query, 10, 0.0)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected search results, but got none")
	}

	// 验证结果结构
	for _, result := range results {
		if result.ID == "" {
			t.Error("Result ID should not be empty")
		}

		if result.Content == "" {
			t.Error("Result content should not be empty")
		}

		if result.Score < 0 || result.Score > 1 {
			t.Errorf("Result score %f should be between 0 and 1", result.Score)
		}

		if result.Source != "programming_docs" {
			t.Errorf("Expected source 'programming_docs', got '%s'", result.Source)
		}
	}
}

func TestKnowledgeStorage_SearchWithThreshold(t *testing.T) {
	logger := logger.NewConsoleLogger()
	storage := NewKnowledgeStorage(nil, "test_collection", logger)
	storage.InitializeKnowledgeStorage()

	documents := []string{
		"Relevant document about programming",
		"Another relevant programming document",
	}

	storage.Save(documents)

	// 测试高阈值搜索
	query := []string{"programming"}
	results, err := storage.Search(query, 10, 0.9)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// 由于是模拟实现，这里主要验证阈值逻辑工作正常
	// 在真实实现中，高阈值会过滤掉低相关性的结果
	for _, result := range results {
		if result.Score < 0.9 {
			t.Errorf("Result score %f should be >= threshold 0.9", result.Score)
		}
	}
}

func TestKnowledgeStorage_SearchNotInitialized(t *testing.T) {
	logger := logger.NewConsoleLogger()
	storage := NewKnowledgeStorage(nil, "test_collection", logger)
	// 故意不初始化

	query := []string{"test"}
	_, err := storage.Search(query, 10, 0.5)
	if err == nil {
		t.Error("Expected error when searching uninitialized storage, but got nil")
	}
}

func TestKnowledgeStorage_Reset(t *testing.T) {
	logger := logger.NewConsoleLogger()
	storage := NewKnowledgeStorage(nil, "test_collection", logger)
	storage.InitializeKnowledgeStorage()

	// 保存一些文档
	documents := []string{"doc1", "doc2", "doc3"}
	storage.Save(documents)

	// 验证文档已保存
	if len(storage.documents) == 0 {
		t.Error("Documents should be saved before reset")
	}

	// 重置存储
	err := storage.Reset()
	if err != nil {
		t.Fatalf("Reset failed: %v", err)
	}

	// 验证文档已清空
	if len(storage.documents) != 0 {
		t.Errorf("Expected 0 documents after reset, got %d", len(storage.documents))
	}
}

func TestKnowledgeStorage_Close(t *testing.T) {
	logger := logger.NewConsoleLogger()
	storage := NewKnowledgeStorage(nil, "test_collection", logger)
	storage.InitializeKnowledgeStorage()

	// 保存一些文档
	documents := []string{"doc1", "doc2"}
	storage.Save(documents)

	err := storage.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// 验证存储已关闭
	if storage.initialized {
		t.Error("Storage should not be initialized after close")
	}

	if storage.documents != nil {
		t.Error("Documents should be nil after close")
	}
}

func TestKnowledgeStorage_generateEmbedding(t *testing.T) {
	logger := logger.NewConsoleLogger()
	storage := NewKnowledgeStorage(nil, "test_collection", logger)

	text := "This is a test text for embedding generation"
	embedding := storage.generateEmbedding(text)

	// 验证向量维度
	if len(embedding) != storage.vectorDim {
		t.Errorf("Expected embedding dimension %d, got %d", storage.vectorDim, len(embedding))
	}

	// 验证向量归一化
	norm := 0.0
	for _, val := range embedding {
		norm += val * val
	}
	norm = math.Sqrt(norm)

	if math.Abs(norm-1.0) > 0.001 {
		t.Errorf("Expected normalized vector (norm=1.0), got norm=%f", norm)
	}

	// 验证相同文本生成相同向量
	embedding2 := storage.generateEmbedding(text)
	for i := 0; i < len(embedding); i++ {
		if math.Abs(embedding[i]-embedding2[i]) > 0.001 {
			t.Errorf("Same text should generate same embedding at position %d", i)
			break
		}
	}

	// 验证不同文本生成不同向量
	differentText := "This is completely different text"
	differentEmbedding := storage.generateEmbedding(differentText)

	identical := true
	for i := 0; i < len(embedding); i++ {
		if math.Abs(embedding[i]-differentEmbedding[i]) > 0.001 {
			identical = false
			break
		}
	}

	if identical {
		t.Error("Different texts should generate different embeddings")
	}
}

func TestKnowledgeStorage_cosineSimilarity(t *testing.T) {
	logger := logger.NewConsoleLogger()
	storage := NewKnowledgeStorage(nil, "test_collection", logger)

	// 测试相同向量的相似度
	vec1 := []float64{1.0, 0.0, 0.0}
	vec2 := []float64{1.0, 0.0, 0.0}

	similarity := storage.cosineSimilarity(vec1, vec2)
	if math.Abs(similarity-1.0) > 0.001 {
		t.Errorf("Identical vectors should have similarity 1.0, got %f", similarity)
	}

	// 测试正交向量的相似度
	vec3 := []float64{1.0, 0.0, 0.0}
	vec4 := []float64{0.0, 1.0, 0.0}

	similarity = storage.cosineSimilarity(vec3, vec4)
	if math.Abs(similarity-0.0) > 0.001 {
		t.Errorf("Orthogonal vectors should have similarity 0.0, got %f", similarity)
	}

	// 测试相反向量的相似度
	vec5 := []float64{1.0, 0.0, 0.0}
	vec6 := []float64{-1.0, 0.0, 0.0}

	similarity = storage.cosineSimilarity(vec5, vec6)
	if math.Abs(similarity-(-1.0)) > 0.001 {
		t.Errorf("Opposite vectors should have similarity -1.0, got %f", similarity)
	}

	// 测试不同长度的向量
	vec7 := []float64{1.0, 0.0}
	vec8 := []float64{1.0, 0.0, 0.0}

	similarity = storage.cosineSimilarity(vec7, vec8)
	if similarity != 0.0 {
		t.Errorf("Vectors of different lengths should have similarity 0.0, got %f", similarity)
	}
}

func TestKnowledgeStorage_normalizeVector(t *testing.T) {
	logger := logger.NewConsoleLogger()
	storage := NewKnowledgeStorage(nil, "test_collection", logger)

	// 测试正常向量归一化
	vector := []float64{3.0, 4.0, 0.0}
	storage.normalizeVector(vector)

	// 计算归一化后的模长
	norm := 0.0
	for _, val := range vector {
		norm += val * val
	}
	norm = math.Sqrt(norm)

	if math.Abs(norm-1.0) > 0.001 {
		t.Errorf("Normalized vector should have norm 1.0, got %f", norm)
	}

	// 验证归一化后的值
	expectedX := 3.0 / 5.0 // 3/sqrt(3^2+4^2)
	expectedY := 4.0 / 5.0 // 4/sqrt(3^2+4^2)

	if math.Abs(vector[0]-expectedX) > 0.001 {
		t.Errorf("Expected x component %f, got %f", expectedX, vector[0])
	}

	if math.Abs(vector[1]-expectedY) > 0.001 {
		t.Errorf("Expected y component %f, got %f", expectedY, vector[1])
	}

	// 测试零向量
	zeroVector := []float64{0.0, 0.0, 0.0}
	storage.normalizeVector(zeroVector)

	// 零向量归一化后应该保持为零
	for i, val := range zeroVector {
		if val != 0.0 {
			t.Errorf("Zero vector component %d should remain 0.0 after normalization, got %f", i, val)
		}
	}
}

func TestKnowledgeStorage_sqrt(t *testing.T) {
	logger := logger.NewConsoleLogger()
	storage := NewKnowledgeStorage(nil, "test_collection", logger)

	tests := []struct {
		input    float64
		expected float64
	}{
		{0.0, 0.0},
		{1.0, 1.0},
		{4.0, 2.0},
		{9.0, 3.0},
		{16.0, 4.0},
		{25.0, 5.0},
	}

	for _, test := range tests {
		result := storage.sqrt(test.input)
		if math.Abs(result-test.expected) > 0.001 {
			t.Errorf("sqrt(%f) = %f, expected %f", test.input, result, test.expected)
		}
	}
}

func TestKnowledgeStorage_sortResultsByScore(t *testing.T) {
	logger := logger.NewConsoleLogger()
	storage := NewKnowledgeStorage(nil, "test_collection", logger)

	results := []ScoredDocument{
		{Document: KnowledgeDocument{ID: "doc1"}, Score: 0.5},
		{Document: KnowledgeDocument{ID: "doc2"}, Score: 0.8},
		{Document: KnowledgeDocument{ID: "doc3"}, Score: 0.3},
		{Document: KnowledgeDocument{ID: "doc4"}, Score: 0.9},
	}

	storage.sortResultsByScore(results)

	// 验证结果按分数降序排列
	expectedOrder := []string{"doc4", "doc2", "doc1", "doc3"}
	expectedScores := []float64{0.9, 0.8, 0.5, 0.3}

	for i, result := range results {
		if result.Document.ID != expectedOrder[i] {
			t.Errorf("Position %d: expected ID %s, got %s",
				i, expectedOrder[i], result.Document.ID)
		}

		if math.Abs(result.Score-expectedScores[i]) > 0.001 {
			t.Errorf("Position %d: expected score %f, got %f",
				i, expectedScores[i], result.Score)
		}
	}
}

func TestKnowledgeStorage_GetStats(t *testing.T) {
	logger := logger.NewConsoleLogger()
	embedder := &EmbedderConfig{
		Provider: "test_provider",
		Config:   map[string]interface{}{"test": "config"},
	}

	storage := NewKnowledgeStorage(embedder, "test_collection", logger)
	storage.InitializeKnowledgeStorage()

	// 添加一些文档
	metadata1 := map[string]interface{}{"source": "source1"}
	metadata2 := map[string]interface{}{"source": "source2"}

	storage.Save([]string{"doc1", "doc2"}, metadata1)
	storage.Save([]string{"doc3"}, metadata2)

	stats := storage.GetStats()

	// 验证统计信息
	if stats["collection_name"] != "test_collection" {
		t.Errorf("Expected collection_name 'test_collection', got '%v'", stats["collection_name"])
	}

	if stats["documents_count"] != 3 {
		t.Errorf("Expected documents_count 3, got '%v'", stats["documents_count"])
	}

	if stats["vector_dimension"] != 384 {
		t.Errorf("Expected vector_dimension 384, got '%v'", stats["vector_dimension"])
	}

	if stats["initialized"] != true {
		t.Errorf("Expected initialized true, got '%v'", stats["initialized"])
	}

	// 验证源统计
	sourceStats, ok := stats["source_stats"].(map[string]int)
	if !ok {
		t.Error("Expected source_stats to be map[string]int")
	} else {
		if sourceStats["source1"] != 2 {
			t.Errorf("Expected source1 count 2, got %d", sourceStats["source1"])
		}
		if sourceStats["source2"] != 1 {
			t.Errorf("Expected source2 count 1, got %d", sourceStats["source2"])
		}
	}

	// 验证嵌入器配置
	if embedderStats, ok := stats["embedder"].(*EmbedderConfig); !ok {
		t.Error("Expected embedder to be *EmbedderConfig")
	} else if embedderStats.Provider != "test_provider" {
		t.Errorf("Expected embedder provider 'test_provider', got '%s'", embedderStats.Provider)
	}
}

// 性能基准测试
func BenchmarkKnowledgeStorage_generateEmbedding(b *testing.B) {
	logger := logger.NewConsoleLogger()
	storage := NewKnowledgeStorage(nil, "benchmark_collection", logger)

	text := "This is a test text for embedding generation benchmarking purposes"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.generateEmbedding(text)
	}
}

func BenchmarkKnowledgeStorage_cosineSimilarity(b *testing.B) {
	logger := logger.NewConsoleLogger()
	storage := NewKnowledgeStorage(nil, "benchmark_collection", logger)

	// 生成两个测试向量
	vec1 := make([]float64, 384)
	vec2 := make([]float64, 384)

	for i := 0; i < 384; i++ {
		vec1[i] = float64(i) / 384.0
		vec2[i] = float64(i*2) / 384.0
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.cosineSimilarity(vec1, vec2)
	}
}

func BenchmarkKnowledgeStorage_Search(b *testing.B) {
	logger := logger.NewConsoleLogger()
	storage := NewKnowledgeStorage(nil, "benchmark_collection", logger)
	storage.InitializeKnowledgeStorage()

	// 添加大量文档
	documents := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		documents[i] = "This is test document content for benchmarking search performance"
	}

	storage.Save(documents)

	query := []string{"test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.Search(query, 10, 0.5)
	}
}
