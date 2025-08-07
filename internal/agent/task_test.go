package agent

import (
	"testing"
)

func TestNewBaseTask(t *testing.T) {
	description := "Test task description"
	expectedOutput := "Expected test output"

	task := NewBaseTask(description, expectedOutput)

	if task == nil {
		t.Fatal("expected task, got nil")
	}

	if task.GetDescription() != description {
		t.Errorf("expected description %s, got %s", description, task.GetDescription())
	}

	if task.GetExpectedOutput() != expectedOutput {
		t.Errorf("expected output %s, got %s", expectedOutput, task.GetExpectedOutput())
	}

	if task.GetID() == "" {
		t.Error("expected non-empty task ID")
	}

	if task.IsHumanInputRequired() {
		t.Error("expected human input to be false by default")
	}

	if task.GetOutputFormat() != OutputFormatRAW {
		t.Errorf("expected output format RAW, got %v", task.GetOutputFormat())
	}
}

func TestTaskWithOptions(t *testing.T) {
	description := "Test task with options"
	expectedOutput := "Expected output with options"
	context := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}

	task := NewTaskWithOptions(
		description,
		expectedOutput,
		WithContext(context),
		WithHumanInput(true),
		WithOutputFormat(OutputFormatJSON),
	)

	if task.GetDescription() != description {
		t.Errorf("expected description %s, got %s", description, task.GetDescription())
	}

	if task.GetExpectedOutput() != expectedOutput {
		t.Errorf("expected output %s, got %s", expectedOutput, task.GetExpectedOutput())
	}

	if !task.IsHumanInputRequired() {
		t.Error("expected human input to be required")
	}

	if task.GetOutputFormat() != OutputFormatJSON {
		t.Errorf("expected output format JSON, got %v", task.GetOutputFormat())
	}

	taskContext := task.GetContext()
	if len(taskContext) != len(context) {
		t.Errorf("expected context length %d, got %d", len(context), len(taskContext))
	}

	for k, v := range context {
		if taskValue, exists := taskContext[k]; !exists || taskValue != v {
			t.Errorf("expected context[%s] = %v, got %v", k, v, taskValue)
		}
	}
}

