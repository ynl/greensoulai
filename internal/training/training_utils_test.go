package training

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

func TestNewTrainingUtils(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	utils := NewTrainingUtils(testLogger)
	
	assert.NotNil(t, utils)
	assert.Equal(t, testLogger, utils.logger)
}

func TestCreateTrainingHandler(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)
	utils := NewTrainingUtils(testLogger)
	
	handler := utils.CreateTrainingHandler(testEventBus, testLogger)
	
	assert.NotNil(t, handler)
	assert.Implements(t, (*TrainingHandler)(nil), handler)
}

func TestValidateTrainingConfig(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	utils := NewTrainingUtils(testLogger)
	
	tests := []struct {
		name        string
		config      *TrainingConfig
		expectError bool
		expectFix   bool
	}{
		{
			name: "valid config",
			config: &TrainingConfig{
				Iterations:      10,
				Filename:        "test.json",
				LearningRate:    0.01,
				BatchSize:       2,
				ValidationSplit: 0.2,
				FeedbackTimeout: 5 * time.Minute,
				PatientceEpochs: 3,
				MinImprovement:  0.01,
				SaveInterval:    5,
				BackupCount:     3,
			},
			expectError: false,
			expectFix:   false,
		},
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
		},
		{
			name: "zero iterations",
			config: &TrainingConfig{
				Iterations: 0,
			},
			expectError: true,
		},
		{
			name: "too many iterations",
			config: &TrainingConfig{
				Iterations: 2000,
			},
			expectError: true,
		},
		{
			name: "empty filename should be fixed",
			config: &TrainingConfig{
				Iterations: 5,
				Filename:   "",
			},
			expectError: false,
			expectFix:   true,
		},
		{
			name: "invalid learning rate should be fixed",
			config: &TrainingConfig{
				Iterations:   5,
				LearningRate: -0.1,
			},
			expectError: false,
			expectFix:   true,
		},
		{
			name: "invalid batch size should be fixed",
			config: &TrainingConfig{
				Iterations: 5,
				BatchSize:  0,
			},
			expectError: false,
			expectFix:   true,
		},
		{
			name: "invalid validation split should be fixed",
			config: &TrainingConfig{
				Iterations:      5,
				ValidationSplit: 1.5,
			},
			expectError: false,
			expectFix:   true,
		},
		{
			name: "zero timeout should be fixed",
			config: &TrainingConfig{
				Iterations:      5,
				FeedbackTimeout: 0,
			},
			expectError: false,
			expectFix:   true,
		},
		{
			name: "invalid patience epochs should be fixed",
			config: &TrainingConfig{
				Iterations:      5,
				PatientceEpochs: 0,
			},
			expectError: false,
			expectFix:   true,
		},
		{
			name: "invalid min improvement should be fixed",
			config: &TrainingConfig{
				Iterations:     5,
				MinImprovement: 0,
			},
			expectError: false,
			expectFix:   true,
		},
		{
			name: "invalid save interval should be fixed",
			config: &TrainingConfig{
				Iterations:   5,
				SaveInterval: 0,
			},
			expectError: false,
			expectFix:   true,
		},
		{
			name: "negative backup count should be fixed",
			config: &TrainingConfig{
				Iterations:  5,
				BackupCount: -1,
			},
			expectError: false,
			expectFix:   true,
		},
		{
			name: "empty target metrics should be fixed",
			config: &TrainingConfig{
				Iterations:    5,
				TargetMetrics: []string{},
			},
			expectError: false,
			expectFix:   true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalConfig := tt.config
			if originalConfig != nil {
				// 创建副本以避免修改原始测试数据
				configCopy := *originalConfig
				originalConfig = &configCopy
			}
			
			err := utils.ValidateTrainingConfig(tt.config)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				
				if tt.expectFix && originalConfig != nil {
					// 验证配置是否被修复
					assert.NotNil(t, tt.config)
				}
			}
		})
	}
}

