package contextual

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ynl/greensoulai/internal/memory"
	"github.com/ynl/greensoulai/internal/memory/entity"
	"github.com/ynl/greensoulai/internal/memory/external"
	"github.com/ynl/greensoulai/internal/memory/long_term"
	"github.com/ynl/greensoulai/internal/memory/short_term"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// IntegrationTestContextualMemory 集成测试
// 使用真实的记忆实例测试ContextualMemory的功能
func TestContextualMemory_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	ctx := context.Background()
	logger := logger.NewConsoleLogger()
	eventBus := events.NewEventBus(logger)

	// 创建真实的记忆实例
	config := &memory.EmbedderConfig{
		Provider: "test",
		Config: map[string]interface{}{
			"model": "test-model",
		},
	}

	// 创建各类记忆实例
	stm := short_term.NewShortTermMemory(nil, config, nil, "", eventBus, logger)
	ltm := long_term.NewLongTermMemory(nil, "", eventBus, logger)
	em := entity.NewEntityMemory(nil, config, nil, "", eventBus, logger)
	exm := external.NewExternalMemory(nil, config, nil, "", eventBus, logger)

	// 创建ContextualMemory
	cm := NewContextualMemory(stm, ltm, em, exm, eventBus, logger, nil)

	// 创建mock任务
	mockTask := &MockTask{}
	mockTask.On("GetDescription").Return("集成测试任务")
	mockTask.On("GetID").Return("integration-task-1")

	// 测试BuildContextForTask
	result, err := cm.BuildContextForTask(ctx, mockTask, "集成测试上下文")

	// 验证结果
	assert.NoError(t, err)
	// 在空的记忆实例下，结果应该为空
	assert.Equal(t, "", result)

	mockTask.AssertExpectations(t)
}

// TestContextualMemory_WithRealData 使用真实数据的集成测试
func TestContextualMemory_WithRealData(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	ctx := context.Background()
	logger := logger.NewConsoleLogger()
	eventBus := events.NewEventBus(logger)

	// 创建真实的记忆实例
	config := &memory.EmbedderConfig{
		Provider: "test",
		Config: map[string]interface{}{
			"model": "test-model",
		},
	}

	stm := short_term.NewShortTermMemory(nil, config, nil, "", eventBus, logger)
	ltm := long_term.NewLongTermMemory(nil, "", eventBus, logger)
	em := entity.NewEntityMemory(nil, config, nil, "", eventBus, logger)
	exm := external.NewExternalMemory(nil, config, nil, "", eventBus, logger)

	// 创建ContextualMemory
	cm := NewContextualMemory(stm, ltm, em, exm, eventBus, logger, nil)

	// 向短期记忆添加测试数据
	err := stm.Save(ctx, "测试洞察1", map[string]interface{}{
		"context": "这是第一个测试洞察",
		"type":    "insight",
	}, "test-agent")
	assert.NoError(t, err)

	// 向实体记忆添加测试数据
	err = em.Save(ctx, "测试实体", map[string]interface{}{
		"context": "这是一个测试实体信息",
		"type":    "entity",
	}, "test-agent")
	assert.NoError(t, err)

	// 创建mock任务
	mockTask := &MockTask{}
	mockTask.On("GetDescription").Return("数据集成测试任务")
	mockTask.On("GetID").Return("data-integration-task-1")

	// 测试BuildContextForTask - 注意：由于使用的是测试embeder配置，可能无法找到相似内容
	result, err := cm.BuildContextForTask(ctx, mockTask, "集成测试")

	// 验证结果
	assert.NoError(t, err)
	// 根据实际的记忆实现，可能会返回相关内容或空内容
	t.Logf("集成测试结果长度: %d", len(result))
	t.Logf("集成测试结果: %s", result)

	mockTask.AssertExpectations(t)
}

// TestContextualMemory_EventEmission 测试事件发射的集成
func TestContextualMemory_EventEmission(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	ctx := context.Background()
	logger := logger.NewConsoleLogger()
	eventBus := events.NewEventBus(logger)

	// 设置事件监听器（使用通道来同步事件接收）
	eventChan := make(chan events.Event, 10)
	eventHandler := func(ctx context.Context, event events.Event) error {
		eventChan <- event
		return nil
	}

	err := eventBus.Subscribe(EventTypeContextBuildStarted, eventHandler)
	assert.NoError(t, err)

	err = eventBus.Subscribe(EventTypeContextBuildCompleted, eventHandler)
	assert.NoError(t, err)

	// 创建ContextualMemory
	cm := NewContextualMemory(nil, nil, nil, nil, eventBus, logger, nil)

	// 创建mock任务
	mockTask := &MockTask{}
	mockTask.On("GetDescription").Return("事件测试任务")
	mockTask.On("GetID").Return("event-test-task-1")

	// 执行BuildContextForTask
	_, err = cm.BuildContextForTask(ctx, mockTask, "事件测试")
	assert.NoError(t, err)

	// 收集事件（使用超时机制）
	var receivedEvents []events.Event
	timeout := time.After(5 * time.Second)

	for i := 0; i < 2; i++ {
		select {
		case event := <-eventChan:
			receivedEvents = append(receivedEvents, event)
		case <-timeout:
			t.Fatalf("超时等待事件，已收到 %d 个事件", len(receivedEvents))
		}
	}

	// 验证事件发射
	assert.Len(t, receivedEvents, 2)
	assert.Equal(t, EventTypeContextBuildStarted, receivedEvents[0].GetType())
	assert.Equal(t, EventTypeContextBuildCompleted, receivedEvents[1].GetType())

	mockTask.AssertExpectations(t)
}

