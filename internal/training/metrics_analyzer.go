package training

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"time"

	"github.com/ynl/greensoulai/pkg/logger"
)

// MetricsAnalyzer 性能指标分析器
type MetricsAnalyzer struct {
	logger logger.Logger
}

// NewMetricsAnalyzer 创建新的性能分析器
func NewMetricsAnalyzer(logger logger.Logger) *MetricsAnalyzer {
	return &MetricsAnalyzer{
		logger: logger,
	}
}

// AnalyzeIteration 分析单次迭代的性能指标
func (ma *MetricsAnalyzer) AnalyzeIteration(ctx context.Context, iteration *IterationData) (*PerformanceMetrics, error) {
	ma.logger.Debug("analyzing iteration metrics",
		logger.Field{Key: "iteration_id", Value: iteration.IterationID},
		logger.Field{Key: "index", Value: iteration.Index},
	)

	metrics := &PerformanceMetrics{
		ExecutionTime:    iteration.Duration,
		AgentPerformance: make(map[string]*AgentMetrics),
		TaskPerformance:  make(map[string]*TaskMetrics),
	}

	// 分析基础指标
	ma.analyzeBasicMetrics(metrics, iteration)

	// 分析资源使用
	ma.analyzeResourceUsage(metrics, iteration)

	// 分析质量指标
	ma.analyzeQualityMetrics(metrics, iteration)

	// 分析Agent性能
	ma.analyzeAgentPerformance(metrics, iteration)

	// 分析任务性能
	ma.analyzeTaskPerformance(metrics, iteration)

	// 计算综合分数
	ma.calculateOverallScores(metrics, iteration)

	ma.logger.Debug("metrics analysis completed",
		logger.Field{Key: "iteration_id", Value: iteration.IterationID},
		logger.Field{Key: "average_score", Value: metrics.AverageScore},
		logger.Field{Key: "execution_time", Value: metrics.ExecutionTime},
	)

	return metrics, nil
}

// analyzeBasicMetrics 分析基础指标
func (ma *MetricsAnalyzer) analyzeBasicMetrics(metrics *PerformanceMetrics, iteration *IterationData) {
	// 成功率
	if iteration.Success {
		metrics.SuccessRate = 1.0
		metrics.ErrorRate = 0.0
	} else {
		metrics.SuccessRate = 0.0
		metrics.ErrorRate = 1.0
	}

	// 执行时间已经在主函数中设置
}

// analyzeResourceUsage 分析资源使用情况
func (ma *MetricsAnalyzer) analyzeResourceUsage(metrics *PerformanceMetrics, iteration *IterationData) {
	// 获取内存使用情况
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	metrics.MemoryUsage = int64(memStats.Alloc)

	// CPU使用率（简化计算）
	// 在实际实现中，可以使用更精确的CPU监控
	metrics.CPUUsage = ma.estimateCPUUsage(iteration.Duration)

	// Token使用统计
	var totalTokens int
	for _, agentData := range iteration.AgentData {
		totalTokens += agentData.TokensUsed
	}
	metrics.TokensUsed = totalTokens
}

// analyzeQualityMetrics 分析质量指标
func (ma *MetricsAnalyzer) analyzeQualityMetrics(metrics *PerformanceMetrics, iteration *IterationData) {
	var scores []float64

	// 从反馈中获取分数
	if iteration.Feedback != nil {
		scores = append(scores, iteration.Feedback.QualityScore)
		scores = append(scores, iteration.Feedback.AccuracyScore)
		scores = append(scores, iteration.Feedback.Usefulness)

		metrics.FeedbackScore = (iteration.Feedback.QualityScore +
			iteration.Feedback.AccuracyScore +
			iteration.Feedback.Usefulness) / 3.0
	}

	// 从任务验证中获取分数
	var validationScores []float64
	for _, taskData := range iteration.TaskData {
		if taskData.ValidationScore > 0 {
			validationScores = append(validationScores, taskData.ValidationScore)
		}
	}

	if len(validationScores) > 0 {
		sum := 0.0
		for _, score := range validationScores {
			sum += score
		}
		metrics.ValidationScore = sum / float64(len(validationScores))
		scores = append(scores, metrics.ValidationScore)
	}

	// 计算平均分数
	if len(scores) > 0 {
		sum := 0.0
		for _, score := range scores {
			sum += score
		}
		metrics.AverageScore = sum / float64(len(scores))
	} else {
		// 基于成功率的默认分数
		if iteration.Success {
			metrics.AverageScore = 7.0 // 成功但没有具体评分
		} else {
			metrics.AverageScore = 3.0 // 失败
		}
	}
}

