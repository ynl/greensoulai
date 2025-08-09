package commands

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/ynl/greensoulai/pkg/logger"
)

func TestNewCreateCommand(t *testing.T) {
	log := logger.NewTestLogger()
	cmd := NewCreateCommand(log)

	if cmd.Use != "create" {
		t.Errorf("Expected command use 'create', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected command to have short description")
	}

	// 检查子命令是否存在
	expectedSubcommands := []string{"crew", "flow", "agent", "task", "tool"}
	actualSubcommands := make(map[string]bool)

	for _, subcmd := range cmd.Commands() {
		// 提取命令名称（去掉参数部分）
		cmdName := strings.Split(subcmd.Use, " ")[0]
		actualSubcommands[cmdName] = true
	}

	for _, expected := range expectedSubcommands {
		if !actualSubcommands[expected] {
			t.Errorf("Expected subcommand '%s' not found", expected)
		}
	}
}

func TestCreateCrewCommand(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	log := logger.NewTestLogger()
	createCmd := NewCreateCommand(log)

	// 查找crew子命令
	var crewCmd *cobra.Command
	for _, cmd := range createCmd.Commands() {
		if cmd.Use == "crew [name]" {
			crewCmd = cmd
			break
		}
	}

	if crewCmd == nil {
		t.Fatal("crew subcommand not found")
	}

	tests := []struct {
		name        string
		args        []string
		flags       map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid crew creation",
			args:        []string{"test-crew"},
			flags:       map[string]string{},
			expectError: false,
		},
		{
			name:        "crew with custom module",
			args:        []string{"my-crew"},
			flags:       map[string]string{"module": "github.com/myuser/my-crew"},
			expectError: false,
		},
		{
			name:        "invalid project name",
			args:        []string{"123invalid"},
			flags:       map[string]string{},
			expectError: true,
			errorMsg:    "invalid project name",
		},
		{
			name:        "no arguments",
			args:        []string{},
			flags:       map[string]string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建命令副本
			testCmd := &cobra.Command{
				Use:  crewCmd.Use,
				Args: crewCmd.Args,
				RunE: crewCmd.RunE,
			}

			// 复制flags
			testCmd.Flags().AddFlagSet(crewCmd.Flags())

			// 设置flags
			for flag, value := range tt.flags {
				if err := testCmd.Flags().Set(flag, value); err != nil {
					t.Fatalf("Failed to set flag %s: %v", flag, err)
				}
			}

			// 创建缓冲区捕获输出
			var buf bytes.Buffer
			testCmd.SetOut(&buf)
			testCmd.SetErr(&buf)

			// 执行命令
			testCmd.SetArgs(tt.args)
			err := testCmd.ExecuteContext(context.Background())

			// 检查错误
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tt.expectError && err != nil && tt.errorMsg != "" {
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			}

			// 如果没有错误，检查项目是否创建成功
			if !tt.expectError && err == nil && len(tt.args) > 0 {
				projectDir := tt.args[0]
				if tt.flags["output"] != "" {
					projectDir = tt.flags["output"]
				}

				// 检查目录是否存在
				if _, err := os.Stat(projectDir); os.IsNotExist(err) {
					t.Errorf("Project directory '%s' was not created", projectDir)
				}

				// 检查关键文件是否存在
				expectedFiles := []string{
					"greensoulai.yaml",
					"go.mod",
					"README.md",
					"Makefile",
					".env.example",
				}

				for _, file := range expectedFiles {
					filePath := filepath.Join(projectDir, file)
					if _, err := os.Stat(filePath); os.IsNotExist(err) {
						t.Errorf("Expected file '%s' was not created", filePath)
					}
				}

				// 清理
				os.RemoveAll(projectDir)
			}
		})
	}
}

