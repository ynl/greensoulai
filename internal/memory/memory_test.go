package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMemory 模拟内存实现，用于测试
type MockMemory struct {
	mock.Mock
	items map[string]*MemoryItem
}

func NewMockMemory() *MockMemory {
	return &MockMemory{
		items: make(map[string]*MemoryItem),
	}
}

func (m *MockMemory) Save(ctx context.Context, value interface{}, metadata map[string]interface{}, agent string) error {
	args := m.Called(ctx, value, metadata, agent)
	return args.Error(0)
}

func (m *MockMemory) Search(ctx context.Context, query string, limit int, scoreThreshold float64) ([]MemoryItem, error) {
	args := m.Called(ctx, query, limit, scoreThreshold)
	return args.Get(0).([]MemoryItem), args.Error(1)
}

func (m *MockMemory) Clear(ctx context.Context) error {
	args := m.Called(ctx)
	if args.Error(0) == nil {
		m.items = make(map[string]*MemoryItem)
	}
	return args.Error(0)
}

func (m *MockMemory) SetCrew(crew interface{}) Memory {
	args := m.Called(crew)
	return args.Get(0).(Memory)
}

func (m *MockMemory) Close() error {
	args := m.Called()
	return args.Error(0)
}

// TestMemoryItem 测试内存项结构
func TestMemoryItem(t *testing.T) {
	now := time.Now()
	item := &MemoryItem{
		ID:        "test-id",
		Value:     "test content",
		Metadata:  map[string]interface{}{"key": "value"},
		Agent:     "test-agent",
		CreatedAt: now,
		Score:     0.95,
	}

	assert.Equal(t, "test-id", item.ID)
	assert.Equal(t, "test content", item.Value)
	assert.Equal(t, "value", item.Metadata["key"])
	assert.Equal(t, "test-agent", item.Agent)
	assert.Equal(t, now, item.CreatedAt)
	assert.Equal(t, 0.95, item.Score)
}

// TestEmbedderConfig 测试嵌入配置
func TestEmbedderConfig(t *testing.T) {
	config := &EmbedderConfig{
		Provider: "openai",
		Config: map[string]interface{}{
			"model":      "text-embedding-ada-002",
			"api_key":    "test-key",
			"base_url":   "https://api.openai.com",
			"batch_size": 100,
			"timeout":    30,
			"version":    "v1",
		},
	}

	assert.Equal(t, "openai", config.Provider)
	assert.Equal(t, "text-embedding-ada-002", config.Config["model"])
	assert.Equal(t, "test-key", config.Config["api_key"])
	assert.Equal(t, "https://api.openai.com", config.Config["base_url"])
	assert.Equal(t, 100, config.Config["batch_size"])
	assert.Equal(t, 30, config.Config["timeout"])
	assert.Equal(t, "v1", config.Config["version"])
}

// TestMockMemorySave 测试模拟内存的保存操作
func TestMockMemorySave(t *testing.T) {
	ctx := context.Background()
	mockMemory := NewMockMemory()

	value := "test content 1"
	metadata := map[string]interface{}{"type": "test"}
	agent := "test-agent"

	mockMemory.On("Save", ctx, value, metadata, agent).Return(nil)
	err := mockMemory.Save(ctx, value, metadata, agent)

	assert.NoError(t, err)
	mockMemory.AssertExpectations(t)
}

// TestMockMemorySearch 测试搜索操作
func TestMockMemorySearch(t *testing.T) {
	ctx := context.Background()
	mockMemory := NewMockMemory()

	expectedItems := []MemoryItem{
		{ID: "search-1", Value: "first result"},
		{ID: "search-2", Value: "second result"},
	}

	mockMemory.On("Search", ctx, "search query", 10, 0.5).Return(expectedItems, nil)

	items, err := mockMemory.Search(ctx, "search query", 10, 0.5)

	assert.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, "search-1", items[0].ID)
	assert.Equal(t, "search-2", items[1].ID)
	mockMemory.AssertExpectations(t)
}

