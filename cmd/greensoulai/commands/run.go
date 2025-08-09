package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/ynl/greensoulai/internal/cli/config"
	"github.com/ynl/greensoulai/internal/cli/utils"
	"github.com/ynl/greensoulai/pkg/logger"
)

// NewRunCommand 创建run命令
func NewRunCommand(log logger.Logger) *cobra.Command {
	var (
		configPath  string
		verbose     bool
		inputFile   string
		outputDir   string
		timeout     time.Duration
		iterations  int
		development bool
	)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "运行GreenSoulAI项目",
		Long: `运行当前目录的GreenSoulAI项目。
会自动检测项目类型（Crew或Flow）并执行相应的运行逻辑。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 查找项目根目录
			projectRoot, err := config.GetProjectRoot()
			if err != nil {
				return fmt.Errorf("not in a greensoulai project: %w", err)
			}

			// 设置配置文件路径
			if configPath == "" {
				configPath = filepath.Join(projectRoot, "greensoulai.yaml")
			}

			// 加载项目配置
			projectConfig, err := config.LoadProjectConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load project config: %w", err)
			}

			// 验证配置
			if err := projectConfig.Validate(); err != nil {
				return fmt.Errorf("invalid project configuration: %w", err)
			}

			log.Info("运行GreenSoulAI项目",
				logger.Field{Key: "name", Value: projectConfig.Name},
				logger.Field{Key: "type", Value: string(projectConfig.Type)},
				logger.Field{Key: "root", Value: projectRoot},
			)

			// 根据项目类型执行不同的运行逻辑
			switch projectConfig.Type {
			case config.ProjectTypeCrew:
				return runCrewProject(cmd.Context(), projectConfig, projectRoot,
					verbose, inputFile, outputDir, timeout, development, log)
			case config.ProjectTypeFlow:
				return runFlowProject(cmd.Context(), projectConfig, projectRoot,
					verbose, inputFile, outputDir, timeout, development, log)
			default:
				return fmt.Errorf("unsupported project type: %s", projectConfig.Type)
			}
		},
	}

	// 添加选项
	cmd.Flags().StringVarP(&configPath, "config", "c", "", "配置文件路径")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "详细输出模式")
	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "输入文件路径")
	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "输出目录")
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", 30*time.Minute, "执行超时时间")
	cmd.Flags().IntVarP(&iterations, "iterations", "n", 1, "执行迭代次数")
	cmd.Flags().BoolVarP(&development, "dev", "d", false, "开发模式（启用热重载）")

	return cmd
}

// runCrewProject 运行Crew项目
func runCrewProject(ctx context.Context, config *config.ProjectConfig,
	projectRoot string, verbose bool, inputFile, outputDir string,
	timeout time.Duration, development bool, log logger.Logger) error {

	log.Info("运行Crew项目", logger.Field{Key: "name", Value: config.Name})

	// 检查必要的环境变量
	if err := checkEnvironmentVariables(config, log); err != nil {
		return fmt.Errorf("environment check failed: %w", err)
	}

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 切换到项目目录
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(projectRoot); err != nil {
		return fmt.Errorf("failed to change to project directory: %w", err)
	}

	// 构建运行命令
	var cmd *exec.Cmd

	// 检查是否有预编译的二进制文件
	mainPath := filepath.Join(projectRoot, "cmd", "main.go")
	if _, err := os.Stat(mainPath); err == nil {
		// 使用go run
		cmd = exec.CommandContext(ctx, "go", "run", "cmd/main.go")
	} else {
		// 查找其他可能的入口点
		possiblePaths := []string{
			"main.go",
			"cmd/main.go",
			fmt.Sprintf("cmd/%s/main.go", config.Name),
		}

		var foundPath string
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				foundPath = path
				break
			}
		}

		if foundPath == "" {
			return fmt.Errorf("no main.go file found, please ensure project structure is correct")
		}

		cmd = exec.CommandContext(ctx, "go", "run", foundPath)
	}

	// 设置环境变量
	cmd.Env = os.Environ()
	if verbose {
		cmd.Env = append(cmd.Env, "LOG_LEVEL=debug")
		cmd.Env = append(cmd.Env, "CREW_VERBOSE=true")
	}

	// 设置输入输出
	if inputFile != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("INPUT_FILE=%s", inputFile))
	}
	if outputDir != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("OUTPUT_DIR=%s", outputDir))
	}

	// 设置标准输入输出
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	log.Info("启动Crew执行...")

	// 记录开始时间
	startTime := time.Now()

	// 运行命令
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("crew execution failed: %w", err)
	}

	duration := time.Since(startTime)
	log.Info("Crew执行完成",
		logger.Field{Key: "duration", Value: duration},
	)

	return nil
}

// runFlowProject 运行Flow项目
func runFlowProject(ctx context.Context, config *config.ProjectConfig,
	projectRoot string, verbose bool, inputFile, outputDir string,
	timeout time.Duration, development bool, log logger.Logger) error {

	log.Info("运行Flow项目", logger.Field{Key: "name", Value: config.Name})

	// TODO: 实现Flow项目运行逻辑
	// 目前Flow功能正在开发中

	log.Warn("Flow项目运行功能正在开发中")
	fmt.Printf(`
⚠️  Flow项目运行功能正在开发中

📋 当前支持的功能：
- ✅ 基础项目结构创建
- ⏳ Flow工作流执行 (开发中)
- ⏳ 复杂流程编排 (开发中)

🚀 替代方案：
1. 将Flow项目转换为Crew项目
2. 使用基础的Go程序运行

💡 获取更新：
- 关注项目仓库获取最新进展
- 查看文档了解开发路线图

`)

	return fmt.Errorf("flow project execution is not yet implemented")
}

// checkEnvironmentVariables 检查必要的环境变量
func checkEnvironmentVariables(config *config.ProjectConfig, log logger.Logger) error {
	requiredEnvVars := make(map[string]string)

	// 根据LLM提供商检查相应的环境变量
	switch config.LLM.Provider {
	case "openai":
		requiredEnvVars["OPENAI_API_KEY"] = "OpenAI API密钥"
	case "anthropic":
		requiredEnvVars["ANTHROPIC_API_KEY"] = "Anthropic API密钥"
	case "openrouter":
		requiredEnvVars["OPENROUTER_API_KEY"] = "OpenRouter API密钥"
	}

	var missingVars []string
	for envVar, description := range requiredEnvVars {
		if value := os.Getenv(envVar); value == "" {
			missingVars = append(missingVars, fmt.Sprintf("%s (%s)", envVar, description))
		} else {
			// 验证API密钥格式
			if err := utils.ValidateAPIKey(value); err != nil {
				log.Warn("API密钥格式可能有问题",
					logger.Field{Key: "env_var", Value: envVar},
					logger.Field{Key: "error", Value: err.Error()},
				)
			}
		}
	}

	if len(missingVars) > 0 {
		log.Error("缺少必要的环境变量")
		fmt.Printf(`
❌ 缺少必要的环境变量：

`)
		for _, envVar := range missingVars {
			fmt.Printf("   - %s\n", envVar)
		}
		fmt.Printf(`
🔧 解决方案：
1. 创建 .env 文件：cp .env.example .env
2. 编辑 .env 文件，设置相应的API密钥
3. 重新运行项目

💡 获取API密钥：
- OpenAI: https://platform.openai.com/api-keys
- Anthropic: https://console.anthropic.com/
- OpenRouter: https://openrouter.ai/keys

`)
		return fmt.Errorf("missing required environment variables: %v", missingVars)
	}

	return nil
}

// watchForChanges 监听文件变化（开发模式）
func watchForChanges(ctx context.Context, projectRoot string,
	restartFunc func() error, log logger.Logger) error {

	// TODO: 实现文件监听和热重载功能
	// 可以使用fsnotify包来实现文件系统监听

	log.Info("文件监听功能正在开发中...")
	return nil
}
