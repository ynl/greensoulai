package evaluation

import (
	"encoding/json"
	"fmt"
	"time"
)

// EvaluationScore 评估分数结构，对应Python版本的EvaluationScore
type EvaluationScore struct {
	Score    float64 `json:"score" validate:"min=0,max=10"` // 评分（0-10）
	Feedback string  `json:"feedback" validate:"required"`  // 反馈信息
	Category string  `json:"category,omitempty"`            // 评估类别
	Criteria string  `json:"criteria,omitempty"`            // 评估标准
}

// String 返回评估分数的字符串表示，与Python版本保持一致
func (e *EvaluationScore) String() string {
	return fmt.Sprintf("Score: %.1f/10 - %s", e.Score, e.Feedback)
}

// IsPass 判断是否通过评估（分数>=6.0视为通过）
func (e *EvaluationScore) IsPass() bool {
	return e.Score >= 6.0
}

// TaskEvaluationPydanticOutput 任务评估输出，对应Python版本的TaskEvaluationPydanticOutput
// 保持与Python版本的业务逻辑一致
type TaskEvaluationPydanticOutput struct {
	Quality float64 `json:"quality" validate:"min=0,max=10"` // 质量评分
}

// ToJSON 将TaskEvaluationPydanticOutput转换为JSON字符串
func (t *TaskEvaluationPydanticOutput) ToJSON() (string, error) {
	bytes, err := json.Marshal(t)
	if err != nil {
		return "", fmt.Errorf("failed to marshal TaskEvaluationPydanticOutput to JSON: %w", err)
	}
	return string(bytes), nil
}

// FromJSON 从JSON字符串解析TaskEvaluationPydanticOutput
func (t *TaskEvaluationPydanticOutput) FromJSON(jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), t)
}

// TaskEvaluation 任务评估结果，对应Python版本的TaskEvaluation
// 包含更详细的评估信息
type TaskEvaluation struct {
	Score            float64                `json:"score" validate:"min=0,max=10"`             // 整体评分
	CompletionScore  float64                `json:"completion_score" validate:"min=0,max=10"`  // 完成度评分
	QualityScore     float64                `json:"quality_score" validate:"min=0,max=10"`     // 质量评分
	PerformanceScore float64                `json:"performance_score" validate:"min=0,max=10"` // 性能评分
	Suggestions      []string               `json:"suggestions"`                               // 改进建议
	Feedback         string                 `json:"feedback" validate:"required"`              // 详细反馈
	Entities         []EntityExtraction     `json:"entities,omitempty"`                        // 提取的实体
	ExecutionTimeMs  float64                `json:"execution_time_ms"`                         // 执行时间（毫秒）
	Timestamp        time.Time              `json:"timestamp"`                                 // 评估时间戳
	EvaluatorVersion string                 `json:"evaluator_version,omitempty"`               // 评估器版本
	Metadata         map[string]interface{} `json:"metadata,omitempty"`                        // 元数据
}

// EntityExtraction 实体提取结果，对应Python版本的实体提取逻辑
type EntityExtraction struct {
	Name          string                 `json:"name" validate:"required"`          // 实体名称
	Type          string                 `json:"type" validate:"required"`          // 实体类型
	Description   string                 `json:"description"`                       // 实体描述
	Relationships []EntityRelationship   `json:"relationships,omitempty"`           // 实体关系
	Confidence    float64                `json:"confidence" validate:"min=0,max=1"` // 置信度
	Metadata      map[string]interface{} `json:"metadata,omitempty"`                // 实体元数据
}

// EntityRelationship 实体关系
type EntityRelationship struct {
	RelationType string  `json:"relation_type" validate:"required"` // 关系类型
	TargetEntity string  `json:"target_entity" validate:"required"` // 目标实体
	Confidence   float64 `json:"confidence" validate:"min=0,max=1"` // 置信度
}

// GetOverallScore 计算综合评分
func (t *TaskEvaluation) GetOverallScore() float64 {
	if t.Score > 0 {
		return t.Score
	}
	// 如果没有设置整体评分，则根据各项评分计算
	return (t.CompletionScore + t.QualityScore + t.PerformanceScore) / 3.0
}

