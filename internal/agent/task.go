package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

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

	// 新增字段，对标Python版本
	assignedAgent  Agent // 对标Python版本的task.agent
	asyncExecution bool  // 对标Python版本的task.async_execution

	// Python版本对标的高级功能
	name            string                                   // 对标Python的name
	outputFile      string                                   // 对标Python的output_file
	createDirectory bool                                     // 对标Python的create_directory
	callback        func(context.Context, *TaskOutput) error // 对标Python的callback
	contextTasks    []Task                                   // 对标Python的context: List[Task]
	retryCount      int                                      // 对标Python的retry_count
	maxRetries      int                                      // 对标Python的max_retries
	guardrail       TaskGuardrail                            // 对标Python的_guardrail
	markdownOutput  bool                                     // 对标Python的markdown

	// 并发安全
	mu sync.RWMutex
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

		// 新增字段默认值，完全对标Python版本
		assignedAgent:   nil,   // 默认没有预分配Agent
		asyncExecution:  false, // 默认同步执行
		name:            "",
		outputFile:      "",
		createDirectory: true, // 对标Python默认值
		callback:        nil,
		contextTasks:    make([]Task, 0),
		retryCount:      0,
		maxRetries:      3, // 对标Python默认重试次数
		guardrail:       nil,
		markdownOutput:  false,
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

// SetDescription 设置任务描述，支持推理功能修改任务描述
func (t *BaseTask) SetDescription(description string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.description = description
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

// BaseConditionalTask 条件任务的基础实现，实现ConditionalTask接口
type BaseConditionalTask struct {
	*BaseTask
	condition         func(*TaskOutput) bool
	originalCondition func(context map[string]interface{}) bool // 保存原始条件函数
	skippedOutput     *TaskOutput
}

// NewConditionalTask 创建条件任务实例
func NewConditionalTask(description, expectedOutput string, condition func(context map[string]interface{}) bool) *BaseConditionalTask {
	// 适配函数签名，从map转换为TaskOutput检查
	adaptedCondition := func(output *TaskOutput) bool {
		if condition == nil {
			return true
		}
		// 为了保持向后兼容，我们直接使用一个简单的实现
		// 实际上我们需要把原始的condition函数存储起来，在ShouldExecuteSimple中使用
		return true // 这里暂时返回true，实际逻辑在ShouldExecuteSimple中
	}

	task := &BaseConditionalTask{
		BaseTask:  NewBaseTask(description, expectedOutput),
		condition: adaptedCondition,
	}

	// 保存原始条件函数以供ShouldExecuteSimple使用
	task.originalCondition = condition

	return task
}

// 实现ConditionalTask接口的方法

// ShouldExecute 检查是否应该执行任务
func (bct *BaseConditionalTask) ShouldExecute(ctx context.Context, context *TaskOutput) (bool, error) {
	if bct.condition == nil {
		return true, nil
	}
	return bct.condition(context), nil
}

// GetCondition 获取条件函数
func (bct *BaseConditionalTask) GetCondition() func(*TaskOutput) bool {
	return bct.condition
}

// SetCondition 设置条件函数
func (bct *BaseConditionalTask) SetCondition(condition func(*TaskOutput) bool) {
	bct.mu.Lock()
	defer bct.mu.Unlock()
	bct.condition = condition
}

// GetSkippedTaskOutput 获取跳过任务时的默认输出
func (bct *BaseConditionalTask) GetSkippedTaskOutput() *TaskOutput {
	if bct.skippedOutput != nil {
		return bct.skippedOutput
	}

	// 返回默认的跳过输出
	return &TaskOutput{
		Raw:            "Task skipped due to condition not met",
		Agent:          "",
		Description:    bct.description,
		ExpectedOutput: bct.expectedOutput,
		OutputFormat:   OutputFormatRAW,
		CreatedAt:      time.Now(),
		Metadata: map[string]interface{}{
			"skipped": true,
			"reason":  "condition_not_met",
		},
	}
}

// 为了向后兼容，保留原有的ShouldExecute方法签名
func (bct *BaseConditionalTask) ShouldExecuteSimple(context map[string]interface{}) bool {
	if bct.originalCondition == nil {
		return true
	}
	// 直接使用原始条件函数
	return bct.originalCondition(context)
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

// 实现新的Task接口方法，对标Python版本

// GetAssignedAgent 获取任务预分配的Agent（对标Python的task.agent）
func (t *BaseTask) GetAssignedAgent() Agent {
	return t.assignedAgent
}

// SetAssignedAgent 设置任务预分配的Agent（对标Python的task.agent）
func (t *BaseTask) SetAssignedAgent(agent Agent) error {
	t.assignedAgent = agent
	return nil
}

// IsAsyncExecution 检查任务是否为异步执行（对标Python的task.async_execution）
func (t *BaseTask) IsAsyncExecution() bool {
	return t.asyncExecution
}

// SetAsyncExecution 设置任务异步执行模式（对标Python的task.async_execution）
func (t *BaseTask) SetAsyncExecution(async bool) {
	t.asyncExecution = async
}

// SetContext 设置任务上下文（对标Python的task.context设置）
func (t *BaseTask) SetContext(context map[string]interface{}) {
	if t.context == nil {
		t.context = make(map[string]interface{})
	}

	// 合并新的上下文到现有上下文中
	for key, value := range context {
		t.context[key] = value
	}
}

// 新增Python版本对标功能实现

// GetName 获取任务名称
func (t *BaseTask) GetName() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.name
}

// SetName 设置任务名称
func (t *BaseTask) SetName(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.name = name
}

// GetOutputFile 获取输出文件路径
func (t *BaseTask) GetOutputFile() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.outputFile
}