// analyzeAgentPerformance 分析Agent性能
func (ma *MetricsAnalyzer) analyzeAgentPerformance(metrics *PerformanceMetrics, iteration *IterationData) {
	for _, agentData := range iteration.AgentData {
		agentMetrics := &AgentMetrics{
			ExecutionCount:     1,
			AverageTime:        agentData.ExecutionTime,
			TokensPerExecution: agentData.TokensUsed,
			ToolUsageCount:     len(agentData.ToolsUsed),
		}

		if agentData.Success {
			agentMetrics.SuccessRate = 1.0
		} else {
			agentMetrics.SuccessRate = 0.0
		}

		metrics.AgentPerformance[agentData.AgentRole] = agentMetrics
	}
}

// analyzeTaskPerformance 分析任务性能
func (ma *MetricsAnalyzer) analyzeTaskPerformance(metrics *PerformanceMetrics, iteration *IterationData) {
	for _, taskData := range iteration.TaskData {
		taskMetrics := &TaskMetrics{
			ExecutionCount:      1,
			AverageTime:         taskData.ExecutionTime,
			AverageOutputLength: taskData.OutputLength,
			ValidationScore:     taskData.ValidationScore,
		}

		if taskData.Success {
			taskMetrics.SuccessRate = 1.0
		} else {
			taskMetrics.SuccessRate = 0.0
		}

		metrics.TaskPerformance[taskData.TaskDescription] = taskMetrics
	}
}

// calculateOverallScores 计算综合分数
func (ma *MetricsAnalyzer) calculateOverallScores(metrics *PerformanceMetrics, iteration *IterationData) {
	// 改进率（单次迭代无法计算，需要历史数据）
	metrics.ImprovementRate = 0.0

	// 一致性分数（基于错误率和执行时间的稳定性）
	consistencyScore := 1.0

	// 如果有错误，降低一致性分数
	if !iteration.Success {
		consistencyScore *= 0.5
	}

	// 基于执行时间的一致性（简化计算）
	expectedTime := 30 * time.Second // 假设的期望执行时间
	timeVariance := math.Abs(float64(iteration.Duration-expectedTime)) / float64(expectedTime)
	if timeVariance > 0.5 { // 如果时间偏差超过50%
		consistencyScore *= 0.8
	}

	metrics.ConsistencyScore = consistencyScore
}

// estimateCPUUsage 估算CPU使用率
func (ma *MetricsAnalyzer) estimateCPUUsage(executionTime time.Duration) float64 {
	// 这是一个简化的估算，实际实现中应该使用更精确的CPU监控
	// 基于执行时间推算CPU使用率

	if executionTime < 1*time.Second {
		return 10.0 // 轻负载
	} else if executionTime < 10*time.Second {
		return 30.0 // 中等负载
	} else if executionTime < 30*time.Second {
		return 60.0 // 高负载
	} else {
		return 80.0 // 非常高的负载
	}
}

// AnalyzeTrend 分析性能趋势
func (ma *MetricsAnalyzer) AnalyzeTrend(ctx context.Context, iterations []*IterationData) (*TrendAnalysis, error) {
	if len(iterations) < 2 {
		return nil, fmt.Errorf("need at least 2 iterations for trend analysis")
	}

	ma.logger.Debug("analyzing performance trend",
		logger.Field{Key: "iterations_count", Value: len(iterations)},
	)

	trend := &TrendAnalysis{
		TotalIterations: len(iterations),
		TimeRange: TimeRange{
			Start: iterations[0].Timestamp,
			End:   iterations[len(iterations)-1].Timestamp,
		},
		Scores: make([]float64, 0, len(iterations)),
		Times:  make([]time.Duration, 0, len(iterations)),
	}

	// 收集数据点
	var totalScore, totalTime float64
	successCount := 0

	for _, iteration := range iterations {
		// 计算迭代分数
		score := ma.calculateIterationScore(iteration)
		trend.Scores = append(trend.Scores, score)
		trend.Times = append(trend.Times, iteration.Duration)

		totalScore += score
		totalTime += float64(iteration.Duration)

		if iteration.Success {
			successCount++
		}
	}

	// 计算趋势统计
	trend.AverageScore = totalScore / float64(len(iterations))
	trend.AverageTime = time.Duration(totalTime / float64(len(iterations)))
	trend.SuccessRate = float64(successCount) / float64(len(iterations))

	// 计算改进率
	if len(trend.Scores) >= 2 {
		initialScore := trend.Scores[0]
		finalScore := trend.Scores[len(trend.Scores)-1]
		if initialScore > 0 {
			trend.ImprovementRate = (finalScore - initialScore) / initialScore * 100
		}
	}

	// 计算线性回归趋势
	trend.TrendSlope = ma.calculateTrendSlope(trend.Scores)

	// 分析变异性
	trend.ScoreVariance = ma.calculateVariance(trend.Scores)
	trend.TimeVariance = ma.calculateTimeVariance(trend.Times)

	ma.logger.Debug("trend analysis completed",
		logger.Field{Key: "average_score", Value: trend.AverageScore},
		logger.Field{Key: "improvement_rate", Value: trend.ImprovementRate},
		logger.Field{Key: "trend_slope", Value: trend.TrendSlope},
	)

	return trend, nil
}

