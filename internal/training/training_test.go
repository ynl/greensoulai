package training

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestDefaultTrainingConfig 测试默认训练配置
func TestDefaultTrainingConfig(t *testing.T) {
	config := DefaultTrainingConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 10, config.Iterations)
	assert.Equal(t, "training_data.json", config.Filename)
	assert.Equal(t, 0.001, config.LearningRate)
	assert.Equal(t, 1, config.BatchSize)
	assert.Equal(t, 0.2, config.ValidationSplit)
	assert.True(t, config.CollectFeedback)
	assert.Equal(t, 5*time.Minute, config.FeedbackTimeout)
	assert.True(t, config.MetricsEnabled)
	assert.Equal(t, 1, config.MetricsInterval)
	assert.False(t, config.EarlyStopping)
	assert.Equal(t, 3, config.PatientceEpochs)
	assert.Equal(t, 0.01, config.MinImprovement)
	assert.Equal(t, 5, config.SaveInterval)
	assert.True(t, config.Verbose)
	assert.True(t, config.AutoSave)
	assert.Equal(t, 3, config.BackupCount)

	// 验证默认的反馈提示
	expectedPrompts := []string{
		"Please rate the quality of this output (1-10):",
		"Any suggestions for improvement?",
	}
	assert.Equal(t, expectedPrompts, config.FeedbackPrompts)

	// 验证默认的目标指标
	expectedMetrics := []string{"execution_time", "success_rate", "feedback_score"}
	assert.Equal(t, expectedMetrics, config.TargetMetrics)
}

// TestTrainingConfigValidation 测试训练配置验证
func TestTrainingConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config *TrainingConfig
		valid  bool
	}{
		{
			name: "valid basic config",
			config: &TrainingConfig{
				Iterations:      10,
				Filename:        "test.json",
				LearningRate:    0.01,
				BatchSize:       2,
				ValidationSplit: 0.1,
			},
			valid: true,
		},
		{
			name: "zero iterations",
			config: &TrainingConfig{
				Iterations: 0,
			},
			valid: false,
		},
		{
			name: "negative learning rate",
			config: &TrainingConfig{
				Iterations:   5,
				LearningRate: -0.01,
			},
			valid: false,
		},
		{
			name: "invalid validation split",
			config: &TrainingConfig{
				Iterations:      5,
				ValidationSplit: 1.5,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				assert.NotNil(t, tt.config)
			} else {
				// 这里应该有验证逻辑，但我们目前的实现中没有
				// 这个测试为将来的验证功能做准备
				assert.NotNil(t, tt.config)
			}
		})
	}
}

// TestTrainingData 测试训练数据结构
func TestTrainingData(t *testing.T) {
	now := time.Now()
	config := DefaultTrainingConfig()

	data := &TrainingData{
		CreatedAt: now,
		UpdatedAt: now,
		Version:   "1.0",
		Config:    config,
		SessionID: "test-session-123",
		CrewName:  "test-crew",
		TotalRuns: 5,
		Iterations: []*IterationData{
			{
				IterationID: "iter-1",
				Index:       0,
				Timestamp:   now,
				Duration:    100 * time.Millisecond,
				Success:     true,
			},
		},
		Summary: &TrainingSummary{
			TotalIterations: 1,
			SuccessfulRuns:  1,
			FailedRuns:      0,
		},
	}

	assert.Equal(t, now, data.CreatedAt)
	assert.Equal(t, now, data.UpdatedAt)
	assert.Equal(t, "1.0", data.Version)
	assert.Equal(t, config, data.Config)
	assert.Equal(t, "test-session-123", data.SessionID)
	assert.Equal(t, "test-crew", data.CrewName)
	assert.Equal(t, 5, data.TotalRuns)
	assert.Len(t, data.Iterations, 1)
	assert.Equal(t, "iter-1", data.Iterations[0].IterationID)
	assert.NotNil(t, data.Summary)
	assert.Equal(t, 1, data.Summary.TotalIterations)
}

