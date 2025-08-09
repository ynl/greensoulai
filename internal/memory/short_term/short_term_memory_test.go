package short_term

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ynl/greensoulai/internal/memory"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

func TestNewShortTermMemory(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)

	// 创建基本配置
	embedderConfig := &memory.EmbedderConfig{
		Provider: "test",
		Config:   map[string]interface{}{"model": "test-model"},
	}

	// 创建存储（这里需要一个mock或真实的存储实现）
	// 由于我们没有具体的存储实现，这个测试主要验证构造函数不会panic
	crew := &struct{ name string }{name: "test-crew"}
	collectionName := "test-collection"

	// 这里会因为需要具体的存储实现而可能失败，但至少验证了类型正确性
	stm := NewShortTermMemory(crew, embedderConfig, nil, collectionName, testEventBus, testLogger)

	// 基本验证
	assert.NotNil(t, stm)
}

// 注意：由于ShortTermMemory依赖于具体的存储实现，
// 完整的功能测试需要在有了存储实现后再添加
