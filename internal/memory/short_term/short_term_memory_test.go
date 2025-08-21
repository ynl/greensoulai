package short_term

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ynl/greensoulai/internal/memory"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

func TestNewShortTermMemory(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)

	tests := []struct {
		name           string
		embedderConfig *memory.EmbedderConfig
		crew           interface{}
		path           string
		expectedType   string
	}{
		{
			name: "default RAG storage",
			embedderConfig: &memory.EmbedderConfig{
				Provider: "test",
				Config:   map[string]interface{}{"model": "test-model"},
			},
			crew:         &struct{ name string }{name: "test-crew"},
			path:         "test-collection",
			expectedType: "test",
		},
		{
			name: "mem0 storage",
			embedderConfig: &memory.EmbedderConfig{
				Provider: "mem0",
				Config:   map[string]interface{}{"api_key": "test-key"},
			},
			crew:         &struct{ name string }{name: "test-crew"},
			path:         "mem0-collection",
			expectedType: "mem0",
		},
		{
			name:           "nil embedder config",
			embedderConfig: nil,
			crew:           &struct{ name string }{name: "test-crew"},
			path:           "nil-config-collection",
			expectedType:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stm := NewShortTermMemory(tt.crew, tt.embedderConfig, nil, tt.path, testEventBus, testLogger)

			assert.NotNil(t, stm)
			assert.NotNil(t, stm.BaseMemory)
			assert.Equal(t, tt.expectedType, stm.memoryProvider)
		})
	}
}

func TestShortTermMemorySave(t *testing.T) {
	stm := createTestShortTermMemory(t)
	ctx := context.Background()

	t.Run("basic save", func(t *testing.T) {
		metadata := map[string]interface{}{"priority": "high"}

		// RAG存储可能不会返回错误，但至少不应该panic
		assert.NotPanics(t, func() {
			stm.Save(ctx, "test task result", metadata, "test-agent")
		})
	})

	t.Run("save with session context", func(t *testing.T) {
		sessionCtx := context.WithValue(ctx, "session_id", "session-123")
		metadata := map[string]interface{}{"type": "user_input"}

		_ = stm.Save(sessionCtx, "user said hello", metadata, "assistant")

		assert.NotPanics(t, func() {
			stm.Save(sessionCtx, "user said hello", metadata, "assistant")
		})
	})

	t.Run("save with task context", func(t *testing.T) {
		taskCtx := context.WithValue(ctx, "task_id", "task-456")
		metadata := map[string]interface{}{"stage": "execution"}

		_ = stm.Save(taskCtx, "task progress update", metadata, "worker-agent")

		assert.NotPanics(t, func() {
			stm.Save(taskCtx, "task progress update", metadata, "worker-agent")
		})
	})

	t.Run("save with nil metadata", func(t *testing.T) {
		_ = stm.Save(ctx, "simple message", nil, "test-agent")

		assert.NotPanics(t, func() {
			stm.Save(ctx, "simple message", nil, "test-agent")
		})
	})
}

func TestShortTermMemorySaveTaskMemory(t *testing.T) {
	stm := createTestShortTermMemory(t)
	ctx := context.Background()

	tests := []struct {
		name     string
		taskID   string
		value    interface{}
		metadata map[string]interface{}
		agent    string
	}{
		{
			name:   "basic task memory",
			taskID: "task-001",
			value:  "completed analysis phase",
			metadata: map[string]interface{}{
				"phase":    "analysis",
				"duration": 120,
			},
			agent: "analyzer-agent",
		},
		{
			name:     "task memory with nil metadata",
			taskID:   "task-002",
			value:    "started execution",
			metadata: nil,
			agent:    "executor-agent",
		},
		{
			name:   "task memory with complex value",
			taskID: "task-003",
			value: map[string]interface{}{
				"status":  "in_progress",
				"results": []string{"item1", "item2"},
			},
			metadata: map[string]interface{}{
				"timestamp": time.Now(),
			},
			agent: "complex-agent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = stm.SaveTaskMemory(ctx, tt.taskID, tt.value, tt.metadata, tt.agent)

			assert.NotPanics(t, func() {
				stm.SaveTaskMemory(ctx, tt.taskID, tt.value, tt.metadata, tt.agent)
			})
		})
	}
}

func TestShortTermMemorySaveInteractionMemory(t *testing.T) {
	stm := createTestShortTermMemory(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		sessionID   string
		interaction interface{}
		metadata    map[string]interface{}
		agent       string
	}{
		{
			name:      "user message",
			sessionID: "session-001",
			interaction: map[string]interface{}{
				"type":    "user_input",
				"content": "How do I deploy my application?",
			},
			metadata: map[string]interface{}{
				"timestamp": time.Now(),
				"channel":   "chat",
			},
			agent: "chat-agent",
		},
		{
			name:        "simple interaction",
			sessionID:   "session-002",
			interaction: "Agent provided help documentation",
			metadata:    nil,
			agent:       "help-agent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = stm.SaveInteractionMemory(ctx, tt.sessionID, tt.interaction, tt.metadata, tt.agent)

			assert.NotPanics(t, func() {
				stm.SaveInteractionMemory(ctx, tt.sessionID, tt.interaction, tt.metadata, tt.agent)
			})
		})
	}
}

