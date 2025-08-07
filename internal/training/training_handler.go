package training

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// CrewTrainingHandler CrewAI训练处理器实现
type CrewTrainingHandler struct {
	// 依赖
	eventBus events.EventBus
	logger   logger.Logger

	// 训练状态
	status       *TrainingStatus
	config       *TrainingConfig
	trainingData *TrainingData

	// 控制信号
	stopChan chan bool
	statusMu sync.RWMutex
	dataMu   sync.RWMutex

	// 反馈收集
	feedbackCollector *FeedbackCollector
	metricsAnalyzer   *MetricsAnalyzer
}

// NewCrewTrainingHandler 创建新的训练处理器
func NewCrewTrainingHandler(eventBus events.EventBus, logger logger.Logger) *CrewTrainingHandler {
	return &CrewTrainingHandler{
		eventBus: eventBus,
		logger:   logger,
		status: &TrainingStatus{
			IsRunning: false,
			Status:    "initialized",
		},
		stopChan:          make(chan bool, 1),
		feedbackCollector: NewFeedbackCollector(logger),
		metricsAnalyzer:   NewMetricsAnalyzer(logger),
	}
}

// StartTraining 开始训练过程
func (th *CrewTrainingHandler) StartTraining(ctx context.Context, config *TrainingConfig) error {
	th.statusMu.Lock()
	if th.status.IsRunning {
		th.statusMu.Unlock()
		return fmt.Errorf("training is already running")
	}

	th.config = config
	th.status = &TrainingStatus{
		IsRunning:        true,
		CurrentIteration: 0,
		TotalIterations:  config.Iterations,
		Progress:         0.0,
		StartTime:        time.Now(),
		Status:           "starting",
	}
	th.statusMu.Unlock()

	// 初始化训练数据
	th.dataMu.Lock()
	th.trainingData = &TrainingData{
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Version:    "1.0",
		Config:     config,
		SessionID:  uuid.New().String(),
		TotalRuns:  0,
		Iterations: make([]*IterationData, 0, config.Iterations),
		Summary:    &TrainingSummary{},
	}
	th.dataMu.Unlock()

	// 发射训练开始事件
	startEvent := NewTrainingStartedEvent(th.trainingData.SessionID, config)
	if err := th.eventBus.Emit(ctx, th, startEvent); err != nil {
		th.logger.Error("failed to emit training started event", logger.Field{Key: "error", Value: err})
	}

	th.logger.Info("training started",
		logger.Field{Key: "session_id", Value: th.trainingData.SessionID},
		logger.Field{Key: "iterations", Value: config.Iterations},
		logger.Field{Key: "filename", Value: config.Filename},
	)

	return nil
}

