package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
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
参考并兼容crewAI的设计理念，提供更高性能和更好的并发支持。`,
		Version: fmt.Sprintf("%s (built %s, commit %s)", Version, BuildTime, GitCommit),
	}

	// Add version flag
	rootCmd.SetVersionTemplate(`{{printf "%s\n" .Version}}`)

	// Add subcommands
	rootCmd.AddCommand(
		newRunCommand(log),
		newCreateCommand(log),
		newTestCommand(log),
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

	// Execute command
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Error("命令执行失败", logger.Field{Key: "error", Value: err})
		os.Exit(1)
	}
}

// newRunCommand creates the run command
func newRunCommand(log logger.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [config-file]",
		Short: "运行GreenSoulAI服务",
		Long:  "启动GreenSoulAI服务，可以指定配置文件路径",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile := "config.yaml"
			if len(args) > 0 {
				configFile = args[0]
			}

			log.Info("启动GreenSoulAI服务", logger.Field{Key: "config", Value: configFile})

			// TODO: 实现服务启动逻辑
			log.Info("服务启动成功")

			// 等待中断信号
			<-cmd.Context().Done()
			log.Info("服务已停止")

			return nil
		},
	}

	return cmd
}

// newCreateCommand creates the create command
func newCreateCommand(log logger.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "创建新的项目或组件",
		Long:  "创建新的GreenSoulAI项目、智能体、任务或团队",
	}

	// Add subcommands for create
	cmd.AddCommand(
		&cobra.Command{
			Use:   "project [name]",
			Short: "创建新项目",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				projectName := args[0]
				log.Info("创建新项目", logger.Field{Key: "name", Value: projectName})

				// TODO: 实现项目创建逻辑
				log.Info("项目创建成功", logger.Field{Key: "name", Value: projectName})

				return nil
			},
		},
		&cobra.Command{
			Use:   "agent [name]",
			Short: "创建新智能体",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				agentName := args[0]
				log.Info("创建新智能体", logger.Field{Key: "name", Value: agentName})

				// TODO: 实现智能体创建逻辑
				log.Info("智能体创建成功", logger.Field{Key: "name", Value: agentName})

				return nil
			},
		},
		&cobra.Command{
			Use:   "crew [name]",
			Short: "创建新团队",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				crewName := args[0]
				log.Info("创建新团队", logger.Field{Key: "name", Value: crewName})

				// TODO: 实现团队创建逻辑
				log.Info("团队创建成功", logger.Field{Key: "name", Value: crewName})

				return nil
			},
		},
	)

	return cmd
}

// newTestCommand creates the test command
func newTestCommand(log logger.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "运行测试",
		Long:  "运行项目测试套件",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info("开始运行测试")

			// TODO: 实现测试运行逻辑
			log.Info("测试完成")

			return nil
		},
	}

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
