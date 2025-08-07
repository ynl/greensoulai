package crew

import (
	"fmt"

	"github.com/ynl/greensoulai/internal/agent"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// crewBuilder 实现CrewBuilder接口
type crewBuilder struct {
	config   *CrewConfig
	eventBus events.EventBus
	logger   logger.Logger
	agents   []agent.Agent
	tasks    []agent.Task
}

// NewCrewBuilder 创建新的crew builder
func NewCrewBuilder() CrewBuilder {
	return &crewBuilder{
		config: DefaultCrewConfig(),
		agents: make([]agent.Agent, 0),
		tasks:  make([]agent.Task, 0),
	}
}

// WithName 设置crew名称
func (b *crewBuilder) WithName(name string) CrewBuilder {
	b.config.Name = name
	return b
}

// WithProcess 设置执行流程
func (b *crewBuilder) WithProcess(process Process) CrewBuilder {
	b.config.Process = process
	return b
}

// WithVerbose 设置详细输出模式
func (b *crewBuilder) WithVerbose(verbose bool) CrewBuilder {
	b.config.Verbose = verbose
	return b
}

// WithMemory 设置内存功能
func (b *crewBuilder) WithMemory(enabled bool) CrewBuilder {
	b.config.MemoryEnabled = enabled
	return b
}

// WithCache 设置缓存功能
func (b *crewBuilder) WithCache(enabled bool) CrewBuilder {
	b.config.CacheEnabled = enabled
	return b
}

// WithMaxRPM 设置最大请求速率
func (b *crewBuilder) WithMaxRPM(rpm int) CrewBuilder {
	b.config.MaxRPM = rpm
	return b
}

// WithAgents 添加agents
func (b *crewBuilder) WithAgents(agents ...agent.Agent) CrewBuilder {
	b.agents = append(b.agents, agents...)
	return b
}

// WithTasks 添加tasks
func (b *crewBuilder) WithTasks(tasks ...agent.Task) CrewBuilder {
	b.tasks = append(b.tasks, tasks...)
	return b
}

// WithEventBus 设置事件总线
func (b *crewBuilder) WithEventBus(eventBus events.EventBus) CrewBuilder {
	b.eventBus = eventBus
	return b
}

// WithLogger 设置日志记录器
func (b *crewBuilder) WithLogger(logger logger.Logger) CrewBuilder {
	b.logger = logger
	return b
}

// WithConfig 使用自定义配置
func (b *crewBuilder) WithConfig(config *CrewConfig) CrewBuilder {
	if config != nil {
		b.config = config
	}
	return b
}

// Build 构建Crew实例
func (b *crewBuilder) Build() (Crew, error) {
	// 验证必需的依赖
	if b.eventBus == nil {
		return nil, fmt.Errorf("event bus is required")
	}

	if b.logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	// 创建crew实例
	crew := NewBaseCrew(b.config, b.eventBus, b.logger)

	// 添加agents
	for _, ag := range b.agents {
		if err := crew.AddAgent(ag); err != nil {
			return nil, fmt.Errorf("failed to add agent: %w", err)
		}
	}

	// 添加tasks
	for _, task := range b.tasks {
		if err := crew.AddTask(task); err != nil {
			return nil, fmt.Errorf("failed to add task: %w", err)
		}
	}

	return crew, nil
}

// CrewBuilderWithDefaults 创建带有默认依赖的builder
func CrewBuilderWithDefaults() CrewBuilder {
	// 创建默认的事件总线和日志记录器
	defaultLogger := logger.NewConsoleLogger()
	defaultEventBus := events.NewEventBus(defaultLogger)

	return NewCrewBuilder().
		WithEventBus(defaultEventBus).
		WithLogger(defaultLogger)
}

// QuickCrew 快速创建crew的便利函数
func QuickCrew(name string, process Process, agents []agent.Agent, tasks []agent.Task) (Crew, error) {
	builder := CrewBuilderWithDefaults().
		WithName(name).
		WithProcess(process).
		WithAgents(agents...).
		WithTasks(tasks...)

	return builder.Build()
}

// SequentialCrew 创建顺序执行的crew
func SequentialCrew(name string, agents []agent.Agent, tasks []agent.Task) (Crew, error) {
	return QuickCrew(name, ProcessSequential, agents, tasks)
}

// HierarchicalCrew 创建层级执行的crew
func HierarchicalCrew(name string, agents []agent.Agent, tasks []agent.Task) (Crew, error) {
	return QuickCrew(name, ProcessHierarchical, agents, tasks)
}

// VerboseCrew 创建详细输出模式的crew
func VerboseCrew(name string, process Process, agents []agent.Agent, tasks []agent.Task) (Crew, error) {
	builder := CrewBuilderWithDefaults().
		WithName(name).
		WithProcess(process).
		WithVerbose(true).
		WithAgents(agents...).
		WithTasks(tasks...)

	return builder.Build()
}

// CrewWithMemory 创建启用内存功能的crew
func CrewWithMemory(name string, process Process, agents []agent.Agent, tasks []agent.Task) (Crew, error) {
	builder := CrewBuilderWithDefaults().
		WithName(name).
		WithProcess(process).
		WithMemory(true).
		WithAgents(agents...).
		WithTasks(tasks...)

	return builder.Build()
}

// CrewWithCache 创建启用缓存功能的crew
func CrewWithCache(name string, process Process, agents []agent.Agent, tasks []agent.Task) (Crew, error) {
	builder := CrewBuilderWithDefaults().
		WithName(name).
		WithProcess(process).
		WithCache(true).
		WithAgents(agents...).
		WithTasks(tasks...)

	return builder.Build()
}

// FullFeaturedCrew 创建全功能crew
func FullFeaturedCrew(name string, process Process, agents []agent.Agent, tasks []agent.Task) (Crew, error) {
	builder := CrewBuilderWithDefaults().
		WithName(name).
		WithProcess(process).
		WithVerbose(true).
		WithMemory(true).
		WithCache(true).
		WithMaxRPM(100).
		WithAgents(agents...).
		WithTasks(tasks...)

	return builder.Build()
}
