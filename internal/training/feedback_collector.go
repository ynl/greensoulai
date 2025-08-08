package training

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ynl/greensoulai/pkg/logger"
)

// FeedbackCollector 人工反馈收集器
type FeedbackCollector struct {
	logger logger.Logger
}

// NewFeedbackCollector 创建新的反馈收集器
func NewFeedbackCollector(logger logger.Logger) *FeedbackCollector {
	return &FeedbackCollector{
		logger: logger,
	}
}

// CollectFeedback 收集人工反馈
func (fc *FeedbackCollector) CollectFeedback(ctx context.Context, iterationID string, outputs interface{}, timeout time.Duration) (*HumanFeedback, error) {
	fc.logger.Info("collecting human feedback",
		logger.Field{Key: "iteration_id", Value: iterationID},
		logger.Field{Key: "timeout", Value: timeout},
	)

	// 创建反馈对象
	feedback := &HumanFeedback{
		IterationID: iterationID,
		Timestamp:   time.Now(),
		Categories:  make(map[string]float64),
		Tags:        make([]string, 0),
		Issues:      make([]string, 0),
	}

	// 显示输出内容
	fmt.Printf("\n" + strings.Repeat("=", 80) + "\n")
	fmt.Printf("🤖 TRAINING ITERATION OUTPUT\n")
	fmt.Printf("Iteration ID: %s\n", iterationID)
	fmt.Printf(strings.Repeat("-", 80) + "\n")

	// 格式化输出显示
	outputStr := fc.formatOutput(outputs)
	fmt.Printf("Output:\n%s\n", outputStr)
	fmt.Printf(strings.Repeat("-", 80) + "\n")

	// 创建超时上下文
	feedbackCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 在goroutine中收集反馈
	feedbackChan := make(chan *HumanFeedback, 1)
	errorChan := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				errorChan <- fmt.Errorf("feedback collection panicked: %v", r)
			}
		}()

		collectedFeedback, err := fc.collectFeedbackInteractive(feedback)
		if err != nil {
			errorChan <- err
			return
		}

		feedbackChan <- collectedFeedback
	}()

	// 等待反馈或超时
	select {
	case feedback := <-feedbackChan:
		fc.logger.Info("feedback collected successfully",
			logger.Field{Key: "iteration_id", Value: iterationID},
			logger.Field{Key: "quality_score", Value: feedback.QualityScore},
		)
		return feedback, nil

	case err := <-errorChan:
		fc.logger.Error("feedback collection error",
			logger.Field{Key: "iteration_id", Value: iterationID},
			logger.Field{Key: "error", Value: err},
		)
		return nil, err

	case <-feedbackCtx.Done():
		fc.logger.Warn("feedback collection timeout",
			logger.Field{Key: "iteration_id", Value: iterationID},
			logger.Field{Key: "timeout", Value: timeout},
		)

		// 返回默认反馈而不是错误，这样训练可以继续
		return &HumanFeedback{
			IterationID:   iterationID,
			Timestamp:     time.Now(),
			QualityScore:  5.0, // 中性分数
			AccuracyScore: 5.0,
			Usefulness:    5.0,
			Comments:      "No feedback provided (timeout)",
			Categories:    make(map[string]float64),
			Tags:          []string{"timeout"},
			Issues:        []string{},
		}, nil
	}
}

// collectFeedbackInteractive 交互式收集反馈
func (fc *FeedbackCollector) collectFeedbackInteractive(feedback *HumanFeedback) (*HumanFeedback, error) {
	reader := bufio.NewReader(os.Stdin)

	// 收集质量分数
	fmt.Printf("\n📊 Please provide feedback (press Enter to skip any question):\n\n")

	qualityScore, err := fc.askForScore(reader, "Quality Score (1-10)", 5.0)
	if err != nil {
		return nil, fmt.Errorf("failed to get quality score: %w", err)
	}
	feedback.QualityScore = qualityScore

	// 收集准确性分数
	accuracyScore, err := fc.askForScore(reader, "Accuracy Score (1-10)", 5.0)
	if err != nil {
		return nil, fmt.Errorf("failed to get accuracy score: %w", err)
	}
	feedback.AccuracyScore = accuracyScore

	// 收集有用性分数
	usefulnessScore, err := fc.askForScore(reader, "Usefulness Score (1-10)", 5.0)
	if err != nil {
		return nil, fmt.Errorf("failed to get usefulness score: %w", err)
	}
	feedback.Usefulness = usefulnessScore

	// 收集文本反馈
	comments, err := fc.askForText(reader, "Comments/Feedback")
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}
	feedback.Comments = comments

	// 收集改进建议
	suggestions, err := fc.askForText(reader, "Suggestions for improvement")
	if err != nil {
		return nil, fmt.Errorf("failed to get suggestions: %w", err)
	}
	feedback.Suggestions = suggestions

	// 收集问题
	issues, err := fc.askForList(reader, "Issues/Problems (comma-separated)")
	if err != nil {
		return nil, fmt.Errorf("failed to get issues: %w", err)
	}
	feedback.Issues = issues

	// 收集标签
	tags, err := fc.askForList(reader, "Tags (comma-separated)")
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}
	feedback.Tags = tags

	// 设置验证信息
	feedback.Verified = true
	feedback.VerifiedBy = "human"

	fmt.Printf("\n✅ Feedback collected successfully!\n")
	fmt.Printf(strings.Repeat("=", 80) + "\n\n")

	return feedback, nil
}

