// 新版本Workflow系统的测试套件
// 使用Job而非Task，避免与Agent系统的Task冲突
package flow

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

// ============================================================================
// 基础功能测试
// ============================================================================

func TestNewWorkflow(t *testing.T) {
	workflow := NewWorkflow("test-workflow")
	if workflow == nil {
		t.Fatal("Workflow should not be nil")
	}
}

func TestSimpleJob(t *testing.T) {
	ctx := context.Background()

	job := NewJob("hello", func(ctx context.Context) (interface{}, error) {
		return "world", nil
	})

	workflow := NewWorkflow("simple-test").
		AddJob(job, Immediately())

	result, err := workflow.Run(ctx)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.FinalResult != "world" {
		t.Errorf("Expected 'world', got: %v", result.FinalResult)
	}

	if len(result.JobTrace) != 1 {
		t.Errorf("Expected 1 job trace, got: %d", len(result.JobTrace))
	}

	if result.JobTrace[0].JobID != "hello" {
		t.Errorf("Expected job ID 'hello', got: %s", result.JobTrace[0].JobID)
	}
}

func TestJobError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("test error")

	job := NewJob("error-job", func(ctx context.Context) (interface{}, error) {
		return nil, expectedErr
	})

	workflow := NewWorkflow("error-test").
		AddJob(job, Immediately())

	result, err := workflow.Run(ctx)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if result.Error == nil {
		t.Fatal("Expected result.Error to be set")
	}
}

// ============================================================================
// 触发器条件测试
// ============================================================================

func TestAfterTrigger(t *testing.T) {
	ctx := context.Background()

	job1 := NewJob("first", func(ctx context.Context) (interface{}, error) {
		return "first-result", nil
	})

	job2 := NewJob("second", func(ctx context.Context) (interface{}, error) {
		return "second-result", nil
	})

	workflow := NewWorkflow("after-test").
		AddJob(job1, Immediately()).
		AddJob(job2, After("first"))

	result, err := workflow.Run(ctx)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result.JobTrace) != 2 {
		t.Fatalf("Expected 2 jobs, got: %d", len(result.JobTrace))
	}

	// 验证执行顺序 - 第一个作业应该是first
	if result.JobTrace[0].JobID != "first" {
		t.Errorf("Expected first job to be 'first', got: %s", result.JobTrace[0].JobID)
	}

	if result.JobTrace[1].JobID != "second" {
		t.Errorf("Expected second job to be 'second', got: %s", result.JobTrace[1].JobID)
	}
}

func TestAllOfTrigger(t *testing.T) {
	ctx := context.Background()

	jobA := NewJob("a", func(ctx context.Context) (interface{}, error) {
		return "a-result", nil
	})

	jobB := NewJob("b", func(ctx context.Context) (interface{}, error) {
		return "b-result", nil
	})

	jobC := NewJob("c", func(ctx context.Context) (interface{}, error) {
		return "c-result", nil
	})

	workflow := NewWorkflow("allof-test").
		AddJob(jobA, Immediately()).
		AddJob(jobB, Immediately()).
		AddJob(jobC, AllOf(After("a"), After("b")))

	result, err := workflow.Run(ctx)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result.JobTrace) != 3 {
		t.Fatalf("Expected 3 jobs, got: %d", len(result.JobTrace))
	}

	// 验证最后一个作业是c（等待a和b都完成）
	lastJob := result.JobTrace[len(result.JobTrace)-1]
	if lastJob.JobID != "c" {
		t.Errorf("Expected last job to be 'c', got: %s", lastJob.JobID)
	}
}

func TestAnyOfTrigger(t *testing.T) {
	ctx := context.Background()

	trigger := NewJob("trigger", func(ctx context.Context) (interface{}, error) {
		return "trigger-result", nil
	})

	listener := NewJob("listener", func(ctx context.Context) (interface{}, error) {
		return "listener-result", nil
	})

	workflow := NewWorkflow("anyof-test").
		AddJob(trigger, Immediately()).
		AddJob(listener, AnyOf(After("trigger"), After("non-existent")))

	result, err := workflow.Run(ctx)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result.JobTrace) != 2 {
		t.Fatalf("Expected 2 jobs, got: %d", len(result.JobTrace))
	}
}

