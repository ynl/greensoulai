// 对比展示：原始API vs 增强状态传递API
// 演示向下兼容性和新功能的使用
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ynl/greensoulai/pkg/flow"
)

func main() {
	fmt.Println("🔄 **API对比演示：原始 vs 状态传递**")
	fmt.Println()

	// ========================================================================
	// 方案1: 原始API - 无状态传递（向下兼容）
	// ========================================================================
	fmt.Println("📋 **方案1: 原始API（向下兼容）**")
	runOriginalAPI()

	fmt.Println()

	// ========================================================================
	// 方案2: 状态传递API - 作业间数据共享
	// ========================================================================
	fmt.Println("📋 **方案2: 状态传递API（新功能）**")
	runStatefulAPI()

	fmt.Println()

	// ========================================================================
	// 方案3: 混合使用 - 普通Job + StatefulJob
	// ========================================================================
	fmt.Println("📋 **方案3: 混合使用（最佳实践）**")
	runMixedAPI()
}

// 原始API - 每个Job是独立的，无状态传递
func runOriginalAPI() {
	fmt.Println("   特点：Job独立执行，无数据传递")

	dataJob := flow.NewJob("collect", func(ctx context.Context) (interface{}, error) {
		fmt.Println("   📥 收集数据（无法传递给其他Job）")
		time.Sleep(50 * time.Millisecond)
		return "收集了100条数据", nil
	})

	processJob := flow.NewJob("process", func(ctx context.Context) (interface{}, error) {
		fmt.Println("   🔄 处理数据（无法访问collect的结果）")
		time.Sleep(80 * time.Millisecond)
		return "处理完成", nil
	})

	workflow := flow.NewWorkflow("original-style").
		AddJob(dataJob, flow.Immediately()).
		AddJob(processJob, flow.After("collect"))

	result, err := workflow.Run(context.Background())
	if err != nil {
		fmt.Printf("   ❌ 失败: %v\n", err)
		return
	}

	fmt.Printf("   ✅ 结果: %s\n", result.FinalResult)
	fmt.Printf("   📊 所有结果: %v\n", result.AllResults)
	fmt.Printf("   🗂️ 状态键数量: %d\n", len(result.FinalState.Keys()))
}

// 状态传递API - Job间可以共享数据
func runStatefulAPI() {
	fmt.Println("   特点：Job间可以传递和共享数据")

	dataJob := flow.NewStatefulJob("collect", func(ctx context.Context, state flow.FlowState) (interface{}, error) {
		fmt.Println("   📥 收集数据并存储到状态")
		time.Sleep(50 * time.Millisecond)

		// 存储数据到状态中
		data := []interface{}{"项目A", "项目B", "项目C"}
		state.Set("collected_data", data)
		state.Set("data_count", len(data))
		state.Set("collect_time", time.Now())

		return fmt.Sprintf("收集了%d条数据", len(data)), nil
	})

	processJob := flow.NewStatefulJob("process", func(ctx context.Context, state flow.FlowState) (interface{}, error) {
		fmt.Println("   🔄 从状态中获取数据进行处理")
		time.Sleep(80 * time.Millisecond)

		// 从状态中获取数据
		data, exists := state.GetSlice("collected_data")
		if !exists {
			return nil, fmt.Errorf("未找到收集的数据")
		}

		count, _ := state.GetInt("data_count")

		// 处理数据并更新状态
		processedData := make([]interface{}, len(data))
		for i, item := range data {
			processedData[i] = fmt.Sprintf("已处理_%s", item)
		}

		state.Set("processed_data", processedData)
		state.Set("process_time", time.Now())

		return fmt.Sprintf("处理了%d条数据", count), nil
	})

	workflow := flow.NewWorkflow("stateful-style").
		AddJob(dataJob, flow.Immediately()).
		AddJob(processJob, flow.After("collect"))

	result, err := workflow.Run(context.Background())
	if err != nil {
		fmt.Printf("   ❌ 失败: %v\n", err)
		return
	}

	fmt.Printf("   ✅ 结果: %s\n", result.FinalResult)
	fmt.Printf("   🗂️ 状态键数量: %d\n", len(result.FinalState.Keys()))
	fmt.Printf("   📋 状态内容:\n")
	for _, key := range result.FinalState.Keys() {
		if value, exists := result.FinalState.Get(key); exists {
			fmt.Printf("      • %s: %v\n", key, value)
		}
	}
}

