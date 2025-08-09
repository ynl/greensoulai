package async

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ynl/greensoulai/pkg/logger"
)

func TestAsyncExecutor_BasicExecution(t *testing.T) {
	logger := logger.NewTestLogger()
	executor := NewAsyncExecutor(2, logger)
	defer executor.Stop()

	// 测试基本异步执行
	resultChan := executor.ExecuteAsync(context.Background(), func() (interface{}, error) {
		time.Sleep(10 * time.Millisecond)
		return "test result", nil
	})

	select {
	case result := <-resultChan:
		if result.Error != nil {
			t.Fatalf("expected no error, got %v", result.Error)
		}
		if result.Value != "test result" {
			t.Errorf("expected 'test result', got '%v'", result.Value)
		}
		if result.Duration <= 0 {
			t.Errorf("expected positive duration, got %v", result.Duration)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("execution timeout")
	}
}

func TestAsyncExecutor_ErrorHandling(t *testing.T) {
	logger := logger.NewTestLogger()
	executor := NewAsyncExecutor(2, logger)
	defer executor.Stop()

	testError := errors.New("test error")

	// 测试错误处理
	resultChan := executor.ExecuteAsync(context.Background(), func() (interface{}, error) {
		return nil, testError
	})

	select {
	case result := <-resultChan:
		if result.Error == nil {
			t.Fatal("expected error, got nil")
		}
		if result.Error.Error() != testError.Error() {
			t.Errorf("expected '%v', got '%v'", testError, result.Error)
		}
		if result.Value != nil {
			t.Errorf("expected nil value, got %v", result.Value)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("execution timeout")
	}
}

func TestAsyncExecutor_TimeoutExecution(t *testing.T) {
	logger := logger.NewTestLogger()
	executor := NewAsyncExecutor(2, logger)
	defer executor.Stop()

	// 测试超时执行
	resultChan := executor.ExecuteWithTimeout(context.Background(), func() (interface{}, error) {
		time.Sleep(100 * time.Millisecond)
		return "delayed result", nil
	}, 50*time.Millisecond)

	select {
	case result := <-resultChan:
		if result.Error == nil {
			t.Fatal("expected timeout error, got nil")
		}
		// 修复：接受两种可能的错误消息
		if result.Error.Error() != "context deadline exceeded" && result.Error.Error() != "context canceled" {
			t.Errorf("expected 'context deadline exceeded' or 'context canceled', got '%v'", result.Error)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("execution timeout")
	}
}

func TestAsyncExecutor_ConcurrentExecution(t *testing.T) {
	logger := logger.NewTestLogger()
	executor := NewAsyncExecutor(4, logger)
	defer executor.Stop()

	// 测试并发执行
	const numTasks = 10
	results := make(chan Result, numTasks)

	for i := 0; i < numTasks; i++ {
		taskID := i
		resultChan := executor.ExecuteAsync(context.Background(), func() (interface{}, error) {
			time.Sleep(10 * time.Millisecond)
			return taskID, nil
		})

		go func() {
			result := <-resultChan
			results <- result
		}()
	}

	// 收集所有结果
	completedTasks := 0
	for i := 0; i < numTasks; i++ {
		select {
		case result := <-results:
			if result.Error != nil {
				t.Errorf("task %d failed: %v", i, result.Error)
			}
			if result.Value == nil {
				t.Errorf("task %d returned nil value", i)
			}
			completedTasks++
		case <-time.After(2 * time.Second):
			t.Fatal("concurrent execution timeout")
		}
	}

	if completedTasks != numTasks {
		t.Errorf("expected %d completed tasks, got %d", numTasks, completedTasks)
	}
}

func TestAsyncExecutor_Stats(t *testing.T) {
	logger := logger.NewTestLogger()
	executor := NewAsyncExecutor(3, logger)
	defer executor.Stop()

	stats := executor.GetStats()
	if stats["max_workers"] != 3 {
		t.Errorf("expected max_workers to be 3, got %v", stats["max_workers"])
	}

	if stats["queue_size"] == nil {
		t.Error("expected queue_size to be present")
	}

	if stats["active_workers"] == nil {
		t.Error("expected active_workers to be present")
	}
}

func TestTaskOutput_Creation(t *testing.T) {
	output := NewTaskOutput("test raw", "test agent", "test description")

	if output.Raw != "test raw" {
		t.Errorf("expected 'test raw', got '%s'", output.Raw)
	}

	if output.Agent != "test agent" {
		t.Errorf("expected 'test agent', got '%s'", output.Agent)
	}

	if output.Description != "test description" {
		t.Errorf("expected 'test description', got '%s'", output.Description)
	}

	if output.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	if output.Metadata == nil {
		t.Error("expected Metadata to be initialized")
	}
}
