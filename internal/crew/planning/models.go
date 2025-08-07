package planning

import (
	"encoding/json"
	"fmt"
)

// PlanPerTask 单个任务的规划，对应Python版本的PlanPerTask
// 保持与Python版本的业务逻辑一致
type PlanPerTask struct {
	Task string `json:"task" validate:"required"` // 任务描述
	Plan string `json:"plan" validate:"required"` // 详细的步骤规划
}

// PlannerTaskPydanticOutput 规划任务的输出格式，对应Python版本的PlannerTaskPydanticOutput
// 保持与Python版本的数据结构一致
type PlannerTaskPydanticOutput struct {
	ListOfPlansPerTask []PlanPerTask `json:"list_of_plans_per_task" validate:"required"`
}

// TaskSummary 任务摘要，用于生成规划时的上下文信息
type TaskSummary struct {
	TaskNumber        int                    `json:"task_number"`
	Description       string                 `json:"task_description"`
	ExpectedOutput    string                 `json:"task_expected_output"`
	AgentRole         string                 `json:"agent"`
	AgentGoal         string                 `json:"agent_goal"`
	TaskTools         []string               `json:"task_tools"`
	AgentTools        []string               `json:"agent_tools"`
	AgentKnowledge    []string               `json:"agent_knowledge,omitempty"`
	AdditionalContext map[string]interface{} `json:"additional_context,omitempty"`
}

// PlanningRequest 规划请求，包含所有必要的规划参数
type PlanningRequest struct {
	Tasks         []TaskInfo             `json:"tasks" validate:"required"`
	PlanningLLM   string                 `json:"planning_llm,omitempty"`
	Context       map[string]interface{} `json:"context,omitempty"`
	MaxRetries    int                    `json:"max_retries,omitempty"`
	TimeoutSec    int                    `json:"timeout_sec,omitempty"`
	CustomPrompts map[string]string      `json:"custom_prompts,omitempty"`
}

// TaskInfo 任务信息，简化的任务表示
type TaskInfo struct {
	ID             string                 `json:"id"`
	Description    string                 `json:"description" validate:"required"`
	ExpectedOutput string                 `json:"expected_output" validate:"required"`
	AgentRole      string                 `json:"agent_role,omitempty"`
	AgentGoal      string                 `json:"agent_goal,omitempty"`
	Tools          []string               `json:"tools,omitempty"`
	Context        []string               `json:"context,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// PlanningResult 规划结果，包含规划输出和元数据
type PlanningResult struct {
	Output          PlannerTaskPydanticOutput `json:"output"`
	ExecutionTime   float64                   `json:"execution_time_ms"`
	TokensUsed      int                       `json:"tokens_used,omitempty"`
	ModelUsed       string                    `json:"model_used,omitempty"`
	Success         bool                      `json:"success"`
	ErrorMessage    string                    `json:"error_message,omitempty"`
	RetryCount      int                       `json:"retry_count"`
	PlanningAgentID string                    `json:"planning_agent_id,omitempty"`
}

// PlanningConfig 规划配置，对应Python版本的配置选项
type PlanningConfig struct {
	PlanningAgentLLM string                 `json:"planning_agent_llm"`
	MaxRetries       int                    `json:"max_retries"`
	TimeoutSeconds   int                    `json:"timeout_seconds"`
	EnableVerbose    bool                   `json:"enable_verbose"`
	CustomPrompts    map[string]string      `json:"custom_prompts,omitempty"`
	AdditionalConfig map[string]interface{} `json:"additional_config,omitempty"`
}

// DefaultPlanningConfig 默认规划配置，与Python版本保持一致
func DefaultPlanningConfig() *PlanningConfig {
	return &PlanningConfig{
		PlanningAgentLLM: "gpt-4o-mini", // 与Python版本默认值一致
		MaxRetries:       3,
		TimeoutSeconds:   300, // 5分钟
		EnableVerbose:    false,
		CustomPrompts:    make(map[string]string),
		AdditionalConfig: make(map[string]interface{}),
	}
}

// ToJSON 将对象转换为JSON字符串
func (p *PlanPerTask) ToJSON() (string, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToJSON 将对象转换为JSON字符串
func (p *PlannerTaskPydanticOutput) ToJSON() (string, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON 从JSON字符串创建PlannerTaskPydanticOutput
func (p *PlannerTaskPydanticOutput) FromJSON(jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), p)
}

// Validate 验证规划输出的有效性
func (p *PlannerTaskPydanticOutput) Validate() error {
	if len(p.ListOfPlansPerTask) == 0 {
		return ErrEmptyPlanList
	}

	for i, plan := range p.ListOfPlansPerTask {
		if plan.Task == "" {
			return &PlanValidationError{
				Field:   "Task",
				Index:   i,
				Message: "task description cannot be empty",
			}
		}
		if plan.Plan == "" {
			return &PlanValidationError{
				Field:   "Plan",
				Index:   i,
				Message: "plan cannot be empty",
			}
		}
	}

	return nil
}

// GetTaskCount 获取任务数量
func (p *PlannerTaskPydanticOutput) GetTaskCount() int {
	return len(p.ListOfPlansPerTask)
}

// GetPlanByTaskDescription 根据任务描述查找计划
func (p *PlannerTaskPydanticOutput) GetPlanByTaskDescription(taskDescription string) (*PlanPerTask, bool) {
	for i := range p.ListOfPlansPerTask {
		if p.ListOfPlansPerTask[i].Task == taskDescription {
			return &p.ListOfPlansPerTask[i], true
		}
	}
	return nil, false
}

// AddPlan 添加新的任务计划
func (p *PlannerTaskPydanticOutput) AddPlan(plan PlanPerTask) {
	p.ListOfPlansPerTask = append(p.ListOfPlansPerTask, plan)
}

// String 返回可读的字符串表示
func (p *PlannerTaskPydanticOutput) String() string {
	result := "Planning Output:\n"
	for i, plan := range p.ListOfPlansPerTask {
		result += fmt.Sprintf("Task %d: %s\nPlan: %s\n\n", i+1, plan.Task, plan.Plan)
	}
	return result
}
