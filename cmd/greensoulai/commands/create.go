package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/ynl/greensoulai/internal/cli/config"
	"github.com/ynl/greensoulai/internal/cli/generator"
	"github.com/ynl/greensoulai/internal/cli/utils"
	"github.com/ynl/greensoulai/pkg/logger"
)

// NewCreateCommand 创建create命令
func NewCreateCommand(log logger.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "创建新的项目或组件",
		Long:  "创建新的GreenSoulAI项目、智能体、任务或团队",
	}

	// 添加子命令
	cmd.AddCommand(
		newCreateCrewCommand(log),
		newCreateFlowCommand(log),
		newCreateAgentCommand(log),
		newCreateTaskCommand(log),
		newCreateToolCommand(log),
	)

	return cmd
}

// newCreateCrewCommand 创建crew项目命令
func newCreateCrewCommand(log logger.Logger) *cobra.Command {
	var (
		outputDir   string
		goModule    string
		provider    string
		skipPrompt  bool
		interactive bool
	)

	cmd := &cobra.Command{
		Use:   "crew [name]",
		Short: "创建新的Crew项目",
		Long: `创建一个新的GreenSoulAI Crew项目，包含完整的项目结构、
示例智能体、任务和配置文件。`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]

			// 验证项目名称
			if err := utils.ValidateProjectName(projectName); err != nil {
				return fmt.Errorf("invalid project name: %w", err)
			}

			// 设置输出目录
			if outputDir == "" {
				outputDir = utils.NormalizeName(projectName)
			}

			// 检查目录是否存在
			absOutputDir, err := utils.FormatPath(outputDir)
			if err != nil {
				return fmt.Errorf("failed to format output directory: %w", err)
			}

			exists, err := utils.CheckDirectoryExists(absOutputDir)
			if err != nil {
				return fmt.Errorf("failed to check output directory: %w", err)
			}

			if exists {
				isEmpty, err := utils.IsDirectoryEmpty(absOutputDir)
				if err != nil {
					return fmt.Errorf("failed to check if directory is empty: %w", err)
				}
				if !isEmpty && !skipPrompt {
					return fmt.Errorf("directory %s already exists and is not empty", absOutputDir)
				}
			}

			// 设置Go模块名
			if goModule == "" {
				goModule = utils.GenerateGoModule(projectName)
				if interactive {
					fmt.Printf("建议的Go模块名: %s\n", goModule)
					fmt.Print("请输入Go模块名 (按回车使用建议值): ")
					var input string
					if _, err := fmt.Scanln(&input); err != nil {
						// 如果用户直接按回车或输入无效，使用默认值
						input = ""
					}
					if input != "" {
						goModule = input
					}
				}
			}

			// 验证Go模块名
			if err := utils.ValidateGoModule(goModule); err != nil {
				log.Warn("Go module name validation warning",
					logger.Field{Key: "warning", Value: err.Error()})
			}

			// 创建项目配置
			projectConfig := config.DefaultCrewProjectConfig(projectName, goModule)

			// 如果是交互模式，允许用户自定义配置
			if interactive {
				if err := configureProjectInteractively(projectConfig); err != nil {
					return fmt.Errorf("failed to configure project: %w", err)
				}
			}

			// 验证配置
			if err := projectConfig.Validate(); err != nil {
				return fmt.Errorf("invalid project configuration: %w", err)
			}

			log.Info("创建Crew项目",
				logger.Field{Key: "name", Value: projectName},
				logger.Field{Key: "output", Value: absOutputDir},
				logger.Field{Key: "module", Value: goModule},
			)

			// 生成项目
			gen := generator.NewCrewGenerator(projectConfig, absOutputDir)
			if err := gen.Generate(); err != nil {
				return fmt.Errorf("failed to generate project: %w", err)
			}

			// 成功消息
			log.Info("Crew项目创建成功!")
			fmt.Printf(`
✅ Crew项目 '%s' 创建成功！

📁 项目目录: %s
🔗 Go模块: %s

🚀 下一步：
1. 进入项目目录: cd %s
2. 设置环境变量: cp .env.example .env
3. 编辑 .env 文件，设置你的 OPENAI_API_KEY
4. 安装依赖: go mod download
5. 运行项目: greensoulai run

📚 文档：
- 项目配置: greensoulai.yaml
- README: README.md
- API文档: docs/api.md

Happy coding! 🎉
`, projectName, absOutputDir, goModule, outputDir)

			return nil
		},
	}

	// 添加选项
	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "输出目录 (默认为项目名)")
	cmd.Flags().StringVarP(&goModule, "module", "m", "", "Go模块名 (例如: github.com/user/project)")
	cmd.Flags().StringVarP(&provider, "provider", "p", "openai", "LLM提供商 (openai, anthropic)")
	cmd.Flags().BoolVar(&skipPrompt, "skip-prompt", false, "跳过确认提示")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "交互式配置")

	return cmd
}

