package planning

import (
	"context"
)

// CrewPlanner 接口定义，对应Python版本的CrewPlanner类
// 负责协调和规划crew任务的执行
type CrewPlanner interface {
	// HandleCrewPlanning 处理crew规划，核心方法，对应Python版本的_handle_crew_planning()
	HandleCrewPlanning(ctx context.Context) (*PlannerTaskPydanticOutput, error)

	// CreateTasksSummary 创建任务摘要，对应Python版本的_create_tasks_summary()
	CreateTasksSummary(ctx context.Context) (string, error)

	// SetConfig 设置规划配置
	SetConfig(config *PlanningConfig) error

	// GetConfig 获取当前规划配置
	GetConfig() *PlanningConfig

	// ValidateConfiguration 验证规划配置
	ValidateConfiguration() error
}

// PlanningAgent 规划代理接口，负责创建和管理规划代理
type PlanningAgent interface {
	// CreatePlanningAgent 创建规划代理，对应Python版本的_create_planning_agent()
	CreatePlanningAgent(ctx context.Context, config *PlanningConfig) (Agent, error)

	// CreatePlannerTask 创建规划任务，对应Python版本的_create_planner_task()
	CreatePlannerTask(ctx context.Context, planningAgent Agent, tasksSummary string) (Task, error)

	// GetAgentKnowledge 获取代理知识，对应Python版本的_get_agent_knowledge()
	GetAgentKnowledge(ctx context.Context, taskInfo *TaskInfo) ([]string, error)
}

// TaskSummaryGenerator 任务摘要生成器接口
type TaskSummaryGenerator interface {
	// GenerateTaskSummary 为单个任务生成摘要
	GenerateTaskSummary(ctx context.Context, taskInfo *TaskInfo, index int) (*TaskSummary, error)

	// GenerateTasksSummary 为多个任务生成摘要字符串
	GenerateTasksSummary(ctx context.Context, tasks []TaskInfo) (string, error)

	// FormatTaskSummary 格式化任务摘要为字符串
	FormatTaskSummary(summary *TaskSummary) string
}

// PlanValidator 计划验证器接口
type PlanValidator interface {
	// ValidatePlan 验证单个计划
	ValidatePlan(ctx context.Context, plan *PlanPerTask) error

	// ValidatePlanOutput 验证完整的规划输出
	ValidatePlanOutput(ctx context.Context, output *PlannerTaskPydanticOutput) error

	// ValidateTaskInfo 验证任务信息
	ValidateTaskInfo(ctx context.Context, taskInfo *TaskInfo) error

	// ValidateTaskInfos 验证多个任务信息
	ValidateTaskInfos(ctx context.Context, tasks []TaskInfo) error
}

// PlanExecutor 计划执行器接口，负责执行规划逻辑
type PlanExecutor interface {
	// ExecutePlanning 执行规划逻辑
	ExecutePlanning(ctx context.Context, request *PlanningRequest) (*PlanningResult, error)

	// ExecutePlanningWithRetry 带重试的规划执行
	ExecutePlanningWithRetry(ctx context.Context, request *PlanningRequest) (*PlanningResult, error)

	// SetMaxRetries 设置最大重试次数
	SetMaxRetries(maxRetries int)

	// SetTimeout 设置超时时间
	SetTimeout(timeoutSeconds int)
}

// PlanStorage 计划存储接口，用于持久化规划结果
type PlanStorage interface {
	// SavePlan 保存规划结果
	SavePlan(ctx context.Context, planID string, result *PlanningResult) error

	// LoadPlan 加载规划结果
	LoadPlan(ctx context.Context, planID string) (*PlanningResult, error)

	// DeletePlan 删除规划结果
	DeletePlan(ctx context.Context, planID string) error

	// ListPlans 列出所有规划结果
	ListPlans(ctx context.Context) ([]string, error)
}

// PlanningMetrics 规划指标接口，用于收集规划性能指标
type PlanningMetrics interface {
	// RecordPlanningStart 记录规划开始
	RecordPlanningStart(ctx context.Context, taskCount int)

	// RecordPlanningEnd 记录规划结束
	RecordPlanningEnd(ctx context.Context, success bool, duration float64)

	// RecordPlanningError 记录规划错误
	RecordPlanningError(ctx context.Context, errorType string, retryCount int)

	// GetMetrics 获取规划指标
	GetMetrics(ctx context.Context) (*PlanningMetricsData, error)
}

// PlanningMetricsData 规划指标数据
type PlanningMetricsData struct {
	TotalPlannings      int64            `json:"total_plannings"`
	SuccessfulPlannings int64            `json:"successful_plannings"`
	FailedPlannings     int64            `json:"failed_plannings"`
	AverageDuration     float64          `json:"average_duration_ms"`
	AverageTaskCount    float64          `json:"average_task_count"`
	TotalRetries        int64            `json:"total_retries"`
	CommonErrors        map[string]int64 `json:"common_errors"`
}

// PlanningEventHandler 规划事件处理器接口
type PlanningEventHandler interface {
	// OnPlanningStarted 规划开始事件
	OnPlanningStarted(ctx context.Context, event *PlanningStartedEvent)

	// OnPlanningCompleted 规划完成事件
	OnPlanningCompleted(ctx context.Context, event *PlanningCompletedEvent)

	// OnPlanningFailed 规划失败事件
	OnPlanningFailed(ctx context.Context, event *PlanningFailedEvent)

	// OnTaskSummaryGenerated 任务摘要生成事件
	OnTaskSummaryGenerated(ctx context.Context, event *TaskSummaryGeneratedEvent)
}
