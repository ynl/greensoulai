package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ynl/greensoulai/internal/cli/config"
)

// CrewGenerator Crew项目生成器
type CrewGenerator struct {
	config *config.ProjectConfig
	output string
}

// NewCrewGenerator 创建Crew项目生成器
func NewCrewGenerator(cfg *config.ProjectConfig, outputDir string) *CrewGenerator {
	return &CrewGenerator{
		config: cfg,
		output: outputDir,
	}
}

// Generate 生成Crew项目
func (g *CrewGenerator) Generate() error {
	// 创建项目目录
	if err := g.createDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// 生成配置文件
	if err := g.generateConfig(); err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	// 生成Go模块文件
	if err := g.generateGoMod(); err != nil {
		return fmt.Errorf("failed to generate go.mod: %w", err)
	}

	// 生成主文件
	if err := g.generateMain(); err != nil {
		return fmt.Errorf("failed to generate main.go: %w", err)
	}

	// 生成Agent文件
	if err := g.generateAgents(); err != nil {
		return fmt.Errorf("failed to generate agents: %w", err)
	}

	// 生成Task文件
	if err := g.generateTasks(); err != nil {
		return fmt.Errorf("failed to generate tasks: %w", err)
	}

	// 生成Crew文件
	if err := g.generateCrew(); err != nil {
		return fmt.Errorf("failed to generate crew: %w", err)
	}

	// 生成工具文件
	if err := g.generateTools(); err != nil {
		return fmt.Errorf("failed to generate tools: %w", err)
	}

	// 生成README
	if err := g.generateReadme(); err != nil {
		return fmt.Errorf("failed to generate README: %w", err)
	}

	// 生成环境文件
	if err := g.generateEnv(); err != nil {
		return fmt.Errorf("failed to generate .env: %w", err)
	}

	// 生成Makefile
	if err := g.generateMakefile(); err != nil {
		return fmt.Errorf("failed to generate Makefile: %w", err)
	}

	return nil
}

