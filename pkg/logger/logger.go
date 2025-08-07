package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger 日志接口
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
}

// Field 日志字段
type Field struct {
	Key   string
	Value interface{}
}

// ConsoleLogger 控制台日志实现
type ConsoleLogger struct {
	logger *logrus.Logger
}

// NewConsoleLogger 创建控制台日志
func NewConsoleLogger() *ConsoleLogger {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(logrus.InfoLevel)

	return &ConsoleLogger{logger: logger}
}

// NewTestLogger 创建测试日志
func NewTestLogger() *ConsoleLogger {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(logrus.DebugLevel)

	return &ConsoleLogger{logger: logger}
}

func (l *ConsoleLogger) Debug(msg string, fields ...Field) {
	l.logger.WithFields(l.convertFields(fields)).Debug(msg)
}

func (l *ConsoleLogger) Info(msg string, fields ...Field) {
	l.logger.WithFields(l.convertFields(fields)).Info(msg)
}

func (l *ConsoleLogger) Warn(msg string, fields ...Field) {
	l.logger.WithFields(l.convertFields(fields)).Warn(msg)
}

func (l *ConsoleLogger) Error(msg string, fields ...Field) {
	l.logger.WithFields(l.convertFields(fields)).Error(msg)
}

func (l *ConsoleLogger) Fatal(msg string, fields ...Field) {
	l.logger.WithFields(l.convertFields(fields)).Fatal(msg)
}

func (l *ConsoleLogger) convertFields(fields []Field) logrus.Fields {
	result := make(logrus.Fields)
	for _, field := range fields {
		result[field.Key] = field.Value
	}
	return result
}
