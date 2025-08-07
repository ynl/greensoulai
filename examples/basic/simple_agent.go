package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// SimpleAgentExample 演示如何创建和使用一个简单的智能体
func main() {
	// 创建日志器
	log := logger.NewConsoleLogger()

	// 创建事件总线
	eventBus := events.NewEventBus(log)

	// 注册事件监听器
	err := eventBus.Subscribe(events.EventTypeAgentStarted, func(ctx context.Context, event events.Event) error {
		log.Info("智能体开始执行",
			logger.Field{Key: "event_type", Value: event.GetType()},
			logger.Field{Key: "timestamp", Value: event.GetTimestamp()},
		)
		return nil
	})
	if err != nil {
		log.Fatal("注册事件监听器失败", logger.Field{Key: "error", Value: err})
	}

	// 创建并发射一个简单的事件
	event := &events.AgentExecutionStartedEvent{
		BaseEvent: events.BaseEvent{
			Type:              events.EventTypeAgentStarted,
			Timestamp:         time.Now(),
			SourceFingerprint: "simple-agent-001",
			SourceType:        "agent",
		},
		Agent:      "SimpleAgent",
		Task:       "Hello World Task",
		TaskPrompt: "Say hello to the world",
	}

	ctx := context.Background()
	err = eventBus.Emit(ctx, nil, event)
	if err != nil {
		log.Fatal("发射事件失败", logger.Field{Key: "error", Value: err})
	}

	fmt.Println("简单智能体示例运行完成！")

	// 等待一点时间让异步事件处理完成
	time.Sleep(100 * time.Millisecond)
}
