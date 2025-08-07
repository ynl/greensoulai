package events

import (
	"context"
	"fmt"
	"sync"

	"github.com/ynl/greensoulai/pkg/logger"
)

// eventBus 事件总线实现
type eventBus struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
	logger   logger.Logger
}

// scopedEventBus 作用域事件总线，用于临时处理器管理
type scopedEventBus struct {
	*eventBus
	originalHandlers map[string][]EventHandler
}

// NewEventBus 创建新的事件总线
func NewEventBus(logger logger.Logger) EventBus {
	return &eventBus{
		handlers: make(map[string][]EventHandler),
		logger:   logger,
	}
}

// Emit 发射事件
func (eb *eventBus) Emit(ctx context.Context, source interface{}, event Event) error {
	eb.mu.RLock()
	handlers, exists := eb.handlers[event.GetType()]
	eb.mu.RUnlock()

	if !exists {
		return nil
	}

	eb.logger.Debug("emitting event",
		logger.Field{Key: "event_type", Value: event.GetType()},
		logger.Field{Key: "handler_count", Value: len(handlers)},
		logger.Field{Key: "source_fingerprint", Value: event.GetSourceFingerprint()},
		logger.Field{Key: "source_type", Value: event.GetSourceType()},
	)

	for _, handler := range handlers {
		go func(h EventHandler) {
			if err := h(ctx, event); err != nil {
				eb.logger.Error("event handler error",
					logger.Field{Key: "error", Value: err},
					logger.Field{Key: "event_type", Value: event.GetType()},
				)
			}
		}(handler)
	}

	return nil
}

// Subscribe 订阅事件
func (eb *eventBus) Subscribe(eventType string, handler EventHandler) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.handlers[eventType] == nil {
		eb.handlers[eventType] = make([]EventHandler, 0)
	}
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)

	eb.logger.Info("event handler registered",
		logger.Field{Key: "event_type", Value: eventType},
	)
	return nil
}

// RegisterHandler 注册事件处理器（别名方法，匹配crewAI API）
func (eb *eventBus) RegisterHandler(eventType string, handler EventHandler) error {
	return eb.Subscribe(eventType, handler)
}

// Unsubscribe 取消订阅
func (eb *eventBus) Unsubscribe(eventType string, handler EventHandler) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	handlers, exists := eb.handlers[eventType]
	if !exists {
		return fmt.Errorf("no handlers for event type: %s", eventType)
	}

	// 移除指定的处理器
	for i, h := range handlers {
		if fmt.Sprintf("%p", h) == fmt.Sprintf("%p", handler) {
			eb.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			eb.logger.Info("event handler unregistered",
				logger.Field{Key: "event_type", Value: eventType},
			)
			return nil
		}
	}

	return fmt.Errorf("handler not found for event type: %s", eventType)
}

// GetHandlerCount 获取指定事件类型的处理器数量
func (eb *eventBus) GetHandlerCount(eventType string) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	handlers, exists := eb.handlers[eventType]
	if !exists {
		return 0
	}
	return len(handlers)
}

// GetRegisteredEventTypes 获取已注册的事件类型
func (eb *eventBus) GetRegisteredEventTypes() []string {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	types := make([]string, 0, len(eb.handlers))
	for eventType := range eb.handlers {
		types = append(types, eventType)
	}
	return types
}

// WithScopedHandlers 创建作用域事件总线，用于临时处理器管理
func (eb *eventBus) WithScopedHandlers() EventBus {
	eb.mu.RLock()
	originalHandlers := make(map[string][]EventHandler)
	for k, v := range eb.handlers {
		originalHandlers[k] = make([]EventHandler, len(v))
		copy(originalHandlers[k], v)
	}
	eb.mu.RUnlock()

	scoped := &scopedEventBus{
		eventBus:         eb,
		originalHandlers: originalHandlers,
	}

	// 清空当前处理器
	eb.mu.Lock()
	eb.handlers = make(map[string][]EventHandler)
	eb.mu.Unlock()

	return scoped
}

// Close 关闭作用域事件总线，恢复原始处理器
func (seb *scopedEventBus) Close() {
	seb.mu.Lock()
	seb.handlers = seb.originalHandlers
	seb.mu.Unlock()
}

// 实现作用域事件总线的其他方法
func (seb *scopedEventBus) Subscribe(eventType string, handler EventHandler) error {
	seb.mu.Lock()
	defer seb.mu.Unlock()

	if seb.handlers[eventType] == nil {
		seb.handlers[eventType] = make([]EventHandler, 0)
	}
	seb.handlers[eventType] = append(seb.handlers[eventType], handler)

	seb.logger.Info("scoped event handler registered",
		logger.Field{Key: "event_type", Value: eventType},
	)
	return nil
}

func (seb *scopedEventBus) RegisterHandler(eventType string, handler EventHandler) error {
	return seb.Subscribe(eventType, handler)
}

func (seb *scopedEventBus) Unsubscribe(eventType string, handler EventHandler) error {
	seb.mu.Lock()
	defer seb.mu.Unlock()

	handlers, exists := seb.handlers[eventType]
	if !exists {
		return fmt.Errorf("no handlers for event type: %s", eventType)
	}

	// 移除指定的处理器
	for i, h := range handlers {
		if fmt.Sprintf("%p", h) == fmt.Sprintf("%p", handler) {
			seb.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			seb.logger.Info("scoped event handler unregistered",
				logger.Field{Key: "event_type", Value: eventType},
			)
			return nil
		}
	}

	return fmt.Errorf("handler not found for event type: %s", eventType)
}

func (seb *scopedEventBus) GetHandlerCount(eventType string) int {
	seb.mu.RLock()
	defer seb.mu.RUnlock()

	handlers, exists := seb.handlers[eventType]
	if !exists {
		return 0
	}
	return len(handlers)
}

func (seb *scopedEventBus) GetRegisteredEventTypes() []string {
	seb.mu.RLock()
	defer seb.mu.RUnlock()

	types := make([]string, 0, len(seb.handlers))
	for eventType := range seb.handlers {
		types = append(types, eventType)
	}
	return types
}

func (seb *scopedEventBus) Emit(ctx context.Context, source interface{}, event Event) error {
	seb.mu.RLock()
	handlers, exists := seb.handlers[event.GetType()]
	seb.mu.RUnlock()

	if !exists {
		return nil
	}

	seb.logger.Debug("emitting scoped event",
		logger.Field{Key: "event_type", Value: event.GetType()},
		logger.Field{Key: "handler_count", Value: len(handlers)},
		logger.Field{Key: "source_fingerprint", Value: event.GetSourceFingerprint()},
		logger.Field{Key: "source_type", Value: event.GetSourceType()},
	)

	for _, handler := range handlers {
		go func(h EventHandler) {
			if err := h(ctx, event); err != nil {
				seb.logger.Error("scoped event handler error",
					logger.Field{Key: "error", Value: err},
					logger.Field{Key: "event_type", Value: event.GetType()},
				)
			}
		}(handler)
	}

	return nil
}

func (seb *scopedEventBus) WithScopedHandlers() EventBus {
	return seb.eventBus.WithScopedHandlers()
}
