package agent

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ynl/greensoulai/pkg/logger"
)

// ConsoleInputHandler 实现了控制台人工输入处理器
type ConsoleInputHandler struct {
	timeout time.Duration
	logger  logger.Logger
}

// NewConsoleInputHandler 创建控制台输入处理器
func NewConsoleInputHandler(log logger.Logger) *ConsoleInputHandler {
	if log == nil {
		log = logger.NewConsoleLogger()
	}

	return &ConsoleInputHandler{
		timeout: 5 * time.Minute, // 默认5分钟超时
		logger:  log,
	}
}

// RequestInput 请求用户输入
func (c *ConsoleInputHandler) RequestInput(ctx context.Context, prompt string, options []string) (string, error) {
	// 创建超时上下文
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// 显示提示信息
	fmt.Printf("\n🤖 %s\n", prompt)

	if len(options) > 0 {
		fmt.Println("\n选项:")
		for i, option := range options {
			fmt.Printf("  %d) %s\n", i+1, option)
		}
		fmt.Println()
	}

	fmt.Print("请输入: ")

	// 创建输入通道
	inputChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	// 在goroutine中等待用户输入
	go func() {
		defer close(inputChan)
		defer close(errorChan)

		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			errorChan <- fmt.Errorf("reading input failed: %w", err)
			return
		}

		input = strings.TrimSpace(input)
		inputChan <- input
	}()

	// 等待输入或超时
	select {
	case input := <-inputChan:
		c.logger.Info("Human input received",
			logger.Field{Key: "input_length", Value: len(input)},
		)
		return input, nil

	case err := <-errorChan:
		c.logger.Error("Human input error",
			logger.Field{Key: "error", Value: err},
		)
		return "", err

	case <-ctx.Done():
		fmt.Println("\n⏰ 输入超时")
		c.logger.Warn("Human input timeout",
			logger.Field{Key: "timeout", Value: c.timeout},
		)
		return "", fmt.Errorf("input timeout after %v: %w", c.timeout, ctx.Err())
	}
}

// IsInteractive 返回是否为交互式
func (c *ConsoleInputHandler) IsInteractive() bool {
	return true
}

// SetTimeout 设置超时时间
func (c *ConsoleInputHandler) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// GetTimeout 获取超时时间
func (c *ConsoleInputHandler) GetTimeout() time.Duration {
	return c.timeout
}

// MockInputHandler 用于测试的模拟输入处理器
type MockInputHandler struct {
	responses []string
	index     int
	timeout   time.Duration
	logger    logger.Logger
}

// NewMockInputHandler 创建模拟输入处理器
func NewMockInputHandler(responses []string, log logger.Logger) *MockInputHandler {
	if log == nil {
		log = logger.NewConsoleLogger()
	}

	return &MockInputHandler{
		responses: responses,
		index:     0,
		timeout:   1 * time.Second, // 测试用，很短的超时
		logger:    log,
	}
}

// RequestInput 返回预设的响应
func (m *MockInputHandler) RequestInput(ctx context.Context, prompt string, options []string) (string, error) {
	m.logger.Info("Mock input requested",
		logger.Field{Key: "prompt", Value: prompt},
		logger.Field{Key: "options_count", Value: len(options)},
	)

	if m.index >= len(m.responses) {
		return "", fmt.Errorf("no more mock responses available")
	}

	response := m.responses[m.index]
	m.index++

	m.logger.Info("Mock input provided",
		logger.Field{Key: "response", Value: response},
	)

	return response, nil
}

// IsInteractive 返回是否为交互式
func (m *MockInputHandler) IsInteractive() bool {
	return false // Mock不是真正的交互式
}

// SetTimeout 设置超时时间
func (m *MockInputHandler) SetTimeout(timeout time.Duration) {
	m.timeout = timeout
}

// GetTimeout 获取超时时间
func (m *MockInputHandler) GetTimeout() time.Duration {
	return m.timeout
}

// AddResponse 添加响应（用于动态修改测试数据）
func (m *MockInputHandler) AddResponse(response string) {
	m.responses = append(m.responses, response)
}

// Reset 重置索引
func (m *MockInputHandler) Reset() {
	m.index = 0
}

// HasMoreResponses 检查是否还有更多响应
func (m *MockInputHandler) HasMoreResponses() bool {
	return m.index < len(m.responses)
}

// GetResponseCount 获取响应总数
func (m *MockInputHandler) GetResponseCount() int {
	return len(m.responses)
}

// GetCurrentIndex 获取当前索引
func (m *MockInputHandler) GetCurrentIndex() int {
	return m.index
}

// PrefilledInputHandler 预填充输入处理器（用于批量处理）
type PrefilledInputHandler struct {
	inputMap map[string]string // prompt -> response 映射
	timeout  time.Duration
	logger   logger.Logger
}

// NewPrefilledInputHandler 创建预填充输入处理器
func NewPrefilledInputHandler(inputMap map[string]string, log logger.Logger) *PrefilledInputHandler {
	if log == nil {
		log = logger.NewConsoleLogger()
	}

	return &PrefilledInputHandler{
		inputMap: inputMap,
		timeout:  1 * time.Second,
		logger:   log,
	}
}

