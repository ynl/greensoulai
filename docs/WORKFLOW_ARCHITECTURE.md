# 工作流系统架构 - 并行作业编排

## 🎯 **系统概述**

**GreenSoulAI 工作流系统** 是一个**高性能并行作业编排引擎**，专门设计用于处理复杂的AI工作流场景。

### 核心特性
- ⚡ **真正的并行执行** - 独立作业并行处理，显著提升性能
- 🏗️ **层次清晰** - Job(作业) / Agent Task(业务任务) 概念分明
- 🚀 **Go原生** - 充分利用Go并发优势，类型安全
- 📊 **详细指标** - 完整的并行执行统计和性能分析
- 🔧 **简洁API** - 最小化接口，易于使用

---

## 🏗️ **架构设计**

### 核心概念层次

```
🎯 工作流系统架构层次
Workflow (工作流编排)     ← 顶层：协调多个作业
    ↓
Job (作业单元)            ← 中层：可并行执行的工作单元  
    ↓
Agent Task (业务任务)     ← 底层：具体的AI任务执行
```

### 核心接口

```go
// Job - 工作流中的作业单元
type Job interface {
    ID() string
    Execute(ctx context.Context) (interface{}, error)
}

// Trigger - 作业触发条件
type Trigger interface {
    Ready(completed JobResults) bool
    String() string
}

// Workflow - 并行作业编排
type Workflow interface {
    AddJob(job Job, trigger Trigger) Workflow
    Run(ctx context.Context) (*ExecutionResult, error)
    RunAsync(ctx context.Context) <-chan *ExecutionResult
}
```

---

## 🚀 **快速开始**

### 基础使用

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/ynl/greensoulai/pkg/flow"
)

func main() {
    // 创建工作流
    workflow := flow.NewWorkflow("ai-analysis")
    
    // 定义作业
    dataJob := flow.NewJob("data-prep", func(ctx context.Context) (interface{}, error) {
        time.Sleep(100 * time.Millisecond)
        return "数据已准备", nil
    })
    
    // 三个可并行执行的分析作业
    qualityJob := flow.NewJob("quality", func(ctx context.Context) (interface{}, error) {
        time.Sleep(150 * time.Millisecond)  
        return "质量评分: 85%", nil
    })
    
    sentimentJob := flow.NewJob("sentiment", func(ctx context.Context) (interface{}, error) {
        time.Sleep(120 * time.Millisecond)
        return "情感: 积极", nil  
    })
    
    topicJob := flow.NewJob("topic", func(ctx context.Context) (interface{}, error) {
        time.Sleep(180 * time.Millisecond)
        return "主题: AI技术", nil
    })
    
    reportJob := flow.NewJob("report", func(ctx context.Context) (interface{}, error) {
        time.Sleep(80 * time.Millisecond)
        return "分析报告已生成", nil
    })
    
    // 构建工作流 - 并行执行模式
    workflow.
        AddJob(dataJob, flow.Immediately()).                    // 立即执行
        AddJob(qualityJob, flow.After("data-prep")).           // 三个分析作业
        AddJob(sentimentJob, flow.After("data-prep")).         // 数据准备后
        AddJob(topicJob, flow.After("data-prep")).             // 并行执行！
        AddJob(reportJob, flow.AfterJobs("quality", "sentiment", "topic"))  // 等全部完成
    
    // 执行并获取详细指标
    result, err := workflow.Run(context.Background())
    if err != nil {
        panic(err)
    }
    
    // 输出执行结果
    fmt.Printf("🎯 总耗时: %v\n", result.Duration)
    fmt.Printf("⚡ 并发数: %d\n", result.Metrics.MaxConcurrency)  
    fmt.Printf("📈 加速比: %.2fx\n", result.Metrics.ParallelEfficiency)
    fmt.Printf("📋 最终结果: %s\n", result.FinalResult)
}
```

### 典型输出

```
🎯 总耗时: 363ms
⚡ 并发数: 3  
📈 加速比: 1.75x
📋 最终结果: 分析报告已生成
```

---

## 📊 **触发条件系统**

### 基础触发器

```go
// 立即执行
flow.Immediately()

// 单依赖
flow.After("job-id")