// createDirectories 创建项目目录结构
func (g *CrewGenerator) createDirectories() error {
	dirs := []string{
		g.output,
		filepath.Join(g.output, "cmd"),
		filepath.Join(g.output, "internal"),
		filepath.Join(g.output, "internal", "agents"),
		filepath.Join(g.output, "internal", "tasks"),
		filepath.Join(g.output, "internal", "tools"),
		filepath.Join(g.output, "internal", "crew"),
		filepath.Join(g.output, "config"),
		filepath.Join(g.output, "docs"),
		filepath.Join(g.output, "scripts"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// generateConfig 生成项目配置文件
func (g *CrewGenerator) generateConfig() error {
	configPath := filepath.Join(g.output, "greensoulai.yaml")
	g.config.CreatedAt = time.Now().Format(time.RFC3339)
	return g.config.SaveProjectConfig(configPath)
}

// generateGoMod 生成go.mod文件
func (g *CrewGenerator) generateGoMod() error {
	// 获取当前工作目录作为greensoulai项目根路径
	greensoulaiRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	content := fmt.Sprintf(`module %s

go %s

require (
	github.com/ynl/greensoulai v0.0.0-00010101000000-000000000000
	github.com/spf13/cobra v1.8.0
	github.com/joho/godotenv v1.5.1
	gopkg.in/yaml.v3 v3.0.1
)

// 用于本地开发，指向本地的greensoulai模块
replace github.com/ynl/greensoulai => %s
`, g.config.GoModule, g.config.GoVersion, greensoulaiRoot)

	path := filepath.Join(g.output, "go.mod")
	return os.WriteFile(path, []byte(content), 0644)
}

// generateMain 生成主文件
func (g *CrewGenerator) generateMain() error {
	content := fmt.Sprintf(`package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	
	"%s/internal/crew"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}
	
	// 创建并运行crew
	c, err := crew.New%sCrew()
	if err != nil {
		log.Fatalf("Failed to create crew: %%v", err)
	}
	
	// 运行crew
	if err := c.Run(); err != nil {
		log.Fatalf("Failed to run crew: %%v", err)
	}
}
`, g.config.GoModule, toPascalCase(g.config.Name))

	path := filepath.Join(g.output, "cmd", "main.go")
	return os.WriteFile(path, []byte(content), 0644)
}

// generateAgents 生成Agent文件
func (g *CrewGenerator) generateAgents() error {
	for _, agentCfg := range g.config.Agents {
		content := g.GenerateAgentCode(agentCfg)
		filename := fmt.Sprintf("%s.go", strings.ToLower(agentCfg.Name))
		path := filepath.Join(g.output, "internal", "agents", filename)

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write agent file %s: %w", filename, err)
		}
	}

	return nil
}

// GenerateAgentCode 生成单个Agent的代码
func (g *CrewGenerator) GenerateAgentCode(agentCfg config.AgentConfig) string {
	toolsImports := ""
	toolsSetup := ""

	if len(agentCfg.Tools) > 0 {
		toolsImports = fmt.Sprintf("\n\t\"%s/internal/tools\"", g.config.GoModule)

		var toolSetups []string
		for _, tool := range agentCfg.Tools {
			toolSetups = append(toolSetups, fmt.Sprintf("\t\ttools.New%sTool()", toPascalCase(tool)))
		}
		toolsSetup = fmt.Sprintf(`
	// 添加工具
	tools := []agent.Tool{
%s,
	}
	
	for _, tool := range tools {
		if err := a.AddTool(tool); err != nil {
			return nil, fmt.Errorf("failed to add tool: %%w", err)
		}
	}`, strings.Join(toolSetups, ",\n"))
	}

	return fmt.Sprintf(`package agents

import (
	"fmt"
	
	"github.com/ynl/greensoulai/internal/agent"
	"github.com/ynl/greensoulai/internal/llm"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"%s
)

// New%sAgent 创建%s智能体
func New%sAgent(llmProvider llm.LLM, eventBus events.EventBus, log logger.Logger) (agent.Agent, error) {
	config := agent.AgentConfig{
		Role:      "%s",
		Goal:      "%s", 
		Backstory: "%s",
		LLM:       llmProvider,
		EventBus:  eventBus,
		Logger:    log,
		ExecutionConfig: agent.ExecutionConfig{
			MaxIterations:   25,
			Timeout:        30 * time.Minute,
			VerboseLogging: %t,
		},
	}
	
	a, err := agent.NewBaseAgent(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %%w", err)
	}%s
	
	return a, nil
}
`, toolsImports, toPascalCase(agentCfg.Name), agentCfg.Name, toPascalCase(agentCfg.Name),
		agentCfg.Role, agentCfg.Goal, agentCfg.Backstory, agentCfg.Verbose, toolsSetup)
}

// generateTasks 生成Task文件
func (g *CrewGenerator) generateTasks() error {
	for _, taskCfg := range g.config.Tasks {
		content := g.GenerateTaskCode(taskCfg)
		filename := fmt.Sprintf("%s.go", strings.ToLower(taskCfg.Name))
		path := filepath.Join(g.output, "internal", "tasks", filename)

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write task file %s: %w", filename, err)
		}
	}

	return nil
}

// GenerateTaskCode 生成单个Task的代码
func (g *CrewGenerator) GenerateTaskCode(taskCfg config.TaskConfig) string {
	return fmt.Sprintf(`package tasks

import (
	"github.com/ynl/greensoulai/internal/agent"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
)

// New%sTask 创建%s任务
func New%sTask(eventBus events.EventBus, log logger.Logger) (agent.Task, error) {
	task := &agent.BaseTask{
		Name:           "%s",
		Description:    "%s",
		ExpectedOutput: "%s",
		OutputFile:     "%s",
		EventBus:       eventBus,
		Logger:         log,
	}
	
	return task, nil
}
`, toPascalCase(taskCfg.Name), taskCfg.Name, toPascalCase(taskCfg.Name),
		taskCfg.Name, taskCfg.Description, taskCfg.ExpectedOutput, taskCfg.OutputFile)
}

// generateCrew 生成Crew文件
func (g *CrewGenerator) generateCrew() error {
	content := g.generateCrewCode()
	path := filepath.Join(g.output, "internal", "crew", "crew.go")
	return os.WriteFile(path, []byte(content), 0644)
}

// generateCrewCode 生成Crew代码
func (g *CrewGenerator) generateCrewCode() string {
	agentImports := make([]string, len(g.config.Agents))
	taskImports := make([]string, len(g.config.Tasks))

	agentCreations := make([]string, len(g.config.Agents))
	taskCreations := make([]string, len(g.config.Tasks))

	for i, agentCfg := range g.config.Agents {
		agentImports[i] = fmt.Sprintf("\"%s/internal/agents\"", g.config.GoModule)
		agentCreations[i] = fmt.Sprintf(`	%sAgent, err := agents.New%sAgent(llmProvider, eventBus, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s agent: %%w", err)
	}`, agentCfg.Name, toPascalCase(agentCfg.Name), agentCfg.Name)
	}

	for i, taskCfg := range g.config.Tasks {
		taskImports[i] = fmt.Sprintf("\"%s/internal/tasks\"", g.config.GoModule)
		taskCreations[i] = fmt.Sprintf(`	%sTask, err := tasks.New%sTask(eventBus, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s task: %%w", err)
	}`, taskCfg.Name, toPascalCase(taskCfg.Name), taskCfg.Name)
	}

	// 任务分配逻辑
	taskAssignments := make([]string, len(g.config.Tasks))
	agentsList := make([]string, len(g.config.Agents))
	tasksList := make([]string, len(g.config.Tasks))

	for i, taskCfg := range g.config.Tasks {
		if taskCfg.Agent != "" {
			taskAssignments[i] = fmt.Sprintf(`	// 分配任务给Agent
	if err := %sTask.SetAssignedAgent(%sAgent); err != nil {
		return nil, fmt.Errorf("failed to assign task to agent: %%w", err)
	}`, taskCfg.Name, taskCfg.Agent)
		}
		tasksList[i] = fmt.Sprintf("%sTask", taskCfg.Name)
	}

	for i, agentCfg := range g.config.Agents {
		agentsList[i] = fmt.Sprintf("%sAgent", agentCfg.Name)
	}

	return fmt.Sprintf(`package crew

import (
	"context"
	"fmt"
	"os"
	
	"github.com/ynl/greensoulai/internal/crew"
	"github.com/ynl/greensoulai/internal/llm"
	"github.com/ynl/greensoulai/pkg/events"
	"github.com/ynl/greensoulai/pkg/logger"
	%s
	%s
)

// %sCrew 结构
type %sCrew struct {
	crew crew.Crew
	log  logger.Logger
}

// New%sCrew 创建%s团队
func New%sCrew() (*%sCrew, error) {
	// 创建日志器
	log := logger.NewConsoleLogger()
	
	// 创建事件总线
	eventBus := events.NewEventBus(log)
	
	// 创建LLM提供商
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}
	
	llmProvider := llm.NewOpenAILLM("%s", llm.WithAPIKey(apiKey))
	
	// 创建Agents
%s
	
	// 创建Tasks
%s
	
%s
	
	// 创建Crew配置
	config := &crew.CrewConfig{
		Name:    "%s",
		Process: crew.ProcessSequential,
		Verbose: true,
	}
	
	// 创建Crew
	c, err := crew.NewBaseCrew(config, eventBus, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create crew: %%w", err)
	}
	
	// 添加Agents
	agents := []agent.Agent{%s}
	for _, a := range agents {
		if err := c.AddAgent(a); err != nil {
			return nil, fmt.Errorf("failed to add agent: %%w", err)
		}
	}
	
	// 添加Tasks
	tasks := []agent.Task{%s}
	for _, t := range tasks {
		if err := c.AddTask(t); err != nil {
			return nil, fmt.Errorf("failed to add task: %%w", err)
		}
	}
	
	return &%sCrew{
		crew: c,
		log:  log,
	}, nil
}

// Run 运行Crew
func (c *%sCrew) Run() error {
	c.log.Info("启动%s团队...")
	
	ctx := context.Background()
	inputs := make(map[string]interface{})
	
	output, err := c.crew.Kickoff(ctx, inputs)
	if err != nil {
		return fmt.Errorf("crew execution failed: %%w", err)
	}
	
	c.log.Info("团队执行完成")
	c.log.Info("执行结果", logger.Field{Key: "output", Value: output.Raw})
	
	return nil
}
`, strings.Join(removeDuplicates(agentImports), "\n\t"), strings.Join(removeDuplicates(taskImports), "\n\t"),
		toPascalCase(g.config.Name), toPascalCase(g.config.Name),
		toPascalCase(g.config.Name), g.config.Name, toPascalCase(g.config.Name), toPascalCase(g.config.Name),
		g.config.LLM.Model,
		strings.Join(agentCreations, "\n\n"), strings.Join(taskCreations, "\n\n"),
		strings.Join(taskAssignments, "\n"),
		g.config.Name,
		strings.Join(agentsList, ", "), strings.Join(tasksList, ", "),
		toPascalCase(g.config.Name), toPascalCase(g.config.Name), g.config.Name)
}

// generateTools 生成工具文件
func (g *CrewGenerator) generateTools() error {
	// 收集所有使用的工具
	toolSet := make(map[string]bool)
	for _, agent := range g.config.Agents {
		for _, tool := range agent.Tools {
			toolSet[tool] = true
		}
	}

	// 为每个工具生成代码
	for toolName := range toolSet {
		content := g.GenerateToolCode(toolName)
		filename := fmt.Sprintf("%s.go", strings.ToLower(toolName))
		path := filepath.Join(g.output, "internal", "tools", filename)

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write tool file %s: %w", filename, err)
		}
	}

	return nil
}

