package async

import (
	"context"
	"sync"
	"time"

	"github.com/ynl/greensoulai/pkg/logger"
)

// AsyncExecutor 异步执行器接口
type AsyncExecutor interface {
	ExecuteAsync(ctx context.Context, task func() (interface{}, error)) <-chan Result
	ExecuteWithTimeout(ctx context.Context, task func() (interface{}, error), timeout time.Duration) <-chan Result
	Stop()
	GetStats() map[string]interface{}
}

// asyncExecutor 异步执行器实现
type asyncExecutor struct {
	maxWorkers int
	workQueue  chan work
	quit       chan bool
	wg         sync.WaitGroup
	logger     logger.Logger
}

// work 工作项
type work struct {
	task   func() (interface{}, error)
	result chan Result
}

// NewAsyncExecutor 创建新的异步执行器
func NewAsyncExecutor(maxWorkers int, logger logger.Logger) AsyncExecutor {
	ae := &asyncExecutor{
		maxWorkers: maxWorkers,
		workQueue:  make(chan work, maxWorkers*2),
		quit:       make(chan bool),
		logger:     logger,
	}

	ae.start()
	return ae
}

// start 启动工作池
func (ae *asyncExecutor) start() {
	for i := 0; i < ae.maxWorkers; i++ {
		ae.wg.Add(1)
		go ae.worker()
	}

	ae.logger.Info("async executor started",
		logger.Field{Key: "max_workers", Value: ae.maxWorkers},
	)
}

// worker 工作协程
func (ae *asyncExecutor) worker() {
	defer ae.wg.Done()

	for {
		select {
		case work := <-ae.workQueue:
			start := time.Now()
			result, err := work.task()
			duration := time.Since(start)

			work.result <- Result{
				Value:    result,
				Error:    err,
				Duration: duration,
			}

			ae.logger.Debug("task completed",
				logger.Field{Key: "duration_ms", Value: duration.Milliseconds()},
				logger.Field{Key: "error", Value: err},
			)
		case <-ae.quit:
			return
		}
	}
}

// ExecuteAsync 异步执行任务
func (ae *asyncExecutor) ExecuteAsync(ctx context.Context, task func() (interface{}, error)) <-chan Result {
	result := make(chan Result, 1)

	select {
	case ae.workQueue <- work{task: task, result: result}:
		return result
	case <-ctx.Done():
		go func() {
			result <- Result{Error: ctx.Err()}
		}()
		return result
	}
}

// ExecuteWithTimeout 带超时的异步执行
func (ae *asyncExecutor) ExecuteWithTimeout(ctx context.Context, task func() (interface{}, error), timeout time.Duration) <-chan Result {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resultChan := make(chan Result, 1)

	go func() {
		defer close(resultChan)

		select {
		case result := <-ae.ExecuteAsync(ctx, task):
			resultChan <- result
		case <-ctx.Done():
			resultChan <- Result{Error: ctx.Err()}
		}
	}()

	return resultChan
}

// Stop 停止执行器
func (ae *asyncExecutor) Stop() {
	ae.logger.Info("stopping async executor")
	close(ae.quit)
	ae.wg.Wait()
	ae.logger.Info("async executor stopped")
}

// GetStats 获取执行器统计信息
func (ae *asyncExecutor) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"max_workers":    ae.maxWorkers,
		"queue_size":     len(ae.workQueue),
		"active_workers": ae.maxWorkers,
	}
}
