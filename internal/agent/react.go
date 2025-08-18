package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ynl/greensoulai/internal/llm"
)

// AgentMode 定义Agent的执行模式
type AgentMode int

const (
	// ModeJSON 默认JSON格式模式，保持现有行为
	ModeJSON AgentMode = iota
	// ModeReAct ReAct结构化推理模式
	ModeReAct
	// ModeHybrid 混合模式，根据任务复杂度自动选择
	ModeHybrid
)

// String 返回模式的字符串表示
func (m AgentMode) String() string {
	switch m {
	case ModeJSON:
		return "json"
	case ModeReAct:
		return "react"
	case ModeHybrid:
		return "hybrid"
	default:
		return "unknown"
	}
}

// ReActConfig ReAct模式的配置
type ReActConfig struct {
	// MaxIterations 最大推理迭代次数
	MaxIterations int `json:"max_iterations"`

	// ThoughtTimeout 单次思考的超时时间
	ThoughtTimeout time.Duration `json:"thought_timeout"`

	// EnableDebugOutput 是否启用调试输出
	EnableDebugOutput bool `json:"enable_debug_output"`

	// CustomPromptTemplate 自定义提示词模板
	CustomPromptTemplate string `json:"custom_prompt_template,omitempty"`

	// StrictFormatValidation 是否启用严格格式验证
	StrictFormatValidation bool `json:"strict_format_validation"`

	// AllowFallbackToJSON 当ReAct格式解析失败时，是否允许回退到JSON模式
	AllowFallbackToJSON bool `json:"allow_fallback_to_json"`
}

// DefaultReActConfig 返回默认的ReAct配置
func DefaultReActConfig() *ReActConfig {
	return &ReActConfig{
		MaxIterations:          10,
		ThoughtTimeout:         30 * time.Second,
		EnableDebugOutput:      false,
		StrictFormatValidation: true,
		AllowFallbackToJSON:    true,
	}
}

// ReActStep 表示ReAct推理过程中的一个步骤
type ReActStep struct {
	// StepID 步骤唯一标识符
	StepID string `json:"step_id"`

	// Thought 思考内容
	Thought string `json:"thought"`

	// Action 要执行的动作/工具名称
	Action string `json:"action,omitempty"`

	// ActionInput 动作的输入参数
	ActionInput map[string]interface{} `json:"action_input,omitempty"`

	// Observation 动作执行后的观察结果
	Observation string `json:"observation,omitempty"`

	// FinalAnswer 最终答案（如果有）
	FinalAnswer string `json:"final_answer,omitempty"`

	// IsComplete 是否为完成步骤
	IsComplete bool `json:"is_complete"`

	// Timestamp 步骤时间戳
	Timestamp time.Time `json:"timestamp"`

	// Duration 步骤执行时长
	Duration time.Duration `json:"duration"`

	// Error 执行错误（如果有）
	Error string `json:"error,omitempty"`
}

// ReActTrace 表示完整的ReAct推理轨迹
type ReActTrace struct {
	// TraceID 轨迹唯一标识符
	TraceID string `json:"trace_id"`

	// Steps 推理步骤序列
	Steps []*ReActStep `json:"steps"`

	// StartTime 开始时间
	StartTime time.Time `json:"start_time"`

	// EndTime 结束时间
	EndTime time.Time `json:"end_time"`

	// TotalDuration 总耗时
	TotalDuration time.Duration `json:"total_duration"`

	// IsCompleted 是否已完成
	IsCompleted bool `json:"is_completed"`

	// FinalOutput 最终输出
	FinalOutput string `json:"final_output"`

	// IterationCount 实际迭代次数
	IterationCount int `json:"iteration_count"`
}

// ReActParser ReAct格式解析器接口
type ReActParser interface {
	// Parse 解析LLM输出为ReAct步骤
	Parse(ctx context.Context, output string) (*ReActStep, error)

	// Validate 验证ReAct步骤的格式是否正确
	Validate(step *ReActStep) error

	// Format 格式化ReAct步骤为字符串
	Format(step *ReActStep) string
}

// ReActExecutor ReAct执行器接口
type ReActExecutor interface {
	// ExecuteReAct 执行ReAct推理流程
	ExecuteReAct(ctx context.Context, agent Agent, task Task) (*ReActTrace, error)

	// ExecuteStep 执行单个ReAct步骤
	ExecuteStep(ctx context.Context, agent Agent, step *ReActStep, toolCtx *ToolExecutionContext) error

	// ShouldContinue 判断是否应该继续推理
	ShouldContinue(trace *ReActTrace, config *ReActConfig) bool
}

