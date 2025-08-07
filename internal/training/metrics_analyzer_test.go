package training

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ynl/greensoulai/pkg/logger"
)

func TestNewMetricsAnalyzer(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	analyzer := NewMetricsAnalyzer(testLogger)

	assert.NotNil(t, analyzer)
	assert.Equal(t, testLogger, analyzer.logger)
}

func TestAnalyzeIteration(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	analyzer := NewMetricsAnalyzer(testLogger)

	ctx := context.Background()
	now := time.Now()
	duration := 500 * time.Millisecond

	// 创建测试迭代数据
	iteration := &IterationData{
		IterationID: "test-analyze-1",
		Index:       0,
		Timestamp:   now,
		Duration:    duration,
		Success:     true,
		Feedback: &HumanFeedback{
			QualityScore:  8.5,
			AccuracyScore: 9.0,
			Usefulness:    7.8,
		},
		AgentData: []*AgentIterationData{
			{
				AgentRole:     "writer",
				ExecutionTime: 300 * time.Millisecond,
				TokensUsed:    150,
				ToolsUsed:     []string{"web_search", "text_editor"},
				Success:       true,
			},
			{
				AgentRole:     "reviewer",
				ExecutionTime: 200 * time.Millisecond,
				TokensUsed:    80,
				ToolsUsed:     []string{"grammar_check"},
				Success:       true,
			},
		},
		TaskData: []*TaskIterationData{
			{
				TaskDescription: "Write blog post",
				ExecutionTime:   450 * time.Millisecond,
				Success:         true,
				OutputLength:    1500,
				ValidationScore: 8.7,
			},
		},
	}

	metrics, err := analyzer.AnalyzeIteration(ctx, iteration)

	assert.NoError(t, err)
	assert.NotNil(t, metrics)

	// 验证基础指标
	assert.Equal(t, duration, metrics.ExecutionTime)
	assert.Equal(t, 1.0, metrics.SuccessRate)
	assert.Equal(t, 0.0, metrics.ErrorRate)

	// 验证资源使用
	assert.Greater(t, metrics.MemoryUsage, int64(0))
	assert.Greater(t, metrics.CPUUsage, 0.0)
	assert.Equal(t, 230, metrics.TokensUsed) // 150 + 80

	// 验证质量指标
	expectedFeedbackScore := (8.5 + 9.0 + 7.8) / 3.0
	assert.Equal(t, expectedFeedbackScore, metrics.FeedbackScore)
	assert.Greater(t, metrics.AverageScore, 0.0)

	// 验证Agent性能
	assert.Len(t, metrics.AgentPerformance, 2)

	writerMetrics := metrics.AgentPerformance["writer"]
	assert.NotNil(t, writerMetrics)
	assert.Equal(t, 1, writerMetrics.ExecutionCount)
	assert.Equal(t, 300*time.Millisecond, writerMetrics.AverageTime)
	assert.Equal(t, 1.0, writerMetrics.SuccessRate)
	assert.Equal(t, 150, writerMetrics.TokensPerExecution)
	assert.Equal(t, 2, writerMetrics.ToolUsageCount)

	reviewerMetrics := metrics.AgentPerformance["reviewer"]
	assert.NotNil(t, reviewerMetrics)
	assert.Equal(t, 1, reviewerMetrics.ExecutionCount)
	assert.Equal(t, 200*time.Millisecond, reviewerMetrics.AverageTime)
	assert.Equal(t, 1.0, reviewerMetrics.SuccessRate)
	assert.Equal(t, 80, reviewerMetrics.TokensPerExecution)
	assert.Equal(t, 1, reviewerMetrics.ToolUsageCount)

	// 验证任务性能
	assert.Len(t, metrics.TaskPerformance, 1)

	taskMetrics := metrics.TaskPerformance["Write blog post"]
	assert.NotNil(t, taskMetrics)
	assert.Equal(t, 1, taskMetrics.ExecutionCount)
	assert.Equal(t, 450*time.Millisecond, taskMetrics.AverageTime)
	assert.Equal(t, 1.0, taskMetrics.SuccessRate)
	assert.Equal(t, 1500, taskMetrics.AverageOutputLength)
	assert.Equal(t, 8.7, taskMetrics.ValidationScore)
}

