package crew

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ynl/greensoulai/internal/agent"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
	"github.com/ynl/greensoulai/pkg/security"
)

// BaseCrew 实现Crew接口的基础结构
type BaseCrew struct {
	id               string
	name             string
	agents           []agent.Agent
	tasks            []agent.Task
	process          Process
	verbose          bool
	memoryEnabled    bool
	cacheEnabled     bool
	maxRPM           int
	shareCrewEnabled bool
	planningEnabled  bool
	maxExecutionTime time.Duration
	fullOutput       bool

	// 回调函数
	beforeKickoffCallbacks []KickoffCallback
	afterKickoffCallbacks  []KickoffCallback
	taskCallback           TaskCallback
	stepCallback           StepCallback

	// 管理器相关
	managerAgent       agent.Agent
	managerLLM         interface{}
	functionCallingLLM interface{}
	chatLLM            interface{}

	// 基础设施
	eventBus       events.EventBus
	logger         logger.Logger
	securityConfig security.SecurityConfig
	memory         Memory
	cache          Cache

	// 执行统计
	usageMetrics       *UsageMetrics
	executionCount     int
	lastExecutionTime  time.Time
	totalExecutionTime time.Duration

	// 并发控制
	mu        sync.RWMutex
	executing bool
}

// NewBaseCrew 创建新的BaseCrew实例
func NewBaseCrew(config *CrewConfig, eventBus events.EventBus, logger logger.Logger) *BaseCrew {
	if config == nil {
		config = DefaultCrewConfig()
	}

	return &BaseCrew{
		id:                     uuid.New().String(),
		name:                   config.Name,
		agents:                 make([]agent.Agent, 0),
		tasks:                  make([]agent.Task, 0),
		process:                config.Process,
		verbose:                config.Verbose,
		memoryEnabled:          config.MemoryEnabled,
		cacheEnabled:           config.CacheEnabled,
		maxRPM:                 config.MaxRPM,
		shareCrewEnabled:       config.ShareCrew,
		planningEnabled:        config.PlanningEnabled,
		maxExecutionTime:       config.MaxExecutionTime,
		fullOutput:             config.FullOutput,
		beforeKickoffCallbacks: make([]KickoffCallback, 0),
		afterKickoffCallbacks:  make([]KickoffCallback, 0),
		taskCallback:           config.TaskCallback,
		stepCallback:           config.StepCallback,
		managerAgent:           config.ManagerAgent,
		managerLLM:             config.ManagerLLM,
		functionCallingLLM:     config.FunctionCallingLLM,
		chatLLM:                config.ChatLLM,
		eventBus:               eventBus,
		logger:                 logger,
		securityConfig:         *security.NewSecurityConfig(),
		usageMetrics:           &UsageMetrics{},
		executionCount:         0,
		executing:              false,
	}
}