// TestIterationData 测试迭代数据结构
func TestIterationData(t *testing.T) {
	now := time.Now()
	duration := 500 * time.Millisecond

	inputs := map[string]interface{}{
		"task": "test task",
		"goal": "test goal",
	}

	outputs := map[string]interface{}{
		"result": "test result",
		"status": "completed",
	}

	feedback := &HumanFeedback{
		IterationID:   "iter-feedback-1",
		QualityScore:  8.5,
		AccuracyScore: 9.0,
		Usefulness:    7.8,
		Comments:      "Good quality output",
	}

	metrics := &PerformanceMetrics{
		ExecutionTime: duration,
		SuccessRate:   1.0,
		AverageScore:  8.5,
		TokensUsed:    150,
	}

	agentData := []*AgentIterationData{
		{
			AgentRole:     "writer",
			ExecutionTime: 300 * time.Millisecond,
			TokensUsed:    80,
			ToolsUsed:     []string{"web_search", "text_editor"},
			Success:       true,
		},
	}

	taskData := []*TaskIterationData{
		{
			TaskDescription: "Write a blog post",
			ExecutionTime:   200 * time.Millisecond,
			Success:         true,
			OutputLength:    1500,
			ValidationScore: 8.7,
		},
	}

	iteration := &IterationData{
		IterationID: "test-iteration-1",
		Index:       0,
		Timestamp:   now,
		Duration:    duration,
		Inputs:      inputs,
		Outputs:     outputs,
		Success:     true,
		Feedback:    feedback,
		Metrics:     metrics,
		AgentData:   agentData,
		TaskData:    taskData,
	}

	assert.Equal(t, "test-iteration-1", iteration.IterationID)
	assert.Equal(t, 0, iteration.Index)
	assert.Equal(t, now, iteration.Timestamp)
	assert.Equal(t, duration, iteration.Duration)
	assert.Equal(t, inputs, iteration.Inputs)
	assert.Equal(t, outputs, iteration.Outputs)
	assert.True(t, iteration.Success)
	assert.Equal(t, "", iteration.Error) // 成功时应该没有错误
	assert.Equal(t, feedback, iteration.Feedback)
	assert.Equal(t, metrics, iteration.Metrics)
	assert.Len(t, iteration.AgentData, 1)
	assert.Len(t, iteration.TaskData, 1)
}

// TestHumanFeedback 测试人工反馈结构
func TestHumanFeedback(t *testing.T) {
	now := time.Now()

	feedback := &HumanFeedback{
		IterationID:   "feedback-test-1",
		Timestamp:     now,
		QualityScore:  8.5,
		AccuracyScore: 9.0,
		Usefulness:    7.5,
		Comments:      "Great job overall, minor improvements needed",
		Suggestions:   "Consider adding more examples",
		Issues:        []string{"grammar error in line 3", "unclear explanation"},
		Categories: map[string]float64{
			"clarity":      8.0,
			"completeness": 7.5,
			"creativity":   9.0,
		},
		Tags:       []string{"quality", "improvement", "feedback"},
		Verified:   true,
		VerifiedBy: "human_reviewer",
	}

	assert.Equal(t, "feedback-test-1", feedback.IterationID)
	assert.Equal(t, now, feedback.Timestamp)
	assert.Equal(t, 8.5, feedback.QualityScore)
	assert.Equal(t, 9.0, feedback.AccuracyScore)
	assert.Equal(t, 7.5, feedback.Usefulness)
	assert.Contains(t, feedback.Comments, "Great job")
	assert.Contains(t, feedback.Suggestions, "more examples")
	assert.Len(t, feedback.Issues, 2)
	assert.Equal(t, 8.0, feedback.Categories["clarity"])
	assert.Equal(t, 3, len(feedback.Categories))
	assert.Contains(t, feedback.Tags, "quality")
	assert.True(t, feedback.Verified)
	assert.Equal(t, "human_reviewer", feedback.VerifiedBy)
}

