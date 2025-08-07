package crew

import (
	"context"
	"fmt"
	"time"

	"github.com/ynl/greensoulai/internal/agent"
	"github.com/ynl/greensoulai/internal/llm"
	"github.com/ynl/greensoulai/pkg/logger"
)

// runSequentialProcess 执行顺序流程
// 在Sequential模式中，任务按照预定义的顺序依次执行，前一个任务的输出作为后续任务的上下文
func (c *BaseCrew) runSequentialProcess(ctx context.Context, inputs map[string]interface{}) (*CrewOutput, error) {
	c.logger.Info("starting sequential process execution",
		logger.Field{Key: "crew_name", Value: c.name},
		logger.Field{Key: "tasks_count", Value: len(c.tasks)},
		logger.Field{Key: "agents_count", Value: len(c.agents)},
	)

	// 发射Sequential流程开始事件
	sequentialStartEvent := NewSequentialProcessStartedEvent(c.name, len(c.tasks), len(c.agents))
	c.eventBus.Emit(ctx, c, sequentialStartEvent)

	// 执行任务
	result, err := c.executeTasks(ctx, c.tasks, inputs)

	// 发射Sequential流程完成事件
	if err == nil {
		sequentialCompletedEvent := NewSequentialProcessCompletedEvent(c.name, len(result.TasksOutput))
		c.eventBus.Emit(ctx, c, sequentialCompletedEvent)
	} else {
		sequentialFailedEvent := NewSequentialProcessFailedEvent(c.name, err.Error())
		c.eventBus.Emit(ctx, c, sequentialFailedEvent)
	}

	return result, err
}

// runHierarchicalProcess 执行层级流程
// 在Hierarchical模式中，创建或使用管理器Agent来协调其他Agent执行任务
func (c *BaseCrew) runHierarchicalProcess(ctx context.Context, inputs map[string]interface{}) (*CrewOutput, error) {
	c.logger.Info("starting hierarchical process execution",
		logger.Field{Key: "crew_name", Value: c.name},
		logger.Field{Key: "tasks_count", Value: len(c.tasks)},
		logger.Field{Key: "agents_count", Value: len(c.agents)},
		logger.Field{Key: "has_manager_agent", Value: c.managerAgent != nil},
		logger.Field{Key: "has_manager_llm", Value: c.managerLLM != nil},
	)

	// 发射Hierarchical流程开始事件
	hierarchicalStartEvent := NewHierarchicalProcessStartedEvent(c.name, len(c.tasks), len(c.agents))
	c.eventBus.Emit(ctx, c, hierarchicalStartEvent)

	// 创建或配置管理器agent
	if err := c.createManagerAgent(); err != nil {
		hierarchicalFailedEvent := NewHierarchicalProcessFailedEvent(c.name, fmt.Sprintf("manager creation failed: %s", err.Error()))
		c.eventBus.Emit(ctx, c, hierarchicalFailedEvent)
		return nil, fmt.Errorf("failed to create manager agent: %w", err)
	}

	c.logger.Info("manager agent configured successfully",
		logger.Field{Key: "manager_role", Value: c.managerAgent.GetRole()},
		logger.Field{Key: "manager_id", Value: c.managerAgent.GetID()},
	)

	// 执行任务
	result, err := c.executeTasks(ctx, c.tasks, inputs)

	// 发射Hierarchical流程完成事件
	if err == nil {
		hierarchicalCompletedEvent := NewHierarchicalProcessCompletedEvent(c.name, len(result.TasksOutput))
		c.eventBus.Emit(ctx, c, hierarchicalCompletedEvent)
	} else {
		hierarchicalFailedEvent := NewHierarchicalProcessFailedEvent(c.name, err.Error())
		c.eventBus.Emit(ctx, c, hierarchicalFailedEvent)
	}

	return result, err
}