// ExecuteIteration 执行单次训练迭代
func (th *CrewTrainingHandler) ExecuteIteration(ctx context.Context, executeFunc func(context.Context, map[string]interface{}) (interface{}, error), iterationIndex int) (*IterationData, error) {
	iterationID := uuid.New().String()
	startTime := time.Now()

	th.logger.Info("executing training iteration",
		logger.Field{Key: "iteration", Value: iterationIndex},
		logger.Field{Key: "iteration_id", Value: iterationID},
	)

	// 创建迭代数据
	iteration := &IterationData{
		IterationID: iterationID,
		Index:       iterationIndex,
		Timestamp:   startTime,
		Inputs:      th.config.Inputs,
		AgentData:   make([]*AgentIterationData, 0),
		TaskData:    make([]*TaskIterationData, 0),
	}

	// 更新状态
	th.statusMu.Lock()
	th.status.CurrentIteration = iterationIndex
	th.status.Progress = float64(iterationIndex) / float64(th.config.Iterations)
	th.status.Status = fmt.Sprintf("executing iteration %d", iterationIndex)
	th.statusMu.Unlock()

	// 发射迭代开始事件
	iterationStartEvent := NewTrainingIterationStartedEvent(th.trainingData.SessionID, iterationID, iterationIndex)
	th.eventBus.Emit(ctx, th, iterationStartEvent)

	// 执行传入的执行函数
	var outputs interface{}
	var err error

	if executeFunc != nil {
		outputs, err = executeFunc(ctx, th.config.Inputs)
	} else {
		// 默认模拟执行（当没有提供执行函数时）
		outputs = map[string]interface{}{
			"result":    "Training iteration completed (simulated)",
			"iteration": iterationIndex,
		}
		err = nil
	}

	duration := time.Since(startTime)
	iteration.Duration = duration
	iteration.Outputs = outputs
	iteration.Success = err == nil

	if err != nil {
		iteration.Error = err.Error()
		th.logger.Error("training iteration failed",
			logger.Field{Key: "iteration", Value: iterationIndex},
			logger.Field{Key: "error", Value: err},
		)
	}

	// 收集性能指标
	if th.config.MetricsEnabled && (iterationIndex%th.config.MetricsInterval == 0) {
		metrics, metricsErr := th.metricsAnalyzer.AnalyzeIteration(ctx, iteration)
		if metricsErr != nil {
			th.logger.Error("failed to analyze metrics",
				logger.Field{Key: "iteration", Value: iterationIndex},
				logger.Field{Key: "error", Value: metricsErr},
			)
		} else {
			iteration.Metrics = metrics
		}
	}

	// 收集人工反馈
	if th.config.CollectFeedback {
		feedback, feedbackErr := th.feedbackCollector.CollectFeedback(ctx, iterationID, outputs, th.config.FeedbackTimeout)
		if feedbackErr != nil {
			th.logger.Warn("failed to collect feedback",
				logger.Field{Key: "iteration", Value: iterationIndex},
				logger.Field{Key: "error", Value: feedbackErr},
			)
		} else if feedback != nil {
			iteration.Feedback = feedback
		}
	}

	// 保存迭代数据
	th.dataMu.Lock()
	th.trainingData.Iterations = append(th.trainingData.Iterations, iteration)
	th.trainingData.TotalRuns++
	th.trainingData.UpdatedAt = time.Now()
	th.dataMu.Unlock()

	// 自动保存
	if th.config.AutoSave && (iterationIndex%th.config.SaveInterval == 0) {
		if saveErr := th.SaveTrainingData(ctx, th.trainingData); saveErr != nil {
			th.logger.Error("failed to auto-save training data",
				logger.Field{Key: "iteration", Value: iterationIndex},
				logger.Field{Key: "error", Value: saveErr},
			)
		}
	}

	// 发射迭代完成事件
	iterationCompletedEvent := NewTrainingIterationCompletedEvent(
		th.trainingData.SessionID, iterationID, iterationIndex, duration, iteration.Success)
	th.eventBus.Emit(ctx, th, iterationCompletedEvent)

	th.logger.Info("training iteration completed",
		logger.Field{Key: "iteration", Value: iterationIndex},
		logger.Field{Key: "duration", Value: duration},
		logger.Field{Key: "success", Value: iteration.Success},
	)

	return iteration, nil
}

// CollectFeedback 收集人工反馈
func (th *CrewTrainingHandler) CollectFeedback(ctx context.Context, iterationID string, feedback *HumanFeedback) error {
	th.dataMu.Lock()
	defer th.dataMu.Unlock()

	// 查找对应的迭代
	for _, iteration := range th.trainingData.Iterations {
		if iteration.IterationID == iterationID {
			iteration.Feedback = feedback
			th.trainingData.UpdatedAt = time.Now()

			th.logger.Info("feedback collected",
				logger.Field{Key: "iteration_id", Value: iterationID},
				logger.Field{Key: "quality_score", Value: feedback.QualityScore},
			)

			// 发射反馈收集事件
			feedbackEvent := NewTrainingFeedbackCollectedEvent(th.trainingData.SessionID, iterationID, feedback)
			th.eventBus.Emit(ctx, th, feedbackEvent)

			return nil
		}
	}

	return fmt.Errorf("iteration not found: %s", iterationID)
}