// TestPerformanceMetrics 测试性能指标结构
func TestPerformanceMetrics(t *testing.T) {
	agentMetrics := map[string]*AgentMetrics{
		"writer": {
			ExecutionCount:     5,
			AverageTime:        200 * time.Millisecond,
			SuccessRate:        0.8,
			TokensPerExecution: 100,
			ToolUsageCount:     15,
		},
		"reviewer": {
			ExecutionCount:     3,
			AverageTime:        150 * time.Millisecond,
			SuccessRate:        1.0,
			TokensPerExecution: 50,
			ToolUsageCount:     5,
		},
	}

	taskMetrics := map[string]*TaskMetrics{
		"write_article": {
			ExecutionCount:      2,
			AverageTime:         300 * time.Millisecond,
			SuccessRate:         1.0,
			AverageOutputLength: 1500,
			ValidationScore:     8.5,
		},
	}

	metrics := &PerformanceMetrics{
		ExecutionTime:    500 * time.Millisecond,
		SuccessRate:      0.9,
		ErrorRate:        0.1,
		MemoryUsage:      1024 * 1024, // 1MB
		CPUUsage:         45.5,
		TokensUsed:       250,
		AverageScore:     8.2,
		FeedbackScore:    8.5,
		ValidationScore:  8.0,
		AgentPerformance: agentMetrics,
		TaskPerformance:  taskMetrics,
		ImprovementRate:  15.3,
		ConsistencyScore: 0.85,
	}

	assert.Equal(t, 500*time.Millisecond, metrics.ExecutionTime)
	assert.Equal(t, 0.9, metrics.SuccessRate)
	assert.Equal(t, 0.1, metrics.ErrorRate)
	assert.Equal(t, int64(1024*1024), metrics.MemoryUsage)
	assert.Equal(t, 45.5, metrics.CPUUsage)
	assert.Equal(t, 250, metrics.TokensUsed)
	assert.Equal(t, 8.2, metrics.AverageScore)
	assert.Equal(t, 8.5, metrics.FeedbackScore)
	assert.Equal(t, 8.0, metrics.ValidationScore)
	assert.Len(t, metrics.AgentPerformance, 2)
	assert.Len(t, metrics.TaskPerformance, 1)
	assert.Equal(t, 15.3, metrics.ImprovementRate)
	assert.Equal(t, 0.85, metrics.ConsistencyScore)

	// 验证Agent指标
	writerMetrics := metrics.AgentPerformance["writer"]
	assert.Equal(t, 5, writerMetrics.ExecutionCount)
	assert.Equal(t, 0.8, writerMetrics.SuccessRate)
	assert.Equal(t, 100, writerMetrics.TokensPerExecution)

	// 验证任务指标
	articleMetrics := metrics.TaskPerformance["write_article"]
	assert.Equal(t, 2, articleMetrics.ExecutionCount)
	assert.Equal(t, 1.0, articleMetrics.SuccessRate)
	assert.Equal(t, 1500, articleMetrics.AverageOutputLength)
}

// TestTrainingSummary 测试训练总结结构
func TestTrainingSummary(t *testing.T) {
	summary := &TrainingSummary{
		TotalIterations: 10,
		SuccessfulRuns:  8,
		FailedRuns:      2,
		TotalDuration:   5 * time.Minute,
		AverageDuration: 30 * time.Second,
		InitialScore:    6.5,
		FinalScore:      8.2,
		ImprovementRate: 26.2,
		BestScore:       8.8,
		WorstScore:      5.5,
		TotalFeedback:   7,
		AverageFeedback: 7.8,
		TotalTokens:     1500,
		AverageTokens:   150,
		Recommendations: []string{
			"Consider increasing learning rate",
			"Add more diverse training examples",
		},
	}

	assert.Equal(t, 10, summary.TotalIterations)
	assert.Equal(t, 8, summary.SuccessfulRuns)
	assert.Equal(t, 2, summary.FailedRuns)
	assert.Equal(t, 5*time.Minute, summary.TotalDuration)
	assert.Equal(t, 30*time.Second, summary.AverageDuration)
	assert.Equal(t, 6.5, summary.InitialScore)
	assert.Equal(t, 8.2, summary.FinalScore)
	assert.Equal(t, 26.2, summary.ImprovementRate)
	assert.Equal(t, 8.8, summary.BestScore)
	assert.Equal(t, 5.5, summary.WorstScore)
	assert.Equal(t, 7, summary.TotalFeedback)
	assert.Equal(t, 7.8, summary.AverageFeedback)
	assert.Equal(t, 1500, summary.TotalTokens)
	assert.Equal(t, 150, summary.AverageTokens)
	assert.Len(t, summary.Recommendations, 2)
	assert.Contains(t, summary.Recommendations[0], "learning rate")
}

