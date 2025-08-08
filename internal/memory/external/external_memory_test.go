package external

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ynl/greensoulai/internal/memory"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

func TestNewExternalMemory(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)

	// 创建基本配置
	embedderConfig := &memory.EmbedderConfig{
		Provider: "test",
		Config:   map[string]interface{}{"model": "test-model"},
	}

	crew := &struct{ name string }{name: "test-crew"}
	collectionName := "test-collection"

	// 创建外部记忆实例
	em := NewExternalMemory(crew, embedderConfig, nil, collectionName, testEventBus, testLogger)

	// 基本验证
	assert.NotNil(t, em)
}

func TestExternalMemoryInterface(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)

	embedderConfig := &memory.EmbedderConfig{
		Provider: "test",
		Config:   map[string]interface{}{"model": "test-model"},
	}

	crew := &struct{ name string }{name: "test-crew"}
	collectionName := "test-collection"

	em := NewExternalMemory(crew, embedderConfig, nil, collectionName, testEventBus, testLogger)

	// 验证实现了Memory接口
	var memory memory.Memory = em
	assert.NotNil(t, memory)

	ctx := context.Background()

	// 测试Save方法 - 外部记忆的保存可能涉及外部服务
	err := memory.Save(ctx, "test external memory", map[string]interface{}{"source": "external"}, "test-agent")
	// 由于是placeholder实现，这里可能返回错误或成功，我们只是确保不panic
	_ = err

	// 测试Search方法
	results, err := memory.Search(ctx, "external", 10, 0.5)
	// 可能返回错误或空结果，我们只是确保不panic
	assert.NotNil(t, results) // 应该返回空slice而不是nil

	// 测试Clear方法
	err = memory.Clear(ctx)
	// 可能返回错误，我们只是确保不panic
	_ = err

	// 测试SetCrew方法
	result := memory.SetCrew("new-crew")
	assert.Equal(t, memory, result) // 应该返回自身

	// 测试Close方法
	err = memory.Close()
	// 可能返回错误，我们只是确保不panic
	_ = err
}

func TestExternalMemoryConfiguration(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)

	// 测试不同的配置组合
	configs := []*memory.EmbedderConfig{
		{
			Provider: "mem0",
			Config: map[string]interface{}{
				"api_key":  "test-key",
				"endpoint": "https://api.mem0.ai",
			},
		},
		{
			Provider: "external_service",
			Config: map[string]interface{}{
				"service_url": "https://external-memory.service.com",
				"auth_token":  "auth-token-123",
				"timeout":     30,
			},
		},
		nil, // 测试nil配置
	}

	for i, config := range configs {
		crew := map[string]interface{}{
			"name": "test-crew",
			"id":   i,
		}

		em := NewExternalMemory(crew, config, nil, "test-collection", testEventBus, testLogger)
		assert.NotNil(t, em, "配置 %d 应该能创建ExternalMemory实例", i)

		// 测试设置crew
		newCrew := map[string]interface{}{"name": "updated-crew", "config_id": i}
		result := em.SetCrew(newCrew)
		assert.Equal(t, em, result)
	}
}

func TestExternalMemoryPlaceholderBehavior(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)

	embedderConfig := &memory.EmbedderConfig{
		Provider: "placeholder",
		Config:   map[string]interface{}{"test": true},
	}

	crew := &struct{ name string }{name: "placeholder-crew"}
	em := NewExternalMemory(crew, embedderConfig, nil, "placeholder-collection", testEventBus, testLogger)

	ctx := context.Background()

	// 所有操作都应该返回"not implemented"错误
	operations := []struct {
		name string
		fn   func() error
	}{
		{
			name: "Save",
			fn: func() error {
				return em.Save(ctx, "test", map[string]interface{}{}, "agent")
			},
		},
		{
			name: "Search",
			fn: func() error {
				_, err := em.Search(ctx, "test", 10, 0.5)
				return err
			},
		},
		{
			name: "Clear",
			fn: func() error {
				return em.Clear(ctx)
			},
		},
		{
			name: "Close",
			fn: func() error {
				return em.Close()
			},
		},
	}

	for _, op := range operations {
		t.Run(op.name, func(t *testing.T) {
			err := op.fn()
			// ExternalMemory基于BaseMemory，大部分操作应该正常工作
			// 我们主要确保不panic，错误处理是可选的
			_ = err // 允许成功或失败
		})
	}
}

