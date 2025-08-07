package planning

import (
	"context"
	"fmt"
	"strings"

	"github.com/ynl/greensoulai/pkg/logger"
)

// TaskSummaryGeneratorImpl TaskSummaryGenerator的实现
// 负责生成任务摘要，对应Python版本的_create_tasks_summary()逻辑
type TaskSummaryGeneratorImpl struct {
	logger logger.Logger
}

// NewTaskSummaryGenerator 创建任务摘要生成器
func NewTaskSummaryGenerator(logger logger.Logger) TaskSummaryGenerator {
	return &TaskSummaryGeneratorImpl{
		logger: logger,
	}
}

// GenerateTaskSummary 为单个任务生成摘要
func (tsg *TaskSummaryGeneratorImpl) GenerateTaskSummary(ctx context.Context, taskInfo *TaskInfo, index int) (*TaskSummary, error) {
	if taskInfo == nil {
		return nil, fmt.Errorf("taskInfo cannot be nil")
	}

	// 生成代理知识信息（简化版实现）
	agentKnowledge := tsg.extractAgentKnowledge(taskInfo)

	summary := &TaskSummary{
		TaskNumber:        index + 1,
		Description:       taskInfo.Description,
		ExpectedOutput:    taskInfo.ExpectedOutput,
		AgentRole:         taskInfo.AgentRole,
		AgentGoal:         taskInfo.AgentGoal,
		TaskTools:         taskInfo.Tools,
		AgentTools:        taskInfo.Tools, // 简化实现：假设任务工具即代理工具
		AgentKnowledge:    agentKnowledge,
		AdditionalContext: taskInfo.Metadata,
	}

	// 处理空值情况，与Python版本保持一致
	if summary.AgentRole == "" {
		summary.AgentRole = "None"
	}
	if summary.AgentGoal == "" {
		summary.AgentGoal = "None"
	}
	if len(summary.TaskTools) == 0 {
		summary.TaskTools = []string{"agent has no tools"}
	}
	if len(summary.AgentTools) == 0 {
		summary.AgentTools = []string{"agent has no tools"}
	}

	return summary, nil
}

// GenerateTasksSummary 为多个任务生成摘要字符串，对应Python版本的_create_tasks_summary()
func (tsg *TaskSummaryGeneratorImpl) GenerateTasksSummary(ctx context.Context, tasks []TaskInfo) (string, error) {
	if len(tasks) == 0 {
		return "", fmt.Errorf("tasks list cannot be empty")
	}

	var tasksSummary []string

	for idx, taskInfo := range tasks {
		summary, err := tsg.GenerateTaskSummary(ctx, &taskInfo, idx)
		if err != nil {
			tsg.logger.Error("Failed to generate task summary",
				logger.Field{Key: "task_index", Value: idx},
				logger.Field{Key: "task_id", Value: taskInfo.ID},
				logger.Field{Key: "error", Value: err},
			)
			return "", NewTaskSummaryError(idx, taskInfo.ID, err.Error())
		}

		formattedSummary := tsg.FormatTaskSummary(summary)
		tasksSummary = append(tasksSummary, formattedSummary)
	}

	result := strings.Join(tasksSummary, "\n\n")

	tsg.logger.Debug("Tasks summary generated",
		logger.Field{Key: "task_count", Value: len(tasks)},
		logger.Field{Key: "summary_length", Value: len(result)},
	)

	return result, nil
}

// FormatTaskSummary 格式化任务摘要为字符串，与Python版本的格式保持一致
func (tsg *TaskSummaryGeneratorImpl) FormatTaskSummary(summary *TaskSummary) string {
	// 格式化工具列表
	taskToolsStr := tsg.formatToolsList(summary.TaskTools)
	agentToolsStr := tsg.formatToolsList(summary.AgentTools)

	// 格式化知识信息
	knowledgeStr := ""
	if len(summary.AgentKnowledge) > 0 {
		// 构建知识字符串，包含所有知识项
		var knowledgeItems []string
		for _, knowledge := range summary.AgentKnowledge {
			if knowledge != "" {
				knowledgeItems = append(knowledgeItems, fmt.Sprintf(`\"%s\"`, knowledge))
			}
		}
		if len(knowledgeItems) > 0 {
			knowledgeStr = fmt.Sprintf(`,
                "agent_knowledge": "[%s]"`, strings.Join(knowledgeItems, ", "))
		}
	}

	// 构建任务摘要，与Python版本的格式完全一致
	formattedSummary := fmt.Sprintf(`
                Task Number %d - %s
                "task_description": %s
                "task_expected_output": %s
                "agent": %s
                "agent_goal": %s
                "task_tools": %s
                "agent_tools": %s%s`,
		summary.TaskNumber,
		summary.Description,
		summary.Description,
		summary.ExpectedOutput,
		summary.AgentRole,
		summary.AgentGoal,
		taskToolsStr,
		agentToolsStr,
		knowledgeStr,
	)

	return strings.TrimSpace(formattedSummary)
}

