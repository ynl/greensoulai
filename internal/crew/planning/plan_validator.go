package planning

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// PlanValidatorImpl PlanValidator的实现
// 负责验证规划结果的有效性和完整性
type PlanValidatorImpl struct{}

// NewPlanValidator 创建计划验证器
func NewPlanValidator() PlanValidator {
	return &PlanValidatorImpl{}
}

// ValidatePlan 验证单个计划
func (pv *PlanValidatorImpl) ValidatePlan(ctx context.Context, plan *PlanPerTask) error {
	if plan == nil {
		return fmt.Errorf("plan cannot be nil")
	}

	// 验证任务描述
	if strings.TrimSpace(plan.Task) == "" {
		return fmt.Errorf("task description cannot be empty")
	}

	// 验证计划内容
	if strings.TrimSpace(plan.Plan) == "" {
		return fmt.Errorf("plan content cannot be empty")
	}

	// 验证计划长度（基本合理性检查）
	if len(plan.Plan) < 10 {
		return fmt.Errorf("plan content too short (minimum 10 characters required)")
	}

	// 验证计划质量（基本检查）
	if err := pv.validatePlanQuality(plan.Plan); err != nil {
		return fmt.Errorf("plan quality validation failed: %w", err)
	}

	return nil
}

// ValidatePlanOutput 验证完整的规划输出
func (pv *PlanValidatorImpl) ValidatePlanOutput(ctx context.Context, output *PlannerTaskPydanticOutput) error {
	startTime := time.Now()

	if output == nil {
		return fmt.Errorf("plan output cannot be nil")
	}

	// 使用内置验证方法
	if err := output.Validate(); err != nil {
		return fmt.Errorf("plan output validation failed: %w", err)
	}

	// 验证每个计划
	for i, plan := range output.ListOfPlansPerTask {
		if err := pv.ValidatePlan(ctx, &plan); err != nil {
			return NewPlanValidationError("plan", i, err.Error())
		}
	}

	// 验证任务覆盖度（确保所有任务都有计划）
	if err := pv.validateTaskCoverage(output); err != nil {
		return fmt.Errorf("task coverage validation failed: %w", err)
	}

	// 验证计划一致性
	if err := pv.validatePlanConsistency(output); err != nil {
		return fmt.Errorf("plan consistency validation failed: %w", err)
	}

	_ = time.Since(startTime) // 用于将来的指标收集

	return nil
}

// ValidateTaskInfo 验证任务信息
func (pv *PlanValidatorImpl) ValidateTaskInfo(ctx context.Context, taskInfo *TaskInfo) error {
	if taskInfo == nil {
		return fmt.Errorf("task info cannot be nil")
	}

	// 验证必填字段
	if strings.TrimSpace(taskInfo.Description) == "" {
		return fmt.Errorf("task description cannot be empty")
	}

	if strings.TrimSpace(taskInfo.ExpectedOutput) == "" {
		return fmt.Errorf("task expected output cannot be empty")
	}

	// 验证ID（如果提供）
	if taskInfo.ID != "" && len(taskInfo.ID) < 3 {
		return fmt.Errorf("task ID must be at least 3 characters long")
	}

	// 验证描述长度
	if len(taskInfo.Description) > 10000 {
		return fmt.Errorf("task description too long (maximum 10000 characters)")
	}

	// 验证期望输出长度
	if len(taskInfo.ExpectedOutput) > 5000 {
		return fmt.Errorf("task expected output too long (maximum 5000 characters)")
	}

	// 验证工具列表
	if err := pv.validateTools(taskInfo.Tools); err != nil {
		return fmt.Errorf("invalid tools: %w", err)
	}

	// 验证上下文
	if err := pv.validateContext(taskInfo.Context); err != nil {
		return fmt.Errorf("invalid context: %w", err)
	}

	return nil
}

// ValidateTaskInfos 验证多个任务信息
func (pv *PlanValidatorImpl) ValidateTaskInfos(ctx context.Context, tasks []TaskInfo) error {
	if len(tasks) == 0 {
		return fmt.Errorf("tasks list cannot be empty")
	}

	if len(tasks) > 100 {
		return fmt.Errorf("too many tasks (maximum 100 tasks allowed)")
	}

	// 验证每个任务
	for i, task := range tasks {
		if err := pv.ValidateTaskInfo(ctx, &task); err != nil {
			return fmt.Errorf("validation failed for task %d: %w", i, err)
		}
	}

	// 验证任务ID唯一性（如果提供）
	if err := pv.validateTaskIDUniqueness(tasks); err != nil {
		return fmt.Errorf("task ID uniqueness validation failed: %w", err)
	}

	return nil
}

