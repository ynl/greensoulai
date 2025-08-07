package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ynl/greensoulai/internal/llm"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
	"github.com/ynl/greensoulai/pkg/security"
)

// BaseAgent 实现了Agent接口的基础结构
type BaseAgent struct {
	// 基础属性
	id        string
	role      string
	goal      string
	backstory string

	// 核心组件
	llmProvider       llm.LLM
	tools             []Tool
	memory            Memory
	knowledgeSources  []KnowledgeSource
	humanInputHandler HumanInputHandler

	// 配置
	executionConfig ExecutionConfig
	securityConfig  security.SecurityConfig

	// 系统组件
	eventBus events.EventBus
	logger   logger.Logger

	// 模板和配置
	systemTemplate string
	promptTemplate string
	callbacks      []func(context.Context, *TaskOutput) error

	// 统计和状态
	stats         ExecutionStats
	isInitialized bool
	mu            sync.RWMutex

	// 私有状态
	timesExecuted     int
	lastExecutionTime time.Time
}

// NewBaseAgent 创建新的BaseAgent实例
func NewBaseAgent(config AgentConfig) (*BaseAgent, error) {
	if config.Role == "" {
		return nil, fmt.Errorf("agent role is required")
	}
	if config.Goal == "" {
		return nil, fmt.Errorf("agent goal is required")
	}
	if config.Backstory == "" {
		return nil, fmt.Errorf("agent backstory is required")
	}

	// 使用提供的组件或创建默认组件
	agentLogger := config.Logger
	if agentLogger == nil {
		agentLogger = logger.NewConsoleLogger()
	}

	secConfig := config.SecurityConfig
	if secConfig.Fingerprint == nil {
		secConfig = *security.NewSecurityConfig()
	}

	execConfig := config.ExecutionConfig
	if execConfig.MaxIterations == 0 {
		execConfig = DefaultExecutionConfig()
	}

	agent := &BaseAgent{
		id:                uuid.New().String(),
		role:              config.Role,
		goal:              config.Goal,
		backstory:         config.Backstory,
		llmProvider:       config.LLM,
		tools:             config.Tools,
		memory:            config.Memory,
		knowledgeSources:  config.KnowledgeSources,
		humanInputHandler: config.HumanInputHandler,
		executionConfig:   execConfig,
		securityConfig:    secConfig,
		eventBus:          config.EventBus,
		logger:            agentLogger,
		systemTemplate:    config.SystemTemplate,
		promptTemplate:    config.PromptTemplate,
		callbacks:         config.Callbacks,
		stats: ExecutionStats{
			TotalExecutions:      0,
			SuccessfulExecutions: 0,
			FailedExecutions:     0,
			TotalExecutionTime:   0,
			AverageExecutionTime: 0,
			TokensUsed:           0,
			TotalCost:            0,
			ToolsUsed:            make(map[string]int),
			CreatedAt:            time.Now(),
		},
		isInitialized:     false,
		timesExecuted:     0,
		lastExecutionTime: time.Time{},
	}

	return agent, nil
}

// Initialize 初始化Agent
func (a *BaseAgent) Initialize() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isInitialized {
		return nil
	}

	// 验证必要组件
	if a.llmProvider == nil {
		return fmt.Errorf("LLM provider is required")
	}

	// 设置事件总线到LLM（如果LLM支持）
	if a.eventBus != nil {
		// Note: LLM模块有自己的事件总线集成，这里不需要设置
		a.logger.Debug("Event bus available for agent",
			logger.Field{Key: "agent_id", Value: a.id},
		)
	}

	// 初始化知识源
	for i, source := range a.knowledgeSources {
		if err := source.Initialize(); err != nil {
			a.logger.Error("Failed to initialize knowledge source",
				logger.Field{Key: "index", Value: i},
				logger.Field{Key: "name", Value: source.GetName()},
				logger.Field{Key: "error", Value: err},
			)
			return fmt.Errorf("failed to initialize knowledge source %d: %w", i, err)
		}
	}

	a.isInitialized = true
	a.logger.Info("Agent initialized successfully",
		logger.Field{Key: "id", Value: a.id},
		logger.Field{Key: "role", Value: a.role},
		logger.Field{Key: "tools_count", Value: len(a.tools)},
		logger.Field{Key: "knowledge_sources_count", Value: len(a.knowledgeSources)},
	)

	return nil
}

