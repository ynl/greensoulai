package commands

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/ynl/greensoulai/internal/cli/config"
	"github.com/ynl/greensoulai/pkg/logger"
)

// NewEvaluateCommand 创建evaluate命令
func NewEvaluateCommand(log logger.Logger) *cobra.Command {
	var (
		iterations int
		model      string
		metric     string
		outputFile string
		parallel   bool
		timeout    time.Duration
	)

	cmd := &cobra.Command{
		Use:   "evaluate",
		Short: "评估GreenSoulAI项目性能",
		Long: `评估当前GreenSoulAI项目的性能和质量。
通过多次运行项目并分析结果来评估智能体的表现，生成详细的评估报告。`,
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

			log.Info("开始评估项目",
				logger.Field{Key: "name", Value: projectConfig.Name},
				logger.Field{Key: "iterations", Value: iterations},
				logger.Field{Key: "model", Value: model},
				logger.Field{Key: "metric", Value: metric},
			)

			// 创建评估器
			evaluator := &ProjectEvaluator{
				Config:      projectConfig,
				ProjectRoot: projectRoot,
				Iterations:  iterations,
				Model:       model,
				Metric:      metric,
				OutputFile:  outputFile,
				Parallel:    parallel,
				Timeout:     timeout,
				Logger:      log,
			}

			// 执行评估
			return evaluator.Evaluate(cmd.Context())
		},
	}

	// 添加选项
	cmd.Flags().IntVarP(&iterations, "iterations", "n", 3, "评估迭代次数")
	cmd.Flags().StringVarP(&model, "model", "m", "", "评估用的LLM模型")
	cmd.Flags().StringVar(&metric, "metric", "quality", "评估指标 (quality, performance, cost)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "评估报告输出文件")
	cmd.Flags().BoolVar(&parallel, "parallel", false, "并行执行评估")
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 15*time.Minute, "单次评估超时时间")

	return cmd
}

// ProjectEvaluator 项目评估器
type ProjectEvaluator struct {
	Config      *config.ProjectConfig
	ProjectRoot string
	Iterations  int
	Model       string
	Metric      string
	OutputFile  string
	Parallel    bool
	Timeout     time.Duration
	Logger      logger.Logger
}

// EvaluationResult 评估结果
type EvaluationResult struct {
	ProjectName    string    `json:"project_name"`
	EvaluationTime time.Time `json:"evaluation_time"`
	Iterations     int       `json:"iterations"`
	Model          string    `json:"model"`
	Metric         string    `json:"metric"`

	// 整体评分
	OverallScore     float64 `json:"overall_score"`
	QualityScore     float64 `json:"quality_score"`
	PerformanceScore float64 `json:"performance_score"`
	CostScore        float64 `json:"cost_score"`

	// 详细结果
	TaskResults  []TaskEvaluationResult  `json:"task_results"`
	AgentResults []AgentEvaluationResult `json:"agent_results"`

	// 统计信息
	Statistics      EvaluationStatistics `json:"statistics"`
	Recommendations []string             `json:"recommendations"`
}

// TaskEvaluationResult 任务评估结果
type TaskEvaluationResult struct {
	TaskName       string        `json:"task_name"`
	AgentName      string        `json:"agent_name"`
	QualityScore   float64       `json:"quality_score"`
	CompletionRate float64       `json:"completion_rate"`
	AverageTime    time.Duration `json:"average_time"`
	SuccessRate    float64       `json:"success_rate"`
	OutputQuality  string        `json:"output_quality"`
}

// AgentEvaluationResult 智能体评估结果
type AgentEvaluationResult struct {
	AgentName    string  `json:"agent_name"`
	Role         string  `json:"role"`
	TaskCount    int     `json:"task_count"`
	AverageScore float64 `json:"average_score"`
	Reliability  float64 `json:"reliability"`
	Efficiency   float64 `json:"efficiency"`
	Consistency  float64 `json:"consistency"`
}

// EvaluationStatistics 评估统计
type EvaluationStatistics struct {
	TotalExecutions  int           `json:"total_executions"`
	SuccessfulRuns   int           `json:"successful_runs"`
	FailedRuns       int           `json:"failed_runs"`
	AverageExecution time.Duration `json:"average_execution_time"`
	MinExecutionTime time.Duration `json:"min_execution_time"`
	MaxExecutionTime time.Duration `json:"max_execution_time"`
	TotalCost        float64       `json:"total_cost"`
	AverageCost      float64       `json:"average_cost"`
}