// AnalyzePerformance 分析性能
func (th *CrewTrainingHandler) AnalyzePerformance(ctx context.Context, iterationID string) (*PerformanceMetrics, error) {
	th.dataMu.RLock()
	defer th.dataMu.RUnlock()

	// 查找对应的迭代
	for _, iteration := range th.trainingData.Iterations {
		if iteration.IterationID == iterationID {
			if iteration.Metrics != nil {
				return iteration.Metrics, nil
			}

			// 如果没有缓存的指标，重新分析
			metrics, err := th.metricsAnalyzer.AnalyzeIteration(ctx, iteration)
			if err != nil {
				return nil, fmt.Errorf("failed to analyze performance: %w", err)
			}

			return metrics, nil
		}
	}

	return nil, fmt.Errorf("iteration not found: %s", iterationID)
}

// SaveTrainingData 保存训练数据
func (th *CrewTrainingHandler) SaveTrainingData(ctx context.Context, data *TrainingData) error {
	th.dataMu.RLock()
	defer th.dataMu.RUnlock()

	filename := th.config.Filename
	if filename == "" {
		filename = fmt.Sprintf("training_data_%s.json", data.SessionID)
	}

	// 确保目录存在
	dir := filepath.Dir(filename)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// 备份现有文件
	if th.config.BackupCount > 0 {
		th.createBackup(filename)
	}

	// 序列化数据
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal training data: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write training data: %w", err)
	}

	th.logger.Info("training data saved",
		logger.Field{Key: "filename", Value: filename},
		logger.Field{Key: "iterations", Value: len(data.Iterations)},
	)

	return nil
}

// LoadTrainingData 加载训练数据
func (th *CrewTrainingHandler) LoadTrainingData(ctx context.Context, filename string) (*TrainingData, error) {
	if filename == "" {
		return nil, fmt.Errorf("filename cannot be empty")
	}

	// 读取文件
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read training data file: %w", err)
	}

	// 反序列化数据
	var data TrainingData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal training data: %w", err)
	}

	th.logger.Info("training data loaded",
		logger.Field{Key: "filename", Value: filename},
		logger.Field{Key: "iterations", Value: len(data.Iterations)},
		logger.Field{Key: "session_id", Value: data.SessionID},
	)

	return &data, nil
}

// GetTrainingStatus 获取训练状态
func (th *CrewTrainingHandler) GetTrainingStatus(ctx context.Context) *TrainingStatus {
	th.statusMu.RLock()
	defer th.statusMu.RUnlock()

	// 创建副本避免并发问题
	status := *th.status

	// 更新运行时信息
	if status.IsRunning {
		status.ElapsedTime = time.Since(status.StartTime)
		if status.CurrentIteration > 0 {
			avgDuration := status.ElapsedTime / time.Duration(status.CurrentIteration)
			remaining := time.Duration(status.TotalIterations-status.CurrentIteration) * avgDuration
			status.EstimatedRemaining = remaining
		}
	}

	return &status
}

// StopTraining 停止训练
func (th *CrewTrainingHandler) StopTraining(ctx context.Context) error {
	th.statusMu.Lock()
	if !th.status.IsRunning {
		th.statusMu.Unlock()
		return fmt.Errorf("training is not running")
	}

	th.status.IsRunning = false
	th.status.Status = "stopping"
	th.statusMu.Unlock()

	// 发送停止信号
	select {
	case th.stopChan <- true:
	default: // 非阻塞发送
	}

	// 生成训练总结
	th.generateTrainingSummary()

	// 最终保存
	if th.config.AutoSave {
		if err := th.SaveTrainingData(ctx, th.trainingData); err != nil {
			th.logger.Error("failed to save training data on stop",
				logger.Field{Key: "error", Value: err},
			)
		}
	}

	// 发射训练停止事件
	stopEvent := NewTrainingStoppedEvent(th.trainingData.SessionID, "manual_stop")
	th.eventBus.Emit(ctx, th, stopEvent)

	th.logger.Info("training stopped",
		logger.Field{Key: "session_id", Value: th.trainingData.SessionID},
		logger.Field{Key: "completed_iterations", Value: th.status.CurrentIteration},
	)

	return nil
}