// Execute 执行任务的核心方法
func (a *BaseAgent) Execute(ctx context.Context, task Task) (*TaskOutput, error) {
	if err := a.ensureInitialized(); err != nil {
		return nil, fmt.Errorf("agent initialization failed: %w", err)
	}

	// 更新执行统计
	a.mu.Lock()
	a.timesExecuted++
	executionID := a.timesExecuted
	a.lastExecutionTime = time.Now()
	a.mu.Unlock()

	startTime := time.Now()

	// 发射开始事件
	if a.eventBus != nil {
		startEvent := NewAgentExecutionStartedEvent(
			a.id,
			a.role,
			task.GetID(),
			task.GetDescription(),
			executionID,
		)
		a.eventBus.Emit(ctx, a, startEvent)
	}

	a.logger.Info("Starting task execution",
		logger.Field{Key: "agent", Value: a.role},
		logger.Field{Key: "task_id", Value: task.GetID()},
		logger.Field{Key: "execution_id", Value: executionID},
	)

	// 检查人工输入需求
	if task.IsHumanInputRequired() {
		if err := a.handleHumanInput(ctx, task); err != nil {
			return nil, fmt.Errorf("human input handling failed: %w", err)
		}
	}

	// 执行核心任务逻辑
	output, err := a.executeCore(ctx, task)
	duration := time.Since(startTime)

	// 更新统计信息
	a.updateStats(output, err, duration)

	// 发射完成事件
	if a.eventBus != nil {
		completedEvent := NewAgentExecutionCompletedEvent(
			a.id,
			a.role,
			task.GetID(),
			task.GetDescription(),
			executionID,
			duration,
			err == nil,
			output,
		)
		a.eventBus.Emit(ctx, a, completedEvent)
	}

	if err != nil {
		a.logger.Error("Task execution failed",
			logger.Field{Key: "agent", Value: a.role},
			logger.Field{Key: "task_id", Value: task.GetID()},
			logger.Field{Key: "execution_id", Value: executionID},
			logger.Field{Key: "duration", Value: duration},
			logger.Field{Key: "error", Value: err},
		)
		return nil, err
	}

	a.logger.Info("Task execution completed successfully",
		logger.Field{Key: "agent", Value: a.role},
		logger.Field{Key: "task_id", Value: task.GetID()},
		logger.Field{Key: "execution_id", Value: executionID},
		logger.Field{Key: "duration", Value: duration},
		logger.Field{Key: "tokens_used", Value: output.TokensUsed},
		logger.Field{Key: "cost", Value: output.Cost},
	)

	return output, nil
}

// ExecuteAsync 异步执行任务
func (a *BaseAgent) ExecuteAsync(ctx context.Context, task Task) (<-chan TaskResult, error) {
	resultChan := make(chan TaskResult, 1)

	go func() {
		defer close(resultChan)
		output, err := a.Execute(ctx, task)
		resultChan <- TaskResult{Output: output, Error: err}
	}()

	return resultChan, nil
}

// ExecuteWithTimeout 带超时执行任务
func (a *BaseAgent) ExecuteWithTimeout(ctx context.Context, task Task, timeout time.Duration) (*TaskOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 使用通道来处理超时
	resultChan := make(chan TaskResult, 1)

	go func() {
		output, err := a.Execute(ctx, task)
		select {
		case resultChan <- TaskResult{Output: output, Error: err}:
		case <-ctx.Done():
			// Context已取消，不发送结果
		}
	}()

	select {
	case result := <-resultChan:
		return result.Output, result.Error
	case <-ctx.Done():
		return nil, fmt.Errorf("task execution timeout after %v: %w", timeout, ctx.Err())
	}
}

// executeCore 执行任务的核心逻辑
func (a *BaseAgent) executeCore(ctx context.Context, task Task) (*TaskOutput, error) {
	// 构建任务提示
	prompt, err := a.buildTaskPrompt(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to build task prompt: %w", err)
	}

	// 准备LLM消息
	messages := a.buildMessages(prompt)

	// 调用LLM
	callOptions := a.buildLLMCallOptions()
	response, err := a.llmProvider.Call(ctx, messages, callOptions)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	// 处理响应并构建输出
	output := a.buildTaskOutput(task, response)

	// 执行回调
	if err := a.executeCallbacks(ctx, output); err != nil {
		a.logger.Error("Callback execution failed",
			logger.Field{Key: "error", Value: err},
		)
		// 回调失败不应该阻止任务完成
	}

	return output, nil
}