// newCreateFlowCommand 创建flow项目命令
func newCreateFlowCommand(log logger.Logger) *cobra.Command {
	var (
		outputDir  string
		goModule   string
		skipPrompt bool
	)

	cmd := &cobra.Command{
		Use:   "flow [name]",
		Short: "创建新的Flow项目",
		Long: `创建一个新的GreenSoulAI Flow项目，用于复杂的工作流编排。
Flow项目专注于多阶段的工作流程，支持条件分支和并行执行。`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]

			// 验证项目名称
			if err := utils.ValidateProjectName(projectName); err != nil {
				return fmt.Errorf("invalid project name: %w", err)
			}

			// 设置输出目录
			if outputDir == "" {
				outputDir = utils.NormalizeName(projectName)
			}

			// 设置Go模块名
			if goModule == "" {
				goModule = utils.GenerateGoModule(projectName)
			}

			// 验证Go模块名
			if err := utils.ValidateGoModule(goModule); err != nil {
				log.Warn("Go module name validation warning",
					logger.Field{Key: "warning", Value: err.Error()})
			}

			log.Info("创建Flow项目",
				logger.Field{Key: "name", Value: projectName},
				logger.Field{Key: "output", Value: outputDir},
				logger.Field{Key: "module", Value: goModule},
			)

			// TODO: 实现Flow项目生成逻辑
			// 现在先创建基本的项目结构
			absOutputDir, err := utils.FormatPath(outputDir)
			if err != nil {
				return fmt.Errorf("failed to format output directory: %w", err)
			}

			if err := utils.EnsureDirectoryExists(absOutputDir); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			// 创建基本的Flow项目配置
			projectConfig := config.DefaultFlowProjectConfig(projectName, goModule)
			configPath := filepath.Join(absOutputDir, "greensoulai.yaml")
			if err := projectConfig.SaveProjectConfig(configPath); err != nil {
				return fmt.Errorf("failed to save project config: %w", err)
			}

			log.Info("Flow项目创建成功!", logger.Field{Key: "path", Value: absOutputDir})

			fmt.Printf(`
✅ Flow项目 '%s' 创建成功！

📁 项目目录: %s
🔗 Go模块: %s

⚠️  注意: Flow项目功能正在开发中，当前版本提供基础项目结构。

🚀 下一步：
1. 进入项目目录: cd %s
2. 查看配置文件: greensoulai.yaml
3. 关注项目更新获取完整Flow功能

`, projectName, absOutputDir, goModule, outputDir)

			return nil
		},
	}

	// 添加选项
	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "输出目录 (默认为项目名)")
	cmd.Flags().StringVarP(&goModule, "module", "m", "", "Go模块名 (例如: github.com/user/project)")
	cmd.Flags().BoolVar(&skipPrompt, "skip-prompt", false, "跳过确认提示")

	return cmd
}

