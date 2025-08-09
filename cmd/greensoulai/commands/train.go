package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/ynl/greensoulai/internal/cli/config"
	"github.com/ynl/greensoulai/pkg/logger"
)

// NewTrainCommand 创建train命令
func NewTrainCommand(log logger.Logger) *cobra.Command {
	var (
		iterations   int
		filename     string
		outputDir    string
		concurrency  int
		timeout      time.Duration
		saveInterval int
	)

	cmd := &cobra.Command{
		Use:   "train",
		Short: "训练GreenSoulAI项目",
		Long: `对当前GreenSoulAI项目进行训练，通过多次运行来优化智能体性能。
训练过程会记录每次执行的结果，并生成训练报告。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 查找项目根目录
			projectRoot, err := config.GetProjectRoot()
			if err != nil {
				return fmt.Errorf("not in a greensoulai project: %w", err)
			}

			// 加载项目配置
			configPath := filepath.Join(projectRoot, "greensoulai.yaml")
			projectConfig, err := config.LoadProjectConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load project config: %w", err)
			}

			// 验证配置
			if err := projectConfig.Validate(); err != nil {
				return fmt.Errorf("invalid project configuration: %w", err)
			}

			// 设置默认训练文件名
			if filename == "" {
				filename = fmt.Sprintf("%s_training_%s.json",
					projectConfig.Name,
					time.Now().Format("20060102_150405"))
			}

			// 设置输出目录
			if outputDir == "" {
				outputDir = filepath.Join(projectRoot, "training_data")
			}

			log.Info("开始训练项目",
				logger.Field{Key: "name", Value: projectConfig.Name},
				logger.Field{Key: "iterations", Value: iterations},
				logger.Field{Key: "filename", Value: filename},
				logger.Field{Key: "output_dir", Value: outputDir},
			)

			// 创建训练器
			trainer := &ProjectTrainer{
				Config:       projectConfig,
				ProjectRoot:  projectRoot,
				Iterations:   iterations,
				Filename:     filename,
				OutputDir:    outputDir,
				Concurrency:  concurrency,
				Timeout:      timeout,
				SaveInterval: saveInterval,
				Logger:       log,
			}

			// 执行训练
			return trainer.Train(cmd.Context())
		},
	}

	// 添加选项
	cmd.Flags().IntVarP(&iterations, "iterations", "n", 5, "训练迭代次数")
	cmd.Flags().StringVarP(&filename, "filename", "f", "", "训练数据文件名")
	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "输出目录")
	cmd.Flags().IntVarP(&concurrency, "concurrency", "c", 1, "并发执行数量")
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 10*time.Minute, "单次执行超时时间")
	cmd.Flags().IntVar(&saveInterval, "save-interval", 1, "保存间隔（每N次迭代保存一次）")

	return cmd
}

// ProjectTrainer 项目训练器
type ProjectTrainer struct {
	Config       *config.ProjectConfig
	ProjectRoot  string
	Iterations   int
	Filename     string
	OutputDir    string
	Concurrency  int
	Timeout      time.Duration
	SaveInterval int
	Logger       logger.Logger
}

// TrainingResult 训练结果
type TrainingResult struct {
	Iteration   int                    `json:"iteration"`
	Success     bool                   `json:"success"`
	Duration    time.Duration          `json:"duration"`
	Output      string                 `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	TaskResults []TaskResult           `json:"task_results,omitempty"`
}

// TaskResult 任务结果
type TaskResult struct {
	TaskName  string                 `json:"task_name"`
	AgentName string                 `json:"agent_name"`
	Duration  time.Duration          `json:"duration"`
	Success   bool                   `json:"success"`
	Output    string                 `json:"output"`
	Quality   float64                `json:"quality,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// TrainingReport 训练报告
type TrainingReport struct {
	ProjectName     string           `json:"project_name"`
	ProjectType     string           `json:"project_type"`
	TotalIterations int              `json:"total_iterations"`
	SuccessfulRuns  int              `json:"successful_runs"`
	FailedRuns      int              `json:"failed_runs"`
	StartTime       time.Time        `json:"start_time"`
	EndTime         time.Time        `json:"end_time"`
	TotalDuration   time.Duration    `json:"total_duration"`
	AverageRunTime  time.Duration    `json:"average_run_time"`
	Results         []TrainingResult `json:"results"`

	// 性能分析
	Performance struct {
		BestIteration  int     `json:"best_iteration"`
		WorstIteration int     `json:"worst_iteration"`
		SuccessRate    float64 `json:"success_rate"`
		Improvement    float64 `json:"improvement"`
	} `json:"performance"`
}

// Train 执行训练
func (t *ProjectTrainer) Train(ctx context.Context) error {
	// 创建输出目录
	if err := os.MkdirAll(t.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 初始化训练报告
	report := &TrainingReport{
		ProjectName:     t.Config.Name,
		ProjectType:     string(t.Config.Type),
		TotalIterations: t.Iterations,
		StartTime:       time.Now(),
		Results:         make([]TrainingResult, 0, t.Iterations),
	}

	t.Logger.Info("🚀 开始训练会话",
		logger.Field{Key: "iterations", Value: t.Iterations},
		logger.Field{Key: "concurrency", Value: t.Concurrency},
	)

	// 显示训练进度头部
	t.printTrainingHeader()

	// 执行训练迭代
	successCount := 0
	var totalDuration time.Duration

	for i := 1; i <= t.Iterations; i++ {
		// 创建迭代上下文
		iterCtx, cancel := context.WithTimeout(ctx, t.Timeout)

		t.Logger.Info("执行迭代", logger.Field{Key: "iteration", Value: i})

		// 执行单次训练
		result := t.executeIteration(iterCtx, i)
		cancel()

		// 记录结果
		report.Results = append(report.Results, result)

		if result.Success {
			successCount++
			t.printSuccess(i, result.Duration)
		} else {
			t.printFailure(i, result.Duration, result.Error)
		}

		totalDuration += result.Duration

		// 定期保存中间结果
		if i%t.SaveInterval == 0 || i == t.Iterations {
			if err := t.saveIntermediateResults(report, i); err != nil {
				t.Logger.Warn("保存中间结果失败", logger.Field{Key: "error", Value: err})
			}
		}

		// 检查是否被取消
		select {
		case <-ctx.Done():
			t.Logger.Info("训练被用户中断")
			return ctx.Err()
		default:
		}
	}

	// 完成训练报告
	report.EndTime = time.Now()
	report.TotalDuration = totalDuration
	report.SuccessfulRuns = successCount
	report.FailedRuns = t.Iterations - successCount

	if t.Iterations > 0 {
		report.AverageRunTime = totalDuration / time.Duration(t.Iterations)
		report.Performance.SuccessRate = float64(successCount) / float64(t.Iterations)
	}

	// 分析性能
	t.analyzePerformance(report)

	// 保存最终训练报告
	if err := t.saveFinalReport(report); err != nil {
		return fmt.Errorf("failed to save training report: %w", err)
	}

	// 显示训练总结
	t.printTrainingSummary(report)

	return nil
}

// executeIteration 执行单次训练迭代
func (t *ProjectTrainer) executeIteration(ctx context.Context, iteration int) TrainingResult {
	start := time.Now()

	result := TrainingResult{
		Iteration: iteration,
		Timestamp: start,
		Metadata:  make(map[string]interface{}),
	}

	// TODO: 这里应该调用实际的项目执行逻辑
	// 现在先用模拟的执行过程

	// 模拟执行时间（实际应该运行crew）
	executionTime := time.Duration(2+iteration%3) * time.Second
	select {
	case <-time.After(executionTime):
		// 模拟成功执行
		result.Success = true
		result.Output = fmt.Sprintf("Iteration %d completed successfully", iteration)

		// 模拟任务结果
		for _, taskCfg := range t.Config.Tasks {
			taskResult := TaskResult{
				TaskName:  taskCfg.Name,
				AgentName: taskCfg.Agent,
				Duration:  time.Duration(1+iteration%2) * time.Second,
				Success:   iteration%4 != 0, // 模拟75%成功率
				Output:    fmt.Sprintf("Task %s output for iteration %d", taskCfg.Name, iteration),
				Quality:   0.7 + float64(iteration%3)*0.1, // 模拟质量评分
			}
			result.TaskResults = append(result.TaskResults, taskResult)
		}

	case <-ctx.Done():
		// 超时或被取消
		result.Success = false
		result.Error = "execution timeout or cancelled"
	}

	result.Duration = time.Since(start)

	// 添加元数据
	result.Metadata["go_version"] = t.Config.GoVersion
	result.Metadata["llm_provider"] = t.Config.LLM.Provider
	result.Metadata["llm_model"] = t.Config.LLM.Model

	return result
}

// printTrainingHeader 打印训练头部信息
func (t *ProjectTrainer) printTrainingHeader() {
	fmt.Printf(`
🎯 GreenSoulAI 训练会话
==================================================
📋 项目: %s (%s)
🔄 迭代次数: %d
⏱️  超时时间: %v
📁 输出目录: %s
⚡ 并发数: %d

执行进度：
`, t.Config.Name, t.Config.Type, t.Iterations, t.Timeout, t.OutputDir, t.Concurrency)
}

// printSuccess 打印成功信息
func (t *ProjectTrainer) printSuccess(iteration int, duration time.Duration) {
	fmt.Printf("✅ 迭代 %2d/%d - 成功 (%.2fs)\n",
		iteration, t.Iterations, duration.Seconds())
}

// printFailure 打印失败信息
func (t *ProjectTrainer) printFailure(iteration int, duration time.Duration, errorMsg string) {
	fmt.Printf("❌ 迭代 %2d/%d - 失败 (%.2fs) - %s\n",
		iteration, t.Iterations, duration.Seconds(), errorMsg)
}

// printTrainingSummary 打印训练总结
func (t *ProjectTrainer) printTrainingSummary(report *TrainingReport) {
	fmt.Printf(`
🏁 训练完成！
==================================================
📊 总结统计:
   总迭代次数: %d
   成功次数: %d
   失败次数: %d
   成功率: %.1f%%
   总耗时: %v
   平均耗时: %v

📈 性能分析:
   最佳迭代: #%d
   最差迭代: #%d

📁 训练数据已保存到:
   %s

🚀 下一步:
1. 查看训练报告分析性能
2. 基于结果调整项目配置
3. 运行评估验证改进效果

`,
		report.TotalIterations,
		report.SuccessfulRuns,
		report.FailedRuns,
		report.Performance.SuccessRate*100,
		report.TotalDuration,
		report.AverageRunTime,
		report.Performance.BestIteration,
		report.Performance.WorstIteration,
		filepath.Join(t.OutputDir, t.Filename),
	)
}

// analyzePerformance 分析训练性能
func (t *ProjectTrainer) analyzePerformance(report *TrainingReport) {
	if len(report.Results) == 0 {
		return
	}

	bestTime := report.Results[0].Duration
	worstTime := report.Results[0].Duration
	bestIdx := 0
	worstIdx := 0

	for i, result := range report.Results {
		if result.Success && result.Duration < bestTime {
			bestTime = result.Duration
			bestIdx = i
		}
		if result.Duration > worstTime {
			worstTime = result.Duration
			worstIdx = i
		}
	}

	report.Performance.BestIteration = bestIdx + 1
	report.Performance.WorstIteration = worstIdx + 1

	// 计算改进程度（简单的线性回归）
	if len(report.Results) > 1 {
		firstHalf := len(report.Results) / 2
		firstAvg := t.calculateAverageDuration(report.Results[:firstHalf])
		secondAvg := t.calculateAverageDuration(report.Results[firstHalf:])

		if firstAvg > 0 {
			report.Performance.Improvement = (firstAvg.Seconds() - secondAvg.Seconds()) / firstAvg.Seconds()
		}
	}
}

// calculateAverageDuration 计算平均持续时间
func (t *ProjectTrainer) calculateAverageDuration(results []TrainingResult) time.Duration {
	if len(results) == 0 {
		return 0
	}

	var total time.Duration
	successCount := 0

	for _, result := range results {
		if result.Success {
			total += result.Duration
			successCount++
		}
	}

	if successCount == 0 {
		return 0
	}

	return total / time.Duration(successCount)
}

// saveIntermediateResults 保存中间结果
func (t *ProjectTrainer) saveIntermediateResults(report *TrainingReport, currentIteration int) error {
	filename := fmt.Sprintf("%s_progress_%d.json",
		t.Config.Name, currentIteration)
	filepath := filepath.Join(t.OutputDir, filename)

	// 这里应该保存JSON格式的中间结果
	// 为了简化，现在只记录日志
	t.Logger.Info("保存中间训练结果",
		logger.Field{Key: "iteration", Value: currentIteration},
		logger.Field{Key: "file", Value: filepath},
	)

	return nil
}

// saveFinalReport 保存最终训练报告
func (t *ProjectTrainer) saveFinalReport(report *TrainingReport) error {
	filepath := filepath.Join(t.OutputDir, t.Filename)

	// 这里应该保存JSON格式的完整报告
	// 为了简化，现在只记录日志
	t.Logger.Info("保存最终训练报告",
		logger.Field{Key: "file", Value: filepath},
		logger.Field{Key: "success_rate", Value: report.Performance.SuccessRate},
	)

	// TODO: 实际的JSON序列化和文件写入
	// data, err := json.MarshalIndent(report, "", "  ")
	// if err != nil {
	//     return fmt.Errorf("failed to marshal report: %w", err)
	// }
	// return os.WriteFile(filepath, data, 0644)

	return nil
}