// buildTaskPrompt 构建任务提示
func (a *BaseAgent) buildTaskPrompt(ctx context.Context, task Task) (string, error) {
	prompt := task.GetDescription()

	// 添加期望输出
	if expectedOutput := task.GetExpectedOutput(); expectedOutput != "" {
		prompt += fmt.Sprintf("\n\nExpected Output: %s", expectedOutput)
	}

	// 添加人工输入（如果有）
	if task.IsHumanInputRequired() && task.GetHumanInput() != "" {
		prompt += fmt.Sprintf("\n\nHuman Input: %s", task.GetHumanInput())
	}

	// 添加可用工具信息
	if len(a.tools) > 0 {
		toolsDesc := a.buildToolsDescription()
		prompt += fmt.Sprintf("\n\nAvailable Tools:\n%s", toolsDesc)
	}

	// 查询记忆系统
	if a.memory != nil {
		memoryContext, err := a.queryMemory(ctx, task)
		if err != nil {
			a.logger.Warn("Failed to query memory",
				logger.Field{Key: "error", Value: err},
			)
		} else if memoryContext != "" {
			prompt += fmt.Sprintf("\n\nRelevant Memory:\n%s", memoryContext)
		}
	}

	// 查询知识源
	if len(a.knowledgeSources) > 0 {
		knowledgeContext, err := a.queryKnowledge(ctx, task)
		if err != nil {
			a.logger.Warn("Failed to query knowledge sources",
				logger.Field{Key: "error", Value: err},
			)
		} else if knowledgeContext != "" {
			prompt += fmt.Sprintf("\n\nRelevant Knowledge:\n%s", knowledgeContext)
		}
	}

	return prompt, nil
}

// buildMessages 构建LLM消息
func (a *BaseAgent) buildMessages(prompt string) []llm.Message {
	messages := []llm.Message{}

	// 系统消息
	if a.executionConfig.UseSystemPrompt {
		systemPrompt := a.buildSystemPrompt()
		messages = append(messages, llm.Message{
			Role:    llm.RoleSystem,
			Content: systemPrompt,
		})
	}

	// 用户消息
	messages = append(messages, llm.Message{
		Role:    llm.RoleUser,
		Content: prompt,
	})

	return messages
}

// buildSystemPrompt 构建系统提示
func (a *BaseAgent) buildSystemPrompt() string {
	if a.systemTemplate != "" {
		// 使用自定义模板
		template := a.systemTemplate
		template = strings.ReplaceAll(template, "{role}", a.role)
		template = strings.ReplaceAll(template, "{goal}", a.goal)
		template = strings.ReplaceAll(template, "{backstory}", a.backstory)
		return template
	}

	// 默认系统提示
	return fmt.Sprintf(`You are %s.

Your goal: %s

Your backstory: %s

You are working with a team of other agents to complete complex tasks. Always provide detailed, accurate responses based on your role and expertise. Use the available tools when necessary and be precise in your reasoning.`,
		a.role, a.goal, a.backstory)
}

// buildToolsDescription 构建工具描述
func (a *BaseAgent) buildToolsDescription() string {
	var descriptions []string
	for _, tool := range a.tools {
		desc := fmt.Sprintf("- %s: %s", tool.GetName(), tool.GetDescription())
		descriptions = append(descriptions, desc)
	}
	return strings.Join(descriptions, "\n")
}

// buildLLMCallOptions 构建LLM调用选项
func (a *BaseAgent) buildLLMCallOptions() *llm.CallOptions {
	options := &llm.CallOptions{}

	if a.executionConfig.MaxTokens > 0 {
		maxTokens := a.executionConfig.MaxTokens
		options.MaxTokens = &maxTokens
	}

	if a.executionConfig.Temperature > 0 {
		temp := a.executionConfig.Temperature
		options.Temperature = &temp
	}

	if a.executionConfig.Timeout > 0 {
		// Note: 这里可能需要在LLM接口中添加超时支持
	}

	return options
}

// buildTaskOutput 构建任务输出
func (a *BaseAgent) buildTaskOutput(task Task, response *llm.Response) *TaskOutput {
	output := &TaskOutput{
		Raw:            response.Content,
		Agent:          a.role,
		Task:           task.GetID(),
		Description:    task.GetDescription(),
		ExpectedOutput: task.GetExpectedOutput(),
		OutputFormat:   task.GetOutputFormat(),
		ExecutionTime:  0, // 将在上层设置
		CreatedAt:      time.Now(),
		TokensUsed:     response.Usage.TotalTokens,
		Cost:           response.Usage.Cost,
		Model:          response.Model,
		IsValid:        true,       // 默认有效，可以后续添加验证逻辑
		ToolsUsed:      []string{}, // TODO: 从响应中提取工具使用信息
		Metadata:       make(map[string]interface{}),
	}

	// 尝试解析JSON输出
	if task.GetOutputFormat() == OutputFormatJSON ||
		(strings.Contains(response.Content, "{") && strings.Contains(response.Content, "}")) {
		// 尝试解析为JSON
		var jsonData map[string]interface{}
		if err := json.Unmarshal([]byte(response.Content), &jsonData); err == nil {
			output.JSON = jsonData
			// 如果任务期望JSON格式，更新输出格式
			if task.GetOutputFormat() == OutputFormatJSON {
				output.OutputFormat = OutputFormatJSON
			}
		}
	}

	// 生成摘要
	output.Summary = a.generateSummary(response.Content)

	// 添加元数据
	output.Metadata["finish_reason"] = response.FinishReason
	output.Metadata["prompt_tokens"] = response.Usage.PromptTokens
	output.Metadata["completion_tokens"] = response.Usage.CompletionTokens
	output.Metadata["agent_id"] = a.id

	return output
}

