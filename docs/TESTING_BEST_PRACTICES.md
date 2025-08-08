# Go 测试最佳实践指南

## 概述

本文档概述了项目中 Go 测试的最佳实践，特别是关于 Mock 对象和测试辅助工具的组织和使用。

## 🏆 当前采用的最佳实践

### 1. 文件命名约定

#### ✅ 推荐的命名

```
# 测试文件
*_test.go                    # 标准测试文件
*_integration_test.go        # 集成测试文件  
*_benchmark_test.go          # 基准测试文件

# Mock 对象文件
mock_test.go                 # 包内Mock对象（推荐）
*_mock_test.go              # 特定功能的Mock对象
```

#### ❌ 避免的命名

```
testing_mocks.go            # 不清晰的命名
mock_dependencies.go        # 容易引起循环导入
mocks.go                    # 太通用，不明确
```

### 2. 构建标签（Build Tags）

所有测试相关文件都应包含构建标签：

```go
//go:build test
// +build test

package yourpackage
```

**作用**：
- 防止测试代码被包含在生产构建中
- 明确标识测试相关代码
- 支持条件编译

### 3. 包内 Mock 组织

每个包的 Mock 对象应该放在 `mock_test.go` 文件中：

```
internal/agent/
├── agent.go
├── agent_test.go
├── mock_test.go          # Agent相关的Mock对象
├── tools.go
├── tools_test.go
└── ...
```

### 4. Mock 对象设计原则

#### ✅ 好的设计

```go
// 1. 明确的命名
type MockLLM struct { ... }
type ExtendedMockLLM struct { ... }

// 2. 链式配置支持
func (m *ExtendedMockLLM) WithCallHandler(handler func([]llm.Message)) *ExtendedMockLLM {
    m.onCall = handler
    return m
}

// 3. 接口实现验证
var _ LLM = (*MockLLM)(nil)

// 4. 状态管理方法
func (m *MockLLM) GetCallCount() int { return m.callCount }
func (m *MockLLM) ResetCallCount() { m.callCount = 0 }
```

#### ❌ 避免的设计

```go
// 1. 模糊的命名
type TestLLM struct { ... }
type Helper struct { ... }

// 2. 难以配置的Mock
type MockLLM struct {
    response string  // 太简单，不够灵活
}

// 3. 没有状态管理
// 无法验证调用次数或重置状态
```

## 📁 文件组织结构

### 当前结构（推荐）

```
internal/agent/
├── mock_test.go              # Agent包的Mock对象
├── base_agent_test.go        # Agent核心测试
├── tools_test.go             # 工具测试
├── tools_integration_test.go # 工具集成测试
└── tool_utils_test.go        # 工具辅助函数测试

internal/crew/
├── mock_test.go              # Crew包的Mock对象
├── base_crew_test.go         # Crew核心测试
└── ...
```

### 避免的结构

```
# ❌ 会导致循环导入
internal/testutil/
├── mocks.go                  # 导入其他包，被其他包导入
└── helpers.go

# ❌ 分散的Mock定义
internal/crew/
├── base_crew_test.go         # 包含Mock定义
├── planning/
│   └── mock_dependencies.go # 重复的Mock定义
```

## 🔧 实际使用示例

### 基础测试

```go
//go:build test
// +build test

package agent

func TestAgentExecution(t *testing.T) {
    // 使用包内的Mock对象
    mockResponse := &llm.Response{
        Content: "test response",
        Usage:   llm.Usage{TotalTokens: 20},
        Model:   "mock-model",
    }
    mockLLM := NewMockLLM(mockResponse, false)
    
    // 创建配置
    config := AgentConfig{
        Role:      "Test Agent",
        Goal:      "Test Goal", 
        Backstory: "Test Backstory",
        LLM:       mockLLM,
        Logger:    logger.NewTestLogger(),
        EventBus:  events.NewEventBus(logger.NewTestLogger()),
    }
    
    agent, err := NewBaseAgent(config)
    require.NoError(t, err)
    
    // 执行测试...
}
```

