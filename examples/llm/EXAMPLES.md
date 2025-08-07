# LLM模块示例列表

本目录包含了LLM模块的各种使用示例，从基础到高级功能的完整演示。

## 📁 示例文件

### 1. `basic/main.go` - LLM基础使用
**功能**: 演示LLM模块的基本功能
- ✅ 基础对话调用
- ✅ 流式响应
- ✅ 函数调用
- ✅ 错误处理
- ✅ 事件监听

**运行方式**:
```bash
# 设置OpenAI API Key
export OPENAI_API_KEY="your-openai-key"
cd basic && go run main.go
```

**适用场景**: 
- 初学者入门
- 理解LLM模块基本概念
- 快速验证功能

---

### 2. `openrouter/basic/main.go` - OpenRouter基础集成
**功能**: 演示如何使用OpenRouter作为LLM提供商
- ✅ OpenRouter API配置
- ✅ 免费模型使用
- ✅ 自定义Headers
- ✅ 中英文对话
- ✅ 成本追踪

**运行方式**:
```bash
# 使用环境变量（推荐）
export OPENROUTER_API_KEY="sk-or-v1-your-key"
cd openrouter/basic && go run main.go

# 或直接运行（使用示例Key）
cd openrouter/basic && go run main.go
```

**适用场景**:
- 需要更多模型选择
- 成本敏感的应用
- 想要使用免费模型

---

### 3. `openrouter/advanced/main.go` - OpenRouter高级功能
**功能**: 演示OpenRouter的全部功能
- ✅ 多种测试场景
- ✅ 同步和流式调用
- ✅ 多模型切换测试
- ✅ 详细的错误处理
- ✅ 完整的统计信息

**运行方式**:
```bash
export OPENROUTER_API_KEY="sk-or-v1-your-key"
cd openrouter/advanced && go run main.go
```

**适用场景**:
- 生产环境部署前测试
- 性能和稳定性验证
- 多模型比较

---

## 🚀 快速开始

### 选择合适的示例

1. **完全新手** → 从 `basic_usage.go` 开始
2. **想要免费模型** → 使用 `openrouter_basic.go`
3. **生产环境准备** → 运行 `openrouter_advanced.go`

### 环境配置

#### OpenAI
```bash
export OPENAI_API_KEY="sk-your-openai-key"
```

#### OpenRouter
```bash
export OPENROUTER_API_KEY="sk-or-v1-your-openrouter-key"
```

### 依赖安装
```bash
# 确保在项目根目录
cd /path/to/greensoulai
go mod tidy
```

## 📊 功能对比

| 功能 | basic_usage | openrouter_basic | openrouter_advanced |
|------|-------------|------------------|---------------------|
| 基础对话 | ✅ | ✅ | ✅ |
| 流式响应 | ✅ | ✅ | ✅ |
| 函数调用 | ✅ | ❌ | ❌ |
| 事件系统 | ✅ | ❌ | ❌ |
| 多模型测试 | ❌ | ❌ | ✅ |
| 成本追踪 | ✅ | ✅ | ✅ |
| 错误处理 | 基础 | 基础 | 详细 |
| 自定义Headers | ❌ | ✅ | ✅ |

## 💡 使用技巧

### 1. API Key管理
- 使用环境变量而非硬编码
- 生产环境使用密钥管理服务
- 定期轮换API密钥

### 2. 模型选择
**OpenAI推荐**:
- `gpt-4o-mini` - 成本效益最佳
- `gpt-4` - 最高质量

**OpenRouter推荐**:
- `moonshotai/kimi-k2:free` - 中文免费
- `google/gemini-flash-1.5` - 快速响应
- `anthropic/claude-3-haiku` - 平衡性能

### 3. 性能优化
- 复用LLM实例
- 设置合理的超时时间
- 使用流式响应减少感知延迟
- 监控Token使用量

### 4. 错误处理
- 实现指数退避重试
- 区分不同错误类型
- 设置断路器模式
- 记录详细错误日志

## 🔧 故障排除

### 常见问题

#### 1. "API key not found"
```bash
# 检查环境变量
echo $OPENAI_API_KEY
echo $OPENROUTER_API_KEY

# 重新设置
export OPENAI_API_KEY="your-key"
```

#### 2. "context deadline exceeded"
- 增加超时时间
- 检查网络连接
- 使用较小的模型

#### 3. "rate limit exceeded"
- 降低请求频率
- 使用不同的模型
- 升级API套餐

#### 4. 编译错误
```bash
# 更新依赖
go mod tidy

# 检查Go版本
go version

# 重新构建
go clean -cache
go build
```

## 📈 下一步

运行示例后，您可以：

1. **集成到项目** - 参考示例代码集成到您的应用
2. **自定义配置** - 根据需求调整参数
3. **扩展功能** - 添加自定义工具和处理逻辑
4. **性能测试** - 在您的环境中进行压力测试
5. **监控部署** - 添加监控和告警机制

## 🤝 贡献

欢迎提交新的示例或改进现有示例：

1. Fork项目
2. 创建功能分支
3. 添加示例和文档
4. 提交Pull Request

---

**需要帮助?** 
- 查看 [README.md](./README.md) 获取详细文档
- 运行测试: `go test ./internal/llm/... -v`
- 提交Issue: [GitHub Issues](https://github.com/ynl/greensoulai/issues)
