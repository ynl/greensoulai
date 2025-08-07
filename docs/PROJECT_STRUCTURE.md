# GreenSoulAI 项目结构说明

## 📁 目录结构

本项目遵循Go语言的标准项目布局和最佳实践：

```
greensoulai/
├── cmd/                    # 命令行应用程序
│   └── greensoulai/       # 主程序入口
├── internal/              # 私有应用程序代码
├── pkg/                   # 公共库代码
├── examples/              # 示例代码
├── docs/                  # 文档
├── scripts/               # 构建和部署脚本
├── deployments/           # 部署配置
├── crewAI/               # Python版本参考（仅供开发参考）
└── tests/                # 集成测试
```

## 📂 详细说明

### `/cmd` - 命令行应用程序

存放项目的主要应用程序。每个应用程序都有自己的目录，目录名对应可执行文件名。

- `cmd/greensoulai/` - 主程序入口点
- 不要在这个目录中放置太多代码，导入并调用 `/internal` 和 `/pkg` 中的代码

### `/internal` - 私有应用程序代码

存放不希望被其他应用程序或库导入的代码。这是由Go编译器强制执行的布局模式。

- `internal/agent/` - 智能体实现
- `internal/crew/` - 团队协作逻辑
- `internal/task/` - 任务管理
- `internal/flow/` - 工作流引擎
- `internal/tools/` - 工具系统
- `internal/memory/` - 记忆管理
- `internal/llm/` - 语言模型集成
- `internal/knowledge/` - 知识管理

### `/pkg` - 公共库代码

存放可以被外部应用程序使用的库代码。其他项目可以导入这些库。

- `pkg/events/` - 事件系统
- `pkg/logger/` - 日志系统
- `pkg/security/` - 安全模块
- `pkg/async/` - 异步执行
- `pkg/errors/` - 错误定义
- `pkg/config/` - 配置管理

### `/examples` - 示例代码

应用程序或公共库的示例代码。

- `examples/basic/` - 基础使用示例
- `examples/advanced/` - 高级功能示例
- `examples/enterprise/` - 企业级示例

### `/docs` - 文档

设计文档和用户文档。

- `docs/api/` - API文档
- `docs/guides/` - 使用指南
- `docs/examples/` - 示例文档

### `/scripts` - 构建和部署脚本

用于执行各种构建、安装、分析等操作的脚本。

### `/deployments` - 部署配置

IaaS、PaaS、系统和容器编排部署配置和模板。

### `/crewAI` - Python版本参考

这个目录包含crewAI的Python版本实现，仅供开发时参考对照：

- 用于理解原始设计理念
- 确保API兼容性
- 功能对比和验证
- **注意**: 这个目录不参与Go项目的构建过程

### `/tests` - 集成测试

额外的外部测试应用程序和测试数据。

## 🎯 设计原则

### 1. 关注点分离

- `cmd/` - 应用程序入口点
- `internal/` - 业务逻辑实现
- `pkg/` - 可重用的库代码

### 2. 依赖方向

```
cmd/ → internal/ → pkg/
```

- `cmd/` 可以导入 `internal/` 和 `pkg/`
- `internal/` 可以导入 `pkg/`
- `pkg/` 应该是自包含的，最小化外部依赖

### 3. 包命名约定

- 使用简短、描述性的包名
- 避免使用 `common`、`util`、`shared` 等通用名称
- 包名应该反映其功能而非实现

### 4. 测试策略

- 单元测试与源代码放在同一包中（`*_test.go`）
- 集成测试放在 `/tests` 目录中
- 基准测试使用 `*_bench_test.go` 命名

## 🔄 与crewAI Python版本的关系

### 参考模式

1. **API兼容性**: 保持与Python版本的API一致性
2. **功能对等**: 实现相同的核心功能
3. **设计理念**: 遵循相同的架构设计
4. **扩展增强**: 利用Go语言特性进行优化

### 开发流程

1. 查看 `crewAI/` 中的Python实现
2. 理解功能需求和设计理念
3. 在Go版本中实现对应功能
4. 确保API兼容性
5. 添加Go特有的优化

## 📋 最佳实践

### 1. 代码组织

- 每个包应该有明确的职责
- 避免循环依赖
- 使用接口定义契约

### 2. 错误处理

- 使用 `pkg/errors` 定义标准错误类型
- 提供有意义的错误信息
- 支持错误链追踪

### 3. 日志记录

- 使用结构化日志（`pkg/logger`）
- 记录关键操作和错误
- 支持不同日志级别

### 4. 配置管理

- 使用 `pkg/config` 统一配置管理
- 支持环境变量和配置文件
- 提供配置验证

### 5. 测试覆盖

- 目标覆盖率：80%+
- 包含单元测试、集成测试和基准测试
- 使用表驱动测试模式

## 🔧 开发工具

### Makefile 目标

- `make build` - 构建项目
- `make test` - 运行测试
- `make coverage` - 生成覆盖率报告
- `make lint` - 代码检查
- `make fmt` - 代码格式化

### 推荐工具

- **IDE**: VS Code with Go extension
- **Linter**: golangci-lint
- **Testing**: testify
- **Mocking**: gomock
- **Documentation**: godoc
