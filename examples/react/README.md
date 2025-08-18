# ReAct模式使用指南

这个示例展示了如何在greensoulai中使用ReAct（Reasoning and Acting）模式。

## ReAct模式简介

ReAct是一种让AI模型能够进行结构化推理和行动的模式，它将AI的思考过程分解为：

1. **Thought（思考）**: AI分析问题的推理过程
2. **Action（行动）**: 选择要执行的工具或动作
3. **Observation（观察）**: 工具执行后的结果观察
4. **Final Answer（最终答案）**: 基于推理过程得出的最终结论

## 运行示例

```bash
cd /Users/linyi/research/greensoulai/examples/react
go run react_example.go
```

## 核心特性

### 1. 模式切换

```go
// 启用ReAct模式
agent.SetReActMode(true)

// 或者在创建时指定
config := agent.AgentConfig{
    // ... 其他配置
    ExecutionConfig: agent.ExecutionConfig{
        Mode: agent.ModeReAct,  // 设置为ReAct模式
        ReActConfig: &agent.ReActConfig{
            MaxIterations: 10,
            ThoughtTimeout: 30 * time.Second,
            EnableDebugOutput: true,
        },
    },
}
```

### 2. 配置选项

```go
reactConfig := &agent.ReActConfig{
    MaxIterations:          10,                // 最大推理迭代次数
    ThoughtTimeout:         30 * time.Second, // 单次思考超时时间
    EnableDebugOutput:      true,             // 启用调试输出
    StrictFormatValidation: true,             // 严格格式验证
    AllowFallbackToJSON:    true,             // 允许回退到JSON模式
    CustomPromptTemplate:   "",               // 自定义提示词模板
}
```

### 3. 执行方式

```go
// 使用ReAct模式执行
output, trace, err := agent.ExecuteWithReAct(ctx, task)
if err != nil {
    // 处理错误
}

// 查看推理轨迹
for i, step := range trace.Steps {
    fmt.Printf("Step %d:\n", i+1)
    fmt.Printf("  Thought: %s\n", step.Thought)
    if step.Action != "" {
        fmt.Printf("  Action: %s\n", step.Action)
        fmt.Printf("  Observation: %s\n", step.Observation)
    }
    if step.FinalAnswer != "" {
        fmt.Printf("  Final Answer: %s\n", step.FinalAnswer)
    }
}
```

## ReAct输出格式示例

```
Thought: 我需要分析用户的问题，确定需要使用什么工具
Action: web_search
Action Input: {"query": "人工智能2024年发展趋势", "limit": 5}
Observation: 搜索到5篇相关文章，包含最新的AI发展动态

Thought: 根据搜索结果，我现在有了足够的信息来回答问题
Final Answer: 基于最新资料，2024年人工智能发展呈现以下趋势...
```

## 与JSON模式的对比

| 特性 | ReAct模式 | JSON模式 |
|------|-----------|-----------|
| **可解释性** | 高（每步推理可见） | 中等 |
| **调试能力** | 强（步骤追踪） | 一般 |
| **执行效率** | 中等（多轮交互） | 高 |
| **复杂任务** | 擅长 | 一般 |
| **简单任务** | 过度设计 | 合适 |

## 最佳实践

### 1. 选择合适的模式

- **使用ReAct模式的场景**:
  - 需要多步推理的复杂任务
  - 需要工具链组合的任务
  - 要求高度可解释性的任务
  - 调试和开发阶段

- **使用JSON模式的场景**:
  - 简单直接的任务
  - 对性能要求高的场景
  - 批量处理任务

### 2. 配置优化

```go
// 生产环境配置
productionConfig := &agent.ReActConfig{
    MaxIterations:          8,     // 适中的迭代次数
    ThoughtTimeout:         15 * time.Second,
    EnableDebugOutput:      false, // 关闭调试输出
    StrictFormatValidation: true,
    AllowFallbackToJSON:    true,  // 允许错误恢复
}

// 开发环境配置  
developmentConfig := &agent.ReActConfig{
    MaxIterations:          15,    // 更多迭代用于调试
    ThoughtTimeout:         60 * time.Second,
    EnableDebugOutput:      true,  // 开启详细日志
    StrictFormatValidation: false, // 宽松验证
    AllowFallbackToJSON:    true,
}
```

### 3. 错误处理

```go
output, trace, err := agent.ExecuteWithReAct(ctx, task)
if err != nil {
    // 检查是否是特定类型的错误
    if strings.Contains(err.Error(), "maximum iterations") {
        // 处理迭代次数超限
        log.Warn("Task too complex, consider breaking down")
    } else if strings.Contains(err.Error(), "parse failed") {
        // 处理解析错误
        log.Warn("LLM output format issue")
    }
    return err
}

// 检查轨迹质量
if !trace.IsCompleted {
    log.Warn("Task incomplete, check trace for issues")
}
```

## 性能考量

1. **内存使用**: ReAct模式会保存完整的推理轨迹，内存使用比JSON模式高
2. **执行时间**: 多轮LLM调用导致执行时间较长
3. **Token消耗**: 包含思考过程的输出会消耗更多tokens

## 故障排除

### 常见问题

1. **无限循环**: 设置合理的MaxIterations
2. **解析失败**: 启用AllowFallbackToJSON
3. **超时**: 调整ThoughtTimeout设置
4. **格式错误**: 检查LLM的ReAct格式训练程度

### 调试技巧

```go
// 启用详细日志
config.ReActConfig.EnableDebugOutput = true

// 检查推理轨迹
for _, step := range trace.Steps {
    if step.Error != "" {
        fmt.Printf("Step error: %s\n", step.Error)
    }
}

// 监控性能
fmt.Printf("Total steps: %d\n", trace.IterationCount)
fmt.Printf("Total time: %v\n", trace.TotalDuration)
```

## 总结

ReAct模式为greensoulai提供了强大的可解释性和复杂推理能力。虽然在性能上有一定开销，但在需要透明决策过程的应用场景中具有重要价值。

通过合理的配置和使用，ReAct模式可以显著提升AI Agent在复杂任务中的表现。