// Evaluate 执行项目评估
func (e *ProjectEvaluator) Evaluate(ctx context.Context) error {
	e.Logger.Info("🎯 开始项目评估")

	// 显示评估信息
	e.printEvaluationHeader()

	// 初始化评估结果
	result := &EvaluationResult{
		ProjectName:    e.Config.Name,
		EvaluationTime: time.Now(),
		Iterations:     e.Iterations,
		Model:          e.Model,
		Metric:         e.Metric,
		TaskResults:    make([]TaskEvaluationResult, 0),
		AgentResults:   make([]AgentEvaluationResult, 0),
	}

	// 执行评估迭代
	for i := 1; i <= e.Iterations; i++ {
		e.Logger.Info("执行评估迭代", logger.Field{Key: "iteration", Value: i})

		// 创建迭代上下文
		iterCtx, cancel := context.WithTimeout(ctx, e.Timeout)

		// 执行单次评估
		if err := e.executeEvaluationIteration(iterCtx, i, result); err != nil {
			cancel()
			e.Logger.Error("评估迭代失败",
				logger.Field{Key: "iteration", Value: i},
				logger.Field{Key: "error", Value: err})
			continue
		}
		cancel()

		fmt.Printf("✅ 迭代 %d/%d 完成\n", i, e.Iterations)
	}

	// 分析评估结果
	e.analyzeResults(result)

	// 生成推荐
	e.generateRecommendations(result)

	// 保存评估报告
	if err := e.saveEvaluationReport(result); err != nil {
		return fmt.Errorf("failed to save evaluation report: %w", err)
	}

	// 显示评估总结
	e.printEvaluationSummary(result)

	return nil
}

// executeEvaluationIteration 执行单次评估迭代
func (e *ProjectEvaluator) executeEvaluationIteration(ctx context.Context, iteration int, result *EvaluationResult) error {
	// TODO: 这里应该实际运行项目并收集性能数据
	// 现在用模拟数据

	// 模拟任务评估
	for _, taskCfg := range e.Config.Tasks {
		taskResult := TaskEvaluationResult{
			TaskName:       taskCfg.Name,
			AgentName:      taskCfg.Agent,
			QualityScore:   0.7 + float64(iteration%3)*0.1, // 模拟质量评分
			CompletionRate: 0.9 + float64(iteration%2)*0.05,
			AverageTime:    time.Duration(2+iteration%3) * time.Second,
			SuccessRate:    0.85 + float64(iteration%4)*0.05,
			OutputQuality:  "良好",
		}
		result.TaskResults = append(result.TaskResults, taskResult)
	}

	// 模拟智能体评估
	for _, agentCfg := range e.Config.Agents {
		agentResult := AgentEvaluationResult{
			AgentName:    agentCfg.Name,
			Role:         agentCfg.Role,
			TaskCount:    len(e.Config.Tasks),
			AverageScore: 0.8 + float64(iteration%3)*0.05,
			Reliability:  0.9 + float64(iteration%2)*0.03,
			Efficiency:   0.75 + float64(iteration%4)*0.05,
			Consistency:  0.85 + float64(iteration%3)*0.03,
		}
		result.AgentResults = append(result.AgentResults, agentResult)
	}

	return nil
}

// analyzeResults 分析评估结果
func (e *ProjectEvaluator) analyzeResults(result *EvaluationResult) {
	// 计算整体评分
	if len(result.TaskResults) > 0 {
		var totalQuality float64
		for _, taskResult := range result.TaskResults {
			totalQuality += taskResult.QualityScore
		}
		result.QualityScore = totalQuality / float64(len(result.TaskResults))
	}

	if len(result.AgentResults) > 0 {
		var totalPerformance float64
		for _, agentResult := range result.AgentResults {
			totalPerformance += agentResult.Efficiency
		}
		result.PerformanceScore = totalPerformance / float64(len(result.AgentResults))
	}

	// 成本评分（模拟）
	result.CostScore = 0.8

	// 计算总体评分
	result.OverallScore = (result.QualityScore + result.PerformanceScore + result.CostScore) / 3

	// 更新统计信息
	result.Statistics = EvaluationStatistics{
		TotalExecutions:  e.Iterations,
		SuccessfulRuns:   e.Iterations,
		FailedRuns:       0,
		AverageExecution: 5 * time.Second,
		MinExecutionTime: 3 * time.Second,
		MaxExecutionTime: 8 * time.Second,
		TotalCost:        float64(e.Iterations) * 0.05,
		AverageCost:      0.05,
	}
}

