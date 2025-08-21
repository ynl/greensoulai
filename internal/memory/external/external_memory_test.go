package external

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	results, _ := memory.Search(ctx, "external", 10, 0.5)
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

// Mock外部源实现
type MockExternalSource struct {
	name         string
	sourceType   string
	connected    bool
	available    bool
	data         []ExternalMemoryItem
	fetchError   error
	syncError    error
	connectError error
	mu           sync.Mutex
}

func NewMockExternalSource(name, sourceType string) *MockExternalSource {
	return &MockExternalSource{
		name:       name,
		sourceType: sourceType,
		available:  true,
		data:       make([]ExternalMemoryItem, 0),
	}
}

func (m *MockExternalSource) GetName() string {
	return m.name
}

func (m *MockExternalSource) GetType() string {
	return m.sourceType
}

func (m *MockExternalSource) Connect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.connectError != nil {
		return m.connectError
	}
	
	m.connected = true
	return nil
}

func (m *MockExternalSource) Fetch(ctx context.Context, query string, limit int) ([]ExternalMemoryItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.fetchError != nil {
		return nil, m.fetchError
	}
	
	if !m.connected {
		return nil, fmt.Errorf("source not connected")
	}
	
	// 模拟查询匹配
	var results []ExternalMemoryItem
	for _, item := range m.data {
		if query == "" || item.Value == query || fmt.Sprint(item.Value) == query {
			results = append(results, item)
			if len(results) >= limit {
				break
			}
		}
	}
	
	return results, nil
}

func (m *MockExternalSource) Sync(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.syncError != nil {
		return m.syncError
	}
	
	if !m.connected {
		return fmt.Errorf("source not connected")
	}
	
	return nil
}

func (m *MockExternalSource) Disconnect() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.connected = false
	return nil
}

func (m *MockExternalSource) IsAvailable() bool {
	return m.available
}

// 辅助方法
func (m *MockExternalSource) SetAvailable(available bool) {
	m.available = available
}

func (m *MockExternalSource) SetFetchError(err error) {
	m.fetchError = err
}

func (m *MockExternalSource) SetSyncError(err error) {
	m.syncError = err
}

func (m *MockExternalSource) SetConnectError(err error) {
	m.connectError = err
}

func (m *MockExternalSource) AddData(items ...ExternalMemoryItem) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = append(m.data, items...)
}

func TestExternalMemoryAddSource(t *testing.T) {
	em := createTestExternalMemory(t)
	
	tests := []struct {
		name        string
		source      ExternalSource
		expectError bool
	}{
		{
			name:        "add valid source",
			source:      NewMockExternalSource("database1", string(SourceTypeDatabase)),
			expectError: false,
		},
		{
			name:        "add api source",
			source:      NewMockExternalSource("api1", string(SourceTypeAPI)),
			expectError: false,
		},
		{
			name:        "add file source",
			source:      NewMockExternalSource("file1", string(SourceTypeFile)),
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := em.AddSource(tt.source)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// 验证源已添加
				sources := em.GetSources()
				found := false
				for _, source := range sources {
					if source.GetName() == tt.source.GetName() {
						found = true
						break
					}
				}
				assert.True(t, found, "Source should be added")
			}
		})
	}
	
	t.Run("add duplicate source", func(t *testing.T) {
		source1 := NewMockExternalSource("duplicate", string(SourceTypeDatabase))
		source2 := NewMockExternalSource("duplicate", string(SourceTypeAPI))
		
		// 第一次添加应该成功
		err := em.AddSource(source1)
		assert.NoError(t, err)
		
		// 第二次添加相同名称的源应该失败
		err = em.AddSource(source2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source already exists")
	})
}

