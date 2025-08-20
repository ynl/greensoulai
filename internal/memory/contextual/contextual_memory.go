package contextual

import (
	"context"
	"fmt"
	"strings"

	"github.com/ynl/greensoulai/internal/agent"
	"github.com/ynl/greensoulai/internal/memory/entity"
	"github.com/ynl/greensoulai/internal/memory/external"
	"github.com/ynl/greensoulai/internal/memory/long_term"
	"github.com/ynl/greensoulai/internal/memory/short_term"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// ContextualMemory 上下文记忆系统
// 参考crewAI的ContextualMemory实现，统一管理所有记忆类型
// 自动构建任务相关的最小且高相关性的上下文信息
type ContextualMemory struct {
	// 各种记忆类型实例
	stm *short_term.ShortTermMemory // 短期记忆
	ltm *long_term.LongTermMemory   // 长期记忆
	em  *entity.EntityMemory        // 实体记忆
	exm *external.ExternalMemory    // 外部记忆

	// 基础设施
	eventBus events.EventBus
	logger   logger.Logger

	// 配置选项
	config ContextualMemoryConfig
}

// ContextualMemoryConfig 上下文记忆配置
type ContextualMemoryConfig struct {
	// 默认搜索限制
	DefaultSTMLimit      int `json:"default_stm_limit"`      // 短期记忆默认搜索数量
	DefaultLTMLimit      int `json:"default_ltm_limit"`      // 长期记忆默认搜索数量
	DefaultEntityLimit   int `json:"default_entity_limit"`   // 实体记忆默认搜索数量
	DefaultExternalLimit int `json:"default_external_limit"` // 外部记忆默认搜索数量

	// 搜索阈值
	STMScoreThreshold      float64 `json:"stm_score_threshold"`      // 短期记忆分数阈值
	EntityScoreThreshold   float64 `json:"entity_score_threshold"`   // 实体记忆分数阈值
	ExternalScoreThreshold float64 `json:"external_score_threshold"` // 外部记忆分数阈值

	// 上下文组装选项
	EnableFormatting     bool `json:"enable_formatting"`      // 启用格式化输出
	EnableSectionHeaders bool `json:"enable_section_headers"` // 启用章节标题
	MaxContextLength     int  `json:"max_context_length"`     // 最大上下文长度

	// 过滤选项
	FilterEmptyResults  bool `json:"filter_empty_results"` // 过滤空结果
	EnableDeduplication bool `json:"enable_deduplication"` // 启用去重
}

// DefaultContextualMemoryConfig 默认上下文记忆配置
func DefaultContextualMemoryConfig() ContextualMemoryConfig {
	return ContextualMemoryConfig{
		// 默认搜索限制（参考crewAI默认值）
		DefaultSTMLimit:      3,
		DefaultLTMLimit:      2,
		DefaultEntityLimit:   3,
		DefaultExternalLimit: 3,

		// 搜索阈值
		STMScoreThreshold:      0.35,
		EntityScoreThreshold:   0.35,
		ExternalScoreThreshold: 0.35,

		// 上下文组装选项
		EnableFormatting:     true,
		EnableSectionHeaders: true,
		MaxContextLength:     8000, // 避免上下文过长

		// 过滤选项
		FilterEmptyResults:  true,
		EnableDeduplication: true,
	}
}

// NewContextualMemory 创建上下文记忆实例
func NewContextualMemory(
	stm *short_term.ShortTermMemory,
	ltm *long_term.LongTermMemory,
	em *entity.EntityMemory,
	exm *external.ExternalMemory,
	eventBus events.EventBus,
	logger logger.Logger,
	config *ContextualMemoryConfig,
) *ContextualMemory {
	// 使用默认配置如果没有提供
	var cfg ContextualMemoryConfig
	if config != nil {
		cfg = *config
	} else {
		cfg = DefaultContextualMemoryConfig()
	}

	return &ContextualMemory{
		stm:      stm,
		ltm:      ltm,
		em:       em,
		exm:      exm,
		eventBus: eventBus,
		logger:   logger,
		config:   cfg,
	}
}

// BuildContextForTask 为任务构建上下文信息
// 这是核心方法，参考crewAI的build_context_for_task实现
// 自动构建最小且高相关性的上下文信息集合
func (cm *ContextualMemory) BuildContextForTask(ctx context.Context, task agent.Task, context string) (string, error) {
	// 构建查询字符串，与crewAI逻辑保持一致
	query := strings.TrimSpace(fmt.Sprintf("%s %s", task.GetDescription(), context))

	cm.logger.Debug("building context for task",
		logger.Field{Key: "task_id", Value: task.GetID()},
		logger.Field{Key: "query", Value: query},
	)

	// 发射上下文构建开始事件
	if cm.eventBus != nil {
		startEvent := NewContextBuildStartedEvent(task.GetID(), query)
		cm.eventBus.Emit(ctx, cm, startEvent)
	}

	if query == "" {
		cm.logger.Debug("empty query provided for context building")
		// 即使查询为空，也要发射完成事件
		if cm.eventBus != nil {
			completedEvent := NewContextBuildCompletedEvent(task.GetID(), query, 0, 0)
			cm.eventBus.Emit(ctx, cm, completedEvent)
		}
		return "", nil
	}

	var contextParts []string

	// 按照crewAI的顺序获取各类记忆上下文
	// 1. 长期记忆上下文（历史数据）
	if ltmContext, err := cm.fetchLTMContext(ctx, task.GetDescription()); err == nil && ltmContext != "" {
		contextParts = append(contextParts, ltmContext)
	} else if err != nil {
		cm.logger.Warn("failed to fetch LTM context", logger.Field{Key: "error", Value: err})
	}

	// 2. 短期记忆上下文（最近洞察）
	if stmContext, err := cm.fetchSTMContext(ctx, query); err == nil && stmContext != "" {
		contextParts = append(contextParts, stmContext)
	} else if err != nil {
		cm.logger.Warn("failed to fetch STM context", logger.Field{Key: "error", Value: err})
	}

	// 3. 实体记忆上下文
	if entityContext, err := cm.fetchEntityContext(ctx, query); err == nil && entityContext != "" {
		contextParts = append(contextParts, entityContext)
	} else if err != nil {
		cm.logger.Warn("failed to fetch entity context", logger.Field{Key: "error", Value: err})
	}

	// 4. 外部记忆上下文
	if externalContext, err := cm.fetchExternalContext(ctx, query); err == nil && externalContext != "" {
		contextParts = append(contextParts, externalContext)
	} else if err != nil {
		cm.logger.Warn("failed to fetch external context", logger.Field{Key: "error", Value: err})
	}

	// 过滤和去重
	if cm.config.FilterEmptyResults {
		contextParts = cm.filterEmptyParts(contextParts)
	}

	if cm.config.EnableDeduplication {
		contextParts = cm.deduplicateParts(contextParts)
	}

	// 组装最终上下文
	finalContext := strings.Join(contextParts, "\n")

	// 检查长度限制
	if cm.config.MaxContextLength > 0 && len(finalContext) > cm.config.MaxContextLength {
		finalContext = finalContext[:cm.config.MaxContextLength]
		cm.logger.Warn("context truncated due to length limit",
			logger.Field{Key: "original_length", Value: len(strings.Join(contextParts, "\n"))},
			logger.Field{Key: "max_length", Value: cm.config.MaxContextLength},
		)
	}

	// 发射上下文构建完成事件
	if cm.eventBus != nil {
		completedEvent := NewContextBuildCompletedEvent(task.GetID(), query, len(contextParts), len(finalContext))
		cm.eventBus.Emit(ctx, cm, completedEvent)
	}

	cm.logger.Debug("context built successfully",
		logger.Field{Key: "task_id", Value: task.GetID()},
		logger.Field{Key: "context_parts", Value: len(contextParts)},
		logger.Field{Key: "final_length", Value: len(finalContext)},
	)

	return finalContext, nil
}

// fetchSTMContext 获取短期记忆上下文
// 参考crewAI的_fetch_stm_context实现
func (cm *ContextualMemory) fetchSTMContext(ctx context.Context, query string) (string, error) {
	if cm.stm == nil {
		return "", nil
	}

	cm.logger.Debug("fetching STM context", logger.Field{Key: "query", Value: query})

	// 搜索短期记忆
	results, err := cm.stm.Search(ctx, query, cm.config.DefaultSTMLimit, cm.config.STMScoreThreshold)
	if err != nil {
		return "", fmt.Errorf("STM search failed: %w", err)
	}

	if len(results) == 0 {
		cm.logger.Debug("no STM results found")
		return "", nil
	}

	// 格式化结果为项目符号列表，与crewAI格式保持一致
	var formattedResults []string
	for _, result := range results {
		// 尝试从metadata中获取context字段（与crewAI保持一致）
		if contextStr, ok := result.Metadata["context"].(string); ok {
			formattedResults = append(formattedResults, fmt.Sprintf("- %s", contextStr))
		} else {
			// 回退到使用Value字段
			formattedResults = append(formattedResults, fmt.Sprintf("- %v", result.Value))
		}
	}

	if len(formattedResults) == 0 {
		return "", nil
	}

	// 添加章节标题（与crewAI格式保持一致）
	header := ""
	if cm.config.EnableSectionHeaders {
		header = "Recent Insights:\n"
	}

	context := header + strings.Join(formattedResults, "\n")

	cm.logger.Debug("STM context fetched",
		logger.Field{Key: "results_count", Value: len(results)},
		logger.Field{Key: "formatted_count", Value: len(formattedResults)},
	)

	return context, nil
}

// fetchLTMContext 获取长期记忆上下文
// 参考crewAI的_fetch_ltm_context实现
func (cm *ContextualMemory) fetchLTMContext(ctx context.Context, task string) (string, error) {
	if cm.ltm == nil {
		return "", nil
	}

	cm.logger.Debug("fetching LTM context", logger.Field{Key: "task", Value: task})

	// 使用长期记忆的专用搜索方法
	results, err := cm.ltm.Search(ctx, task, cm.config.DefaultLTMLimit)
	if err != nil {
		return "", fmt.Errorf("LTM search failed: %w", err)
	}

	if len(results) == 0 {
		cm.logger.Debug("no LTM results found")
		return "", nil
	}

	// 提取建议列表，与crewAI逻辑保持一致
	var suggestions []string
	for _, result := range results {
		if metadata, ok := result["metadata"].(map[string]interface{}); ok {
			if suggestionsList, ok := metadata["suggestions"].([]interface{}); ok {
				for _, suggestion := range suggestionsList {
					if suggestionStr, ok := suggestion.(string); ok {
						suggestions = append(suggestions, suggestionStr)
					}
				}
			}
		}
	}

	// 去重处理，与crewAI保持一致
	suggestions = cm.removeDuplicateStrings(suggestions)

	if len(suggestions) == 0 {
		return "", nil
	}

	// 格式化为项目符号列表
	formattedResults := make([]string, len(suggestions))
	for i, suggestion := range suggestions {
		formattedResults[i] = fmt.Sprintf("- %s", suggestion)
	}

	// 添加章节标题
	header := ""
	if cm.config.EnableSectionHeaders {
		header = "Historical Data:\n"
	}

	context := header + strings.Join(formattedResults, "\n")

	cm.logger.Debug("LTM context fetched",
		logger.Field{Key: "results_count", Value: len(results)},
		logger.Field{Key: "suggestions_count", Value: len(suggestions)},
	)

	return context, nil
}

// fetchEntityContext 获取实体记忆上下文
// 参考crewAI的_fetch_entity_context实现
func (cm *ContextualMemory) fetchEntityContext(ctx context.Context, query string) (string, error) {
	if cm.em == nil {
		return "", nil
	}

	cm.logger.Debug("fetching entity context", logger.Field{Key: "query", Value: query})

	// 搜索实体记忆
	results, err := cm.em.Search(ctx, query, cm.config.DefaultEntityLimit, cm.config.EntityScoreThreshold)
	if err != nil {
		return "", fmt.Errorf("entity memory search failed: %w", err)
	}

	if len(results) == 0 {
		cm.logger.Debug("no entity results found")
		return "", nil
	}

	// 格式化结果，与crewAI格式保持一致
	var formattedResults []string
	for _, result := range results {
		// 尝试从metadata中获取context字段
		if contextStr, ok := result.Metadata["context"].(string); ok {
			formattedResults = append(formattedResults, fmt.Sprintf("- %s", contextStr))
		} else {
			// 回退到使用Value字段
			formattedResults = append(formattedResults, fmt.Sprintf("- %v", result.Value))
		}
	}

	if len(formattedResults) == 0 {
		return "", nil
	}

	// 添加章节标题
	header := ""
	if cm.config.EnableSectionHeaders {
		header = "Entities:\n"
	}

	context := header + strings.Join(formattedResults, "\n")

	cm.logger.Debug("entity context fetched",
		logger.Field{Key: "results_count", Value: len(results)},
		logger.Field{Key: "formatted_count", Value: len(formattedResults)},
	)

	return context, nil
}

// fetchExternalContext 获取外部记忆上下文
// 参考crewAI的_fetch_external_context实现
func (cm *ContextualMemory) fetchExternalContext(ctx context.Context, query string) (string, error) {
	if cm.exm == nil {
		return "", nil
	}

	cm.logger.Debug("fetching external context", logger.Field{Key: "query", Value: query})

	// 搜索外部记忆
	results, err := cm.exm.Search(ctx, query, cm.config.DefaultExternalLimit, cm.config.ExternalScoreThreshold)
	if err != nil {
		return "", fmt.Errorf("external memory search failed: %w", err)
	}

	if len(results) == 0 {
		cm.logger.Debug("no external results found")
		return "", nil
	}

	// 格式化结果，与crewAI格式保持一致
	var formattedResults []string
	for _, result := range results {
		// 尝试从metadata中获取context字段
		if contextStr, ok := result.Metadata["context"].(string); ok {
			formattedResults = append(formattedResults, fmt.Sprintf("- %s", contextStr))
		} else {
			// 回退到使用Value字段
			formattedResults = append(formattedResults, fmt.Sprintf("- %v", result.Value))
		}
	}

	if len(formattedResults) == 0 {
		return "", nil
	}

	// 添加章节标题
	header := ""
	if cm.config.EnableSectionHeaders {
		header = "External memories:\n"
	}

	context := header + strings.Join(formattedResults, "\n")

	cm.logger.Debug("external context fetched",
		logger.Field{Key: "results_count", Value: len(results)},
		logger.Field{Key: "formatted_count", Value: len(formattedResults)},
	)

	return context, nil
}

// 辅助方法

// filterEmptyParts 过滤空的上下文部分
func (cm *ContextualMemory) filterEmptyParts(parts []string) []string {
	var filtered []string
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			filtered = append(filtered, part)
		}
	}
	return filtered
}