// GetGrade 获取评估等级
func (t *TaskEvaluation) GetGrade() string {
	score := t.GetOverallScore()
	switch {
	case score >= 9.0:
		return "A+"
	case score >= 8.0:
		return "A"
	case score >= 7.0:
		return "B"
	case score >= 6.0:
		return "C"
	case score >= 5.0:
		return "D"
	default:
		return "F"
	}
}

// TrainingTaskEvaluation 训练任务评估结果，对应Python版本的TrainingTaskEvaluation
type TrainingTaskEvaluation struct {
	TaskID          string                 `json:"task_id" validate:"required"`    // 任务ID
	AgentRole       string                 `json:"agent_role" validate:"required"` // 代理角色
	TaskDescription string                 `json:"task_description"`               // 任务描述
	ExpectedOutput  string                 `json:"expected_output"`                // 期望输出
	ActualOutput    string                 `json:"actual_output"`                  // 实际输出
	Score           float64                `json:"score" validate:"min=0,max=10"`  // 评分
	Feedback        string                 `json:"feedback" validate:"required"`   // 反馈
	Improvements    []string               `json:"improvements"`                   // 改进建议
	ExecutionTimeMs float64                `json:"execution_time_ms"`              // 执行时间（毫秒）
	Iteration       int                    `json:"iteration"`                      // 训练迭代次数
	Timestamp       time.Time              `json:"timestamp"`                      // 评估时间戳
	ModelUsed       string                 `json:"model_used,omitempty"`           // 使用的模型
	TokensUsed      int                    `json:"tokens_used,omitempty"`          // 使用的token数量
	Metadata        map[string]interface{} `json:"metadata,omitempty"`             // 元数据
}

// AgentEvaluationResult Agent评估结果
type AgentEvaluationResult struct {
	AgentID   string                      `json:"agent_id" validate:"required"` // Agent ID
	TaskID    string                      `json:"task_id,omitempty"`            // 任务ID
	Metrics   map[string]*EvaluationScore `json:"metrics"`                      // 评估指标
	Timestamp time.Time                   `json:"timestamp"`                    // 评估时间戳
	Metadata  map[string]interface{}      `json:"metadata,omitempty"`           // 元数据
}

// GetAverageScore 获取平均评分
func (a *AgentEvaluationResult) GetAverageScore() float64 {
	if len(a.Metrics) == 0 {
		return 0.0
	}

	total := 0.0
	for _, score := range a.Metrics {
		total += score.Score
	}
	return total / float64(len(a.Metrics))
}

// CrewEvaluationResult Crew评估结果
type CrewEvaluationResult struct {
	CrewName       string                 `json:"crew_name" validate:"required"` // Crew名称
	Iteration      int                    `json:"iteration"`                     // 评估迭代次数
	TasksScores    map[int][]float64      `json:"tasks_scores"`                  // 各迭代任务评分
	ExecutionTimes map[int][]float64      `json:"execution_times"`               // 各迭代执行时间
	AverageScore   float64                `json:"average_score"`                 // 平均分数
	AverageTime    float64                `json:"average_time"`                  // 平均执行时间
	TotalTasks     int                    `json:"total_tasks"`                   // 总任务数
	PassedTasks    int                    `json:"passed_tasks"`                  // 通过的任务数
	SuccessRate    float64                `json:"success_rate"`                  // 成功率
	ModelUsed      string                 `json:"model_used,omitempty"`          // 使用的评估模型
	EvaluatedAt    time.Time              `json:"evaluated_at"`                  // 评估时间
	Metadata       map[string]interface{} `json:"metadata,omitempty"`            // 元数据
}

// CalculateStats 计算统计信息
func (c *CrewEvaluationResult) CalculateStats() {
	if len(c.TasksScores) == 0 {
		return
	}

	totalScore := 0.0
	totalTime := 0.0
	totalTasks := 0
	passedTasks := 0

	for _, scores := range c.TasksScores {
		for _, score := range scores {
			totalScore += score
			totalTasks++
			if score >= 6.0 { // 6.0分以上算通过
				passedTasks++
			}
		}
	}

	for _, times := range c.ExecutionTimes {
		for _, time := range times {
			totalTime += time
		}
	}

	if totalTasks > 0 {
		c.AverageScore = totalScore / float64(totalTasks)
		c.AverageTime = totalTime / float64(totalTasks)
		c.TotalTasks = totalTasks
		c.PassedTasks = passedTasks
		c.SuccessRate = float64(passedTasks) / float64(totalTasks) * 100.0
	}
}

