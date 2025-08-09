# AI研究助手 - 完整端到端示例

## 🎯 概述

这是一个完整的端到端示例，展示了 GreenSoulAI 框架的核心功能：
- **Agent + LLM 完整集成**
- **智能工具使用**
- **Crew 团队协作**
- **事件系统监控**
- **真实的 OpenAI API 调用**

## 🚀 快速开始

### 1. 前置要求

- Go 1.21+
- OpenRouter API 密钥（支持免费的 Kimi 模型）

### 2. 设置 API 密钥

```bash
# 方法1: OpenRouter API 密钥（推荐，支持Kimi免费模型）
export OPENROUTER_API_KEY="sk-or-your-openrouter-api-key-here"

# 方法2: OpenAI API 密钥（传统方式）
export OPENAI_API_KEY="sk-your-openai-api-key-here"

# 获取免费API密钥：https://openrouter.ai/
```

### 3. 运行示例

#### 🌟 简化测试（推荐新手）

```bash
cd examples/complete
go run simple_kimi_demo.go
```

**预期输出：**
```
🚀 OpenRouter Kimi API 简化测试
✅ 模型创建成功: moonshotai/kimi-k2:free
💬 测试1: 基本中文对话
🤖 回复: 你好！我是Kimi，一个由月之暗面科技有限公司训练的大语言模型...
📊 统计: 69 tokens (提示: 28, 完成: 41)
```

#### 🔥 完整功能演示

```bash
cd examples/complete
go run ai_research_assistant.go       # 完整研究助手示例
go run quick_start.go                 # 快速开始示例
```

## 📊 演示场景

### 场景1: 单个 Agent 使用工具进行研究
- **研究员 Agent** 配备多种研究工具
- 展示工具的智能选择和使用
- 真实的 LLM 推理和决策过程
- 详细的执行统计和监控

### 场景2: Crew 团队协作研究
- **多专业 Agent** 协作完成复杂任务
- 数据收集专家 + 趋势分析师 + 技术评估专家
- **Sequential 执行模式**展示任务依赖关系
- 团队协作成果整合

### 场景3: 复杂工作流
- **市场研究 → 产品需求分析**工作流
- 展示任务间的上下文传递
- 多阶段决策过程
- 业务流程自动化示例

## 🔧 核心功能展示

### Agent 功能
```go
// 创建专业 Agent
researcherConfig := agent.AgentConfig{
    Role:      "高级技术研究员",
    Goal:      "对新兴技术进行全面研究并提供详细洞察",
    Backstory: "你是一位经验丰富的技术研究专家，在AI、软件开发和新兴技术趋势方面有深度专业知识。你总是用中文回答。",
    LLM:       llmInstance,  // 真实的 Kimi LLM
    EventBus:  eventBus,
    Logger:    baseLogger,
}
```

### 工具集成
```go
// Agent 自动选择和使用工具
- 网络搜索工具 (模拟)
- 数据分析工具  
- 文档生成工具
```

### LLM 集成
```go
// 真实的 OpenAI API 调用
config := &llm.Config{
    Provider: "openai",
    Model:    "gpt-4o-mini",
    APIKey:   apiKey,
    // ... 其他配置
}
```

### 事件监控
```go
// 完整的事件系统监控
- Agent 执行事件
- LLM 调用事件  
- 工具使用事件
- Crew 协作事件
```

## 📈 输出示例

```
🚀 GreenSoulAI 完整端到端示例：AI研究助手
===============================================

🔧 初始化系统组件...
🤖 创建OpenAI LLM实例...
✅ 成功创建 gpt-4o-mini 实例
🎯 支持函数调用: true

==================================================
📊 场景1: 单个Agent使用工具进行技术研究
🤖 Agent开始执行任务: map[agent:Senior Technology Researcher ...]
🧠 LLM调用开始: gpt-4o-mini
🔧 工具调用: web_search
🔧 工具调用: data_analysis
🧠 LLM调用完成: 1250ms
✅ Agent任务完成: Senior Technology Researcher
✅ 任务完成! 耗时: 2.3s
📄 生成内容长度: 1847 字符
🔢 使用Token: 432

📋 研究结果摘要:
----------------------------------------
# Large Language Models (LLMs) State & Trends Report 2024

## Executive Summary
Based on comprehensive research and analysis, the current state of Large Language Models...

## Latest Model Architectures
1. **Transformer Variants**: Continued evolution of the transformer architecture...
2. **Mixture of Experts (MoE)**: Enhanced efficiency through specialized sub-networks...
... (更多内容已省略)

🔧 工具使用统计:
   - web_search: 2次使用
   - data_analysis: 1次使用
   - document_generator: 1次使用
```

## 💡 关键特性

### ✅ 完全可工作
- 使用真实的 OpenAI API
- 完整的错误处理
- 实际的 Token 计费和统计

### ✅ 生产就绪
- 完善的日志记录
- 事件系统监控
- 超时控制和重试机制

### ✅ 易于扩展
- 模块化的 Agent 设计
- 可插拔的工具系统
- 灵活的 Crew 配置

## 🛠️ 自定义和扩展

### 添加新工具
```go
func createMyCustomTool() agent.Tool {
    return agent.NewBaseTool(
        "my_tool",
        "Description of my tool",
        func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
            // 工具实现逻辑
            return result, nil
        },
    )
}
```

### 创建新 Agent
```go
config := agent.AgentConfig{
    Role:      "你的Agent角色",
    Goal:      "你的Agent目标", 
    Backstory: "你的Agent背景故事，记得要求用中文回答",
    LLM:       llmInstance,
    EventBus:  eventBus,
    Logger:    logger,
}
```

### 配置不同的执行模式
```go
// Sequential 模式 - 按顺序执行
crewConfig := &crew.CrewConfig{
    Process: crew.ProcessSequential,
    Verbose: true,
}

// Hierarchical 模式 - 层级管理
crewConfig := &crew.CrewConfig{
    Process: crew.ProcessHierarchical,
    ManagerLLM: managerLLM,
}
```

## 🔍 学习要点

这个示例完整展示了：

1. **Agent 设计模式** - 如何定义专业化的 AI Agent
2. **工具集成** - 如何让 Agent 智能使用工具
3. **LLM 交互** - 如何与真实的 LLM API 交互
4. **团队协作** - 如何让多个 Agent 协作完成复杂任务
5. **事件监控** - 如何监控整个 AI 工作流的执行
6. **错误处理** - 如何处理各种异常情况
7. **性能优化** - 如何控制成本和提升效率

## 📚 相关文档

- [Agent 系统文档](../../docs/api/agents.md)
- [LLM 集成指南](../../docs/guides/llm-integration.md) 
- [工具开发指南](../../docs/guides/tool-development.md)
- [Crew 协作模式](../../docs/guides/crew-collaboration.md)

## 💬 问题和反馈

如果在运行示例时遇到问题，请检查：

1. **API 密钥设置是否正确**
2. **网络连接是否正常**
3. **Go 版本是否兼容**
4. **依赖包是否完整**

更多帮助请查看 [故障排除指南](../../docs/troubleshooting.md) 或提交 Issue。