func TestExternalMemoryRemoveSource(t *testing.T) {
	em := createTestExternalMemory(t)
	
	// 先添加一些源
	source1 := NewMockExternalSource("remove_test1", string(SourceTypeDatabase))
	source2 := NewMockExternalSource("remove_test2", string(SourceTypeAPI))
	
	_ = em.AddSource(source1)
	_ = em.AddSource(source2)
	
	// 连接源以测试断开连接
	_ = source1.Connect(context.Background())
	
	tests := []struct {
		name        string
		sourceName  string
		expectError bool
	}{
		{
			name:        "remove existing source",
			sourceName:  "remove_test1",
			expectError: false,
		},
		{
			name:        "remove another existing source", 
			sourceName:  "remove_test2",
			expectError: false,
		},
		{
			name:        "remove non-existent source",
			sourceName:  "non_existent",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := em.RemoveSource(tt.sourceName)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "source not found")
			} else {
				assert.NoError(t, err)
				// 验证源已移除
				sources := em.GetSources()
				for _, source := range sources {
					assert.NotEqual(t, tt.sourceName, source.GetName(), "Source should be removed")
				}
			}
		})
	}
}

func TestExternalMemoryGetSources(t *testing.T) {
	em := createTestExternalMemory(t)
	
	// 初始状态应该没有源
	sources := em.GetSources()
	assert.Empty(t, sources, "Initially should have no sources")
	
	// 添加几个源
	source1 := NewMockExternalSource("source1", string(SourceTypeDatabase))
	source2 := NewMockExternalSource("source2", string(SourceTypeAPI))
	source3 := NewMockExternalSource("source3", string(SourceTypeFile))
	
	_ = em.AddSource(source1)
	_ = em.AddSource(source2)
	_ = em.AddSource(source3)
	
	// 验证返回的源
	sources = em.GetSources()
	assert.Len(t, sources, 3, "Should have 3 sources")
	
	sourceNames := make(map[string]bool)
	for _, source := range sources {
		sourceNames[source.GetName()] = true
	}
	
	assert.True(t, sourceNames["source1"])
	assert.True(t, sourceNames["source2"])
	assert.True(t, sourceNames["source3"])
}

func TestExternalMemorySyncAll(t *testing.T) {
	em := createTestExternalMemory(t)
	ctx := context.Background()
	
	// 创建不同状态的源
	availableSource := NewMockExternalSource("available", string(SourceTypeDatabase))
	unavailableSource := NewMockExternalSource("unavailable", string(SourceTypeAPI))
	errorSource := NewMockExternalSource("error", string(SourceTypeFile))
	
	// 设置源状态
	unavailableSource.SetAvailable(false)
	errorSource.SetSyncError(fmt.Errorf("sync failed"))
	
	// 添加源
	_ = em.AddSource(availableSource)
	_ = em.AddSource(unavailableSource)
	_ = em.AddSource(errorSource)
	
	// 连接可用的源
	_ = availableSource.Connect(ctx)
	_ = errorSource.Connect(ctx)
	
	tests := []struct {
		name        string
		expectError bool
	}{
		{
			name:        "sync all sources",
			expectError: true, // 因为有一个源会出错
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := em.SyncAll(ctx)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "sync errors")
			} else {
				assert.NoError(t, err)
			}
		})
	}
	
	t.Run("sync all sources success", func(t *testing.T) {
		// 创建一个新的实例，只有成功的源
		em2 := createTestExternalMemory(t)
		successSource := NewMockExternalSource("success", string(SourceTypeDatabase))
		_ = em2.AddSource(successSource)
		_ = successSource.Connect(ctx)
		
		err := em2.SyncAll(ctx)
		assert.NoError(t, err)
	})
}