// executeTasks 执行任务列表
// 支持任务上下文传递，前一个任务的输出会作为后续任务的上下文
func (c *BaseCrew) executeTasks(ctx context.Context, tasks []agent.Task, inputs map[string]interface{}) (*CrewOutput, error) {
	tasksOutput := make([]*agent.TaskOutput, 0, len(tasks))
	var lastOutput *agent.TaskOutput
	var combinedRaw string

	for i, task := range tasks {
		c.logger.Info("executing task",
			logger.Field{Key: "crew_name", Value: c.name},
			logger.Field{Key: "task_index", Value: i},
			logger.Field{Key: "task_id", Value: task.GetID()},
			logger.Field{Key: "task_description", Value: task.GetDescription()},
		)

		// 选择执行该任务的agent
		selectedAgent, err := c.selectAgentForTask(task, i)
		if err != nil {
			c.logger.Error("failed to select agent for task",
				logger.Field{Key: "task_index", Value: i},
				logger.Field{Key: "task_id", Value: task.GetID()},
				logger.Field{Key: "error", Value: err},
			)
			return nil, fmt.Errorf("failed to select agent for task %d (%s): %w", i, task.GetID(), err)
		}

		c.logger.Debug("agent selected for task",
			logger.Field{Key: "task_index", Value: i},
			logger.Field{Key: "task_id", Value: task.GetID()},
			logger.Field{Key: "selected_agent", Value: selectedAgent.GetRole()},
		)

		// 准备任务上下文
		taskContext := c.prepareTaskContext(inputs, tasksOutput, lastOutput)

		// 将上下文应用到任务中
		if len(taskContext) > 0 {
			// 设置任务上下文 - 检查task是否支持SetContext方法
			if contextSetter, ok := task.(interface{ SetContext(map[string]interface{}) }); ok {
				contextSetter.SetContext(taskContext)
				c.logger.Debug("task context applied",
					logger.Field{Key: "task_id", Value: task.GetID()},
					logger.Field{Key: "context_keys", Value: len(taskContext)},
				)
			} else {
				// 如果task不支持SetContext，我们记录一个debug消息但不报错
				c.logger.Debug("task does not support context setting",
					logger.Field{Key: "task_id", Value: task.GetID()},
					logger.Field{Key: "task_type", Value: fmt.Sprintf("%T", task)},
				)
			}
		}

		// 发射任务开始事件
		taskStartEvent := NewTaskExecutionStartedEvent(i, task.GetDescription(), selectedAgent.GetRole())
		c.eventBus.Emit(ctx, c, taskStartEvent)

		// 执行任务
		start := time.Now()
		output, err := selectedAgent.Execute(ctx, task)
		duration := time.Since(start)

		if err != nil {
			c.logger.Error("task execution failed",
				logger.Field{Key: "task_index", Value: i},
				logger.Field{Key: "agent_role", Value: selectedAgent.GetRole()},
				logger.Field{Key: "error", Value: err},
				logger.Field{Key: "duration", Value: duration},
			)

			// 发射任务失败事件
			taskFailedEvent := NewTaskExecutionFailedEvent(i, task.GetDescription(), selectedAgent.GetRole(), err.Error(), duration)
			c.eventBus.Emit(ctx, c, taskFailedEvent)

			return nil, fmt.Errorf("task %d execution failed: %w", i, err)
		}

		// 执行任务回调
		if c.taskCallback != nil {
			if callbackErr := c.taskCallback(ctx, task, output); callbackErr != nil {
				c.logger.Error("task callback failed",
					logger.Field{Key: "task_index", Value: i},
					logger.Field{Key: "error", Value: callbackErr},
				)
			}
		}

		// 发射任务完成事件
		taskCompletedEvent := NewTaskExecutionCompletedEvent(i, task.GetDescription(), selectedAgent.GetRole(), duration, true)
		c.eventBus.Emit(ctx, c, taskCompletedEvent)

		c.logger.Info("task execution completed",
			logger.Field{Key: "task_index", Value: i},
			logger.Field{Key: "agent_role", Value: selectedAgent.GetRole()},
			logger.Field{Key: "duration", Value: duration},
		)

		// 存储输出
		tasksOutput = append(tasksOutput, output)
		lastOutput = output

		// 累积原始输出
		if output != nil && output.Raw != "" {
			if combinedRaw != "" {
				combinedRaw += "\n\n"
			}
			combinedRaw += output.Raw
		}

		// 检查上下文取消
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("execution cancelled: %w", ctx.Err())
		default:
			// 继续执行
		}
	}

	// 构建最终输出
	crewOutput := &CrewOutput{
		Raw:         combinedRaw,
		TasksOutput: tasksOutput,
		CreatedAt:   time.Now(),
		Success:     true,
		Metadata: map[string]interface{}{
			"process":      c.process.String(),
			"tasks_count":  len(tasks),
			"agents_count": len(c.agents),
		},
	}

	// 如果最后一个任务有JSON输出，使用它作为crew的JSON输出
	if lastOutput != nil && lastOutput.JSON != nil {
		crewOutput.JSON = lastOutput.JSON
	}

	// 如果最后一个任务有Pydantic输出，使用它作为crew的Pydantic输出
	if lastOutput != nil && lastOutput.Pydantic != nil {
		crewOutput.Pydantic = lastOutput.Pydantic
	}

	return crewOutput, nil
}