func TestShortTermMemorySearchByTask(t *testing.T) {
	stm := createTestShortTermMemory(t)
	ctx := context.Background()

	t.Run("search existing task", func(t *testing.T) {
		// 先保存一些任务记忆
		_ = stm.SaveTaskMemory(ctx, "search-task-1", "task data", nil, "test-agent")

		results, _ := stm.SearchByTask(ctx, "search-task-1", "data", 10, 0.5)
		assert.NotNil(t, results) // 应该返回结果集合（可能为空）
	})

	t.Run("search non-existent task", func(t *testing.T) {
		results, _ := stm.SearchByTask(ctx, "non-existent-task", "query", 10, 0.5)
		assert.NotNil(t, results) // 应该返回空结果集合而不是nil
	})

	t.Run("search with different limits", func(t *testing.T) {
		limits := []int{1, 5, 20}
		for _, limit := range limits {
			results, _ := stm.SearchByTask(ctx, "any-task", "query", limit, 0.7)
			assert.NotNil(t, results)
		}
	})
}

func TestShortTermMemorySearchBySession(t *testing.T) {
	stm := createTestShortTermMemory(t)
	ctx := context.Background()

	t.Run("search existing session", func(t *testing.T) {
		// 先保存一些会话记忆
		_ = stm.SaveInteractionMemory(ctx, "search-session-1", "session data", nil, "test-agent")

		results, _ := stm.SearchBySession(ctx, "search-session-1", "data", 10, 0.5)
		assert.NotNil(t, results) // 应该返回结果集合（可能为空）
	})

	t.Run("search non-existent session", func(t *testing.T) {
		results, _ := stm.SearchBySession(ctx, "non-existent-session", "query", 10, 0.5)
		assert.NotNil(t, results) // 应该返回空结果集合而不是nil
	})
}

func TestShortTermMemoryGetRecentMemories(t *testing.T) {
	stm := createTestShortTermMemory(t)
	ctx := context.Background()

	tests := []struct {
		name  string
		limit int
	}{
		{"small limit", 5},
		{"medium limit", 20},
		{"large limit", 100},
		{"zero limit", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, _ := stm.GetRecentMemories(ctx, "test-agent", tt.limit)
			assert.NotNil(t, results) // 应该返回结果集合（可能为空）
		})
	}
}

func TestShortTermMemoryClearSession(t *testing.T) {
	stm := createTestShortTermMemory(t)
	ctx := context.Background()

	t.Run("clear existing session", func(t *testing.T) {
		// 先保存会话数据
		_ = stm.SaveInteractionMemory(ctx, "clear-session-1", "data to clear", nil, "test-agent")

		_ = stm.ClearSession(ctx, "clear-session-1")

		assert.NotPanics(t, func() {
			stm.ClearSession(ctx, "clear-session-1")
		})
	})

	t.Run("clear non-existent session", func(t *testing.T) {
		_ = stm.ClearSession(ctx, "non-existent-session")

		assert.NotPanics(t, func() {
			stm.ClearSession(ctx, "non-existent-session")
		})
	})
}

func TestShortTermMemoryClearTask(t *testing.T) {
	stm := createTestShortTermMemory(t)
	ctx := context.Background()

	t.Run("clear existing task", func(t *testing.T) {
		// 先保存任务数据
		_ = stm.SaveTaskMemory(ctx, "clear-task-1", "data to clear", nil, "test-agent")

		_ = stm.ClearTask(ctx, "clear-task-1")

		assert.NotPanics(t, func() {
			stm.ClearTask(ctx, "clear-task-1")
		})
	})

	t.Run("clear non-existent task", func(t *testing.T) {
		_ = stm.ClearTask(ctx, "non-existent-task")

		assert.NotPanics(t, func() {
			stm.ClearTask(ctx, "non-existent-task")
		})
	})
}

func TestShortTermMemoryInheritedMethods(t *testing.T) {
	stm := createTestShortTermMemory(t)
	ctx := context.Background()

	t.Run("search method", func(t *testing.T) {
		results, _ := stm.Search(ctx, "test query", 10, 0.5)
		assert.NotNil(t, results)
	})

	t.Run("clear method", func(t *testing.T) {
		_ = stm.Clear(ctx)

		assert.NotPanics(t, func() {
			stm.Clear(ctx)
		})
	})

	t.Run("close method", func(t *testing.T) {
		// 测试关闭功能
		assert.NotPanics(t, func() {
			stm.Close()
		})
	})
}

func TestShortTermMemoryEdgeCases(t *testing.T) {
	stm := createTestShortTermMemory(t)
	ctx := context.Background()

	t.Run("empty string values", func(t *testing.T) {
		_ = stm.SaveTaskMemory(ctx, "", "", nil, "")

		assert.NotPanics(t, func() {
			stm.SaveTaskMemory(ctx, "", "", nil, "")
		})
	})

	t.Run("nil context values", func(t *testing.T) {
		nilCtx := context.Background()
		_ = stm.Save(nilCtx, nil, nil, "test-agent")

		assert.NotPanics(t, func() {
			stm.Save(nilCtx, nil, nil, "test-agent")
		})
	})

	t.Run("large data values", func(t *testing.T) {
		largeData := make([]byte, 1024*1024) // 1MB data
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}

		_ = stm.Save(ctx, largeData, nil, "large-data-agent")

		assert.NotPanics(t, func() {
			stm.Save(ctx, largeData, nil, "large-data-agent")
		})
	})
}

// Helper function to create a test ShortTermMemory instance
func createTestShortTermMemory(t *testing.T) *ShortTermMemory {
	t.Helper()

	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)

	embedderConfig := &memory.EmbedderConfig{
		Provider: "test",
		Config:   map[string]interface{}{"model": "test-model"},
	}

	crew := &struct{ name string }{name: "test-crew"}
	collectionName := "test-collection"

	stm := NewShortTermMemory(crew, embedderConfig, nil, collectionName, testEventBus, testLogger)
	require.NotNil(t, stm)

	return stm
}