// SetOutputFile 设置输出文件路径
func (t *BaseTask) SetOutputFile(filename string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.outputFile = filename
	return nil
}

// GetCreateDirectory 获取是否自动创建目录
func (t *BaseTask) GetCreateDirectory() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.createDirectory
}

// SetCreateDirectory 设置是否自动创建目录
func (t *BaseTask) SetCreateDirectory(create bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.createDirectory = create
}

// GetCallback 获取回调函数
func (t *BaseTask) GetCallback() func(context.Context, *TaskOutput) error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.callback
}

// SetCallback 设置回调函数
func (t *BaseTask) SetCallback(callback func(context.Context, *TaskOutput) error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.callback = callback
}

// GetContextTasks 获取上下文任务列表（对标Python的context: List[Task]）
func (t *BaseTask) GetContextTasks() []Task {
	t.mu.RLock()
	defer t.mu.RUnlock()
	// 返回副本以防止外部修改
	result := make([]Task, len(t.contextTasks))
	copy(result, t.contextTasks)
	return result
}

// SetContextTasks 设置上下文任务列表
func (t *BaseTask) SetContextTasks(tasks []Task) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.contextTasks = make([]Task, len(tasks))
	copy(t.contextTasks, tasks)
}

// GetRetryCount 获取当前重试次数
func (t *BaseTask) GetRetryCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.retryCount
}

// GetMaxRetries 获取最大重试次数
func (t *BaseTask) GetMaxRetries() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.maxRetries
}

// SetMaxRetries 设置最大重试次数
func (t *BaseTask) SetMaxRetries(maxRetries int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.maxRetries = maxRetries
}

// HasGuardrail 检查是否有护栏
func (t *BaseTask) HasGuardrail() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.guardrail != nil
}

// SetGuardrail 设置护栏
func (t *BaseTask) SetGuardrail(guardrail TaskGuardrail) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.guardrail = guardrail
}

// GetGuardrail 获取护栏
func (t *BaseTask) GetGuardrail() TaskGuardrail {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.guardrail
}

// IsMarkdownOutput 检查是否输出Markdown格式
func (t *BaseTask) IsMarkdownOutput() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.markdownOutput
}

// SetMarkdownOutput 设置是否输出Markdown格式
func (t *BaseTask) SetMarkdownOutput(markdown bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.markdownOutput = markdown
}

// 新增任务选项，支持Agent预分配和异步执行

// WithAssignedAgent 设置任务预分配的Agent
func WithAssignedAgent(agent Agent) TaskOption {
	return func(task *BaseTask) {
		task.assignedAgent = agent
	}
}

// WithAsyncExecution 设置任务为异步执行
func WithAsyncExecution(async bool) TaskOption {
	return func(task *BaseTask) {
		task.asyncExecution = async
	}
}
