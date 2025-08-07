// Package flow 提供并行作业编排系统
// 重新设计：使用Job而非Task，避免与Agent系统的Task冲突
package flow

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// 核心接口 - 使用Job避免与Agent Task冲突
// ============================================================================

// Job 定义工作流中的作业单元 - 可以并行执行
// 注意：Job与Agent的Task不同，Job是工作流编排单元，Task是Agent执行的业务任务
type Job interface {
	ID() string
	Execute(ctx context.Context) (interface{}, error)
}

// Trigger 定义作业触发条件
type Trigger interface {
	Ready(completed JobResults) bool
	String() string
}

// Workflow 定义并行作业编排接口
type Workflow interface {
	AddJob(job Job, trigger Trigger) Workflow
	Run(ctx context.Context) (*ExecutionResult, error)
	RunAsync(ctx context.Context) <-chan *ExecutionResult
}

// JobResults 已完成作业的结果集
type JobResults map[string]interface{}

// ============================================================================
// 执行结果和指标
// ============================================================================

// ExecutionResult 工作流执行结果
type ExecutionResult struct {
	FinalResult interface{}      // 最后完成的作业结果
	AllResults  JobResults       // 所有作业结果
	JobTrace    []JobExecution   // 作业执行追踪
	Metrics     *ParallelMetrics // 并行执行指标
	Duration    time.Duration    // 总执行时间
	Error       error            // 执行错误
}

// JobExecution 单个作业执行记录
type JobExecution struct {
	JobID     string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Result    interface{}
	Error     error
	BatchID   int // 所属的并行批次ID
}

// ParallelMetrics 并行执行指标
type ParallelMetrics struct {
	TotalJobs          int            // 总作业数
	ParallelBatches    int            // 并行批次数量
	BatchInfo          []BatchMetrics // 每个批次的详细信息
	MaxConcurrency     int            // 最大并发数
	ParallelEfficiency float64        // 并行效率
	SerialTime         time.Duration  // 假设串行执行的时间
	ParallelTime       time.Duration  // 实际并行执行时间
}

// BatchMetrics 批次执行指标
type BatchMetrics struct {
	BatchID        int
	JobCount       int
	Duration       time.Duration
	Concurrency    int
	EfficiencyGain float64 // 相对于串行的效率提升
}

// ============================================================================
// 核心实现：并行作业引擎
// ============================================================================

// ParallelEngine 并行作业执行引擎
// 注意：Engine设计用于编排和执行工作流作业，不同于Agent的任务执行
type ParallelEngine struct {
	name      string
	jobs      []jobWithTrigger
	maxCycles int
	mu        sync.RWMutex
}

type jobWithTrigger struct {
	job     Job
	trigger Trigger
}

// NewWorkflow 创建新的并行工作流
func NewWorkflow(name string) Workflow {
	return &ParallelEngine{
		name:      name,
		jobs:      make([]jobWithTrigger, 0),
		maxCycles: 100,
	}
}

// AddJob 添加作业和触发条件
func (e *ParallelEngine) AddJob(job Job, trigger Trigger) Workflow {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.jobs = append(e.jobs, jobWithTrigger{
		job:     job,
		trigger: trigger,
	})
	return e
}

// Run 执行工作流 - 重点：并行执行所有就绪的作业
func (e *ParallelEngine) Run(ctx context.Context) (*ExecutionResult, error) {
	startTime := time.Now()

	result := &ExecutionResult{
		AllResults: make(JobResults),
		JobTrace:   make([]JobExecution, 0),
		Metrics:    &ParallelMetrics{},
	}

	var totalSerialTime time.Duration
	cycle := 0
	batchID := 0

	for cycle < e.maxCycles {
		cycle++

		// 🚀 关键：获取所有就绪的作业（可能有多个）
		readyJobs := e.getReadyJobs(result.AllResults)
		if len(readyJobs) == 0 {
			break // 没有更多就绪的作业
		}

		// 🚀 关键：并行执行所有就绪的作业
		batchID++
		batchResults, batchMetrics, err := e.executeJobBatch(ctx, readyJobs, batchID)
		if err != nil {
			result.Error = err
			result.Duration = time.Since(startTime)
			return result, err
		}

		// 更新结果
		for _, jobExec := range batchResults {
			result.AllResults[jobExec.JobID] = jobExec.Result
			result.JobTrace = append(result.JobTrace, jobExec)
			result.FinalResult = jobExec.Result // 最后一个作为最终结果
			totalSerialTime += jobExec.Duration
		}

		// 更新批次指标
		result.Metrics.BatchInfo = append(result.Metrics.BatchInfo, batchMetrics)
		if batchMetrics.Concurrency > result.Metrics.MaxConcurrency {
			result.Metrics.MaxConcurrency = batchMetrics.Concurrency
		}
	}

	// 完善执行指标
	result.Duration = time.Since(startTime)
	result.Metrics.TotalJobs = len(result.JobTrace)
	result.Metrics.ParallelBatches = batchID
	result.Metrics.SerialTime = totalSerialTime
	result.Metrics.ParallelTime = result.Duration

	if result.Duration > 0 {
		result.Metrics.ParallelEfficiency = float64(totalSerialTime) / float64(result.Duration)
	}

	return result, nil
}