// generateSummary 生成输出摘要
func (a *BaseAgent) generateSummary(content string) string {
	words := strings.Fields(content)
	if len(words) <= 15 {
		return content
	}
	return strings.Join(words[:15], " ") + "..."
}

// handleHumanInput 处理人工输入
func (a *BaseAgent) handleHumanInput(ctx context.Context, task Task) error {
	if a.humanInputHandler == nil {
		return fmt.Errorf("human input required but no handler configured")
	}

	prompt := fmt.Sprintf("Task requires your input: %s", task.GetDescription())
	input, err := a.humanInputHandler.RequestInput(ctx, prompt, nil)
	if err != nil {
		return fmt.Errorf("human input request failed: %w", err)
	}

	task.SetHumanInput(input)
	a.logger.Info("Received human input",
		logger.Field{Key: "task_id", Value: task.GetID()},
		logger.Field{Key: "input_length", Value: len(input)},
	)

	return nil
}

// executeCallbacks 执行回调函数
func (a *BaseAgent) executeCallbacks(ctx context.Context, output *TaskOutput) error {
	for i, callback := range a.callbacks {
		if err := callback(ctx, output); err != nil {
			return fmt.Errorf("callback %d failed: %w", i, err)
		}
	}
	return nil
}

// queryMemory 查询记忆系统
func (a *BaseAgent) queryMemory(ctx context.Context, task Task) (string, error) {
	if a.memory == nil {
		return "", nil
	}

	// 使用任务描述作为查询
	query := task.GetDescription()
	items, err := a.memory.Search(ctx, query, 5) // 获取最相关的5个记忆
	if err != nil {
		return "", err
	}

	if len(items) == 0 {
		return "", nil
	}

	var contexts []string
	for _, item := range items {
		if str, ok := item.Value.(string); ok {
			contexts = append(contexts, str)
		}
	}

	return strings.Join(contexts, "\n"), nil
}

// queryKnowledge 查询知识源
func (a *BaseAgent) queryKnowledge(ctx context.Context, task Task) (string, error) {
	if len(a.knowledgeSources) == 0 {
		return "", nil
	}

	query := task.GetDescription()
	options := DefaultQueryOptions()
	options.Limit = 3 // 每个知识源获取3个结果

	var allKnowledge []string
	for _, source := range a.knowledgeSources {
		items, err := source.Query(ctx, query, options)
		if err != nil {
			a.logger.Warn("Knowledge source query failed",
				logger.Field{Key: "source", Value: source.GetName()},
				logger.Field{Key: "error", Value: err},
			)
			continue
		}

		for _, item := range items {
			allKnowledge = append(allKnowledge,
				fmt.Sprintf("[%s] %s", source.GetName(), item.Content))
		}
	}

	return strings.Join(allKnowledge, "\n"), nil
}

// updateStats 更新执行统计
func (a *BaseAgent) updateStats(output *TaskOutput, err error, duration time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.stats.TotalExecutions++
	if err == nil {
		a.stats.SuccessfulExecutions++
		if output != nil {
			a.stats.TokensUsed += output.TokensUsed
			a.stats.TotalCost += output.Cost

			// 更新工具使用统计
			for _, tool := range output.ToolsUsed {
				a.stats.ToolsUsed[tool]++
			}
		}
	} else {
		a.stats.FailedExecutions++
	}

	a.stats.TotalExecutionTime += duration
	a.stats.AverageExecutionTime = a.stats.TotalExecutionTime / time.Duration(a.stats.TotalExecutions)
	a.stats.LastExecutionTime = time.Now()

	if output != nil {
		output.ExecutionTime = duration
	}
}

// ensureInitialized 确保Agent已初始化
func (a *BaseAgent) ensureInitialized() error {
	a.mu.RLock()
	initialized := a.isInitialized
	a.mu.RUnlock()

	if !initialized {
		return a.Initialize()
	}
	return nil
}