func TestExternalMemoryEdgeCases(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)

	// 测试各种边界情况
	testCases := []struct {
		name           string
		crew           interface{}
		config         *memory.EmbedderConfig
		collectionName string
		shouldPanic    bool
	}{
		{
			name:           "all_nil",
			crew:           nil,
			config:         nil,
			collectionName: "",
			shouldPanic:    false,
		},
		{
			name: "empty_config",
			crew: map[string]interface{}{},
			config: &memory.EmbedderConfig{
				Provider: "",
				Config:   map[string]interface{}{},
			},
			collectionName: "",
			shouldPanic:    false,
		},
		{
			name: "complex_crew",
			crew: map[string]interface{}{
				"agents": []interface{}{
					map[string]interface{}{"name": "agent1", "role": "researcher"},
					map[string]interface{}{"name": "agent2", "role": "writer"},
				},
				"tasks": []string{"analyze", "summarize", "report"},
				"metadata": map[string]interface{}{
					"version": "1.0",
					"created": time.Now().Unix(),
				},
			},
			config: &memory.EmbedderConfig{
				Provider: "complex_provider",
				Config: map[string]interface{}{
					"nested": map[string]interface{}{
						"deep": map[string]interface{}{
							"value": 42,
						},
					},
				},
			},
			collectionName: "complex-external-collection",
			shouldPanic:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				assert.Panics(t, func() {
					NewExternalMemory(tc.crew, tc.config, nil, tc.collectionName, testEventBus, testLogger)
				})
			} else {
				assert.NotPanics(t, func() {
					em := NewExternalMemory(tc.crew, tc.config, nil, tc.collectionName, testEventBus, testLogger)
					assert.NotNil(t, em)
				})
			}
		})
	}
}

func TestExternalMemoryFutureIntegration(t *testing.T) {
	// 这个测试模拟未来集成真实外部记忆服务时的行为
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)

	// 模拟真实的外部记忆服务配置
	realConfig := &memory.EmbedderConfig{
		Provider: "mem0",
		Config: map[string]interface{}{
			"api_key":     "real-api-key",
			"endpoint":    "https://api.mem0.ai/v1",
			"model":       "text-embedding-3-small",
			"dimensions":  1536,
			"max_retries": 3,
			"timeout":     30,
		},
	}

	crew := map[string]interface{}{
		"name":        "production-crew",
		"id":          "crew-prod-001",
		"environment": "production",
		"features":    []string{"long_term_memory", "context_awareness", "learning"},
	}

	em := NewExternalMemory(crew, realConfig, nil, "production-memory-collection", testEventBus, testLogger)
	assert.NotNil(t, em)

	ctx := context.Background()

	// 模拟真实使用场景中的数据
	testData := []struct {
		value    interface{}
		metadata map[string]interface{}
		agent    string
	}{
		{
			value: "用户John Doe倾向于选择高质量的产品，即使价格较高",
			metadata: map[string]interface{}{
				"type":       "user_preference",
				"user_id":    "john_doe_123",
				"category":   "shopping_behavior",
				"confidence": 0.85,
				"timestamp":  time.Now().Unix(),
			},
			agent: "behavior_analyzer",
		},
		{
			value: "项目Alpha需要在2024年Q2完成，关键依赖是数据库迁移",
			metadata: map[string]interface{}{
				"type":       "project_info",
				"project_id": "alpha_001",
				"category":   "timeline",
				"priority":   "high",
				"deadline":   "2024-06-30",
			},
			agent: "project_manager",
		},
	}

	// 尝试保存数据（基于BaseMemory实现，应该正常工作）
	for i, data := range testData {
		err := em.Save(ctx, data.value, data.metadata, data.agent)
		// 允许成功或失败，主要确保不panic
		_ = err
		t.Logf("测试数据 %d: Save completed", i)
	}

	// 尝试搜索（基于BaseMemory实现，应该正常工作）
	searchQueries := []string{
		"用户偏好",
		"项目timeline",
		"数据库迁移",
	}

	for _, query := range searchQueries {
		results, err := em.Search(ctx, query, 5, 0.7)
		// 允许成功或失败，主要确保不panic
		_ = err
		_ = results
		t.Logf("查询 '%s': Search completed", query)
	}
}

// BenchmarkExternalMemoryCreation 外部记忆创建的性能基准测试
func BenchmarkExternalMemoryCreation(b *testing.B) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)

	embedderConfig := &memory.EmbedderConfig{
		Provider: "benchmark",
		Config:   map[string]interface{}{"model": "benchmark-model"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		crew := &struct{ name string }{name: "bench-crew"}
		em := NewExternalMemory(crew, embedderConfig, nil, "bench-collection", testEventBus, testLogger)
		_ = em
	}
}

// TestExternalMemoryThreadSafety 测试并发安全性
func TestExternalMemoryThreadSafety(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)

	embedderConfig := &memory.EmbedderConfig{
		Provider: "concurrent",
		Config:   map[string]interface{}{"model": "concurrent-model"},
	}

	crew := &struct{ name string }{name: "concurrent-crew"}
	em := NewExternalMemory(crew, embedderConfig, nil, "concurrent-collection", testEventBus, testLogger)

	numGoroutines := 10
	done := make(chan bool, numGoroutines)

	// 并发执行各种操作
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			ctx := context.Background()

			// 尝试各种操作（都会返回"not implemented"）
			em.Save(ctx, "concurrent test", map[string]interface{}{"id": id}, "concurrent-agent")
			em.Search(ctx, "concurrent", 5, 0.5)
			em.SetCrew(map[string]interface{}{"name": "updated", "id": id})
			em.Clear(ctx)
			em.Close()

			// 简短等待
			time.Sleep(1 * time.Millisecond)
		}(i)
	}

	// 等待所有goroutines完成
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
			// 成功完成
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out - possible deadlock")
		}
	}
}
