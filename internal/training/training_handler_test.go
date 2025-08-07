package training

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// MockEventBus 模拟事件总线
type MockEventBus struct {
	mock.Mock
}

func (m *MockEventBus) Emit(ctx context.Context, source interface{}, event events.Event) error {
	args := m.Called(ctx, source, event)
	return args.Error(0)
}

func (m *MockEventBus) Subscribe(eventType string, handler events.EventHandler) error {
	args := m.Called(eventType, handler)
	return args.Error(0)
}

func (m *MockEventBus) Unsubscribe(eventType string, handler events.EventHandler) error {
	args := m.Called(eventType, handler)
	return args.Error(0)
}

// MockLogger 模拟日志器
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, fields ...logger.Field) {
	args := []interface{}{msg}
	for _, field := range fields {
		args = append(args, field)
	}
	m.Called(args...)
}

func (m *MockLogger) Info(msg string, fields ...logger.Field) {
	args := []interface{}{msg}
	for _, field := range fields {
		args = append(args, field)
	}
	m.Called(args...)
}

func (m *MockLogger) Warn(msg string, fields ...logger.Field) {
	args := []interface{}{msg}
	for _, field := range fields {
		args = append(args, field)
	}
	m.Called(args...)
}

func (m *MockLogger) Error(msg string, fields ...logger.Field) {
	args := []interface{}{msg}
	for _, field := range fields {
		args = append(args, field)
	}
	m.Called(args...)
}

func (m *MockLogger) Fatal(msg string, fields ...logger.Field) {
	args := []interface{}{msg}
	for _, field := range fields {
		args = append(args, field)
	}
	m.Called(args...)
}

// TestNewCrewTrainingHandler 测试创建训练处理器
func TestNewCrewTrainingHandler(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)
	
	handler := NewCrewTrainingHandler(testEventBus, testLogger)
	
	assert.NotNil(t, handler)
	assert.Equal(t, testLogger, handler.logger)
	assert.Equal(t, testEventBus, handler.eventBus)
	assert.NotNil(t, handler.status)
	assert.False(t, handler.status.IsRunning)
	assert.Equal(t, "initialized", handler.status.Status)
	assert.NotNil(t, handler.feedbackCollector)
	assert.NotNil(t, handler.metricsAnalyzer)
}

// TestStartTraining 测试开始训练
func TestStartTraining(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)
	handler := NewCrewTrainingHandler(testEventBus, testLogger)
	
	config := &TrainingConfig{
		Iterations:      5,
		Filename:        "test_training.json",
		CollectFeedback: true,
		MetricsEnabled:  true,
		AutoSave:        true,
		Inputs: map[string]interface{}{
			"task": "test task",
		},
	}
	
	ctx := context.Background()
	err := handler.StartTraining(ctx, config)
	
	assert.NoError(t, err)
	
	// 验证训练状态
	status := handler.GetTrainingStatus(ctx)
	assert.True(t, status.IsRunning)
	assert.Equal(t, 5, status.TotalIterations)
	assert.Equal(t, "starting", status.Status)
	
	// 验证训练数据初始化
	assert.NotNil(t, handler.trainingData)
	assert.Equal(t, config, handler.trainingData.Config)
	assert.NotEmpty(t, handler.trainingData.SessionID)
	assert.Equal(t, "1.0", handler.trainingData.Version)
}

// TestStartTrainingAlreadyRunning 测试重复开始训练
func TestStartTrainingAlreadyRunning(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)
	handler := NewCrewTrainingHandler(testEventBus, testLogger)
	
	config := DefaultTrainingConfig()
	ctx := context.Background()
	
	// 第一次开始训练
	err := handler.StartTraining(ctx, config)
	assert.NoError(t, err)
	
	// 第二次开始训练应该失败
	err = handler.StartTraining(ctx, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")
}

// TestExecuteIteration 测试执行迭代
func TestExecuteIteration(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)
	handler := NewCrewTrainingHandler(testEventBus, testLogger)
	
	config := DefaultTrainingConfig()
	config.CollectFeedback = false // 禁用反馈以简化测试
	config.MetricsEnabled = false
	
	ctx := context.Background()
	err := handler.StartTraining(ctx, config)
	require.NoError(t, err)
	
	// 创建模拟执行函数
	executeFunc := func(ctx context.Context, inputs map[string]interface{}) (interface{}, error) {
		return map[string]interface{}{
			"result": "test execution result",
			"status": "success",
		}, nil
	}
	
	iteration, err := handler.ExecuteIteration(ctx, executeFunc, 0)
	
	assert.NoError(t, err)
	assert.NotNil(t, iteration)
	assert.Equal(t, 0, iteration.Index)
	assert.True(t, iteration.Success)
	assert.NotEmpty(t, iteration.IterationID)
	assert.NotNil(t, iteration.Outputs)
	assert.Greater(t, iteration.Duration, time.Duration(0))
	
	// 验证输出内容
	outputs, ok := iteration.Outputs.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "test execution result", outputs["result"])
	assert.Equal(t, "success", outputs["status"])
}

