package errors

import (
	"testing"
)

func TestCrewAIError_Creation(t *testing.T) {
	cause := NewCrewAIError(ErrorTypeValidation, "test message", nil)

	if cause.Type != ErrorTypeValidation {
		t.Errorf("expected type %s, got %s", ErrorTypeValidation, cause.Type)
	}

	if cause.Message != "test message" {
		t.Errorf("expected message 'test message', got '%s'", cause.Message)
	}

	if cause.Cause != nil {
		t.Error("expected nil cause")
	}
}

func TestCrewAIError_WithCause(t *testing.T) {
	originalErr := NewCrewAIError(ErrorTypeInternal, "original error", nil)
	cause := NewCrewAIError(ErrorTypeValidation, "test message", originalErr)

	if cause.Type != ErrorTypeValidation {
		t.Errorf("expected type %s, got %s", ErrorTypeValidation, cause.Type)
	}

	if cause.Cause != originalErr {
		t.Error("expected cause to be the original error")
	}
}

func TestCrewAIError_Error(t *testing.T) {
	// 测试无原因的错误
	err := NewCrewAIError(ErrorTypeTimeout, "timeout occurred", nil)
	errorStr := err.Error()

	if errorStr != "timeout: timeout occurred" {
		t.Errorf("expected 'timeout: timeout occurred', got '%s'", errorStr)
	}

	// 测试有原因的错误
	cause := NewCrewAIError(ErrorTypeInternal, "internal error", nil)
	err = NewCrewAIError(ErrorTypeValidation, "validation failed", cause)
	errorStr = err.Error()

	expected := "validation: validation failed (caused by: internal: internal error)"
	if errorStr != expected {
		t.Errorf("expected '%s', got '%s'", expected, errorStr)
	}
}

func TestCrewAIError_Unwrap(t *testing.T) {
	originalErr := NewCrewAIError(ErrorTypeInternal, "original error", nil)
	cause := NewCrewAIError(ErrorTypeValidation, "test message", originalErr)

	unwrapped := cause.Unwrap()
	if unwrapped != originalErr {
		t.Error("expected Unwrap to return the original error")
	}
}

func TestErrorType_Constants(t *testing.T) {
	// 测试错误类型常量
	types := []ErrorType{
		ErrorTypeValidation,
		ErrorTypeTimeout,
		ErrorTypeNotFound,
		ErrorTypePermission,
		ErrorTypeInternal,
		ErrorTypeExternal,
	}

	for _, errorType := range types {
		if errorType == "" {
			t.Errorf("error type should not be empty")
		}
	}
}

func TestError_Checking(t *testing.T) {
	// 测试错误类型检查
	validationErr := NewCrewAIError(ErrorTypeValidation, "validation error", nil)
	timeoutErr := NewCrewAIError(ErrorTypeTimeout, "timeout error", nil)
	notFoundErr := NewCrewAIError(ErrorTypeNotFound, "not found error", nil)

	// 测试IsValidationError
	if !IsValidationError(validationErr) {
		t.Error("expected validation error to be detected")
	}
	if IsValidationError(timeoutErr) {
		t.Error("expected timeout error to not be detected as validation error")
	}

	// 测试IsTimeoutError
	if !IsTimeoutError(timeoutErr) {
		t.Error("expected timeout error to be detected")
	}
	if IsTimeoutError(validationErr) {
		t.Error("expected validation error to not be detected as timeout error")
	}

	// 测试IsNotFoundError
	if !IsNotFoundError(notFoundErr) {
		t.Error("expected not found error to be detected")
	}
	if IsNotFoundError(validationErr) {
		t.Error("expected validation error to not be detected as not found error")
	}
}

func TestPredefinedErrors(t *testing.T) {
	// 测试预定义错误
	predefinedErrors := []error{
		ErrAgentNotFound,
		ErrTaskTimeout,
		ErrToolUsageLimitExceeded,
		ErrHumanInputRequired,
		ErrInvalidOutputFormat,
		ErrEventHandlerNotFound,
		ErrAsyncExecutorStopped,
		ErrSecurityConfigInvalid,
	}

	for _, err := range predefinedErrors {
		if err == nil {
			t.Error("predefined error should not be nil")
		}
		if err.Error() == "" {
			t.Error("predefined error should have a message")
		}
	}
}
