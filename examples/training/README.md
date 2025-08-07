# CrewAI-Go 训练系统演示

这个示例展示了如何使用 CrewAI-Go 的完整训练系统来提升 Crew 的性能。

## 功能特性

### 🎯 训练核心功能

- **完整训练会话管理** - 自动化的训练流程控制
- **人工反馈收集** - 交互式质量评估和改进建议
- **性能指标分析** - 自动生成执行时间、成功率等关键指标
- **训练数据持久化** - JSON格式保存训练历史和结果
- **早停机制** - 基于性能停滞的智能训练终止
- **事件驱动架构** - 实时监控训练进程

### 📊 监控和分析

- **实时进度跟踪** - 训练过程的可视化反馈
- **性能趋势分析** - 改进率和一致性评估
- **智能建议生成** - 基于训练结果的优化建议
- **详细训练报告** - 包含洞察、警告和建议的综合报告

## 使用方法

### 基本训练

```bash
# 运行基本训练演示
cd examples/training
go run main.go
```

### 高级配置

```go
// 创建高级训练配置
config := training.CreateAdvancedTrainingConfig(20, "advanced_training.json", map[string]interface{}{
    "task": "Write a technical documentation",
    "style": "clear and comprehensive",
    "audience": "developers",
})

// 启用早停
config.EarlyStopping = true
config.PatientceEpochs = 5
config.MinImprovement = 0.05

// 自定义反馈超时
config.FeedbackTimeout = 3 * time.Minute
```

## 训练配置参数

| 参数 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `Iterations` | int | 10 | 训练迭代次数 |
| `Filename` | string | auto | 训练数据保存文件名 |
| `CollectFeedback` | bool | true | 是否收集人工反馈 |
| `FeedbackTimeout` | time.Duration | 5分钟 | 反馈收集超时时间 |
| `MetricsEnabled` | bool | true | 是否启用性能指标分析 |
| `EarlyStopping` | bool | false | 是否启用早停机制 |
| `AutoSave` | bool | true | 是否自动保存训练数据 |
| `SaveInterval` | int | 5 | 自动保存间隔（迭代数） |

## 训练流程

1. **配置验证** - 检查训练参数的有效性
2. **训练初始化** - 创建训练会话和数据结构
3. **迭代执行** - 循环执行训练任务
4. **反馈收集** - 每次迭代后收集人工评估
5. **性能分析** - 计算关键性能指标
6. **数据保存** - 持久化训练历史
7. **早停检查** - 评估是否需要提前终止
8. **报告生成** - 创建详细的训练总结

## 事件监听

训练系统支持以下事件类型：

- `training_started` - 训练开始
- `training_iteration_started` - 单次迭代开始  
- `training_iteration_completed` - 单次迭代完成
- `training_feedback_collected` - 反馈收集完成
- `training_metrics_analyzed` - 性能分析完成
- `training_stopped` - 训练停止
- `training_error` - 训练错误

## 输出示例

```
🚀 CrewAI-Go 训练系统演示
=================================
🎯 训练开始: 12345-67890 (迭代次数: 10)
  ✅ 迭代 1/10 完成 (耗时: 234ms)
  💬 反馈收集完成 (质量评分: 7.2)
  ✅ 迭代 2/10 完成 (耗时: 198ms)
  ...
🛑 训练停止: 12345-67890 (原因: completed)

============================================================
📊 训练报告
============================================================
会话ID: 12345-67890
状态: completed
总迭代次数: 10
成功迭代: 9
失败迭代: 1
成功率: 90.0%
改进率: 15.3%
平均反馈分: 8.1
总用时: 2.1s
平均用时: 210ms

💡 洞察:
  • Significant improvement of 15.3% achieved
  • High feedback scores indicate excellent output quality

📋 建议:
  • Performance is stable and improving, continue current approach
============================================================
```

## 自定义训练

### 创建自定义执行函数

```go
// 创建真实的crew执行函数
func createCrewExecuteFunc(crew *crew.BaseCrew) func(context.Context, map[string]interface{}) (interface{}, error) {
    return func(ctx context.Context, inputs map[string]interface{}) (interface{}, error) {
        return crew.Kickoff(ctx, inputs)
    }
}

// 使用自定义执行函数
trainingUtils.RunTrainingSession(ctx, handler, config, createCrewExecuteFunc(myCrew))
```

### 自定义反馈收集

```go
// 批量反馈模式（非交互式）
feedbackData := map[string]interface{}{
    "quality_score": 8.5,
    "accuracy_score": 9.0,
    "usefulness": 7.8,
    "comments": "Good overall quality with minor improvements needed",
}

collector.CollectBatchFeedback(ctx, iterationID, outputs, feedbackData)
```

## 最佳实践

1. **合理设置迭代次数** - 通常10-50次迭代足够看到改进
2. **启用早停** - 避免过度训练和时间浪费
3. **收集高质量反馈** - 详细的反馈有助于更好的改进
4. **定期保存数据** - 防止长时间训练中的数据丢失
5. **监控性能指标** - 关注成功率、执行时间等关键指标
6. **分析训练报告** - 基于建议调整training参数

## 故障排除

### 常见问题

**训练反馈超时**
```
解决方案：增加 FeedbackTimeout 时间或切换到批量反馈模式
```

**早停过于频繁**
```
解决方案：降低 MinImprovement 阈值或增加 PatientceEpochs
```

**训练数据保存失败**
```
解决方案：检查文件路径权限和磁盘空间
```

### 调试模式

```go
// 启用详细日志
config.Verbose = true

// 禁用反馈收集进行快速测试
config.CollectFeedback = false

// 减少迭代次数进行调试
config.Iterations = 3
```

## 扩展开发

训练系统采用模块化设计，支持以下扩展：

- **自定义反馈收集器** - 实现不同的反馈收集策略
- **自定义性能分析器** - 添加新的性能指标
- **自定义存储后端** - 支持数据库等持久化方案
- **自定义事件处理器** - 集成外部监控系统

查看 `internal/training/` 目录中的接口定义了解更多扩展可能性。