// TestExecuteIterationWithError 测试执行迭代时发生错误
func TestExecuteIterationWithError(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)
	handler := NewCrewTrainingHandler(testEventBus, testLogger)
	
	config := DefaultTrainingConfig()
	config.CollectFeedback = false
	config.MetricsEnabled = false
	
	ctx := context.Background()
	err := handler.StartTraining(ctx, config)
	require.NoError(t, err)
	
	// 创建会失败的执行函数
	executeFunc := func(ctx context.Context, inputs map[string]interface{}) (interface{}, error) {
		return nil, fmt.Errorf("execution failed for testing")
	}
	
	iteration, err := handler.ExecuteIteration(ctx, executeFunc, 0)
	
	assert.NoError(t, err) // ExecuteIteration本身不应该返回错误
	assert.NotNil(t, iteration)
	assert.False(t, iteration.Success)
	assert.Equal(t, "execution failed for testing", iteration.Error)
	assert.Nil(t, iteration.Outputs)
}

// TestGetTrainingStatus 测试获取训练状态
func TestGetTrainingStatus(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)
	handler := NewCrewTrainingHandler(testEventBus, testLogger)
	
	ctx := context.Background()
	
	// 测试初始状态
	status := handler.GetTrainingStatus(ctx)
	assert.False(t, status.IsRunning)
	assert.Equal(t, "initialized", status.Status)
	assert.Equal(t, 0, status.CurrentIteration)
	
	// 开始训练后的状态
	config := DefaultTrainingConfig()
	err := handler.StartTraining(ctx, config)
	require.NoError(t, err)
	
	status = handler.GetTrainingStatus(ctx)
	assert.True(t, status.IsRunning)
	assert.Equal(t, config.Iterations, status.TotalIterations)
	assert.Equal(t, "starting", status.Status)
	assert.NotEqual(t, time.Time{}, status.StartTime)
}

// TestStopTraining 测试停止训练
func TestStopTraining(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)
	handler := NewCrewTrainingHandler(testEventBus, testLogger)
	
	ctx := context.Background()
	config := DefaultTrainingConfig()
	config.AutoSave = false // 禁用自动保存以简化测试
	
	// 开始训练
	err := handler.StartTraining(ctx, config)
	require.NoError(t, err)
	
	// 验证训练正在运行
	status := handler.GetTrainingStatus(ctx)
	assert.True(t, status.IsRunning)
	
	// 停止训练
	err = handler.StopTraining(ctx)
	assert.NoError(t, err)
	
	// 验证训练已停止
	status = handler.GetTrainingStatus(ctx)
	assert.False(t, status.IsRunning)
}

// TestStopTrainingNotRunning 测试停止未运行的训练
func TestStopTrainingNotRunning(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)
	handler := NewCrewTrainingHandler(testEventBus, testLogger)
	
	ctx := context.Background()
	
	// 尝试停止未开始的训练
	err := handler.StopTraining(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

// TestSaveTrainingData 测试保存训练数据
func TestSaveTrainingData(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)
	handler := NewCrewTrainingHandler(testEventBus, testLogger)
	
	// 创建测试目录
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "test_training_data.json")
	
	// 创建测试数据
	trainingData := &TrainingData{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   "1.0",
		Config:    DefaultTrainingConfig(),
		SessionID: "test-session-123",
		CrewName:  "test-crew",
		TotalRuns: 3,
		Iterations: []*IterationData{
			{
				IterationID: "iter-1",
				Index:       0,
				Success:     true,
				Duration:    100 * time.Millisecond,
			},
		},
		Summary: &TrainingSummary{
			TotalIterations: 1,
			SuccessfulRuns:  1,
			FailedRuns:      0,
		},
	}
	
	handler.trainingData = trainingData
	handler.config = &TrainingConfig{
		Filename:    filename,
		BackupCount: 0, // 禁用备份以简化测试
	}
	
	ctx := context.Background()
	err := handler.SaveTrainingData(ctx, trainingData)
	
	assert.NoError(t, err)
	
	// 验证文件是否存在
	_, err = os.Stat(filename)
	assert.NoError(t, err)
	
	// 验证可以加载数据
	loadedData, err := handler.LoadTrainingData(ctx, filename)
	assert.NoError(t, err)
	assert.Equal(t, trainingData.SessionID, loadedData.SessionID)
	assert.Equal(t, trainingData.Version, loadedData.Version)
	assert.Equal(t, len(trainingData.Iterations), len(loadedData.Iterations))
}