func TestTaskValidation(t *testing.T) {
	tests := []struct {
		name        string
		task        Task
		expectError bool
	}{
		{
			name:        "valid task",
			task:        NewBaseTask("Valid description", "Valid expected output"),
			expectError: false,
		},
		{
			name:        "empty description",
			task:        NewBaseTask("", "Valid expected output"),
			expectError: true,
		},
		{
			name:        "empty expected output",
			task:        NewBaseTask("Valid description", ""),
			expectError: true,
		},
		{
			name: "human input required but not provided",
			task: NewTaskWithOptions(
				"Valid description",
				"Valid expected output",
				WithHumanInput(true),
			),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()

			if tt.expectError && err == nil {
				t.Error("expected validation error, got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestTaskHumanInput(t *testing.T) {
	task := NewBaseTask("Human input task", "Expected output")

	// 初始状态
	if task.IsHumanInputRequired() {
		t.Error("expected human input to be false initially")
	}

	if task.GetHumanInput() != "" {
		t.Errorf("expected empty human input initially, got %s", task.GetHumanInput())
	}

	// 设置需要人工输入
	task.SetHumanInputRequired(true)
	if !task.IsHumanInputRequired() {
		t.Error("expected human input to be required after setting")
	}

	// 设置人工输入
	humanInput := "This is human input"
	task.SetHumanInput(humanInput)
	if task.GetHumanInput() != humanInput {
		t.Errorf("expected human input %s, got %s", humanInput, task.GetHumanInput())
	}

	// 验证任务现在有效
	err := task.Validate()
	if err != nil {
		t.Errorf("task should be valid after setting human input: %v", err)
	}
}

func TestTaskContext(t *testing.T) {
	baseTask := NewBaseTask("Context task", "Expected output")
	task := Task(baseTask)

	// 初始上下文应为空
	context := task.GetContext()
	if len(context) != 0 {
		t.Errorf("expected empty context initially, got %d items", len(context))
	}

	// 添加上下文项
	baseTask.AddContext("key1", "value1")
	baseTask.AddContext("key2", 42)
	baseTask.AddContext("key3", true)

	context = task.GetContext()
	if len(context) != 3 {
		t.Errorf("expected 3 context items, got %d", len(context))
	}

	// 验证上下文值
	if value, exists := baseTask.GetContextValue("key1"); !exists || value != "value1" {
		t.Errorf("expected key1 = value1, got %v (exists: %v)", value, exists)
	}

	if value, exists := baseTask.GetContextValue("key2"); !exists || value != 42 {
		t.Errorf("expected key2 = 42, got %v (exists: %v)", value, exists)
	}

	if value, exists := baseTask.GetContextValue("key3"); !exists || value != true {
		t.Errorf("expected key3 = true, got %v (exists: %v)", value, exists)
	}

	// 测试不存在的键
	if value, exists := baseTask.GetContextValue("nonexistent"); exists {
		t.Errorf("expected nonexistent key to not exist, got %v", value)
	}
}

func TestTaskTools(t *testing.T) {
	baseTask := NewBaseTask("Tools task", "Expected output")
	task := Task(baseTask)

	// 初始工具应为空
	tools := task.GetTools()
	if len(tools) != 0 {
		t.Errorf("expected no tools initially, got %d", len(tools))
	}

	// 添加工具
	calculator := NewCalculatorTool()
	fileReader := NewFileReaderTool()

	baseTask.AddTool(calculator)
	baseTask.AddTool(fileReader)

	tools = task.GetTools()
	if len(tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(tools))
	}

	// 验证工具
	foundCalculator := false
	foundFileReader := false

	for _, tool := range tools {
		switch tool.GetName() {
		case "calculator":
			foundCalculator = true
		case "file_reader":
			foundFileReader = true
		}
	}

	if !foundCalculator {
		t.Error("calculator tool not found")
	}

	if !foundFileReader {
		t.Error("file_reader tool not found")
	}
}

func TestTaskClone(t *testing.T) {
	originalBase := NewTaskWithOptions(
		"Original task",
		"Original expected output",
		WithContext(map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		}),
		WithHumanInput(true),
		WithOutputFormat(OutputFormatJSON),
	)
	original := Task(originalBase)

	// 设置人工输入
	original.SetHumanInput("Original human input")

	// 添加工具
	calculator := NewCalculatorTool()
	originalBase.AddTool(calculator)

	// 克隆任务
	cloned := originalBase.Clone()

	// 验证基础属性
	if cloned.GetDescription() != original.GetDescription() {
		t.Errorf("cloned description mismatch: expected %s, got %s",
			original.GetDescription(), cloned.GetDescription())
	}

	if cloned.GetExpectedOutput() != original.GetExpectedOutput() {
		t.Errorf("cloned expected output mismatch: expected %s, got %s",
			original.GetExpectedOutput(), cloned.GetExpectedOutput())
	}

	if cloned.IsHumanInputRequired() != original.IsHumanInputRequired() {
		t.Errorf("cloned human input requirement mismatch: expected %v, got %v",
			original.IsHumanInputRequired(), cloned.IsHumanInputRequired())
	}

	if cloned.GetHumanInput() != original.GetHumanInput() {
		t.Errorf("cloned human input mismatch: expected %s, got %s",
			original.GetHumanInput(), cloned.GetHumanInput())
	}

	if cloned.GetOutputFormat() != original.GetOutputFormat() {
		t.Errorf("cloned output format mismatch: expected %v, got %v",
			original.GetOutputFormat(), cloned.GetOutputFormat())
	}

	// 验证ID不同
	if cloned.GetID() == original.GetID() {
		t.Error("cloned task should have different ID")
	}

	// 验证上下文
	originalContext := original.GetContext()
	clonedContext := cloned.GetContext()

	if len(clonedContext) != len(originalContext) {
		t.Errorf("cloned context length mismatch: expected %d, got %d",
			len(originalContext), len(clonedContext))
	}

	for k, v := range originalContext {
		if clonedValue, exists := clonedContext[k]; !exists || clonedValue != v {
			t.Errorf("cloned context[%s] mismatch: expected %v, got %v", k, v, clonedValue)
		}
	}

	// 验证工具
	originalTools := original.GetTools()
	clonedTools := cloned.GetTools()

	if len(clonedTools) != len(originalTools) {
		t.Errorf("cloned tools count mismatch: expected %d, got %d",
			len(originalTools), len(clonedTools))
	}
}

func TestConditionalTask(t *testing.T) {
	condition := func(context map[string]interface{}) bool {
		value, exists := context["execute"]
		if !exists {
			return false
		}
		execute, ok := value.(bool)
		return ok && execute
	}

	task := NewConditionalTask("Conditional task", "Expected output", condition)

	// 测试条件为false的情况
	context1 := map[string]interface{}{
		"execute": false,
	}

	if task.ShouldExecuteSimple(context1) {
		t.Error("task should not execute when condition is false")
	}

	// 测试条件为true的情况
	context2 := map[string]interface{}{
		"execute": true,
	}

	if !task.ShouldExecuteSimple(context2) {
		t.Error("task should execute when condition is true")
	}

	// 测试缺少键的情况
	context3 := map[string]interface{}{
		"other_key": "value",
	}

	if task.ShouldExecuteSimple(context3) {
		t.Error("task should not execute when key is missing")
	}
}

func TestTaskBuilder(t *testing.T) {
	calculator := NewCalculatorTool()

	task := NewTaskBuilder().
		WithDescription("Builder task").
		WithExpectedOutput("Builder expected output").
		WithContext(map[string]interface{}{
			"builder_key": "builder_value",
		}).
		WithHumanInput(true).
		WithOutputFormat(OutputFormatJSON).
		WithTools(calculator).
		Build()

	if task.GetDescription() != "Builder task" {
		t.Errorf("expected description 'Builder task', got %s", task.GetDescription())
	}

	if task.GetExpectedOutput() != "Builder expected output" {
		t.Errorf("expected output 'Builder expected output', got %s", task.GetExpectedOutput())
	}

	if !task.IsHumanInputRequired() {
		t.Error("expected human input to be required")
	}

	if task.GetOutputFormat() != OutputFormatJSON {
		t.Errorf("expected output format JSON, got %v", task.GetOutputFormat())
	}

	context := task.GetContext()
	if value, exists := context["builder_key"]; !exists || value != "builder_value" {
		t.Errorf("expected context[builder_key] = builder_value, got %v (exists: %v)", value, exists)
	}

	tools := task.GetTools()
	if len(tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(tools))
	}

	if tools[0].GetName() != "calculator" {
		t.Errorf("expected calculator tool, got %s", tools[0].GetName())
	}
}

func TestTaskCollection(t *testing.T) {
	collection := NewTaskCollection()

	// 初始状态
	if !collection.IsEmpty() {
		t.Error("expected collection to be empty initially")
	}

	if collection.Size() != 0 {
		t.Errorf("expected size 0, got %d", collection.Size())
	}

	// 添加任务
	task1 := NewBaseTask("Task 1", "Output 1")
	task2 := NewBaseTask("Task 2", "Output 2")
	task3 := NewBaseTask("Task 3", "Output 3")

	collection.Add(task1).Add(task2).AddTasks(task3)

	if collection.IsEmpty() {
		t.Error("expected collection to not be empty after adding tasks")
	}

	if collection.Size() != 3 {
		t.Errorf("expected size 3, got %d", collection.Size())
	}

	// 获取任务
	retrieved, found := collection.Get(0)
	if !found {
		t.Error("expected to find task at index 0")
	}
	if retrieved.GetID() != task1.GetID() {
		t.Error("retrieved task mismatch")
	}

	// 按ID获取任务
	retrieved, found = collection.GetByID(task2.GetID())
	if !found {
		t.Error("expected to find task by ID")
	}
	if retrieved.GetID() != task2.GetID() {
		t.Error("retrieved task by ID mismatch")
	}

	// 检查包含
	if !collection.Contains(task3.GetID()) {
		t.Error("expected collection to contain task3")
	}

	// 获取所有任务
	allTasks := collection.All()
	if len(allTasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(allTasks))
	}

	// 过滤任务
	filtered := collection.Filter(func(task Task) bool {
		return task.GetDescription() == "Task 1" || task.GetDescription() == "Task 3"
	})

	if filtered.Size() != 2 {
		t.Errorf("expected filtered size 2, got %d", filtered.Size())
	}

	// 移除任务
	removed := collection.Remove(task2.GetID())
	if !removed {
		t.Error("expected to remove task2")
	}

	if collection.Size() != 2 {
		t.Errorf("expected size 2 after removal, got %d", collection.Size())
	}

	if collection.Contains(task2.GetID()) {
		t.Error("expected task2 to be removed")
	}

	// 清空集合
	collection.Clear()
	if !collection.IsEmpty() {
		t.Error("expected collection to be empty after clear")
	}
}

