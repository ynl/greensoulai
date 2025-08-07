package agent

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// BaseTask 实现了Task接口的基础结构
type BaseTask struct {
	id                 string
	description        string
	expectedOutput     string
	context            map[string]interface{}
	humanInput         string
	humanInputRequired bool
	outputFormat       OutputFormat
	tools              []Tool
}

// NewBaseTask 创建新的BaseTask实例
func NewBaseTask(description, expectedOutput string) *BaseTask {
	return &BaseTask{
		id:                 uuid.New().String(),
		description:        description,
		expectedOutput:     expectedOutput,
		context:            make(map[string]interface{}),
		humanInput:         "",
		humanInputRequired: false,
		outputFormat:       OutputFormatRAW,
		tools:              make([]Tool, 0),
	}
}

// NewTaskWithOptions 使用选项创建任务
func NewTaskWithOptions(description, expectedOutput string, options ...TaskOption) *BaseTask {
	task := NewBaseTask(description, expectedOutput)

	for _, option := range options {
		option(task)
	}

	return task
}

// TaskOption 定义任务配置选项
type TaskOption func(*BaseTask)

// WithContext 设置任务上下文
func WithContext(context map[string]interface{}) TaskOption {
	return func(t *BaseTask) {
		t.context = context
	}
}

// WithHumanInput 设置任务需要人工输入
func WithHumanInput(required bool) TaskOption {
	return func(t *BaseTask) {
		t.humanInputRequired = required
	}
}

// WithOutputFormat 设置输出格式
func WithOutputFormat(format OutputFormat) TaskOption {
	return func(t *BaseTask) {
		t.outputFormat = format
	}
}

// WithTools 设置任务专用工具
func WithTools(tools ...Tool) TaskOption {
	return func(t *BaseTask) {
		t.tools = append(t.tools, tools...)
	}
}

// WithID 设置任务ID (用于测试或特殊情况)
func WithID(id string) TaskOption {
	return func(t *BaseTask) {
		t.id = id
	}
}

// Task接口实现
func (t *BaseTask) GetID() string {
	return t.id
}

func (t *BaseTask) GetDescription() string {
	return t.description
}

func (t *BaseTask) GetExpectedOutput() string {
	return t.expectedOutput
}

func (t *BaseTask) GetContext() map[string]interface{} {
	return t.context
}

func (t *BaseTask) IsHumanInputRequired() bool {
	return t.humanInputRequired
}

func (t *BaseTask) SetHumanInput(input string) {
	t.humanInput = strings.TrimSpace(input)
}

func (t *BaseTask) GetHumanInput() string {
	return t.humanInput
}

func (t *BaseTask) GetOutputFormat() OutputFormat {
	return t.outputFormat
}

func (t *BaseTask) GetTools() []Tool {
	return t.tools
}

func (t *BaseTask) Validate() error {
	if t.description == "" {
		return fmt.Errorf("task description is required")
	}

	if t.expectedOutput == "" {
		return fmt.Errorf("expected output is required")
	}

	if t.humanInputRequired && t.humanInput == "" {
		return fmt.Errorf("human input is required but not provided")
	}

	return nil
}

// SetContext 设置任务上下文
func (t *BaseTask) SetContext(context map[string]interface{}) {
	t.context = context
}

// AddContext 添加上下文项
func (t *BaseTask) AddContext(key string, value interface{}) {
	if t.context == nil {
		t.context = make(map[string]interface{})
	}
	t.context[key] = value
}

// GetContextValue 获取上下文值
func (t *BaseTask) GetContextValue(key string) (interface{}, bool) {
	value, exists := t.context[key]
	return value, exists
}

// SetHumanInputRequired 设置是否需要人工输入
func (t *BaseTask) SetHumanInputRequired(required bool) {
	t.humanInputRequired = required
}

// SetOutputFormat 设置输出格式
func (t *BaseTask) SetOutputFormat(format OutputFormat) {
	t.outputFormat = format
}

