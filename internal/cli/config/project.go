package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ProjectType 项目类型
type ProjectType string

const (
	ProjectTypeCrew ProjectType = "crew"
	ProjectTypeFlow ProjectType = "flow"
)

// ProjectConfig 项目配置结构
type ProjectConfig struct {
	Name        string      `yaml:"name"`
	Type        ProjectType `yaml:"type"`
	Description string      `yaml:"description,omitempty"`
	Version     string      `yaml:"version"`
	Author      string      `yaml:"author,omitempty"`
	CreatedAt   string      `yaml:"created_at"`

	// Go特定配置
	GoModule  string `yaml:"go_module"`
	GoVersion string `yaml:"go_version"`

	// Crew特定配置
	Agents []AgentConfig `yaml:"agents,omitempty"`
	Tasks  []TaskConfig  `yaml:"tasks,omitempty"`

	// LLM配置
	LLM LLMConfig `yaml:"llm"`

	// 依赖配置
	Dependencies []string `yaml:"dependencies,omitempty"`
}

// AgentConfig Agent配置
type AgentConfig struct {
	Name      string   `yaml:"name"`
	Role      string   `yaml:"role"`
	Goal      string   `yaml:"goal"`
	Backstory string   `yaml:"backstory"`
	Tools     []string `yaml:"tools,omitempty"`
	LLM       string   `yaml:"llm,omitempty"`
	Verbose   bool     `yaml:"verbose,omitempty"`
}

// TaskConfig Task配置
type TaskConfig struct {
	Name           string   `yaml:"name"`
	Description    string   `yaml:"description"`
	ExpectedOutput string   `yaml:"expected_output"`
	Agent          string   `yaml:"agent"`
	Context        []string `yaml:"context,omitempty"`
	Tools          []string `yaml:"tools,omitempty"`
	OutputFormat   string   `yaml:"output_format,omitempty"`
	OutputFile     string   `yaml:"output_file,omitempty"`
}

// LLMConfig LLM配置
type LLMConfig struct {
	Provider    string  `yaml:"provider"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature,omitempty"`
	MaxTokens   int     `yaml:"max_tokens,omitempty"`
	BaseURL     string  `yaml:"base_url,omitempty"`
}

// LoadProjectConfig 加载项目配置
func LoadProjectConfig(configPath string) (*ProjectConfig, error) {
	if configPath == "" {
		configPath = "greensoulai.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ProjectConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 设置默认值
	if config.GoVersion == "" {
		config.GoVersion = "1.21"
	}

	if config.LLM.Provider == "" {
		config.LLM.Provider = "openai"
		config.LLM.Model = "gpt-4o-mini"
		config.LLM.Temperature = 0.7
	}

	return &config, nil
}

// SaveProjectConfig 保存项目配置
func (pc *ProjectConfig) SaveProjectConfig(configPath string) error {
	if configPath == "" {
		configPath = "greensoulai.yaml"
	}

	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(pc)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate 验证配置
func (pc *ProjectConfig) Validate() error {
	if pc.Name == "" {
		return fmt.Errorf("project name is required")
	}

	if pc.Type != ProjectTypeCrew && pc.Type != ProjectTypeFlow {
		return fmt.Errorf("invalid project type: %s", pc.Type)
	}

	if pc.GoModule == "" {
		return fmt.Errorf("go module is required")
	}

	// 验证Agent配置
	agentNames := make(map[string]bool)
	for _, agent := range pc.Agents {
		if agent.Name == "" {
			return fmt.Errorf("agent name is required")
		}
		if agent.Role == "" {
			return fmt.Errorf("agent role is required for agent %s", agent.Name)
		}
		if agent.Goal == "" {
			return fmt.Errorf("agent goal is required for agent %s", agent.Name)
		}
		if agentNames[agent.Name] {
			return fmt.Errorf("duplicate agent name: %s", agent.Name)
		}
		agentNames[agent.Name] = true
	}

	// 验证Task配置
	taskNames := make(map[string]bool)
	for _, task := range pc.Tasks {
		if task.Name == "" {
			return fmt.Errorf("task name is required")
		}
		if task.Description == "" {
			return fmt.Errorf("task description is required for task %s", task.Name)
		}
		if task.Agent != "" && !agentNames[task.Agent] {
			return fmt.Errorf("task %s references unknown agent: %s", task.Name, task.Agent)
		}
		if taskNames[task.Name] {
			return fmt.Errorf("duplicate task name: %s", task.Name)
		}
		taskNames[task.Name] = true
	}

	return nil
}

// GetProjectRoot 获取项目根目录
func GetProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// 向上查找项目配置文件
	dir := cwd
	for {
		configPath := filepath.Join(dir, "greensoulai.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // 到达根目录
		}
		dir = parent
	}

	return "", fmt.Errorf("not in a greensoulai project directory")
}

// DefaultCrewProjectConfig 默认Crew项目配置
func DefaultCrewProjectConfig(name, module string) *ProjectConfig {
	return &ProjectConfig{
		Name:        name,
		Type:        ProjectTypeCrew,
		Description: fmt.Sprintf("%s crew project", name),
		Version:     "1.0.0",
		GoModule:    module,
		GoVersion:   "1.21",
		LLM: LLMConfig{
			Provider:    "openai",
			Model:       "gpt-4o-mini",
			Temperature: 0.7,
			MaxTokens:   4096,
		},
		Agents: []AgentConfig{
			{
				Name:      "researcher",
				Role:      "高级研究员",
				Goal:      "进行深入的研究和分析",
				Backstory: "你是一位经验丰富的研究专家，擅长收集、分析和总结信息。",
				Tools:     []string{"search_tool", "analysis_tool"},
				Verbose:   true,
			},
		},
		Tasks: []TaskConfig{
			{
				Name:           "research_task",
				Description:    "进行主题研究",
				ExpectedOutput: "详细的研究报告",
				Agent:          "researcher",
				OutputFormat:   "markdown",
				OutputFile:     "research_report.md",
			},
		},
		Dependencies: []string{
			"github.com/ynl/greensoulai",
		},
	}
}

// DefaultFlowProjectConfig 默认Flow项目配置
func DefaultFlowProjectConfig(name, module string) *ProjectConfig {
	return &ProjectConfig{
		Name:        name,
		Type:        ProjectTypeFlow,
		Description: fmt.Sprintf("%s flow project", name),
		Version:     "1.0.0",
		GoModule:    module,
		GoVersion:   "1.21",
		LLM: LLMConfig{
			Provider:    "openai",
			Model:       "gpt-4o-mini",
			Temperature: 0.7,
			MaxTokens:   4096,
		},
		Dependencies: []string{
			"github.com/ynl/greensoulai",
		},
	}
}