// ReActAgent ReAct模式Agent的扩展接口
type ReActAgent interface {
	Agent

	// SetReActMode 设置是否启用ReAct模式
	SetReActMode(enabled bool)

	// GetReActMode 获取当前是否启用ReAct模式
	GetReActMode() bool

	// SetReActConfig 设置ReAct配置
	SetReActConfig(config *ReActConfig) error

	// GetReActConfig 获取ReAct配置
	GetReActConfig() *ReActConfig

	// GetReActTrace 获取最近的ReAct推理轨迹
	GetReActTrace() *ReActTrace

	// ExecuteWithReAct 使用ReAct模式执行任务
	ExecuteWithReAct(ctx context.Context, task Task) (*TaskOutput, *ReActTrace, error)
}

// StandardReActParser 标准ReAct格式解析器实现
type StandardReActParser struct {
	// 正则表达式模式
	thoughtPattern     *regexp.Regexp
	actionPattern      *regexp.Regexp
	actionInputPattern *regexp.Regexp
	observationPattern *regexp.Regexp
	finalAnswerPattern *regexp.Regexp
}

// NewStandardReActParser 创建标准ReAct解析器
func NewStandardReActParser() *StandardReActParser {
	return &StandardReActParser{
		thoughtPattern:     regexp.MustCompile(`(?i)thought:\s*(.+?)(?:\n|$)`),
		actionPattern:      regexp.MustCompile(`(?i)action:\s*(.+?)(?:\n|$)`),
		actionInputPattern: regexp.MustCompile(`(?i)action\s+input:\s*(\{.*?\})`),
		observationPattern: regexp.MustCompile(`(?i)observation:\s*(.+?)(?:\n|$)`),
		finalAnswerPattern: regexp.MustCompile(`(?i)final\s+answer:\s*(.+)`),
	}
}

