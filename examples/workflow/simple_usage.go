// 新Flow系统的简单使用示例
// 展示Job而非Task，避免与Agent系统Task冲突，强调工作流编排能力
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ynl/greensoulai/pkg/flow"
)

func main() {
	fmt.Println("🚀 **精简工作流系统使用示例**")
	fmt.Println("   使用Job而非Task，避免与Agent系统冲突")
	fmt.Println("   Job = 工作流作业单元，Task = Agent执行的业务任务")
	fmt.Println()

	// 创建工作流
	workflow := flow.NewWorkflow("ai-processing")

	// 定义作业 - 注意这里是Job，不是Task，避免与Agent系统冲突
	dataCollection := flow.NewJob("data-collection", func(ctx context.Context) (interface{}, error) {
		fmt.Println("   📥 数据收集作业执行中...")
		time.Sleep(100 * time.Millisecond)
		return "收集了1000条数据", nil
	})

	// 三个可以并行执行的分析作业
	qualityAnalysis := flow.NewJob("quality-analysis", func(ctx context.Context) (interface{}, error) {
		fmt.Println("   🔍 质量分析作业执行中...")
		time.Sleep(150 * time.Millisecond)
		return "质量评分: 85%", nil
	})

	sentimentAnalysis := flow.NewJob("sentiment-analysis", func(ctx context.Context) (interface{}, error) {
		fmt.Println("   😊 情感分析作业执行中...")
		time.Sleep(120 * time.Millisecond)
		return "正面情感: 78%", nil
	})

	topicAnalysis := flow.NewJob("topic-analysis", func(ctx context.Context) (interface{}, error) {
		fmt.Println("   📊 主题分析作业执行中...")
		time.Sleep(180 * time.Millisecond)
		return "主要话题: AI, 技术, 创新", nil
	})

	// 报告生成作业
	reportGeneration := flow.NewJob("report-generation", func(ctx context.Context) (interface{}, error) {
		fmt.Println("   📝 报告生成作业执行中...")
		time.Sleep(80 * time.Millisecond)
		return "AI分析报告已生成", nil
	})

	// 构建工作流 - API清晰表达执行逻辑
	workflow.
		AddJob(dataCollection, flow.Immediately()).               // 立即开始数据收集
		AddJob(qualityAnalysis, flow.After("data-collection")).   // 这三个作业会在
		AddJob(sentimentAnalysis, flow.After("data-collection")). // 数据收集完成后
		AddJob(topicAnalysis, flow.After("data-collection")).     // 并行执行！
		AddJob(reportGeneration, flow.AfterJobs("quality-analysis", "sentiment-analysis", "topic-analysis"))

	fmt.Println("   🔄 执行并行工作流...")
	start := time.Now()

	// 执行工作流
	result, err := workflow.Run(context.Background())
	totalTime := time.Since(start)

	if err != nil {
		fmt.Printf("❌ 执行失败: %v\n", err)
		return
	}

	fmt.Printf("\n   ✅ 工作流执行完成！\n")
	fmt.Printf("   📊 最终结果: %s\n", result.FinalResult)
	fmt.Printf("   ⏱️  总执行时间: %v\n", totalTime)
	fmt.Printf("   🔢 总作业数: %d\n", result.Metrics.TotalJobs)
	fmt.Printf("   📦 并行批次数: %d\n", result.Metrics.ParallelBatches)
	fmt.Printf("   🚀 最大并发数: %d\n", result.Metrics.MaxConcurrency)
	fmt.Printf("   📈 并行效率: %.2fx\n", result.Metrics.ParallelEfficiency)

	fmt.Printf("\n   🎯 **关键验证**:\n")
	fmt.Printf("   • 如果串行执行预计: %v\n", result.Metrics.SerialTime)
	fmt.Printf("   • 实际并行执行用时: %v\n", result.Metrics.ParallelTime)
	fmt.Printf("   • 并行加速效果: %.2fx 倍\n", result.Metrics.ParallelEfficiency)

	fmt.Printf("\n   📋 **执行轨迹**:\n")
	for i, trace := range result.JobTrace {
		fmt.Printf("   %d. %s (批次%d): %v → %s\n",
			i+1, trace.JobID, trace.BatchID, trace.Duration, trace.Result)
	}

	fmt.Printf("\n   ✨ **设计亮点**:\n")
	fmt.Printf("   • Job而非Task - 避免与Agent系统冲突\n")
	fmt.Printf("   • Job = 工作流作业单元，Task = Agent业务任务\n")
	fmt.Printf("   • 三个分析作业真正并行执行（批次2）\n")
	fmt.Printf("   • API语义清晰，层次分明\n")
	fmt.Printf("   • 完整的并行执行指标\n")
	fmt.Printf("   • 精简的接口设计\n")

	fmt.Printf("\n   🏗️ **架构层次**:\n")
	fmt.Printf("   Workflow (工作流) \n")
	fmt.Printf("   └── Job (作业单元) \n")
	fmt.Printf("       └── Agent Task (智能体任务) [未来集成]\n")
}
