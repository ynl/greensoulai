# 🌿 GreenSoul AI

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/ynl/greensoulai/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/ynl/greensoulai)](https://goreportcard.com/report/github.com/ynl/greensoulai)

> **让AI智能体在Go生态中自然生长**

## 🌱 关于 GreenSoul AI

GreenSoul AI 是一个用 Go 语言构建的多智能体协作框架，灵感来源于 crewAI 的设计理念。我们相信 AI 的未来不仅在于单个模型的能力，更在于多个智能体如何优雅地协同工作。

这是一个正在成长的开源项目，我们的目标是为 Go 开发者提供一个简洁、高效、可扩展的多智能体开发框架。如果您也对多智能体系统感兴趣，欢迎加入我们的旅程。

## 🎯 我们的愿景

我们梦想着构建一个充满活力的 Go 语言多智能体生态系统，让开发者能够：

- **轻松创建智能体** - 用简洁的 Go 代码定义智能体的角色、目标和能力
- **自然地协作** - 让多个智能体像团队一样自然地协同工作
- **与现有系统融合** - 无缝集成到您的 Go 应用和微服务架构中
- **持续学习成长** - 通过社区的力量不断完善和扩展框架能力

## ✨ 核心特性

### 当前已实现
- 🤖 **智能体系统** - 基础的 Agent 接口和实现框架
- 👥 **团队协作** - Crew 管理和任务分配机制
- 📋 **任务执行** - 灵活的任务定义和执行引擎
- 🔄 **工作流编排** - 并行作业调度和状态管理
- 🎯 **事件驱动** - 完整的事件总线和监听器系统
- 🔧 **LLM 集成** - 支持 OpenAI 和可扩展的 Provider 机制
- 📝 **结构化日志** - 统一的日志和错误处理

### 正在开发中
- 🧠 **记忆系统** - 短期、长期和实体记忆管理
- 🛠️ **工具生态** - 丰富的内置工具和自定义工具支持
- 📚 **知识管理** - 文档检索和知识库集成
- 🎓 **持续学习** - 智能体训练和优化机制
- 🌐 **分布式支持** - 跨服务的智能体协作

## 🚀 快速开始

### 安装

```bash
go get github.com/ynl/greensoulai
```

### 第一个智能体

创建一个简单的研究助手：

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/ynl/greensoulai/internal/agent"
    "github.com/ynl/greensoulai/internal/llm"
    "github.com/ynl/greensoulai/pkg/events"
    "github.com/ynl/greensoulai/pkg/logger"
)

func main() {
    // 初始化基础组件
    logger := logger.NewConsoleLogger()
    eventBus := events.NewEventBus(logger)
    
    // 配置 LLM（需要设置 OPENAI_API_KEY 环境变量）
    llmConfig := &llm.Config{
        Provider: "openai",
        Model:    "gpt-3.5-turbo",
        APIKey:   os.Getenv("OPENAI_API_KEY"),
    }
    
    llmProvider, err := llm.CreateLLM(llmConfig)
    if err != nil {
        log.Fatal(err)
    }
    
    // 创建研究助手智能体
    researcher := agent.NewBaseAgent(
        agent.AgentConfig{
            Role:      "研究助手",
            Goal:      "帮助用户研究和分析信息",
            Backstory: "你是一位经验丰富的研究员，善于收集和分析各种信息",
            LLM:       llmProvider,
            EventBus:  eventBus,
            Logger:    logger,
        },
    )
    
    // 创建并执行任务
    task := agent.NewTask(
        "研究 Go 语言的并发模型",
        "提供 Go 语言并发模型的详细说明，包括 goroutine 和 channel 的工作原理",
    )
    
    ctx := context.Background()
    output, err := researcher.Execute(ctx, task)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("研究结果:\n%s\n", output.Raw)
}
```

### 团队协作示例

多个智能体协同工作：

```go
package main

import (
    "context"
    "log"
    
    "github.com/ynl/greensoulai/internal/agent"
    "github.com/ynl/greensoulai/internal/crew"
)