// validatePlanQuality 验证计划质量
func (pv *PlanValidatorImpl) validatePlanQuality(plan string) error {
	plan = strings.ToLower(strings.TrimSpace(plan))

	// 检查是否包含基本的规划要素
	requiredElements := []string{"step", "task", "tool", "output"}
	foundElements := 0

	for _, element := range requiredElements {
		if strings.Contains(plan, element) {
			foundElements++
		}
	}

	// 至少应该包含一半的必要要素
	if foundElements < len(requiredElements)/2 {
		return fmt.Errorf("plan lacks basic planning elements (found %d/%d)", foundElements, len(requiredElements))
	}

	// 检查计划结构（简单检查）
	if !strings.Contains(plan, "step") && !strings.Contains(plan, "1.") && !strings.Contains(plan, "first") {
		return fmt.Errorf("plan appears to lack step-by-step structure")
	}

	return nil
}

// validateTaskCoverage 验证任务覆盖度
func (pv *PlanValidatorImpl) validateTaskCoverage(output *PlannerTaskPydanticOutput) error {
	if len(output.ListOfPlansPerTask) == 0 {
		return fmt.Errorf("no plans found in output")
	}

	// 检查是否有重复的任务
	taskDescriptions := make(map[string]int)
	for i, plan := range output.ListOfPlansPerTask {
		if count, exists := taskDescriptions[plan.Task]; exists {
			return fmt.Errorf("duplicate task found at index %d: '%s' (previously seen at index %d)", i, plan.Task, count)
		}
		taskDescriptions[plan.Task] = i
	}

	return nil
}

// validatePlanConsistency 验证计划一致性
func (pv *PlanValidatorImpl) validatePlanConsistency(output *PlannerTaskPydanticOutput) error {
	// 检查计划长度的一致性（不应该有过短或过长的异常值）
	var planLengths []int
	for _, plan := range output.ListOfPlansPerTask {
		planLengths = append(planLengths, len(plan.Plan))
	}

	if len(planLengths) > 1 {
		avg := 0
		for _, length := range planLengths {
			avg += length
		}
		avg /= len(planLengths)

		// 检查是否有过于偏离平均值的计划
		for i, length := range planLengths {
			if length < avg/10 { // 长度小于平均值的1/10
				return fmt.Errorf("plan %d is suspiciously short (%d characters, average: %d)", i, length, avg)
			}
			if length > avg*10 { // 长度大于平均值的10倍
				return fmt.Errorf("plan %d is suspiciously long (%d characters, average: %d)", i, length, avg)
			}
		}
	}

	return nil
}

// validateTools 验证工具列表
func (pv *PlanValidatorImpl) validateTools(tools []string) error {
	if len(tools) > 50 {
		return fmt.Errorf("too many tools (maximum 50 tools allowed)")
	}

	// 检查工具名称的有效性
	for i, tool := range tools {
		if strings.TrimSpace(tool) == "" {
			return fmt.Errorf("tool %d cannot be empty", i)
		}
		if len(tool) > 200 {
			return fmt.Errorf("tool %d name too long (maximum 200 characters)", i)
		}
	}

	return nil
}

// validateContext 验证上下文
func (pv *PlanValidatorImpl) validateContext(context []string) error {
	if len(context) > 100 {
		return fmt.Errorf("too many context items (maximum 100 items allowed)")
	}

	for i, item := range context {
		if len(item) > 1000 {
			return fmt.Errorf("context item %d too long (maximum 1000 characters)", i)
		}
	}

	return nil
}

// validateTaskIDUniqueness 验证任务ID唯一性
func (pv *PlanValidatorImpl) validateTaskIDUniqueness(tasks []TaskInfo) error {
	ids := make(map[string]int)

	for i, task := range tasks {
		if task.ID == "" {
			continue // 跳过空ID
		}

		if existingIndex, exists := ids[task.ID]; exists {
			return fmt.Errorf("duplicate task ID '%s' found at indices %d and %d", task.ID, existingIndex, i)
		}
		ids[task.ID] = i
	}

	return nil
}

// GetValidationStatistics 获取验证统计信息
func (pv *PlanValidatorImpl) GetValidationStatistics(output *PlannerTaskPydanticOutput) map[string]interface{} {
	if output == nil {
		return map[string]interface{}{"error": "output is nil"}
	}

	stats := map[string]interface{}{
		"total_plans":         len(output.ListOfPlansPerTask),
		"average_plan_length": 0,
		"min_plan_length":     0,
		"max_plan_length":     0,
		"total_characters":    0,
	}

	if len(output.ListOfPlansPerTask) == 0 {
		return stats
	}

	totalChars := 0
	minLength := len(output.ListOfPlansPerTask[0].Plan)
	maxLength := len(output.ListOfPlansPerTask[0].Plan)

	for _, plan := range output.ListOfPlansPerTask {
		planLength := len(plan.Plan)
		totalChars += planLength

		if planLength < minLength {
			minLength = planLength
		}
		if planLength > maxLength {
			maxLength = planLength
		}
	}

	stats["total_characters"] = totalChars
	stats["average_plan_length"] = totalChars / len(output.ListOfPlansPerTask)
	stats["min_plan_length"] = minLength
	stats["max_plan_length"] = maxLength

	return stats
}
