package errors

import (
	"fmt"
)

// 预定义错误
var (
	ErrAgentNotFound          = fmt.Errorf("agent not found")
	ErrTaskTimeout            = fmt.Errorf("task execution timeout")
	ErrToolUsageLimitExceeded = fmt.Errorf("tool usage limit exceeded")
	ErrHumanInputRequired     = fmt.Errorf("human input required")
	ErrInvalidOutputFormat    = fmt.Errorf("invalid output format")
	ErrEventHandlerNotFound   = fmt.Errorf("event handler not found")
	ErrAsyncExecutorStopped   = fmt.Errorf("async executor stopped")
	ErrSecurityConfigInvalid  = fmt.Errorf("security config invalid")
)

// ErrorType 错误类型
type ErrorType string

const (
	ErrorTypeValidation ErrorType = "validation"
	ErrorTypeTimeout    ErrorType = "timeout"
	ErrorTypeNotFound   ErrorType = "not_found"
	ErrorTypePermission ErrorType = "permission"
	ErrorTypeInternal   ErrorType = "internal"
	ErrorTypeExternal   ErrorType = "external"
)

// CrewAIError 自定义错误
type CrewAIError struct {
	Type    ErrorType
	Message string
	Cause   error
}

// NewCrewAIError 创建新的CrewAI错误
func NewCrewAIError(errorType ErrorType, message string, cause error) *CrewAIError {
	return &CrewAIError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
	}
}

// Error 实现error接口
func (e *CrewAIError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap 返回原始错误
func (e *CrewAIError) Unwrap() error {
	return e.Cause
}

// IsValidationError 检查是否为验证错误
func IsValidationError(err error) bool {
	if crewAIErr, ok := err.(*CrewAIError); ok {
		return crewAIErr.Type == ErrorTypeValidation
	}
	return false
}

// IsTimeoutError 检查是否为超时错误
func IsTimeoutError(err error) bool {
	if crewAIErr, ok := err.(*CrewAIError); ok {
		return crewAIErr.Type == ErrorTypeTimeout
	}
	return false
}

// IsNotFoundError 检查是否为未找到错误
func IsNotFoundError(err error) bool {
	if crewAIErr, ok := err.(*CrewAIError); ok {
		return crewAIErr.Type == ErrorTypeNotFound
	}
	return false
}
