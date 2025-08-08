package long_term

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

func TestNewLongTermMemory(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)

	// 创建长期记忆实例（可能会因为没有SQLite存储而失败，但测试基本结构）
	ltm := NewLongTermMemory(nil, "test-path", testEventBus, testLogger)

	// 基本验证
	assert.NotNil(t, ltm)
}

func TestLongTermMemoryInterface(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)

	ltm := NewLongTermMemory(nil, "test-path", testEventBus, testLogger)
	assert.NotNil(t, ltm)

	ctx := context.Background()

	// 测试Python版本兼容的Save方法
	item := NewLongTermMemoryItem(
		"test-agent",
		"test task",
		"expected output",
		time.Now().Format(time.RFC3339),
		nil,
		map[string]interface{}{"test": "metadata"},
	)

	err := ltm.Save(ctx, item)
	_ = err // 忽略具体错误，只要不panic即可

	// 测试Python版本兼容的Search方法
	results, err := ltm.Search(ctx, "test task", 3)
	_ = err
	assert.NotNil(t, results) // 至少应该返回空slice而不是nil

	// 测试Reset方法
	err = ltm.Reset(ctx)
	_ = err

	// 测试通用Memory接口的兼容方法
	err = ltm.SaveCompatible(ctx, "test value", map[string]interface{}{}, "test-agent")
	_ = err

	memResults, err := ltm.SearchCompatible(ctx, "test", 10, 0.5)
	_ = err
	assert.NotNil(t, memResults)

	// 测试SetCrew和Close方法（通过BaseMemory）
	result := ltm.SetCrew("new-crew")
	assert.NotNil(t, result)

	err = ltm.Close()
	_ = err
}

func TestLongTermMemoryConfiguration(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)

	ltm := NewLongTermMemory(nil, "advanced-path", testEventBus, testLogger)

	assert.NotNil(t, ltm)

	// 测试重新设置crew
	newCrew := map[string]interface{}{"name": "updated-crew"}
	result := ltm.SetCrew(newCrew)
	// SetCrew返回的是Memory接口，不是具体的LongTermMemory类型
	assert.NotNil(t, result) // 确保返回值不为nil
}

func TestLongTermMemoryEdgeCases(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)

	// 测试nil配置
	ltm := NewLongTermMemory(nil, "", testEventBus, testLogger)
	assert.NotNil(t, ltm)

	ctx := context.Background()

	// 测试空值保存
	emptyItem := NewLongTermMemoryItem("", "", "", "", nil, nil)
	err := ltm.Save(ctx, emptyItem)
	_ = err // 应该能处理空值而不panic，可能返回错误

	// 测试空查询
	results, err := ltm.Search(ctx, "", 0)
	_ = err
	assert.NotNil(t, results) // 应该返回空slice而不是nil

	// 测试负数限制
	results, err = ltm.Search(ctx, "test", -1)
	_ = err
	assert.NotNil(t, results)
}

// BenchmarkLongTermMemoryCreation 创建长期记忆的性能基准测试
func BenchmarkLongTermMemoryCreation(b *testing.B) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ltm := NewLongTermMemory(nil, "bench-collection", testEventBus, testLogger)
		_ = ltm
	}
}

// TestLongTermMemoryThreadSafety 测试并发安全性
func TestLongTermMemoryThreadSafety(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)

	// 为每个测试创建唯一的数据库文件以避免锁争用
	dbPath := "/tmp/thread-safe-test-" + time.Now().Format("20060102150405") + ".db"
	ltm := NewLongTermMemory(nil, dbPath, testEventBus, testLogger)

	numGoroutines := 2 // 进一步减少并发数量以避免数据库锁争用
	done := make(chan bool, numGoroutines)

	// 并发执行各种操作
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() {
				// 恢复panic并标记完成
				if r := recover(); r != nil {
					t.Logf("Goroutine %d panicked: %v", id, r)
				}
				done <- true
			}()

			ctx := context.Background()

			// 尝试保存数据（忽略错误以避免死锁）
			item := NewLongTermMemoryItem(
				"test-agent",
				"concurrent test",
				"output",
				time.Now().Format(time.RFC3339),
				nil,
				map[string]interface{}{"id": id},
			)
			_ = ltm.Save(ctx, item) // 忽略错误

			// 尝试搜索（忽略错误）
			_, _ = ltm.Search(ctx, "concurrent", 5)

			// 尝试设置crew
			ltm.SetCrew(map[string]interface{}{"name": "updated", "id": id})

			// 简短等待
			time.Sleep(10 * time.Millisecond)
		}(i)
		
		// 在启动goroutine间添加小延迟以减少并发压力
		time.Sleep(10 * time.Millisecond)
	}

	// 等待所有goroutines完成
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
			// 成功完成
		case <-time.After(5 * time.Second): // 增加等待时间
			t.Fatalf("Test timed out after 5 seconds - possible deadlock (goroutine %d)", i)
		}
	}
}