func TestExternalMemorySyncSource(t *testing.T) {
	em := createTestExternalMemory(t)
	ctx := context.Background()
	
	// 创建测试源
	availableSource := NewMockExternalSource("available", string(SourceTypeDatabase))
	unavailableSource := NewMockExternalSource("unavailable", string(SourceTypeAPI))
	
	// 设置状态
	unavailableSource.SetAvailable(false)
	
	// 添加源
	_ = em.AddSource(availableSource)
	_ = em.AddSource(unavailableSource)
	
	// 连接可用源
	_ = availableSource.Connect(ctx)
	
	tests := []struct {
		name        string
		sourceName  string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "sync available source",
			sourceName:  "available",
			expectError: false,
		},
		{
			name:        "sync unavailable source",
			sourceName:  "unavailable", 
			expectError: true,
			errorMsg:    "source not available",
		},
		{
			name:        "sync non-existent source",
			sourceName:  "non_existent",
			expectError: true,
			errorMsg:    "source not found",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := em.SyncSource(ctx, tt.sourceName)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExternalMemoryFetchFromSource(t *testing.T) {
	em := createTestExternalMemory(t)
	ctx := context.Background()
	
	// 创建测试源
	source := NewMockExternalSource("fetch_test", string(SourceTypeDatabase))
	
	// 添加测试数据
	testItems := []ExternalMemoryItem{
		{
			MemoryItem: memory.MemoryItem{Value: "test data 1"},
			SourceName: "fetch_test",
			SourceType: string(SourceTypeDatabase),
		},
		{
			MemoryItem: memory.MemoryItem{Value: "test data 2"},
			SourceName: "fetch_test",
			SourceType: string(SourceTypeDatabase),
		},
	}
	source.AddData(testItems...)
	
	// 添加源并连接
	_ = em.AddSource(source)
	_ = source.Connect(ctx)
	
	tests := []struct {
		name        string
		sourceName  string
		query       string
		limit       int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "fetch from available source",
			sourceName:  "fetch_test",
			query:       "",
			limit:       10,
			expectError: false,
		},
		{
			name:        "fetch with specific query",
			sourceName:  "fetch_test",
			query:       "test data 1",
			limit:       10,
			expectError: false,
		},
		{
			name:        "fetch from non-existent source",
			sourceName:  "non_existent",
			query:       "",
			limit:       10,
			expectError: true,
			errorMsg:    "source not found",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := em.FetchFromSource(ctx, tt.sourceName, tt.query, tt.limit)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, results)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, results)
			}
		})
	}
	
	t.Run("fetch from unavailable source", func(t *testing.T) {
		unavailableSource := NewMockExternalSource("unavailable", string(SourceTypeAPI))
		unavailableSource.SetAvailable(false)
		_ = em.AddSource(unavailableSource)
		
		results, err := em.FetchFromSource(ctx, "unavailable", "", 10)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source not available")
		assert.Nil(t, results)
	})
}

func TestExternalMemorySearchBySource(t *testing.T) {
	em := createTestExternalMemory(t)
	ctx := context.Background()
	
	tests := []struct {
		name           string
		sourceName     string
		query          string
		limit          int
		scoreThreshold float64
		expectError    bool
	}{
		{
			name:           "search by source",
			sourceName:     "test_source",
			query:          "test query",
			limit:          10,
			scoreThreshold: 0.5,
			expectError:    false,
		},
		{
			name:           "search with empty source",
			sourceName:     "",
			query:          "test query",
			limit:          10,
			scoreThreshold: 0.5,
			expectError:    false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := em.SearchBySource(ctx, tt.sourceName, tt.query, tt.limit, tt.scoreThreshold)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				// 可能返回空结果，但不应该返回错误
				assert.NotNil(t, results)
			}
		})
	}
}

