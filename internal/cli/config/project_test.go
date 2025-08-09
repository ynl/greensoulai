package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *ProjectConfig
		wantErr bool
	}{
		{
			name: "valid crew project",
			config: &ProjectConfig{
				Name:      "test-project",
				Type:      ProjectTypeCrew,
				GoModule:  "github.com/user/test-project",
				GoVersion: "1.21",
				Agents: []AgentConfig{
					{
						Name:      "researcher",
						Role:      "Research Expert",
						Goal:      "Conduct research",
						Backstory: "Experienced researcher",
					},
				},
				Tasks: []TaskConfig{
					{
						Name:           "research_task",
						Description:    "Conduct research",
						ExpectedOutput: "Research report",
						Agent:          "researcher",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing project name",
			config: &ProjectConfig{
				Type:      ProjectTypeCrew,
				GoModule:  "github.com/user/test",
				GoVersion: "1.21",
			},
			wantErr: true,
		},
		{
			name: "invalid project type",
			config: &ProjectConfig{
				Name:      "test-project",
				Type:      "invalid",
				GoModule:  "github.com/user/test-project",
				GoVersion: "1.21",
			},
			wantErr: true,
		},
		{
			name: "missing go module",
			config: &ProjectConfig{
				Name:      "test-project",
				Type:      ProjectTypeCrew,
				GoVersion: "1.21",
			},
			wantErr: true,
		},
		{
			name: "agent with missing role",
			config: &ProjectConfig{
				Name:      "test-project",
				Type:      ProjectTypeCrew,
				GoModule:  "github.com/user/test-project",
				GoVersion: "1.21",
				Agents: []AgentConfig{
					{
						Name:      "researcher",
						Goal:      "Conduct research",
						Backstory: "Experienced researcher",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "task with unknown agent",
			config: &ProjectConfig{
				Name:      "test-project",
				Type:      ProjectTypeCrew,
				GoModule:  "github.com/user/test-project",
				GoVersion: "1.21",
				Agents: []AgentConfig{
					{
						Name:      "researcher",
						Role:      "Research Expert",
						Goal:      "Conduct research",
						Backstory: "Experienced researcher",
					},
				},
				Tasks: []TaskConfig{
					{
						Name:           "research_task",
						Description:    "Conduct research",
						ExpectedOutput: "Research report",
						Agent:          "unknown_agent",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ProjectConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadProjectConfig(t *testing.T) {
	// 创建临时目录和配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "greensoulai.yaml")

	// 测试配置内容
	configContent := `name: test-project
type: crew
description: Test project
version: 1.0.0
go_module: github.com/user/test-project
go_version: "1.21"
created_at: "2024-01-01T00:00:00Z"

agents:
  - name: researcher
    role: Research Expert
    goal: Conduct comprehensive research
    backstory: You are an experienced researcher
    tools:
      - search_tool
    verbose: true

tasks:
  - name: research_task
    description: Conduct research on the topic
    expected_output: Detailed research report
    agent: researcher
    output_format: markdown
    output_file: research_output.md

llm:
  provider: openai
  model: gpt-4o-mini
  temperature: 0.7
  max_tokens: 4096

dependencies:
  - github.com/ynl/greensoulai
`

	// 写入配置文件
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// 测试加载配置
	config, err := LoadProjectConfig(configPath)
	if err != nil {
		t.Fatalf("LoadProjectConfig() error = %v", err)
	}

	// 验证配置内容
	if config.Name != "test-project" {
		t.Errorf("Expected name 'test-project', got '%s'", config.Name)
	}

	if config.Type != ProjectTypeCrew {
		t.Errorf("Expected type '%s', got '%s'", ProjectTypeCrew, config.Type)
	}

	if config.GoModule != "github.com/user/test-project" {
		t.Errorf("Expected module 'github.com/user/test-project', got '%s'", config.GoModule)
	}

	if len(config.Agents) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(config.Agents))
	}

	if config.Agents[0].Name != "researcher" {
		t.Errorf("Expected agent name 'researcher', got '%s'", config.Agents[0].Name)
	}

	if len(config.Tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(config.Tasks))
	}

	if config.Tasks[0].Name != "research_task" {
		t.Errorf("Expected task name 'research_task', got '%s'", config.Tasks[0].Name)
	}

	if config.LLM.Provider != "openai" {
		t.Errorf("Expected LLM provider 'openai', got '%s'", config.LLM.Provider)
	}
}

func TestSaveProjectConfig(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.yaml")

	// 创建测试配置
	config := &ProjectConfig{
		Name:        "test-save",
		Type:        ProjectTypeCrew,
		Description: "Test save configuration",
		Version:     "1.0.0",
		GoModule:    "github.com/user/test-save",
		GoVersion:   "1.21",
		CreatedAt:   time.Now().Format(time.RFC3339),
		LLM: LLMConfig{
			Provider:    "openai",
			Model:       "gpt-4o-mini",
			Temperature: 0.7,
			MaxTokens:   4096,
		},
		Agents: []AgentConfig{
			{
				Name:      "test-agent",
				Role:      "Test Role",
				Goal:      "Test goal",
				Backstory: "Test backstory",
				Verbose:   true,
			},
		},
		Tasks: []TaskConfig{
			{
				Name:           "test-task",
				Description:    "Test description",
				ExpectedOutput: "Test output",
				Agent:          "test-agent",
				OutputFormat:   "markdown",
			},
		},
		Dependencies: []string{"github.com/ynl/greensoulai"},
	}

	// 保存配置
	if err := config.SaveProjectConfig(configPath); err != nil {
		t.Fatalf("SaveProjectConfig() error = %v", err)
	}

	// 验证文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Config file was not created")
	}

	// 重新加载配置验证内容
	loadedConfig, err := LoadProjectConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}

	if loadedConfig.Name != config.Name {
		t.Errorf("Expected name '%s', got '%s'", config.Name, loadedConfig.Name)
	}

	if loadedConfig.Type != config.Type {
		t.Errorf("Expected type '%s', got '%s'", config.Type, loadedConfig.Type)
	}
}

func TestDefaultCrewProjectConfig(t *testing.T) {
	name := "test-crew"
	module := "github.com/user/test-crew"

	config := DefaultCrewProjectConfig(name, module)

	if config.Name != name {
		t.Errorf("Expected name '%s', got '%s'", name, config.Name)
	}

	if config.Type != ProjectTypeCrew {
		t.Errorf("Expected type '%s', got '%s'", ProjectTypeCrew, config.Type)
	}

	if config.GoModule != module {
		t.Errorf("Expected module '%s', got '%s'", module, config.GoModule)
	}

	if config.GoVersion != "1.21" {
		t.Errorf("Expected Go version '1.21', got '%s'", config.GoVersion)
	}

	if len(config.Agents) == 0 {
		t.Error("Expected at least one default agent")
	}

	if len(config.Tasks) == 0 {
		t.Error("Expected at least one default task")
	}

	if config.LLM.Provider != "openai" {
		t.Errorf("Expected LLM provider 'openai', got '%s'", config.LLM.Provider)
	}

	// 验证配置有效性
	if err := config.Validate(); err != nil {
		t.Errorf("Default config validation failed: %v", err)
	}
}

func TestDefaultFlowProjectConfig(t *testing.T) {
	name := "test-flow"
	module := "github.com/user/test-flow"

	config := DefaultFlowProjectConfig(name, module)

	if config.Name != name {
		t.Errorf("Expected name '%s', got '%s'", name, config.Name)
	}

	if config.Type != ProjectTypeFlow {
		t.Errorf("Expected type '%s', got '%s'", ProjectTypeFlow, config.Type)
	}

	if config.GoModule != module {
		t.Errorf("Expected module '%s', got '%s'", module, config.GoModule)
	}

	if config.GoVersion != "1.21" {
		t.Errorf("Expected Go version '1.21', got '%s'", config.GoVersion)
	}

	if config.LLM.Provider != "openai" {
		t.Errorf("Expected LLM provider 'openai', got '%s'", config.LLM.Provider)
	}
}

func TestGetProjectRoot(t *testing.T) {
	// 创建临时目录结构
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-project")
	subDir := filepath.Join(projectDir, "subdir")

	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}

	// 创建配置文件
	configPath := filepath.Join(projectDir, "greensoulai.yaml")
	configContent := `name: test
type: crew
go_module: github.com/user/test
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// 保存当前目录
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// 测试从项目根目录
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("Failed to change to project directory: %v", err)
	}

	root, err := GetProjectRoot()
	if err != nil {
		t.Errorf("GetProjectRoot() from project root error = %v", err)
	}

	expectedRoot, _ := filepath.Abs(projectDir)
	// 规范化路径以处理符号链接（如macOS上的/tmp -> /private/tmp）
	expectedRoot, _ = filepath.EvalSymlinks(expectedRoot)
	root, _ = filepath.EvalSymlinks(root)

	if root != expectedRoot {
		t.Errorf("Expected root '%s', got '%s'", expectedRoot, root)
	}

	// 测试从子目录
	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("Failed to change to sub directory: %v", err)
	}

	root, err = GetProjectRoot()
	if err != nil {
		t.Errorf("GetProjectRoot() from subdirectory error = %v", err)
	}

	// 同样规范化路径
	root, _ = filepath.EvalSymlinks(root)
	if root != expectedRoot {
		t.Errorf("Expected root '%s', got '%s'", expectedRoot, root)
	}

	// 测试从非项目目录
	nonProjectDir := t.TempDir()
	if err := os.Chdir(nonProjectDir); err != nil {
		t.Fatalf("Failed to change to non-project directory: %v", err)
	}

	_, err = GetProjectRoot()
	if err == nil {
		t.Error("Expected error when not in project directory")
	}
}
