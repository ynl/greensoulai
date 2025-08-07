package training

import (
	"time"

	"github.com/ynl/greensoulai/pkg/events"
)

// 训练事件类型常量
const (
	TrainingStartedEventType            = "training_started"
	TrainingStoppedEventType            = "training_stopped"
	TrainingCompletedEventType          = "training_completed"
	TrainingIterationStartedEventType   = "training_iteration_started"
	TrainingIterationCompletedEventType = "training_iteration_completed"
	TrainingFeedbackCollectedEventType  = "training_feedback_collected"
	TrainingMetricsAnalyzedEventType    = "training_metrics_analyzed"
	TrainingErrorEventType              = "training_error"
)

// TrainingStartedEvent 训练开始事件
type TrainingStartedEvent struct {
	events.BaseEvent
	SessionID string          `json:"session_id"`
	Config    *TrainingConfig `json:"config"`
}

// NewTrainingStartedEvent 创建训练开始事件
func NewTrainingStartedEvent(sessionID string, config *TrainingConfig) *TrainingStartedEvent {
	return &TrainingStartedEvent{
		BaseEvent: events.BaseEvent{
			Type:      TrainingStartedEventType,
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"session_id":       sessionID,
				"iterations":       config.Iterations,
				"filename":         config.Filename,
				"collect_feedback": config.CollectFeedback,
			},
		},
		SessionID: sessionID,
		Config:    config,
	}
}

// TrainingStoppedEvent 训练停止事件
type TrainingStoppedEvent struct {
	events.BaseEvent
	SessionID string `json:"session_id"`
	Reason    string `json:"reason"`
}

// NewTrainingStoppedEvent 创建训练停止事件
func NewTrainingStoppedEvent(sessionID string, reason string) *TrainingStoppedEvent {
	return &TrainingStoppedEvent{
		BaseEvent: events.BaseEvent{
			Type:      TrainingStoppedEventType,
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"session_id": sessionID,
				"reason":     reason,
			},
		},
		SessionID: sessionID,
		Reason:    reason,
	}
}

// TrainingCompletedEvent 训练完成事件
type TrainingCompletedEvent struct {
	events.BaseEvent
	SessionID       string           `json:"session_id"`
	TotalIterations int              `json:"total_iterations"`
	SuccessfulRuns  int              `json:"successful_runs"`
	TotalDuration   time.Duration    `json:"total_duration"`
	Summary         *TrainingSummary `json:"summary"`
}

// NewTrainingCompletedEvent 创建训练完成事件
func NewTrainingCompletedEvent(sessionID string, totalIterations, successfulRuns int,
	totalDuration time.Duration, summary *TrainingSummary) *TrainingCompletedEvent {
	return &TrainingCompletedEvent{
		BaseEvent: events.BaseEvent{
			Type:      TrainingCompletedEventType,
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"session_id":        sessionID,
				"total_iterations":  totalIterations,
				"successful_runs":   successfulRuns,
				"total_duration_ms": totalDuration.Milliseconds(),
				"success_rate":      float64(successfulRuns) / float64(totalIterations),
			},
		},
		SessionID:       sessionID,
		TotalIterations: totalIterations,
		SuccessfulRuns:  successfulRuns,
		TotalDuration:   totalDuration,
		Summary:         summary,
	}
}

// TrainingIterationStartedEvent 训练迭代开始事件
type TrainingIterationStartedEvent struct {
	events.BaseEvent
	SessionID      string `json:"session_id"`
	IterationID    string `json:"iteration_id"`
	IterationIndex int    `json:"iteration_index"`
}

// NewTrainingIterationStartedEvent 创建迭代开始事件
func NewTrainingIterationStartedEvent(sessionID, iterationID string, index int) *TrainingIterationStartedEvent {
	return &TrainingIterationStartedEvent{
		BaseEvent: events.BaseEvent{
			Type:      TrainingIterationStartedEventType,
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"session_id":      sessionID,
				"iteration_id":    iterationID,
				"iteration_index": index,
			},
		},
		SessionID:      sessionID,
		IterationID:    iterationID,
		IterationIndex: index,
	}
}

// TrainingIterationCompletedEvent 训练迭代完成事件
type TrainingIterationCompletedEvent struct {
	events.BaseEvent
	SessionID      string        `json:"session_id"`
	IterationID    string        `json:"iteration_id"`
	IterationIndex int           `json:"iteration_index"`
	Duration       time.Duration `json:"duration"`
	Success        bool          `json:"success"`
}

// NewTrainingIterationCompletedEvent 创建迭代完成事件
func NewTrainingIterationCompletedEvent(sessionID, iterationID string, index int,
	duration time.Duration, success bool) *TrainingIterationCompletedEvent {
	return &TrainingIterationCompletedEvent{
		BaseEvent: events.BaseEvent{
			Type:      TrainingIterationCompletedEventType,
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"session_id":      sessionID,
				"iteration_id":    iterationID,
				"iteration_index": index,
				"duration_ms":     duration.Milliseconds(),
				"success":         success,
			},
		},
		SessionID:      sessionID,
		IterationID:    iterationID,
		IterationIndex: index,
		Duration:       duration,
		Success:        success,
	}
}

// TrainingFeedbackCollectedEvent 训练反馈收集事件
type TrainingFeedbackCollectedEvent struct {
	events.BaseEvent
	SessionID   string         `json:"session_id"`
	IterationID string         `json:"iteration_id"`
	Feedback    *HumanFeedback `json:"feedback"`
}

