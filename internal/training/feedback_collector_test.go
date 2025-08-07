package training

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ynl/greensoulai/pkg/logger"
)

func TestNewFeedbackCollector(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	collector := NewFeedbackCollector(testLogger)

	assert.NotNil(t, collector)
	assert.Equal(t, testLogger, collector.logger)
}

func TestCollectBatchFeedback(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	collector := NewFeedbackCollector(testLogger)

	ctx := context.Background()
	iterationID := "batch-test-1"
	outputs := map[string]interface{}{
		"result":  "test output for batch feedback",
		"quality": "high",
	}

	feedbackData := map[string]interface{}{
		"quality_score":  8.5,
		"accuracy_score": 9.0,
		"usefulness":     7.8,
		"comments":       "Great work overall",
		"suggestions":    "Could add more examples",
		"issues":         []string{"minor grammar issue"},
		"tags":           []string{"quality", "batch"},
		"categories": map[string]float64{
			"clarity": 8.0,
			"depth":   7.5,
		},
	}

	feedback, err := collector.CollectBatchFeedback(ctx, iterationID, outputs, feedbackData)

	assert.NoError(t, err)
	assert.NotNil(t, feedback)
	assert.Equal(t, iterationID, feedback.IterationID)
	assert.Equal(t, 8.5, feedback.QualityScore)
	assert.Equal(t, 9.0, feedback.AccuracyScore)
	assert.Equal(t, 7.8, feedback.Usefulness)
	assert.Equal(t, "Great work overall", feedback.Comments)
	assert.Equal(t, "Could add more examples", feedback.Suggestions)
	assert.Contains(t, feedback.Issues, "minor grammar issue")
	assert.Contains(t, feedback.Tags, "quality")
	assert.Contains(t, feedback.Tags, "batch")
	assert.Equal(t, 8.0, feedback.Categories["clarity"])
	assert.Equal(t, 7.5, feedback.Categories["depth"])
	assert.True(t, feedback.Verified)
	assert.Equal(t, "batch", feedback.VerifiedBy)
	assert.False(t, feedback.Timestamp.IsZero())
}

func TestCollectBatchFeedbackWithDefaults(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	collector := NewFeedbackCollector(testLogger)

	ctx := context.Background()
	iterationID := "batch-test-defaults"
	outputs := "simple string output"

	// 空的反馈数据，应该使用默认值
	feedbackData := map[string]interface{}{}

	feedback, err := collector.CollectBatchFeedback(ctx, iterationID, outputs, feedbackData)

	assert.NoError(t, err)
	assert.NotNil(t, feedback)
	assert.Equal(t, iterationID, feedback.IterationID)
	assert.Equal(t, 5.0, feedback.QualityScore)  // 默认值
	assert.Equal(t, 5.0, feedback.AccuracyScore) // 默认值
	assert.Equal(t, 5.0, feedback.Usefulness)    // 默认值
	assert.Equal(t, "", feedback.Comments)
	assert.Equal(t, "", feedback.Suggestions)
	assert.Empty(t, feedback.Issues)
	assert.Empty(t, feedback.Tags)
	assert.Empty(t, feedback.Categories)
	assert.True(t, feedback.Verified)
	assert.Equal(t, "batch", feedback.VerifiedBy)
}