// TestLoadTrainingData 测试加载训练数据
func TestLoadTrainingData(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)
	handler := NewCrewTrainingHandler(testEventBus, testLogger)
	
	ctx := context.Background()
	
	// 测试加载不存在的文件
	_, err := handler.LoadTrainingData(ctx, "nonexistent.json")
	assert.Error(t, err)
	
	// 测试加载空文件名
	_, err = handler.LoadTrainingData(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "filename cannot be empty")
}

// TestCollectFeedback 测试收集反馈
func TestCollectFeedback(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)
	handler := NewCrewTrainingHandler(testEventBus, testLogger)
	
	// 初始化训练数据
	handler.trainingData = &TrainingData{
		Iterations: []*IterationData{
			{
				IterationID: "test-iteration-1",
				Index:       0,
				Success:     true,
			},
		},
	}
	
	feedback := &HumanFeedback{
		IterationID:   "test-iteration-1",
		QualityScore:  8.5,
		AccuracyScore: 9.0,
		Comments:      "Great job!",
	}
	
	ctx := context.Background()
	err := handler.CollectFeedback(ctx, "test-iteration-1", feedback)
	
	assert.NoError(t, err)
	assert.Equal(t, feedback, handler.trainingData.Iterations[0].Feedback)
}

// TestCollectFeedbackNotFound 测试收集不存在迭代的反馈
func TestCollectFeedbackNotFound(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)
	handler := NewCrewTrainingHandler(testEventBus, testLogger)
	
	handler.trainingData = &TrainingData{
		Iterations: []*IterationData{},
	}
	
	feedback := &HumanFeedback{
		IterationID: "nonexistent-iteration",
	}
	
	ctx := context.Background()
	err := handler.CollectFeedback(ctx, "nonexistent-iteration", feedback)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "iteration not found")
}

// TestCheckEarlyStop 测试早停检查
func TestCheckEarlyStop(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)
	handler := NewCrewTrainingHandler(testEventBus, testLogger)
	
	// 设置早停配置
	handler.config = &TrainingConfig{
		EarlyStopping:  true,
		PatientceEpochs: 3,
		MinImprovement: 0.1,
	}
	
	tests := []struct {
		name     string
		scores   []float64
		expected bool
	}{
		{
			name:     "not enough scores",
			scores:   []float64{7.0, 7.5},
			expected: false,
		},
		{
			name:     "good improvement",
			scores:   []float64{6.0, 7.0, 8.0},
			expected: false,
		},
		{
			name:     "no improvement",
			scores:   []float64{7.0, 7.0, 7.0},
			expected: true,
		},
		{
			name:     "declining scores",
			scores:   []float64{8.0, 7.5, 7.0},
			expected: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.CheckEarlyStop(tt.scores)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCheckEarlyStopDisabled 测试禁用早停
func TestCheckEarlyStopDisabled(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)
	handler := NewCrewTrainingHandler(testEventBus, testLogger)
	
	// 禁用早停
	handler.config = &TrainingConfig{
		EarlyStopping: false,
	}
	
	// 即使分数没有改进，也不应该触发早停
	result := handler.CheckEarlyStop([]float64{7.0, 7.0, 7.0, 7.0})
	assert.False(t, result)
}

// BenchmarkExecuteIteration 基准测试迭代执行
func BenchmarkExecuteIteration(b *testing.B) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)
	handler := NewCrewTrainingHandler(testEventBus, testLogger)
	
	config := DefaultTrainingConfig()
	config.CollectFeedback = false
	config.MetricsEnabled = false
	config.AutoSave = false
	
	ctx := context.Background()
	handler.StartTraining(ctx, config)
	
	executeFunc := func(ctx context.Context, inputs map[string]interface{}) (interface{}, error) {
		return map[string]interface{}{"result": "benchmark"}, nil
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ExecuteIteration(ctx, executeFunc, i)
	}
}

// TestConcurrentAccess 测试并发访问
func TestConcurrentAccess(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	testEventBus := events.NewEventBus(testLogger)
	handler := NewCrewTrainingHandler(testEventBus, testLogger)
	
	config := DefaultTrainingConfig()
	config.CollectFeedback = false
	config.MetricsEnabled = false
	
	ctx := context.Background()
	handler.StartTraining(ctx, config)
	
	numGoroutines := 5
	done := make(chan bool, numGoroutines)
	
	// 并发获取状态
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- true }()
			
			for j := 0; j < 10; j++ {
				status := handler.GetTrainingStatus(ctx)
				assert.NotNil(t, status)
			}
		}()
	}
	
	// 等待所有goroutines完成
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}