// selectAgentForTask 为任务选择合适的agent
// 优先级：任务指定Agent > Hierarchical管理器Agent > Sequential按序分配 > 智能匹配
func (c *BaseCrew) selectAgentForTask(task agent.Task, taskIndex int) (agent.Agent, error) {
	// 优先检查任务是否已经指定了Agent（类似Python版本中的task.agent）
	if taskAgent := c.getTaskAssignedAgent(task); taskAgent != nil {
		c.logger.Debug("using task-assigned agent",
			logger.Field{Key: "task_id", Value: task.GetID()},
			logger.Field{Key: "agent_role", Value: taskAgent.GetRole()},
		)
		return taskAgent, nil
	}

	// 如果是层级模式且有管理器agent，使用管理器进行委托
	if c.process == ProcessHierarchical && c.managerAgent != nil {
		c.logger.Debug("using manager agent for hierarchical delegation",
			logger.Field{Key: "task_id", Value: task.GetID()},
			logger.Field{Key: "manager_role", Value: c.managerAgent.GetRole()},
		)
		return c.managerAgent, nil
	}

	// 顺序模式：按索引分配agent
	if c.process == ProcessSequential {
		if len(c.agents) == 0 {
			return nil, fmt.Errorf("no agents available for sequential execution")
		}

		// 优先使用一对一映射
		if taskIndex < len(c.agents) {
			selectedAgent := c.agents[taskIndex]
			c.logger.Debug("sequential agent assignment (1:1)",
				logger.Field{Key: "task_index", Value: taskIndex},
				logger.Field{Key: "agent_index", Value: taskIndex},
				logger.Field{Key: "agent_role", Value: selectedAgent.GetRole()},
			)
			return selectedAgent, nil
		}

		// 如果任务数量超过agent数量，循环使用agent
		agentIndex := taskIndex % len(c.agents)
		selectedAgent := c.agents[agentIndex]
		c.logger.Debug("sequential agent assignment (cyclic)",
			logger.Field{Key: "task_index", Value: taskIndex},
			logger.Field{Key: "agent_index", Value: agentIndex},
			logger.Field{Key: "agent_role", Value: selectedAgent.GetRole()},
			logger.Field{Key: "total_agents", Value: len(c.agents)},
		)
		return selectedAgent, nil
	}

	// TODO: 实现更智能的agent选择逻辑
	// 可以基于：
	// 1. 任务类型与agent技能匹配
	// 2. Agent当前负载
	// 3. 任务优先级
	// 4. Agent专业领域
	// 5. Agent工具兼容性

	// 默认使用第一个agent
	if len(c.agents) > 0 {
		defaultAgent := c.agents[0]
		c.logger.Debug("using default (first) agent",
			logger.Field{Key: "task_id", Value: task.GetID()},
			logger.Field{Key: "agent_role", Value: defaultAgent.GetRole()},
		)
		return defaultAgent, nil
	}

	return nil, fmt.Errorf("no available agents for task %s", task.GetID())
}

// getTaskAssignedAgent 检查任务是否已指定Agent
// 这个方法用于模拟Python版本中的task.agent属性
func (c *BaseCrew) getTaskAssignedAgent(task agent.Task) agent.Agent {
	// 如果任务有预分配的agent，在这里检查
	// 目前BaseTask没有agent字段，这是一个扩展点
	// TODO: 可以考虑在Task接口中添加GetAssignedAgent()方法
	return nil
}

// prepareTaskContext 准备任务执行上下文
func (c *BaseCrew) prepareTaskContext(inputs map[string]interface{}, tasksOutput []*agent.TaskOutput, lastOutput *agent.TaskOutput) map[string]interface{} {
	context := make(map[string]interface{})

	// 添加初始输入
	for key, value := range inputs {
		context[key] = value
	}

	// 添加之前任务的输出作为上下文
	if len(tasksOutput) > 0 {
		context["previous_tasks_output"] = tasksOutput
	}

	// 添加最后一个任务的输出
	if lastOutput != nil {
		context["last_task_output"] = lastOutput.Raw
		if lastOutput.JSON != nil {
			context["last_task_json"] = lastOutput.JSON
		}
	}

	// 添加crew信息
	context["crew_name"] = c.name
	context["crew_process"] = c.process.String()
	context["total_tasks"] = len(c.tasks)
	context["completed_tasks"] = len(tasksOutput)

	return context
}

