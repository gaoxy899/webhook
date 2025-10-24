// logger/logger.go
package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	once   sync.Once
	logger *slog.Logger
)

// 配置结构体
type Config struct {
	Filename   string // 日志文件路径
	Level      string
	MaxSize    int  // 单文件最大MB
	MaxBackups int  // 保留几个旧日志
	MaxAge     int  // 保留天数
	Compress   bool // 是否压缩归档
	JSON       bool // 是否使用 JSON 格式
}

// 初始化日志（只执行一次）
func Init(cfg Config) {
	once.Do(func() {
		writer := &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
			LocalTime:  true,
		}

		level := levelFromString(cfg.Level)
		var handler slog.Handler
		if cfg.JSON {
			handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: level})
		} else {
			handler = slog.NewTextHandler(writer, &slog.HandlerOptions{Level: level})
		}

		logger = slog.New(handler)
	})
}

func levelFromString(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		fmt.Printf("unknown log level: %s, defaulting to INFO\n", s)
		return slog.LevelInfo
	}
}

// 获取 slog.Logger 实例
func Slog() *slog.Logger {
	return logger
}

// 快捷调用封装（结构化日志）
func Debug(msg string, args ...any) { logger.Debug(msg, args...) }
func Info(msg string, args ...any)  { logger.Info(msg, args...) }
func Warn(msg string, args ...any)  { logger.Warn(msg, args...) }
func Error(msg string, args ...any) { logger.Error(msg, args...) }

// Panic 和 Fatal 用原生 fmt + os 处理
func Panic(msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	logger.Error("PANIC", slog.String("msg", formatted))
	panic(formatted)
}

func Fatal(msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	logger.Error("FATAL", slog.String("msg", formatted))
	os.Exit(1)
}

/*

logger.Init(logger.Config{
		Filename:   "./logs/app.log",
		Level:      "info",
		MaxSize:    20, // 每个文件最多 20MB
		MaxBackups: 7,
		MaxAge:     3, // 最多保存 3 天
		Compress:   true,
		JSON:       true, // 结构化日志
	})

	logger.Info("app started", "version", "1.0.0")
    logger.Debug("debug message", "detail", "init complete")
    logger.Warn("disk almost full")
    logger.Error("db connection failed", "retrying", true)
*/
