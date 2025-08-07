# GreenSoulAI

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Test Coverage](https://img.shields.io/badge/coverage-80%25-brightgreen.svg)](https://github.com/ynl/greensoulai)

**GreenSoulAI** 是一个基于Go语言实现的多智能体协作AI框架，参考并兼容crewAI的设计理念，提供更高性能和更好的并发支持。

## 🌟 特性

- 🚀 **高性能**: 基于Go语言，提供比Python版本2-3倍的性能提升
- 🔄 **并发友好**: 充分利用Go的goroutine和channel机制
- 🛡️ **类型安全**: 强类型系统减少运行时错误
- 📦 **单文件部署**: 编译为单个二进制文件，无运行时依赖
- 🔧 **企业就绪**: 内置监控、安全、容错等企业级功能
- 🔌 **兼容性**: 与crewAI Python版本API兼容

## 🏗️ 项目结构

```
greensoulai/
├── cmd/                    # 命令行应用程序
│   └── greensoulai/       # 主程序入口
├── internal/              # 私有应用程序代码
│   ├── agent/            # 智能体实现
│   ├── crew/             # 团队协作
│   ├── task/             # 任务管理
│   ├── flow/             # 工作流引擎
│   ├── tools/            # 工具系统
│   ├── memory/           # 记忆管理
│   ├── llm/              # 语言模型
│   └── knowledge/        # 知识管理
├── pkg/                   # 公共库代码
│   ├── events/           # 事件系统
│   ├── logger/           # 日志系统
│   ├── security/         # 安全模块
│   ├── async/            # 异步执行
│   ├── errors/           # 错误定义
│   └── config/           # 配置管理
├── examples/             # 示例代码
│   ├── basic/           # 基础示例
│   ├── advanced/        # 高级示例
│   └── enterprise/      # 企业级示例
├── docs/                 # 文档
│   ├── api/             # API文档
│   ├── guides/          # 使用指南
│   └── examples/        # 示例文档
├── scripts/              # 构建和部署脚本
├── deployments/          # 部署配置
├── crewAI/              # Python版本参考（仅供开发参考）
└── tests/               # 集成测试
```

## 🚀 快速开始

### 安装

```bash
go install github.com/ynl/greensoulai/cmd/greensoulai@latest
```

### 基本使用

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/ynl/greensoulai/internal/agent"
    "github.com/ynl/greensoulai/internal/crew"
    "github.com/ynl/greensoulai/internal/task"
    "github.com/ynl/greensoulai/pkg/logger"
)

func main() {
    // 创建日志器
    logger := logger.NewConsoleLogger()
    
    // 创建智能体
    researcher := agent.NewAgent(
        "Researcher",
        "收集和分析信息",
        "你是一个专业的研究员",
        logger,
    )
    
    // 创建任务
    researchTask := task.NewTask(
        "研究AI发展趋势",
        "提供详细的AI发展趋势报告",
        logger,
    )
    
    // 创建团队
    crew := crew.NewCrew("AI Research Team", logger)
    crew.AddAgent(researcher)
    crew.AddTask(researchTask)
    
    // 执行
    ctx := context.Background()
    result, err := crew.Kickoff(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("结果: %s\n", result.Output)
}
```

## 📚 文档

- [API文档](docs/api/README.md)
- [使用指南](docs/guides/README.md)
- [示例代码](examples/README.md)
- [部署指南](docs/deployment.md)

## 🔄 与crewAI Python版本的关系

- `crewAI/` 目录包含Python版本的完整实现，仅供开发时参考对照
- Go版本保持API兼容性，便于从Python版本迁移
- 设计理念和架构保持一致，功能增强

## 🧪 开发

### 构建

```bash
make build
```

### 测试

```bash
make test
```

### 代码覆盖率

```bash
make coverage
```

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🤝 贡献

欢迎贡献代码！请查看 [CONTRIBUTING.md](CONTRIBUTING.md) 了解详情。

## 📞 支持

- 问题反馈: [GitHub Issues](https://github.com/ynl/greensoulai/issues)
- 功能请求: [GitHub Discussions](https://github.com/ynl/greensoulai/discussions)
- 文档: [官方文档](https://greensoulai.dev)