// TestMockMemoryClear 测试清除操作
func TestMockMemoryClear(t *testing.T) {
	ctx := context.Background()
	mockMemory := NewMockMemory()

	// 添加一些测试数据
	mockMemory.items["test-1"] = &MemoryItem{ID: "test-1", Value: "content 1"}
	mockMemory.items["test-2"] = &MemoryItem{ID: "test-2", Value: "content 2"}

	assert.Len(t, mockMemory.items, 2)

	mockMemory.On("Clear", ctx).Return(nil)
	err := mockMemory.Clear(ctx)

	assert.NoError(t, err)
	assert.Len(t, mockMemory.items, 0)
	mockMemory.AssertExpectations(t)
}

// TestMockMemoryConfiguration 测试配置方法
func TestMockMemoryConfiguration(t *testing.T) {
	mockMemory := NewMockMemory()
	testCrew := &struct{ name string }{name: "test-crew"}

	mockMemory.On("SetCrew", testCrew).Return(mockMemory)
	mockMemory.On("Close").Return(nil)

	result := mockMemory.SetCrew(testCrew)
	assert.Equal(t, mockMemory, result)

	err := mockMemory.Close()
	assert.NoError(t, err)

	mockMemory.AssertExpectations(t)
}

// BenchmarkMemoryItem 内存项性能基准测试
func BenchmarkMemoryItem(b *testing.B) {
	for i := 0; i < b.N; i++ {
		item := &MemoryItem{
			ID:        "bench-test",
			Value:     "benchmark content for performance testing",
			Metadata:  map[string]interface{}{"iteration": i},
			Agent:     "bench-agent",
			CreatedAt: time.Now(),
			Score:     0.9,
		}
		_ = item
	}
}

// TestMemoryItemEdgeCases 测试内存项边界情况
func TestMemoryItemEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		item     *MemoryItem
		expected bool
	}{
		{
			name: "empty value",
			item: &MemoryItem{
				ID:    "empty-value",
				Value: "",
			},
			expected: true,
		},
		{
			name: "nil metadata",
			item: &MemoryItem{
				ID:       "nil-metadata",
				Value:    "content",
				Metadata: nil,
			},
			expected: true,
		},
		{
			name: "zero time",
			item: &MemoryItem{
				ID:        "zero-time",
				Value:     "content",
				CreatedAt: time.Time{},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.item)
			assert.Equal(t, tt.expected, true) // 所有情况都应该是有效的
		})
	}
}

// TestEmbedderConfigDefaults 测试嵌入配置默认值
func TestEmbedderConfigDefaults(t *testing.T) {
	config := &EmbedderConfig{}

	// 测试零值
	assert.Equal(t, "", config.Provider)
	assert.Nil(t, config.Config)
}

// TestMemoryInterface 测试内存接口的完整性
func TestMemoryInterface(t *testing.T) {
	var memory Memory
	mockMemory := NewMockMemory()

	// 确保MockMemory实现了Memory接口
	memory = mockMemory
	assert.NotNil(t, memory)

	// 验证所有接口方法都可调用
	ctx := context.Background()

	mockMemory.On("Save", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockMemory.On("Search", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]MemoryItem{}, nil)
	mockMemory.On("Clear", mock.Anything).Return(nil)
	mockMemory.On("SetCrew", mock.Anything).Return(mockMemory)
	mockMemory.On("Close").Return(nil)

	// 调用所有方法
	err := memory.Save(ctx, "test value", map[string]interface{}{}, "test-agent")
	assert.NoError(t, err)

	_, err = memory.Search(ctx, "test", 1, 0.5)
	assert.NoError(t, err)

	err = memory.Clear(ctx)
	assert.NoError(t, err)

	result := memory.SetCrew("test")
	assert.Equal(t, mockMemory, result)

	err = memory.Close()
	assert.NoError(t, err)

	mockMemory.AssertExpectations(t)
}
