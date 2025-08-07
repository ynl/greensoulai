package events

import (
	"context"
	"testing"
	"time"

	"github.com/ynl/greensoulai/pkg/logger"
)

func TestBasic(t *testing.T) {
	// 基础测试，验证包可以正常导入和运行
	t.Log("Events package is working correctly")
}

func TestEventBus_BasicOperations(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := NewEventBus(logger)

	// 测试事件发射和订阅
	received := make(chan bool, 1)

	err := eventBus.Subscribe("test_event", func(ctx context.Context, event Event) error {
		received <- true
		return nil
	})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	event := &BaseEvent{
		Type:      "test_event",
		Timestamp: time.Now(),
	}

	err = eventBus.Emit(context.Background(), nil, event)
	if err != nil {
		t.Fatalf("failed to emit event: %v", err)
	}

	select {
	case <-received:
		// 成功
	case <-time.After(1 * time.Second):
		t.Fatal("event not received")
	}
}

func TestEventBus_MultipleHandlers(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := NewEventBus(logger)

	receivedCount := 0

	// 注册多个处理器
	for i := 0; i < 3; i++ {
		err := eventBus.Subscribe("multi_event", func(ctx context.Context, event Event) error {
			receivedCount++
			return nil
		})
		if err != nil {
			t.Fatalf("failed to subscribe: %v", err)
		}
	}

	event := &BaseEvent{
		Type:      "multi_event",
		Timestamp: time.Now(),
	}

	err := eventBus.Emit(context.Background(), nil, event)
	if err != nil {
		t.Fatalf("failed to emit event: %v", err)
	}

	// 等待所有处理器执行
	time.Sleep(100 * time.Millisecond)

	if receivedCount != 3 {
		t.Errorf("expected 3 handlers to be called, got %d", receivedCount)
	}
}

func TestEventBus_Unsubscribe(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := NewEventBus(logger)

	received := make(chan bool, 1)

	handler := func(ctx context.Context, event Event) error {
		received <- true
		return nil
	}

	// 订阅
	err := eventBus.Subscribe("unsub_event", handler)
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	// 取消订阅
	err = eventBus.Unsubscribe("unsub_event", handler)
	if err != nil {
		t.Fatalf("failed to unsubscribe: %v", err)
	}

	// 发射事件，应该不会被接收
	event := &BaseEvent{
		Type:      "unsub_event",
		Timestamp: time.Now(),
	}

	err = eventBus.Emit(context.Background(), nil, event)
	if err != nil {
		t.Fatalf("failed to emit event: %v", err)
	}

	select {
	case <-received:
		t.Fatal("event received after unsubscribe")
	case <-time.After(100 * time.Millisecond):
		// 成功，没有接收到事件
	}
}

func TestEventBus_GetStats(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := NewEventBus(logger)

	// 注册一些处理器
	for i := 0; i < 2; i++ {
		err := eventBus.Subscribe("stats_event", func(ctx context.Context, event Event) error {
			return nil
		})
		if err != nil {
			t.Fatalf("failed to subscribe: %v", err)
		}
	}

	// 检查统计信息
	if eventBus.GetHandlerCount("stats_event") != 2 {
		t.Errorf("expected 2 handlers, got %d", eventBus.GetHandlerCount("stats_event"))
	}

	if eventBus.GetHandlerCount("nonexistent_event") != 0 {
		t.Errorf("expected 0 handlers, got %d", eventBus.GetHandlerCount("nonexistent_event"))
	}

	registeredTypes := eventBus.GetRegisteredEventTypes()
	if len(registeredTypes) != 1 {
		t.Errorf("expected 1 registered type, got %d", len(registeredTypes))
	}

	if registeredTypes[0] != "stats_event" {
		t.Errorf("expected 'stats_event', got '%s'", registeredTypes[0])
	}
}
func TestBaseEvent_Interface(t *testing.T) {
	event := &BaseEvent{
		Type:      "test_event",
		Timestamp: time.Now(),
		Source:    "test_source",
		Payload:   map[string]interface{}{"key": "value"},
	}

	// 测试接口实现
	if event.GetType() != "test_event" {
		t.Errorf("expected 'test_event', got '%s'", event.GetType())
	}

	if event.GetSource() != "test_source" {
		t.Errorf("expected 'test_source', got '%v'", event.GetSource())
	}

	payload := event.GetPayload()
	if payload["key"] != "value" {
		t.Errorf("expected 'value', got '%v'", payload["key"])
	}
}

