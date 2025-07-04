package logger

import (
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"ai-api-gateway/internal/infrastructure/config"
)

// Logger 日志记录器接口
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
}

// logrusLogger logrus日志记录器实现
type logrusLogger struct {
	logger *logrus.Logger
	entry  *logrus.Entry
}

// NewLogger 创建新的日志记录器
func NewLogger(config *config.LoggingConfig) Logger {
	logger := logrus.New()
	
	// 设置日志级别
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)
	
	// 设置日志格式
	switch strings.ToLower(config.Format) {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	default:
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}
	
	// 设置输出
	switch strings.ToLower(config.Output) {
	case "stdout":
		logger.SetOutput(os.Stdout)
	case "stderr":
		logger.SetOutput(os.Stderr)
	default:
		// 如果是文件路径，创建文件
		if config.Output != "" && config.Output != "stdout" && config.Output != "stderr" {
			file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				logger.SetOutput(os.Stdout)
			} else {
				logger.SetOutput(io.MultiWriter(os.Stdout, file))
			}
		} else {
			logger.SetOutput(os.Stdout)
		}
	}
	
	return &logrusLogger{
		logger: logger,
		entry:  logrus.NewEntry(logger),
	}
}

// Debug 记录调试日志
func (l *logrusLogger) Debug(args ...interface{}) {
	l.entry.Debug(args...)
}

// Debugf 记录格式化调试日志
func (l *logrusLogger) Debugf(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

// Info 记录信息日志
func (l *logrusLogger) Info(args ...interface{}) {
	l.entry.Info(args...)
}

// Infof 记录格式化信息日志
func (l *logrusLogger) Infof(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

// Warn 记录警告日志
func (l *logrusLogger) Warn(args ...interface{}) {
	l.entry.Warn(args...)
}

// Warnf 记录格式化警告日志
func (l *logrusLogger) Warnf(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

// Error 记录错误日志
func (l *logrusLogger) Error(args ...interface{}) {
	l.entry.Error(args...)
}

// Errorf 记录格式化错误日志
func (l *logrusLogger) Errorf(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

// Fatal 记录致命错误日志
func (l *logrusLogger) Fatal(args ...interface{}) {
	l.entry.Fatal(args...)
}

// Fatalf 记录格式化致命错误日志
func (l *logrusLogger) Fatalf(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

// WithField 添加字段
func (l *logrusLogger) WithField(key string, value interface{}) Logger {
	return &logrusLogger{
		logger: l.logger,
		entry:  l.entry.WithField(key, value),
	}
}

// WithFields 添加多个字段
func (l *logrusLogger) WithFields(fields map[string]interface{}) Logger {
	return &logrusLogger{
		logger: l.logger,
		entry:  l.entry.WithFields(logrus.Fields(fields)),
	}
}

// 全局日志记录器实例
var globalLogger Logger

// InitGlobalLogger 初始化全局日志记录器
func InitGlobalLogger(config *config.LoggingConfig) {
	globalLogger = NewLogger(config)
}

// GetLogger 获取全局日志记录器
func GetLogger() Logger {
	if globalLogger == nil {
		// 如果没有初始化，使用默认配置
		defaultConfig := &config.LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		}
		globalLogger = NewLogger(defaultConfig)
	}
	return globalLogger
}

// 便捷方法
func Debug(args ...interface{}) {
	GetLogger().Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}

func Info(args ...interface{}) {
	GetLogger().Info(args...)
}

func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

func Warn(args ...interface{}) {
	GetLogger().Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

func Error(args ...interface{}) {
	GetLogger().Error(args...)
}

func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	GetLogger().Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}

func WithField(key string, value interface{}) Logger {
	return GetLogger().WithField(key, value)
}

func WithFields(fields map[string]interface{}) Logger {
	return GetLogger().WithFields(fields)
}