// Parse 实现ReActParser接口
func (p *StandardReActParser) Parse(ctx context.Context, output string) (*ReActStep, error) {
	step := &ReActStep{
		StepID:    fmt.Sprintf("step_%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
	}

	// 解析Thought
	if matches := p.thoughtPattern.FindStringSubmatch(output); len(matches) > 1 {
		step.Thought = strings.TrimSpace(matches[1])
	}

	// 检查是否有Final Answer
	if matches := p.finalAnswerPattern.FindStringSubmatch(output); len(matches) > 1 {
		step.FinalAnswer = strings.TrimSpace(matches[1])
		step.IsComplete = true
		return step, nil
	}

	// 解析Action
	if matches := p.actionPattern.FindStringSubmatch(output); len(matches) > 1 {
		step.Action = strings.TrimSpace(matches[1])
	}

	// 解析Action Input
	if matches := p.actionInputPattern.FindStringSubmatch(output); len(matches) > 1 {
		var actionInput map[string]interface{}
		if err := json.Unmarshal([]byte(matches[1]), &actionInput); err != nil {
			return nil, fmt.Errorf("failed to parse action input JSON: %w", err)
		}
		step.ActionInput = actionInput
	}

	// 解析Observation
	if matches := p.observationPattern.FindStringSubmatch(output); len(matches) > 1 {
		step.Observation = strings.TrimSpace(matches[1])
	}

	return step, nil
}

// Validate 验证ReAct步骤格式
func (p *StandardReActParser) Validate(step *ReActStep) error {
	if step == nil {
		return fmt.Errorf("step cannot be nil")
	}

	// 完成步骤必须有FinalAnswer
	if step.IsComplete {
		if step.FinalAnswer == "" {
			return fmt.Errorf("complete step must have final_answer")
		}
		return nil
	}

	// 非完成步骤必须有Thought
	if step.Thought == "" {
		return fmt.Errorf("step must have thought")
	}

	// 如果有Action，必须有ActionInput
	if step.Action != "" && len(step.ActionInput) == 0 {
		return fmt.Errorf("action step must have action_input")
	}

	return nil
}

// Format 格式化ReAct步骤为字符串
func (p *StandardReActParser) Format(step *ReActStep) string {
	var parts []string

	if step.Thought != "" {
		parts = append(parts, fmt.Sprintf("Thought: %s", step.Thought))
	}

	if step.IsComplete && step.FinalAnswer != "" {
		parts = append(parts, fmt.Sprintf("Final Answer: %s", step.FinalAnswer))
		return strings.Join(parts, "\n")
	}

	if step.Action != "" {
		parts = append(parts, fmt.Sprintf("Action: %s", step.Action))
	}

	if len(step.ActionInput) > 0 {
		if inputJSON, err := json.Marshal(step.ActionInput); err == nil {
			parts = append(parts, fmt.Sprintf("Action Input: %s", string(inputJSON)))
		}
	}

	if step.Observation != "" {
		parts = append(parts, fmt.Sprintf("Observation: %s", step.Observation))
	}

	return strings.Join(parts, "\n")
}

// NewReActTrace 创建新的ReAct轨迹
func NewReActTrace() *ReActTrace {
	return &ReActTrace{
		TraceID:   fmt.Sprintf("trace_%d", time.Now().UnixNano()),
		Steps:     make([]*ReActStep, 0),
		StartTime: time.Now(),
	}
}

// AddStep 添加步骤到轨迹
func (t *ReActTrace) AddStep(step *ReActStep) {
	t.Steps = append(t.Steps, step)
	t.IterationCount = len(t.Steps)

	if step.IsComplete {
		t.IsCompleted = true
		t.FinalOutput = step.FinalAnswer
		t.EndTime = time.Now()
		t.TotalDuration = t.EndTime.Sub(t.StartTime)
	}
}

// GetLastStep 获取最后一个步骤
func (t *ReActTrace) GetLastStep() *ReActStep {
	if len(t.Steps) == 0 {
		return nil
	}
	return t.Steps[len(t.Steps)-1]
}

// HasCompletedStep 检查是否有完成步骤
func (t *ReActTrace) HasCompletedStep() bool {
	for _, step := range t.Steps {
		if step.IsComplete {
			return true
		}
	}
	return false
}

// StandardReActExecutor 标准ReAct执行器实现
type StandardReActExecutor struct {
	parser ReActParser
}

// NewStandardReActExecutor 创建标准ReAct执行器
func NewStandardReActExecutor() *StandardReActExecutor {
	return &StandardReActExecutor{
		parser: NewStandardReActParser(),
	}
}

// ExecuteReAct 实现ReActExecutor接口
func (e *StandardReActExecutor) ExecuteReAct(ctx context.Context, agent Agent, task Task) (*ReActTrace, error) {
	// 获取ReAct配置
	config := DefaultReActConfig()
	if reactAgent, ok := agent.(ReActAgent); ok {
		if reactConfig := reactAgent.GetReActConfig(); reactConfig != nil {
			config = reactConfig
		}
	}

	// 创建轨迹
	trace := NewReActTrace()

	// 创建工具执行上下文
	toolCtx := NewToolExecutionContext(agent, task)

	// 构建初始提示
	initialPrompt, err := e.buildReActPrompt(task, toolCtx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build initial prompt: %w", err)
	}

	// 执行ReAct循环
	for trace.IterationCount < config.MaxIterations && !trace.IsCompleted {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return trace, ctx.Err()
		default:
		}

		// 调用LLM
		response, err := e.callLLM(ctx, agent, initialPrompt, trace)
		if err != nil {
			return trace, fmt.Errorf("LLM call failed at iteration %d: %w", trace.IterationCount, err)
		}

		// 解析响应
		step, err := e.parser.Parse(ctx, response)
		if err != nil {
			// 如果解析失败且允许回退，尝试作为普通响应处理
			if config.AllowFallbackToJSON {
				step = &ReActStep{
					StepID:      fmt.Sprintf("fallback_%d", trace.IterationCount),
					FinalAnswer: response,
					IsComplete:  true,
					Timestamp:   time.Now(),
					Error:       fmt.Sprintf("Parse failed, fallback to direct response: %v", err),
				}
			} else {
				return trace, fmt.Errorf("failed to parse response at iteration %d: %w", trace.IterationCount, err)
			}
		}

		// 验证步骤
		if err := e.parser.Validate(step); err != nil {
			step.Error = fmt.Sprintf("Validation failed: %v", err)
		}

		// 执行步骤（如果有动作）
		if step.Action != "" && step.Error == "" {
			if err := e.ExecuteStep(ctx, agent, step, toolCtx); err != nil {
				step.Error = err.Error()
			}
		}

		// 添加步骤到轨迹
		trace.AddStep(step)

		// 如果步骤完成或有错误，跳出循环
		if step.IsComplete || step.Error != "" {
			break
		}

		// 更新提示以包含新的观察结果
		initialPrompt = e.updatePromptWithStep(initialPrompt, step)
	}

	// 如果达到最大迭代次数但未完成，创建一个强制完成步骤
	if !trace.IsCompleted && trace.IterationCount >= config.MaxIterations {
		finalStep := &ReActStep{
			StepID:      fmt.Sprintf("force_final_%d", time.Now().UnixNano()),
			Thought:     "Reached maximum iterations, providing best answer available",
			FinalAnswer: "Task execution reached maximum iterations. Unable to complete within the allowed time.",
			IsComplete:  true,
			Timestamp:   time.Now(),
			Error:       "Reached maximum iterations",
		}
		trace.AddStep(finalStep)
	}

	return trace, nil
}