func main() {
    // 创建一个研究团队
    researchCrew := crew.NewBaseCrew(
        &crew.CrewConfig{
            Name:    "产品研究团队",
            Process: crew.ProcessSequential,
        },
        eventBus,
        logger,
    )
    
    // 添加不同角色的智能体
    researchCrew.AddAgent(marketAnalyst)   // 市场分析师
    researchCrew.AddAgent(techExpert)      // 技术专家
    researchCrew.AddAgent(contentWriter)   // 内容撰写者
    
    // 定义任务链
    researchCrew.AddTask(marketResearchTask)
    researchCrew.AddTask(technicalAnalysisTask)
    researchCrew.AddTask(reportWritingTask)
    
    // 启动团队工作
    ctx := context.Background()
    result, err := researchCrew.Kickoff(ctx, map[string]interface{}{
        "product": "AI 助手应用",
        "target":  "开发者群体",
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("团队工作完成: %s", result.FinalOutput)
}
```

## 🏗️ 项目结构

```
greensoulai/
├── cmd/                    # 命令行工具
├── internal/              # 核心实现
│   ├── agent/            # 智能体系统
│   ├── crew/             # 团队协作
│   ├── llm/              # 语言模型集成
│   ├── memory/           # 记忆管理
│   └── knowledge/        # 知识管理
├── pkg/                   # 公共库
│   ├── events/           # 事件系统
│   ├── logger/           # 日志系统
│   └── flow/             # 工作流引擎
├── examples/              # 示例代码
└── docs/                  # 文档
```

## 🛠️ 开发路线图

我们正在积极开发以下功能：

### 近期目标（v0.1.0）
- [x] 基础智能体系统
- [x] 事件驱动架构
- [x] OpenAI 集成
- [ ] 完整的任务执行流程
- [ ] 基础工具系统
- [ ] 简单的记忆管理

### 中期目标（v0.2.0）
- [ ] 更多 LLM Provider 支持
- [ ] 高级工具集成
- [ ] 知识库管理
- [ ] Web UI 界面
- [ ] 性能优化

### 长期愿景
- [ ] 分布式智能体协作
- [ ] 智能体市场
- [ ] 可视化工作流设计器
- [ ] 与主流框架集成

## 🤝 参与贡献

GreenSoul AI 是一个开源项目，我们欢迎所有形式的贡献！

### 如何贡献

1. **报告问题** - 发现 bug 或有建议？[提交 Issue](https://github.com/ynl/greensoulai/issues)
2. **贡献代码** - Fork 项目，创建分支，提交 PR
3. **完善文档** - 帮助改进文档和示例
4. **分享经验** - 在项目中使用？分享您的经验和最佳实践
5. **传播项目** - Star 项目，向朋友推荐

### 开发环境设置

```bash
# 克隆项目
git clone https://github.com/ynl/greensoulai.git
cd greensoulai

# 安装依赖
go mod download

# 运行测试
make test

# 构建项目
make build
```

### 行为准则

我们致力于提供友好、包容的社区环境。请阅读并遵守我们的[行为准则](CODE_OF_CONDUCT.md)。

## 📚 文档和资源

- [入门指南](docs/getting-started.md) - 详细的入门教程
- [API 文档](docs/api-reference.md) - 完整的 API 参考
- [架构设计](docs/WORKFLOW_ARCHITECTURE.md) - 系统架构说明
- [示例代码](examples/) - 各种使用场景的示例

## 🙏 致谢

- 感谢 [crewAI](https://github.com/joaomdmoura/crewAI) 项目的灵感和设计理念
- 感谢所有贡献者的努力和支持
- 感谢 Go 社区提供的优秀工具和库

## 📄 许可证

GreenSoul AI 采用 MIT 许可证，详见 [LICENSE](LICENSE) 文件。

## 💬 联系我们

- **GitHub Issues**: [问题和建议](https://github.com/ynl/greensoulai/issues)
- **Discussions**: [社区讨论](https://github.com/ynl/greensoulai/discussions)
- **Email**: greensoulai@example.com

---

<div align="center">

**🌿 GreenSoul AI - 让智能体在 Go 生态中自然生长**

如果这个项目对您有帮助，请给我们一个 ⭐️ Star！

</div>