// ============================================================================
// 并行执行测试
// ============================================================================

func TestParallelExecution(t *testing.T) {
	ctx := context.Background()

	var counter int64

	// 创建多个可以并行执行的作业
	createJob := func(name string) Job {
		return NewJob(name, func(ctx context.Context) (interface{}, error) {
			atomic.AddInt64(&counter, 1)
			time.Sleep(50 * time.Millisecond) // 模拟工作
			return name + "-result", nil
		})
	}

	workflow := NewWorkflow("parallel-test").
		AddJob(createJob("job1"), Immediately()).
		AddJob(createJob("job2"), Immediately()).
		AddJob(createJob("job3"), Immediately())

	start := time.Now()
	result, err := workflow.Run(ctx)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 验证并行执行（应该比串行快很多）
	if duration > 100*time.Millisecond {
		t.Errorf("Expected parallel execution to be faster, took: %v", duration)
	}

	if len(result.JobTrace) != 3 {
		t.Fatalf("Expected 3 jobs, got: %d", len(result.JobTrace))
	}

	// 验证所有作业都执行了
	if atomic.LoadInt64(&counter) != 3 {
		t.Errorf("Expected counter to be 3, got: %d", counter)
	}

	// 验证并发指标
	if result.Metrics.MaxConcurrency != 3 {
		t.Errorf("Expected max concurrency to be 3, got: %d", result.Metrics.MaxConcurrency)
	}
}

func TestMultipleListeners(t *testing.T) {
	ctx := context.Background()

	var listenerCount int64

	triggerJob := NewJob("trigger", func(ctx context.Context) (interface{}, error) {
		return "triggered", nil
	})

	createListener := func(name string) Job {
		return NewJob(name, func(ctx context.Context) (interface{}, error) {
			atomic.AddInt64(&listenerCount, 1)
			return name + "-done", nil
		})
	}

	workflow := NewWorkflow("multi-listener-test").
		AddJob(triggerJob, Immediately()).
		AddJob(createListener("listener1"), After("trigger")).
		AddJob(createListener("listener2"), After("trigger")).
		AddJob(createListener("listener3"), After("trigger"))

	result, err := workflow.Run(ctx)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result.JobTrace) != 4 {
		t.Fatalf("Expected 4 jobs, got: %d", len(result.JobTrace))
	}

	// 验证所有监听器都被触发
	if atomic.LoadInt64(&listenerCount) != 3 {
		t.Errorf("Expected 3 listeners to execute, got: %d", listenerCount)
	}

	// 验证有两个批次：第一个批次是trigger，第二个批次是3个listener
	if result.Metrics.ParallelBatches != 2 {
		t.Errorf("Expected 2 parallel batches, got: %d", result.Metrics.ParallelBatches)
	}
}

// ============================================================================
// 异步执行测试
// ============================================================================

func TestAsyncExecution(t *testing.T) {
	ctx := context.Background()

	job := NewJob("async-job", func(ctx context.Context) (interface{}, error) {
		time.Sleep(100 * time.Millisecond)
		return "async-result", nil
	})

	workflow := NewWorkflow("async-test").
		AddJob(job, Immediately())

	start := time.Now()
	resultChan := workflow.RunAsync(ctx)

	// 应该立即返回
	if time.Since(start) > 10*time.Millisecond {
		t.Errorf("RunAsync should return immediately")
	}

	// 等待结果
	select {
	case result := <-resultChan:
		if result.Error != nil {
			t.Fatalf("Unexpected error: %v", result.Error)
		}
		if result.FinalResult != "async-result" {
			t.Errorf("Expected 'async-result', got: %v", result.FinalResult)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Async execution timed out")
	}
}

// ============================================================================
// 组合作业测试
// ============================================================================

func TestParallelJobGroup(t *testing.T) {
	ctx := context.Background()

	job1 := NewJob("sub1", func(ctx context.Context) (interface{}, error) {
		return "result1", nil
	})

	job2 := NewJob("sub2", func(ctx context.Context) (interface{}, error) {
		return "result2", nil
	})

	parallelGroup := NewParallelGroup("parallel-group", job1, job2)

	workflow := NewWorkflow("parallel-group-test").
		AddJob(parallelGroup, Immediately())

	result, err := workflow.Run(ctx)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result.JobTrace) != 1 {
		t.Fatalf("Expected 1 job trace, got: %d", len(result.JobTrace))
	}

	// 验证并行结果
	results, ok := result.FinalResult.([]interface{})
	if !ok {
		t.Fatal("Expected parallel results to be []interface{}")
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 parallel results, got: %d", len(results))
	}
}

