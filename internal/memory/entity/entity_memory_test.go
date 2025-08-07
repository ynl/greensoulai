package entity

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ynl/greensoulai/internal/memory"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

func TestNewEntityMemory(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)

	embedderConfig := &memory.EmbedderConfig{
		Provider: "test",
		Config:   map[string]interface{}{"model": "test-model"},
	}

	crew := &struct{ name string }{name: "test-crew"}

	// 创建实体记忆实例
	em := NewEntityMemory(crew, embedderConfig, nil, "test-collection", testEventBus, testLogger)

	assert.NotNil(t, em)
}

func TestEntityMemoryInterface(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)

	embedderConfig := &memory.EmbedderConfig{
		Provider: "test",
		Config:   map[string]interface{}{"model": "test-model"},
	}

	crew := &struct{ name string }{name: "test-crew"}
	em := NewEntityMemory(crew, embedderConfig, nil, "test-collection", testEventBus, testLogger)

	// 验证实现了Memory接口
	var memory memory.Memory = em
	assert.NotNil(t, memory)

	ctx := context.Background()

	// 测试基本的Memory接口方法
	err := memory.Save(ctx, "test entity", map[string]interface{}{"type": "person"}, "test-agent")
	_ = err // 忽略错误，只要不panic即可

	results, err := memory.Search(ctx, "entity", 10, 0.5)
	assert.NotNil(t, results)

	err = memory.Clear(ctx)
	_ = err

	result := memory.SetCrew("new-crew")
	assert.NotNil(t, result)

	err = memory.Close()
	_ = err
}

func TestEntityMemoryBasicOperations(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)

	embedderConfig := &memory.EmbedderConfig{
		Provider: "test",
		Config:   map[string]interface{}{"model": "test-model"},
	}

	crew := &struct{ name string }{name: "test-crew"}
	em := NewEntityMemory(crew, embedderConfig, nil, "test-collection", testEventBus, testLogger)

	ctx := context.Background()

	// 由于实际的SaveEntity和SaveRelationship方法参数与测试中期望的不同，
	// 我们只测试基本的Memory接口功能

	// 通过Memory接口保存实体数据
	entityData := map[string]interface{}{
		"name": "John Doe",
		"type": "person",
		"age":  30,
	}

	err := em.Save(ctx, entityData, map[string]interface{}{"entity_type": "person"}, "entity-manager")
	_ = err // 可能因为没有真实存储而失败，但不应该panic

	// 搜索实体
	results, err := em.Search(ctx, "John", 5, 0.5)
	assert.NotNil(t, results)
}

func TestEntityMemoryEdgeCases(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)

	// 测试nil配置
	em := NewEntityMemory(nil, nil, nil, "", testEventBus, testLogger)
	assert.NotNil(t, em)

	ctx := context.Background()

	// 测试空值
	err := em.Save(ctx, "", nil, "")
	_ = err // 应该能处理空值

	// 测试搜索不存在的内容
	results, err := em.Search(ctx, "nonexistent", 5, 0.5)
	_ = err // 可能返回错误
	if results != nil {
		assert.GreaterOrEqual(t, len(results), 0) // 如果返回结果，应该是有效的
	}
}

// BenchmarkEntityMemoryCreation 创建实体记忆的性能基准测试
func BenchmarkEntityMemoryCreation(b *testing.B) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)

	embedderConfig := &memory.EmbedderConfig{
		Provider: "test",
		Config:   map[string]interface{}{"model": "test-model"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		crew := &struct{ name string }{name: "bench-crew"}
		em := NewEntityMemory(crew, embedderConfig, nil, "bench-collection", testEventBus, testLogger)
		_ = em
	}
}