func TestTaskFactoryFunctions(t *testing.T) {
	// 测试简单任务创建
	simple := CreateSimpleTask("Simple description", "Simple output")
	if simple.GetDescription() != "Simple description" {
		t.Errorf("expected description 'Simple description', got %s", simple.GetDescription())
	}

	// 测试分析任务创建
	analysis := CreateAnalysisTask("test subject", "statistical")
	if analysis.GetOutputFormat() != OutputFormatJSON {
		t.Errorf("expected JSON output format for analysis task, got %v", analysis.GetOutputFormat())
	}

	// 测试研究任务创建
	research := CreateResearchTask("AI technology")
	if research.GetDescription() == "" {
		t.Error("expected non-empty description for research task")
	}

	// 测试审查任务创建
	review := CreateReviewTask("test content", "quality criteria")
	if review.GetOutputFormat() != OutputFormatJSON {
		t.Errorf("expected JSON output format for review task, got %v", review.GetOutputFormat())
	}

	// 测试交互式任务创建
	interactive := CreateInteractiveTask("Interactive task", "Interactive output")
	if !interactive.IsHumanInputRequired() {
		t.Error("expected interactive task to require human input")
	}
}

func BenchmarkTaskCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		task := NewBaseTask("Benchmark task", "Benchmark output")
		_ = task.GetID()
	}
}

func BenchmarkTaskWithOptions(b *testing.B) {
	context := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}

	for i := 0; i < b.N; i++ {
		task := NewTaskWithOptions(
			"Benchmark task with options",
			"Benchmark output",
			WithContext(context),
			WithHumanInput(true),
			WithOutputFormat(OutputFormatJSON),
		)
		_ = task.GetID()
	}
}
