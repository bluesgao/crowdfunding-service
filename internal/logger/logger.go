package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// Logger 自定义日志器
type Logger struct {
	level  LogLevel
	logger *log.Logger
}

var (
	// 默认日志器
	defaultLogger *Logger
)

// 初始化默认日志器
func init() {
	defaultLogger = New(INFO, os.Stdout)
}

// New 创建新的日志器
func New(level LogLevel, output io.Writer) *Logger {
	return &Logger{
		level:  level,
		logger: log.New(output, "", 0), // 不使用默认前缀，我们自己控制格式
	}
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// SetOutput 设置输出目标
func (l *Logger) SetOutput(output io.Writer) {
	l.logger.SetOutput(output)
}

// getCallerInfo 获取调用者信息
func (l *Logger) getCallerInfo() (string, int) {
	// 跳过当前函数和调用日志函数的函数
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return "unknown", 0
	}

	// 只保留文件名，不包含完整路径
	filename := filepath.Base(file)
	return filename, line
}

// formatMessage 格式化日志消息
func (l *Logger) formatMessage(level LogLevel, format string, args ...interface{}) string {
	// 获取调用者信息
	filename, line := l.getCallerInfo()

	// 获取当前时间
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// 日志级别字符串
	levelStr := ""
	switch level {
	case DEBUG:
		levelStr = "DEBUG"
	case INFO:
		levelStr = "INFO"
	case WARN:
		levelStr = "WARN"
	case ERROR:
		levelStr = "ERROR"
	case FATAL:
		levelStr = "FATAL"
	}

	// 格式化消息
	message := format
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	}

	// 组合最终格式: [时间] [级别] [文件:行号] 消息
	return fmt.Sprintf("[%s] [%s] [%s:%d] %s", timestamp, levelStr, filename, line, message)
}

// Debug 调试日志
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= DEBUG {
		l.logger.Println(l.formatMessage(DEBUG, format, args...))
	}
}

// Info 信息日志
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= INFO {
		l.logger.Println(l.formatMessage(INFO, format, args...))
	}
}

// Warn 警告日志
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= WARN {
		l.logger.Println(l.formatMessage(WARN, format, args...))
	}
}

// Error 错误日志
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= ERROR {
		l.logger.Println(l.formatMessage(ERROR, format, args...))
	}
}

// Fatal 致命错误日志（会调用os.Exit(1)）
func (l *Logger) Fatal(format string, args ...interface{}) {
	if l.level <= FATAL {
		l.logger.Println(l.formatMessage(FATAL, format, args...))
		os.Exit(1)
	}
}

// 全局函数，使用默认日志器

// SetLevel 设置默认日志器级别
func SetLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

// SetOutput 设置默认日志器输出
func SetOutput(output io.Writer) {
	defaultLogger.SetOutput(output)
}

// Debug 调试日志
func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

// Info 信息日志
func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

// Warn 警告日志
func Warn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

// Error 错误日志
func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

// Fatal 致命错误日志
func Fatal(format string, args ...interface{}) {
	defaultLogger.Fatal(format, args...)
}

// 兼容标准库log包的函数

// Printf 兼容标准库的Printf
func Printf(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

// Println 兼容标准库的Println
func Println(args ...interface{}) {
	message := fmt.Sprint(args...)
	defaultLogger.Info(message)
}

// Fatalf 兼容标准库的Fatalf
func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatal(format, args...)
}

// 日志级别字符串转换
func ParseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return INFO
	}
}
