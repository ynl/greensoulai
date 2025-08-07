// 演示工作流状态传递的完整示例
// 展示如何在作业间传递和共享数据
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ynl/greensoulai/pkg/flow"
)

func main() {
	fmt.Println("🔄 **工作流状态传递演示**")
	fmt.Println("   展示作业间如何传递和共享数据")
	fmt.Println()

	// 创建工作流
	workflow := flow.NewWorkflow("stateful-processing")

	// 1. 数据收集作业 - 将数据存储到状态中
	dataCollectionJob := flow.NewStatefulJob("data-collection", func(ctx context.Context, state flow.FlowState) (interface{}, error) {
		fmt.Println("   📥 收集数据...")
		time.Sleep(100 * time.Millisecond)

		// 将收集的数据存储到工作流状态中
		rawData := []string{"用户评论1", "用户评论2", "用户评论3", "产品描述"}
		// 转换为[]interface{}以便后续使用GetSlice方法
		rawDataInterface := make([]interface{}, len(rawData))
		for i, v := range rawData {
			rawDataInterface[i] = v
		}
		state.Set("raw_data", rawDataInterface)
		state.Set("data_count", len(rawData))
		state.Set("collection_time", time.Now())

		fmt.Printf("   📦 已收集 %d 条数据并存储到状态中\n", len(rawData))
		return fmt.Sprintf("收集了%d条数据", len(rawData)), nil
	})

	// 2. 数据清洗作业 - 从状态中读取数据，清洗后更新状态
	dataCleaningJob := flow.NewStatefulJob("data-cleaning", func(ctx context.Context, state flow.FlowState) (interface{}, error) {
		fmt.Println("   🧹 清洗数据...")
		time.Sleep(80 * time.Millisecond)

		// 从状态中获取原始数据
		rawData, exists := state.GetSlice("raw_data")
		if !exists {
			return nil, fmt.Errorf("未找到原始数据")
		}

		// 模拟数据清洗
		cleanedData := make([]string, 0)
		for _, item := range rawData {
			if str, ok := item.(string); ok && len(str) > 0 {
				cleaned := fmt.Sprintf("[已清洗] %s", str)
				cleanedData = append(cleanedData, cleaned)
			}
		}

		// 将清洗后的数据存储回状态（转换为[]interface{}）
		cleanedDataInterface := make([]interface{}, len(cleanedData))
		for i, v := range cleanedData {
			cleanedDataInterface[i] = v
		}
		state.Set("cleaned_data", cleanedDataInterface)
		state.Set("cleaning_time", time.Now())

		fmt.Printf("   ✨ 已清洗 %d 条数据\n", len(cleanedData))
		return fmt.Sprintf("清洗了%d条数据", len(cleanedData)), nil
	})

	// 3. 质量分析作业 - 基于清洗后的数据进行分析
	qualityAnalysisJob := flow.NewStatefulJob("quality-analysis", func(ctx context.Context, state flow.FlowState) (interface{}, error) {
		fmt.Println("   🔍 质量分析...")
		time.Sleep(120 * time.Millisecond)

		// 从状态中获取清洗后的数据
		cleanedData, exists := state.GetSlice("cleaned_data")
		if !exists {
			return nil, fmt.Errorf("未找到清洗后的数据")
		}

		// 模拟质量分析
		qualityScore := float64(len(cleanedData)) * 20.5 // 简单的质量评分逻辑

		// 将分析结果存储到状态
		state.Set("quality_score", qualityScore)
		state.Set("quality_analysis_time", time.Now())

		analysisResult := fmt.Sprintf("质量评分: %.1f%%", qualityScore)
		fmt.Printf("   📊 %s\n", analysisResult)
		return analysisResult, nil
	})

	// 4. 情感分析作业 - 同时基于清洗后的数据进行分析
	sentimentAnalysisJob := flow.NewStatefulJob("sentiment-analysis", func(ctx context.Context, state flow.FlowState) (interface{}, error) {
		fmt.Println("   😊 情感分析...")
		time.Sleep(100 * time.Millisecond)

		// 从状态中获取清洗后的数据
		cleanedData, exists := state.GetSlice("cleaned_data")
		if !exists {
			return nil, fmt.Errorf("未找到清洗后的数据")
		}

		// 模拟情感分析
		positiveCount := 0
		for _, item := range cleanedData {
			if str, ok := item.(string); ok && len(str) > 10 { // 简单逻辑：长文本倾向正面
				positiveCount++
			}
		}

		positiveRatio := float64(positiveCount) / float64(len(cleanedData)) * 100

		// 将分析结果存储到状态
		state.Set("positive_ratio", positiveRatio)
		state.Set("sentiment_analysis_time", time.Now())

		sentimentResult := fmt.Sprintf("正面情感: %.1f%%", positiveRatio)
		fmt.Printf("   💚 %s\n", sentimentResult)
		return sentimentResult, nil
	})

	// 5. 报告生成作业 - 汇总所有分析结果
	reportGenerationJob := flow.NewStatefulJob("report-generation", func(ctx context.Context, state flow.FlowState) (interface{}, error) {
		fmt.Println("   📝 生成综合报告...")
		time.Sleep(60 * time.Millisecond)

		// 从状态中获取所有分析结果
		dataCount, _ := state.GetInt("data_count")
		qualityScore, _ := state.GetFloat64("quality_score")
		positiveRatio, _ := state.GetFloat64("positive_ratio")
		collectionTime, _ := state.Get("collection_time")

		// 生成综合报告
		report := fmt.Sprintf(`
📋 **数据分析综合报告**
   • 数据量: %d 条
   • 质量评分: %.1f%%
   • 正面情感比例: %.1f%%
   • 收集时间: %v
   • 处理状态: 完成`,
			dataCount, qualityScore, positiveRatio, collectionTime)

		// 将最终报告存储到状态
		state.Set("final_report", report)
		state.Set("generation_time", time.Now())

		fmt.Println("   📊 综合报告已生成")
		return "综合分析报告已完成", nil
	})

	// 构建工作流 - 展示复杂的依赖关系
	workflow.
		AddJob(dataCollectionJob, flow.Immediately()).                                        // 1. 首先收集数据
		AddJob(dataCleaningJob, flow.After("data-collection")).                               // 2. 基于收集的数据进行清洗
		AddJob(qualityAnalysisJob, flow.After("data-cleaning")).                              // 3. 基于清洗数据并行分析
		AddJob(sentimentAnalysisJob, flow.After("data-cleaning")).                            // 4. 基于清洗数据并行分析
		AddJob(reportGenerationJob, flow.AfterJobs("quality-analysis", "sentiment-analysis")) // 5. 等待所有分析完成

	fmt.Println("   🚀 执行带状态传递的工作流...")
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

	fmt.Printf("\n   🗂️ **最终工作流状态**:\n")
	finalState := result.FinalState
	for _, key := range finalState.Keys() {
		if value, exists := finalState.Get(key); exists {
			fmt.Printf("   • %s: %v\n", key, value)
		}
	}

	fmt.Printf("\n   📋 **执行轨迹** (展示状态传递过程):\n")
	for i, trace := range result.JobTrace {
		fmt.Printf("   %d. %s (批次%d): %v\n",
			i+1, trace.JobID, trace.BatchID, trace.Duration)
	}

	fmt.Printf("\n   ✨ **状态传递亮点**:\n")
	fmt.Printf("   • 数据在作业间无缝传递\n")
	fmt.Printf("   • 支持并发安全的状态访问\n")
	fmt.Printf("   • 提供类型安全的状态获取方法\n")
	fmt.Printf("   • 完整保留工作流执行过程中的所有状态\n")
	fmt.Printf("   • 向下兼容：普通Job无需修改即可使用\n")

	fmt.Printf("\n   🔧 **技术特性**:\n")
	fmt.Printf("   • FlowState: 线程安全的状态存储\n")
	fmt.Printf("   • StatefulJob: 支持状态传递的作业接口\n")
	fmt.Printf("   • 自动适配: 普通Job和StatefulJob混合使用\n")
	fmt.Printf("   • 状态持久化: 执行完成后状态完整保留\n")
}
