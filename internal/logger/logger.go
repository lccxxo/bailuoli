package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
)

var (
	Logger      *zap.Logger
	atomicLevel zap.AtomicLevel // 全局日志级别控制器
)

// RotationConfig 日志轮转策略配置
type RotationConfig struct {
	MaxSize    int // MB
	MaxAge     int // Days
	MaxBackups int
	Compress   bool
}

func InitLogger(level string, outPutPaths []string, rotation RotationConfig) error {
	var logLevel zapcore.Level
	switch level {
	case "debug":
		logLevel = zapcore.DebugLevel
	case "info":
		logLevel = zapcore.InfoLevel
	case "warn":
		logLevel = zapcore.WarnLevel
	case "error":
		logLevel = zapcore.ErrorLevel
	case "panic":
		logLevel = zapcore.PanicLevel
	default:
		logLevel = zapcore.InfoLevel
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	atomicLevel = zap.NewAtomicLevelAt(logLevel)

	var cores []zapcore.Core
	for _, outPath := range outPutPaths {
		var writer zapcore.WriteSyncer
		if outPath == "stdout" {
			writer = zapcore.AddSync(os.Stdout)
		} else {
			writer = zapcore.AddSync(newLogWriter(outPath, rotation))
		}
		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			writer,
			atomicLevel,
		)
		cores = append(cores, core)
	}

	// todo 添加 syslog 输出

	Logger = zap.New(zapcore.NewTee(cores...), zap.AddCaller())
	return nil
}

// 日志切割
func newLogWriter(path string, rotation RotationConfig) *lumberjack.Logger {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create log directory: %v", err))
	}

	return &lumberjack.Logger{
		Filename:   path,
		MaxSize:    rotation.MaxSize,    // MB
		MaxBackups: rotation.MaxBackups, // 保留旧日志文件数
		MaxAge:     rotation.MaxAge,     // 保留天数
		Compress:   rotation.Compress,   // 启用GZIP压缩
	}
}

func UpdateLogLevel(level string) error {
	var newLevel zapcore.Level
	switch level {
	case "debug":
		newLevel = zapcore.DebugLevel
	case "info":
		newLevel = zapcore.InfoLevel
	case "warn":
		newLevel = zapcore.WarnLevel
	case "error":
		newLevel = zapcore.ErrorLevel
	default:
		return fmt.Errorf("invalid log level: %s", level)
	}

	// 原子化更新日志级别（线程安全）
	atomicLevel.SetLevel(newLevel)
	return nil
}

// Sync 刷新未写入的日志
func Sync() {
	_ = Logger.Sync()
}
