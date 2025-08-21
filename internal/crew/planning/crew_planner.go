package planning

import (
	"context"
	"fmt"
	"time"

	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// CrewPlannerImpl CrewPlanner接口的实现，对应Python版本的CrewPlanner类
// 负责规划和协调crew任务的执行
type CrewPlannerImpl struct {
	tasks            []TaskInfo           // 任务列表
	config           *PlanningConfig      // 规划配置
	eventBus         events.EventBus      // 事件总线
	logger           logger.Logger        // 日志器
	agentFactory     AgentFactory         // 代理工厂
	taskFactory      TaskFactory          // 任务工厂
	summaryGenerator TaskSummaryGenerator // 摘要生成器
	validator        PlanValidator        // 计划验证器
	executor         PlanExecutor         // 计划执行器
}

// NewCrewPlanner 创建新的CrewPlanner实例，对应Python版本的__init__()
func NewCrewPlanner(
	tasks []TaskInfo,
	config *PlanningConfig,
	eventBus events.EventBus,
	logger logger.Logger,
	agentFactory AgentFactory,
	taskFactory TaskFactory,
) *CrewPlannerImpl {
	// 如果配置为空，使用默认配置
	if config == nil {
		config = DefaultPlanningConfig()
	}

	planner := &CrewPlannerImpl{
		tasks:        tasks,
		config:       config,
		eventBus:     eventBus,
		logger:       logger,
		agentFactory: agentFactory,
		taskFactory:  taskFactory,
	}

	// 初始化组件
	planner.summaryGenerator = NewTaskSummaryGenerator(logger)
	planner.validator = NewPlanValidator()
	planner.executor = NewPlanExecutor(agentFactory, taskFactory, eventBus, logger)

	return planner
}

// HandleCrewPlanning 处理crew规划，核心方法，对应Python版本的_handle_crew_planning()
// 创建详细的任务步骤规划
func (cp *CrewPlannerImpl) HandleCrewPlanning(ctx context.Context) (*PlannerTaskPydanticOutput, error) {
	startTime := time.Now()

	// 发射规划开始事件
	if cp.eventBus != nil {
		event := NewPlanningStartedEvent(len(cp.tasks), cp.config.PlanningAgentLLM, cp.config)
		if err := cp.eventBus.Emit(ctx, cp, event); err != nil {
			cp.logger.Error("Failed to emit planning started event",
				logger.Field{Key: "error", Value: err})
		}
	}

	cp.logger.Info("Starting crew planning",
		logger.Field{Key: "task_count", Value: len(cp.tasks)},
		logger.Field{Key: "planning_llm", Value: cp.config.PlanningAgentLLM},
	)

	// 验证任务信息
	if err := cp.validator.ValidateTaskInfos(ctx, cp.tasks); err != nil {
		return nil, cp.handlePlanningError(ctx, "validation", err, startTime)
	}

	// 创建规划代理
	planningAgent, err := cp.createPlanningAgent(ctx)
	if err != nil {
		return nil, cp.handlePlanningError(ctx, "agent_creation", err, startTime)
	}

	// 创建任务摘要
	tasksSummary, err := cp.CreateTasksSummary(ctx)
	if err != nil {
		return nil, cp.handlePlanningError(ctx, "tasks_summary", err, startTime)
	}

	// 创建规划任务
	plannerTask, err := cp.createPlannerTask(ctx, planningAgent, tasksSummary)
	if err != nil {
		return nil, cp.handlePlanningError(ctx, "planner_task", err, startTime)
	}

	// 执行规划任务
	result, err := cp.executePlannerTask(ctx, plannerTask)
	if err != nil {
		return nil, cp.handlePlanningError(ctx, "execution", err, startTime)
	}

	// 验证规划结果
	if err := cp.validator.ValidatePlanOutput(ctx, result); err != nil {
		return nil, cp.handlePlanningError(ctx, "result_validation", err, startTime)
	}

	executionTime := float64(time.Since(startTime).Nanoseconds()) / 1e6

	// 发射规划完成事件
	if cp.eventBus != nil {
		event := NewPlanningCompletedEvent(len(cp.tasks), result.GetTaskCount(), executionTime, true)
		event.Result = result
		if err := cp.eventBus.Emit(ctx, cp, event); err != nil {
			cp.logger.Error("Failed to emit planning completed event",
				logger.Field{Key: "error", Value: err})
		}
	}

	cp.logger.Info("Crew planning completed successfully",
		logger.Field{Key: "task_count", Value: len(cp.tasks)},
		logger.Field{Key: "plans_generated", Value: result.GetTaskCount()},
		logger.Field{Key: "execution_time_ms", Value: executionTime},
	)

	return result, nil
}

// CreateTasksSummary 创建任务摘要，对应Python版本的_create_tasks_summary()
func (cp *CrewPlannerImpl) CreateTasksSummary(ctx context.Context) (string, error) {
	if cp.summaryGenerator == nil {
		return "", fmt.Errorf("task summary generator not initialized")
	}

	summary, err := cp.summaryGenerator.GenerateTasksSummary(ctx, cp.tasks)
	if err != nil {
		cp.logger.Error("Failed to create tasks summary", logger.Field{Key: "error", Value: err})
		return "", fmt.Errorf("failed to create tasks summary: %w", err)
	}

	cp.logger.Debug("Tasks summary created",
		logger.Field{Key: "summary_length", Value: len(summary)},
		logger.Field{Key: "task_count", Value: len(cp.tasks)},
	)

	return summary, nil
}

// createPlanningAgent 创建规划代理，对应Python版本的_create_planning_agent()
func (cp *CrewPlannerImpl) createPlanningAgent(ctx context.Context) (Agent, error) {
	startTime := time.Now()

	// 与Python版本保持一致的规划代理配置
	agentConfig := Config{
		Role: "Task Execution Planner",
		Goal: "Your goal is to create an extremely detailed, step-by-step plan based on the tasks and tools " +
			"available to each agent so that they can perform the tasks in an exemplary manner",
		Backstory:   "Planner agent for crew planning",
		LLM:         cp.config.PlanningAgentLLM,
		Verbose:     cp.config.EnableVerbose,
		MaxIter:     10,  // 默认最大迭代次数
		Temperature: 0.1, // 保持规划的一致性
	}

	planningAgent, err := cp.agentFactory.CreateAgent(ctx, AdaptAgentConfig(&agentConfig))
	if err != nil {
		return nil, NewAgentCreationError("Task Execution Planner", err.Error())
	}

	creationTime := float64(time.Since(startTime).Nanoseconds()) / 1e6

	// 发射代理创建事件
	if cp.eventBus != nil {
		event := NewPlanningAgentCreatedEvent(
			planningAgent.ID(),
			agentConfig.Role,
			agentConfig.LLM,
			creationTime,
		)
		if err := cp.eventBus.Emit(ctx, cp, event); err != nil {
			cp.logger.Error("Failed to emit planning agent created event",
				logger.Field{Key: "error", Value: err})
		}
	}

	cp.logger.Debug("Planning agent created",
		logger.Field{Key: "agent_id", Value: planningAgent.ID()},
		logger.Field{Key: "llm", Value: agentConfig.LLM},
		logger.Field{Key: "creation_time_ms", Value: creationTime},
	)

	return planningAgent, nil
}

// createPlannerTask 创建规划任务，对应Python版本的_create_planner_task()
func (cp *CrewPlannerImpl) createPlannerTask(ctx context.Context, planningAgent Agent, tasksSummary string) (Task, error) {
	startTime := time.Now()

	// 构建规划任务描述，与Python版本保持一致的格式
	description := fmt.Sprintf(`
Based on the following tasks and the available tools for each agent, create an extremely detailed step-by-step plan for each task:

%s

For each task, provide:
1. A clear breakdown of what the agent needs to do
2. The specific steps in order
3. How to use the available tools effectively
4. Any dependencies or prerequisites
5. Expected outputs at each step

Return the result in JSON format with the following structure:
{
  "list_of_plans_per_task": [
    {
      "task": "task description",
      "plan": "detailed step-by-step plan"
    }
  ]
}`, tasksSummary)

	expectedOutput := "JSON object containing detailed plans for each task"

	taskConfig := AdaptTaskConfig(description, expectedOutput, planningAgent, true, cp.config.EnableVerbose)

	plannerTask, err := cp.taskFactory.CreateTask(ctx, taskConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create planner task: %w", err)
	}

	creationTime := float64(time.Since(startTime).Nanoseconds()) / 1e6

	// 发射任务创建事件
	if cp.eventBus != nil {
		event := NewPlanningTaskCreatedEvent(
			plannerTask.ID(),
			"planner",
			planningAgent.ID(),
			creationTime,
		)
		if err := cp.eventBus.Emit(ctx, cp, event); err != nil {
			cp.logger.Error("Failed to emit planning task created event",
				logger.Field{Key: "error", Value: err})
		}
	}

	cp.logger.Debug("Planner task created",
		logger.Field{Key: "task_id", Value: plannerTask.ID()},
		logger.Field{Key: "agent_id", Value: planningAgent.ID()},
		logger.Field{Key: "creation_time_ms", Value: creationTime},
	)

	return plannerTask, nil
}

// executePlannerTask 执行规划任务
func (cp *CrewPlannerImpl) executePlannerTask(ctx context.Context, plannerTask Task) (*PlannerTaskPydanticOutput, error) {
	// 创建带超时的上下文
	if cp.config.TimeoutSeconds > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(cp.config.TimeoutSeconds)*time.Second)
		defer cancel()
	}

	// 同步执行任务，对应Python版本的execute_sync()
	mockTaskOutput, err := plannerTask.ExecuteSync(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute planning task: %w", err)
	}

	// 适配输出格式
	taskOutput := AdaptTaskOutput(mockTaskOutput)

	// 解析任务输出
	var result PlannerTaskPydanticOutput
	if taskOutput.JSONDict != nil {
		// 从JSON字典解析
		if err := cp.parseFromJSONDict(taskOutput.JSONDict, &result); err != nil {
			return nil, fmt.Errorf("failed to parse planning output from JSON dict: %w", err)
		}
	} else if taskOutput.Raw != "" {
		// 从原始文本解析
		if err := result.FromJSON(taskOutput.Raw); err != nil {
			return nil, fmt.Errorf("failed to parse planning output from raw text: %w", err)
		}
	} else {
		return nil, ErrInvalidPlanOutput
	}

	return &result, nil
}