func TestCreateFlowCommand(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	log := logger.NewTestLogger()
	createCmd := NewCreateCommand(log)

	// 查找flow子命令
	var flowCmd *cobra.Command
	for _, cmd := range createCmd.Commands() {
		if cmd.Use == "flow [name]" {
			flowCmd = cmd
			break
		}
	}

	if flowCmd == nil {
		t.Fatal("flow subcommand not found")
	}

	// 测试flow创建
	testCmd := &cobra.Command{
		Use:  flowCmd.Use,
		Args: flowCmd.Args,
		RunE: flowCmd.RunE,
	}
	testCmd.Flags().AddFlagSet(flowCmd.Flags())

	var buf bytes.Buffer
	testCmd.SetOut(&buf)
	testCmd.SetErr(&buf)

	testCmd.SetArgs([]string{"test-flow"})
	err := testCmd.ExecuteContext(context.Background())

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// 检查项目目录是否创建
	if _, err := os.Stat("test-flow"); os.IsNotExist(err) {
		t.Error("Flow project directory was not created")
	}

	// 检查配置文件是否存在
	configPath := filepath.Join("test-flow", "greensoulai.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Flow project config file was not created")
	}

	// 清理
	os.RemoveAll("test-flow")
}

func TestCreateAgentCommand(t *testing.T) {
	// 创建临时项目目录
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// 创建模拟项目结构
	projectDir := "test-project"
	if err := os.MkdirAll(filepath.Join(projectDir, "internal", "agents"), 0755); err != nil {
		t.Fatalf("Failed to create project structure: %v", err)
	}

	// 创建项目配置文件
	configContent := `name: test-project
type: crew
go_module: github.com/user/test-project
go_version: "1.21"
agents: []
tasks: []
llm:
  provider: openai
  model: gpt-4o-mini
`
	configPath := filepath.Join(projectDir, "greensoulai.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// 切换到项目目录
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("Failed to change to project directory: %v", err)
	}

	log := logger.NewTestLogger()
	createCmd := NewCreateCommand(log)

	// 查找agent子命令
	var agentCmd *cobra.Command
	for _, cmd := range createCmd.Commands() {
		if cmd.Use == "agent [name]" {
			agentCmd = cmd
			break
		}
	}

	if agentCmd == nil {
		t.Fatal("agent subcommand not found")
	}

	tests := []struct {
		name        string
		args        []string
		flags       map[string]string
		expectError bool
	}{
		{
			name: "create basic agent",
			args: []string{"researcher"},
			flags: map[string]string{
				"role":      "Research Specialist",
				"goal":      "Conduct thorough research",
				"backstory": "Expert researcher with years of experience",
			},
			expectError: false,
		},
		{
			name:        "create agent without flags",
			args:        []string{"analyst"},
			flags:       map[string]string{},
			expectError: false,
		},
		{
			name:        "invalid agent name",
			args:        []string{"123invalid"},
			flags:       map[string]string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCmd := &cobra.Command{
				Use:  agentCmd.Use,
				Args: agentCmd.Args,
				RunE: agentCmd.RunE,
			}
			testCmd.Flags().AddFlagSet(agentCmd.Flags())

			// 设置flags
			for flag, value := range tt.flags {
				if err := testCmd.Flags().Set(flag, value); err != nil {
					t.Fatalf("Failed to set flag %s: %v", flag, err)
				}
			}

			var buf bytes.Buffer
			testCmd.SetOut(&buf)
			testCmd.SetErr(&buf)

			testCmd.SetArgs(tt.args)
			err := testCmd.ExecuteContext(context.Background())

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// 如果成功，检查agent文件是否创建
			if !tt.expectError && err == nil && len(tt.args) > 0 {
				agentName := tt.args[0]
				agentFile := filepath.Join("internal", "agents", strings.ToLower(agentName)+".go")

				if _, err := os.Stat(agentFile); os.IsNotExist(err) {
					t.Errorf("Agent file '%s' was not created", agentFile)
				}
			}
		})
	}
}