// RunAsync 异步执行工作流
func (e *ParallelEngine) RunAsync(ctx context.Context) <-chan *ExecutionResult {
	resultChan := make(chan *ExecutionResult, 1)

	go func() {
		defer close(resultChan)
		result, err := e.Run(ctx)
		if err != nil && result.Error == nil {
			result.Error = err
		}
		resultChan <- result
	}()

	return resultChan
}

// getReadyJobs 获取所有就绪的作业 - 强调可能有多个并行就绪
func (e *ParallelEngine) getReadyJobs(completed JobResults) []Job {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var ready []Job

	for _, jt := range e.jobs {
		// 跳过已完成的作业
		if _, done := completed[jt.job.ID()]; done {
			continue
		}

		// 检查触发条件
		if jt.trigger.Ready(completed) {
			ready = append(ready, jt.job)
		}
	}

	return ready
}

// executeJobBatch 批量并行执行作业 - 核心并行逻辑
func (e *ParallelEngine) executeJobBatch(ctx context.Context, jobs []Job, batchID int) ([]JobExecution, BatchMetrics, error) {
	if len(jobs) == 0 {
		return nil, BatchMetrics{}, nil
	}

	batchStartTime := time.Now()
	resultChan := make(chan JobExecution, len(jobs))
	var wg sync.WaitGroup

	// 🚀 关键：为每个作业启动独立的goroutine并行执行
	for _, job := range jobs {
		wg.Add(1)
		go func(j Job) {
			defer wg.Done()

			execution := JobExecution{
				JobID:     j.ID(),
				StartTime: time.Now(),
				BatchID:   batchID,
			}

			// 执行作业
			result, err := j.Execute(ctx)

			execution.EndTime = time.Now()
			execution.Duration = execution.EndTime.Sub(execution.StartTime)
			execution.Result = result
			execution.Error = err

			resultChan <- execution
		}(job)
	}

	// 等待所有作业完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	var executions []JobExecution
	for execution := range resultChan {
		if execution.Error != nil {
			return nil, BatchMetrics{}, fmt.Errorf("job %s failed: %w", execution.JobID, execution.Error)
		}
		executions = append(executions, execution)
	}

	// 计算批次指标
	batchDuration := time.Since(batchStartTime)
	totalJobTime := time.Duration(0)
	for _, exec := range executions {
		totalJobTime += exec.Duration
	}

	var efficiencyGain float64
	if batchDuration > 0 {
		efficiencyGain = float64(totalJobTime) / float64(batchDuration)
	}

	batchMetrics := BatchMetrics{
		BatchID:        batchID,
		JobCount:       len(jobs),
		Duration:       batchDuration,
		Concurrency:    len(jobs),
		EfficiencyGain: efficiencyGain,
	}

	return executions, batchMetrics, nil
}

// ============================================================================
// 作业实现 - 清晰表达这是并行执行单元
// ============================================================================

// SimpleJob 简单作业实现
type SimpleJob struct {
	id string
	fn func(ctx context.Context) (interface{}, error)
}

func (j SimpleJob) ID() string                                       { return j.id }
func (j SimpleJob) Execute(ctx context.Context) (interface{}, error) { return j.fn(ctx) }

// NewJob 创建简单作业
func NewJob(id string, fn func(ctx context.Context) (interface{}, error)) Job {
	return SimpleJob{id: id, fn: fn}
}

// ============================================================================
// 触发条件实现 - 控制作业何时就绪执行
// ============================================================================

// ImmediateTrigger 立即触发（用于起始作业）
type ImmediateTrigger struct{}

func (ImmediateTrigger) Ready(completed JobResults) bool { return true }
func (ImmediateTrigger) String() string                  { return "immediate" }

// AfterTrigger 在指定作业完成后触发
type AfterTrigger struct{ jobID string }

func (t AfterTrigger) Ready(completed JobResults) bool {
	_, exists := completed[t.jobID]
	return exists
}

func (t AfterTrigger) String() string { return fmt.Sprintf("after:%s", t.jobID) }

// AllOfTrigger 所有指定作业都完成后触发（AND逻辑）
type AllOfTrigger struct{ triggers []Trigger }