// newCreateAgentCommand 创建智能体命令
func newCreateAgentCommand(log logger.Logger) *cobra.Command {
	var (
		role      string
		goal      string
		backstory string
		tools     []string
	)

	cmd := &cobra.Command{
		Use:   "agent [name]",
		Short: "在当前项目中创建新的智能体",
		Long: `在当前GreenSoulAI项目中创建一个新的智能体。
需要在项目根目录中运行此命令。`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentName := args[0]

			// 验证智能体名称
			if err := utils.ValidateProjectName(agentName); err != nil {
				return fmt.Errorf("invalid agent name: %w", err)
			}

			// 查找项目根目录
			projectRoot, err := config.GetProjectRoot()
			if err != nil {
				return fmt.Errorf("not in a greensoulai project: %w", err)
			}

			// 加载项目配置
			configPath := filepath.Join(projectRoot, "greensoulai.yaml")
			projectConfig, err := config.LoadProjectConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load project config: %w", err)
			}

			// 检查智能体是否已存在
			for _, agent := range projectConfig.Agents {
				if agent.Name == agentName {
					return fmt.Errorf("agent '%s' already exists", agentName)
				}
			}

			// 如果没有提供参数，使用默认值或提示用户输入
			if role == "" {
				role = fmt.Sprintf("%s专家", utils.ToPascalCase(agentName))
			}
			if goal == "" {
				goal = fmt.Sprintf("协助完成与%s相关的任务", agentName)
			}
			if backstory == "" {
				backstory = fmt.Sprintf("你是一位经验丰富的%s，擅长处理相关领域的复杂问题。", role)
			}

			// 创建智能体配置
			newAgent := config.AgentConfig{
				Name:      agentName,
				Role:      role,
				Goal:      goal,
				Backstory: backstory,
				Tools:     tools,
				Verbose:   true,
			}

			// 添加到项目配置
			projectConfig.Agents = append(projectConfig.Agents, newAgent)

			// 保存配置
			if err := projectConfig.SaveProjectConfig(configPath); err != nil {
				return fmt.Errorf("failed to save project config: %w", err)
			}

			// 生成智能体代码文件
			gen := generator.NewCrewGenerator(projectConfig, projectRoot)
			agentCode := gen.GenerateAgentCode(newAgent)

			agentFileName := fmt.Sprintf("%s.go", utils.ToSnakeCase(agentName))
			agentFilePath := filepath.Join(projectRoot, "internal", "agents", agentFileName)

			if err := os.WriteFile(agentFilePath, []byte(agentCode), 0644); err != nil {
				return fmt.Errorf("failed to write agent file: %w", err)
			}

			log.Info("智能体创建成功!",
				logger.Field{Key: "name", Value: agentName},
				logger.Field{Key: "role", Value: role},
				logger.Field{Key: "file", Value: agentFilePath},
			)

			fmt.Printf(`
✅ 智能体 '%s' 创建成功！

👤 角色: %s
🎯 目标: %s
📝 背景: %s
🛠️  工具: %v

📁 文件位置: %s
⚙️  配置已更新: greensoulai.yaml

🚀 下一步：
1. 编辑智能体文件自定义逻辑
2. 在任务中引用该智能体
3. 运行项目测试智能体功能

`, agentName, role, goal, backstory, tools, agentFilePath)

			return nil
		},
	}

	// 添加选项
	cmd.Flags().StringVarP(&role, "role", "r", "", "智能体角色")
	cmd.Flags().StringVarP(&goal, "goal", "g", "", "智能体目标")
	cmd.Flags().StringVarP(&backstory, "backstory", "b", "", "智能体背景故事")
	cmd.Flags().StringSliceVarP(&tools, "tools", "t", nil, "智能体工具列表")

	return cmd
}

// newCreateTaskCommand 创建任务命令
func newCreateTaskCommand(log logger.Logger) *cobra.Command {
	var (
		description    string
		expectedOutput string
		agent          string
		outputFormat   string
		outputFile     string
	)

	cmd := &cobra.Command{
		Use:   "task [name]",
		Short: "在当前项目中创建新的任务",
		Long: `在当前GreenSoulAI项目中创建一个新的任务。
需要在项目根目录中运行此命令。`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskName := args[0]

			// 验证任务名称
			if err := utils.ValidateProjectName(taskName); err != nil {
				return fmt.Errorf("invalid task name: %w", err)
			}

			// 查找项目根目录
			projectRoot, err := config.GetProjectRoot()
			if err != nil {
				return fmt.Errorf("not in a greensoulai project: %w", err)
			}

			// 加载项目配置
			configPath := filepath.Join(projectRoot, "greensoulai.yaml")
			projectConfig, err := config.LoadProjectConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load project config: %w", err)
			}

			// 检查任务是否已存在
			for _, task := range projectConfig.Tasks {
				if task.Name == taskName {
					return fmt.Errorf("task '%s' already exists", taskName)
				}
			}

			// 验证智能体是否存在
			if agent != "" {
				agentExists := false
				for _, a := range projectConfig.Agents {
					if a.Name == agent {
						agentExists = true
						break
					}
				}
				if !agentExists {
					return fmt.Errorf("agent '%s' does not exist", agent)
				}
			}

			// 设置默认值
			if description == "" {
				description = fmt.Sprintf("执行%s相关的任务", taskName)
			}
			if expectedOutput == "" {
				expectedOutput = "详细的任务执行结果"
			}
			if outputFormat == "" {
				outputFormat = "markdown"
			}
			if outputFile == "" {
				outputFile = fmt.Sprintf("%s_output.md", utils.ToSnakeCase(taskName))
			}

			// 创建任务配置
			newTask := config.TaskConfig{
				Name:           taskName,
				Description:    description,
				ExpectedOutput: expectedOutput,
				Agent:          agent,
				OutputFormat:   outputFormat,
				OutputFile:     outputFile,
			}

			// 添加到项目配置
			projectConfig.Tasks = append(projectConfig.Tasks, newTask)

			// 保存配置
			if err := projectConfig.SaveProjectConfig(configPath); err != nil {
				return fmt.Errorf("failed to save project config: %w", err)
			}

			// 生成任务代码文件
			gen := generator.NewCrewGenerator(projectConfig, projectRoot)
			taskCode := gen.GenerateTaskCode(newTask)

			taskFileName := fmt.Sprintf("%s.go", utils.ToSnakeCase(taskName))
			taskFilePath := filepath.Join(projectRoot, "internal", "tasks", taskFileName)

			if err := os.WriteFile(taskFilePath, []byte(taskCode), 0644); err != nil {
				return fmt.Errorf("failed to write task file: %w", err)
			}

			log.Info("任务创建成功!",
				logger.Field{Key: "name", Value: taskName},
				logger.Field{Key: "description", Value: description},
				logger.Field{Key: "agent", Value: agent},
				logger.Field{Key: "file", Value: taskFilePath},
			)

			fmt.Printf(`
✅ 任务 '%s' 创建成功！

📝 描述: %s
🎯 期望输出: %s
👤 分配智能体: %s
📊 输出格式: %s
📁 输出文件: %s

📁 文件位置: %s
⚙️  配置已更新: greensoulai.yaml

🚀 下一步：
1. 编辑任务文件自定义逻辑
2. 运行项目测试任务功能
3. 查看输出结果

`, taskName, description, expectedOutput, agent, outputFormat, outputFile, taskFilePath)

			return nil
		},
	}

	// 添加选项
	cmd.Flags().StringVarP(&description, "description", "d", "", "任务描述")
	cmd.Flags().StringVarP(&expectedOutput, "expected-output", "e", "", "期望输出")
	cmd.Flags().StringVarP(&agent, "agent", "a", "", "分配的智能体")
	cmd.Flags().StringVar(&outputFormat, "format", "markdown", "输出格式 (markdown, json, raw)")
	cmd.Flags().StringVarP(&outputFile, "output-file", "f", "", "输出文件名")

	return cmd
}