// Kickoff 启动Crew执行
func (c *BaseCrew) Kickoff(ctx context.Context, inputs map[string]interface{}) (*CrewOutput, error) {
	c.mu.Lock()
	if c.executing {
		c.mu.Unlock()
		return nil, fmt.Errorf("crew is already executing")
	}
	c.executing = true
	c.executionCount++
	executionID := c.executionCount
	c.lastExecutionTime = time.Now()
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.executing = false
		c.mu.Unlock()
	}()

	// 发射开始事件
	startEvent := NewCrewKickoffStartedEvent(c.id, c.name, executionID, c.process.String())
	c.eventBus.Emit(ctx, c, startEvent)

	c.logger.Info("crew kickoff started",
		logger.Field{Key: "crew_id", Value: c.id},
		logger.Field{Key: "crew_name", Value: c.name},
		logger.Field{Key: "execution_id", Value: executionID},
		logger.Field{Key: "process", Value: c.process.String()},
		logger.Field{Key: "agents_count", Value: len(c.agents)},
		logger.Field{Key: "tasks_count", Value: len(c.tasks)},
	)

	// 验证配置
	if err := c.validateConfiguration(); err != nil {
		c.logger.Error("crew configuration validation failed",
			logger.Field{Key: "error", Value: err},
		)
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// 执行前回调
	for _, callback := range c.beforeKickoffCallbacks {
		if _, err := callback(ctx, c, nil); err != nil {
			c.logger.Error("before kickoff callback failed",
				logger.Field{Key: "error", Value: err},
			)
			return nil, fmt.Errorf("before kickoff callback failed: %w", err)
		}
	}

	// 规划处理
	if c.planningEnabled {
		if err := c.handleCrewPlanning(ctx, inputs); err != nil {
			c.logger.Error("crew planning failed",
				logger.Field{Key: "error", Value: err},
			)
			return nil, fmt.Errorf("crew planning failed: %w", err)
		}
	}

	// 执行任务
	start := time.Now()
	var result *CrewOutput
	var err error

	switch c.process {
	case ProcessSequential:
		result, err = c.runSequentialProcess(ctx, inputs)
	case ProcessHierarchical:
		result, err = c.runHierarchicalProcess(ctx, inputs)
	default:
		err = fmt.Errorf("unsupported process: %v", c.process)
	}

	duration := time.Since(start)
	c.mu.Lock()
	c.totalExecutionTime += duration
	c.mu.Unlock()

	if result != nil {
		result.Duration = duration
		result.Success = err == nil
		result.Error = err
		result.CreatedAt = time.Now()
	}

	// 执行后回调
	for _, callback := range c.afterKickoffCallbacks {
		if result, err = callback(ctx, c, result); err != nil {
			c.logger.Error("after kickoff callback failed",
				logger.Field{Key: "error", Value: err},
			)
			return result, fmt.Errorf("after kickoff callback failed: %w", err)
		}
	}

	// 计算使用统计
	c.calculateUsageMetrics(result)

	// 发射完成事件
	completedEvent := NewCrewKickoffCompletedEvent(c.id, c.name, executionID, duration, err == nil)
	c.eventBus.Emit(ctx, c, completedEvent)

	if err != nil {
		c.logger.Error("crew execution failed",
			logger.Field{Key: "crew_id", Value: c.id},
			logger.Field{Key: "crew_name", Value: c.name},
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "duration", Value: duration},
		)
	} else {
		c.logger.Info("crew execution completed",
			logger.Field{Key: "crew_id", Value: c.id},
			logger.Field{Key: "crew_name", Value: c.name},
			logger.Field{Key: "duration", Value: duration},
			logger.Field{Key: "tasks_completed", Value: len(result.TasksOutput)},
		)
	}

	return result, err
}

// KickoffAsync 异步启动Crew执行
func (c *BaseCrew) KickoffAsync(ctx context.Context, inputs map[string]interface{}) (<-chan CrewResult, error) {
	resultChan := make(chan CrewResult, 1)

	go func() {
		defer close(resultChan)
		output, err := c.Kickoff(ctx, inputs)
		resultChan <- CrewResult{Output: output, Error: err}
	}()

	return resultChan, nil
}

// KickoffWithTimeout 带超时的Crew执行
func (c *BaseCrew) KickoffWithTimeout(ctx context.Context, inputs map[string]interface{}, timeout time.Duration) (*CrewOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resultChan := make(chan CrewResult, 1)

	go func() {
		output, err := c.Kickoff(ctx, inputs)
		resultChan <- CrewResult{Output: output, Error: err}
	}()

	select {
	case result := <-resultChan:
		return result.Output, result.Error
	case <-ctx.Done():
		return nil, fmt.Errorf("crew execution timeout after %v: %w", timeout, ctx.Err())
	}
}

// KickoffForEach 为每个输入执行Crew
func (c *BaseCrew) KickoffForEach(ctx context.Context, inputsList []map[string]interface{}) ([]*CrewOutput, error) {
	results := make([]*CrewOutput, 0, len(inputsList))
	totalUsageMetrics := &UsageMetrics{}

	for i, inputs := range inputsList {
		c.logger.Info("executing crew for input set",
			logger.Field{Key: "crew_name", Value: c.name},
			logger.Field{Key: "input_index", Value: i},
			logger.Field{Key: "total_inputs", Value: len(inputsList)},
		)

		// 创建crew副本以避免状态冲突
		crewCopy, err := c.Clone()
		if err != nil {
			return results, fmt.Errorf("failed to clone crew for input %d: %w", i, err)
		}

		output, err := crewCopy.Kickoff(ctx, inputs)
		if err != nil {
			c.logger.Error("crew execution failed for input set",
				logger.Field{Key: "input_index", Value: i},
				logger.Field{Key: "error", Value: err},
			)
			return results, fmt.Errorf("execution failed for input %d: %w", i, err)
		}

		results = append(results, output)

		// 累计使用统计
		if crewUsage := crewCopy.GetUsageMetrics(); crewUsage != nil {
			totalUsageMetrics.AddUsageMetrics(crewUsage)
		}
	}

	// 更新总的使用统计
	c.mu.Lock()
	c.usageMetrics = totalUsageMetrics
	c.mu.Unlock()

	return results, nil
}

