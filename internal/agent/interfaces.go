package agent

import (
	"context"
	"time"

	"github.com/ynl/greensoulai/internal/llm"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
	"github.com/ynl/greensoulai/pkg/security"
)

// Agent 代表一个智能体的核心接口
type Agent interface {
	// 核心执行方法
	Execute(ctx context.Context, task Task) (*TaskOutput, error)
	ExecuteAsync(ctx context.Context, task Task) (<-chan TaskResult, error)
	ExecuteWithTimeout(ctx context.Context, task Task, timeout time.Duration) (*TaskOutput, error)

	// 基础属性获取
	GetID() string
	GetRole() string
	GetGoal() string
	GetBackstory() string

	// 配置和工具管理
	AddTool(tool Tool) error
	GetTools() []Tool
	SetLLM(llm llm.LLM) error
	GetLLM() llm.LLM

	// 记忆和知识管理
	SetMemory(memory Memory) error
	GetMemory() Memory
	SetKnowledgeSources(sources []KnowledgeSource) error
	GetKnowledgeSources() []KnowledgeSource

	// 执行配置
	SetExecutionConfig(config ExecutionConfig) error
	GetExecutionConfig() ExecutionConfig
	SetHumanInputHandler(handler HumanInputHandler) error
	GetHumanInputHandler() HumanInputHandler

	// 事件和监控
	SetEventBus(eventBus events.EventBus) error
	GetEventBus() events.EventBus
	SetLogger(logger logger.Logger) error
	GetLogger() logger.Logger

	// 生命周期管理
	Initialize() error
	Close() error
	Clone() Agent

	// 统计和监控
	GetExecutionStats() ExecutionStats
	ResetStats() error
}

// Task 代表一个任务的接口
type Task interface {
	GetID() string
	GetDescription() string
	GetExpectedOutput() string
	GetContext() map[string]interface{}
	IsHumanInputRequired() bool
	SetHumanInput(input string)
	GetHumanInput() string
	GetOutputFormat() OutputFormat
	GetTools() []Tool
	Validate() error
}

// Tool 代表工具的接口
type Tool interface {
	GetName() string
	GetDescription() string
	GetSchema() ToolSchema
	Execute(ctx context.Context, args map[string]interface{}) (interface{}, error)
	ExecuteAsync(ctx context.Context, args map[string]interface{}) (<-chan ToolResult, error)
	GetUsageCount() int
	GetUsageLimit() int
	ResetUsage()
	IsUsageLimitExceeded() bool
}

// Memory 代表记忆系统的接口
type Memory interface {
	Store(ctx context.Context, key string, value interface{}) error
	Retrieve(ctx context.Context, key string) (interface{}, error)
	Search(ctx context.Context, query string, limit int) ([]MemoryItem, error)
	Clear(ctx context.Context) error
	GetStats() MemoryStats
}

// KnowledgeSource 代表知识源的接口
type KnowledgeSource interface {
	GetName() string
	GetDescription() string
	Query(ctx context.Context, query string, options QueryOptions) ([]KnowledgeItem, error)
	Initialize() error
	Close() error
	GetStats() KnowledgeStats
}

// HumanInputHandler 代表人工输入处理器的接口
type HumanInputHandler interface {
	RequestInput(ctx context.Context, prompt string, options []string) (string, error)
	IsInteractive() bool
	SetTimeout(timeout time.Duration)
	GetTimeout() time.Duration
}

// OutputFormat 定义任务输出格式
type OutputFormat int

const (
	OutputFormatRAW OutputFormat = iota
	OutputFormatJSON
	OutputFormatPydantic
)

// ExecutionConfig 定义Agent执行配置
type ExecutionConfig struct {
	MaxIterations    int           `json:"max_iterations"`
	MaxRPM           int           `json:"max_rpm"`
	Timeout          time.Duration `json:"timeout"`
	MaxExecutionTime time.Duration `json:"max_execution_time"`
	AllowDelegation  bool          `json:"allow_delegation"`
	VerboseLogging   bool          `json:"verbose_logging"`
	HumanInput       bool          `json:"human_input"`
	UseSystemPrompt  bool          `json:"use_system_prompt"`
	MaxTokens        int           `json:"max_tokens"`
	Temperature      float64       `json:"temperature"`
	CacheEnabled     bool          `json:"cache_enabled"`
	MaxRetryLimit    int           `json:"max_retry_limit"`
}