// parseFromJSONDict 从JSON字典解析规划结果
func (cp *CrewPlannerImpl) parseFromJSONDict(jsonDict map[string]interface{}, result *PlannerTaskPydanticOutput) error {
	if listInterface, ok := jsonDict["list_of_plans_per_task"]; ok {
		if listSlice, ok := listInterface.([]interface{}); ok {
			result.ListOfPlansPerTask = make([]PlanPerTask, len(listSlice))
			for i, item := range listSlice {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if taskStr, ok := itemMap["task"].(string); ok {
						result.ListOfPlansPerTask[i].Task = taskStr
					}
					if planStr, ok := itemMap["plan"].(string); ok {
						result.ListOfPlansPerTask[i].Plan = planStr
					}
				}
			}
			return nil
		}
	}
	return ErrInvalidLLMResponse
}

// handlePlanningError 处理规划错误
func (cp *CrewPlannerImpl) handlePlanningError(ctx context.Context, phase string, err error, startTime time.Time) error {
	executionTime := float64(time.Since(startTime).Nanoseconds()) / 1e6

	// 发射规划失败事件
	if cp.eventBus != nil {
		event := NewPlanningFailedEvent(
			len(cp.tasks),
			err.Error(),
			fmt.Sprintf("%T", err),
			phase,
			executionTime,
		)
		if emitErr := cp.eventBus.Emit(ctx, cp, event); emitErr != nil {
			cp.logger.Error("Failed to emit planning failed event",
				logger.Field{Key: "error", Value: emitErr})
		}
	}

	cp.logger.Error("Planning failed",
		logger.Field{Key: "phase", Value: phase},
		logger.Field{Key: "error", Value: err},
		logger.Field{Key: "execution_time_ms", Value: executionTime},
	)

	return NewPlanningExecutionError(phase, err, 0, len(cp.tasks), "", cp.config.PlanningAgentLLM)
}

