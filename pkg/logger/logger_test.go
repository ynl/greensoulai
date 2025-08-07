package logger

import (
	"testing"
)

func TestNewConsoleLogger(t *testing.T) {
	logger := NewConsoleLogger()

	if logger == nil {
		t.Fatal("expected logger to be created")
	}
}

func TestNewTestLogger(t *testing.T) {
	logger := NewTestLogger()

	if logger == nil {
		t.Fatal("expected test logger to be created")
	}
}

func TestConsoleLogger_Methods(t *testing.T) {
	logger := NewTestLogger()

	// 测试所有日志级别方法
	logger.Debug("debug message", Field{Key: "key1", Value: "value1"})
	logger.Info("info message", Field{Key: "key2", Value: "value2"})
	logger.Warn("warn message", Field{Key: "key3", Value: "value3"})
	logger.Error("error message", Field{Key: "key4", Value: "value4"})

	// 这些方法不应该panic
	// 注意：Fatal方法在测试中会导致程序退出，所以不测试
}

func TestField_Structure(t *testing.T) {
	field := Field{Key: "test_key", Value: "test_value"}

	if field.Key != "test_key" {
		t.Errorf("expected key 'test_key', got '%s'", field.Key)
	}

	if field.Value != "test_value" {
		t.Errorf("expected value 'test_value', got '%v'", field.Value)
	}
}

func TestConsoleLogger_ConvertFields(t *testing.T) {
	consoleLogger := NewTestLogger()

	fields := []Field{
		{Key: "string_key", Value: "string_value"},
		{Key: "int_key", Value: 42},
		{Key: "bool_key", Value: true},
	}

	converted := consoleLogger.convertFields(fields)

	if len(converted) != 3 {
		t.Errorf("expected 3 fields, got %d", len(converted))
	}

	if converted["string_key"] != "string_value" {
		t.Errorf("expected 'string_value', got '%v'", converted["string_key"])
	}

	if converted["int_key"] != 42 {
		t.Errorf("expected 42, got '%v'", converted["int_key"])
	}

	if converted["bool_key"] != true {
		t.Errorf("expected true, got '%v'", converted["bool_key"])
	}
}