// TaskOutput 代表任务执行的输出
type TaskOutput struct {
	Raw             string                 `json:"raw"`
	JSON            map[string]interface{} `json:"json,omitempty"`
	Pydantic        interface{}            `json:"pydantic,omitempty"`
	Agent           string                 `json:"agent"`
	Task            string                 `json:"task"`
	Description     string                 `json:"description"`
	Summary         string                 `json:"summary"`
	ExpectedOutput  string                 `json:"expected_output"`
	OutputFormat    OutputFormat           `json:"output_format"`
	ExecutionTime   time.Duration          `json:"execution_time"`
	CreatedAt       time.Time              `json:"created_at"`
	TokensUsed      int                    `json:"tokens_used"`
	Cost            float64                `json:"cost"`
	Model           string                 `json:"model"`
	IsValid         bool                   `json:"is_valid"`
	ValidationError string                 `json:"validation_error,omitempty"`
	ToolsUsed       []string               `json:"tools_used"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// TaskResult 代表异步任务执行结果
type TaskResult struct {
	Output *TaskOutput
	Error  error
}

// ToolSchema 定义工具的模式
type ToolSchema struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Required    []string               `json:"required"`
}

// ToolResult 代表工具执行结果
type ToolResult struct {
	Output   interface{}            `json:"output"`
	Error    error                  `json:"error,omitempty"`
	Duration time.Duration          `json:"duration"`
	Metadata map[string]interface{} `json:"metadata"`
}

// MemoryItem 代表记忆中的一个项目
type MemoryItem struct {
	Key       string                 `json:"key"`
	Value     interface{}            `json:"value"`
	Timestamp time.Time              `json:"timestamp"`
	Score     float64                `json:"score"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// KnowledgeItem 代表知识源中的一个项目
type KnowledgeItem struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Source    string                 `json:"source"`
	Score     float64                `json:"score"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}

// QueryOptions 定义知识查询选项
type QueryOptions struct {
	Limit     int                    `json:"limit"`
	Threshold float64                `json:"threshold"`
	Metadata  map[string]interface{} `json:"metadata"`
	Filters   []QueryFilter          `json:"filters"`
}

// QueryFilter 定义查询过滤器
type QueryFilter struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// ExecutionStats 代表Agent执行统计
type ExecutionStats struct {
	TotalExecutions      int            `json:"total_executions"`
	SuccessfulExecutions int            `json:"successful_executions"`
	FailedExecutions     int            `json:"failed_executions"`
	TotalExecutionTime   time.Duration  `json:"total_execution_time"`
	AverageExecutionTime time.Duration  `json:"average_execution_time"`
	LastExecutionTime    time.Time      `json:"last_execution_time"`
	TokensUsed           int            `json:"tokens_used"`
	TotalCost            float64        `json:"total_cost"`
	ToolsUsed            map[string]int `json:"tools_used"`
	CreatedAt            time.Time      `json:"created_at"`
}

// MemoryStats 代表记忆系统统计
type MemoryStats struct {
	TotalItems   int       `json:"total_items"`
	TotalSize    int64     `json:"total_size"`
	HitRate      float64   `json:"hit_rate"`
	MissRate     float64   `json:"miss_rate"`
	LastAccessed time.Time `json:"last_accessed"`
}

// KnowledgeStats 代表知识源统计
type KnowledgeStats struct {
	TotalItems   int       `json:"total_items"`
	TotalQueries int       `json:"total_queries"`
	AverageScore float64   `json:"average_score"`
	LastQueried  time.Time `json:"last_queried"`
	IndexSize    int64     `json:"index_size"`
}

// AgentRole 定义预设的Agent角色
type AgentRole string

const (
	RoleSoftwareEngineer   AgentRole = "Software Engineer"
	RoleDataAnalyst        AgentRole = "Data Analyst"
	RoleProjectManager     AgentRole = "Project Manager"
	RoleResearcher         AgentRole = "Researcher"
	RoleContentWriter      AgentRole = "Content Writer"
	RoleQualityAssurance   AgentRole = "Quality Assurance"
	RoleDevOpsEngineer     AgentRole = "DevOps Engineer"
	RoleProductManager     AgentRole = "Product Manager"
	RoleUIUXDesigner       AgentRole = "UI/UX Designer"
	RoleSecuritySpecialist AgentRole = "Security Specialist"
)

// AgentConfig 定义Agent创建配置
type AgentConfig struct {
	Role              string                                     `json:"role"`
	Goal              string                                     `json:"goal"`
	Backstory         string                                     `json:"backstory"`
	LLM               llm.LLM                                    `json:"-"`
	Tools             []Tool                                     `json:"-"`
	ExecutionConfig   ExecutionConfig                            `json:"execution_config"`
	Memory            Memory                                     `json:"-"`
	KnowledgeSources  []KnowledgeSource                          `json:"-"`
	HumanInputHandler HumanInputHandler                          `json:"-"`
	EventBus          events.EventBus                            `json:"-"`
	Logger            logger.Logger                              `json:"-"`
	SecurityConfig    security.SecurityConfig                    `json:"security_config"`
	SystemTemplate    string                                     `json:"system_template"`
	PromptTemplate    string                                     `json:"prompt_template"`
	Callbacks         []func(context.Context, *TaskOutput) error `json:"-"`
}

// DefaultExecutionConfig 返回默认的执行配置
func DefaultExecutionConfig() ExecutionConfig {
	return ExecutionConfig{
		MaxIterations:    25,
		MaxRPM:           60,
		Timeout:          30 * time.Minute,
		MaxExecutionTime: 10 * time.Minute,
		AllowDelegation:  false,
		VerboseLogging:   false,
		HumanInput:       false,
		UseSystemPrompt:  true,
		MaxTokens:        4096,
		Temperature:      0.7,
		CacheEnabled:     true,
		MaxRetryLimit:    3,
	}
}

// DefaultQueryOptions 返回默认的查询选项
func DefaultQueryOptions() QueryOptions {
	return QueryOptions{
		Limit:     10,
		Threshold: 0.7,
		Metadata:  make(map[string]interface{}),
		Filters:   make([]QueryFilter, 0),
	}
}