// calculateIterationScore 计算迭代分数
func (ma *MetricsAnalyzer) calculateIterationScore(iteration *IterationData) float64 {
	if !iteration.Success {
		return 0.0
	}

	score := 5.0 // 基础分数

	// 如果有反馈分数
	if iteration.Feedback != nil {
		feedbackScore := (iteration.Feedback.QualityScore +
			iteration.Feedback.AccuracyScore +
			iteration.Feedback.Usefulness) / 3.0
		score = feedbackScore
	}

	// 如果有性能指标
	if iteration.Metrics != nil && iteration.Metrics.AverageScore > 0 {
		score = iteration.Metrics.AverageScore
	}

	return score
}

// calculateTrendSlope 计算趋势斜率
func (ma *MetricsAnalyzer) calculateTrendSlope(scores []float64) float64 {
	n := float64(len(scores))
	if n < 2 {
		return 0
	}

	// 计算线性回归斜率 y = mx + b
	var sumX, sumY, sumXY, sumXX float64

	for i, score := range scores {
		x := float64(i)
		sumX += x
		sumY += score
		sumXY += x * score
		sumXX += x * x
	}

	// 斜率 m = (n*∑xy - ∑x*∑y) / (n*∑x² - (∑x)²)
	denominator := n*sumXX - sumX*sumX
	if denominator == 0 {
		return 0
	}

	slope := (n*sumXY - sumX*sumY) / denominator
	return slope
}

// calculateVariance 计算方差
func (ma *MetricsAnalyzer) calculateVariance(scores []float64) float64 {
	if len(scores) < 2 {
		return 0
	}

	// 计算均值
	sum := 0.0
	for _, score := range scores {
		sum += score
	}
	mean := sum / float64(len(scores))

	// 计算方差
	sumSquaredDiff := 0.0
	for _, score := range scores {
		diff := score - mean
		sumSquaredDiff += diff * diff
	}

	return sumSquaredDiff / float64(len(scores))
}

// calculateTimeVariance 计算时间方差
func (ma *MetricsAnalyzer) calculateTimeVariance(times []time.Duration) time.Duration {
	if len(times) < 2 {
		return 0
	}

	// 计算均值
	sum := int64(0)
	for _, t := range times {
		sum += int64(t)
	}
	mean := time.Duration(sum / int64(len(times)))

	// 计算方差
	sumSquaredDiff := int64(0)
	for _, t := range times {
		diff := int64(t - mean)
		sumSquaredDiff += diff * diff
	}

	variance := time.Duration(sumSquaredDiff / int64(len(times)))
	return variance
}

// TrendAnalysis 趋势分析结果
type TrendAnalysis struct {
	TotalIterations int             `json:"total_iterations"`
	TimeRange       TimeRange       `json:"time_range"`
	AverageScore    float64         `json:"average_score"`
	AverageTime     time.Duration   `json:"average_time"`
	SuccessRate     float64         `json:"success_rate"`
	ImprovementRate float64         `json:"improvement_rate"`
	TrendSlope      float64         `json:"trend_slope"`
	ScoreVariance   float64         `json:"score_variance"`
	TimeVariance    time.Duration   `json:"time_variance"`
	Scores          []float64       `json:"scores"`
	Times           []time.Duration `json:"times"`
}

// TimeRange 时间范围
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// GenerateRecommendations 生成性能改进建议
func (ma *MetricsAnalyzer) GenerateRecommendations(trend *TrendAnalysis) []string {
	var recommendations []string

	// 基于成功率的建议
	if trend.SuccessRate < 0.8 {
		recommendations = append(recommendations,
			fmt.Sprintf("Success rate is %.1f%%, consider reviewing task configuration", trend.SuccessRate*100))
	}

	// 基于改进趋势的建议
	if trend.ImprovementRate < 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Performance declining by %.1f%%, review training parameters", math.Abs(trend.ImprovementRate)))
	} else if trend.ImprovementRate < 5 {
		recommendations = append(recommendations,
			"Low improvement rate, consider adjusting learning parameters")
	}

	// 基于趋势斜率的建议
	if trend.TrendSlope < -0.1 {
		recommendations = append(recommendations,
			"Negative performance trend detected, investigate recent changes")
	} else if trend.TrendSlope < 0.05 {
		recommendations = append(recommendations,
			"Performance plateau detected, consider changing training strategy")
	}

	// 基于方差的建议
	if trend.ScoreVariance > 2.0 {
		recommendations = append(recommendations,
			"High score variance indicates inconsistent performance")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations,
			"Performance is stable and improving, continue current approach")
	}

	return recommendations
}