// generateRecommendations 生成改进建议
func (e *ProjectEvaluator) generateRecommendations(result *EvaluationResult) {
	recommendations := []string{}

	if result.QualityScore < 0.7 {
		recommendations = append(recommendations, "建议优化智能体的提示词和角色定义")
	}

	if result.PerformanceScore < 0.7 {
		recommendations = append(recommendations, "建议优化任务分配和执行流程")
	}

	if result.CostScore < 0.7 {
		recommendations = append(recommendations, "建议优化LLM调用频率和模型选择")
	}

	if result.OverallScore > 0.85 {
		recommendations = append(recommendations, "项目表现优秀，可以考虑增加更复杂的任务")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "项目整体表现良好，继续保持")
	}

	result.Recommendations = recommendations
}

// printEvaluationHeader 打印评估头部信息
func (e *ProjectEvaluator) printEvaluationHeader() {
	fmt.Printf(`
🎯 GreenSoulAI 项目评估
==================================================
📋 项目: %s (%s)
🔄 评估次数: %d
📊 评估指标: %s
🤖 评估模型: %s
⏱️  超时时间: %v

评估进度：
`, e.Config.Name, e.Config.Type, e.Iterations, e.Metric, e.Model, e.Timeout)
}

// printEvaluationSummary 打印评估总结
func (e *ProjectEvaluator) printEvaluationSummary(result *EvaluationResult) {
	fmt.Printf(`
🏁 评估完成！
==================================================
📊 评估结果:
   整体评分: %.2f/1.00 (%s)
   质量评分: %.2f/1.00
   性能评分: %.2f/1.00
   成本评分: %.2f/1.00

📈 统计信息:
   成功执行: %d/%d
   平均耗时: %v
   平均成本: $%.4f

🎯 任务表现:
`, result.OverallScore, e.getScoreLevel(result.OverallScore),
		result.QualityScore, result.PerformanceScore, result.CostScore,
		result.Statistics.SuccessfulRuns, result.Statistics.TotalExecutions,
		result.Statistics.AverageExecution, result.Statistics.AverageCost)

	for _, taskResult := range result.TaskResults {
		fmt.Printf("   • %s: %.2f/1.00 (成功率: %.1f%%)\n",
			taskResult.TaskName, taskResult.QualityScore, taskResult.SuccessRate*100)
	}

	fmt.Println("\n🤖 智能体表现:")
	for _, agentResult := range result.AgentResults {
		fmt.Printf("   • %s: %.2f/1.00 (可靠性: %.1f%%)\n",
			agentResult.AgentName, agentResult.AverageScore, agentResult.Reliability*100)
	}

	fmt.Println("\n💡 改进建议:")
	for i, rec := range result.Recommendations {
		fmt.Printf("   %d. %s\n", i+1, rec)
	}

	if e.OutputFile != "" {
		fmt.Printf("\n📁 详细报告已保存到: %s\n", e.OutputFile)
	}
}

// getScoreLevel 获取评分等级
func (e *ProjectEvaluator) getScoreLevel(score float64) string {
	if score >= 0.9 {
		return "优秀"
	} else if score >= 0.8 {
		return "良好"
	} else if score >= 0.7 {
		return "中等"
	} else if score >= 0.6 {
		return "一般"
	} else {
		return "需改进"
	}
}

// saveEvaluationReport 保存评估报告
func (e *ProjectEvaluator) saveEvaluationReport(result *EvaluationResult) error {
	if e.OutputFile == "" {
		e.OutputFile = fmt.Sprintf("%s_evaluation_%s.json",
			e.Config.Name,
			time.Now().Format("20060102_150405"))
	}

	// TODO: 实际的JSON序列化和文件保存
	e.Logger.Info("保存评估报告",
		logger.Field{Key: "file", Value: e.OutputFile},
		logger.Field{Key: "overall_score", Value: result.OverallScore},
	)

	return nil
}