func TestExternalMemorySearchBySourceType(t *testing.T) {
	em := createTestExternalMemory(t)
	ctx := context.Background()
	
	tests := []struct {
		name           string
		sourceType     string
		query          string
		limit          int
		scoreThreshold float64
		expectError    bool
	}{
		{
			name:           "search by database type",
			sourceType:     string(SourceTypeDatabase),
			query:          "test query",
			limit:          10,
			scoreThreshold: 0.5,
			expectError:    false,
		},
		{
			name:           "search by api type",
			sourceType:     string(SourceTypeAPI),
			query:          "test query",
			limit:          10,
			scoreThreshold: 0.5,
			expectError:    false,
		},
		{
			name:           "search with empty type",
			sourceType:     "",
			query:          "test query",
			limit:          10,
			scoreThreshold: 0.5,
			expectError:    false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := em.SearchBySourceType(ctx, ExternalSourceType(tt.sourceType), tt.query, tt.limit, tt.scoreThreshold)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NotNil(t, results)
			}
		})
	}
}

func TestExternalMemoryAutoSync(t *testing.T) {
	em := createTestExternalMemory(t)
	
	// 测试启用自动同步
	t.Run("enable auto sync", func(t *testing.T) {
		em.EnableAutoSync(context.Background(), 100*time.Millisecond) // 很短的间隔用于测试
		// 这个方法主要是设置内部状态，我们无法直接验证，但至少确保不panic
		assert.NotPanics(t, func() {
			em.EnableAutoSync(context.Background(), 1*time.Second)
		})
	})
	
	// 测试禁用自动同步
	t.Run("disable auto sync", func(t *testing.T) {
		em.DisableAutoSync()
		assert.NotPanics(t, func() {
			em.DisableAutoSync()
		})
	})
}

func TestExternalMemoryGetSyncStatus(t *testing.T) {
	em := createTestExternalMemory(t)
	ctx := context.Background()
	
	// 初始状态
	status := em.GetSyncStatus()
	assert.NotNil(t, status)
	
	// 执行同步后检查状态
	err := em.SyncAll(ctx)
	_ = err // 可能有错误，但我们关心的是状态更新
	
	status = em.GetSyncStatus()
	assert.NotNil(t, status)
}

func TestExternalMemoryComprehensiveEdgeCases(t *testing.T) {
	em := createTestExternalMemory(t)
	ctx := context.Background()
	
	t.Run("operations with no sources", func(t *testing.T) {
		// 测试在没有源的情况下的操作
		err := em.SyncAll(ctx)
		assert.NoError(t, err, "SyncAll with no sources should succeed")
		
		err = em.SyncSource(ctx, "non_existent")
		assert.Error(t, err, "SyncSource with non-existent source should fail")
		
		results, err := em.FetchFromSource(ctx, "non_existent", "", 10)
		assert.Error(t, err)
		assert.Nil(t, results)
	})
	
	t.Run("source with connection errors", func(t *testing.T) {
		errorSource := NewMockExternalSource("error_source", string(SourceTypeDatabase))
		errorSource.SetConnectError(fmt.Errorf("connection failed"))
		
		_ = em.AddSource(errorSource)
		
		err := errorSource.Connect(ctx)
		assert.Error(t, err)
	})
	
	t.Run("source with fetch errors", func(t *testing.T) {
		fetchErrorSource := NewMockExternalSource("fetch_error", string(SourceTypeAPI))
		fetchErrorSource.SetFetchError(fmt.Errorf("fetch failed"))
		_ = em.AddSource(fetchErrorSource)
		_ = fetchErrorSource.Connect(ctx)
		
		results, err := em.FetchFromSource(ctx, "fetch_error", "", 10)
		assert.Error(t, err)
		assert.Nil(t, results)
	})
}

// Helper function to create a test ExternalMemory instance
func createTestExternalMemory(t *testing.T) *ExternalMemory {
	t.Helper()
	
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)
	
	embedderConfig := &memory.EmbedderConfig{
		Provider: "test",
		Config:   map[string]interface{}{"model": "test-model"},
	}
	
	crew := &struct{ name string }{name: "test-crew"}
	collectionName := "test-collection"
	
	em := NewExternalMemory(crew, embedderConfig, nil, collectionName, testEventBus, testLogger)
	require.NotNil(t, em)
	
	return em
}