// KickoffForEachAsync 异步为每个输入执行Crew
func (c *BaseCrew) KickoffForEachAsync(ctx context.Context, inputsList []map[string]interface{}) (<-chan []*CrewOutput, error) {
	resultChan := make(chan []*CrewOutput, 1)

	go func() {
		defer close(resultChan)
		results, err := c.KickoffForEach(ctx, inputsList)
		if err != nil {
			c.logger.Error("async kickoff for each failed", logger.Field{Key: "error", Value: err})
			resultChan <- nil
			return
		}
		resultChan <- results
	}()

	return resultChan, nil
}

// Train 训练Crew模型
func (c *BaseCrew) Train(ctx context.Context, nIterations int, filename string, inputs map[string]interface{}) error {
	// 需要导入训练包，但为了避免循环依赖，这里使用接口
	return c.TrainWithConfig(ctx, &TrainingConfig{
		Iterations: nIterations,
		Filename:   filename,
		Inputs:     inputs,
	})
}

// TrainingConfig 训练配置（临时定义，实际应该使用training包的配置）
type TrainingConfig struct {
	Iterations      int
	Filename        string
	Inputs          map[string]interface{}
	CollectFeedback bool
	MetricsEnabled  bool
	AutoSave        bool
}

// TrainWithConfig 使用配置进行训练
func (c *BaseCrew) TrainWithConfig(ctx context.Context, config *TrainingConfig) error {
	c.logger.Info("starting crew training with config",
		logger.Field{Key: "crew_name", Value: c.name},
		logger.Field{Key: "iterations", Value: config.Iterations},
		logger.Field{Key: "filename", Value: config.Filename},
	)

	// 创建训练副本
	trainCrew, err := c.Copy()
	if err != nil {
		return fmt.Errorf("failed to create training copy: %w", err)
	}

	// 设置默认配置
	if config.CollectFeedback == false {
		config.CollectFeedback = true
	}
	if config.MetricsEnabled == false {
		config.MetricsEnabled = true
	}
	if config.AutoSave == false {
		config.AutoSave = true
	}

	// 创建执行函数
	executeFunc := func(ctx context.Context, inputs map[string]interface{}) (interface{}, error) {
		output, err := trainCrew.Kickoff(ctx, inputs)
		return output, err
	}

	// 执行训练迭代
	var trainingErrors []error
	for iteration := 0; iteration < config.Iterations; iteration++ {
		c.logger.Info("executing training iteration",
			logger.Field{Key: "iteration", Value: iteration},
			logger.Field{Key: "total_iterations", Value: config.Iterations},
		)

		// 模拟训练处理器行为（实际使用时应该创建真正的训练处理器）
		start := time.Now()
		output, err := executeFunc(ctx, config.Inputs)
		duration := time.Since(start)

		if err != nil {
			c.logger.Error("training iteration failed",
				logger.Field{Key: "iteration", Value: iteration},
				logger.Field{Key: "error", Value: err},
			)
			trainingErrors = append(trainingErrors, err)
			continue
		}

		c.logger.Info("training iteration completed",
			logger.Field{Key: "iteration", Value: iteration},
			logger.Field{Key: "duration", Value: duration},
			logger.Field{Key: "success", Value: output != nil},
		)

		// 模拟反馈收集（实际使用时应该调用训练处理器）
		if config.CollectFeedback {
			c.logger.Debug("feedback collection would occur here",
				logger.Field{Key: "iteration", Value: iteration},
			)
		}

		// 模拟性能分析（实际使用时应该调用训练处理器）
		if config.MetricsEnabled {
			c.logger.Debug("metrics analysis would occur here",
				logger.Field{Key: "iteration", Value: iteration},
				logger.Field{Key: "execution_time", Value: duration},
			)
		}

		// 模拟自动保存（实际使用时应该调用训练处理器）
		if config.AutoSave && iteration%5 == 0 {
			c.logger.Debug("auto-save would occur here",
				logger.Field{Key: "iteration", Value: iteration},
				logger.Field{Key: "filename", Value: config.Filename},
			)
		}
	}

	// 训练完成统计
	successRate := float64(config.Iterations-len(trainingErrors)) / float64(config.Iterations) * 100
	c.logger.Info("crew training completed",
		logger.Field{Key: "crew_name", Value: c.name},
		logger.Field{Key: "total_iterations", Value: config.Iterations},
		logger.Field{Key: "success_rate", Value: successRate},
		logger.Field{Key: "failed_iterations", Value: len(trainingErrors)},
	)

	// TODO: 最终保存训练数据
	// TODO: 生成训练报告
	// TODO: 应用训练结果到crew

	if len(trainingErrors) > 0 && len(trainingErrors) == config.Iterations {
		return fmt.Errorf("all training iterations failed")
	}

	return nil
}