// GenerateToolCode 生成工具代码
func (g *CrewGenerator) GenerateToolCode(toolName string) string {
	return fmt.Sprintf(`package tools

import (
	"context"
	"fmt"
	
	"github.com/ynl/greensoulai/internal/agent"
)

// New%sTool 创建%s工具
func New%sTool() agent.Tool {
	return agent.NewBaseTool(
		"%s",
		"%s工具的描述",
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			// TODO: 实现%s工具的具体逻辑
			
			// 示例实现
			input, ok := args["input"]
			if !ok {
				return nil, fmt.Errorf("missing input parameter")
			}
			
			result := fmt.Sprintf("%s工具处理结果: %%v", input)
			return result, nil
		},
	)
}
`, toPascalCase(toolName), toolName, toPascalCase(toolName), toolName, toolName, toolName, toolName)
}

// generateReadme 生成README文件
func (g *CrewGenerator) generateReadme() error {
	content := fmt.Sprintf(`# %s

%s

## 🚀 快速开始

### 1. 环境准备

确保你已安装Go %s或更高版本。

### 2. 安装依赖

`+"```"+`bash
go mod download
`+"```"+`

### 3. 配置环境变量

复制 `+"`"+`.env.example`+"`"+` 到 `+"`"+`.env`+"`"+` 并设置你的API密钥：

`+"```"+`bash
cp .env.example .env
# 编辑 .env 文件，设置 OPENAI_API_KEY
`+"```"+`

### 4. 运行项目

`+"```"+`bash
# 使用 greensoulai CLI
greensoulai run

# 或直接运行
go run cmd/main.go

# 或使用 Makefile
make run
`+"```"+`

## 📁 项目结构

`+"```"+`
%s/
├── cmd/
│   └── main.go              # 程序入口
├── internal/
│   ├── agents/              # 智能体定义
│   ├── tasks/               # 任务定义
│   ├── tools/               # 工具实现
│   └── crew/                # 团队配置
├── config/                  # 配置文件
├── docs/                    # 文档
├── scripts/                 # 脚本文件
├── greensoulai.yaml         # 项目配置
├── go.mod                   # Go模块文件
├── .env                     # 环境变量
├── Makefile                 # 构建脚本
└── README.md               # 说明文档
`+"```"+`

## 🤖 智能体配置

本项目包含以下智能体：

%s

## 📋 任务配置

定义的任务：

%s

## 🛠️ 工具集

可用工具：

%s

## ⚙️ 配置说明

主要配置文件：

- `+"`"+`greensoulai.yaml`+"`"+`: 项目主配置
- `+"`"+`.env`+"`"+`: 环境变量配置

### LLM 配置

当前使用的LLM配置：
- 提供商: %s
- 模型: %s
- 温度: %.2f

## 🔧 自定义开发

### 添加新的智能体

1. 在 `+"`"+`internal/agents/`+"`"+` 中创建新的智能体文件
2. 实现智能体逻辑
3. 在 `+"`"+`greensoulai.yaml`+"`"+` 中配置智能体

### 添加新的任务

1. 在 `+"`"+`internal/tasks/`+"`"+` 中创建新的任务文件
2. 实现任务逻辑
3. 在 `+"`"+`greensoulai.yaml`+"`"+` 中配置任务

### 添加新的工具

1. 在 `+"`"+`internal/tools/`+"`"+` 中创建新的工具文件
2. 实现工具逻辑
3. 在智能体配置中引用工具

## 📚 文档

- [API文档](docs/api.md)
- [开发指南](docs/development.md)
- [部署指南](docs/deployment.md)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

本项目采用 MIT 许可证。
`,
		g.config.Name, g.config.Description, g.config.GoVersion, g.config.Name,
		g.generateAgentsList(), g.generateTasksList(), g.generateToolsList(),
		g.config.LLM.Provider, g.config.LLM.Model, g.config.LLM.Temperature)

	path := filepath.Join(g.output, "README.md")
	return os.WriteFile(path, []byte(content), 0644)
}