// GetPerformanceGrade 获取性能等级
func (c *CrewEvaluationResult) GetPerformanceGrade() string {
	switch {
	case c.AverageScore >= 9.0:
		return "Excellent"
	case c.AverageScore >= 8.0:
		return "Very Good"
	case c.AverageScore >= 7.0:
		return "Good"
	case c.AverageScore >= 6.0:
		return "Satisfactory"
	case c.AverageScore >= 5.0:
		return "Needs Improvement"
	default:
		return "Poor"
	}
}

// MetricCategory 评估指标类别，对应Python版本的MetricCategory枚举
type MetricCategory string

const (
	MetricCategoryGoalAlignment   MetricCategory = "goal_alignment"   // 目标对齐度
	MetricCategorySemanticQuality MetricCategory = "semantic_quality" // 语义质量
	MetricCategoryTaskCompletion  MetricCategory = "task_completion"  // 任务完成度
	MetricCategoryEfficiency      MetricCategory = "efficiency"       // 效率
	MetricCategoryAccuracy        MetricCategory = "accuracy"         // 准确性
	MetricCategoryCreativity      MetricCategory = "creativity"       // 创造性
	MetricCategoryCoherence       MetricCategory = "coherence"        // 连贯性
	MetricCategoryRelevance       MetricCategory = "relevance"        // 相关性
)

// String 返回指标类别的字符串表示
func (m MetricCategory) String() string {
	return string(m)
}

// IsValid 检查指标类别是否有效
func (m MetricCategory) IsValid() bool {
	switch m {
	case MetricCategoryGoalAlignment, MetricCategorySemanticQuality, MetricCategoryTaskCompletion,
		MetricCategoryEfficiency, MetricCategoryAccuracy, MetricCategoryCreativity,
		MetricCategoryCoherence, MetricCategoryRelevance:
		return true
	default:
		return false
	}
}

// EvaluationConfig 评估配置
type EvaluationConfig struct {
	EvaluatorLLM   string                 `json:"evaluator_llm"`             // 评估使用的LLM模型
	EnableVerbose  bool                   `json:"enable_verbose"`            // 是否启用详细输出
	MaxRetries     int                    `json:"max_retries"`               // 最大重试次数
	TimeoutSeconds int                    `json:"timeout_seconds"`           // 超时时间（秒）
	PassingScore   float64                `json:"passing_score"`             // 及格分数
	Categories     []MetricCategory       `json:"categories"`                // 评估类别
	CustomCriteria map[string]string      `json:"custom_criteria,omitempty"` // 自定义评估标准
	Metadata       map[string]interface{} `json:"metadata,omitempty"`        // 配置元数据
}

// DefaultEvaluationConfig 返回默认的评估配置
func DefaultEvaluationConfig() *EvaluationConfig {
	return &EvaluationConfig{
		EvaluatorLLM:   "gpt-4o-mini", // 默认使用gpt-4o-mini进行评估
		EnableVerbose:  false,
		MaxRetries:     3,
		TimeoutSeconds: 60,
		PassingScore:   6.0, // 默认6分及格
		Categories: []MetricCategory{
			MetricCategoryGoalAlignment,
			MetricCategorySemanticQuality,
			MetricCategoryTaskCompletion,
		},
		CustomCriteria: make(map[string]string),
		Metadata:       make(map[string]interface{}),
	}
}

// Validate 验证评估配置
func (c *EvaluationConfig) Validate() error {
	if c.EvaluatorLLM == "" {
		return fmt.Errorf("evaluator LLM cannot be empty")
	}
	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}
	if c.TimeoutSeconds < 0 {
		return fmt.Errorf("timeout seconds cannot be negative")
	}
	if c.PassingScore < 0 || c.PassingScore > 10 {
		return fmt.Errorf("passing score must be between 0 and 10")
	}
	for _, category := range c.Categories {
		if !category.IsValid() {
			return fmt.Errorf("invalid metric category: %s", category)
		}
	}
	return nil
}