// 多依赖 - AND逻辑（所有完成）  
flow.AfterJobs("job1", "job2", "job3")

// 多依赖 - OR逻辑（任一完成）
flow.AfterAnyJob("job1", "job2", "job3")
```

### 复合触发器

```go
// 复杂组合条件
allCondition := flow.AllOf(
    flow.After("data-prep"),
    flow.After("model-ready")
)

anyCondition := flow.AnyOf(
    flow.After("backup-data"),  
    flow.After("cache-data")
)
```

---

## 🔧 **高级功能**

### 并行作业组

```go
// 创建并行作业组 - 内部作业同时执行
parallelGroup := flow.NewParallelGroup("analysis-group",
    qualityJob,
    sentimentJob, 
    topicJob,
)

workflow.AddJob(parallelGroup, flow.After("data-prep"))
```

### 顺序作业链

```go  
// 创建顺序作业链 - 内部作业依次执行
sequentialChain := flow.NewSequentialChain("preprocessing",
    validateJob,
    cleanJob,
    normalizeJob,
)

workflow.AddJob(sequentialChain, flow.Immediately())
```

### 异步执行

```go
// 异步执行工作流
resultChan := workflow.RunAsync(context.Background())

// 处理其他任务...

// 获取结果
result := <-resultChan
if result.Error != nil {
    log.Fatal(result.Error)
}
```

---

## 🔗 **与Agent系统集成**

### Job包装Agent任务

```go
import "github.com/ynl/greensoulai/internal/agent"

// 创建Agent任务包装器
func createAnalysisJob(agentInstance agent.Agent) flow.Job {
    return flow.NewJob("ai-analysis", func(ctx context.Context) (interface{}, error) {
        // 创建Agent业务任务
        task := agent.NewBaseTask(
            "分析用户反馈数据",
            "生成情感分析和主题提取报告"
        )
        
        // 执行Agent任务
        return agentInstance.Execute(ctx, task)
    })
}

// 在工作流中使用
analysisJob := createAnalysisJob(myAgent)
workflow.AddJob(analysisJob, flow.After("data-collection"))
```

---

## 📈 **性能指标**

### ExecutionResult 结构

```go
type ExecutionResult struct {
    FinalResult   interface{}        // 最终结果
    AllResults    JobResults         // 所有作业结果
    JobTrace      []JobExecution     // 执行轨迹
    Metrics       *ParallelMetrics   // 性能指标
    Duration      time.Duration      // 总耗时
}

type ParallelMetrics struct {
    TotalJobs         int             // 总作业数
    ParallelBatches   int             // 并行批次数
    MaxConcurrency    int             // 最大并发数
    ParallelEfficiency float64        // 并行效率
    SerialTime        time.Duration   // 假设串行时间
    ParallelTime      time.Duration   // 实际并行时间
}
```

### 性能分析示例

```go
result, _ := workflow.Run(ctx)

fmt.Printf("📊 执行统计:\n")
fmt.Printf("   总作业数: %d\n", result.Metrics.TotalJobs)
fmt.Printf("   并行批次: %d\n", result.Metrics.ParallelBatches)  
fmt.Printf("   最大并发: %d\n", result.Metrics.MaxConcurrency)
fmt.Printf("   串行耗时: %v\n", result.Metrics.SerialTime)
fmt.Printf("   并行耗时: %v\n", result.Metrics.ParallelTime)
fmt.Printf("   加速效果: %.2fx\n", result.Metrics.ParallelEfficiency)

fmt.Printf("\n🔍 执行轨迹:\n")
for i, trace := range result.JobTrace {
    fmt.Printf("   %d. %s (批次%d): %v\n", 
        i+1, trace.JobID, trace.BatchID, trace.Duration)
}
```

---

## 🛠️ **最佳实践**

### 1. 作业设计原则

```go
// ✅ 好的作业设计 - 职责单一，无副作用
func createDataValidationJob() flow.Job {
    return flow.NewJob("validate", func(ctx context.Context) (interface{}, error) {
        // 职责明确：仅做数据验证
        return validateData(ctx)
    })
}