// NewTrainingFeedbackCollectedEvent 创建反馈收集事件
func NewTrainingFeedbackCollectedEvent(sessionID, iterationID string, feedback *HumanFeedback) *TrainingFeedbackCollectedEvent {
	return &TrainingFeedbackCollectedEvent{
		BaseEvent: events.BaseEvent{
			Type:      TrainingFeedbackCollectedEventType,
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"session_id":     sessionID,
				"iteration_id":   iterationID,
				"quality_score":  feedback.QualityScore,
				"accuracy_score": feedback.AccuracyScore,
				"usefulness":     feedback.Usefulness,
				"has_comments":   feedback.Comments != "",
			},
		},
		SessionID:   sessionID,
		IterationID: iterationID,
		Feedback:    feedback,
	}
}

// TrainingMetricsAnalyzedEvent 训练指标分析事件
type TrainingMetricsAnalyzedEvent struct {
	events.BaseEvent
	SessionID   string              `json:"session_id"`
	IterationID string              `json:"iteration_id"`
	Metrics     *PerformanceMetrics `json:"metrics"`
}

// NewTrainingMetricsAnalyzedEvent 创建指标分析事件
func NewTrainingMetricsAnalyzedEvent(sessionID, iterationID string, metrics *PerformanceMetrics) *TrainingMetricsAnalyzedEvent {
	return &TrainingMetricsAnalyzedEvent{
		BaseEvent: events.BaseEvent{
			Type:      TrainingMetricsAnalyzedEventType,
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"session_id":        sessionID,
				"iteration_id":      iterationID,
				"average_score":     metrics.AverageScore,
				"success_rate":      metrics.SuccessRate,
				"execution_time_ms": metrics.ExecutionTime.Milliseconds(),
				"tokens_used":       metrics.TokensUsed,
			},
		},
		SessionID:   sessionID,
		IterationID: iterationID,
		Metrics:     metrics,
	}
}

// TrainingErrorEvent 训练错误事件
type TrainingErrorEvent struct {
	events.BaseEvent
	SessionID   string `json:"session_id"`
	IterationID string `json:"iteration_id,omitempty"`
	Error       string `json:"error"`
	ErrorType   string `json:"error_type"`
}

// NewTrainingErrorEvent 创建训练错误事件
func NewTrainingErrorEvent(sessionID, iterationID, errorType, errorMsg string) *TrainingErrorEvent {
	return &TrainingErrorEvent{
		BaseEvent: events.BaseEvent{
			Type:      TrainingErrorEventType,
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"session_id":   sessionID,
				"iteration_id": iterationID,
				"error":        errorMsg,
				"error_type":   errorType,
			},
		},
		SessionID:   sessionID,
		IterationID: iterationID,
		Error:       errorMsg,
		ErrorType:   errorType,
	}
}

// 实现Event接口
func (e *TrainingStartedEvent) GetType() string {
	return e.Type
}

func (e *TrainingStartedEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

func (e *TrainingStartedEvent) GetSource() interface{} {
	return e.Source
}

func (e *TrainingStartedEvent) GetPayload() map[string]interface{} {
	return e.Payload
}

func (e *TrainingStoppedEvent) GetType() string {
	return e.Type
}

func (e *TrainingStoppedEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

func (e *TrainingStoppedEvent) GetSource() interface{} {
	return e.Source
}

func (e *TrainingStoppedEvent) GetPayload() map[string]interface{} {
	return e.Payload
}

func (e *TrainingCompletedEvent) GetType() string {
	return e.Type
}

func (e *TrainingCompletedEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

func (e *TrainingCompletedEvent) GetSource() interface{} {
	return e.Source
}

func (e *TrainingCompletedEvent) GetPayload() map[string]interface{} {
	return e.Payload
}

func (e *TrainingIterationStartedEvent) GetType() string {
	return e.Type
}

func (e *TrainingIterationStartedEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

func (e *TrainingIterationStartedEvent) GetSource() interface{} {
	return e.Source
}

func (e *TrainingIterationStartedEvent) GetPayload() map[string]interface{} {
	return e.Payload
}

func (e *TrainingIterationCompletedEvent) GetType() string {
	return e.Type
}

func (e *TrainingIterationCompletedEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

func (e *TrainingIterationCompletedEvent) GetSource() interface{} {
	return e.Source
}

func (e *TrainingIterationCompletedEvent) GetPayload() map[string]interface{} {
	return e.Payload
}

func (e *TrainingFeedbackCollectedEvent) GetType() string {
	return e.Type
}

func (e *TrainingFeedbackCollectedEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

func (e *TrainingFeedbackCollectedEvent) GetSource() interface{} {
	return e.Source
}

func (e *TrainingFeedbackCollectedEvent) GetPayload() map[string]interface{} {
	return e.Payload
}

func (e *TrainingMetricsAnalyzedEvent) GetType() string {
	return e.Type
}

func (e *TrainingMetricsAnalyzedEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

func (e *TrainingMetricsAnalyzedEvent) GetSource() interface{} {
	return e.Source
}

func (e *TrainingMetricsAnalyzedEvent) GetPayload() map[string]interface{} {
	return e.Payload
}

func (e *TrainingErrorEvent) GetType() string {
	return e.Type
}

func (e *TrainingErrorEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

func (e *TrainingErrorEvent) GetSource() interface{} {
	return e.Source
}

func (e *TrainingErrorEvent) GetPayload() map[string]interface{} {
	return e.Payload
}