// ExecuteStep 实现ReActExecutor接口
func (e *StandardReActExecutor) ExecuteStep(ctx context.Context, agent Agent, step *ReActStep, toolCtx *ToolExecutionContext) error {
	if step.Action == "" {
		return fmt.Errorf("no action specified")
	}

	startTime := time.Now()
	defer func() {
		step.Duration = time.Since(startTime)
	}()

	// 执行工具
	result, err := toolCtx.ExecuteTool(ctx, step.Action, step.ActionInput)
	if err != nil {
		return fmt.Errorf("tool execution failed: %w", err)
	}

	// 将结果转换为观察结果
	if result != nil {
		step.Observation = fmt.Sprintf("%v", result)
	} else {
		step.Observation = "Tool executed successfully with no output"
	}

	return nil
}

// ShouldContinue 实现ReActExecutor接口
func (e *StandardReActExecutor) ShouldContinue(trace *ReActTrace, config *ReActConfig) bool {
	// 如果已完成，不继续
	if trace.IsCompleted || trace.HasCompletedStep() {
		return false
	}

	// 如果达到最大迭代次数，不继续
	if trace.IterationCount >= config.MaxIterations {
		return false
	}

	// 检查最后一个步骤是否有错误
	if lastStep := trace.GetLastStep(); lastStep != nil && lastStep.Error != "" {
		return false
	}

	return true
}

// buildReActPrompt 构建ReAct提示
func (e *StandardReActExecutor) buildReActPrompt(task Task, toolCtx *ToolExecutionContext, trace *ReActTrace) (string, error) {
	var prompt strings.Builder

	// 任务描述
	prompt.WriteString(fmt.Sprintf("Task: %s\n\n", task.GetDescription()))

	// 期望输出
	if expectedOutput := task.GetExpectedOutput(); expectedOutput != "" {
		prompt.WriteString(fmt.Sprintf("Expected Output: %s\n\n", expectedOutput))
	}

	// 可用工具
	if toolCtx.HasTools() {
		prompt.WriteString("Available Tools:\n")
		prompt.WriteString(toolCtx.GetToolsDescription())
		prompt.WriteString("\n\n")
	}

	// ReAct格式说明
	prompt.WriteString("Use the following format for your response:\n\n")
	prompt.WriteString("Thought: [your reasoning about what to do]\n")
	if toolCtx.HasTools() {
		prompt.WriteString("Action: [the action/tool to use]\n")
		prompt.WriteString("Action Input: [the input for the action as JSON]\n")
		prompt.WriteString("Observation: [the result of the action]\n")
		prompt.WriteString("... (this Thought/Action/Action Input/Observation can repeat N times)\n")
	}
	prompt.WriteString("Thought: [final reasoning]\n")
	prompt.WriteString("Final Answer: [your final answer to the task]\n\n")

	// 历史步骤（如果有）
	if trace != nil && len(trace.Steps) > 0 {
		prompt.WriteString("Previous steps:\n")
		for _, step := range trace.Steps {
			prompt.WriteString(e.parser.Format(step))
			prompt.WriteString("\n\n")
		}
	}

	prompt.WriteString("Begin!\n")

	return prompt.String(), nil
}

// callLLM 调用LLM获取响应
func (e *StandardReActExecutor) callLLM(ctx context.Context, agent Agent, prompt string, trace *ReActTrace) (string, error) {
	llmProvider := agent.GetLLM()
	if llmProvider == nil {
		return "", fmt.Errorf("no LLM provider available")
	}

	// 构建消息
	messages := []llm.Message{
		{
			Role:    llm.RoleUser,
			Content: prompt,
		},
	}

	// 调用LLM
	response, err := llmProvider.Call(ctx, messages, &llm.CallOptions{})
	if err != nil {
		return "", err
	}

	return response.Content, nil
}

// updatePromptWithStep 使用步骤更新提示
func (e *StandardReActExecutor) updatePromptWithStep(originalPrompt string, step *ReActStep) string {
	stepText := e.parser.Format(step)
	return originalPrompt + "\n" + stepText + "\n"
}