func TestSequentialJobChain(t *testing.T) {
	ctx := context.Background()

	job1 := NewJob("seq1", func(ctx context.Context) (interface{}, error) {
		return "first", nil
	})

	job2 := NewJob("seq2", func(ctx context.Context) (interface{}, error) {
		return "second", nil
	})

	sequentialChain := NewSequentialChain("seq-chain", job1, job2)

	workflow := NewWorkflow("sequential-chain-test").
		AddJob(sequentialChain, Immediately())

	result, err := workflow.Run(ctx)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 应该返回最后一个作业的结果
	if result.FinalResult != "second" {
		t.Errorf("Expected 'second', got: %v", result.FinalResult)
	}
}

// ============================================================================
// 便捷函数测试
// ============================================================================

func TestAfterJobs(t *testing.T) {
	completed := JobResults{
		"job1": "result1",
		"job2": "result2",
	}

	trigger := AfterJobs("job1", "job2")

	if !trigger.Ready(completed) {
		t.Error("AfterJobs should be ready when all jobs are completed")
	}

	// 删除一个作业
	delete(completed, "job2")

	if trigger.Ready(completed) {
		t.Error("AfterJobs should not be ready when not all jobs are completed")
	}
}

func TestAfterAnyJob(t *testing.T) {
	completed := JobResults{
		"job1": "result1",
	}

	trigger := AfterAnyJob("job1", "job2")

	if !trigger.Ready(completed) {
		t.Error("AfterAnyJob should be ready when at least one job is completed")
	}

	// 清空
	completed = make(JobResults)

	if trigger.Ready(completed) {
		t.Error("AfterAnyJob should not be ready when no jobs are completed")
	}
}

// ============================================================================
// Agent集成测试
// ============================================================================

func TestAgentJobAdapter(t *testing.T) {
	ctx := context.Background()

	agentJob := NewAgentJob("agent-task-1", "Analyze sentiment of text")

	workflow := NewWorkflow("agent-integration-test").
		AddJob(agentJob, Immediately())

	result, err := workflow.Run(ctx)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 验证Agent作业执行
	if result.FinalResult == nil {
		t.Error("Expected agent job to return a result")
	}

	// 验证结果包含预期信息
	resultStr, ok := result.FinalResult.(string)
	if !ok {
		t.Fatal("Expected agent job result to be string")
	}

	if !contains(resultStr, "agent-task-1") {
		t.Errorf("Expected result to contain job ID, got: %s", resultStr)
	}
}

// ============================================================================
// 边界条件测试
// ============================================================================

