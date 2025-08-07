package training

import (
	"context"
	"time"
)

// TrainingHandler 定义训练处理器接口
type TrainingHandler interface {
	// StartTraining 开始训练
	StartTraining(ctx context.Context, config *TrainingConfig) error

	// CollectFeedback 收集人工反馈
	CollectFeedback(ctx context.Context, iterationID string, feedback *HumanFeedback) error

	// AnalyzePerformance 分析性能
	AnalyzePerformance(ctx context.Context, iterationID string) (*PerformanceMetrics, error)

	// SaveTrainingData 保存训练数据
	SaveTrainingData(ctx context.Context, data *TrainingData) error

	// LoadTrainingData 加载训练数据
	LoadTrainingData(ctx context.Context, filename string) (*TrainingData, error)

	// GetTrainingStatus 获取训练状态
	GetTrainingStatus(ctx context.Context) *TrainingStatus

	// StopTraining 停止训练
	StopTraining(ctx context.Context) error
}

// TrainingConfig 训练配置
type TrainingConfig struct {
	// 基础配置
	Iterations int                    `json:"iterations"`
	Filename   string                 `json:"filename"`
	Inputs     map[string]interface{} `json:"inputs"`

	// 训练参数
	LearningRate    float64 `json:"learning_rate"`
	BatchSize       int     `json:"batch_size"`
	ValidationSplit float64 `json:"validation_split"`

	// 反馈配置
	CollectFeedback bool          `json:"collect_feedback"`
	FeedbackTimeout time.Duration `json:"feedback_timeout"`
	FeedbackPrompts []string      `json:"feedback_prompts"`

	// 性能监控
	MetricsEnabled  bool     `json:"metrics_enabled"`
	MetricsInterval int      `json:"metrics_interval"` // 每N次迭代收集一次指标
	TargetMetrics   []string `json:"target_metrics"`

	// 早停条件
	EarlyStopping   bool    `json:"early_stopping"`
	PatientceEpochs int     `json:"patience_epochs"`
	MinImprovement  float64 `json:"min_improvement"`

	// 其他配置
	SaveInterval int  `json:"save_interval"` // 每N次迭代保存一次
	Verbose      bool `json:"verbose"`
	AutoSave     bool `json:"auto_save"`
	BackupCount  int  `json:"backup_count"`
}