// AddTool 添加任务专用工具
func (t *BaseTask) AddTool(tool Tool) {
	t.tools = append(t.tools, tool)
}

// SetTools 设置任务工具
func (t *BaseTask) SetTools(tools []Tool) {
	t.tools = tools
}

// Clone 创建任务副本
func (t *BaseTask) Clone() Task {
	clone := &BaseTask{
		id:                 uuid.New().String(), // 新的ID
		description:        t.description,
		expectedOutput:     t.expectedOutput,
		context:            make(map[string]interface{}),
		humanInput:         t.humanInput,
		humanInputRequired: t.humanInputRequired,
		outputFormat:       t.outputFormat,
		tools:              make([]Tool, len(t.tools)),
	}

	// 深拷贝上下文
	for k, v := range t.context {
		clone.context[k] = v
	}

	// 拷贝工具切片
	copy(clone.tools, t.tools)

	return clone
}

// String 返回任务的字符串表示
func (t *BaseTask) String() string {
	return fmt.Sprintf("Task{ID: %s, Description: %s, HumanInput: %v}",
		t.id, t.description, t.humanInputRequired)
}

// ConditionalTask 表示有条件执行的任务
type ConditionalTask struct {
	*BaseTask
	condition func(context map[string]interface{}) bool
}

// NewConditionalTask 创建条件任务
func NewConditionalTask(description, expectedOutput string, condition func(context map[string]interface{}) bool) *ConditionalTask {
	return &ConditionalTask{
		BaseTask:  NewBaseTask(description, expectedOutput),
		condition: condition,
	}
}

// ShouldExecute 检查是否应该执行任务
func (ct *ConditionalTask) ShouldExecute(context map[string]interface{}) bool {
	if ct.condition == nil {
		return true
	}
	return ct.condition(context)
}

// TaskBuilder 任务构建器
type TaskBuilder struct {
	task *BaseTask
}

// NewTaskBuilder 创建任务构建器
func NewTaskBuilder() *TaskBuilder {
	return &TaskBuilder{
		task: &BaseTask{
			id:      uuid.New().String(),
			context: make(map[string]interface{}),
			tools:   make([]Tool, 0),
		},
	}
}

// WithDescription 设置描述
func (tb *TaskBuilder) WithDescription(description string) *TaskBuilder {
	tb.task.description = description
	return tb
}

// WithExpectedOutput 设置期望输出
func (tb *TaskBuilder) WithExpectedOutput(expectedOutput string) *TaskBuilder {
	tb.task.expectedOutput = expectedOutput
	return tb
}

// WithContext 设置上下文
func (tb *TaskBuilder) WithContext(context map[string]interface{}) *TaskBuilder {
	tb.task.context = context
	return tb
}

// WithHumanInput 设置人工输入需求
func (tb *TaskBuilder) WithHumanInput(required bool) *TaskBuilder {
	tb.task.humanInputRequired = required
	return tb
}

// WithOutputFormat 设置输出格式
func (tb *TaskBuilder) WithOutputFormat(format OutputFormat) *TaskBuilder {
	tb.task.outputFormat = format
	return tb
}

// WithTools 设置工具
func (tb *TaskBuilder) WithTools(tools ...Tool) *TaskBuilder {
	tb.task.tools = append(tb.task.tools, tools...)
	return tb
}

// Build 构建任务
func (tb *TaskBuilder) Build() Task {
	return tb.task
}

// TaskCollection 任务集合
type TaskCollection struct {
	tasks []Task
}

// NewTaskCollection 创建任务集合
func NewTaskCollection() *TaskCollection {
	return &TaskCollection{
		tasks: make([]Task, 0),
	}
}

// Add 添加任务
func (tc *TaskCollection) Add(task Task) *TaskCollection {
	tc.tasks = append(tc.tasks, task)
	return tc
}

