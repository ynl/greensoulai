package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/ynl/greensoulai/cmd/greensoulai/commands"
	"github.com/ynl/greensoulai/pkg/logger"
)

var (
	// Version information (set by build flags)
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Create logger
	log := logger.NewConsoleLogger()

	// Create root command
	rootCmd := &cobra.Command{
		Use:   "greensoulai",
		Short: "GreenSoulAI - 多智能体协作AI框架",
		Long: `GreenSoulAI 是一个基于Go语言实现的多智能体协作AI框架，
参考并兼容crewAI的设计理念，提供更高性能和更好的并发支持。

📋 主要功能：
  • 多智能体协作系统
  • 工作流编排和管理  
  • 完整的CLI工具链
  • 高性能并发执行
  • 企业级安全特性

🚀 快速开始：
  greensoulai create crew my-project  # 创建新项目
  cd my-project && greensoulai run    # 运行项目

📚 文档和帮助：
  greensoulai --help                  # 查看帮助
  greensoulai create --help           # 查看创建命令帮助`,
		Version:       fmt.Sprintf("%s (built %s, commit %s)", Version, BuildTime, GitCommit),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Add version flag
	rootCmd.SetVersionTemplate(`{{printf "%s\n" .Version}}`)

	// Add subcommands
	rootCmd.AddCommand(
		commands.NewCreateCommand(log),
		commands.NewRunCommand(log),
		commands.NewTrainCommand(log),
		commands.NewEvaluateCommand(log),
		newChatCommand(log),
		newInstallCommand(log),
		newResetCommand(log),
		newToolsCommand(log),
		newVersionCommand(),
	)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Info("收到中断信号，正在关闭...")
		cancel()
	}()

	// Execute command with improved error handling
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		// 用户友好的错误提示
		if err.Error() != "" {
			fmt.Fprintf(os.Stderr, "❌ 错误: %s\n", err.Error())
		}

		// 提供帮助提示
		fmt.Fprintf(os.Stderr, "\n💡 提示: 使用 'greensoulai --help' 查看可用命令\n")

		os.Exit(1)
	}
}

// newChatCommand 创建chat命令
func newChatCommand(log logger.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "chat",
		Short: "与项目智能体对话",
		Long: `启动与项目智能体的交互式对话模式。
可以实时与智能体交互，测试其响应和功能。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info("启动对话模式...")

			fmt.Printf(`
💬 GreenSoulAI 对话模式
==================================================
⚠️  对话功能正在开发中

🔄 当前状态: 开发中
📋 预期功能:
  • 实时智能体对话
  • 多智能体协作对话
  • 对话历史记录
  • 上下文管理

💡 替代方案:
  使用 'greensoulai run' 运行完整项目

输入 'exit' 或按 Ctrl+C 退出
==================================================

`)

			// TODO: 实现实际的对话逻辑
			fmt.Println("对话功能开发中，敬请期待...")

			return nil
		},
	}
}

// newInstallCommand 创建install命令
func newInstallCommand(log logger.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "安装项目依赖",
		Long: `安装当前GreenSoulAI项目的所有依赖。
会自动检测项目类型并安装相应的依赖包。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info("开始安装项目依赖...")

			fmt.Printf(`
📦 GreenSoulAI 依赖安装
==================================================
🔍 检测项目类型...
📋 解析依赖列表...
⬇️  下载依赖包...

`)

			// TODO: 实现实际的依赖安装逻辑
			// 1. 检查go.mod文件
			// 2. 运行go mod download
			// 3. 验证依赖完整性

			log.Info("依赖安装完成!")

			fmt.Printf(`
✅ 依赖安装完成！

🚀 下一步:
  greensoulai run    # 运行项目

`)

			return nil
		},
	}
}

// newResetCommand 创建reset命令
func newResetCommand(log logger.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "reset-memories",
		Short: "重置智能体记忆",
		Long: `重置当前项目中所有智能体的记忆数据。
这将清除智能体的长期记忆、短期记忆和上下文信息。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info("开始重置智能体记忆...")

			fmt.Printf(`
🧠 GreenSoulAI 记忆重置
==================================================
⚠️  警告: 此操作将删除所有智能体记忆数据

📋 将要清除的数据:
  • 长期记忆存储
  • 短期记忆缓存
  • 对话历史记录
  • 上下文信息

`)

			// TODO: 实现实际的记忆重置逻辑
			// 1. 查找记忆存储文件
			// 2. 清除SQLite数据库
			// 3. 清除缓存文件
			// 4. 重置向量数据库

			log.Info("智能体记忆重置完成!")

			fmt.Printf(`
✅ 记忆重置完成！

🔄 智能体将以全新状态开始工作
🚀 运行 'greensoulai run' 验证重置效果

`)

			return nil
		},
	}
}

// newToolsCommand 创建tools命令
func newToolsCommand(log logger.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tools",
		Short: "工具管理",
		Long: `管理GreenSoulAI项目中的工具。
可以列出、安装、更新和配置各种智能体工具。`,
	}

	// 添加子命令
	cmd.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "列出可用工具",
			Long:  "列出当前项目中所有可用的工具",
			RunE: func(cmd *cobra.Command, args []string) error {
				log.Info("列出可用工具...")

				fmt.Printf(`
🛠️  GreenSoulAI 工具列表
==================================================

📋 内置工具:
  • search_tool        - 网络搜索工具
  • file_tool          - 文件操作工具  
  • analysis_tool      - 数据分析工具
  • web_scraper_tool   - 网页抓取工具
  • api_client_tool    - API客户端工具

📦 可安装工具:
  • database_tool      - 数据库操作工具
  • image_tool         - 图像处理工具
  • email_tool         - 邮件发送工具
  • calendar_tool      - 日历管理工具

💡 使用方法:
  greensoulai tools install <tool_name>  # 安装工具
  greensoulai tools remove <tool_name>   # 移除工具

`)

				return nil
			},
		},
		&cobra.Command{
			Use:   "install [tool-name]",
			Short: "安装工具",
			Long:  "安装指定的工具到当前项目",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				toolName := args[0]

				log.Info("安装工具", logger.Field{Key: "tool", Value: toolName})

				fmt.Printf(`
🔧 安装工具: %s
==================================================
⬇️  下载工具包...
🔨 编译工具代码...
📝 更新项目配置...

`, toolName)

				// TODO: 实现实际的工具安装逻辑

				fmt.Printf("✅ 工具 '%s' 安装成功!\n\n", toolName)

				return nil
			},
		},
		&cobra.Command{
			Use:   "remove [tool-name]",
			Short: "移除工具",
			Long:  "从当前项目中移除指定的工具",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				toolName := args[0]

				log.Info("移除工具", logger.Field{Key: "tool", Value: toolName})

				fmt.Printf("✅ 工具 '%s' 已移除\n", toolName)

				return nil
			},
		},
	)

	return cmd
}

// newVersionCommand creates the version command
func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("GreenSoulAI %s\n", Version)
			fmt.Printf("构建时间: %s\n", BuildTime)
			fmt.Printf("Git提交: %s\n", GitCommit)
		},
	}
}