### 高级Mock配置

```go
func TestAgentWithComplexScenario(t *testing.T) {
    // 使用ExtendedMockLLM进行复杂场景测试
    responses := []llm.Response{
        {Content: "First response", Usage: llm.Usage{TotalTokens: 15}},
        {Content: "Second response", Usage: llm.Usage{TotalTokens: 25}},
    }
    
    var capturedPrompts []string
    mockLLM := NewExtendedMockLLM(responses).
        WithCallHandler(func(messages []llm.Message) {
            if len(messages) > 0 {
                if content, ok := messages[len(messages)-1].Content.(string); ok {
                    capturedPrompts = append(capturedPrompts, content)
                }
            }
        })
    
    // 执行测试并验证提示捕获
    // ...
    
    assert.Len(t, capturedPrompts, 1)
    assert.Contains(t, capturedPrompts[0], "expected prompt content")
}
```

### 工具集成测试

```go
func TestAgentWithMockTools(t *testing.T) {
    // Mock LLM返回工具调用
    mockLLM := NewExtendedMockLLM([]llm.Response{
        {
            Content: `{"tool_name": "calculator", "arguments": {"operation": "add", "a": 5, "b": 3}}`,
            Usage:   llm.Usage{TotalTokens: 50},
        },
    })
    
    agent := createTestAgent(mockLLM)
    
    // 添加Mock工具
    mockTool := NewMockTool("calculator", "Math calculator").
        WithExecuteFunc(func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
            op := args["operation"].(string)
            a := args["a"].(float64)
            b := args["b"].(float64)
            if op == "add" {
                return a + b, nil
            }
            return nil, fmt.Errorf("unsupported operation: %s", op)
        })
    
    err := agent.AddTool(mockTool)
    require.NoError(t, err)
    
    // 执行测试...
}
```

## 🚀 迁移指南

### 从旧结构迁移到新结构

1. **重命名文件**：
   ```bash
   # 将testing_mocks.go重命名为mock_test.go
   mv testing_mocks.go mock_test.go
   ```

2. **添加构建标签**：
   ```go
   //go:build test
   // +build test
   
   package yourpackage
   ```

3. **移除重复定义**：
   - 检查所有测试文件中的重复Mock定义
   - 统一使用包内的mock_test.go中的Mock对象
   - 删除不必要的单独Mock文件

4. **更新测试导入**：
   ```go
   // 不需要导入外部测试包
   // 直接使用包内的Mock对象
   mockLLM := NewMockLLM(response, false)
   ```

## 🔍 验证清单

在重构测试代码时，检查以下项目：

- [ ] 所有测试文件都有适当的构建标签
- [ ] Mock对象命名清晰且一致
- [ ] 没有循环导入问题
- [ ] 重复的Mock定义已被移除
- [ ] Mock对象实现了正确的接口
- [ ] 提供了状态管理方法（如调用计数）
- [ ] 支持链式配置（如果适用）
- [ ] 文件命名遵循标准约定

## 📚 参考资源

- [Effective Go - Testing](https://golang.org/doc/effective_go#testing)
- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify - Testing Toolkit](https://github.com/stretchr/testify)
- [Build Constraints](https://pkg.go.dev/go/build#hdr-Build_Constraints)

## 🏁 结论

通过遵循这些最佳实践：

1. **避免循环导入** - 将Mock对象保持在各自的包内
2. **清晰的命名约定** - 使用标准的文件命名模式
3. **适当的构建标签** - 确保测试代码不会泄漏到生产环境
4. **良好的Mock设计** - 提供灵活且易于使用的Mock对象
5. **统一的组织结构** - 减少代码重复，提高可维护性

这样的结构既符合 Go 语言的惯例，又能有效地组织测试代码，提高开发效率和代码质量。