// 混合使用 - 展示兼容性和灵活性
func runMixedAPI() {
	fmt.Println("   特点：普通Job和StatefulJob可以混合使用")

	// 普通Job - 简单任务，无需状态
	simpleJob := flow.NewJob("simple", func(ctx context.Context) (interface{}, error) {
		fmt.Println("   ⚡ 简单任务（无状态）")
		time.Sleep(30 * time.Millisecond)
		return "简单任务完成", nil
	})

	// StatefulJob - 需要存储数据
	setupJob := flow.NewStatefulJob("setup", func(ctx context.Context, state flow.FlowState) (interface{}, error) {
		fmt.Println("   🔧 设置任务（存储配置到状态）")
		time.Sleep(40 * time.Millisecond)

		config := map[string]interface{}{
			"batch_size": 100,
			"timeout":    "30s",
			"retries":    3,
		}
		state.Set("config", config)
		state.Set("setup_complete", true)

		return "配置已设置", nil
	})

	// StatefulJob - 使用配置数据
	workerJob := flow.NewStatefulJob("worker", func(ctx context.Context, state flow.FlowState) (interface{}, error) {
		fmt.Println("   👷 工作任务（使用状态中的配置）")
		time.Sleep(60 * time.Millisecond)

		// 获取配置
		config, exists := state.GetMap("config")
		if !exists {
			return nil, fmt.Errorf("未找到配置")
		}

		batchSize := config["batch_size"].(int)

		// 模拟使用配置进行工作
		result := fmt.Sprintf("使用批次大小%d处理任务", batchSize)
		state.Set("worker_result", result)

		return result, nil
	})

	// 普通Job - 清理任务，无需访问状态
	cleanupJob := flow.NewJob("cleanup", func(ctx context.Context) (interface{}, error) {
		fmt.Println("   🧹 清理任务（无状态）")
		time.Sleep(20 * time.Millisecond)
		return "清理完成", nil
	})

	workflow := flow.NewWorkflow("mixed-style").
		AddJob(simpleJob, flow.Immediately()).
		AddJob(setupJob, flow.After("simple")).
		AddJob(workerJob, flow.After("setup")).
		AddJob(cleanupJob, flow.After("worker"))

	result, err := workflow.Run(context.Background())
	if err != nil {
		fmt.Printf("   ❌ 失败: %v\n", err)
		return
	}

	fmt.Printf("   ✅ 最终结果: %s\n", result.FinalResult)
	fmt.Printf("   📊 Job结果: %v\n", result.AllResults)
	fmt.Printf("   🗂️ 有状态的Job设置了%d个状态键\n", len(result.FinalState.Keys()))

	// 展示状态内容
	if setupComplete, exists := result.FinalState.GetBool("setup_complete"); exists && setupComplete {
		fmt.Printf("   ✅ 配置设置成功\n")
	}

	if workerResult, exists := result.FinalState.GetString("worker_result"); exists {
		fmt.Printf("   💼 工作结果: %s\n", workerResult)
	}

	fmt.Printf("\n   ✨ **混合使用的优势**:\n")
	fmt.Printf("   • 向下兼容: 现有Job代码无需修改\n")
	fmt.Printf("   • 按需使用: 只在需要时使用状态传递\n")
	fmt.Printf("   • 性能优化: 简单Job避免状态开销\n")
	fmt.Printf("   • 灵活组合: 普通Job和StatefulJob自由搭配\n")

	fmt.Printf("\n   🎯 **使用建议**:\n")
	fmt.Printf("   • 简单独立任务 → 使用Job\n")
	fmt.Printf("   • 需要数据传递 → 使用StatefulJob\n")
	fmt.Printf("   • 配置型任务 → 使用StatefulJob\n")
	fmt.Printf("   • 清理型任务 → 使用Job\n")
}
