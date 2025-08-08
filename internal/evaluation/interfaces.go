package evaluation

import (
	"context"
	"time"

	"github.com/ynl/greensoulai/internal/agent"
	"github.com/ynl/greensoulai/internal/llm"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// CrewEvaluator 接口定义，对应Python版本的CrewEvaluator类
// 负责评估crew的整体性能和任务执行质量
type CrewEvaluator interface {
	// SetupForEvaluating 设置crew进行评估，对应Python版本的_setup_for_evaluating()
	SetupForEvaluating(ctx context.Context, crew Crew) error

	// Evaluate 评估任务输出，对应Python版本的evaluate()方法
	Evaluate(ctx context.Context, taskOutput *TaskOutput) (*TaskEvaluationPydanticOutput, error)

	// SetIteration 设置评估迭代次数
	SetIteration(iteration int)

	// GetIteration 获取当前评估迭代次数
	GetIteration() int

	// PrintCrewEvaluationResult 打印crew评估结果
	PrintCrewEvaluationResult(ctx context.Context) error

	// GetEvaluationResult 获取评估结果
	GetEvaluationResult() *CrewEvaluationResult

	// GetTasksScores 获取任务评分
	GetTasksScores() map[int][]float64

	// GetExecutionTimes 获取执行时间
	GetExecutionTimes() map[int][]float64

	// Reset 重置评估状态
	Reset() error

	// SetConfig 设置评估配置
	SetConfig(config *EvaluationConfig)

	// GetConfig 获取评估配置
	GetConfig() *EvaluationConfig
}

// TaskEvaluator 接口定义，对应Python版本的TaskEvaluator类
// 负责评估单个任务的执行质量和性能
type TaskEvaluator interface {
	// Evaluate 评估任务执行结果，对应Python版本的evaluate()方法
	Evaluate(ctx context.Context, task Task, output string) (*TaskEvaluation, error)

	// EvaluateTrainingData 评估训练数据，对应Python版本的evaluate_training_data()方法
	EvaluateTrainingData(ctx context.Context, trainingData map[string]interface{}, agentID string) (*TrainingTaskEvaluation, error)

	// SetOriginalAgent 设置原始agent
	SetOriginalAgent(agent agent.Agent)

	// GetOriginalAgent 获取原始agent
	GetOriginalAgent() agent.Agent

	// SetLLM 设置评估用的LLM
	SetLLM(llm llm.LLM) error

	// GetLLM 获取评估用的LLM
	GetLLM() llm.LLM

	// SetConfig 设置评估配置
	SetConfig(config *EvaluationConfig)

	// GetConfig 获取评估配置
	GetConfig() *EvaluationConfig
}

// BaseEvaluator 基础评估器接口，对应Python版本的BaseEvaluator抽象类
// 定义评估器的通用接口
type BaseEvaluator interface {
	// GetMetricCategory 获取评估指标类别
	GetMetricCategory() MetricCategory

	// Evaluate 评估agent执行情况
	Evaluate(ctx context.Context, agent agent.Agent, executionTrace map[string]interface{}, finalOutput interface{}, task Task) (*EvaluationScore, error)

	// SetLLM 设置LLM
	SetLLM(llm llm.LLM) error

	// GetLLM 获取LLM
	GetLLM() llm.LLM
}

// AgentEvaluator Agent评估器接口，对应Python版本的AgentEvaluator类
// 负责评估agent的性能和行为
type AgentEvaluator interface {
	// AddEvaluator 添加评估器
	AddEvaluator(evaluator BaseEvaluator) error

	// RemoveEvaluator 移除评估器
	RemoveEvaluator(category MetricCategory) error

	// GetEvaluators 获取所有评估器
	GetEvaluators() []BaseEvaluator

	// Evaluate 评估agent
	Evaluate(ctx context.Context, agent agent.Agent, executionTrace map[string]interface{}, finalOutput interface{}, task Task) (*AgentEvaluationResult, error)

	// EvaluateAsync 异步评估agent
	EvaluateAsync(ctx context.Context, agent agent.Agent, executionTrace map[string]interface{}, finalOutput interface{}, task Task) (<-chan *AgentEvaluationResult, error)

	// DisplayEvaluationWithFeedback 显示评估结果和反馈
	DisplayEvaluationWithFeedback(ctx context.Context) error

	// GetIterationsResults 获取迭代结果
	GetIterationsResults() []*AgentEvaluationResult

	// Reset 重置评估状态
	Reset() error

	// SetConfig 设置配置
	SetConfig(config *EvaluationConfig)

	// GetConfig 获取配置
	GetConfig() *EvaluationConfig
}

// EvaluationSession 评估会话接口
// 管理整个评估过程的生命周期
type EvaluationSession interface {
	// StartSession 开始评估会话
	StartSession(ctx context.Context, sessionID string) error

	// EndSession 结束评估会话
	EndSession(ctx context.Context) error

	// GetSessionID 获取会话ID
	GetSessionID() string

	// GetStartTime 获取开始时间
	GetStartTime() time.Time

	// GetEndTime 获取结束时间
	GetEndTime() time.Time

	// GetDuration 获取持续时间
	GetDuration() time.Duration

	// AddEvaluationResult 添加评估结果
	AddEvaluationResult(result interface{}) error

	// GetEvaluationResults 获取所有评估结果
	GetEvaluationResults() []interface{}

	// GenerateReport 生成评估报告
	GenerateReport(ctx context.Context) (string, error)

	// SaveReport 保存评估报告
	SaveReport(ctx context.Context, filename string) error

	// LoadReport 加载评估报告
	LoadReport(ctx context.Context, filename string) error
}

// 依赖接口定义，用于解耦外部依赖

// Crew 接口定义（简化版本，用于评估）
type Crew interface {
	GetName() string
	GetTasks() []Task
	GetAgents() []agent.Agent
	Execute(ctx context.Context, inputs map[string]interface{}) (*CrewOutput, error)
	SetTaskCallback(callback func(*TaskOutput)) error
}

// Task 接口定义（简化版本，用于评估）
type Task interface {
	GetID() string
	GetDescription() string
	GetExpectedOutput() string
	GetAgent() agent.Agent
	GetExecutionDuration() time.Duration
	Execute(ctx context.Context) (*TaskOutput, error)
	ExecuteSync(ctx context.Context) (*TaskOutput, error)
}

// TaskOutput 任务输出（简化版本，用于评估）
type TaskOutput struct {
	TaskID      string                 `json:"task_id"`
	Description string                 `json:"description"`
	Raw         string                 `json:"raw"`                 // 原始输出
	Agent       string                 `json:"agent"`               // 执行的agent
	Summary     string                 `json:"summary"`             // 输出摘要
	JSONDict    map[string]interface{} `json:"json_dict,omitempty"` // JSON字典输出
	Pydantic    interface{}            `json:"pydantic,omitempty"`  // Pydantic模型输出
	Metadata    map[string]interface{} `json:"metadata,omitempty"`  // 元数据
}

// CrewOutput crew执行输出（简化版本，用于评估）
type CrewOutput struct {
	Raw         string                 `json:"raw"`
	JSONDict    map[string]interface{} `json:"json_dict,omitempty"`
	Pydantic    interface{}            `json:"pydantic,omitempty"`
	TasksOutput []*TaskOutput          `json:"tasks_output"`
	TokenUsage  map[string]int         `json:"token_usage,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// EvaluationEventHandler 评估事件处理器接口
type EvaluationEventHandler interface {
	// HandleEvaluationStarted 处理评估开始事件
	HandleEvaluationStarted(ctx context.Context, event *EvaluationStartedEvent) error

	// HandleEvaluationCompleted 处理评估完成事件
	HandleEvaluationCompleted(ctx context.Context, event *EvaluationCompletedEvent) error

	// HandleEvaluationFailed 处理评估失败事件
	HandleEvaluationFailed(ctx context.Context, event *EvaluationFailedEvent) error

	// HandleTaskEvaluated 处理任务评估事件
	HandleTaskEvaluated(ctx context.Context, event *TaskEvaluatedEvent) error

	// HandleAgentEvaluated 处理agent评估事件
	HandleAgentEvaluated(ctx context.Context, event *AgentEvaluatedEvent) error
}

// EvaluatorFactory 评估器工厂接口
type EvaluatorFactory interface {
	// CreateCrewEvaluator 创建crew评估器
	CreateCrewEvaluator(ctx context.Context, config *EvaluationConfig) (CrewEvaluator, error)

	// CreateTaskEvaluator 创建任务评估器
	CreateTaskEvaluator(ctx context.Context, originalAgent agent.Agent, config *EvaluationConfig) (TaskEvaluator, error)

	// CreateAgentEvaluator 创建agent评估器
	CreateAgentEvaluator(ctx context.Context, agents []agent.Agent, evaluators []BaseEvaluator, config *EvaluationConfig) (AgentEvaluator, error)

	// CreateBaseEvaluator 创建基础评估器
	CreateBaseEvaluator(ctx context.Context, category MetricCategory, llmModel llm.LLM) (BaseEvaluator, error)

	// SetEventBus 设置事件总线
	SetEventBus(eventBus events.EventBus)

	// SetLogger 设置日志器
	SetLogger(logger logger.Logger)
}

