package crew

import (
	"context"
	"time"

	"github.com/ynl/greensoulai/internal/agent"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// Process 定义Crew的执行模式
type Process int

const (
	ProcessSequential Process = iota
	ProcessHierarchical
	// TODO: ProcessConsensual
)

func (p Process) String() string {
	switch p {
	case ProcessSequential:
		return "sequential"
	case ProcessHierarchical:
		return "hierarchical"
	default:
		return "unknown"
	}
}

// Crew 定义团队协作接口
type Crew interface {
	// 核心执行方法
	Kickoff(ctx context.Context, inputs map[string]interface{}) (*CrewOutput, error)
	KickoffAsync(ctx context.Context, inputs map[string]interface{}) (<-chan CrewResult, error)
	KickoffForEach(ctx context.Context, inputsList []map[string]interface{}) ([]*CrewOutput, error)
	KickoffForEachAsync(ctx context.Context, inputsList []map[string]interface{}) (<-chan []*CrewOutput, error)
	KickoffWithTimeout(ctx context.Context, inputs map[string]interface{}, timeout time.Duration) (*CrewOutput, error)

	// 训练方法
	Train(ctx context.Context, nIterations int, filename string, inputs map[string]interface{}) error

	// 配置方法
	AddAgent(agent agent.Agent) error
	AddTask(task agent.Task) error
	SetProcess(process Process)
	SetVerbose(verbose bool)
	SetMemoryEnabled(enabled bool)
	SetCacheEnabled(enabled bool)

	// 回调管理
	AddBeforeKickoffCallback(callback KickoffCallback) error
	AddAfterKickoffCallback(callback KickoffCallback) error
	AddTaskCallback(callback TaskCallback) error
	AddStepCallback(callback StepCallback) error

	// 状态查询
	GetAgents() []agent.Agent
	GetTasks() []agent.Task
	GetProcess() Process
	IsMemoryEnabled() bool
	IsCacheEnabled() bool
	GetUsageMetrics() *UsageMetrics

	// 生命周期管理
	Clone() (Crew, error)
	Copy() (Crew, error)
	Close() error
}

// CrewOutput 定义Crew执行的输出结果
type CrewOutput struct {
	Raw         string                 `json:"raw"`
	JSON        map[string]interface{} `json:"json,omitempty"`
	Pydantic    interface{}            `json:"pydantic,omitempty"`
	TasksOutput []*agent.TaskOutput    `json:"tasks_output"`
	TokenUsage  *UsageMetrics          `json:"token_usage"`
	CreatedAt   time.Time              `json:"created_at"`
	Duration    time.Duration          `json:"duration"`
	Success     bool                   `json:"success"`
	Error       error                  `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// CrewResult 定义异步执行的结果
type CrewResult struct {
	Output *CrewOutput
	Error  error
}

// UsageMetrics 定义使用统计
type UsageMetrics struct {
	TotalTokens      int           `json:"total_tokens"`
	PromptTokens     int           `json:"prompt_tokens"`
	CompletionTokens int           `json:"completion_tokens"`
	TotalCost        float64       `json:"total_cost"`
	SuccessfulTasks  int           `json:"successful_tasks"`
	FailedTasks      int           `json:"failed_tasks"`
	TotalTasks       int           `json:"total_tasks"`
	ExecutionTime    time.Duration `json:"execution_time"`
}

// AddUsageMetrics 累加使用统计
func (u *UsageMetrics) AddUsageMetrics(other *UsageMetrics) {
	if other == nil {
		return
	}
	u.TotalTokens += other.TotalTokens
	u.PromptTokens += other.PromptTokens
	u.CompletionTokens += other.CompletionTokens
	u.TotalCost += other.TotalCost
	u.SuccessfulTasks += other.SuccessfulTasks
	u.FailedTasks += other.FailedTasks
	u.TotalTasks += other.TotalTasks
	u.ExecutionTime += other.ExecutionTime
}

// 回调函数类型定义
type KickoffCallback func(ctx context.Context, crew Crew, output *CrewOutput) (*CrewOutput, error)
type TaskCallback func(ctx context.Context, task agent.Task, output *agent.TaskOutput) error
type StepCallback func(ctx context.Context, agent agent.Agent, step *StepInfo) error

// StepInfo 定义步骤信息
type StepInfo struct {
	Agent       string                 `json:"agent"`
	Task        string                 `json:"task"`
	Step        int                    `json:"step"`
	Action      string                 `json:"action"`
	Observation string                 `json:"observation"`
	Thought     string                 `json:"thought"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
}

// CrewConfig 定义Crew配置
type CrewConfig struct {
	Name                   string                 `json:"name"`
	Process                Process                `json:"process"`
	Verbose                bool                   `json:"verbose"`
	MemoryEnabled          bool                   `json:"memory_enabled"`
	CacheEnabled           bool                   `json:"cache_enabled"`
	MaxRPM                 int                    `json:"max_rpm"`
	ShareCrew              bool                   `json:"share_crew"`
	PlanningEnabled        bool                   `json:"planning_enabled"`
	MaxExecutionTime       time.Duration          `json:"max_execution_time"`
	FullOutput             bool                   `json:"full_output"`
	StepCallback           StepCallback           `json:"-"`
	TaskCallback           TaskCallback           `json:"-"`
	BeforeKickoffCallbacks []KickoffCallback      `json:"-"`
	AfterKickoffCallbacks  []KickoffCallback      `json:"-"`
	ManagerAgent           agent.Agent            `json:"-"`
	ManagerLLM             interface{}            `json:"-"`
	FunctionCallingLLM     interface{}            `json:"-"`
	ChatLLM                interface{}            `json:"-"`
	PromptFile             string                 `json:"prompt_file"`
	OutputLogFile          string                 `json:"output_log_file"`
	Metadata               map[string]interface{} `json:"metadata"`
}

// DefaultCrewConfig 返回默认配置
func DefaultCrewConfig() *CrewConfig {
	return &CrewConfig{
		Name:                   "crew",
		Process:                ProcessSequential,
		Verbose:                false,
		MemoryEnabled:          false,
		CacheEnabled:           true,
		MaxRPM:                 60,
		ShareCrew:              false,
		PlanningEnabled:        false,
		MaxExecutionTime:       30 * time.Minute,
		FullOutput:             false,
		BeforeKickoffCallbacks: make([]KickoffCallback, 0),
		AfterKickoffCallbacks:  make([]KickoffCallback, 0),
		Metadata:               make(map[string]interface{}),
	}
}

// PlanningConfig 定义规划配置
type PlanningConfig struct {
	Enabled       bool        `json:"enabled"`
	PlannerLLM    interface{} `json:"-"`
	PlanningAgent agent.Agent `json:"-"`
}

// Memory 接口定义（占位符，后续实现）
type Memory interface {
	Store(ctx context.Context, key string, value interface{}) error
	Retrieve(ctx context.Context, key string) (interface{}, error)
	Search(ctx context.Context, query string, limit int) ([]interface{}, error)
	Clear(ctx context.Context) error
}

// Cache 接口定义（占位符，后续实现）
type Cache interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string) (interface{}, error)
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
}

// CrewBuilder 定义Crew构建器模式
type CrewBuilder interface {
	WithName(name string) CrewBuilder
	WithProcess(process Process) CrewBuilder
	WithVerbose(verbose bool) CrewBuilder
	WithMemory(enabled bool) CrewBuilder
	WithCache(enabled bool) CrewBuilder
	WithMaxRPM(rpm int) CrewBuilder
	WithAgents(agents ...agent.Agent) CrewBuilder
	WithTasks(tasks ...agent.Task) CrewBuilder
	WithEventBus(eventBus events.EventBus) CrewBuilder
	WithLogger(logger logger.Logger) CrewBuilder
	WithConfig(config *CrewConfig) CrewBuilder
	Build() (Crew, error)
}
