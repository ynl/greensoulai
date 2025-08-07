package events

import (
	"context"
)

// On 装饰器函数，用于注册事件处理器
// 用法类似于crewAI的 @crewai_event_bus.on(AgentExecutionCompletedEvent)
func On(eventType string) func(EventHandler) EventHandler {
	return func(handler EventHandler) EventHandler {
		// 这里需要全局事件总线实例
		// 在实际使用中，可以通过依赖注入或全局变量获取
		return handler
	}
}

// GlobalEventBus 全局事件总线实例
var GlobalEventBus EventBus

// SetGlobalEventBus 设置全局事件总线
func SetGlobalEventBus(bus EventBus) {
	GlobalEventBus = bus
}

// GetGlobalEventBus 获取全局事件总线
func GetGlobalEventBus() EventBus {
	return GlobalEventBus
}

// OnWithBus 带事件总线的装饰器
func OnWithBus(bus EventBus, eventType string) func(EventHandler) EventHandler {
	return func(handler EventHandler) EventHandler {
		if bus != nil {
			bus.RegisterHandler(eventType, handler)
		}
		return handler
	}
}

// EventListener 事件监听器接口，匹配crewAI的BaseEventListener
type EventListener interface {
	SetupListeners(bus EventBus) error
	GetName() string
}

// BaseEventListener 基础事件监听器实现
type BaseEventListener struct {
	name string
	bus  EventBus
}

// NewBaseEventListener 创建基础事件监听器
func NewBaseEventListener(name string) *BaseEventListener {
	return &BaseEventListener{
		name: name,
	}
}

// SetupListeners 设置监听器（子类需要重写）
func (bel *BaseEventListener) SetupListeners(bus EventBus) error {
	bel.bus = bus
	return nil
}

// GetName 获取监听器名称
func (bel *BaseEventListener) GetName() string {
	return bel.name
}

// GetBus 获取事件总线
func (bel *BaseEventListener) GetBus() EventBus {
	return bel.bus
}

// RegisterListener 注册事件监听器
func RegisterListener(listener EventListener, bus EventBus) error {
	return listener.SetupListeners(bus)
}

// 示例：控制台日志监听器
type ConsoleLogListener struct {
	*BaseEventListener
}

// NewConsoleLogListener 创建控制台日志监听器
func NewConsoleLogListener() *ConsoleLogListener {
	return &ConsoleLogListener{
		BaseEventListener: NewBaseEventListener("console_log"),
	}
}

// SetupListeners 设置控制台日志监听器
func (cl *ConsoleLogListener) SetupListeners(bus EventBus) error {
	cl.BaseEventListener.SetupListeners(bus)

	// 注册各种事件处理器
	events := []string{
		EventTypeAgentStarted,
		EventTypeAgentCompleted,
		EventTypeAgentError,
		EventTypeTaskStarted,
		EventTypeTaskCompleted,
		EventTypeTaskFailed,
		EventTypeCrewStarted,
		EventTypeCrewCompleted,
		EventTypeCrewFailed,
		EventTypeLLMCallStarted,
		EventTypeLLMCallCompleted,
		EventTypeLLMCallFailed,
		EventTypeToolUsageStarted,
		EventTypeToolUsageFinished,
		EventTypeToolUsageError,
	}

	for _, eventType := range events {
		bus.RegisterHandler(eventType, cl.handleEvent)
	}

	return nil
}

// handleEvent 处理事件
func (cl *ConsoleLogListener) handleEvent(ctx context.Context, event Event) error {
	// 这里可以实现具体的日志记录逻辑
	// 例如：写入文件、发送到监控系统等
	return nil
}

// 示例：性能监控监听器
type PerformanceMonitorListener struct {
	*BaseEventListener
	metrics map[string]interface{}
}

// NewPerformanceMonitorListener 创建性能监控监听器
func NewPerformanceMonitorListener() *PerformanceMonitorListener {
	return &PerformanceMonitorListener{
		BaseEventListener: NewBaseEventListener("performance_monitor"),
		metrics:           make(map[string]interface{}),
	}
}

// SetupListeners 设置性能监控监听器
func (pm *PerformanceMonitorListener) SetupListeners(bus EventBus) error {
	pm.BaseEventListener.SetupListeners(bus)

	// 注册性能相关事件
	performanceEvents := []string{
		EventTypeAgentStarted,
		EventTypeAgentCompleted,
		EventTypeTaskStarted,
		EventTypeTaskCompleted,
		EventTypeLLMCallStarted,
		EventTypeLLMCallCompleted,
		EventTypeToolUsageStarted,
		EventTypeToolUsageFinished,
	}

	for _, eventType := range performanceEvents {
		bus.RegisterHandler(eventType, pm.handlePerformanceEvent)
	}

	return nil
}

// handlePerformanceEvent 处理性能事件
func (pm *PerformanceMonitorListener) handlePerformanceEvent(ctx context.Context, event Event) error {
	// 收集性能指标
	eventType := event.GetType()

	// 这里可以实现具体的性能监控逻辑
	// 例如：记录执行时间、统计成功率等
	_ = eventType // 使用变量避免linter警告

	return nil
}

// GetMetrics 获取性能指标
func (pm *PerformanceMonitorListener) GetMetrics() map[string]interface{} {
	return pm.metrics
}