// deduplicateParts 去重上下文部分
func (cm *ContextualMemory) deduplicateParts(parts []string) []string {
	seen := make(map[string]bool)
	var deduplicated []string

	for _, part := range parts {
		if !seen[part] {
			seen[part] = true
			deduplicated = append(deduplicated, part)
		}
	}

	return deduplicated
}

// removeDuplicateStrings 去重字符串数组
func (cm *ContextualMemory) removeDuplicateStrings(strs []string) []string {
	keys := make(map[string]bool)
	result := make([]string, 0) // 确保返回非nil的空切片

	for _, str := range strs {
		if !keys[str] {
			keys[str] = true
			result = append(result, str)
		}
	}

	return result
}

// GetConfig 获取配置
func (cm *ContextualMemory) GetConfig() ContextualMemoryConfig {
	return cm.config
}

// UpdateConfig 更新配置
func (cm *ContextualMemory) UpdateConfig(config ContextualMemoryConfig) {
	cm.config = config
	cm.logger.Info("contextual memory config updated")
}

// GetMemoryInstances 获取记忆实例（用于测试和调试）
func (cm *ContextualMemory) GetMemoryInstances() (
	*short_term.ShortTermMemory,
	*long_term.LongTermMemory,
	*entity.EntityMemory,
	*external.ExternalMemory,
) {
	return cm.stm, cm.ltm, cm.em, cm.exm
}