// RequestInput 根据提示返回预填充的输入
func (p *PrefilledInputHandler) RequestInput(ctx context.Context, prompt string, options []string) (string, error) {
	response, exists := p.inputMap[prompt]
	if !exists {
		return "", fmt.Errorf("no prefilled input found for prompt: %s", prompt)
	}

	p.logger.Info("Prefilled input provided",
		logger.Field{Key: "prompt", Value: prompt},
		logger.Field{Key: "response", Value: response},
	)

	return response, nil
}

// IsInteractive 返回是否为交互式
func (p *PrefilledInputHandler) IsInteractive() bool {
	return false
}

// SetTimeout 设置超时时间
func (p *PrefilledInputHandler) SetTimeout(timeout time.Duration) {
	p.timeout = timeout
}

// GetTimeout 获取超时时间
func (p *PrefilledInputHandler) GetTimeout() time.Duration {
	return p.timeout
}

// AddInput 添加输入映射
func (p *PrefilledInputHandler) AddInput(prompt, response string) {
	p.inputMap[prompt] = response
}

// RemoveInput 移除输入映射
func (p *PrefilledInputHandler) RemoveInput(prompt string) {
	delete(p.inputMap, prompt)
}

// HasInput 检查是否有指定提示的输入
func (p *PrefilledInputHandler) HasInput(prompt string) bool {
	_, exists := p.inputMap[prompt]
	return exists
}

// Clear 清空所有输入映射
func (p *PrefilledInputHandler) Clear() {
	p.inputMap = make(map[string]string)
}

// NoInputHandler 不需要人工输入的处理器
type NoInputHandler struct {
	timeout time.Duration
	logger  logger.Logger
}

// NewNoInputHandler 创建无输入处理器
func NewNoInputHandler(log logger.Logger) *NoInputHandler {
	if log == nil {
		log = logger.NewConsoleLogger()
	}

	return &NoInputHandler{
		timeout: 0,
		logger:  log,
	}
}

// RequestInput 总是返回错误，因为不支持输入
func (n *NoInputHandler) RequestInput(ctx context.Context, prompt string, options []string) (string, error) {
	n.logger.Warn("Input requested but not supported",
		logger.Field{Key: "prompt", Value: prompt},
	)
	return "", fmt.Errorf("human input is not supported by this handler")
}

// IsInteractive 返回是否为交互式
func (n *NoInputHandler) IsInteractive() bool {
	return false
}

// SetTimeout 设置超时时间
func (n *NoInputHandler) SetTimeout(timeout time.Duration) {
	n.timeout = timeout
}

// GetTimeout 获取超时时间
func (n *NoInputHandler) GetTimeout() time.Duration {
	return n.timeout
}

// 确保所有处理器都实现了HumanInputHandler接口
var (
	_ HumanInputHandler = (*ConsoleInputHandler)(nil)
	_ HumanInputHandler = (*MockInputHandler)(nil)
	_ HumanInputHandler = (*PrefilledInputHandler)(nil)
	_ HumanInputHandler = (*NoInputHandler)(nil)
)

// InputHandlerFactory 输入处理器工厂
type InputHandlerFactory struct{}

// NewInputHandlerFactory 创建输入处理器工厂
func NewInputHandlerFactory() *InputHandlerFactory {
	return &InputHandlerFactory{}
}

// CreateConsoleHandler 创建控制台处理器
func (f *InputHandlerFactory) CreateConsoleHandler(log logger.Logger) HumanInputHandler {
	return NewConsoleInputHandler(log)
}

// CreateMockHandler 创建模拟处理器
func (f *InputHandlerFactory) CreateMockHandler(responses []string, log logger.Logger) HumanInputHandler {
	return NewMockInputHandler(responses, log)
}

// CreatePrefilledHandler 创建预填充处理器
func (f *InputHandlerFactory) CreatePrefilledHandler(inputMap map[string]string, log logger.Logger) HumanInputHandler {
	return NewPrefilledInputHandler(inputMap, log)
}

// CreateNoInputHandler 创建无输入处理器
func (f *InputHandlerFactory) CreateNoInputHandler(log logger.Logger) HumanInputHandler {
	return NewNoInputHandler(log)
}

// CreateHandler 根据类型创建处理器
func (f *InputHandlerFactory) CreateHandler(handlerType string, config interface{}, log logger.Logger) (HumanInputHandler, error) {
	switch handlerType {
	case "console":
		return f.CreateConsoleHandler(log), nil

	case "mock":
		if responses, ok := config.([]string); ok {
			return f.CreateMockHandler(responses, log), nil
		}
		return nil, fmt.Errorf("invalid config for mock handler: expected []string")

	case "prefilled":
		if inputMap, ok := config.(map[string]string); ok {
			return f.CreatePrefilledHandler(inputMap, log), nil
		}
		return nil, fmt.Errorf("invalid config for prefilled handler: expected map[string]string")

	case "none":
		return f.CreateNoInputHandler(log), nil

	default:
		return nil, fmt.Errorf("unknown handler type: %s", handlerType)
	}
}
