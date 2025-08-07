# crewAI-Go 完整实施指引

## 目录

1. [项目概述](#1-项目概述)
2. [技术架构分析](#2-技术架构分析)
3. [Go版本架构设计](#3-go版本架构设计)
4. [详细实施步骤](#4-详细实施步骤)
5. [质量保证体系](#5-质量保证体系)
6. [项目管理](#6-项目管理)
7. [风险管理](#7-风险管理)
8. [成功标准](#8-成功标准)

---

## 1. 项目概述

### 1.1 项目目标

使用 Go 语言重新实现 crewAI 框架，这是一个基于多智能体协作的 AI 任务处理系统。Go 版本将提供更好的性能、并发处理能力和部署便利性。

### 1.2 项目价值

- **性能提升**: 预期比 Python 版本提升 2-3 倍执行效率
- **部署简化**: 单文件部署，无运行时依赖
- **并发优势**: 充分利用 Go 的 goroutine 和 channel 机制
- **生产就绪**: 内置监控、安全、容错等企业级功能

### 1.3 项目规模

- **开发周期**: 12周（3个月）
- **团队规模**: 2-3名 Go 开发工程师
- **代码估算**: 约 20,000+ 行 Go 代码
- **测试覆盖**: 目标 80%+ 覆盖率

---

## 2. 技术架构分析

### 2.1 crewAI Python 核心架构

基于深度源码分析，crewAI 采用分层架构：

```
┌─────────────────────────────────────────────────────────┐
│                    Flow 工作流层                          │
├─────────────────────────────────────────────────────────┤
│                    Crew 团队协作层                         │
├─────────────────────────────────────────────────────────┤
│      Agent 智能体层      │        Task 任务执行层         │
├─────────────────────────────────────────────────────────┤
│  Tools 工具层  │  Memory 记忆层  │  Knowledge 知识层   │
├─────────────────────────────────────────────────────────┤
│              LLM 语言模型抽象层                           │
└─────────────────────────────────────────────────────────┘
```

### 2.2 核心组件功能

#### Agent（智能体）
- 角色定义：role、goal、backstory
- 工具使用和函数调用
- 多类型记忆集成
- 知识源查询
- 执行参数配置

#### Crew（团队）
- 多智能体协作管理
- 执行模式：Sequential、Hierarchical
- 任务调度和结果聚合
- 记忆、缓存、知识管理

#### Task（任务）
- 多输出格式：RAW、JSON、Pydantic
- 条件任务执行
- 护栏验证机制
- 回调和监控

#### Flow（工作流）
- 复杂工作流编排
- 状态管理和持久化
- 路由和监听机制
- 装饰器系统：@start、@listen、@router

### 2.3 新发现的关键功能

**事件系统**
- 基于 blinker 的事件总线
- 完整的组件生命周期事件
- 支持外部监听器

**异步执行**
- 任务异步执行支持
- 线程池管理
- Future 模式结果处理

**安全机制**
- 组件指纹识别
- 安全配置管理
- 确定性 UUID 生成

**人工干预**
- 执行过程中的用户输入
- 交互式任务支持
- 超时控制

---

## 3. Go版本架构设计

### 3.1 包结构设计

```
greensoulai/
├── cmd/                           # 命令行工具
│   └── greensoulai/
│       └── main.go
├── internal/                      # 私有应用程序代码
│   ├── agent/                     # 智能体模块
│   │   ├── agent.go
│   │   ├── executor.go
│   │   ├── cache.go
│   │   └── human_input.go
│   ├── crew/                      # 团队模块
│   │   ├── crew.go
│   │   ├── output.go
│   │   ├── process.go
│   │   └── concurrent.go
│   ├── task/                      # 任务模块
│   │   ├── task.go
│   │   ├── output.go
│   │   ├── conditional.go
│   │   └── guardrail.go
│   ├── flow/                      # 工作流模块
│   │   ├── flow.go
│   │   ├── state.go
│   │   ├── persistence.go
│   │   └── decorators.go
│   ├── tools/                     # 工具模块
│   │   ├── base.go
│   │   ├── structured.go
│   │   ├── usage.go
│   │   └── async.go
│   ├── memory/                    # 记忆模块
│   │   ├── memory.go
│   │   ├── shortterm/
│   │   │   └── shortterm.go
│   │   ├── longterm/
│   │   │   └── longterm.go
│   │   ├── entity/
│   │   │   └── entity.go
│   │   └── storage/
│   │       ├── interface.go
│   │       ├── sqlite.go
│   │       └── vector.go
│   ├── llm/                       # 语言模型模块
│   │   ├── base.go
│   │   ├── llm.go
│   │   ├── rpm_controller.go
│   │   └── providers/
│   │       ├── openai.go
│   │       └── anthropic.go
│   └── knowledge/                 # 知识模块
│       ├── knowledge.go
│       └── sources/
│           ├── document.go
│           └── vector.go
├── pkg/                           # 公共库代码
│   ├── events/                    # 事件系统
│   │   ├── bus.go
│   │   ├── types.go
│   │   └── listener.go
│   ├── security/                  # 安全模块
│   │   ├── fingerprint.go
│   │   └── config.go
│   ├── async/                     # 异步基础
│   │   ├── executor.go
│   │   └── result.go
│   ├── config/                    # 配置管理
│   │   └── config.go
│   ├── logger/                    # 日志系统
│   │   └── logger.go
│   └── errors/                    # 错误定义
│       └── errors.go
├── examples/                      # 示例代码
│   ├── basic/
│   ├── advanced/
│   └── enterprise/
├── docs/                          # 文档
│   ├── api/
│   ├── guides/
│   └── examples/
├── tests/                         # 测试
│   ├── integration/
│   ├── e2e/
│   └── benchmarks/
├── scripts/                       # 构建脚本
├── .github/                       # CI/CD
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
└── README.md
```

### 3.2 核心接口设计

#### 事件系统接口

```go
// pkg/common/events/bus.go
type EventBus interface {
    Emit(ctx context.Context, source interface{}, event Event) error
    Subscribe(eventType string, handler EventHandler) error
    Unsubscribe(eventType string, handler EventHandler) error
}

type Event interface {
    GetType() string
    GetTimestamp() time.Time
    GetSource() interface{}
    GetPayload() map[string]interface{}
}

type EventHandler func(ctx context.Context, event Event) error
```

#### 异步执行接口

```go
// pkg/common/async/executor.go
type AsyncExecutor interface {
    ExecuteAsync(ctx context.Context, task func() (interface{}, error)) <-chan Result
    ExecuteWithTimeout(ctx context.Context, task func() (interface{}, error), timeout time.Duration) <-chan Result
}

type Result struct {
    Value interface{}
    Error error
}
```

#### Agent 接口

```go
// pkg/agent/agent.go
type Agent interface {
    Execute(ctx context.Context, task Task) (*TaskOutput, error)
    ExecuteAsync(ctx context.Context, task Task) (<-chan TaskResult, error)
    ExecuteWithTimeout(ctx context.Context, task Task, timeout time.Duration) (*TaskOutput, error)
    GetRole() string
    GetGoal() string
    AddTool(tool Tool) error
    SetMemory(memory Memory) error
    SetHumanInputHandler(handler HumanInputHandler) error
    SetExecutionConfig(config ExecutionConfig) error
}

type HumanInputHandler interface {
    RequestInput(ctx context.Context, prompt string, options []string) (string, error)
    IsInteractive() bool
    SetTimeout(timeout time.Duration)
}
```

#### Task 接口

```go
// pkg/task/task.go
type Task interface {
    Execute(ctx context.Context, agent Agent, context map[string]interface{}) (*TaskOutput, error)
    ExecuteAsync(ctx context.Context, agent Agent, context map[string]interface{}) (<-chan TaskResult, error)
    GetDescription() string
    GetExpectedOutput() string
    Validate(output *TaskOutput) error
    SetHumanInputRequired(required bool)
    GetHumanInputRequired() bool
    SetCallback(callback func(ctx context.Context, output *TaskOutput) error)
    SetGuardrails(guardrails []TaskGuardrail)
}

type TaskGuardrail interface {
    Validate(ctx context.Context, output *TaskOutput) (*GuardrailResult, error)
}
```

#### LLM 接口

```go
// pkg/llm/base.go
type LLM interface {
    Call(ctx context.Context, messages []Message, opts ...CallOption) (*Response, error)
    CallWithTimeout(ctx context.Context, messages []Message, timeout time.Duration) (*Response, error)
    SupportsFunctionCalling() bool
    GetModel() string
    SetRPMController(controller RPMController)
}

type RPMController interface {
    AllowRequest(ctx context.Context) error
    GetCurrentRate() int
    Reset()
}
```

#### Tools 接口

```go
// pkg/tools/base.go
type Tool interface {
    Name() string
    Description() string
    Execute(ctx context.Context, args map[string]interface{}) (interface{}, error)
    ExecuteAsync(ctx context.Context, args map[string]interface{}) (<-chan ToolResult, error)
    Schema() ToolSchema
    GetUsageLimit() int
    GetCurrentUsage() int
    ResetUsage()
}

type ToolResult struct {
    Output interface{}
    Error  error
    Usage  ToolUsageInfo
}
```

---

## 4. 详细实施步骤

### 阶段1: 基础架构搭建（第1-2周）

#### 1.1 项目初始化

```bash
# 创建项目
mkdir greensoulai && cd greensoulai
go mod init github.com/ynl/greensoulai

# 创建目录结构
mkdir -p {cmd/greensoulai,internal/{agent,crew,task,flow,tools,memory,llm,knowledge},pkg/{events,security,async,config,logger,errors},examples/{basic,advanced,enterprise},docs/{api,guides,examples},tests/{integration,e2e,benchmarks},scripts,deployments}

# 初始化基础文件
touch {go.mod,go.sum,Makefile,Dockerfile,README.md}
```

#### 1.2 通用基础模块

**配置管理 (pkg/config/)**
```go
type Config struct {
    LogLevel    string `mapstructure:"log_level"`
    EventBus    EventBusConfig `mapstructure:"event_bus"`
    Security    SecurityConfig `mapstructure:"security"`
    Performance PerformanceConfig `mapstructure:"performance"`
}

type PerformanceConfig struct {
    MaxConcurrency int `mapstructure:"max_concurrency"`
    TimeoutDefault time.Duration `mapstructure:"timeout_default"`
}
```

**日志系统 (pkg/logger/)**
```go
type Logger interface {
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    Fatal(msg string, fields ...Field)
}

type Field struct {
    Key   string
    Value interface{}
}
```

**错误定义 (pkg/errors/)**
```go
var (
    ErrAgentNotFound         = errors.New("agent not found")
    ErrTaskTimeout          = errors.New("task execution timeout")
    ErrToolUsageLimitExceeded = errors.New("tool usage limit exceeded")
    ErrHumanInputRequired    = errors.New("human input required")
    ErrInvalidOutputFormat   = errors.New("invalid output format")
)
```

#### 1.3 事件系统实现

**事件总线 (pkg/events/bus.go)**
```go
type EventBus struct {
    handlers map[string][]EventHandler
    mu       sync.RWMutex
    logger   logger.Logger
}

func NewEventBus(logger logger.Logger) *EventBus {
    return &EventBus{
        handlers: make(map[string][]EventHandler),
        logger:   logger,
    }
}

func (eb *EventBus) Emit(ctx context.Context, source interface{}, event Event) error {
    eb.mu.RLock()
    handlers, exists := eb.handlers[event.GetType()]
    eb.mu.RUnlock()
    
    if !exists {
        return nil
    }
    
    for _, handler := range handlers {
        go func(h EventHandler) {
            if err := h(ctx, event); err != nil {
                eb.logger.Error("event handler error", 
                    logger.Field{Key: "error", Value: err},
                    logger.Field{Key: "event_type", Value: event.GetType()},
                )
            }
        }(handler)
    }
    
    return nil
}
```

**事件类型 (pkg/events/types.go)**
```go
type BaseEvent struct {
    Type      string                 `json:"type"`
    Timestamp time.Time             `json:"timestamp"`
    Source    interface{}           `json:"source"`
    Payload   map[string]interface{} `json:"payload"`
}

type AgentExecutionStartedEvent struct {
    BaseEvent
    Agent string `json:"agent"`
    Task  string `json:"task"`
}

type TaskCompletedEvent struct {
    BaseEvent
    Task     string        `json:"task"`
    Agent    string        `json:"agent"`
    Duration time.Duration `json:"duration"`
    Success  bool         `json:"success"`
}
```

#### 1.4 异步基础设施

**异步执行器 (pkg/async/executor.go)**
```go
type AsyncExecutor struct {
    maxWorkers   int
    workerPool   chan chan work
    workQueue    chan work
    quit         chan bool
}

type work struct {
    task   func() (interface{}, error)
    result chan Result
}

func NewAsyncExecutor(maxWorkers int) *AsyncExecutor {
    ae := &AsyncExecutor{
        maxWorkers: maxWorkers,
        workerPool: make(chan chan work, maxWorkers),
        workQueue:  make(chan work, maxWorkers*2),
        quit:       make(chan bool),
    }
    
    ae.start()
    return ae
}

func (ae *AsyncExecutor) ExecuteAsync(ctx context.Context, task func() (interface{}, error)) <-chan Result {
    result := make(chan Result, 1)
    
    select {
    case ae.workQueue <- work{task: task, result: result}:
        return result
    case <-ctx.Done():
        go func() {
            result <- Result{Error: ctx.Err()}
        }()
        return result
    }
}
```

#### 1.5 安全基础模块

**指纹系统 (pkg/security/fingerprint.go)**
```go
type Fingerprint struct {
    UUID      string                 `json:"uuid"`
    CreatedAt time.Time             `json:"created_at"`
    Metadata  map[string]interface{} `json:"metadata"`
}

func NewFingerprint() *Fingerprint {
    return &Fingerprint{
        UUID:      uuid.New().String(),
        CreatedAt: time.Now(),
        Metadata:  make(map[string]interface{}),
    }
}

func GenerateDeterministic(seed string) *Fingerprint {
    namespace := uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479")
    deterministicUUID := uuid.NewSHA1(namespace, []byte(seed))
    
    return &Fingerprint{
        UUID:      deterministicUUID.String(),
        CreatedAt: time.Now(),
        Metadata:  map[string]interface{}{"seed": seed},
    }
}
```

**安全配置 (pkg/security/config.go)**
```go
type SecurityConfig struct {
    Version     string      `json:"version"`
    Fingerprint *Fingerprint `json:"fingerprint"`
    EnableAudit bool        `json:"enable_audit"`
}

func NewSecurityConfig() *SecurityConfig {
    return &SecurityConfig{
        Version:     "1.0.0",
        Fingerprint: NewFingerprint(),
        EnableAudit: false,
    }
}
```

#### 验收标准
- [ ] 项目结构完整，符合 Go 项目最佳实践
- [ ] 事件系统基础功能正常，支持订阅和发布
- [ ] 异步执行器能够处理并发任务
- [ ] 安全指纹生成和验证功能正常
- [ ] 基础模块的单元测试覆盖率达到80%+
- [ ] 日志、配置、错误处理等通用功能完备

---

### 阶段2: LLM和工具系统（第3-4周）

#### 2.1 LLM 抽象层

**基础接口 (pkg/llm/base.go)**
```go
type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type Response struct {
    Content      string                 `json:"content"`
    TokensUsed   int                   `json:"tokens_used"`
    Model        string                `json:"model"`
    FinishReason string                `json:"finish_reason"`
    Metadata     map[string]interface{} `json:"metadata"`
}

type CallOption func(*CallConfig)

type CallConfig struct {
    Temperature      *float64
    MaxTokens        *int
    Tools            []ToolSchema
    ResponseFormat   *ResponseFormat
    Stream           bool
    Timeout          time.Duration
}

func WithTemperature(temp float64) CallOption {
    return func(c *CallConfig) {
        c.Temperature = &temp
    }
}

func WithMaxTokens(tokens int) CallOption {
    return func(c *CallConfig) {
        c.MaxTokens = &tokens
    }
}

func WithTimeout(timeout time.Duration) CallOption {
    return func(c *CallConfig) {
        c.Timeout = timeout
    }
}
```

**RPM控制器 (pkg/llm/rpm_controller.go)**
```go
type RPMController struct {
    maxRPM     int
    window     time.Duration
    requests   []time.Time
    mu         sync.RWMutex
    logger     logger.Logger
}

func NewRPMController(maxRPM int, logger logger.Logger) *RPMController {
    return &RPMController{
        maxRPM:   maxRPM,
        window:   time.Minute,
        requests: make([]time.Time, 0, maxRPM),
        logger:   logger,
    }
}

func (r *RPMController) AllowRequest(ctx context.Context) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    now := time.Now()
    
    // 清理过期请求
    cutoff := now.Add(-r.window)
    validRequests := make([]time.Time, 0, len(r.requests))
    for _, req := range r.requests {
        if req.After(cutoff) {
            validRequests = append(validRequests, req)
        }
    }
    r.requests = validRequests
    
    // 检查是否超出限制
    if len(r.requests) >= r.maxRPM {
        return fmt.Errorf("rate limit exceeded: %d requests in the last minute", len(r.requests))
    }
    
    // 记录新请求
    r.requests = append(r.requests, now)
    return nil
}
```

**OpenAI实现 (pkg/llm/providers/openai.go)**
```go
type OpenAILLM struct {
    client        *openai.Client
    model         string
    rpmController RPMController
    eventBus      events.EventBus
    security      security.SecurityConfig
    logger        logger.Logger
}

func NewOpenAILLM(apiKey, model string, eventBus events.EventBus, logger logger.Logger) *OpenAILLM {
    client := openai.NewClient(apiKey)
    
    return &OpenAILLM{
        client:        client,
        model:         model,
        rpmController: NewRPMController(60, logger), // 默认60 RPM
        eventBus:      eventBus,
        security:      *security.NewSecurityConfig(),
        logger:        logger,
    }
}

func (o *OpenAILLM) Call(ctx context.Context, messages []Message, opts ...CallOption) (*Response, error) {
    // 应用配置选项
    config := &CallConfig{
        Timeout: 30 * time.Second,
    }
    for _, opt := range opts {
        opt(config)
    }
    
    // RPM控制
    if err := o.rpmController.AllowRequest(ctx); err != nil {
        return nil, fmt.Errorf("rate limit: %w", err)
    }
    
    // 发射开始事件
    startEvent := &events.LLMCallStartedEvent{
        BaseEvent: events.BaseEvent{
            Type:      "llm_call_started",
            Timestamp: time.Now(),
            Source:    o,
            Payload:   map[string]interface{}{"model": o.model, "message_count": len(messages)},
        },
        Model: o.model,
    }
    o.eventBus.Emit(ctx, o, startEvent)
    
    // 设置超时
    if config.Timeout > 0 {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, config.Timeout)
        defer cancel()
    }
    
    // 转换消息格式
    openaiMessages := make([]openai.ChatCompletionMessage, len(messages))
    for i, msg := range messages {
        openaiMessages[i] = openai.ChatCompletionMessage{
            Role:    msg.Role,
            Content: msg.Content,
        }
    }
    
    // 构建请求
    request := openai.ChatCompletionRequest{
        Model:    o.model,
        Messages: openaiMessages,
    }
    
    if config.Temperature != nil {
        request.Temperature = *config.Temperature
    }
    if config.MaxTokens != nil {
        request.MaxTokens = *config.MaxTokens
    }
    
    // 执行调用
    start := time.Now()
    resp, err := o.client.CreateChatCompletion(ctx, request)
    duration := time.Since(start)
    
    // 处理响应
    var response *Response
    if err == nil && len(resp.Choices) > 0 {
        response = &Response{
            Content:      resp.Choices[0].Message.Content,
            TokensUsed:   resp.Usage.TotalTokens,
            Model:        resp.Model,
            FinishReason: string(resp.Choices[0].FinishReason),
            Metadata: map[string]interface{}{
                "prompt_tokens":     resp.Usage.PromptTokens,
                "completion_tokens": resp.Usage.CompletionTokens,
                "duration_ms":       duration.Milliseconds(),
            },
        }
    }
    
    // 发射完成事件
    completedEvent := &events.LLMCallCompletedEvent{
        BaseEvent: events.BaseEvent{
            Type:      "llm_call_completed",
            Timestamp: time.Now(),
            Source:    o,
            Payload: map[string]interface{}{
                "model":       o.model,
                "duration_ms": duration.Milliseconds(),
                "success":     err == nil,
            },
        },
        Model:    o.model,
        Duration: duration,
        Success:  err == nil,
    }
    o.eventBus.Emit(ctx, o, completedEvent)
    
    if err != nil {
        o.logger.Error("OpenAI API call failed", 
            logger.Field{Key: "error", Value: err},
            logger.Field{Key: "model", Value: o.model},
            logger.Field{Key: "duration", Value: duration},
        )
        return nil, fmt.Errorf("openai call failed: %w", err)
    }
    
    return response, nil
}

func (o *OpenAILLM) CallWithTimeout(ctx context.Context, messages []Message, timeout time.Duration) (*Response, error) {
    return o.Call(ctx, messages, WithTimeout(timeout))
}

func (o *OpenAILLM) SupportsFunctionCalling() bool {
    return true
}

func (o *OpenAILLM) GetModel() string {
    return o.model
}

func (o *OpenAILLM) SetRPMController(controller RPMController) {
    o.rpmController = controller
}
```

#### 2.2 工具系统

**工具基础 (pkg/tools/base.go)**
```go
type ToolSchema struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Parameters  map[string]interface{} `json:"parameters"`
}

type ToolUsageInfo struct {
    Count     int           `json:"count"`
    Duration  time.Duration `json:"duration"`
    Timestamp time.Time     `json:"timestamp"`
}

type BaseTool struct {
    name           string
    description    string
    schema         ToolSchema
    handler        func(ctx context.Context, args map[string]interface{}) (interface{}, error)
    maxUsage       int
    currentUsage   int
    eventBus       events.EventBus
    securityConfig security.SecurityConfig
    logger         logger.Logger
    mu             sync.RWMutex
}

func NewBaseTool(
    name, description string,
    handler func(ctx context.Context, args map[string]interface{}) (interface{}, error),
    eventBus events.EventBus,
    logger logger.Logger,
) *BaseTool {
    return &BaseTool{
        name:           name,
        description:    description,
        handler:        handler,
        maxUsage:       -1, // 无限制
        currentUsage:   0,
        eventBus:       eventBus,
        securityConfig: *security.NewSecurityConfig(),
        logger:         logger,
    }
}

func (bt *BaseTool) Name() string {
    return bt.name
}

func (bt *BaseTool) Description() string {
    return bt.description
}

func (bt *BaseTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    bt.mu.Lock()
    
    // 检查使用限制
    if bt.maxUsage >= 0 && bt.currentUsage >= bt.maxUsage {
        bt.mu.Unlock()
        return nil, fmt.Errorf("tool usage limit exceeded: %d/%d", bt.currentUsage, bt.maxUsage)
    }
    
    bt.currentUsage++
    usageCount := bt.currentUsage
    bt.mu.Unlock()
    
    // 发射开始事件
    startEvent := &events.ToolUsageStartedEvent{
        BaseEvent: events.BaseEvent{
            Type:      "tool_usage_started",
            Timestamp: time.Now(),
            Source:    bt,
            Payload: map[string]interface{}{
                "tool_name":     bt.name,
                "usage_count":   usageCount,
                "args":          args,
            },
        },
        ToolName: bt.name,
        Args:     args,
    }
    bt.eventBus.Emit(ctx, bt, startEvent)
    
    // 执行工具
    start := time.Now()
    result, err := bt.handler(ctx, args)
    duration := time.Since(start)
    
    // 发射完成事件
    finishedEvent := &events.ToolUsageFinishedEvent{
        BaseEvent: events.BaseEvent{
            Type:      "tool_usage_finished",
            Timestamp: time.Now(),
            Source:    bt,
            Payload: map[string]interface{}{
                "tool_name":     bt.name,
                "duration_ms":   duration.Milliseconds(),
                "success":       err == nil,
                "usage_count":   usageCount,
            },
        },
        ToolName: bt.name,
        Duration: duration,
        Success:  err == nil,
    }
    bt.eventBus.Emit(ctx, bt, finishedEvent)
    
    if err != nil {
        bt.logger.Error("tool execution failed",
            logger.Field{Key: "tool", Value: bt.name},
            logger.Field{Key: "error", Value: err},
            logger.Field{Key: "duration", Value: duration},
        )
    }
    
    return result, err
}

func (bt *BaseTool) ExecuteAsync(ctx context.Context, args map[string]interface{}) (<-chan ToolResult, error) {
    resultChan := make(chan ToolResult, 1)
    
    go func() {
        defer close(resultChan)
        
        output, err := bt.Execute(ctx, args)
        
        resultChan <- ToolResult{
            Output: output,
            Error:  err,
            Usage: ToolUsageInfo{
                Count:     bt.GetCurrentUsage(),
                Timestamp: time.Now(),
            },
        }
    }()
    
    return resultChan, nil
}

func (bt *BaseTool) Schema() ToolSchema {
    return bt.schema
}

func (bt *BaseTool) GetUsageLimit() int {
    bt.mu.RLock()
    defer bt.mu.RUnlock()
    return bt.maxUsage
}

func (bt *BaseTool) GetCurrentUsage() int {
    bt.mu.RLock()
    defer bt.mu.RUnlock()
    return bt.currentUsage
}

func (bt *BaseTool) ResetUsage() {
    bt.mu.Lock()
    defer bt.mu.Unlock()
    bt.currentUsage = 0
}

func (bt *BaseTool) SetUsageLimit(limit int) {
    bt.mu.Lock()
    defer bt.mu.Unlock()
    bt.maxUsage = limit
}
```

**结构化工具 (pkg/tools/structured.go)**
```go
type StructuredTool struct {
    *BaseTool
    argsSchema reflect.Type
    validator  *validator.Validate
}

func NewStructuredTool(
    name, description string,
    argsSchema interface{},
    handler func(ctx context.Context, args interface{}) (interface{}, error),
    eventBus events.EventBus,
    logger logger.Logger,
) (*StructuredTool, error) {
    schemaType := reflect.TypeOf(argsSchema)
    if schemaType.Kind() == reflect.Ptr {
        schemaType = schemaType.Elem()
    }
    
    // 将类型化的handler转换为通用handler
    genericHandler := func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
        // 创建结构体实例
        argsValue := reflect.New(schemaType).Interface()
        
        // 转换参数
        if err := mapstructure.Decode(args, argsValue); err != nil {
            return nil, fmt.Errorf("failed to decode args: %w", err)
        }
        
        // 验证参数
        if err := validator.New().Struct(argsValue); err != nil {
            return nil, fmt.Errorf("validation failed: %w", err)
        }
        
        return handler(ctx, argsValue)
    }
    
    baseTool := NewBaseTool(name, description, genericHandler, eventBus, logger)
    
    return &StructuredTool{
        BaseTool:   baseTool,
        argsSchema: schemaType,
        validator:  validator.New(),
    }, nil
}
```

#### 验收标准
- [ ] OpenAI LLM 调用成功，支持所有常用参数
- [ ] RPM 控制器有效限制请求频率
- [ ] 工具系统支持同步和异步执行
- [ ] 工具使用限制和监控功能正常
- [ ] 事件系统完整集成（LLM和工具事件）
- [ ] 错误处理和重试机制完善
- [ ] 结构化工具参数验证正确
- [ ] 性能测试通过（并发调用、内存使用）

---

### 阶段3: Agent和Task系统（第5-6周）

#### 3.1 人工输入处理

**人工输入接口 (pkg/agent/human_input.go)**
```go
type HumanInputHandler interface {
    RequestInput(ctx context.Context, prompt string, options []string) (string, error)
    IsInteractive() bool
    SetTimeout(timeout time.Duration)
}

type ConsoleInputHandler struct {
    timeout time.Duration
    logger  logger.Logger
}

func NewConsoleInputHandler(logger logger.Logger) *ConsoleInputHandler {
    return &ConsoleInputHandler{
        timeout: 5 * time.Minute,
        logger:  logger,
    }
}

func (c *ConsoleInputHandler) RequestInput(ctx context.Context, prompt string, options []string) (string, error) {
    // 创建超时上下文
    ctx, cancel := context.WithTimeout(ctx, c.timeout)
    defer cancel()
    
    // 显示提示
    fmt.Printf("\n🤖 %s\n", prompt)
    if len(options) > 0 {
        fmt.Println("Options:")
        for i, option := range options {
            fmt.Printf("  %d) %s\n", i+1, option)
        }
    }
    fmt.Print("Your input: ")
    
    // 在goroutine中等待用户输入
    inputChan := make(chan string, 1)
    errorChan := make(chan error, 1)
    
    go func() {
        reader := bufio.NewReader(os.Stdin)
        input, err := reader.ReadString('\n')
        if err != nil {
            errorChan <- err
            return
        }
        inputChan <- strings.TrimSpace(input)
    }()
    
    // 等待输入或超时
    select {
    case input := <-inputChan:
        c.logger.Info("received human input", logger.Field{Key: "input_length", Value: len(input)})
        return input, nil
    case err := <-errorChan:
        return "", fmt.Errorf("failed to read input: %w", err)
    case <-ctx.Done():
        return "", fmt.Errorf("input timeout: %w", ctx.Err())
    }
}

func (c *ConsoleInputHandler) IsInteractive() bool {
    return true
}

func (c *ConsoleInputHandler) SetTimeout(timeout time.Duration) {
    c.timeout = timeout
}
```

#### 3.2 Agent实现

**Agent配置 (pkg/agent/agent.go)**
```go
type ExecutionConfig struct {
    MaxIterations   int           `json:"max_iterations"`
    MaxRPM         int           `json:"max_rpm"`
    Timeout        time.Duration `json:"timeout"`
    AllowDelegation bool          `json:"allow_delegation"`
    VerboseLogging  bool          `json:"verbose_logging"`
    HumanInput     bool          `json:"human_input"`
}

type BaseAgent struct {
    role               string
    goal               string
    backstory          string
    llm                llm.LLM
    tools              []tools.Tool
    memory             memory.Memory
    knowledgeSource    knowledge.KnowledgeSource
    securityConfig     security.SecurityConfig
    executionConfig    ExecutionConfig
    eventBus           events.EventBus
    humanInputHandler  HumanInputHandler
    rpmController      *llm.RPMController
    logger             logger.Logger
    
    // 执行统计
    executionCount     int
    lastExecutionTime  time.Time
    totalExecutionTime time.Duration
    mu                 sync.RWMutex
}

func NewBaseAgent(
    role, goal, backstory string,
    llmProvider llm.LLM,
    eventBus events.EventBus,
    logger logger.Logger,
) *BaseAgent {
    return &BaseAgent{
        role:               role,
        goal:               goal,
        backstory:          backstory,
        llm:                llmProvider,
        tools:              make([]tools.Tool, 0),
        securityConfig:     *security.NewSecurityConfig(),
        executionConfig:    ExecutionConfig{
            MaxIterations:   50,
            Timeout:         30 * time.Minute,
            AllowDelegation: false,
            VerboseLogging:  false,
            HumanInput:     false,
        },
        eventBus:           eventBus,
        logger:             logger,
        executionCount:     0,
    }
}

func (a *BaseAgent) Execute(ctx context.Context, task Task) (*TaskOutput, error) {
    // 更新执行统计
    a.mu.Lock()
    a.executionCount++
    executionID := a.executionCount
    a.lastExecutionTime = time.Now()
    a.mu.Unlock()
    
    // 发射开始事件
    startEvent := &events.AgentExecutionStartedEvent{
        BaseEvent: events.BaseEvent{
            Type:      "agent_execution_started",
            Timestamp: time.Now(),
            Source:    a,
            Payload: map[string]interface{}{
                "agent":        a.role,
                "task":         task.GetDescription(),
                "execution_id": executionID,
            },
        },
        Agent: a.role,
        Task:  task.GetDescription(),
    }
    a.eventBus.Emit(ctx, a, startEvent)
    
    // 检查人工输入需求
    if task.GetHumanInputRequired() {
        if a.humanInputHandler == nil {
            err := fmt.Errorf("human input required but no handler configured")
            a.logger.Error("human input handler missing", logger.Field{Key: "task", Value: task.GetDescription()})
            return nil, err
        }
        
        input, err := a.humanInputHandler.RequestInput(ctx, 
            fmt.Sprintf("Task requires your input: %s", task.GetDescription()), 
            nil)
        if err != nil {
            a.logger.Error("human input failed", 
                logger.Field{Key: "error", Value: err},
                logger.Field{Key: "task", Value: task.GetDescription()},
            )
            return nil, fmt.Errorf("human input failed: %w", err)
        }
        
        task.SetHumanInput(input)
        a.logger.Info("received human input for task", 
            logger.Field{Key: "task", Value: task.GetDescription()},
            logger.Field{Key: "input_length", Value: len(input)},
        )
    }
    
    // 执行任务
    start := time.Now()
    output, err := a.executeTask(ctx, task)
    duration := time.Since(start)
    
    // 更新执行时间统计
    a.mu.Lock()
    a.totalExecutionTime += duration
    a.mu.Unlock()
    
    // 发射完成事件
    completedEvent := &events.AgentExecutionCompletedEvent{
        BaseEvent: events.BaseEvent{
            Type:      "agent_execution_completed",
            Timestamp: time.Now(),
            Source:    a,
            Payload: map[string]interface{}{
                "agent":        a.role,
                "task":         task.GetDescription(),
                "execution_id": executionID,
                "duration_ms":  duration.Milliseconds(),
                "success":      err == nil,
            },
        },
        Agent:    a.role,
        Task:     task.GetDescription(),
        Duration: duration,
        Success:  err == nil,
    }
    a.eventBus.Emit(ctx, a, completedEvent)
    
    if err != nil {
        a.logger.Error("agent execution failed",
            logger.Field{Key: "agent", Value: a.role},
            logger.Field{Key: "task", Value: task.GetDescription()},
            logger.Field{Key: "error", Value: err},
            logger.Field{Key: "duration", Value: duration},
        )
    } else {
        a.logger.Info("agent execution completed",
            logger.Field{Key: "agent", Value: a.role},
            logger.Field{Key: "task", Value: task.GetDescription()},
            logger.Field{Key: "duration", Value: duration},
        )
    }
    
    return output, err
}

func (a *BaseAgent) ExecuteWithTimeout(ctx context.Context, task Task, timeout time.Duration) (*TaskOutput, error) {
    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()
    
    resultChan := make(chan TaskResult, 1)
    
    go func() {
        output, err := a.Execute(ctx, task)
        resultChan <- TaskResult{Output: output, Error: err}
    }()
    
    select {
    case result := <-resultChan:
        return result.Output, result.Error
    case <-ctx.Done():
        return nil, fmt.Errorf("agent execution timeout after %v: %w", timeout, ctx.Err())
    }
}

func (a *BaseAgent) ExecuteAsync(ctx context.Context, task Task) (<-chan TaskResult, error) {
    resultChan := make(chan TaskResult, 1)
    
    go func() {
        defer close(resultChan)
        output, err := a.Execute(ctx, task)
        resultChan <- TaskResult{Output: output, Error: err}
    }()
    
    return resultChan, nil
}

func (a *BaseAgent) executeTask(ctx context.Context, task Task) (*TaskOutput, error) {
    // 构建任务提示
    prompt := a.buildTaskPrompt(task)
    
    // 准备消息
    messages := []llm.Message{
        {Role: "system", Content: a.buildSystemPrompt()},
        {Role: "user", Content: prompt},
    }
    
    // 调用LLM
    response, err := a.llm.Call(ctx, messages)
    if err != nil {
        return nil, fmt.Errorf("llm call failed: %w", err)
    }
    
    // 处理响应并构建输出
    return a.processLLMResponse(task, response)
}

func (a *BaseAgent) buildSystemPrompt() string {
    return fmt.Sprintf(`You are %s.

Your goal: %s

Your backstory: %s

You are working with a team of other agents to complete complex tasks. Always provide detailed, accurate responses based on your role and expertise.`, 
        a.role, a.goal, a.backstory)
}

func (a *BaseAgent) buildTaskPrompt(task Task) string {
    prompt := fmt.Sprintf("Task: %s\n\nExpected Output: %s\n", 
        task.GetDescription(), 
        task.GetExpectedOutput())
    
    // 添加人工输入（如果有）
    if task.GetHumanInputRequired() && task.(*BaseTask).HumanInput != "" {
        prompt += fmt.Sprintf("\nHuman Input: %s\n", task.(*BaseTask).HumanInput)
    }
    
    // 添加可用工具信息
    if len(a.tools) > 0 {
        prompt += "\nAvailable Tools:\n"
        for _, tool := range a.tools {
            prompt += fmt.Sprintf("- %s: %s\n", tool.Name(), tool.Description())
        }
    }
    
    return prompt
}

func (a *BaseAgent) processLLMResponse(task Task, response *llm.Response) (*TaskOutput, error) {
    output := &TaskOutput{
        Raw:           response.Content,
        Agent:         a.role,
        Description:   task.GetDescription(),
        ExpectedOutput: task.GetExpectedOutput(),
        CreatedAt:     time.Now(),
        Metadata: map[string]interface{}{
            "model":       response.Model,
            "tokens_used": response.TokensUsed,
            "finish_reason": response.FinishReason,
        },
    }
    
    // 尝试解析JSON输出（如果适用）
    if strings.Contains(response.Content, "{") {
        var jsonData map[string]interface{}
        if err := json.Unmarshal([]byte(response.Content), &jsonData); err == nil {
            output.JSON = jsonData
            output.OutputFormat = OutputFormat.JSON
        }
    }
    
    // 生成摘要
    contentWords := strings.Fields(output.Raw)
    if len(contentWords) > 10 {
        output.Summary = strings.Join(contentWords[:10], " ") + "..."
    } else {
        output.Summary = output.Raw
    }
    
    return output, nil
}

// Agent接口实现
func (a *BaseAgent) GetRole() string {
    return a.role
}

func (a *BaseAgent) GetGoal() string {
    return a.goal
}

func (a *BaseAgent) AddTool(tool tools.Tool) error {
    a.tools = append(a.tools, tool)
    a.logger.Info("tool added to agent",
        logger.Field{Key: "agent", Value: a.role},
        logger.Field{Key: "tool", Value: tool.Name()},
    )
    return nil
}

func (a *BaseAgent) SetMemory(mem memory.Memory) error {
    a.memory = mem
    return nil
}

func (a *BaseAgent) SetHumanInputHandler(handler HumanInputHandler) error {
    a.humanInputHandler = handler
    return nil
}

func (a *BaseAgent) SetExecutionConfig(config ExecutionConfig) error {
    a.executionConfig = config
    return nil
}
```

#### 3.3 Task系统实现

**任务基础 (pkg/task/task.go)**
```go
type OutputFormat int

const (
    OutputFormatRAW OutputFormat = iota
    OutputFormatJSON
    OutputFormatPydantic
)

type TaskOutput struct {
    Raw             string                 `json:"raw"`
    JSON            map[string]interface{} `json:"json,omitempty"`
    Pydantic        interface{}           `json:"pydantic,omitempty"`
    Agent           string                 `json:"agent"`
    Description     string                 `json:"description"`
    Summary         string                 `json:"summary"`
    Name            string                 `json:"name,omitempty"`
    ExpectedOutput  string                 `json:"expected_output"`
    OutputFormat    OutputFormat          `json:"output_format"`
    ExecutionTime   time.Duration         `json:"execution_time"`
    CreatedAt       time.Time             `json:"created_at"`
    Metadata        map[string]interface{} `json:"metadata"`
    
    // 验证和护栏
    IsValid         bool             `json:"is_valid"`
    ValidationError error            `json:"validation_error,omitempty"`
    GuardrailResult *GuardrailResult `json:"guardrail_result,omitempty"`
}

type TaskResult struct {
    Output *TaskOutput
    Error  error
}

type GuardrailResult struct {
    Valid    bool                   `json:"valid"`
    Message  string                 `json:"message"`
    Score    float64               `json:"score"`
    Details  map[string]interface{} `json:"details"`
}

type BaseTask struct {
    description        string
    expectedOutput     string
    agent             Agent
    tools             []tools.Tool
    context           []Task
    
    // 新增功能
    humanInputRequired bool
    humanInput         string
    callback           func(ctx context.Context, output *TaskOutput) error
    outputFile         string
    asyncExecution     bool
    securityConfig     security.SecurityConfig
    eventBus           events.EventBus
    guardrails         []TaskGuardrail
    logger             logger.Logger
    
    // 执行统计
    startTime          time.Time
    endTime            time.Time
    executionDuration  time.Duration
    mu                 sync.RWMutex
}

func NewBaseTask(description, expectedOutput string, eventBus events.EventBus, logger logger.Logger) *BaseTask {
    return &BaseTask{
        description:        description,
        expectedOutput:     expectedOutput,
        tools:             make([]tools.Tool, 0),
        context:           make([]Task, 0),
        humanInputRequired: false,
        asyncExecution:     false,
        securityConfig:     *security.NewSecurityConfig(),
        eventBus:           eventBus,
        guardrails:         make([]TaskGuardrail, 0),
        logger:             logger,
    }
}

func (bt *BaseTask) Execute(ctx context.Context, agent Agent, taskContext map[string]interface{}) (*TaskOutput, error) {
    bt.mu.Lock()
    bt.startTime = time.Now()
    bt.mu.Unlock()
    
    // 发射任务开始事件
    startEvent := &events.TaskStartedEvent{
        BaseEvent: events.BaseEvent{
            Type:      "task_started",
            Timestamp: time.Now(),
            Source:    bt,
            Payload: map[string]interface{}{
                "task":        bt.description,
                "agent":       agent.GetRole(),
                "has_context": len(taskContext) > 0,
            },
        },
        Task:  bt.description,
        Agent: agent.GetRole(),
    }
    bt.eventBus.Emit(ctx, bt, startEvent)
    
    // 执行任务
    output, err := agent.Execute(ctx, bt)
    
    bt.mu.Lock()
    bt.endTime = time.Now()
    bt.executionDuration = bt.endTime.Sub(bt.startTime)
    bt.mu.Unlock()
    
    if output != nil {
        output.ExecutionTime = bt.executionDuration
    }
    
    // 执行护栏验证
    if err == nil && len(bt.guardrails) > 0 {
        bt.logger.Info("executing guardrails", 
            logger.Field{Key: "task", Value: bt.description},
            logger.Field{Key: "guardrail_count", Value: len(bt.guardrails)},
        )
        
        for i, guardrail := range bt.guardrails {
            result, guardErr := guardrail.Validate(ctx, output)
            if guardErr != nil {
                bt.logger.Error("guardrail validation error",
                    logger.Field{Key: "task", Value: bt.description},
                    logger.Field{Key: "guardrail_index", Value: i},
                    logger.Field{Key: "error", Value: guardErr},
                )
                output.IsValid = false
                output.ValidationError = guardErr
                output.GuardrailResult = result
                break
            }
            
            if !result.Valid {
                bt.logger.Warn("guardrail validation failed",
                    logger.Field{Key: "task", Value: bt.description},
                    logger.Field{Key: "guardrail_index", Value: i},
                    logger.Field{Key: "message", Value: result.Message},
                )
                output.IsValid = false
                output.GuardrailResult = result
                break
            }
        }
        
        if output.GuardrailResult == nil || output.GuardrailResult.Valid {
            output.IsValid = true
        }
    } else {
        output.IsValid = err == nil
    }
    
    // 执行回调
    if bt.callback != nil && err == nil {
        if callbackErr := bt.callback(ctx, output); callbackErr != nil {
            bt.logger.Error("task callback failed",
                logger.Field{Key: "task", Value: bt.description},
                logger.Field{Key: "error", Value: callbackErr},
            )
        }
    }
    
    // 写入输出文件
    if bt.outputFile != "" && err == nil {
        if writeErr := bt.writeOutputToFile(output); writeErr != nil {
            bt.logger.Error("failed to write output file",
                logger.Field{Key: "task", Value: bt.description},
                logger.Field{Key: "file", Value: bt.outputFile},
                logger.Field{Key: "error", Value: writeErr},
            )
        }
    }
    
    // 发射任务完成事件
    completedEvent := &events.TaskCompletedEvent{
        BaseEvent: events.BaseEvent{
            Type:      "task_completed",
            Timestamp: time.Now(),
            Source:    bt,
            Payload: map[string]interface{}{
                "task":        bt.description,
                "agent":       agent.GetRole(),
                "duration_ms": bt.executionDuration.Milliseconds(),
                "success":     err == nil,
                "valid":       output != nil && output.IsValid,
            },
        },
        Task:     bt.description,
        Agent:    agent.GetRole(),
        Duration: bt.executionDuration,
        Success:  err == nil,
    }
    bt.eventBus.Emit(ctx, bt, completedEvent)
    
    if err != nil {
        bt.logger.Error("task execution failed",
            logger.Field{Key: "task", Value: bt.description},
            logger.Field{Key: "agent", Value: agent.GetRole()},
            logger.Field{Key: "error", Value: err},
            logger.Field{Key: "duration", Value: bt.executionDuration},
        )
    } else {
        bt.logger.Info("task execution completed",
            logger.Field{Key: "task", Value: bt.description},
            logger.Field{Key: "agent", Value: agent.GetRole()},
            logger.Field{Key: "duration", Value: bt.executionDuration},
            logger.Field{Key: "valid", Value: output.IsValid},
        )
    }
    
    return output, err
}

func (bt *BaseTask) ExecuteAsync(ctx context.Context, agent Agent, taskContext map[string]interface{}) (<-chan TaskResult, error) {
    resultChan := make(chan TaskResult, 1)
    
    go func() {
        defer close(resultChan)
        output, err := bt.Execute(ctx, agent, taskContext)
        resultChan <- TaskResult{Output: output, Error: err}
    }()
    
    return resultChan, nil
}

func (bt *BaseTask) writeOutputToFile(output *TaskOutput) error {
    var content []byte
    var err error
    
    switch output.OutputFormat {
    case OutputFormatJSON:
        if output.JSON != nil {
            content, err = json.MarshalIndent(output.JSON, "", "  ")
        } else {
            content = []byte(output.Raw)
        }
    default:
        content = []byte(output.Raw)
    }
    
    if err != nil {
        return fmt.Errorf("failed to marshal output: %w", err)
    }
    
    // 确保目录存在
    dir := filepath.Dir(bt.outputFile)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("failed to create output directory: %w", err)
    }
    
    return os.WriteFile(bt.outputFile, content, 0644)
}

// Task接口实现
func (bt *BaseTask) GetDescription() string {
    return bt.description
}

func (bt *BaseTask) GetExpectedOutput() string {
    return bt.expectedOutput
}

func (bt *BaseTask) Validate(output *TaskOutput) error {
    if output == nil {
        return fmt.Errorf("output is nil")
    }
    
    if output.Raw == "" {
        return fmt.Errorf("output is empty")
    }
    
    return nil
}

func (bt *BaseTask) SetHumanInputRequired(required bool) {
    bt.mu.Lock()
    defer bt.mu.Unlock()
    bt.humanInputRequired = required
}

func (bt *BaseTask) GetHumanInputRequired() bool {
    bt.mu.RLock()
    defer bt.mu.RUnlock()
    return bt.humanInputRequired
}

func (bt *BaseTask) SetHumanInput(input string) {
    bt.mu.Lock()
    defer bt.mu.Unlock()
    bt.humanInput = input
}

func (bt *BaseTask) SetCallback(callback func(ctx context.Context, output *TaskOutput) error) {
    bt.mu.Lock()
    defer bt.mu.Unlock()
    bt.callback = callback
}

func (bt *BaseTask) SetOutputFile(filename string) error {
    bt.mu.Lock()
    defer bt.mu.Unlock()
    bt.outputFile = filename
    return nil
}

func (bt *BaseTask) SetAsyncExecution(async bool) {
    bt.mu.Lock()
    defer bt.mu.Unlock()
    bt.asyncExecution = async
}

func (bt *BaseTask) GetExecutionDuration() time.Duration {
    bt.mu.RLock()
    defer bt.mu.RUnlock()
    return bt.executionDuration
}

func (bt *BaseTask) SetGuardrails(guardrails []TaskGuardrail) {
    bt.mu.Lock()
    defer bt.mu.Unlock()
    bt.guardrails = guardrails
}
```

#### 3.4 护栏系统

**护栏接口 (pkg/task/guardrail.go)**
```go
type TaskGuardrail interface {
    Validate(ctx context.Context, output *TaskOutput) (*GuardrailResult, error)
    GetDescription() string
}

type LLMGuardrail struct {
    description string
    llm         llm.LLM
    logger      logger.Logger
}

func NewLLMGuardrail(description string, llmProvider llm.LLM, logger logger.Logger) *LLMGuardrail {
    return &LLMGuardrail{
        description: description,
        llm:         llmProvider,
        logger:      logger,
    }
}

func (lg *LLMGuardrail) Validate(ctx context.Context, output *TaskOutput) (*GuardrailResult, error) {
    prompt := fmt.Sprintf(`Ensure the following task result complies with the given guardrail.

Task result:
%s

Guardrail:
%s

Your task:
- Confirm if the Task result complies with the guardrail.
- If not, provide clear feedback explaining what is wrong.
- Focus only on identifying issues — do not propose corrections.
- If the Task result complies with the guardrail, say that it is valid.

Respond with a JSON object with the following structure:
{
    "valid": boolean,
    "message": "explanation of validation result",
    "score": number between 0 and 1 indicating compliance level,
    "details": {}
}`, output.Raw, lg.description)
    
    messages := []llm.Message{
        {Role: "system", Content: "You are a validation expert. Always respond with valid JSON."},
        {Role: "user", Content: prompt},
    }
    
    response, err := lg.llm.Call(ctx, messages)
    if err != nil {
        return nil, fmt.Errorf("llm validation call failed: %w", err)
    }
    
    // 解析JSON响应
    var result GuardrailResult
    if err := json.Unmarshal([]byte(response.Content), &result); err != nil {
        lg.logger.Error("failed to parse guardrail response",
            logger.Field{Key: "response", Value: response.Content},
            logger.Field{Key: "error", Value: err},
        )
        
        // 回退到简单的文本分析
        result = GuardrailResult{
            Valid:   !strings.Contains(strings.ToLower(response.Content), "not valid") && 
                    !strings.Contains(strings.ToLower(response.Content), "invalid") &&
                    !strings.Contains(strings.ToLower(response.Content), "fails"),
            Message: response.Content,
            Score:   0.5, // 默认分数
            Details: map[string]interface{}{
                "raw_response": response.Content,
                "parsing_error": err.Error(),
            },
        }
    }
    
    lg.logger.Info("guardrail validation completed",
        logger.Field{Key: "description", Value: lg.description},
        logger.Field{Key: "valid", Value: result.Valid},
        logger.Field{Key: "score", Value: result.Score},
    )
    
    return &result, nil
}

func (lg *LLMGuardrail) GetDescription() string {
    return lg.description
}

// 长度护栏
type LengthGuardrail struct {
    minLength int
    maxLength int
    logger    logger.Logger
}

func NewLengthGuardrail(minLength, maxLength int, logger logger.Logger) *LengthGuardrail {
    return &LengthGuardrail{
        minLength: minLength,
        maxLength: maxLength,
        logger:    logger,
    }
}

func (lg *LengthGuardrail) Validate(ctx context.Context, output *TaskOutput) (*GuardrailResult, error) {
    length := len(output.Raw)
    
    valid := true
    message := "Output length is within acceptable range"
    score := 1.0
    
    if lg.minLength > 0 && length < lg.minLength {
        valid = false
        message = fmt.Sprintf("Output too short: %d characters (minimum: %d)", length, lg.minLength)
        score = float64(length) / float64(lg.minLength)
    } else if lg.maxLength > 0 && length > lg.maxLength {
        valid = false
        message = fmt.Sprintf("Output too long: %d characters (maximum: %d)", length, lg.maxLength)
        score = float64(lg.maxLength) / float64(length)
    }
    
    result := &GuardrailResult{
        Valid:   valid,
        Message: message,
        Score:   score,
        Details: map[string]interface{}{
            "actual_length": length,
            "min_length":    lg.minLength,
            "max_length":    lg.maxLength,
        },
    }
    
    lg.logger.Info("length guardrail validation",
        logger.Field{Key: "length", Value: length},
        logger.Field{Key: "valid", Value: valid},
        logger.Field{Key: "score", Value: score},
    )
    
    return result, nil
}

func (lg *LengthGuardrail) GetDescription() string {
    return fmt.Sprintf("Output length must be between %d and %d characters", lg.minLength, lg.maxLength)
}
```

#### 验收标准
- [ ] Agent 可执行简单任务并返回正确格式输出
- [ ] 异步任务执行支持，能够处理并发请求
- [ ] 人工干预机制工作正常，支持控制台和自定义输入
- [ ] 超时控制有效，能够在指定时间内终止执行
- [ ] 事件系统完整集成，所有执行阶段都有对应事件
- [ ] 支持工具调用（基于之前实现的工具系统）
- [ ] 任务输出支持多种格式（RAW、JSON、结构化）
- [ ] 条件任务执行逻辑正确
- [ ] 任务护栏验证功能正常，包括LLM和规则护栏
- [ ] 回调机制正常，支持任务完成后的自定义处理
- [ ] 执行统计和监控数据完整准确
- [ ] 输出文件写入功能正常
- [ ] 错误处理完善，所有异常都有合适的处理

---

## 5. 质量保证体系

### 5.1 测试策略

#### 单元测试
```go
// 示例: Agent测试 (pkg/agent/agent_test.go)
func TestBaseAgent_Execute(t *testing.T) {
    // 设置测试依赖
    logger := logger.NewTestLogger()
    eventBus := events.NewEventBus(logger)
    mockLLM := &MockLLM{}
    
    agent := NewBaseAgent("test-agent", "test goal", "test backstory", mockLLM, eventBus, logger)
    
    // 创建测试任务
    task := NewBaseTask("test task", "expected output", eventBus, logger)
    
    // 执行测试
    ctx := context.Background()
    output, err := agent.Execute(ctx, task)
    
    // 验证结果
    assert.NoError(t, err)
    assert.NotNil(t, output)
    assert.Equal(t, "test-agent", output.Agent)
    assert.True(t, mockLLM.Called)
}

func TestAgent_ExecuteWithTimeout(t *testing.T) {
    // 测试超时场景
    logger := logger.NewTestLogger()
    eventBus := events.NewEventBus(logger)
    slowLLM := &SlowMockLLM{delay: 2 * time.Second}
    
    agent := NewBaseAgent("test-agent", "test goal", "test backstory", slowLLM, eventBus, logger)
    task := NewBaseTask("slow task", "expected output", eventBus, logger)
    
    ctx := context.Background()
    _, err := agent.ExecuteWithTimeout(ctx, task, 1*time.Second)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "timeout")
}
```

#### 集成测试
```go
// 示例: 端到端集成测试 (tests/integration/agent_llm_test.go)
func TestAgent_LLM_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // 使用真实的OpenAI API (需要API密钥)
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        t.Skip("OPENAI_API_KEY not set")
    }
    
    logger := logger.NewConsoleLogger()
    eventBus := events.NewEventBus(logger)
    
    openaiLLM := NewOpenAILLM(apiKey, "gpt-3.5-turbo", eventBus, logger)
    agent := NewBaseAgent("Software Engineer", 
        "Write clean, efficient code", 
        "You are an experienced software engineer", 
        openaiLLM, eventBus, logger)
    
    task := NewBaseTask(
        "Write a simple Hello World function in Go", 
        "A complete Go function that prints 'Hello, World!'", 
        eventBus, logger)
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    output, err := agent.Execute(ctx, task)
    
    assert.NoError(t, err)
    assert.NotNil(t, output)
    assert.Contains(t, strings.ToLower(output.Raw), "hello")
    assert.Contains(t, strings.ToLower(output.Raw), "world")
}
```

#### 性能测试
```go
// 示例: 并发性能测试 (tests/benchmarks/agent_benchmark_test.go)
func BenchmarkAgent_ConcurrentExecution(b *testing.B) {
    logger := logger.NewTestLogger()
    eventBus := events.NewEventBus(logger)
    fastLLM := &FastMockLLM{}
    
    agent := NewBaseAgent("test-agent", "test goal", "test backstory", fastLLM, eventBus, logger)
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            task := NewBaseTask("benchmark task", "expected output", eventBus, logger)
            ctx := context.Background()
            
            _, err := agent.Execute(ctx, task)
            if err != nil {
                b.Fatalf("execution failed: %v", err)
            }
        }
    })
}

func BenchmarkEventBus_EmitPerformance(b *testing.B) {
    logger := logger.NewTestLogger()
    eventBus := events.NewEventBus(logger)
    
    // 注册多个处理器
    for i := 0; i < 10; i++ {
        eventBus.Subscribe("test_event", func(ctx context.Context, event events.Event) error {
            // 模拟处理时间
            time.Sleep(1 * time.Microsecond)
            return nil
        })
    }
    
    event := &events.BaseEvent{
        Type:      "test_event",
        Timestamp: time.Now(),
        Payload:   map[string]interface{}{"test": true},
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        eventBus.Emit(context.Background(), nil, event)
    }
}
```

### 5.2 代码质量标准

#### 静态代码分析
```makefile
# Makefile 中的质量检查命令
.PHONY: lint
lint:
	golangci-lint run ./...
	go vet ./...
	gofmt -s -w .
	go mod tidy

.PHONY: test
test:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: benchmark
benchmark:
	go test -bench=. -benchmem ./tests/benchmarks/

.PHONY: integration
integration:
	go test -tags=integration ./tests/integration/

.PHONY: e2e
e2e:
	go test -tags=e2e ./tests/e2e/
```

#### 代码覆盖率要求
- 核心业务逻辑: 90%+
- 接口实现: 85%+
- 工具和辅助函数: 80%+
- 总体覆盖率: 80%+

### 5.3 持续集成配置

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      redis:
        image: redis:6-alpine
        ports:
          - 6379:6379
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: '1.21'
    
    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run linting
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest
    
    - name: Run unit tests
      run: go test -race -coverprofile=coverage.out ./...
    
    - name: Run integration tests
      run: go test -tags=integration ./tests/integration/
      env:
        REDIS_URL: redis://localhost:6379
    
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
    
    - name: Build
      run: go build -v ./cmd/greensoulai
    
  security:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: '1.21'
    
    - name: Run gosec security scanner
      uses: securecodewarrior/github-action-gosec@master
      with:
        args: ./...
    
    - name: Run govulncheck
      run: |
        go install golang.org/x/vuln/cmd/govulncheck@latest
        govulncheck ./...

  performance:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: '1.21'
    
    - name: Run benchmarks
      run: go test -bench=. -benchmem ./tests/benchmarks/ | tee benchmark.txt
    
    - name: Store benchmark result
      uses: benchmark-action/github-action-benchmark@v1
      with:
        tool: 'go'
        output-file-path: benchmark.txt
```

---

## 6. 项目管理

### 6.1 开发流程

#### Git工作流
```
main          ─────●─────●─────●─────  (稳定版本)
               ╱         ╱
develop    ───●─────●───●─────●─────  (开发集成)
             ╱   ╱     ╱
feature/xxx ●───●─────╱              (功能分支)
           ╱
hotfix/xxx●                          (紧急修复)
```

#### 分支策略
- `main`: 生产就绪版本，只接受来自develop和hotfix的合并
- `develop`: 开发集成分支，最新的开发功能
- `feature/*`: 功能开发分支，从develop创建
- `hotfix/*`: 紧急修复分支，从main创建
- `release/*`: 发布准备分支，从develop创建

#### 提交规范
```
<type>(<scope>): <subject>

<body>

<footer>
```

类型说明:
- `feat`: 新功能
- `fix`: 错误修复  
- `docs`: 文档更新
- `style`: 代码格式化
- `refactor`: 重构
- `test`: 测试
- `chore`: 构建或工具相关

### 6.2 里程碑计划

| 里程碑 | 时间节点 | 主要交付物 | 验收标准 |
|--------|----------|------------|----------|
| **M1: 基础架构** | 第2周结束 | 事件系统、异步基础、安全模块 | 通过基础功能测试 |
| **M2: LLM&工具** | 第4周结束 | OpenAI集成、工具系统、RPM控制 | 通过LLM调用和工具执行测试 |
| **M3: Agent&Task** | 第6周结束 | 智能体和任务系统 | 通过端到端任务执行测试 |
| **M4: 记忆&知识** | 第8周结束 | 多类型记忆系统、知识管理 | 通过记忆存储和检索测试 |
| **M5: Crew&Flow** | 第10周结束 | 团队协作、工作流编排 | 通过复杂工作流测试 |
| **M6: 产品化** | 第12周结束 | 完整测试、文档、示例 | 通过生产就绪检查 |

### 6.3 风险管理

#### 技术风险及缓解措施

| 风险等级 | 风险描述 | 影响 | 概率 | 缓解措施 |
|---------|----------|------|------|----------|
| 🔴 高 | 事件系统性能瓶颈 | 高 | 中 | 早期性能测试、异步处理优化 |
| 🟡 中 | LLM API限制和成本 | 中 | 高 | 实现多Provider支持、本地缓存 |
| 🟡 中 | Go工作流装饰器复杂性 | 中 | 中 | 简化设计、使用反射和标签 |
| 🟢 低 | 第三方依赖变更 | 低 | 低 | 版本锁定、定期更新 |

#### 进度风险管控

**每周检查点:**
- 周一: 计划回顾和本周目标设定
- 周三: 中期进度检查和问题解决
- 周五: 周结总结和下周准备

**风险早期预警:**
- 连续2天未提交代码 → 黄色预警
- 单元测试覆盖率低于70% → 橙色预警
- 里程碑延期风险 → 红色预警

---

## 7. 成功标准

### 7.1 功能完整性
- [ ] **核心功能对标**: 100%实现Python版本的核心功能
- [ ] **API兼容性**: 提供清晰的迁移路径和兼容层
- [ ] **功能扩展**: 利用Go优势增加并发、性能优化功能

### 7.2 性能指标
- [ ] **执行效率**: 比Python版本提升2-3倍
- [ ] **内存使用**: 稳定的内存占用，无内存泄漏
- [ ] **并发能力**: 支持100+并发智能体
- [ ] **响应时间**: 单个任务执行延迟<5秒(不含LLM调用)

### 7.3 质量标准
- [ ] **测试覆盖**: 80%+单元测试覆盖率
- [ ] **代码质量**: 通过golangci-lint检查
- [ ] **文档完整**: API文档、使用指南、示例代码
- [ ] **安全性**: 通过安全扫描，无高危漏洞

### 7.4 用户体验
- [ ] **易用性**: 简化的安装和配置流程
- [ ] **错误处理**: 清晰的错误信息和调试支持
- [ ] **监控能力**: 完整的执行监控和日志记录
- [ ] **部署便利**: 单文件部署，容器化支持

### 7.5 社区接受度
- [ ] **文档质量**: 完整的文档和教程
- [ ] **示例丰富**: 涵盖基础到高级的使用场景
- [ ] **社区反馈**: 积极的社区反馈和贡献
- [ ] **生态集成**: 与现有工具生态的良好集成

---

## 总结

这个完整的实施指引涵盖了crewAI Go版本从概念到交付的全过程。通过深入的源码分析，我们确保了功能的完整性和准确性。分阶段的实施计划、完善的质量保证体系和风险管理措施，为项目成功提供了有力保障。

**关键成功要素:**
1. **技术架构正确** - 基于深入源码分析的精确设计
2. **实施计划合理** - 渐进式开发，风险可控
3. **质量标准严格** - 高覆盖率测试，企业级质量
4. **团队执行到位** - 经验丰富的Go开发团队

项目基础架构已经完成，现在可以开始实现业务逻辑模块。请确认是否开始执行下一阶段的开发工作? 🚀