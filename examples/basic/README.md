# 基础示例

本目录包含GreenSoulAI的基础使用示例，适合初学者了解框架的基本概念和用法。

## 示例列表

### 1. simple_agent.go
演示如何创建一个简单的智能体和事件系统的基本用法。

**运行方式：**
```bash
cd examples/basic
go run simple_agent.go
```

**学习要点：**
- 创建日志器
- 初始化事件总线
- 注册事件监听器
- 发射和处理事件

## 前置条件

确保你已经安装了Go 1.21或更高版本：

```bash
go version
```

## 运行示例

1. 克隆项目：
```bash
git clone https://github.com/ynl/greensoulai.git
cd greensoulai
```

2. 安装依赖：
```bash
go mod tidy
```

3. 运行示例：
```bash
cd examples/basic
go run simple_agent.go
```

## 下一步

- 查看 [高级示例](../advanced/README.md) 了解更复杂的用法
- 阅读 [API文档](../../docs/api/README.md) 了解详细的API说明
- 参考 [使用指南](../../docs/guides/README.md) 获取最佳实践