package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ynl/greensoulai/internal/memory"
	"github.com/ynl/greensoulai/pkg/logger"
)

func TestNewRAGStorage(t *testing.T) {
	testLogger := logger.NewConsoleLogger()

	config := &memory.EmbedderConfig{
		Provider: "test",
		Config: map[string]interface{}{
			"model": "test-model",
		},
	}

	crew := &struct{ name string }{name: "test-crew"}

	storage := NewRAGStorage("test-type", config, crew, "test-path", testLogger)

	assert.NotNil(t, storage)
}

func TestRAGStorageBasicOperations(t *testing.T) {
	testLogger := logger.NewConsoleLogger()

	config := &memory.EmbedderConfig{
		Provider: "test",
		Config:   map[string]interface{}{},
	}

	crew := &struct{ name string }{name: "test-crew"}
	storage := NewRAGStorage("test-type", config, crew, "test-path", testLogger)

	ctx := context.Background()

	// 测试保存
	item := memory.MemoryItem{
		ID:    "test-1",
		Value: "test content",
		Agent: "test-agent",
	}

	err := storage.Save(ctx, item)
	assert.NoError(t, err)

	// 测试搜索
	results, err := storage.Search(ctx, "test", 10, 0.1)
	assert.NoError(t, err)
	assert.NotNil(t, results)

	// 测试删除
	err = storage.Delete(ctx, "test-1")
	assert.NoError(t, err)

	// 测试清除
	err = storage.Clear(ctx)
	assert.NoError(t, err)
}

func TestRAGStorageEdgeCases(t *testing.T) {
	testLogger := logger.NewConsoleLogger()

	config := &memory.EmbedderConfig{
		Provider: "test",
		Config:   map[string]interface{}{},
	}

	storage := NewRAGStorage("test-type", config, nil, "", testLogger)
	ctx := context.Background()

	// 测试空查询
	results, err := storage.Search(ctx, "", 10, 0.5)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(results)) // 空查询应该返回空结果

	// 测试删除不存在的项目 - 可能返回错误，这是正常的
	err = storage.Delete(ctx, "nonexistent")
	_ = err // 允许返回错误
}

// BenchmarkRAGStorageOperations 性能基准测试
func BenchmarkRAGStorageOperations(b *testing.B) {
	testLogger := logger.NewConsoleLogger()
	config := &memory.EmbedderConfig{
		Provider: "benchmark",
		Config:   map[string]interface{}{},
	}

	storage := NewRAGStorage("benchmark", config, nil, "", testLogger)
	ctx := context.Background()

	b.Run("Save", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			item := memory.MemoryItem{
				ID:    "bench-item",
				Value: "benchmark content",
				Agent: "bench-agent",
			}
			storage.Save(ctx, item)
		}
	})

	b.Run("Search", func(b *testing.B) {
		// 预先添加一些数据
		item := memory.MemoryItem{ID: "search-item", Value: "search content", Agent: "test"}
		storage.Save(ctx, item)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			storage.Search(ctx, "search", 10, 0.1)
		}
	})
}

func TestRAGStorageNilConfig(t *testing.T) {
	testLogger := logger.NewConsoleLogger()

	// 测试nil配置
	storage := NewRAGStorage("test", nil, nil, "", testLogger)
	assert.NotNil(t, storage)

	ctx := context.Background()

	// 基本操作应该仍然工作
	err := storage.Save(ctx, memory.MemoryItem{ID: "test", Value: "content", Agent: "agent"})
	assert.NoError(t, err)

	results, err := storage.Search(ctx, "content", 5, 0.1)
	assert.NoError(t, err)
	assert.NotNil(t, results)
}