// TestContextualMemory_ConfigurationIntegration 配置集成测试
func TestContextualMemory_ConfigurationIntegration(t *testing.T) {
	ctx := context.Background()
	logger := logger.NewConsoleLogger()
	eventBus := events.NewEventBus(logger)

	// 创建自定义配置
	customConfig := &ContextualMemoryConfig{
		DefaultSTMLimit:        10,
		DefaultLTMLimit:        5,
		DefaultEntityLimit:     8,
		DefaultExternalLimit:   6,
		STMScoreThreshold:      0.2,
		EntityScoreThreshold:   0.3,
		ExternalScoreThreshold: 0.4,
		EnableFormatting:       false,
		EnableSectionHeaders:   false,
		MaxContextLength:       1000,
		FilterEmptyResults:     false,
		EnableDeduplication:    false,
	}

	// 创建ContextualMemory
	cm := NewContextualMemory(nil, nil, nil, nil, eventBus, logger, customConfig)

	// 验证配置设置正确
	config := cm.GetConfig()
	assert.Equal(t, customConfig.DefaultSTMLimit, config.DefaultSTMLimit)
	assert.Equal(t, customConfig.DefaultLTMLimit, config.DefaultLTMLimit)
	assert.Equal(t, customConfig.DefaultEntityLimit, config.DefaultEntityLimit)
	assert.Equal(t, customConfig.DefaultExternalLimit, config.DefaultExternalLimit)
	assert.Equal(t, customConfig.STMScoreThreshold, config.STMScoreThreshold)
	assert.Equal(t, customConfig.EntityScoreThreshold, config.EntityScoreThreshold)
	assert.Equal(t, customConfig.ExternalScoreThreshold, config.ExternalScoreThreshold)
	assert.Equal(t, customConfig.EnableFormatting, config.EnableFormatting)
	assert.Equal(t, customConfig.EnableSectionHeaders, config.EnableSectionHeaders)
	assert.Equal(t, customConfig.MaxContextLength, config.MaxContextLength)
	assert.Equal(t, customConfig.FilterEmptyResults, config.FilterEmptyResults)
	assert.Equal(t, customConfig.EnableDeduplication, config.EnableDeduplication)

	// 创建mock任务
	mockTask := &MockTask{}
	mockTask.On("GetDescription").Return("配置测试任务")
	mockTask.On("GetID").Return("config-test-task-1")

	// 测试配置影响
	result, err := cm.BuildContextForTask(ctx, mockTask, "配置测试")
	assert.NoError(t, err)
	assert.Equal(t, "", result) // 由于记忆实例为nil，结果应该为空

	mockTask.AssertExpectations(t)
}

// TestContextualMemory_MemoryInstanceAccess 记忆实例访问测试
func TestContextualMemory_MemoryInstanceAccess(t *testing.T) {
	logger := logger.NewConsoleLogger()
	eventBus := events.NewEventBus(logger)

	config := &memory.EmbedderConfig{
		Provider: "test",
		Config: map[string]interface{}{
			"model": "test-model",
		},
	}

	// 创建真实的记忆实例
	stm := short_term.NewShortTermMemory(nil, config, nil, "", eventBus, logger)
	ltm := long_term.NewLongTermMemory(nil, "", eventBus, logger)
	em := entity.NewEntityMemory(nil, config, nil, "", eventBus, logger)
	exm := external.NewExternalMemory(nil, config, nil, "", eventBus, logger)

	// 创建ContextualMemory
	cm := NewContextualMemory(stm, ltm, em, exm, eventBus, logger, nil)

	// 验证记忆实例访问
	returnedSTM, returnedLTM, returnedEM, returnedEXM := cm.GetMemoryInstances()

	assert.Equal(t, stm, returnedSTM)
	assert.Equal(t, ltm, returnedLTM)
	assert.Equal(t, em, returnedEM)
	assert.Equal(t, exm, returnedEXM)
}

// BenchmarkContextualMemory_IntegrationBenchmark 集成性能基准测试
func BenchmarkContextualMemory_IntegrationBenchmark(b *testing.B) {
	ctx := context.Background()
	logger := logger.NewConsoleLogger()
	eventBus := events.NewEventBus(logger)

	config := &memory.EmbedderConfig{
		Provider: "test",
		Config: map[string]interface{}{
			"model": "test-model",
		},
	}

	// 创建真实的记忆实例
	stm := short_term.NewShortTermMemory(nil, config, nil, "", eventBus, logger)
	ltm := long_term.NewLongTermMemory(nil, "", eventBus, logger)
	em := entity.NewEntityMemory(nil, config, nil, "", eventBus, logger)
	exm := external.NewExternalMemory(nil, config, nil, "", eventBus, logger)

	cm := NewContextualMemory(stm, ltm, em, exm, eventBus, logger, nil)

	// 创建mock任务
	mockTask := &MockTask{}
	mockTask.On("GetDescription").Return("性能测试任务")
	mockTask.On("GetID").Return("perf-test-task")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cm.BuildContextForTask(ctx, mockTask, "性能测试上下文")
		if err != nil {
			b.Fatalf("基准测试失败: %v", err)
		}
	}
}

// TestContextualMemory_ErrorHandling 错误处理集成测试
func TestContextualMemory_ErrorHandling(t *testing.T) {
	logger := logger.NewConsoleLogger()

	// 使用nil eventBus测试错误处理
	cm := NewContextualMemory(nil, nil, nil, nil, nil, logger, nil)

	mockTask := &MockTask{}
	mockTask.On("GetDescription").Return("错误处理测试任务")
	mockTask.On("GetID").Return("error-test-task-1")

	ctx := context.Background()

	// 应该能够处理nil eventBus而不崩溃
	result, err := cm.BuildContextForTask(ctx, mockTask, "错误处理测试")

	assert.NoError(t, err)
	assert.Equal(t, "", result)

	mockTask.AssertExpectations(t)
}