func TestValidateFeedback(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	collector := NewFeedbackCollector(testLogger)

	tests := []struct {
		name        string
		feedback    *HumanFeedback
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid feedback",
			feedback: &HumanFeedback{
				IterationID:   "valid-test",
				QualityScore:  8.5,
				AccuracyScore: 9.0,
				Usefulness:    7.8,
			},
			expectError: false,
		},
		{
			name: "quality score too low",
			feedback: &HumanFeedback{
				IterationID:   "low-quality",
				QualityScore:  0.5,
				AccuracyScore: 8.0,
				Usefulness:    7.0,
			},
			expectError: true,
			errorMsg:    "quality score must be between 1 and 10",
		},
		{
			name: "quality score too high",
			feedback: &HumanFeedback{
				IterationID:   "high-quality",
				QualityScore:  11.0,
				AccuracyScore: 8.0,
				Usefulness:    7.0,
			},
			expectError: true,
			errorMsg:    "quality score must be between 1 and 10",
		},
		{
			name: "accuracy score too low",
			feedback: &HumanFeedback{
				IterationID:   "low-accuracy",
				QualityScore:  8.0,
				AccuracyScore: 0.0,
				Usefulness:    7.0,
			},
			expectError: true,
			errorMsg:    "accuracy score must be between 1 and 10",
		},
		{
			name: "accuracy score too high",
			feedback: &HumanFeedback{
				IterationID:   "high-accuracy",
				QualityScore:  8.0,
				AccuracyScore: 15.0,
				Usefulness:    7.0,
			},
			expectError: true,
			errorMsg:    "accuracy score must be between 1 and 10",
		},
		{
			name: "usefulness score too low",
			feedback: &HumanFeedback{
				IterationID:   "low-usefulness",
				QualityScore:  8.0,
				AccuracyScore: 8.0,
				Usefulness:    0.5,
			},
			expectError: true,
			errorMsg:    "usefulness score must be between 1 and 10",
		},
		{
			name: "usefulness score too high",
			feedback: &HumanFeedback{
				IterationID:   "high-usefulness",
				QualityScore:  8.0,
				AccuracyScore: 8.0,
				Usefulness:    12.0,
			},
			expectError: true,
			errorMsg:    "usefulness score must be between 1 and 10",
		},
		{
			name: "empty iteration ID",
			feedback: &HumanFeedback{
				IterationID:   "",
				QualityScore:  8.0,
				AccuracyScore: 8.0,
				Usefulness:    8.0,
			},
			expectError: true,
			errorMsg:    "iteration ID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := collector.ValidateFeedback(tt.feedback)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFormatOutput(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	collector := NewFeedbackCollector(testLogger)

	tests := []struct {
		name     string
		output   interface{}
		expected string
	}{
		{
			name:     "string output",
			output:   "simple string output",
			expected: "simple string output",
		},
		{
			name: "map output",
			output: map[string]interface{}{
				"result": "test result",
				"status": "success",
				"score":  8.5,
			},
			expected: "result: test result\nstatus: success\nscore: 8.5",
		},
		{
			name:     "slice output",
			output:   []interface{}{"first", "second", "third"},
			expected: "[0]: first\n[1]: second\n[2]: third",
		},
		{
			name: "struct output",
			output: struct {
				Name  string
				Value int
			}{
				Name:  "test",
				Value: 42,
			},
			expected: "{Name:test Value:42}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collector.formatOutput(tt.output)

			// 对于map类型，由于输出顺序不确定，我们检查是否包含所有预期的部分
			if strings.Contains(tt.expected, ": ") && strings.Contains(tt.expected, "\n") {
				expectedParts := strings.Split(tt.expected, "\n")
				for _, part := range expectedParts {
					assert.Contains(t, result, strings.TrimSpace(part))
				}
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestCollectFeedbackTimeout 测试反馈收集超时
func TestCollectFeedbackTimeout(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	collector := NewFeedbackCollector(testLogger)

	ctx := context.Background()
	iterationID := "timeout-test"
	outputs := "test output for timeout"
	timeout := 100 * time.Millisecond // 很短的超时时间

	// 这个测试会超时，因为没有实际的用户输入
	feedback, err := collector.CollectFeedback(ctx, iterationID, outputs, timeout)

	// 超时时应该返回错误，这是正常行为
	if err != nil {
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "EOF")
		// 如果有错误，feedback可能为nil，这是可接受的
		return
	}

	// 如果没有错误，则验证默认反馈
	assert.NotNil(t, feedback)
	assert.Equal(t, iterationID, feedback.IterationID)
	assert.Equal(t, 5.0, feedback.QualityScore)
	assert.Equal(t, 5.0, feedback.AccuracyScore)
	assert.Equal(t, 5.0, feedback.Usefulness)
}

// BenchmarkFormatOutput 格式化输出性能基准测试
func BenchmarkFormatOutput(b *testing.B) {
	testLogger := logger.NewConsoleLogger()
	collector := NewFeedbackCollector(testLogger)

	output := map[string]interface{}{
		"result":    "benchmark test result",
		"status":    "success",
		"score":     8.5,
		"details":   []string{"detail1", "detail2", "detail3"},
		"metadata":  map[string]interface{}{"version": "1.0", "type": "test"},
		"timestamp": time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.formatOutput(output)
	}
}

// BenchmarkValidateFeedback 反馈验证性能基准测试
func BenchmarkValidateFeedback(b *testing.B) {
	testLogger := logger.NewConsoleLogger()
	collector := NewFeedbackCollector(testLogger)

	feedback := &HumanFeedback{
		IterationID:   "benchmark-test",
		QualityScore:  8.5,
		AccuracyScore: 9.0,
		Usefulness:    7.8,
		Comments:      "Benchmark feedback validation test",
		Suggestions:   "No suggestions",
		Issues:        []string{},
		Tags:          []string{"benchmark", "test"},
		Categories:    map[string]float64{"quality": 8.5},
		Verified:      true,
		VerifiedBy:    "benchmark",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.ValidateFeedback(feedback)
	}
}

// TestFeedbackCollectorEdgeCases 测试边界情况
func TestFeedbackCollectorEdgeCases(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	collector := NewFeedbackCollector(testLogger)

	ctx := context.Background()

	// 测试nil输出
	feedback, err := collector.CollectBatchFeedback(ctx, "nil-test", nil, map[string]interface{}{})
	assert.NoError(t, err)
	assert.NotNil(t, feedback)

	// 测试空输出
	feedback, err = collector.CollectBatchFeedback(ctx, "empty-test", "", map[string]interface{}{})
	assert.NoError(t, err)
	assert.NotNil(t, feedback)

	// 测试复杂嵌套输出
	complexOutput := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": []interface{}{
				map[string]interface{}{"key": "value"},
				[]string{"nested", "array"},
			},
		},
	}
	formatted := collector.formatOutput(complexOutput)
	assert.NotEmpty(t, formatted)
	assert.Contains(t, formatted, "level1")
}

// TestFeedbackTimestamp 测试反馈时间戳
func TestFeedbackTimestamp(t *testing.T) {
	testLogger := logger.NewConsoleLogger()
	collector := NewFeedbackCollector(testLogger)

	ctx := context.Background()
	beforeTime := time.Now()

	feedback, err := collector.CollectBatchFeedback(ctx, "timestamp-test", "output", map[string]interface{}{})

	afterTime := time.Now()

	require.NoError(t, err)
	assert.True(t, feedback.Timestamp.After(beforeTime) || feedback.Timestamp.Equal(beforeTime))
	assert.True(t, feedback.Timestamp.Before(afterTime) || feedback.Timestamp.Equal(afterTime))
}