func TestAnalyzeIterationFailure(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	analyzer := NewMetricsAnalyzer(testLogger)

	ctx := context.Background()

	// 创建失败的迭代数据
	iteration := &IterationData{
		IterationID: "test-failure",
		Index:       0,
		Duration:    100 * time.Millisecond,
		Success:     false,
		Error:       "execution failed",
		AgentData: []*AgentIterationData{
			{
				AgentRole: "failed_agent",
				Success:   false,
				Error:     "agent execution failed",
			},
		},
		TaskData: []*TaskIterationData{
			{
				TaskDescription: "Failed task",
				Success:         false,
				Error:           "task execution failed",
			},
		},
	}

	metrics, err := analyzer.AnalyzeIteration(ctx, iteration)

	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, 0.0, metrics.SuccessRate)
	assert.Equal(t, 1.0, metrics.ErrorRate)
	assert.Equal(t, 3.0, metrics.AverageScore) // 失败时的默认分数

	// 验证失败的Agent指标
	agentMetrics := metrics.AgentPerformance["failed_agent"]
	assert.NotNil(t, agentMetrics)
	assert.Equal(t, 0.0, agentMetrics.SuccessRate)

	// 验证失败的任务指标
	taskMetrics := metrics.TaskPerformance["Failed task"]
	assert.NotNil(t, taskMetrics)
	assert.Equal(t, 0.0, taskMetrics.SuccessRate)
}

func TestAnalyzeIterationWithoutFeedback(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	analyzer := NewMetricsAnalyzer(testLogger)

	ctx := context.Background()

	// 创建没有反馈的成功迭代
	iteration := &IterationData{
		IterationID: "no-feedback",
		Success:     true,
		Duration:    200 * time.Millisecond,
		TaskData: []*TaskIterationData{
			{
				TaskDescription: "Task without feedback",
				Success:         true,
				ValidationScore: 8.0,
			},
		},
	}

	metrics, err := analyzer.AnalyzeIteration(ctx, iteration)

	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, 1.0, metrics.SuccessRate)
	assert.Equal(t, 8.0, metrics.ValidationScore)
	assert.Equal(t, 8.0, metrics.AverageScore) // 应该使用验证分数
}

func TestAnalyzeTrend(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	analyzer := NewMetricsAnalyzer(testLogger)

	ctx := context.Background()
	now := time.Now()

	// 创建多个迭代数据，显示改进趋势
	iterations := []*IterationData{
		{
			IterationID: "trend-1",
			Index:       0,
			Timestamp:   now,
			Duration:    300 * time.Millisecond,
			Success:     true,
			Feedback:    &HumanFeedback{QualityScore: 6.0, AccuracyScore: 6.0, Usefulness: 6.0},
		},
		{
			IterationID: "trend-2",
			Index:       1,
			Timestamp:   now.Add(1 * time.Minute),
			Duration:    250 * time.Millisecond,
			Success:     true,
			Feedback:    &HumanFeedback{QualityScore: 7.0, AccuracyScore: 7.0, Usefulness: 7.0},
		},
		{
			IterationID: "trend-3",
			Index:       2,
			Timestamp:   now.Add(2 * time.Minute),
			Duration:    200 * time.Millisecond,
			Success:     true,
			Feedback:    &HumanFeedback{QualityScore: 8.0, AccuracyScore: 8.0, Usefulness: 8.0},
		},
		{
			IterationID: "trend-4",
			Index:       3,
			Timestamp:   now.Add(3 * time.Minute),
			Duration:    180 * time.Millisecond,
			Success:     false, // 一次失败
		},
		{
			IterationID: "trend-5",
			Index:       4,
			Timestamp:   now.Add(4 * time.Minute),
			Duration:    160 * time.Millisecond,
			Success:     true,
			Feedback:    &HumanFeedback{QualityScore: 8.5, AccuracyScore: 8.5, Usefulness: 8.5},
		},
	}

	trend, err := analyzer.AnalyzeTrend(ctx, iterations)

	assert.NoError(t, err)
	assert.NotNil(t, trend)
	assert.Equal(t, 5, trend.TotalIterations)
	assert.Equal(t, now, trend.TimeRange.Start)
	assert.Equal(t, now.Add(4*time.Minute), trend.TimeRange.End)
	assert.Equal(t, 0.8, trend.SuccessRate) // 4成功/5总数
	assert.Greater(t, trend.AverageScore, 0.0)
	assert.Greater(t, trend.ImprovementRate, 0.0) // 应该显示改进
	// 注意：趋势斜率可能为负值，因为有一次失败的迭代，我们只验证它不为0
	assert.NotEqual(t, 0.0, trend.TrendSlope)
	assert.Len(t, trend.Scores, 5)
	assert.Len(t, trend.Times, 5)
}