// askForScore 询问评分
func (fc *FeedbackCollector) askForScore(reader *bufio.Reader, prompt string, defaultValue float64) (float64, error) {
	fmt.Printf("%s [default: %.1f]: ", prompt, defaultValue)

	input, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue, nil
	}

	score, err := strconv.ParseFloat(input, 64)
	if err != nil {
		fc.logger.Warn("invalid score input, using default",
			logger.Field{Key: "input", Value: input},
			logger.Field{Key: "default", Value: defaultValue},
		)
		return defaultValue, nil
	}

	// 确保分数在1-10范围内
	if score < 1 {
		score = 1
	} else if score > 10 {
		score = 10
	}

	return score, nil
}

// askForText 询问文本输入
func (fc *FeedbackCollector) askForText(reader *bufio.Reader, prompt string) (string, error) {
	fmt.Printf("%s: ", prompt)

	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(input), nil
}

// askForList 询问列表输入
func (fc *FeedbackCollector) askForList(reader *bufio.Reader, prompt string) ([]string, error) {
	text, err := fc.askForText(reader, prompt)
	if err != nil {
		return nil, err
	}

	if text == "" {
		return []string{}, nil
	}

	items := strings.Split(text, ",")
	result := make([]string, 0, len(items))

	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}

	return result, nil
}

// formatOutput 格式化输出内容
func (fc *FeedbackCollector) formatOutput(outputs interface{}) string {
	switch v := outputs.(type) {
	case string:
		return v
	case map[string]interface{}:
		var parts []string
		for key, value := range v {
			parts = append(parts, fmt.Sprintf("%s: %v", key, value))
		}
		return strings.Join(parts, "\n")
	case []interface{}:
		var parts []string
		for i, item := range v {
			parts = append(parts, fmt.Sprintf("[%d]: %v", i, item))
		}
		return strings.Join(parts, "\n")
	default:
		return fmt.Sprintf("%+v", outputs)
	}
}

// CollectBatchFeedback 批量收集反馈（非交互式）
func (fc *FeedbackCollector) CollectBatchFeedback(ctx context.Context, iterationID string, outputs interface{}, feedbackData map[string]interface{}) (*HumanFeedback, error) {
	feedback := &HumanFeedback{
		IterationID: iterationID,
		Timestamp:   time.Now(),
		Categories:  make(map[string]float64),
		Tags:        make([]string, 0),
		Issues:      make([]string, 0),
		Verified:    true,
		VerifiedBy:  "batch",
	}

	// 从提供的数据中提取反馈
	if score, ok := feedbackData["quality_score"].(float64); ok {
		feedback.QualityScore = score
	} else {
		feedback.QualityScore = 5.0 // 默认中性分数
	}

	if score, ok := feedbackData["accuracy_score"].(float64); ok {
		feedback.AccuracyScore = score
	} else {
		feedback.AccuracyScore = 5.0
	}

	if score, ok := feedbackData["usefulness"].(float64); ok {
		feedback.Usefulness = score
	} else {
		feedback.Usefulness = 5.0
	}

	if comments, ok := feedbackData["comments"].(string); ok {
		feedback.Comments = comments
	}

	if suggestions, ok := feedbackData["suggestions"].(string); ok {
		feedback.Suggestions = suggestions
	}

	if issues, ok := feedbackData["issues"].([]string); ok {
		feedback.Issues = issues
	}

	if tags, ok := feedbackData["tags"].([]string); ok {
		feedback.Tags = tags
	}

	if categories, ok := feedbackData["categories"].(map[string]float64); ok {
		feedback.Categories = categories
	}

	fc.logger.Info("batch feedback collected",
		logger.Field{Key: "iteration_id", Value: iterationID},
		logger.Field{Key: "quality_score", Value: feedback.QualityScore},
	)

	return feedback, nil
}

// ValidateFeedback 验证反馈数据
func (fc *FeedbackCollector) ValidateFeedback(feedback *HumanFeedback) error {
	if feedback.QualityScore < 1 || feedback.QualityScore > 10 {
		return fmt.Errorf("quality score must be between 1 and 10, got %.2f", feedback.QualityScore)
	}

	if feedback.AccuracyScore < 1 || feedback.AccuracyScore > 10 {
		return fmt.Errorf("accuracy score must be between 1 and 10, got %.2f", feedback.AccuracyScore)
	}

	if feedback.Usefulness < 1 || feedback.Usefulness > 10 {
		return fmt.Errorf("usefulness score must be between 1 and 10, got %.2f", feedback.Usefulness)
	}

	if feedback.IterationID == "" {
		return fmt.Errorf("iteration ID cannot be empty")
	}

	return nil
}

