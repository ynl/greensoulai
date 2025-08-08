# 🌿 GreenSoul AI

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/ynl/greensoulai/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/ynl/greensoulai)](https://goreportcard.com/report/github.com/ynl/greensoulai)

> **让AI智能体在Go生态中自然生长**

## 🌱 关于 GreenSoul AI

GreenSoul AI 是一个用 Go 语言构建的多智能体协作框架，灵感来源于 crewAI 的设计理念。我们相信 AI 的未来不仅在于单个模型的能力，更在于多个智能体如何优雅地协同工作。

这是一个正在成长的开源项目，我们的目标是为 Go 开发者提供一个简洁、高效、可扩展的多智能体开发框架。如果您也对多智能体系统感兴趣，欢迎加入我们的旅程。

在 Go 生态中，虽已拥有丰富的 LLM/RAG 组件，但长期缺少一个面向生产的一体化多智能体“应用层”框架。GreenSoul AI 的使命，就是以 Go 的类型安全与高并发为基石，补齐这块关键拼图。

## 🎯 我们的愿景

我们梦想着构建一个充满活力的 Go 语言多智能体生态系统，让开发者能够：

- **补齐生态空白** - 在 Go 中提供完整的应用层抽象（Agent/Crew/Workflow/Memory/Evaluation）
- **轻松创建智能体** - 用简洁的 Go 代码定义智能体的角色、目标和能力
- **自然地协作** - 让多个智能体像团队一样自然地协同工作
- **与现有系统融合** - 无缝集成到您的 Go 应用和微服务架构中
- **持续学习成长** - 通过社区的力量不断完善和扩展框架能力

### 为什么选择 Go

- **原生并发优势**：goroutine + channel 天然适配多智能体并行协作与工作流编排
- **工程与可运维**：类型安全、静态编译、单可执行文件，便于微服务/边缘/内网落地
- **生态与集成**：易与现有 Go 后端、队列、存储、监控体系对接

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

### 与现有库的不同

- **不止是 LLM/RAG SDK**：提供 Agent、Crew、Workflow、Memory、Evaluation 等完整“应用层”抽象
- **Go 原生并发与事件**：工作流与事件总线以并行为一等公民，内建执行轨迹与并行效率指标
- **对齐 crewAI 概念**：API/语义亲和，便于从 Python 生态平滑迁移

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

提示：也可直接运行并行工作流示例 `examples/workflow/simple_usage.go`，快速体验 Job/Trigger 的并行编排与性能指标。

## 🌸 Garden 场景（用户故事与产品应用）

### 用户故事（多轮群聊协作）
- 你希望在 10x10m 的花园里进行生态友好的种植布局，让不同花卉“自行协商”。
- 我们将 5 种花建模为 5 个智能体：`Rose/玫瑰`、`Sunflower/向日葵`、`Lavender/薰衣草`、`Lily/百合`、`Tulip/郁金香`。
- 他们按回合（多轮顺序任务）依次发言：
  1) 自述需求（光照/排水/土壤/季节性）
  2) 基于前文“聚合上下文”回应/补充/妥协
  3) 汇总偏好与约束
  4) 综合协商 + 最终布局
- 输出以“对话形式”实时打印，最终给出结构化的布局建议思路。

### 运行（仅 OpenRouter）
- 环境变量：
  - `OPENROUTER_API_KEY`: OpenRouter 密钥（必需）
  - `GARDEN_MODEL`（可选，默认 `moonshotai/kimi-k2:free`）
  - `GARDEN_VERBOSE=0`：仅打印对话，不暴露实现日志
  - `GARDEN_DELAY_MS=300`：对话停顿（毫秒）
  - `GARDEN_DELAY_JITTER_PCT=30`：延迟抖动百分比（可模拟随机性）
  - `GARDEN_SHUFFLE_ROUNDS=1`：每回合发言随机顺序

```bash
# 示例（只保留对话，带轻微随机停顿与回合内乱序）
OPENROUTER_API_KEY=sk-*** \
GARDEN_VERBOSE=0 GARDEN_DELAY_MS=300 GARDEN_DELAY_JITTER_PCT=30 GARDEN_SHUFFLE_ROUNDS=1 \
go run ./examples/garden/cmd
```

### 🎥 演示视频
您可以直接在 GitHub 上播放或下载：

![GreenSoul Garden Demo](docs/greensoulgarden.gif)

> 若需高清版，见 `docs/greensoulgarden.mp4` 或点击 "View raw" 下载。

### 产出与体验
- 终端滚动输出：
  - `💬 Rose：...`、`💬 Sunflower：...` 等逐条“发言”
  - 对话节奏可被 `GARDEN_DELAY_*` 调整，默认隐藏内部实现日志
- 最终在控制台底部打印“运行成功与聚合文本摘要”。

### 产品应用场景
- **园林规划/景观设计**：多角色（光照/排水/季节/授粉）偏好冲突自动化博弈与妥协生成初稿。
- **智慧农业/设施园艺**：按作物/花卉“角色化”建模，自动给出分区/轮作/邻里组合建议。
- **教育演示/工作坊**：将“多智能体协作”以具象、可视化的对话方式呈现，便于非技术人员理解。
- **A/B 场景对比**：通过修改 `inputs` 与回合设置，快速形成不同布局方案以供评审。

### 技术映射
- Crew（花园）：`examples/garden/garden.go` 内部 `Crew.ProcessSequential` 按回合推进。
- Agents（花朵）：各自 `role/goal/backstory`，使用 OpenRouter LLM 生成“发言”。
- Tasks（多轮顺序）：前文输出自动聚合进 `aggregated_context`，后续发言基于“上下文”回应。
- 事件/输出：通过 `TaskCallback` 以“💬 发言”逐条打印，避免暴露底层实现细节。

> 小贴士：若需要主持人/仲裁者，可将流程切换为分层模式（Hierarchical），引入“园丁”作为 Manager 协调与指派。

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
- [ ] 公共 API 稳定化（逐步将 `internal/` 能力提炼到 `pkg/`，形成可复用接口）

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
- [最终工作流设计说明](docs/FINAL_FLOW_DESIGN.md) - 命名/接口/并行指标的最终版说明
- [项目结构与最佳实践](docs/PROJECT_STRUCTURE.md) - 目录布局与依赖方向

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