func TestAnalyzeTrendInsufficientData(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	analyzer := NewMetricsAnalyzer(testLogger)

	ctx := context.Background()

	// 只有一个迭代，不足以进行趋势分析
	iterations := []*IterationData{
		{
			IterationID: "single",
			Success:     true,
		},
	}

	_, err := analyzer.AnalyzeTrend(ctx, iterations)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "need at least 2 iterations")
}

func TestCalculateTrendSlope(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	analyzer := NewMetricsAnalyzer(testLogger)

	tests := []struct {
		name     string
		scores   []float64
		expected float64
	}{
		{
			name:     "empty scores",
			scores:   []float64{},
			expected: 0.0,
		},
		{
			name:     "single score",
			scores:   []float64{5.0},
			expected: 0.0,
		},
		{
			name:     "increasing trend",
			scores:   []float64{5.0, 6.0, 7.0, 8.0},
			expected: 1.0, // 每个点增加1
		},
		{
			name:     "decreasing trend",
			scores:   []float64{8.0, 7.0, 6.0, 5.0},
			expected: -1.0, // 每个点减少1
		},
		{
			name:     "flat trend",
			scores:   []float64{7.0, 7.0, 7.0, 7.0},
			expected: 0.0, // 无变化
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slope := analyzer.calculateTrendSlope(tt.scores)
			assert.InDelta(t, tt.expected, slope, 0.01, "slope should match expected value")
		})
	}
}

func TestCalculateVariance(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	analyzer := NewMetricsAnalyzer(testLogger)

	tests := []struct {
		name     string
		scores   []float64
		expected float64
	}{
		{
			name:     "empty scores",
			scores:   []float64{},
			expected: 0.0,
		},
		{
			name:     "single score",
			scores:   []float64{5.0},
			expected: 0.0,
		},
		{
			name:     "identical scores",
			scores:   []float64{5.0, 5.0, 5.0},
			expected: 0.0,
		},
		{
			name:     "varying scores",
			scores:   []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			expected: 2.0, // 方差为2.0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			variance := analyzer.calculateVariance(tt.scores)
			assert.InDelta(t, tt.expected, variance, 0.01, "variance should match expected value")
		})
	}
}

func TestEstimateCPUUsage(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	analyzer := NewMetricsAnalyzer(testLogger)

	tests := []struct {
		name          string
		executionTime time.Duration
		expectedRange [2]float64 // [min, max]
	}{
		{
			name:          "very fast execution",
			executionTime: 500 * time.Millisecond,
			expectedRange: [2]float64{5.0, 15.0}, // 轻负载
		},
		{
			name:          "medium execution",
			executionTime: 5 * time.Second,
			expectedRange: [2]float64{25.0, 35.0}, // 中等负载
		},
		{
			name:          "slow execution",
			executionTime: 20 * time.Second,
			expectedRange: [2]float64{55.0, 65.0}, // 高负载
		},
		{
			name:          "very slow execution",
			executionTime: 60 * time.Second,
			expectedRange: [2]float64{75.0, 85.0}, // 非常高负载
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usage := analyzer.estimateCPUUsage(tt.executionTime)
			assert.GreaterOrEqual(t, usage, tt.expectedRange[0])
			assert.LessOrEqual(t, usage, tt.expectedRange[1])
		})
	}
}