func (t AllOfTrigger) Ready(completed JobResults) bool {
	for _, trigger := range t.triggers {
		if !trigger.Ready(completed) {
			return false
		}
	}
	return true
}

func (t AllOfTrigger) String() string { return "all-of" }

// AnyOfTrigger 任一指定作业完成后触发（OR逻辑）
type AnyOfTrigger struct{ triggers []Trigger }

func (t AnyOfTrigger) Ready(completed JobResults) bool {
	for _, trigger := range t.triggers {
		if trigger.Ready(completed) {
			return true
		}
	}
	return false
}

func (t AnyOfTrigger) String() string { return "any-of" }

// ============================================================================
// 便捷构造函数 - 语义清晰的API
// ============================================================================

// Immediately 立即就绪触发器
func Immediately() Trigger { return ImmediateTrigger{} }

// After 在指定作业完成后触发
func After(jobID string) Trigger { return AfterTrigger{jobID} }

// AllOf 所有作业都完成后触发
func AllOf(triggers ...Trigger) Trigger { return AllOfTrigger{triggers} }

// AnyOf 任一作业完成后触发
func AnyOf(triggers ...Trigger) Trigger { return AnyOfTrigger{triggers} }

// AfterJobs 在指定作业都完成后触发 - 语法糖
func AfterJobs(jobIDs ...string) Trigger {
	triggers := make([]Trigger, len(jobIDs))
	for i, id := range jobIDs {
		triggers[i] = After(id)
	}
	return AllOf(triggers...)
}

// AfterAnyJob 在任一指定作业完成后触发 - 语法糖
func AfterAnyJob(jobIDs ...string) Trigger {
	triggers := make([]Trigger, len(jobIDs))
	for i, id := range jobIDs {
		triggers[i] = After(id)
	}
	return AnyOf(triggers...)
}

// ============================================================================
// 高级作业类型 - 组合和并行模式
// ============================================================================

// ParallelJobGroup 并行作业组 - 明确表达并行意图
type ParallelJobGroup struct {
	id   string
	jobs []Job
}

func (pjg ParallelJobGroup) ID() string { return pjg.id }

func (pjg ParallelJobGroup) Execute(ctx context.Context) (interface{}, error) {
	var wg sync.WaitGroup
	results := make([]interface{}, len(pjg.jobs))
	errors := make([]error, len(pjg.jobs))

	for i, job := range pjg.jobs {
		wg.Add(1)
		go func(index int, j Job) {
			defer wg.Done()
			result, err := j.Execute(ctx)
			results[index] = result
			errors[index] = err
		}(i, job)
	}

	wg.Wait()

	// 检查错误
	for i, err := range errors {
		if err != nil {
			return nil, fmt.Errorf("parallel job %s failed: %w", pjg.jobs[i].ID(), err)
		}
	}

	return results, nil
}

// NewParallelGroup 创建并行作业组
func NewParallelGroup(id string, jobs ...Job) Job {
	return ParallelJobGroup{id: id, jobs: jobs}
}

// SequentialJobChain 顺序作业链
type SequentialJobChain struct {
	id   string
	jobs []Job
}

func (sjc SequentialJobChain) ID() string { return sjc.id }

func (sjc SequentialJobChain) Execute(ctx context.Context) (interface{}, error) {
	var lastResult interface{}

	for _, job := range sjc.jobs {
		result, err := job.Execute(ctx)
		if err != nil {
			return nil, fmt.Errorf("sequential job %s failed: %w", job.ID(), err)
		}
		lastResult = result
	}

	return lastResult, nil
}

// NewSequentialChain 创建顺序作业链
func NewSequentialChain(id string, jobs ...Job) Job {
	return SequentialJobChain{id: id, jobs: jobs}
}

// ============================================================================
// 与Agent系统集成的便捷作业类型
// ============================================================================

// AgentJobAdapter 将Agent任务适配为工作流作业
// 这样可以在工作流中执行Agent任务，概念层次清晰：
// Workflow Job -> 包含 -> Agent Task
type AgentJobAdapter struct {
	id          string
	description string
	// 注意：这里可以引用 agent.Agent 和 agent.Task
	// 但现在先用interface{}避免循环导入，后续可以重构
}

func (aja AgentJobAdapter) ID() string { return aja.id }

func (aja AgentJobAdapter) Execute(ctx context.Context) (interface{}, error) {
	// TODO: 集成Agent系统
	// agent := getAgent()
	// task := createAgentTask()
	// return agent.Execute(ctx, task)
	return fmt.Sprintf("Agent job %s executed: %s", aja.id, aja.description), nil
}

// NewAgentJob 创建Agent作业适配器
func NewAgentJob(id, description string) Job {
	return AgentJobAdapter{
		id:          id,
		description: description,
	}
}