// SetConfig 设置规划配置
func (cp *CrewPlannerImpl) SetConfig(config *PlanningConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	cp.config = config

	// 更新执行器配置
	if cp.executor != nil {
		cp.executor.SetMaxRetries(config.MaxRetries)
		cp.executor.SetTimeout(config.TimeoutSeconds)
	}

	cp.logger.Debug("Planning configuration updated",
		logger.Field{Key: "planning_llm", Value: config.PlanningAgentLLM},
		logger.Field{Key: "max_retries", Value: config.MaxRetries},
		logger.Field{Key: "timeout_seconds", Value: config.TimeoutSeconds},
	)

	return nil
}

// GetConfig 获取当前规划配置
func (cp *CrewPlannerImpl) GetConfig() *PlanningConfig {
	return cp.config
}

// ValidateConfiguration 验证规划配置
func (cp *CrewPlannerImpl) ValidateConfiguration() error {
	if cp.config == nil {
		return fmt.Errorf("configuration is nil")
	}

	if cp.config.PlanningAgentLLM == "" {
		return fmt.Errorf("planning_agent_llm cannot be empty")
	}

	if cp.config.MaxRetries < 0 {
		return fmt.Errorf("max_retries cannot be negative")
	}

	if cp.config.TimeoutSeconds <= 0 {
		return fmt.Errorf("timeout_seconds must be positive")
	}

	return nil
}

// GetTaskCount 获取任务数量
func (cp *CrewPlannerImpl) GetTaskCount() int {
	return len(cp.tasks)
}

// GetTasks 获取任务列表（只读）
func (cp *CrewPlannerImpl) GetTasks() []TaskInfo {
	// 返回拷贝以确保不被修改
	tasks := make([]TaskInfo, len(cp.tasks))
	copy(tasks, cp.tasks)
	return tasks
}