// newCreateToolCommand 创建工具命令
func newCreateToolCommand(log logger.Logger) *cobra.Command {
	var (
		description string
		packageName string
	)

	cmd := &cobra.Command{
		Use:   "tool [name]",
		Short: "在当前项目中创建新的工具",
		Long: `在当前GreenSoulAI项目中创建一个新的工具。
需要在项目根目录中运行此命令。`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			toolName := args[0]

			// 验证工具名称
			if err := utils.ValidateProjectName(toolName); err != nil {
				return fmt.Errorf("invalid tool name: %w", err)
			}

			// 查找项目根目录
			projectRoot, err := config.GetProjectRoot()
			if err != nil {
				return fmt.Errorf("not in a greensoulai project: %w", err)
			}

			// 设置默认值
			if description == "" {
				description = fmt.Sprintf("%s工具的描述", toolName)
			}

			// 生成工具代码
			gen := generator.NewCrewGenerator(nil, projectRoot)
			toolCode := gen.GenerateToolCode(toolName)

			toolFileName := fmt.Sprintf("%s.go", utils.ToSnakeCase(toolName))
			toolFilePath := filepath.Join(projectRoot, "internal", "tools", toolFileName)

			// 确保工具目录存在
			if err := utils.EnsureDirectoryExists(filepath.Dir(toolFilePath)); err != nil {
				return fmt.Errorf("failed to create tools directory: %w", err)
			}

			if err := os.WriteFile(toolFilePath, []byte(toolCode), 0644); err != nil {
				return fmt.Errorf("failed to write tool file: %w", err)
			}

			log.Info("工具创建成功!",
				logger.Field{Key: "name", Value: toolName},
				logger.Field{Key: "description", Value: description},
				logger.Field{Key: "file", Value: toolFilePath},
			)

			fmt.Printf(`
✅ 工具 '%s' 创建成功！

📝 描述: %s
📁 文件位置: %s

🚀 下一步：
1. 编辑工具文件实现具体功能
2. 在智能体配置中引用该工具
3. 运行项目测试工具功能

💡 提示：
- 工具函数签名: func(ctx context.Context, args map[string]interface{}) (interface{}, error)
- 可以在args中获取输入参数
- 返回值会传递给智能体

`, toolName, description, toolFilePath)

			return nil
		},
	}

	// 添加选项
	cmd.Flags().StringVarP(&description, "description", "d", "", "工具描述")
	cmd.Flags().StringVar(&packageName, "package", "tools", "工具包名")

	return cmd
}

// configureProjectInteractively 交互式配置项目
func configureProjectInteractively(config *config.ProjectConfig) error {
	// 这里可以添加交互式配置逻辑
	// 例如：询问用户是否要修改默认配置
	fmt.Println("交互式配置模式 (未来实现)")
	return nil
}