func TestEmptyWorkflow(t *testing.T) {
	ctx := context.Background()

	workflow := NewWorkflow("empty-workflow")
	result, err := workflow.Run(ctx)

	if err != nil {
		t.Fatalf("Empty workflow should not error: %v", err)
	}

	if len(result.JobTrace) != 0 {
		t.Errorf("Expected 0 jobs, got: %d", len(result.JobTrace))
	}
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	job := NewJob("slow-job", func(ctx context.Context) (interface{}, error) {
		select {
		case <-time.After(1 * time.Second):
			return "should not reach here", nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	})

	workflow := NewWorkflow("cancellation-test").
		AddJob(job, Immediately())

	// 在短时间后取消
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	result, err := workflow.Run(ctx)

	if err == nil {
		t.Fatal("Expected cancellation error")
	}

	if result.Error == nil {
		t.Fatal("Expected result.Error to be set")
	}
}

// ============================================================================
// 性能基准测试
// ============================================================================

func BenchmarkSimpleWorkflow(b *testing.B) {
	ctx := context.Background()

	job := NewJob("bench-job", func(ctx context.Context) (interface{}, error) {
		return "result", nil
	})

	workflow := NewWorkflow("benchmark").AddJob(job, Immediately())

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := workflow.Run(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParallelWorkflow(b *testing.B) {
	ctx := context.Background()

	workflow := NewWorkflow("parallel-benchmark")

	for i := 0; i < 10; i++ {
		jobName := fmt.Sprintf("job%d", i)
		job := NewJob(jobName, func(ctx context.Context) (interface{}, error) {
			return "result", nil
		})
		workflow.AddJob(job, Immediately())
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := workflow.Run(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ============================================================================
// 状态传递测试
// ============================================================================

func TestFlowStateBasicOperations(t *testing.T) {
	state := NewFlowState()

	// 基础设置和获取
	state.Set("key1", "value1")

	if value, exists := state.Get("key1"); !exists || value != "value1" {
		t.Errorf("Expected value1, got %v, exists: %v", value, exists)
	}

	// 类型安全获取
	if str, exists := state.GetString("key1"); !exists || str != "value1" {
		t.Errorf("Expected value1, got %s, exists: %v", str, exists)
	}

	// 键列表
	keys := state.Keys()
	if len(keys) != 1 || keys[0] != "key1" {
		t.Errorf("Expected keys [key1], got %v", keys)
	}
}

func TestStatefulJobExecution(t *testing.T) {
	ctx := context.Background()

	// 创建支持状态的作业
	setJob := NewStatefulJob("setter", func(ctx context.Context, state FlowState) (interface{}, error) {
		state.Set("shared_data", "from_setter")
		return "set_complete", nil
	})

	getJob := NewStatefulJob("getter", func(ctx context.Context, state FlowState) (interface{}, error) {
		if data, exists := state.GetString("shared_data"); exists {
			return "got_" + data, nil
		}
		return nil, fmt.Errorf("shared_data not found")
	})

	workflow := NewWorkflow("stateful-test").
		AddJob(setJob, Immediately()).
		AddJob(getJob, After("setter"))

	result, err := workflow.Run(ctx)

	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	if result.FinalResult != "got_from_setter" {
		t.Errorf("Expected 'got_from_setter', got %v", result.FinalResult)
	}

	// 验证最终状态
	if data, exists := result.FinalState.GetString("shared_data"); !exists || data != "from_setter" {
		t.Errorf("Expected final state to contain 'from_setter', got %s, exists: %v", data, exists)
	}
}

func TestMixedJobTypes(t *testing.T) {
	ctx := context.Background()

	// 普通作业
	regularJob := NewJob("regular", func(ctx context.Context) (interface{}, error) {
		return "regular_result", nil
	})

	// 支持状态的作业
	statefulJob := NewStatefulJob("stateful", func(ctx context.Context, state FlowState) (interface{}, error) {
		state.Set("stateful_data", "from_stateful")
		return "stateful_result", nil
	})

	workflow := NewWorkflow("mixed-test").
		AddJob(regularJob, Immediately()).
		AddJob(statefulJob, After("regular"))

	result, err := workflow.Run(ctx)

	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	// 验证两种作业都能正常执行
	if len(result.AllResults) != 2 {
		t.Errorf("Expected 2 job results, got %d", len(result.AllResults))
	}

	// 验证状态只包含StatefulJob设置的数据
	if data, exists := result.FinalState.GetString("stateful_data"); !exists || data != "from_stateful" {
		t.Errorf("Expected stateful state data, got %s, exists: %v", data, exists)
	}
}

// ============================================================================
// 工具函数
// ============================================================================

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (len(substr) == 0 || func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}())
}