// createManagerAgent 创建或配置管理器agent
// 在Hierarchical模式中，管理器Agent负责协调和委托任务给其他Agent
func (c *BaseCrew) createManagerAgent() error {
	// 如果已经有管理器agent，配置它
	if c.managerAgent != nil {
		c.logger.Info("using provided manager agent",
			logger.Field{Key: "manager_role", Value: c.managerAgent.GetRole()},
			logger.Field{Key: "manager_id", Value: c.managerAgent.GetID()},
		)

		// 验证管理器agent配置
		if err := c.validateManagerAgent(c.managerAgent); err != nil {
			return fmt.Errorf("manager agent validation failed: %w", err)
		}

		// 确保管理器agent有对其他agents的访问能力
		if err := c.configureManagerAgentForDelegation(c.managerAgent); err != nil {
			return fmt.Errorf("failed to configure manager agent for delegation: %w", err)
		}

		return nil
	}

	// 如果没有管理器agent但有管理器LLM，创建默认管理器
	if c.managerLLM != nil {
		c.logger.Info("creating default manager agent with provided LLM")

		managerAgent, err := c.createDefaultManagerAgent()
		if err != nil {
			return fmt.Errorf("failed to create default manager agent: %w", err)
		}

		c.managerAgent = managerAgent
		c.logger.Info("default manager agent created successfully",
			logger.Field{Key: "manager_role", Value: managerAgent.GetRole()},
			logger.Field{Key: "manager_id", Value: managerAgent.GetID()},
		)

		return nil
	}

	return fmt.Errorf("hierarchical process requires either manager agent or manager LLM")
}

// validateManagerAgent 验证管理器Agent配置
func (c *BaseCrew) validateManagerAgent(manager agent.Agent) error {
	// 检查管理器agent是否有工具（在Python版本中，管理器不应该有工具）
	if tools := manager.GetTools(); len(tools) > 0 {
		c.logger.Warn("manager agent has tools, this may cause issues in hierarchical mode",
			logger.Field{Key: "manager_role", Value: manager.GetRole()},
			logger.Field{Key: "tools_count", Value: len(tools)},
		)
		// 注意：在Go版本中我们先警告但不强制移除工具，可以根据需要调整
	}

	// 验证管理器Agent的LLM是否配置
	if llm := manager.GetLLM(); llm == nil {
		return fmt.Errorf("manager agent must have an LLM configured")
	}

	return nil
}

// configureManagerAgentForDelegation 配置管理器Agent的委托能力
func (c *BaseCrew) configureManagerAgentForDelegation(manager agent.Agent) error {
	// 设置执行配置以允许委托
	config := manager.GetExecutionConfig()
	config.AllowDelegation = true

	if err := manager.SetExecutionConfig(config); err != nil {
		return fmt.Errorf("failed to set delegation config: %w", err)
	}

	// TODO: 创建AgentTools来管理其他agents
	// 这里需要实现类似Python版本中AgentTools的功能
	// agentTools := NewAgentTools(c.agents)
	// manager.SetTools(agentTools.GetTools())

	c.logger.Debug("manager agent configured for delegation",
		logger.Field{Key: "manager_role", Value: manager.GetRole()},
		logger.Field{Key: "available_agents", Value: len(c.agents)},
	)

	return nil
}

// createDefaultManagerAgent 创建默认的管理器Agent
func (c *BaseCrew) createDefaultManagerAgent() (agent.Agent, error) {
	// 创建管理器Agent的配置
	config := agent.AgentConfig{
		Role:      "Crew Manager",
		Goal:      "Coordinate and delegate tasks among team members to achieve the crew's objectives efficiently",
		Backstory: "You are an experienced project manager skilled in coordinating teams, delegating tasks, and ensuring quality deliverables. You understand each team member's strengths and assign tasks accordingly.",
		LLM:       c.managerLLM.(llm.LLM), // 将interface{}转换为llm.LLM
		Tools:     make([]agent.Tool, 0),  // 管理器开始时没有工具
		ExecutionConfig: agent.ExecutionConfig{
			MaxIterations:   25,
			AllowDelegation: true,
			VerboseLogging:  c.verbose,
			HumanInput:      false,
		},
		EventBus:       c.eventBus,
		Logger:         c.logger,
		SecurityConfig: c.securityConfig,
	}

	// 创建管理器Agent
	managerAgent, err := agent.NewBaseAgent(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create manager agent: %w", err)
	}

	// 初始化管理器Agent
	if err := managerAgent.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize manager agent: %w", err)
	}

	// TODO: 添加AgentTools以管理其他agents
	// agentTools := NewAgentTools(c.agents)
	// for _, tool := range agentTools.GetTools() {
	//     managerAgent.AddTool(tool)
	// }

	return managerAgent, nil
}