func TestCreateTaskCommand(t *testing.T) {
	// 创建临时项目目录
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// 创建模拟项目结构
	projectDir := "test-project"
	if err := os.MkdirAll(filepath.Join(projectDir, "internal", "tasks"), 0755); err != nil {
		t.Fatalf("Failed to create project structure: %v", err)
	}

	// 创建项目配置文件
	configContent := `name: test-project
type: crew
go_module: github.com/user/test-project
go_version: "1.21"
agents:
  - name: researcher
    role: Research Expert
    goal: Conduct research
    backstory: Experienced researcher
tasks: []
llm:
  provider: openai
  model: gpt-4o-mini
`
	configPath := filepath.Join(projectDir, "greensoulai.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// 切换到项目目录
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("Failed to change to project directory: %v", err)
	}

	log := logger.NewTestLogger()
	createCmd := NewCreateCommand(log)

	// 查找task子命令
	var taskCmd *cobra.Command
	for _, cmd := range createCmd.Commands() {
		if cmd.Use == "task [name]" {
			taskCmd = cmd
			break
		}
	}

	if taskCmd == nil {
		t.Fatal("task subcommand not found")
	}

	tests := []struct {
		name        string
		args        []string
		flags       map[string]string
		expectError bool
	}{
		{
			name: "create task with agent",
			args: []string{"research"},
			flags: map[string]string{
				"description":     "Conduct research on topic",
				"expected-output": "Comprehensive research report",
				"agent":           "researcher",
			},
			expectError: false,
		},
		{
			name: "create task with unknown agent",
			args: []string{"analysis"},
			flags: map[string]string{
				"agent": "unknown_agent",
			},
			expectError: true,
		},
		{
			name:        "create task without agent",
			args:        []string{"writing"},
			flags:       map[string]string{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCmd := &cobra.Command{
				Use:  taskCmd.Use,
				Args: taskCmd.Args,
				RunE: taskCmd.RunE,
			}
			testCmd.Flags().AddFlagSet(taskCmd.Flags())

			// 设置flags
			for flag, value := range tt.flags {
				if err := testCmd.Flags().Set(flag, value); err != nil {
					t.Fatalf("Failed to set flag %s: %v", flag, err)
				}
			}

			var buf bytes.Buffer
			testCmd.SetOut(&buf)
			testCmd.SetErr(&buf)

			testCmd.SetArgs(tt.args)
			err := testCmd.ExecuteContext(context.Background())

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// 如果成功，检查task文件是否创建
			if !tt.expectError && err == nil && len(tt.args) > 0 {
				taskName := tt.args[0]
				taskFile := filepath.Join("internal", "tasks", strings.ToLower(taskName)+".go")

				if _, err := os.Stat(taskFile); os.IsNotExist(err) {
					t.Errorf("Task file '%s' was not created", taskFile)
				}
			}
		})
	}
}

func TestCreateToolCommand(t *testing.T) {
	// 创建临时项目目录
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// 创建模拟项目结构
	projectDir := "test-project"
	if err := os.MkdirAll(filepath.Join(projectDir, "internal", "tools"), 0755); err != nil {
		t.Fatalf("Failed to create project structure: %v", err)
	}

	// 创建简单的配置文件
	configContent := `name: test-project
type: crew
go_module: github.com/user/test-project
`
	configPath := filepath.Join(projectDir, "greensoulai.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// 切换到项目目录
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("Failed to change to project directory: %v", err)
	}

	log := logger.NewTestLogger()
	createCmd := NewCreateCommand(log)

	// 查找tool子命令
	var toolCmd *cobra.Command
	for _, cmd := range createCmd.Commands() {
		if cmd.Use == "tool [name]" {
			toolCmd = cmd
			break
		}
	}

	if toolCmd == nil {
		t.Fatal("tool subcommand not found")
	}

	// 测试工具创建
	testCmd := &cobra.Command{
		Use:  toolCmd.Use,
		Args: toolCmd.Args,
		RunE: toolCmd.RunE,
	}
	testCmd.Flags().AddFlagSet(toolCmd.Flags())

	var buf bytes.Buffer
	testCmd.SetOut(&buf)
	testCmd.SetErr(&buf)

	testCmd.SetArgs([]string{"search_tool"})
	err := testCmd.ExecuteContext(context.Background())

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// 检查工具文件是否创建
	toolFile := filepath.Join("internal", "tools", "search_tool.go")
	if _, err := os.Stat(toolFile); os.IsNotExist(err) {
		t.Errorf("Tool file '%s' was not created", toolFile)
	}
}