// formatToolsList 格式化工具列表
func (tsg *TaskSummaryGeneratorImpl) formatToolsList(tools []string) string {
	if len(tools) == 0 || (len(tools) == 1 && tools[0] == "agent has no tools") {
		return `"agent has no tools"`
	}

	// 转换工具列表为字符串格式
	var formattedTools []string
	for _, tool := range tools {
		formattedTools = append(formattedTools, fmt.Sprintf(`"%s"`, tool))
	}

	return fmt.Sprintf("[%s]", strings.Join(formattedTools, ", "))
}

// extractAgentKnowledge 提取代理知识（简化实现）
// 对应Python版本的_get_agent_knowledge()方法
func (tsg *TaskSummaryGeneratorImpl) extractAgentKnowledge(taskInfo *TaskInfo) []string {
	var knowledge []string

	// 从任务元数据中提取知识信息
	if taskInfo.Metadata != nil {
		if knowledgeInterface, ok := taskInfo.Metadata["knowledge"]; ok {
			switch k := knowledgeInterface.(type) {
			case string:
				if k != "" {
					knowledge = append(knowledge, k)
				}
			case []string:
				knowledge = append(knowledge, k...)
			case []interface{}:
				for _, item := range k {
					if str, ok := item.(string); ok && str != "" {
						knowledge = append(knowledge, str)
					}
				}
			}
		}

		// 检查其他知识相关字段
		if sources, ok := taskInfo.Metadata["knowledge_sources"].([]string); ok {
			knowledge = append(knowledge, sources...)
		}

		// 检查接口类型的知识源
		if sourcesInterface, ok := taskInfo.Metadata["knowledge_sources"]; ok {
			switch sources := sourcesInterface.(type) {
			case []interface{}:
				for _, source := range sources {
					if sourceStr, ok := source.(string); ok && sourceStr != "" {
						knowledge = append(knowledge, sourceStr)
					}
				}
			}
		}
	}

	// 如果没有找到知识信息，返回空列表
	if len(knowledge) == 0 {
		return []string{}
	}

	// 移除重复项和空字符串
	knowledge = tsg.removeDuplicatesAndEmpty(knowledge)

	tsg.logger.Debug("Agent knowledge extracted",
		logger.Field{Key: "task_id", Value: taskInfo.ID},
		logger.Field{Key: "knowledge_count", Value: len(knowledge)},
	)

	return knowledge
}

// removeDuplicatesAndEmpty 移除重复项和空字符串
func (tsg *TaskSummaryGeneratorImpl) removeDuplicatesAndEmpty(items []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" && !seen[trimmed] {
			seen[trimmed] = true
			result = append(result, trimmed)
		}
	}

	return result
}

// ValidateTaskInfo 验证任务信息的完整性
func (tsg *TaskSummaryGeneratorImpl) ValidateTaskInfo(taskInfo *TaskInfo) error {
	if taskInfo == nil {
		return fmt.Errorf("task info cannot be nil")
	}

	if strings.TrimSpace(taskInfo.Description) == "" {
		return fmt.Errorf("task description cannot be empty")
	}

	if strings.TrimSpace(taskInfo.ExpectedOutput) == "" {
		return fmt.Errorf("task expected output cannot be empty")
	}

	return nil
}

// GetFormattingSummary 获取格式化摘要统计信息
func (tsg *TaskSummaryGeneratorImpl) GetFormattingSummary(tasks []TaskInfo) map[string]interface{} {
	summary := map[string]interface{}{
		"total_tasks":          len(tasks),
		"tasks_with_agent":     0,
		"tasks_with_tools":     0,
		"tasks_with_knowledge": 0,
		"average_desc_length":  0,
	}

	totalDescLength := 0

	for _, task := range tasks {
		totalDescLength += len(task.Description)

		if task.AgentRole != "" {
			summary["tasks_with_agent"] = summary["tasks_with_agent"].(int) + 1
		}

		if len(task.Tools) > 0 {
			summary["tasks_with_tools"] = summary["tasks_with_tools"].(int) + 1
		}

		if task.Metadata != nil {
			if _, hasKnowledge := task.Metadata["knowledge"]; hasKnowledge {
				summary["tasks_with_knowledge"] = summary["tasks_with_knowledge"].(int) + 1
			}
		}
	}

	if len(tasks) > 0 {
		summary["average_desc_length"] = totalDescLength / len(tasks)
	}

	return summary
}