// DefaultTrainingConfig 返回默认训练配置
func DefaultTrainingConfig() *TrainingConfig {
	return &TrainingConfig{
		Iterations:      10,
		Filename:        "training_data.json",
		LearningRate:    0.001,
		BatchSize:       1,
		ValidationSplit: 0.2,
		CollectFeedback: true,
		FeedbackTimeout: 5 * time.Minute,
		FeedbackPrompts: []string{"Please rate the quality of this output (1-10):", "Any suggestions for improvement?"},
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

// TrainingData 训练数据结构
type TrainingData struct {
	// 元数据
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	Version   string          `json:"version"`
	Config    *TrainingConfig `json:"config"`

	// 训练会话信息
	SessionID string `json:"session_id"`
	CrewName  string `json:"crew_name"`
	TotalRuns int    `json:"total_runs"`

	// 迭代数据
	Iterations []*IterationData `json:"iterations"`

	// 汇总信息
	Summary *TrainingSummary `json:"summary"`
}

// IterationData 单次迭代数据
type IterationData struct {
	// 基础信息
	IterationID string        `json:"iteration_id"`
	Index       int           `json:"index"`
	Timestamp   time.Time     `json:"timestamp"`
	Duration    time.Duration `json:"duration"`

	// 输入输出
	Inputs  map[string]interface{} `json:"inputs"`
	Outputs interface{}            `json:"outputs"`
	Success bool                   `json:"success"`
	Error   string                 `json:"error,omitempty"`

	// 反馈数据
	Feedback *HumanFeedback `json:"feedback,omitempty"`

	// 性能指标
	Metrics *PerformanceMetrics `json:"metrics,omitempty"`

	// Agent执行数据
	AgentData []*AgentIterationData `json:"agent_data"`

	// 任务执行数据
	TaskData []*TaskIterationData `json:"task_data"`
}

// AgentIterationData Agent在单次迭代中的数据
type AgentIterationData struct {
	AgentRole     string        `json:"agent_role"`
	ExecutionTime time.Duration `json:"execution_time"`
	TokensUsed    int           `json:"tokens_used"`
	ToolsUsed     []string      `json:"tools_used"`
	Success       bool          `json:"success"`
	Error         string        `json:"error,omitempty"`
}

// TaskIterationData 任务在单次迭代中的数据
type TaskIterationData struct {
	TaskDescription string        `json:"task_description"`
	ExecutionTime   time.Duration `json:"execution_time"`
	Success         bool          `json:"success"`
	OutputLength    int           `json:"output_length"`
	ValidationScore float64       `json:"validation_score"`
	Error           string        `json:"error,omitempty"`
}

// HumanFeedback 人工反馈数据
type HumanFeedback struct {
	// 基础信息
	IterationID string    `json:"iteration_id"`
	Timestamp   time.Time `json:"timestamp"`

	// 评分反馈
	QualityScore  float64 `json:"quality_score"`  // 1-10分
	AccuracyScore float64 `json:"accuracy_score"` // 1-10分
	Usefulness    float64 `json:"usefulness"`     // 1-10分

	// 文本反馈
	Comments    string   `json:"comments"`
	Suggestions string   `json:"suggestions"`
	Issues      []string `json:"issues"`

	// 分类反馈
	Categories map[string]float64 `json:"categories"` // 自定义分类评分
	Tags       []string           `json:"tags"`       // 标签

	// 验证信息
	Verified   bool   `json:"verified"`
	VerifiedBy string `json:"verified_by"`
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	// 基础指标
	ExecutionTime time.Duration `json:"execution_time"`
	SuccessRate   float64       `json:"success_rate"`
	ErrorRate     float64       `json:"error_rate"`

	// 资源使用
	MemoryUsage int64   `json:"memory_usage"`
	CPUUsage    float64 `json:"cpu_usage"`
	TokensUsed  int     `json:"tokens_used"`

	// 质量指标
	AverageScore    float64 `json:"average_score"`
	FeedbackScore   float64 `json:"feedback_score"`
	ValidationScore float64 `json:"validation_score"`

	// Agent指标
	AgentPerformance map[string]*AgentMetrics `json:"agent_performance"`

	// 任务指标
	TaskPerformance map[string]*TaskMetrics `json:"task_performance"`

	// 趋势指标
	ImprovementRate  float64 `json:"improvement_rate"`
	ConsistencyScore float64 `json:"consistency_score"`
}

// AgentMetrics Agent性能指标
type AgentMetrics struct {
	ExecutionCount     int           `json:"execution_count"`
	AverageTime        time.Duration `json:"average_time"`
	SuccessRate        float64       `json:"success_rate"`
	TokensPerExecution int           `json:"tokens_per_execution"`
	ToolUsageCount     int           `json:"tool_usage_count"`
}

// TaskMetrics 任务性能指标
type TaskMetrics struct {
	ExecutionCount      int           `json:"execution_count"`
	AverageTime         time.Duration `json:"average_time"`
	SuccessRate         float64       `json:"success_rate"`
	AverageOutputLength int           `json:"average_output_length"`
	ValidationScore     float64       `json:"validation_score"`
}

// TrainingSummary 训练总结
type TrainingSummary struct {
	// 总体统计
	TotalIterations int           `json:"total_iterations"`
	SuccessfulRuns  int           `json:"successful_runs"`
	FailedRuns      int           `json:"failed_runs"`
	TotalDuration   time.Duration `json:"total_duration"`
	AverageDuration time.Duration `json:"average_duration"`

	// 性能趋势
	InitialScore    float64 `json:"initial_score"`
	FinalScore      float64 `json:"final_score"`
	ImprovementRate float64 `json:"improvement_rate"`
	BestScore       float64 `json:"best_score"`
	WorstScore      float64 `json:"worst_score"`

	// 反馈统计
	TotalFeedback   int     `json:"total_feedback"`
	AverageFeedback float64 `json:"average_feedback"`

	// 资源使用
	TotalTokens   int `json:"total_tokens"`
	AverageTokens int `json:"average_tokens"`

	// 建议
	Recommendations []string `json:"recommendations"`
}

// TrainingStatus 训练状态
type TrainingStatus struct {
	// 状态信息
	IsRunning        bool    `json:"is_running"`
	CurrentIteration int     `json:"current_iteration"`
	TotalIterations  int     `json:"total_iterations"`
	Progress         float64 `json:"progress"`

	// 时间信息
	StartTime          time.Time     `json:"start_time"`
	ElapsedTime        time.Duration `json:"elapsed_time"`
	EstimatedRemaining time.Duration `json:"estimated_remaining"`

	// 当前性能
	CurrentScore      float64 `json:"current_score"`
	BestScore         float64 `json:"best_score"`
	RecentImprovement float64 `json:"recent_improvement"`

	// 状态消息
	Status    string `json:"status"`
	LastError string `json:"last_error,omitempty"`
}