// generateTrainingSummary 生成训练总结
func (th *CrewTrainingHandler) generateTrainingSummary() {
	th.dataMu.Lock()
	defer th.dataMu.Unlock()

	iterations := th.trainingData.Iterations
	if len(iterations) == 0 {
		return
	}

	summary := th.trainingData.Summary
	summary.TotalIterations = len(iterations)

	var totalDuration time.Duration
	var successfulRuns, failedRuns int
	var totalScore, totalTokens float64
	var scores []float64

	for _, iteration := range iterations {
		totalDuration += iteration.Duration

		if iteration.Success {
			successfulRuns++
		} else {
			failedRuns++
		}

		// 收集分数
		if iteration.Feedback != nil {
			score := iteration.Feedback.QualityScore
			totalScore += score
			scores = append(scores, score)
		} else if iteration.Metrics != nil {
			score := iteration.Metrics.AverageScore
			totalScore += score
			scores = append(scores, score)
		}

		// 收集token使用
		if iteration.Metrics != nil {
			totalTokens += float64(iteration.Metrics.TokensUsed)
		}
	}

	summary.SuccessfulRuns = successfulRuns
	summary.FailedRuns = failedRuns
	summary.TotalDuration = totalDuration
	summary.AverageDuration = totalDuration / time.Duration(len(iterations))

	if len(scores) > 0 {
		summary.AverageFeedback = totalScore / float64(len(scores))
		summary.InitialScore = scores[0]
		summary.FinalScore = scores[len(scores)-1]
		summary.ImprovementRate = (summary.FinalScore - summary.InitialScore) / summary.InitialScore * 100

		// 找到最佳和最差分数
		summary.BestScore = scores[0]
		summary.WorstScore = scores[0]
		for _, score := range scores {
			if score > summary.BestScore {
				summary.BestScore = score
			}
			if score < summary.WorstScore {
				summary.WorstScore = score
			}
		}
	}

	summary.TotalTokens = int(totalTokens)
	if len(iterations) > 0 {
		summary.AverageTokens = int(totalTokens / float64(len(iterations)))
	}

	// 生成建议
	summary.Recommendations = th.generateRecommendations(summary)
}

// generateRecommendations 生成训练建议
func (th *CrewTrainingHandler) generateRecommendations(summary *TrainingSummary) []string {
	var recommendations []string

	if summary.ImprovementRate < 5 {
		recommendations = append(recommendations, "Consider adjusting learning rate or training parameters")
	}

	if summary.SuccessfulRuns < summary.TotalIterations/2 {
		recommendations = append(recommendations, "High failure rate detected, review task configuration")
	}

	if summary.AverageFeedback < 5 {
		recommendations = append(recommendations, "Low feedback scores, consider improving prompt engineering")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Training completed successfully with good performance")
	}

	return recommendations
}

// createBackup 创建备份文件
func (th *CrewTrainingHandler) createBackup(filename string) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return
	}

	timestamp := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("%s.backup_%s", filename, timestamp)

	if err := th.copyFile(filename, backupName); err != nil {
		th.logger.Error("failed to create backup",
			logger.Field{Key: "filename", Value: filename},
			logger.Field{Key: "backup", Value: backupName},
			logger.Field{Key: "error", Value: err},
		)
	}
}

// copyFile 复制文件
func (th *CrewTrainingHandler) copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// CheckEarlyStop 检查是否应该早停
func (th *CrewTrainingHandler) CheckEarlyStop(recentScores []float64) bool {
	if !th.config.EarlyStopping || len(recentScores) < th.config.PatientceEpochs {
		return false
	}

	// 检查最近几个epoch是否没有显著改善
	bestScore := recentScores[0]
	for i := 1; i < len(recentScores); i++ {
		improvement := (recentScores[i] - bestScore) / bestScore
		if improvement > th.config.MinImprovement {
			return false
		}
		if recentScores[i] > bestScore {
			bestScore = recentScores[i]
		}
	}

	th.logger.Info("early stopping triggered",
		logger.Field{Key: "patience_epochs", Value: th.config.PatientceEpochs},
		logger.Field{Key: "min_improvement", Value: th.config.MinImprovement},
	)

	return true
}
