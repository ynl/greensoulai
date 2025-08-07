package async

import (
	"time"
)

// Result 异步执行结果
type Result struct {
	Value    interface{}
	Error    error
	Duration time.Duration
}

// TaskResult 任务执行结果
type TaskResult struct {
	Output   *TaskOutput
	Error    error
	Duration time.Duration
}

// TaskOutput 任务输出
type TaskOutput struct {
	Raw           string                 `json:"raw"`
	JSON          map[string]interface{} `json:"json,omitempty"`
	Agent         string                 `json:"agent"`
	Description   string                 `json:"description"`
	Summary       string                 `json:"summary"`
	CreatedAt     time.Time              `json:"created_at"`
	ExecutionTime time.Duration          `json:"execution_time"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// NewTaskOutput 创建新的任务输出
func NewTaskOutput(raw string, agent string, description string) *TaskOutput {
	return &TaskOutput{
		Raw:         raw,
		Agent:       agent,
		Description: description,
		CreatedAt:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}
}
