package long_term

import (
	"time"
)

// LongTermMemoryItem 长期记忆项目，对应Python版本的LongTermMemoryItem
// 保持与Python版本的业务逻辑一致
type LongTermMemoryItem struct {
	Agent          string                 `json:"agent"`           // Agent角色
	Task           string                 `json:"task"`            // 任务描述
	ExpectedOutput string                 `json:"expected_output"` // 期望输出
	DateTime       string                 `json:"datetime"`        // 时间戳（字符串格式，与Python版本一致）
	Quality        *float64               `json:"quality"`         // 质量评分（可选）
	Metadata       map[string]interface{} `json:"metadata"`        // 元数据

	// Go版本额外字段（用于内部管理，不影响业务逻辑）
	ID        string    `json:"id,omitempty"`         // 内部ID
	CreatedAt time.Time `json:"created_at,omitempty"` // 创建时间
	Score     float64   `json:"score,omitempty"`      // 内部评分
}

// NewLongTermMemoryItem 创建新的长期记忆项，遵循Python版本的构造逻辑
func NewLongTermMemoryItem(
	agent string,
	task string,
	expectedOutput string,
	datetime string,
	quality *float64,
	metadata map[string]interface{},
) *LongTermMemoryItem {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	return &LongTermMemoryItem{
		Agent:          agent,
		Task:           task,
		ExpectedOutput: expectedOutput,
		DateTime:       datetime,
		Quality:        quality,
		Metadata:       metadata,
		CreatedAt:      time.Now(),
	}
}

// ToDict 转换为字典格式，便于存储和序列化
func (item *LongTermMemoryItem) ToDict() map[string]interface{} {
	result := map[string]interface{}{
		"agent":           item.Agent,
		"task":            item.Task,
		"expected_output": item.ExpectedOutput,
		"datetime":        item.DateTime,
		"metadata":        item.Metadata,
	}

	if item.Quality != nil {
		result["quality"] = *item.Quality
	}

	return result
}

// FromDict 从字典创建LongTermMemoryItem
func FromDict(data map[string]interface{}) *LongTermMemoryItem {
	item := &LongTermMemoryItem{
		Metadata: make(map[string]interface{}),
	}

	if agent, ok := data["agent"].(string); ok {
		item.Agent = agent
	}

	if task, ok := data["task"].(string); ok {
		item.Task = task
	}

	if expectedOutput, ok := data["expected_output"].(string); ok {
		item.ExpectedOutput = expectedOutput
	}

	if datetime, ok := data["datetime"].(string); ok {
		item.DateTime = datetime
	}

	if quality, ok := data["quality"]; ok {
		switch v := quality.(type) {
		case float64:
			item.Quality = &v
		case int:
			f := float64(v)
			item.Quality = &f
		}
	}

	if metadata, ok := data["metadata"].(map[string]interface{}); ok {
		item.Metadata = metadata
	}

	return item
}