// validateConfiguration 验证Crew配置
func (c *BaseCrew) validateConfiguration() error {
	if len(c.agents) == 0 {
		return fmt.Errorf("crew must have at least one agent")
	}

	if len(c.tasks) == 0 {
		return fmt.Errorf("crew must have at least one task")
	}

	if len(c.tasks) > len(c.agents) && c.process == ProcessSequential {
		c.logger.Warn("more tasks than agents in sequential process",
			logger.Field{Key: "tasks_count", Value: len(c.tasks)},
			logger.Field{Key: "agents_count", Value: len(c.agents)},
		)
	}

	// 验证层级模式配置
	if c.process == ProcessHierarchical {
		if c.managerAgent == nil && c.managerLLM == nil {
			return fmt.Errorf("hierarchical process requires either manager agent or manager LLM")
		}
	}

	return nil
}

// handleCrewPlanning 处理Crew规划
func (c *BaseCrew) handleCrewPlanning(ctx context.Context, inputs map[string]interface{}) error {
	c.logger.Info("handling crew planning",
		logger.Field{Key: "crew_name", Value: c.name},
		logger.Field{Key: "tasks_count", Value: len(c.tasks)},
	)

	// TODO: 实现规划逻辑
	// 这里应该包括：
	// 1. 分析任务依赖关系
	// 2. 优化执行顺序
	// 3. 资源分配
	// 4. 风险评估

	return nil
}

// calculateUsageMetrics 计算使用统计
func (c *BaseCrew) calculateUsageMetrics(result *CrewOutput) {
	if result == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	metrics := &UsageMetrics{
		TotalTasks:    len(c.tasks),
		ExecutionTime: result.Duration,
	}

	// 统计任务结果
	for _, taskOutput := range result.TasksOutput {
		if taskOutput != nil && taskOutput.IsValid {
			metrics.SuccessfulTasks++
		} else {
			metrics.FailedTasks++
		}

		// TODO: 从任务输出中提取token使用情况
		// 这需要与LLM模块集成
	}

	c.usageMetrics = metrics
	result.TokenUsage = metrics
}

// 接口实现方法
func (c *BaseCrew) AddAgent(agentToAdd agent.Agent) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.agents = append(c.agents, agentToAdd)
	c.logger.Info("agent added to crew",
		logger.Field{Key: "crew_name", Value: c.name},
		logger.Field{Key: "agent_role", Value: agentToAdd.GetRole()},
		logger.Field{Key: "total_agents", Value: len(c.agents)},
	)
	return nil
}

func (c *BaseCrew) AddTask(task agent.Task) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.tasks = append(c.tasks, task)
	c.logger.Info("task added to crew",
		logger.Field{Key: "crew_name", Value: c.name},
		logger.Field{Key: "task_description", Value: task.GetDescription()},
		logger.Field{Key: "total_tasks", Value: len(c.tasks)},
	)
	return nil
}

func (c *BaseCrew) SetProcess(process Process) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.process = process
}

func (c *BaseCrew) SetVerbose(verbose bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.verbose = verbose
}

func (c *BaseCrew) SetMemoryEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.memoryEnabled = enabled
}

func (c *BaseCrew) SetCacheEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cacheEnabled = enabled
}

func (c *BaseCrew) AddBeforeKickoffCallback(callback KickoffCallback) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.beforeKickoffCallbacks = append(c.beforeKickoffCallbacks, callback)
	return nil
}

func (c *BaseCrew) AddAfterKickoffCallback(callback KickoffCallback) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.afterKickoffCallbacks = append(c.afterKickoffCallbacks, callback)
	return nil
}

func (c *BaseCrew) AddTaskCallback(callback TaskCallback) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.taskCallback = callback
	return nil
}

func (c *BaseCrew) AddStepCallback(callback StepCallback) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stepCallback = callback
	return nil
}

// 状态查询方法
func (c *BaseCrew) GetAgents() []agent.Agent {
	c.mu.RLock()
	defer c.mu.RUnlock()

	agents := make([]agent.Agent, len(c.agents))
	copy(agents, c.agents)
	return agents
}

func (c *BaseCrew) GetTasks() []agent.Task {
	c.mu.RLock()
	defer c.mu.RUnlock()

	tasks := make([]agent.Task, len(c.tasks))
	copy(tasks, c.tasks)
	return tasks
}

func (c *BaseCrew) GetProcess() Process {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.process
}

func (c *BaseCrew) IsMemoryEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.memoryEnabled
}

