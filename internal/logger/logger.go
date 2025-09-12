package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
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
	zapLogger *zap.Logger
}

// LumberjackConfig lumberjack 配置
type LumberjackConfig struct {
	Filename   string // 日志文件路径
	MaxSize    int    // 每个日志文件的最大大小（MB）
	MaxBackups int    // 保留的旧日志文件数量
	MaxAge     int    // 保留日志文件的天数
	Compress   bool   // 是否压缩旧日志文件
}

var defaultLogger *Logger

func init() {
	var err error
	defaultLogger, err = New(INFO)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
}

// New 创建新的日志器
func New(level LogLevel) (*Logger, error) {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapLevelFromLogLevel(level))

	// 设置输出格式
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.EncodeName = zapcore.FullNameEncoder

	if level == DEBUG {
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zapLevelFromLogLevel(level))
	}

	zapLogger, err := config.Build(zap.AddCallerSkip(2))
	if err != nil {
		return nil, err
	}

	return &Logger{zapLogger: zapLogger}, nil
}

// NewWithConfig 使用自定义配置创建日志器
func NewWithConfig(config zap.Config) (*Logger, error) {
	zapLogger, err := config.Build(zap.AddCallerSkip(2))
	if err != nil {
		return nil, err
	}
	return &Logger{zapLogger: zapLogger}, nil
}

// NewWithFileRotation 创建支持文件轮转的日志器
func NewWithFileRotation(level LogLevel, logFile string) (*Logger, error) {
	config := LumberjackConfig{
		Filename:   logFile,
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}
	return NewWithLumberjackConfig(level, config)
}

// NewWithLumberjackConfig 使用自定义 lumberjack 配置创建日志器
func NewWithLumberjackConfig(level LogLevel, config LumberjackConfig) (*Logger, error) {
	// 设置默认值
	if config.MaxSize == 0 {
		config.MaxSize = 100
	}
	if config.MaxBackups == 0 {
		config.MaxBackups = 3
	}
	if config.MaxAge == 0 {
		config.MaxAge = 28
	}

	lumberjackLogger := &lumberjack.Logger{
		Filename:   config.Filename,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	}

	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(zapLevelFromLogLevel(level))

	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}
	zapConfig.EncoderConfig.CallerKey = "caller"
	zapConfig.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	zapConfig.EncoderConfig.LevelKey = "level"
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	zapConfig.EncoderConfig.MessageKey = "message"
	zapConfig.EncoderConfig.EncodeName = zapcore.FullNameEncoder

	if level == DEBUG {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.Level = zap.NewAtomicLevelAt(zapLevelFromLogLevel(level))
	}

	encoder := zapcore.NewJSONEncoder(zapConfig.EncoderConfig)
	core := zapcore.NewCore(encoder, zapcore.AddSync(lumberjackLogger), zapConfig.Level)
	zapLogger := zap.New(core, zap.AddCallerSkip(2), zap.AddCaller())

	return &Logger{zapLogger: zapLogger}, nil
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	newLogger, err := New(level)
	if err == nil {
		l.zapLogger.Sync()
		l.zapLogger = newLogger.zapLogger
	}
}

// Debug 调试日志
func (l *Logger) Debug(format string, args ...interface{}) {
	l.zapLogger.Debug(fmt.Sprintf(format, args...))
}

// Info 信息日志
func (l *Logger) Info(format string, args ...interface{}) {
	l.zapLogger.Info(fmt.Sprintf(format, args...))
}

// Warn 警告日志
func (l *Logger) Warn(format string, args ...interface{}) {
	l.zapLogger.Warn(fmt.Sprintf(format, args...))
}

// Error 错误日志
func (l *Logger) Error(format string, args ...interface{}) {
	l.zapLogger.Error(fmt.Sprintf(format, args...))
}

// Fatal 致命错误日志
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.zapLogger.Fatal(fmt.Sprintf(format, args...))
}

// Sync 同步日志
func (l *Logger) Sync() {
	l.zapLogger.Sync()
}

// With 添加字段
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{zapLogger: l.zapLogger.With(fields...)}
}

// 全局函数
func SetLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

// SetDefaultLogger 设置默认日志器
func SetDefaultLogger(l *Logger) {
	if defaultLogger != nil {
		defaultLogger.Sync()
	}
	defaultLogger = l
}

func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	defaultLogger.Fatal(format, args...)
}

func Sync() {
	defaultLogger.Sync()
}

func With(fields ...zap.Field) *Logger {
	return defaultLogger.With(fields...)
}

// 兼容性函数
func Printf(format string, args ...interface{}) {
	Info(format, args...)
}

func Println(args ...interface{}) {
	Info(strings.Join(strings.Fields(fmt.Sprint(args...)), " "))
}

func Fatalf(format string, args ...interface{}) {
	Fatal(format, args...)
}

// ParseLogLevel 解析日志级别字符串
func ParseLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn", "warning":
		return WARN
	case "error":
		return ERROR
	case "fatal":
		return FATAL
	default:
		return INFO
	}
}

// zapLevelFromLogLevel 转换日志级别
func zapLevelFromLogLevel(level LogLevel) zapcore.Level {
	switch level {
	case DEBUG:
		return zapcore.DebugLevel
	case INFO:
		return zapcore.InfoLevel
	case WARN:
		return zapcore.WarnLevel
	case ERROR:
		return zapcore.ErrorLevel
	case FATAL:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// GetZapLogger 获取底层的zap logger
func (l *Logger) GetZapLogger() *zap.Logger {
	return l.zapLogger
}

// GetDefaultZapLogger 获取默认的zap logger
func GetDefaultZapLogger() *zap.Logger {
	return defaultLogger.GetZapLogger()
}

// RedirectStdout 重定向标准输出到我们的logger
func RedirectStdout() {
	// 创建一个管道
	reader, writer, err := os.Pipe()
	if err != nil {
		Error("Failed to create pipe: %v", err)
		return
	}

	// 保存原始的stdout
	originalStdout := os.Stdout

	// 重定向stdout到我们的writer
	os.Stdout = writer

	// 启动goroutine来读取管道中的数据
	go func() {
		defer reader.Close()
		buffer := make([]byte, 1024)
		for {
			n, err := reader.Read(buffer)
			if err != nil {
				if err != io.EOF {
					Error("Failed to read from pipe: %v", err)
				}
				break
			}
			if n > 0 {
				// 将读取到的数据写入我们的logger
				message := strings.TrimSpace(string(buffer[:n]))
				if message != "" {
					Info("STDOUT: %s", message)
				}
			}
		}
	}()

	// 恢复stdout的函数
	restoreStdout := func() {
		writer.Close()
		os.Stdout = originalStdout
	}

	// 注册清理函数
	// 注意：这个函数需要在程序退出时调用
	_ = restoreStdout
}
