package training

import (
	"context"
	"fmt"
	"time"

	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// TrainingUtils 训练工具包
type TrainingUtils struct {
	logger logger.Logger
}

// NewTrainingUtils 创建训练工具包
func NewTrainingUtils(logger logger.Logger) *TrainingUtils {
	return &TrainingUtils{
		logger: logger,
	}
}

// CreateTrainingHandler 创建训练处理器的工厂函数
func (tu *TrainingUtils) CreateTrainingHandler(eventBus events.EventBus, logger logger.Logger) TrainingHandler {
	return NewCrewTrainingHandler(eventBus, logger)
}

// RunTrainingSession 运行完整的训练会话
func (tu *TrainingUtils) RunTrainingSession(
	ctx context.Context,
	handler TrainingHandler,
	config *TrainingConfig,
	executeFunc func(context.Context, map[string]interface{}) (interface{}, error),
) (*TrainingSummary, error) {

	tu.logger.Info("starting training session",
		logger.Field{Key: "iterations", Value: config.Iterations},
		logger.Field{Key: "filename", Value: config.Filename},
	)

	// 启动训练
	if err := handler.StartTraining(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to start training: %w", err)
	}

	// 获取训练处理器的具体实现
	trainingHandler, ok := handler.(*CrewTrainingHandler)
	if !ok {
		return nil, fmt.Errorf("unsupported training handler type")
	}

	var recentScores []float64

	// 执行训练迭代
	for i := 0; i < config.Iterations; i++ {
		select {
		case <-ctx.Done():
			tu.logger.Info("training cancelled by context")
			handler.StopTraining(ctx)
			return nil, ctx.Err()
		default:
		}

		// 执行迭代
		iteration, err := trainingHandler.ExecuteIteration(ctx, executeFunc, i)
		if err != nil {
			tu.logger.Error("training iteration failed",
				logger.Field{Key: "iteration", Value: i},
				logger.Field{Key: "error", Value: err},
			)
			continue
		}

		// 收集分数用于早停检查
		if iteration.Feedback != nil {
			recentScores = append(recentScores, iteration.Feedback.QualityScore)
		} else if iteration.Metrics != nil {
			recentScores = append(recentScores, iteration.Metrics.AverageScore)
		} else {
			// 基于成功状态的默认分数
			if iteration.Success {
				recentScores = append(recentScores, 7.0)
			} else {
				recentScores = append(recentScores, 3.0)
			}
		}

		// 保持最近的分数窗口
		if len(recentScores) > config.PatientceEpochs {
			recentScores = recentScores[len(recentScores)-config.PatientceEpochs:]
		}

		// 检查早停条件
		if trainingHandler.CheckEarlyStop(recentScores) {
			tu.logger.Info("early stopping triggered",
				logger.Field{Key: "iteration", Value: i},
				logger.Field{Key: "recent_scores", Value: recentScores},
			)
			break
		}

		// 显示进度
		progress := float64(i+1) / float64(config.Iterations) * 100
		tu.logger.Info("training progress",
			logger.Field{Key: "iteration", Value: i + 1},
			logger.Field{Key: "total", Value: config.Iterations},
			logger.Field{Key: "progress", Value: fmt.Sprintf("%.1f%%", progress)},
		)
	}

	// 停止训练并获取总结
	if err := handler.StopTraining(ctx); err != nil {
		tu.logger.Error("failed to stop training gracefully",
			logger.Field{Key: "error", Value: err})
	}

	// 获取训练数据和总结
	trainingData := trainingHandler.trainingData
	if trainingData != nil && trainingData.Summary != nil {
		tu.logger.Info("training session completed",
			logger.Field{Key: "total_iterations", Value: trainingData.Summary.TotalIterations},
			logger.Field{Key: "success_rate", Value: trainingData.Summary.SuccessfulRuns},
			logger.Field{Key: "improvement_rate", Value: trainingData.Summary.ImprovementRate},
		)

		return trainingData.Summary, nil
	}

	return nil, fmt.Errorf("no training summary available")
}

// ValidateTrainingConfig 验证训练配置
func (tu *TrainingUtils) ValidateTrainingConfig(config *TrainingConfig) error {
	if config == nil {
		return fmt.Errorf("training config cannot be nil")
	}

	if config.Iterations <= 0 {
		return fmt.Errorf("iterations must be positive, got %d", config.Iterations)
	}

	if config.Iterations > 1000 {
		return fmt.Errorf("iterations too large, maximum is 1000, got %d", config.Iterations)
	}

	if config.Filename == "" {
		config.Filename = fmt.Sprintf("training_data_%s.json", time.Now().Format("20060102_150405"))
		tu.logger.Info("using default filename", logger.Field{Key: "filename", Value: config.Filename})
	}

	if config.LearningRate <= 0 || config.LearningRate > 1 {
		config.LearningRate = 0.001 // 默认学习率
		tu.logger.Info("using default learning rate", logger.Field{Key: "learning_rate", Value: config.LearningRate})
	}

	if config.BatchSize <= 0 {
		config.BatchSize = 1 // 默认批大小
	}

	if config.ValidationSplit < 0 || config.ValidationSplit >= 1 {
		config.ValidationSplit = 0.2 // 默认验证集比例
	}

	if config.FeedbackTimeout <= 0 {
		config.FeedbackTimeout = 5 * time.Minute // 默认反馈超时
	}

	if config.PatientceEpochs <= 0 {
		config.PatientceEpochs = 3 // 默认耐心轮数
	}

	if config.MinImprovement <= 0 {
		config.MinImprovement = 0.01 // 默认最小改进
	}

	if config.SaveInterval <= 0 {
		config.SaveInterval = 5 // 默认保存间隔
	}

	if config.BackupCount < 0 {
		config.BackupCount = 3 // 默认备份数量
	}

	if len(config.TargetMetrics) == 0 {
		config.TargetMetrics = []string{"execution_time", "success_rate", "feedback_score"}
	}

	return nil
}

// GenerateTrainingReport 生成训练报告
func (tu *TrainingUtils) GenerateTrainingReport(data *TrainingData) *TrainingReport {
	if data == nil || len(data.Iterations) == 0 {
		return &TrainingReport{
			SessionID: "unknown",
			Status:    "no_data",
			Message:   "No training data available",
		}
	}

	report := &TrainingReport{
		SessionID:       data.SessionID,
		CrewName:        data.CrewName,
		CreatedAt:       data.CreatedAt,
		UpdatedAt:       data.UpdatedAt,
		TotalIterations: len(data.Iterations),
		Config:          data.Config,
		Summary:         data.Summary,
		Status:          "completed",
		Insights:        make([]string, 0),
		Warnings:        make([]string, 0),
		Recommendations: make([]string, 0),
	}

	// 生成洞察
	if data.Summary != nil {
		if data.Summary.ImprovementRate > 10 {
			report.Insights = append(report.Insights,
				fmt.Sprintf("Significant improvement of %.1f%% achieved", data.Summary.ImprovementRate))
		}

		if data.Summary.SuccessfulRuns == data.Summary.TotalIterations {
			report.Insights = append(report.Insights, "Perfect success rate achieved")
		}

		if data.Summary.AverageFeedback > 8 {
			report.Insights = append(report.Insights, "High feedback scores indicate excellent output quality")
		}
	}

	// 生成警告
	failureRate := float64(data.Summary.FailedRuns) / float64(data.Summary.TotalIterations)
	if failureRate > 0.2 {
		report.Warnings = append(report.Warnings,
			fmt.Sprintf("High failure rate of %.1f%% detected", failureRate*100))
	}

	if data.Summary.AverageFeedback < 5 {
		report.Warnings = append(report.Warnings, "Low feedback scores indicate potential quality issues")
	}

	// 生成建议
	if data.Summary != nil && len(data.Summary.Recommendations) > 0 {
		report.Recommendations = append(report.Recommendations, data.Summary.Recommendations...)
	}

	if failureRate > 0.1 {
		report.Recommendations = append(report.Recommendations, "Consider reviewing task configuration and agent prompts")
	}

	if data.Summary.ImprovementRate < 5 {
		report.Recommendations = append(report.Recommendations, "Consider adjusting training parameters for better improvement")
	}

	return report
}

// TrainingReport 训练报告
type TrainingReport struct {
	SessionID       string           `json:"session_id"`
	CrewName        string           `json:"crew_name"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	TotalIterations int              `json:"total_iterations"`
	Config          *TrainingConfig  `json:"config"`
	Summary         *TrainingSummary `json:"summary"`
	Status          string           `json:"status"`
	Message         string           `json:"message,omitempty"`
	Insights        []string         `json:"insights"`
	Warnings        []string         `json:"warnings"`
	Recommendations []string         `json:"recommendations"`
}

// CreateSimpleTrainingConfig 创建简单的训练配置
func CreateSimpleTrainingConfig(iterations int, filename string) *TrainingConfig {
	return &TrainingConfig{
		Iterations:      iterations,
		Filename:        filename,
		Inputs:          make(map[string]interface{}),
		LearningRate:    0.001,
		BatchSize:       1,
		ValidationSplit: 0.2,
		CollectFeedback: true,
		FeedbackTimeout: 5 * time.Minute,
		FeedbackPrompts: []string{"Please rate the quality (1-10):", "Any suggestions?"},
		MetricsEnabled:  true,
		MetricsInterval: 1,
		TargetMetrics:   []string{"execution_time", "success_rate", "feedback_score"},
		EarlyStopping:   false,
		PatientceEpochs: 3,
		MinImprovement:  0.01,
		SaveInterval:    5,
		Verbose:         true,
		AutoSave:        true,
		BackupCount:     3,
	}
}

// CreateAdvancedTrainingConfig 创建高级训练配置
func CreateAdvancedTrainingConfig(iterations int, filename string, inputs map[string]interface{}) *TrainingConfig {
	config := CreateSimpleTrainingConfig(iterations, filename)
	config.Inputs = inputs
	config.EarlyStopping = true
	config.CollectFeedback = true
	config.MetricsEnabled = true
	config.FeedbackTimeout = 3 * time.Minute
	config.SaveInterval = 3

	// 更详细的反馈提示
	config.FeedbackPrompts = []string{
		"Rate overall quality (1-10):",
		"Rate accuracy (1-10):",
		"Rate usefulness (1-10):",
		"What could be improved?",
		"Any specific issues?",
	}

	return config
}