// generateAgentsList 生成智能体列表
func (g *CrewGenerator) generateAgentsList() string {
	if len(g.config.Agents) == 0 {
		return "无智能体配置"
	}

	var agents []string
	for _, agent := range g.config.Agents {
		toolList := "无"
		if len(agent.Tools) > 0 {
			toolList = strings.Join(agent.Tools, ", ")
		}

		agents = append(agents, fmt.Sprintf(`- **%s** (%s)
  - 目标: %s
  - 背景: %s
  - 工具: %s`, agent.Name, agent.Role, agent.Goal, agent.Backstory, toolList))
	}

	return strings.Join(agents, "\n\n")
}

// generateTasksList 生成任务列表
func (g *CrewGenerator) generateTasksList() string {
	if len(g.config.Tasks) == 0 {
		return "无任务配置"
	}

	var tasks []string
	for _, task := range g.config.Tasks {
		assignedAgent := "未分配"
		if task.Agent != "" {
			assignedAgent = task.Agent
		}

		tasks = append(tasks, fmt.Sprintf(`- **%s**
  - 描述: %s
  - 期望输出: %s
  - 分配智能体: %s`, task.Name, task.Description, task.ExpectedOutput, assignedAgent))
	}

	return strings.Join(tasks, "\n\n")
}