func (c *BaseCrew) IsCacheEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cacheEnabled
}

func (c *BaseCrew) GetUsageMetrics() *UsageMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.usageMetrics == nil {
		return &UsageMetrics{}
	}

	// 返回副本以避免并发修改
	metrics := *c.usageMetrics
	return &metrics
}

// Clone 创建Crew的副本
func (c *BaseCrew) Clone() (Crew, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	config := &CrewConfig{
		Name:               c.name + "_clone",
		Process:            c.process,
		Verbose:            c.verbose,
		MemoryEnabled:      c.memoryEnabled,
		CacheEnabled:       c.cacheEnabled,
		MaxRPM:             c.maxRPM,
		ShareCrew:          c.shareCrewEnabled,
		PlanningEnabled:    c.planningEnabled,
		MaxExecutionTime:   c.maxExecutionTime,
		FullOutput:         c.fullOutput,
		TaskCallback:       c.taskCallback,
		StepCallback:       c.stepCallback,
		ManagerAgent:       c.managerAgent,
		ManagerLLM:         c.managerLLM,
		FunctionCallingLLM: c.functionCallingLLM,
		ChatLLM:            c.chatLLM,
	}

	clone := NewBaseCrew(config, c.eventBus, c.logger)

	// 复制agents和tasks
	for _, agentToCopy := range c.agents {
		// TODO: 实现agent的Clone方法
		clone.AddAgent(agentToCopy)
	}

	for _, task := range c.tasks {
		// TODO: 实现task的Clone方法
		clone.AddTask(task)
	}

	// 复制回调
	clone.beforeKickoffCallbacks = make([]KickoffCallback, len(c.beforeKickoffCallbacks))
	copy(clone.beforeKickoffCallbacks, c.beforeKickoffCallbacks)

	clone.afterKickoffCallbacks = make([]KickoffCallback, len(c.afterKickoffCallbacks))
	copy(clone.afterKickoffCallbacks, c.afterKickoffCallbacks)

	return clone, nil
}

// Copy 创建Crew的浅拷贝，用于并行执行
func (c *BaseCrew) Copy() (Crew, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	config := &CrewConfig{
		Name:               c.name + "_copy",
		Process:            c.process,
		Verbose:            c.verbose,
		MemoryEnabled:      c.memoryEnabled,
		CacheEnabled:       c.cacheEnabled,
		MaxRPM:             c.maxRPM,
		ShareCrew:          c.shareCrewEnabled,
		PlanningEnabled:    c.planningEnabled,
		MaxExecutionTime:   c.maxExecutionTime,
		FullOutput:         c.fullOutput,
		TaskCallback:       c.taskCallback,
		StepCallback:       c.stepCallback,
		ManagerAgent:       c.managerAgent,
		ManagerLLM:         c.managerLLM,
		FunctionCallingLLM: c.functionCallingLLM,
		ChatLLM:            c.chatLLM,
	}

	crewCopy := NewBaseCrew(config, c.eventBus, c.logger)

	// 直接复制agents和tasks切片（浅拷贝）
	crewCopy.agents = make([]agent.Agent, len(c.agents))
	copy(crewCopy.agents, c.agents)

	crewCopy.tasks = make([]agent.Task, len(c.tasks))
	copy(crewCopy.tasks, c.tasks)

	// 复制回调
	crewCopy.beforeKickoffCallbacks = make([]KickoffCallback, len(c.beforeKickoffCallbacks))
	copy(crewCopy.beforeKickoffCallbacks, c.beforeKickoffCallbacks)

	crewCopy.afterKickoffCallbacks = make([]KickoffCallback, len(c.afterKickoffCallbacks))
	copy(crewCopy.afterKickoffCallbacks, c.afterKickoffCallbacks)

	// 重置执行状态
	crewCopy.executing = false
	crewCopy.executionCount = 0
	crewCopy.usageMetrics = &UsageMetrics{}

	return crewCopy, nil
}

// Close 清理资源
func (c *BaseCrew) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.Info("closing crew", logger.Field{Key: "crew_name", Value: c.name})

	// 清理agents
	for _, agentToClose := range c.agents {
		// TODO: 如果agent有Close方法，调用它
		_ = agentToClose
	}

	// 清理memory和cache
	if c.memory != nil {
		// TODO: 清理memory
	}

	if c.cache != nil {
		// TODO: 清理cache
	}

	// 清空集合
	c.agents = nil
	c.tasks = nil
	c.beforeKickoffCallbacks = nil
	c.afterKickoffCallbacks = nil

	return nil
}