// ❌ 避免 - 作业过于复杂
func badComplexJob() flow.Job {
    return flow.NewJob("complex", func(ctx context.Context) (interface{}, error) {
        // 做太多事情：读取+验证+转换+存储
        data := readData()
        validated := validateData(data)
        transformed := transformData(validated)  
        return saveData(transformed), nil
    })
}
```

### 2. 错误处理

```go
func robustJob() flow.Job {
    return flow.NewJob("robust", func(ctx context.Context) (interface{}, error) {
        // 检查上下文取消
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
        }
        
        // 业务逻辑包含错误处理
        result, err := riskyOperation()
        if err != nil {
            return nil, fmt.Errorf("作业执行失败: %w", err)
        }
        
        return result, nil
    })
}
```

### 3. 资源管理

```go
func resourceSafeJob() flow.Job {
    return flow.NewJob("resource-safe", func(ctx context.Context) (interface{}, error) {
        // 获取资源
        resource, err := acquireResource()
        if err != nil {
            return nil, err
        }
        
        // 确保资源释放
        defer resource.Close()
        
        // 使用资源
        return processWithResource(ctx, resource)
    })
}
```

---

## 🧪 **测试指南**

### 单元测试示例

```go
func TestWorkflowBasic(t *testing.T) {
    ctx := context.Background()
    
    job := flow.NewJob("test", func(ctx context.Context) (interface{}, error) {
        return "success", nil
    })
    
    workflow := flow.NewWorkflow("test-workflow").
        AddJob(job, flow.Immediately())
    
    result, err := workflow.Run(ctx)
    
    assert.NoError(t, err)
    assert.Equal(t, "success", result.FinalResult)
}
```

### 并行性能测试

```go
func TestParallelPerformance(t *testing.T) {
    ctx := context.Background()
    
    workflow := flow.NewWorkflow("perf-test")
    
    // 添加10个并行作业
    for i := 0; i < 10; i++ {
        job := createSlowJob(100 * time.Millisecond)
        workflow.AddJob(job, flow.Immediately())
    }
    
    start := time.Now()
    result, err := workflow.Run(ctx)
    duration := time.Since(start)
    
    assert.NoError(t, err)
    assert.Less(t, duration, 200*time.Millisecond) // 应比串行快
    assert.Equal(t, 10, result.Metrics.MaxConcurrency)
}
```

---

## 📚 **API参考**

### 创建函数

| 函数 | 说明 | 示例 |
|------|------|------|
| `NewWorkflow(name)` | 创建工作流 | `flow.NewWorkflow("pipeline")` |
| `NewJob(id, func)` | 创建简单作业 | `flow.NewJob("task1", taskFunc)` |
| `NewParallelGroup(id, jobs...)` | 并行作业组 | `flow.NewParallelGroup("group", job1, job2)` |
| `NewSequentialChain(id, jobs...)` | 顺序作业链 | `flow.NewSequentialChain("chain", job1, job2)` |

### 触发器函数

| 触发器 | 说明 | 示例 |
|--------|------|------|
| `Immediately()` | 立即触发 | `flow.Immediately()` |
| `After(jobID)` | 单作业完成后 | `flow.After("job1")` |  
| `AfterJobs(ids...)` | 多作业都完成后 | `flow.AfterJobs("job1", "job2")` |
| `AfterAnyJob(ids...)` | 任一作业完成后 | `flow.AfterAnyJob("job1", "job2")` |
| `AllOf(triggers...)` | 所有条件满足 | `flow.AllOf(trigger1, trigger2)` |
| `AnyOf(triggers...)` | 任一条件满足 | `flow.AnyOf(trigger1, trigger2)` |

---

## 🎯 **总结**

GreenSoulAI工作流系统通过以下设计实现了高效的并行作业编排：

1. **清晰的概念模型** - Job作为工作流作业单元，与Agent Task业务任务层次分明
2. **真正的并行执行** - 利用Go的goroutine实现高效并发处理  
3. **灵活的触发条件** - 支持复杂的依赖关系和执行逻辑
4. **完整的可观测性** - 详细的性能指标和执行轨迹
5. **简洁的API设计** - 最小化接口，易于理解和使用

这个系统特别适合处理**复杂AI工作流场景**，如数据处理管道、多模型推理、结果聚合等，能够显著提升执行效率和系统吞吐量。