func TestEventBus_WithFingerprint(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := NewEventBus(logger)

	// 测试带指纹的事件
	received := make(chan bool, 1)

	err := eventBus.Subscribe(EventTypeAgentStarted, func(ctx context.Context, event Event) error {
		// 验证指纹信息
		if event.GetSourceFingerprint() != "test-fingerprint" {
			t.Errorf("expected fingerprint 'test-fingerprint', got '%s'", event.GetSourceFingerprint())
		}
		if event.GetSourceType() != "agent" {
			t.Errorf("expected source type 'agent', got '%s'", event.GetSourceType())
		}
		received <- true
		return nil
	})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	event := &AgentExecutionStartedEvent{
		BaseEvent: BaseEvent{
			Type:              EventTypeAgentStarted,
			Timestamp:         time.Now(),
			SourceFingerprint: "test-fingerprint",
			SourceType:        "agent",
			FingerprintMetadata: map[string]interface{}{
				"role": "test-agent",
			},
		},
		Agent:      "test-agent",
		Task:       "test-task",
		TaskPrompt: "test prompt",
	}

	err = eventBus.Emit(context.Background(), nil, event)
	if err != nil {
		t.Fatalf("failed to emit event: %v", err)
	}

	select {
	case <-received:
		// 成功
	case <-time.After(1 * time.Second):
		t.Fatal("event not received")
	}
}

func TestEventBus_ScopedHandlers(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := NewEventBus(logger)

	// 注册全局处理器
	globalReceived := make(chan bool, 1)
	err := eventBus.Subscribe("test_event", func(ctx context.Context, event Event) error {
		globalReceived <- true
		return nil
	})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	// 创建作用域事件总线
	scopedBus := eventBus.WithScopedHandlers()

	// 注册作用域处理器
	scopedReceived := make(chan bool, 1)
	err = scopedBus.Subscribe("test_event", func(ctx context.Context, event Event) error {
		scopedReceived <- true
		return nil
	})
	if err != nil {
		t.Fatalf("failed to subscribe to scoped bus: %v", err)
	}

	// 通过作用域总线发射事件
	event := &BaseEvent{
		Type:      "test_event",
		Timestamp: time.Now(),
	}

	err = scopedBus.Emit(context.Background(), nil, event)
	if err != nil {
		t.Fatalf("failed to emit event: %v", err)
	}

	// 验证只有作用域处理器被调用
	select {
	case <-scopedReceived:
		// 成功
	case <-time.After(1 * time.Second):
		t.Fatal("scoped event not received")
	}

	select {
	case <-globalReceived:
		t.Fatal("global event should not be received")
	case <-time.After(100 * time.Millisecond):
		// 成功，全局处理器不应该被调用
	}
}

func TestEventBus_RegisterHandler(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := NewEventBus(logger)

	// 测试RegisterHandler方法（别名）
	received := make(chan bool, 1)

	err := eventBus.RegisterHandler("test_event", func(ctx context.Context, event Event) error {
		received <- true
		return nil
	})
	if err != nil {
		t.Fatalf("failed to register handler: %v", err)
	}

	event := &BaseEvent{
		Type:      "test_event",
		Timestamp: time.Now(),
	}

	err = eventBus.Emit(context.Background(), nil, event)
	if err != nil {
		t.Fatalf("failed to emit event: %v", err)
	}

	select {
	case <-received:
		// 成功
	case <-time.After(1 * time.Second):
		t.Fatal("event not received")
	}
}

func TestEventBus_NewEventTypes(t *testing.T) {
	logger := logger.NewTestLogger()
	eventBus := NewEventBus(logger)

	// 测试新的事件类型
	testCases := []struct {
		eventType string
		event     Event
	}{
		{
			eventType: EventTypeAgentError,
			event: &AgentExecutionErrorEvent{
				BaseEvent: BaseEvent{
					Type:      EventTypeAgentError,
					Timestamp: time.Now(),
				},
				Agent: "test-agent",
				Task:  "test-task",
				Error: "test error",
			},
		},
		{
			eventType: EventTypeTaskFailed,
			event: &TaskFailedEvent{
				BaseEvent: BaseEvent{
					Type:      EventTypeTaskFailed,
					Timestamp: time.Now(),
				},
				Task:  "test-task",
				Agent: "test-agent",
				Error: "task failed",
			},
		},
		{
			eventType: EventTypeLLMCallFailed,
			event: &LLMCallFailedEvent{
				BaseEvent: BaseEvent{
					Type:      EventTypeLLMCallFailed,
					Timestamp: time.Now(),
				},
				Model: "gpt-4",
				Error: "API error",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.eventType, func(t *testing.T) {
			received := make(chan bool, 1)

			err := eventBus.Subscribe(tc.eventType, func(ctx context.Context, event Event) error {
				received <- true
				return nil
			})
			if err != nil {
				t.Fatalf("failed to subscribe: %v", err)
			}

			err = eventBus.Emit(context.Background(), nil, tc.event)
			if err != nil {
				t.Fatalf("failed to emit event: %v", err)
			}

			select {
			case <-received:
				// 成功
			case <-time.After(1 * time.Second):
				t.Fatal("event not received")
			}
		})
	}
}