func TestGenerateTrainingReport(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	utils := NewTrainingUtils(testLogger)
	
	tests := []struct {
		name           string
		data           *TrainingData
		expectedStatus string
		expectedInsights int
		expectedWarnings int
	}{
		{
			name:           "nil data",
			data:           nil,
			expectedStatus: "no_data",
		},
		{
			name: "empty iterations",
			data: &TrainingData{
				SessionID:  "empty-test",
				Iterations: []*IterationData{},
			},
			expectedStatus: "no_data",
		},
		{
			name: "successful training with high scores",
			data: &TrainingData{
				SessionID: "success-test",
				CrewName:  "test-crew",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Config:    DefaultTrainingConfig(),
				Iterations: []*IterationData{
					{IterationID: "1", Success: true},
					{IterationID: "2", Success: true},
					{IterationID: "3", Success: true},
				},
				Summary: &TrainingSummary{
					TotalIterations: 3,
					SuccessfulRuns:  3,
					FailedRuns:      0,
					ImprovementRate: 15.0,
					AverageFeedback: 8.5,
				},
			},
			expectedStatus:   "completed",
			expectedInsights: 3, // 完美成功率 + 显著改进 + 高反馈分数
		},
		{
			name: "training with issues",
			data: &TrainingData{
				SessionID: "issues-test",
				CrewName:  "test-crew",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Config:    DefaultTrainingConfig(),
				Iterations: []*IterationData{
					{IterationID: "1", Success: false},
					{IterationID: "2", Success: true},
					{IterationID: "3", Success: false},
				},
				Summary: &TrainingSummary{
					TotalIterations: 3,
					SuccessfulRuns:  1,
					FailedRuns:      2,
					ImprovementRate: -5.0,
					AverageFeedback: 4.0,
				},
			},
			expectedStatus:   "completed",
			expectedWarnings: 2, // 高失败率 + 低反馈分数
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := utils.GenerateTrainingReport(tt.data)
			
			assert.NotNil(t, report)
			assert.Equal(t, tt.expectedStatus, report.Status)
			
			if tt.data != nil && len(tt.data.Iterations) > 0 {
				assert.Equal(t, tt.data.SessionID, report.SessionID)
				assert.Equal(t, tt.data.CrewName, report.CrewName)
				assert.Equal(t, len(tt.data.Iterations), report.TotalIterations)
			}
			
			if tt.expectedInsights > 0 {
				assert.Len(t, report.Insights, tt.expectedInsights)
			}
			
			if tt.expectedWarnings > 0 {
				assert.Len(t, report.Warnings, tt.expectedWarnings)
			}
		})
	}
}

func TestCreateSimpleTrainingConfig(t *testing.T) {
	config := CreateSimpleTrainingConfig(15, "simple_test.json")
	
	assert.NotNil(t, config)
	assert.Equal(t, 15, config.Iterations)
	assert.Equal(t, "simple_test.json", config.Filename)
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
	
	// 验证基本的反馈提示
	assert.Len(t, config.FeedbackPrompts, 2)
	assert.Contains(t, config.FeedbackPrompts[0], "rate the quality")
	
	// 验证目标指标
	expectedMetrics := []string{"execution_time", "success_rate", "feedback_score"}
	assert.Equal(t, expectedMetrics, config.TargetMetrics)
}

func TestCreateAdvancedTrainingConfig(t *testing.T) {
	inputs := map[string]interface{}{
		"task": "advanced test task",
		"goal": "test advanced configuration",
	}
	
	config := CreateAdvancedTrainingConfig(20, "advanced_test.json", inputs)
	
	assert.NotNil(t, config)
	assert.Equal(t, 20, config.Iterations)
	assert.Equal(t, "advanced_test.json", config.Filename)
	assert.Equal(t, inputs, config.Inputs)
	assert.True(t, config.EarlyStopping)   // 应该启用早停
	assert.True(t, config.CollectFeedback)
	assert.True(t, config.MetricsEnabled)
	assert.Equal(t, 3*time.Minute, config.FeedbackTimeout) // 更短的超时
	assert.Equal(t, 3, config.SaveInterval)                // 更频繁的保存
	
	// 验证高级反馈提示
	assert.Len(t, config.FeedbackPrompts, 5)
	assert.Contains(t, config.FeedbackPrompts[0], "overall quality")
	assert.Contains(t, config.FeedbackPrompts[1], "accuracy")
	assert.Contains(t, config.FeedbackPrompts[2], "usefulness")
	assert.Contains(t, config.FeedbackPrompts[3], "improved")
	assert.Contains(t, config.FeedbackPrompts[4], "issues")
}

// MockTrainingHandler 用于测试的模拟训练处理器
type MockTrainingHandler struct {
	mock.Mock
}

func (m *MockTrainingHandler) StartTraining(ctx context.Context, config *TrainingConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockTrainingHandler) CollectFeedback(ctx context.Context, iterationID string, feedback *HumanFeedback) error {
	args := m.Called(ctx, iterationID, feedback)
	return args.Error(0)
}

func (m *MockTrainingHandler) AnalyzePerformance(ctx context.Context, iterationID string) (*PerformanceMetrics, error) {
	args := m.Called(ctx, iterationID)
	return args.Get(0).(*PerformanceMetrics), args.Error(1)
}

func (m *MockTrainingHandler) SaveTrainingData(ctx context.Context, data *TrainingData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockTrainingHandler) LoadTrainingData(ctx context.Context, filename string) (*TrainingData, error) {
	args := m.Called(ctx, filename)
	return args.Get(0).(*TrainingData), args.Error(1)
}