// Getter方法实现
func (a *BaseAgent) GetID() string {
	return a.id
}

func (a *BaseAgent) GetRole() string {
	return a.role
}

func (a *BaseAgent) GetGoal() string {
	return a.goal
}

func (a *BaseAgent) GetBackstory() string {
	return a.backstory
}

func (a *BaseAgent) GetLLM() llm.LLM {
	return a.llmProvider
}

func (a *BaseAgent) GetTools() []Tool {
	return a.tools
}

func (a *BaseAgent) GetMemory() Memory {
	return a.memory
}

func (a *BaseAgent) GetKnowledgeSources() []KnowledgeSource {
	return a.knowledgeSources
}

func (a *BaseAgent) GetExecutionConfig() ExecutionConfig {
	return a.executionConfig
}

func (a *BaseAgent) GetHumanInputHandler() HumanInputHandler {
	return a.humanInputHandler
}

func (a *BaseAgent) GetEventBus() events.EventBus {
	return a.eventBus
}

func (a *BaseAgent) GetLogger() logger.Logger {
	return a.logger
}

func (a *BaseAgent) GetExecutionStats() ExecutionStats {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.stats
}

// Setter方法实现
func (a *BaseAgent) SetLLM(llmProvider llm.LLM) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.llmProvider = llmProvider
	return nil
}

func (a *BaseAgent) AddTool(tool Tool) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.tools = append(a.tools, tool)
	a.logger.Info("Tool added to agent",
		logger.Field{Key: "agent", Value: a.role},
		logger.Field{Key: "tool", Value: tool.GetName()},
	)
	return nil
}

func (a *BaseAgent) SetMemory(memory Memory) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.memory = memory
	return nil
}

func (a *BaseAgent) SetKnowledgeSources(sources []KnowledgeSource) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.knowledgeSources = sources
	return nil
}

func (a *BaseAgent) SetExecutionConfig(config ExecutionConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.executionConfig = config
	return nil
}

func (a *BaseAgent) SetHumanInputHandler(handler HumanInputHandler) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.humanInputHandler = handler
	return nil
}

func (a *BaseAgent) SetEventBus(eventBus events.EventBus) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.eventBus = eventBus
	return nil
}

func (a *BaseAgent) SetLogger(logger logger.Logger) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.logger = logger
	return nil
}

func (a *BaseAgent) ResetStats() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.stats = ExecutionStats{
		TotalExecutions:      0,
		SuccessfulExecutions: 0,
		FailedExecutions:     0,
		TotalExecutionTime:   0,
		AverageExecutionTime: 0,
		TokensUsed:           0,
		TotalCost:            0,
		ToolsUsed:            make(map[string]int),
		CreatedAt:            time.Now(),
	}

	a.timesExecuted = 0
	a.lastExecutionTime = time.Time{}
	return nil
}

// Close 清理资源
func (a *BaseAgent) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 关闭知识源
	for _, source := range a.knowledgeSources {
		if err := source.Close(); err != nil {
			a.logger.Error("Failed to close knowledge source",
				logger.Field{Key: "source", Value: source.GetName()},
				logger.Field{Key: "error", Value: err},
			)
		}
	}

	// 关闭LLM（如果支持）
	if a.llmProvider != nil {
		if err := a.llmProvider.Close(); err != nil {
			a.logger.Error("Failed to close LLM provider",
				logger.Field{Key: "error", Value: err},
			)
		}
	}

	a.isInitialized = false
	return nil
}

// Clone 创建Agent的副本
func (a *BaseAgent) Clone() Agent {
	a.mu.RLock()
	defer a.mu.RUnlock()

	config := AgentConfig{
		Role:              a.role,
		Goal:              a.goal,
		Backstory:         a.backstory,
		LLM:               a.llmProvider,
		Tools:             make([]Tool, len(a.tools)),
		ExecutionConfig:   a.executionConfig,
		Memory:            a.memory,
		KnowledgeSources:  make([]KnowledgeSource, len(a.knowledgeSources)),
		HumanInputHandler: a.humanInputHandler,
		EventBus:          a.eventBus,
		Logger:            a.logger,
		SecurityConfig:    a.securityConfig,
		SystemTemplate:    a.systemTemplate,
		PromptTemplate:    a.promptTemplate,
		Callbacks:         make([]func(context.Context, *TaskOutput) error, len(a.callbacks)),
	}

	// 复制切片
	copy(config.Tools, a.tools)
	copy(config.KnowledgeSources, a.knowledgeSources)
	copy(config.Callbacks, a.callbacks)

	clonedAgent, _ := NewBaseAgent(config)
	return clonedAgent
}
