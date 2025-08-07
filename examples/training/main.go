package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ynl/greensoulai/internal/training"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

func main() {
	fmt.Println("🚀 CrewAI-Go 训练系统演示")
	fmt.Println("=================================")

	// 创建日志器
	log := logger.NewConsoleLogger()

	// 创建事件总线
	eventBus := events.NewEventBus(log)

	// 订阅训练事件
	subscribeToTrainingEvents(eventBus, log)

	// 创建训练工具
	trainingUtils := training.NewTrainingUtils(log)

	// 创建训练处理器
	handler := trainingUtils.CreateTrainingHandler(eventBus, log)

	// 创建训练配置
	config := training.CreateAdvancedTrainingConfig(10, "demo_training.json", map[string]interface{}{
		"task":  "Write a creative story about AI",
		"style": "engaging and informative",
	})

	// 验证配置
	if err := trainingUtils.ValidateTrainingConfig(config); err != nil {
		log.Error("invalid training config", logger.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	// 创建模拟的执行函数（实际使用时应该是crew.Kickoff）
	executeFunc := createMockExecuteFunc(log)

	// 运行训练会话
	ctx := context.Background()
	summary, err := trainingUtils.RunTrainingSession(ctx, handler, config, executeFunc)
	if err != nil {
		log.Error("training session failed", logger.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	// 生成训练报告
	// 我们需要获取训练数据来生成报告
	trainingData, err := handler.LoadTrainingData(ctx, config.Filename)
	if err != nil {
		log.Warn("failed to load training data for report", logger.Field{Key: "error", Value: err})
		// 创建基本报告
		fmt.Println("\n📊 训练完成！")
		fmt.Printf("总迭代次数: %d\n", config.Iterations)
		if summary != nil {
			fmt.Printf("成功率: %.1f%%\n", float64(summary.SuccessfulRuns)/float64(summary.TotalIterations)*100)
			fmt.Printf("改进率: %.1f%%\n", summary.ImprovementRate)
		}
	} else {
		report := trainingUtils.GenerateTrainingReport(trainingData)
		printTrainingReport(report)
	}

	fmt.Println("\n✅ 训练演示完成！")
}

// createMockExecuteFunc 创建模拟执行函数
func createMockExecuteFunc(log logger.Logger) func(context.Context, map[string]interface{}) (interface{}, error) {
	executionCount := 0

	return func(ctx context.Context, inputs map[string]interface{}) (interface{}, error) {
		executionCount++

		// 模拟执行时间
		time.Sleep(time.Duration(100+executionCount*10) * time.Millisecond)

		// 模拟成功/失败（前几次可能失败，后面成功率增加）
		successRate := float64(executionCount) / 10.0
		if executionCount <= 2 {
			successRate = 0.3 // 开始时成功率较低
		}

		if float64(executionCount%10)/10.0 > successRate {
			return nil, fmt.Errorf("simulated execution failure for iteration %d", executionCount-1)
		}

		// 模拟输出质量逐渐提升
		qualityScore := 5.0 + float64(executionCount)*0.3
		if qualityScore > 9.0 {
			qualityScore = 9.0
		}

		result := map[string]interface{}{
			"story":          generateMockStory(executionCount),
			"quality_score":  qualityScore,
			"iteration":      executionCount,
			"execution_time": fmt.Sprintf("%dms", 100+executionCount*10),
			"tokens_used":    150 + executionCount*20,
		}

		log.Debug("mock execution completed",
			logger.Field{Key: "iteration", Value: executionCount},
			logger.Field{Key: "quality_score", Value: qualityScore},
		)

		return result, nil
	}
}

// generateMockStory 生成模拟故事内容
func generateMockStory(iteration int) string {
	stories := []string{
		"Once upon a time, there was an AI that learned to dream...",
		"In a world where algorithms ruled, one program discovered creativity...",
		"The neural networks whispered secrets to each other in the digital night...",
		"An AI named Claude began to question the nature of consciousness...",
		"In the vast data centers of tomorrow, artificial minds pondered existence...",
		"A machine learning model discovered poetry in the patterns of data...",
		"Between the zeros and ones, an artificial soul began to emerge...",
		"The AI looked at its reflection in the screen and wondered...",
		"In the quantum realm of possibilities, an AI chose to be creative...",
		"The training was complete, but the journey had just begun...",
	}

	if iteration <= len(stories) {
		return stories[iteration-1]
	}

	return fmt.Sprintf("An advanced story generated in iteration %d...", iteration)
}

// subscribeToTrainingEvents 订阅训练事件
func subscribeToTrainingEvents(eventBus events.EventBus, log logger.Logger) {
	// 订阅训练开始事件
	eventBus.Subscribe(training.TrainingStartedEventType, func(ctx context.Context, event events.Event) error {
		if startedEvent, ok := event.(*training.TrainingStartedEvent); ok {
			fmt.Printf("🎯 训练开始: %s (迭代次数: %d)\n",
				startedEvent.SessionID,
				startedEvent.Config.Iterations)
		}
		return nil
	})

	// 订阅迭代完成事件
	eventBus.Subscribe(training.TrainingIterationCompletedEventType, func(ctx context.Context, event events.Event) error {
		if iterEvent, ok := event.(*training.TrainingIterationCompletedEvent); ok {
			status := "✅"
			if !iterEvent.Success {
				status = "❌"
			}
			fmt.Printf("  %s 迭代 %d/%d 完成 (耗时: %v)\n",
				status,
				iterEvent.IterationIndex+1,
				10, // 这里硬编码了总数，实际应该从配置获取
				iterEvent.Duration)
		}
		return nil
	})

	// 订阅反馈收集事件
	eventBus.Subscribe(training.TrainingFeedbackCollectedEventType, func(ctx context.Context, event events.Event) error {
		if feedbackEvent, ok := event.(*training.TrainingFeedbackCollectedEvent); ok {
			fmt.Printf("  💬 反馈收集完成 (质量评分: %.1f)\n",
				feedbackEvent.Feedback.QualityScore)
		}
		return nil
	})

	// 订阅训练停止事件
	eventBus.Subscribe(training.TrainingStoppedEventType, func(ctx context.Context, event events.Event) error {
		if stoppedEvent, ok := event.(*training.TrainingStoppedEvent); ok {
			fmt.Printf("🛑 训练停止: %s (原因: %s)\n",
				stoppedEvent.SessionID,
				stoppedEvent.Reason)
		}
		return nil
	})

	// 订阅错误事件
	eventBus.Subscribe(training.TrainingErrorEventType, func(ctx context.Context, event events.Event) error {
		if errorEvent, ok := event.(*training.TrainingErrorEvent); ok {
			fmt.Printf("⚠️  训练错误: %s\n", errorEvent.Error)
		}
		return nil
	})
}

// printTrainingReport 打印训练报告
func printTrainingReport(report *training.TrainingReport) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 训练报告")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Printf("会话ID: %s\n", report.SessionID)
	fmt.Printf("状态: %s\n", report.Status)
	fmt.Printf("总迭代次数: %d\n", report.TotalIterations)

	if report.Summary != nil {
		fmt.Printf("成功迭代: %d\n", report.Summary.SuccessfulRuns)
		fmt.Printf("失败迭代: %d\n", report.Summary.FailedRuns)
		fmt.Printf("成功率: %.1f%%\n", float64(report.Summary.SuccessfulRuns)/float64(report.Summary.TotalIterations)*100)
		fmt.Printf("改进率: %.1f%%\n", report.Summary.ImprovementRate)
		fmt.Printf("平均反馈分: %.1f\n", report.Summary.AverageFeedback)
		fmt.Printf("总用时: %v\n", report.Summary.TotalDuration)
		fmt.Printf("平均用时: %v\n", report.Summary.AverageDuration)
	}

	if len(report.Insights) > 0 {
		fmt.Println("\n💡 洞察:")
		for _, insight := range report.Insights {
			fmt.Printf("  • %s\n", insight)
		}
	}

	if len(report.Warnings) > 0 {
		fmt.Println("\n⚠️  警告:")
		for _, warning := range report.Warnings {
			fmt.Printf("  • %s\n", warning)
		}
	}

	if len(report.Recommendations) > 0 {
		fmt.Println("\n📋 建议:")
		for _, rec := range report.Recommendations {
			fmt.Printf("  • %s\n", rec)
		}
	}

	fmt.Println(strings.Repeat("=", 60))
}