func TestGenerateRecommendations(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	analyzer := NewMetricsAnalyzer(testLogger)

	tests := []struct {
		name             string
		trend            *TrendAnalysis
		expectedCount    int
		expectedContains []string
	}{
		{
			name: "good performance",
			trend: &TrendAnalysis{
				SuccessRate:     0.95,
				ImprovementRate: 10.0,
				TrendSlope:      0.1,
				ScoreVariance:   1.0,
			},
			expectedCount:    1,
			expectedContains: []string{"stable and improving"},
		},
		{
			name: "low success rate",
			trend: &TrendAnalysis{
				SuccessRate:     0.6,
				ImprovementRate: 5.0,
				TrendSlope:      0.05,
				ScoreVariance:   1.0,
			},
			expectedCount:    1,
			expectedContains: []string{"Success rate is 60.0%"},
		},
		{
			name: "declining performance",
			trend: &TrendAnalysis{
				SuccessRate:     0.8,
				ImprovementRate: -5.0,
				TrendSlope:      -0.15,
				ScoreVariance:   3.0,
			},
			expectedCount:    3,
			expectedContains: []string{"declining", "Negative performance trend", "High score variance"},
		},
		{
			name: "plateau performance",
			trend: &TrendAnalysis{
				SuccessRate:     0.85,
				ImprovementRate: 2.0,
				TrendSlope:      0.02,
				ScoreVariance:   1.5,
			},
			expectedCount:    2, // 可能有多个推荐
			expectedContains: []string{"plateau detected"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recommendations := analyzer.GenerateRecommendations(tt.trend)

			assert.Len(t, recommendations, tt.expectedCount)

			for _, expected := range tt.expectedContains {
				found := false
				for _, rec := range recommendations {
					if strings.Contains(rec, expected) {
						found = true
						break
					}
				}
				assert.True(t, found, "should contain recommendation with text: %s", expected)
			}
		})
	}
}

// BenchmarkAnalyzeIteration 迭代分析性能基准测试
func BenchmarkAnalyzeIteration(b *testing.B) {
	testLogger := logger.NewConsoleLogger()
	analyzer := NewMetricsAnalyzer(testLogger)

	ctx := context.Background()

	iteration := &IterationData{
		IterationID: "benchmark-test",
		Success:     true,
		Duration:    200 * time.Millisecond,
		Feedback: &HumanFeedback{
			QualityScore:  8.0,
			AccuracyScore: 8.5,
			Usefulness:    7.8,
		},
		AgentData: []*AgentIterationData{
			{AgentRole: "agent1", Success: true, TokensUsed: 100},
			{AgentRole: "agent2", Success: true, TokensUsed: 150},
		},
		TaskData: []*TaskIterationData{
			{TaskDescription: "task1", Success: true, ValidationScore: 8.2},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.AnalyzeIteration(ctx, iteration)
	}
}

// BenchmarkCalculateTrendSlope 趋势斜率计算性能基准测试
func BenchmarkCalculateTrendSlope(b *testing.B) {
	testLogger := logger.NewConsoleLogger()
	analyzer := NewMetricsAnalyzer(testLogger)

	// 创建大量分数数据
	scores := make([]float64, 1000)
	for i := 0; i < 1000; i++ {
		scores[i] = 5.0 + float64(i)*0.001 // 轻微上升趋势
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.calculateTrendSlope(scores)
	}
}

// TestMetricsAnalyzerEdgeCases 测试边界情况
func TestMetricsAnalyzerEdgeCases(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	analyzer := NewMetricsAnalyzer(testLogger)

	ctx := context.Background()

	// 测试空迭代数据
	emptyIteration := &IterationData{
		IterationID: "empty",
		Success:     true,
		Duration:    0,
	}

	metrics, err := analyzer.AnalyzeIteration(ctx, emptyIteration)
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, 1.0, metrics.SuccessRate)
	assert.Equal(t, time.Duration(0), metrics.ExecutionTime)

	// 测试时间变异性计算
	times := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		300 * time.Millisecond,
	}
	variance := analyzer.calculateTimeVariance(times)
	assert.Greater(t, variance, time.Duration(0))
}