// generateToolsList 生成工具列表
func (g *CrewGenerator) generateToolsList() string {
	toolSet := make(map[string]bool)
	for _, agent := range g.config.Agents {
		for _, tool := range agent.Tools {
			toolSet[tool] = true
		}
	}

	if len(toolSet) == 0 {
		return "无工具配置"
	}

	var tools []string
	for tool := range toolSet {
		tools = append(tools, fmt.Sprintf("- %s", tool))
	}

	return strings.Join(tools, "\n")
}

// generateEnv 生成环境变量文件
func (g *CrewGenerator) generateEnv() error {
	content := `# OpenAI API配置
OPENAI_API_KEY=your_openai_api_key_here

# 可选：OpenAI Base URL (如果使用代理或其他兼容服务)
# OPENAI_BASE_URL=https://api.openai.com/v1

# 日志级别 (debug, info, warn, error)
LOG_LEVEL=info

# 其他配置
# CREW_VERBOSE=true
`

	path := filepath.Join(g.output, ".env.example")
	return os.WriteFile(path, []byte(content), 0644)
}

// generateMakefile 生成Makefile
func (g *CrewGenerator) generateMakefile() error {
	content := fmt.Sprintf(`.PHONY: build run test clean deps

# Go参数
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# 项目参数
BINARY_NAME=%s
MAIN_PATH=./cmd/main.go

# 构建
build:
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)

# 运行
run:
	$(GOCMD) run $(MAIN_PATH)

# 测试
test:
	$(GOTEST) -v ./...

# 测试覆盖率
test-coverage:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

# 清理
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out

# 依赖管理
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# 更新依赖
update:
	$(GOMOD) download
	$(GOMOD) tidy
	$(GOGET) -u ./...

# 格式化代码
fmt:
	gofmt -s -w .
	$(GOCMD) mod tidy

# 静态检查
lint:
	golangci-lint run

# 开发环境设置
setup:
	cp .env.example .env
	$(MAKE) deps

# 帮助
help:
	@echo "可用命令:"
	@echo "  build        构建项目"
	@echo "  run          运行项目" 
	@echo "  test         运行测试"
	@echo "  test-coverage 运行测试并生成覆盖率报告"
	@echo "  clean        清理构建文件"
	@echo "  deps         下载依赖"
	@echo "  update       更新依赖"
	@echo "  fmt          格式化代码"
	@echo "  lint         静态检查"
	@echo "  setup        设置开发环境"
	@echo "  help         显示帮助信息"
`, g.config.Name)

	path := filepath.Join(g.output, "Makefile")
	return os.WriteFile(path, []byte(content), 0644)
}

// 工具函数

// toPascalCase 转换为帕斯卡命名
func toPascalCase(s string) string {
	if s == "" {
		return s
	}

	// 处理下划线分隔的字符串
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})

	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(part[:1]))
			if len(part) > 1 {
				result.WriteString(strings.ToLower(part[1:]))
			}
		}
	}

	return result.String()
}

// removeDuplicates 移除重复字符串
func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}