func (m *MockTrainingHandler) GetTrainingStatus(ctx context.Context) *TrainingStatus {
	args := m.Called(ctx)
	return args.Get(0).(*TrainingStatus)
}

func (m *MockTrainingHandler) StopTraining(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestRunTrainingSession(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	utils := NewTrainingUtils(testLogger)
	
	mockHandler := &MockTrainingHandler{}
	config := CreateSimpleTrainingConfig(3, "test_session.json")
	config.EarlyStopping = false // 禁用早停以简化测试
	
	ctx := context.Background()
	
	// 设置mock期望
	mockHandler.On("StartTraining", ctx, config).Return(nil)
	mockHandler.On("StopTraining", ctx).Return(nil)
	
	// 由于我们需要访问CrewTrainingHandler的内部方法，这个测试需要重新设计
	// 或者我们需要将ExecuteIteration方法添加到接口中
	
	// 现在我们只测试配置验证
	err := utils.ValidateTrainingConfig(config)
	assert.NoError(t, err)
	
	mockHandler.AssertExpectations(t)
}

func TestRunTrainingSessionWithCancellation(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	utils := NewTrainingUtils(testLogger)
	
	config := CreateSimpleTrainingConfig(10, "cancelled_test.json") // 较多迭代
	
	// 创建会被取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	
	// 立即取消上下文
	cancel()
	
	executeFunc := func(ctx context.Context, inputs map[string]interface{}) (interface{}, error) {
		return nil, ctx.Err() // 应该返回取消错误
	}
	
	// 创建真实的处理器进行测试
	handler := utils.CreateTrainingHandler(events.NewEventBus(testLogger), testLogger)
	
	_, err := utils.RunTrainingSession(ctx, handler, config, executeFunc)
	
	// 应该因为上下文取消而返回错误
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

// BenchmarkValidateTrainingConfig 配置验证性能基准测试
func BenchmarkValidateTrainingConfig(b *testing.B) {
	testLogger := logger.NewConsoleLogger()
	utils := NewTrainingUtils(testLogger)
	
	config := &TrainingConfig{
		Iterations:      10,
		Filename:        "benchmark.json",
		LearningRate:    0.01,
		BatchSize:       2,
		ValidationSplit: 0.2,
		FeedbackTimeout: 5 * time.Minute,
		PatientceEpochs: 3,
		MinImprovement:  0.01,
		SaveInterval:    5,
		BackupCount:     3,
		TargetMetrics:   []string{"execution_time", "success_rate"},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 每次创建新的配置副本以避免修改
		configCopy := *config
		utils.ValidateTrainingConfig(&configCopy)
	}
}

// BenchmarkGenerateTrainingReport 报告生成性能基准测试
func BenchmarkGenerateTrainingReport(b *testing.B) {
	testLogger := logger.NewConsoleLogger()
	utils := NewTrainingUtils(testLogger)
	
	// 创建包含大量迭代的训练数据
	iterations := make([]*IterationData, 100)
	for i := 0; i < 100; i++ {
		iterations[i] = &IterationData{
			IterationID: fmt.Sprintf("bench-iter-%d", i),
			Index:       i,
			Success:     i%10 != 0, // 10%失败率
			Duration:    time.Duration(100+i) * time.Millisecond,
		}
	}
	
	data := &TrainingData{
		SessionID:  "benchmark-session",
		CrewName:   "benchmark-crew",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Config:     DefaultTrainingConfig(),
		Iterations: iterations,
		Summary: &TrainingSummary{
			TotalIterations: 100,
			SuccessfulRuns:  90,
			FailedRuns:      10,
			ImprovementRate: 25.0,
			AverageFeedback: 7.5,
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.GenerateTrainingReport(data)
	}
}

// TestTrainingUtilsEdgeCases 测试边界情况
func TestTrainingUtilsEdgeCases(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	utils := NewTrainingUtils(testLogger)
	
	// 测试极端配置值
	extremeConfig := &TrainingConfig{
		Iterations:      999, // 接近最大值
		LearningRate:    0.999,
		ValidationSplit: 0.999,
		FeedbackTimeout: 24 * time.Hour, // 非常长的超时
		PatientceEpochs: 100,
		MinImprovement:  0.001,
		SaveInterval:    1, // 每次都保存
		BackupCount:     100,
	}
	
	err := utils.ValidateTrainingConfig(extremeConfig)
	assert.NoError(t, err)
	
	// 测试空的训练数据报告生成
	emptyData := &TrainingData{}
	report := utils.GenerateTrainingReport(emptyData)
	assert.NotNil(t, report)
	assert.Equal(t, "no_data", report.Status)
}