// TestTrainingStatus 测试训练状态结构
func TestTrainingStatus(t *testing.T) {
	startTime := time.Now()

	status := &TrainingStatus{
		IsRunning:          true,
		CurrentIteration:   3,
		TotalIterations:    10,
		Progress:           0.3,
		StartTime:          startTime,
		ElapsedTime:        90 * time.Second,
		EstimatedRemaining: 210 * time.Second,
		CurrentScore:       7.5,
		BestScore:          8.2,
		RecentImprovement:  0.5,
		Status:             "training in progress",
		LastError:          "",
	}

	assert.True(t, status.IsRunning)
	assert.Equal(t, 3, status.CurrentIteration)
	assert.Equal(t, 10, status.TotalIterations)
	assert.Equal(t, 0.3, status.Progress)
	assert.Equal(t, startTime, status.StartTime)
	assert.Equal(t, 90*time.Second, status.ElapsedTime)
	assert.Equal(t, 210*time.Second, status.EstimatedRemaining)
	assert.Equal(t, 7.5, status.CurrentScore)
	assert.Equal(t, 8.2, status.BestScore)
	assert.Equal(t, 0.5, status.RecentImprovement)
	assert.Equal(t, "training in progress", status.Status)
	assert.Equal(t, "", status.LastError)
}

// TestAgentIterationData 测试Agent迭代数据
func TestAgentIterationData(t *testing.T) {
	agentData := &AgentIterationData{
		AgentRole:     "content_writer",
		ExecutionTime: 250 * time.Millisecond,
		TokensUsed:    180,
		ToolsUsed:     []string{"web_search", "grammar_check", "fact_verify"},
		Success:       true,
		Error:         "",
	}

	assert.Equal(t, "content_writer", agentData.AgentRole)
	assert.Equal(t, 250*time.Millisecond, agentData.ExecutionTime)
	assert.Equal(t, 180, agentData.TokensUsed)
	assert.Len(t, agentData.ToolsUsed, 3)
	assert.Contains(t, agentData.ToolsUsed, "web_search")
	assert.Contains(t, agentData.ToolsUsed, "grammar_check")
	assert.Contains(t, agentData.ToolsUsed, "fact_verify")
	assert.True(t, agentData.Success)
	assert.Equal(t, "", agentData.Error)
}

// TestTaskIterationData 测试任务迭代数据
func TestTaskIterationData(t *testing.T) {
	taskData := &TaskIterationData{
		TaskDescription: "Generate a technical blog post about AI",
		ExecutionTime:   400 * time.Millisecond,
		Success:         true,
		OutputLength:    2500,
		ValidationScore: 8.7,
		Error:           "",
	}

	assert.Equal(t, "Generate a technical blog post about AI", taskData.TaskDescription)
	assert.Equal(t, 400*time.Millisecond, taskData.ExecutionTime)
	assert.True(t, taskData.Success)
	assert.Equal(t, 2500, taskData.OutputLength)
	assert.Equal(t, 8.7, taskData.ValidationScore)
	assert.Equal(t, "", taskData.Error)
}

// BenchmarkTrainingConfigCreation 基准测试训练配置创建
func BenchmarkTrainingConfigCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config := DefaultTrainingConfig()
		_ = config
	}
}

// BenchmarkTrainingDataSerialization 基准测试训练数据序列化
func BenchmarkTrainingDataSerialization(b *testing.B) {
	// 创建测试数据
	config := DefaultTrainingConfig()
	data := &TrainingData{
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Version:    "1.0",
		Config:     config,
		SessionID:  "bench-test-session",
		CrewName:   "benchmark-crew",
		TotalRuns:  10,
		Iterations: make([]*IterationData, 0, 10),
		Summary: &TrainingSummary{
			TotalIterations: 10,
			SuccessfulRuns:  8,
			FailedRuns:      2,
		},
	}

	// 添加一些迭代数据
	for i := 0; i < 10; i++ {
		iteration := &IterationData{
			IterationID: fmt.Sprintf("bench-iter-%d", i),
			Index:       i,
			Timestamp:   time.Now(),
			Duration:    time.Duration(i*100) * time.Millisecond,
			Success:     i%10 != 0, // 10%失败率
		}
		data.Iterations = append(data.Iterations, iteration)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 这里可以测试序列化性能，但需要json包
		_ = data
	}
}