// AddTasks 批量添加任务
func (tc *TaskCollection) AddTasks(tasks ...Task) *TaskCollection {
	tc.tasks = append(tc.tasks, tasks...)
	return tc
}

// Get 获取指定索引的任务
func (tc *TaskCollection) Get(index int) (Task, bool) {
	if index < 0 || index >= len(tc.tasks) {
		return nil, false
	}
	return tc.tasks[index], true
}

// GetByID 根据ID获取任务
func (tc *TaskCollection) GetByID(id string) (Task, bool) {
	for _, task := range tc.tasks {
		if task.GetID() == id {
			return task, true
		}
	}
	return nil, false
}

// Size 返回任务数量
func (tc *TaskCollection) Size() int {
	return len(tc.tasks)
}

// IsEmpty 检查是否为空
func (tc *TaskCollection) IsEmpty() bool {
	return len(tc.tasks) == 0
}

// All 返回所有任务
func (tc *TaskCollection) All() []Task {
	result := make([]Task, len(tc.tasks))
	copy(result, tc.tasks)
	return result
}

// Filter 根据条件过滤任务
func (tc *TaskCollection) Filter(predicate func(Task) bool) *TaskCollection {
	filtered := NewTaskCollection()
	for _, task := range tc.tasks {
		if predicate(task) {
			filtered.Add(task)
		}
	}
	return filtered
}

// Map 映射任务
func (tc *TaskCollection) Map(mapper func(Task) Task) *TaskCollection {
	mapped := NewTaskCollection()
	for _, task := range tc.tasks {
		mapped.Add(mapper(task))
	}
	return mapped
}

// ForEach 遍历任务
func (tc *TaskCollection) ForEach(action func(Task)) {
	for _, task := range tc.tasks {
		action(task)
	}
}

// Clear 清空任务集合
func (tc *TaskCollection) Clear() {
	tc.tasks = tc.tasks[:0]
}

// Remove 移除指定任务
func (tc *TaskCollection) Remove(taskID string) bool {
	for i, task := range tc.tasks {
		if task.GetID() == taskID {
			tc.tasks = append(tc.tasks[:i], tc.tasks[i+1:]...)
			return true
		}
	}
	return false
}

// Contains 检查是否包含指定任务
func (tc *TaskCollection) Contains(taskID string) bool {
	_, found := tc.GetByID(taskID)
	return found
}

// 预定义的任务模板和工厂函数

// CreateSimpleTask 创建简单任务
func CreateSimpleTask(description, expectedOutput string) Task {
	return NewBaseTask(description, expectedOutput)
}

// CreateAnalysisTask 创建分析任务
func CreateAnalysisTask(subject, analysisType string) Task {
	return NewTaskWithOptions(
		fmt.Sprintf("Analyze %s using %s approach", subject, analysisType),
		fmt.Sprintf("Detailed %s analysis report with insights and recommendations", analysisType),
		WithOutputFormat(OutputFormatJSON),
	)
}

// CreateResearchTask 创建研究任务
func CreateResearchTask(topic string) Task {
	return NewTaskWithOptions(
		fmt.Sprintf("Research comprehensive information about: %s", topic),
		"Well-structured research report with sources, key findings, and summary",
		WithOutputFormat(OutputFormatRAW),
	)
}

// CreateReviewTask 创建审查任务
func CreateReviewTask(content, criteria string) Task {
	return NewTaskWithOptions(
		fmt.Sprintf("Review the following content against criteria: %s\n\nContent: %s", criteria, content),
		"Detailed review with scores, feedback, and improvement suggestions",
		WithOutputFormat(OutputFormatJSON),
	)
}

// CreateInteractiveTask 创建交互式任务
func CreateInteractiveTask(description, expectedOutput string) Task {
	return NewTaskWithOptions(
		description,
		expectedOutput,
		WithHumanInput(true),
	)
}
