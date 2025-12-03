package logger

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"baize-monitor/pkg/config"

	"gorm.io/gorm/logger"
)

var gormLoggerOnce sync.Once
var gormLoggerInstance logger.Interface

func GetGormLogger() logger.Interface {
	gormLoggerOnce.Do(func() {
		gormLoggerInstance = newGormLogger()
	})
	return gormLoggerInstance
}

// gormLogger 实现 gorm.Logger.Interface
type gormLogger struct {
	logger *slog.Logger
	config *config.GORMLogConfig
}

// newGormLogger 创建 GORM 专用日志实例
func newGormLogger() logger.Interface {
	init_logger()

	var writer io.Writer
	if globalConfig.Output == "file" {
		logPath := filepath.Join(globalConfig.LogDir, "gorm", "gorm.log")
		rolling := newRollingFile(
			logPath,
			globalConfig.MaxSizeMB,
			globalConfig.MaxBackups,
			globalConfig.MaxAgeDays,
		)
		writer = rolling
	} else {
		writer = os.Stdout
	}

	// 创建 slog handler
	var handler slog.Handler
	if globalConfig.Format == "json" {
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{
			Level: getGormSlogLevel(globalConfig.GORM.Level),
		})
	} else {
		handler = slog.NewTextHandler(writer, &slog.HandlerOptions{
			Level: getGormSlogLevel(globalConfig.GORM.Level),
		})
	}

	slogLogger := slog.New(handler).With(
		"module", "gorm",
		"component", "database",
	)

	return &gormLogger{
		logger: slogLogger,
		config: &globalConfig.GORM,
	}
}

// getGormSlogLevel 将 GORM 日志级别转换为 slog 级别
func getGormSlogLevel(gormLevel string) slog.Level {
	switch strings.ToLower(gormLevel) {
	case "silent":
		return slog.LevelError + 1 // 高于 Error 级别，基本不记录
	case "error":
		return slog.LevelError
	case "warn":
		return slog.LevelWarn
	case "info":
		return slog.LevelInfo
	default:
		return slog.LevelWarn
	}
}

// LogMode 实现 gorm.Logger.Interface
func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	return &newLogger
}

// Info 实现 gorm.Logger.Interface
func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if !l.config.Enabled {
		return
	}
	l.logger.InfoContext(ctx, fmt.Sprintf(msg, data...))
}

// Warn 实现 gorm.Logger.Interface
func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if !l.config.Enabled {
		return
	}
	l.logger.WarnContext(ctx, fmt.Sprintf(msg, data...))
}

// Error 实现 gorm.Logger.Interface
func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if !l.config.Enabled {
		return
	}
	l.logger.ErrorContext(ctx, fmt.Sprintf(msg, data...))
}

// Trace 实现 gorm.Logger.Interface
func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if !l.config.Enabled {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// 根据配置决定是否记录
	logLevel := l.getTraceLogLevel(elapsed, err)
	if logLevel == logger.Silent {
		return
	}

	// 构建日志字段
	fields := []interface{}{
		"elapsed", elapsed,
		"sql", sql,
		"rows", rows,
	}

	// 添加调用者信息（可选）
	if pc, file, line, ok := runtime.Caller(3); ok {
		funcName := runtime.FuncForPC(pc).Name()
		fields = append(fields, "caller", fmt.Sprintf("%s:%d %s", filepath.Base(file), line, funcName))
	}

	// 根据日志级别记录
	switch logLevel {
	case logger.Error:
		if err != nil {
			fields = append(fields, "error", err.Error())
		}
		l.logger.ErrorContext(ctx, "SQL execution failed", fields...)
	case logger.Warn:
		l.logger.WarnContext(ctx, "SQL executed", fields...)
	case logger.Info:
		l.logger.InfoContext(ctx, "SQL executed", fields...)
	}
}

// getTraceLogLevel 根据执行时间和错误决定日志级别
func (l *gormLogger) getTraceLogLevel(elapsed time.Duration, err error) logger.LogLevel {
	// 如果有错误且不是忽略的记录未找到错误，返回 Error 级别
	if err != nil && !(errors.Is(err, logger.ErrRecordNotFound) && l.config.IgnoreRecordNotFoundError) {
		return logger.Error
	}

	// 如果是慢查询，返回 Warn 级别
	if l.config.SlowThreshold > 0 && elapsed > l.config.SlowThreshold {
		return logger.Warn
	}

	// 根据配置的级别决定是否记录普通 SQL
	switch strings.ToLower(l.config.Level) {
	case "silent":
		return logger.Silent
	case "error":
		// error 级别下，只有错误才记录，正常查询不记录
		return logger.Silent
	case "warn":
		// warn 级别下，错误和慢查询已经在上面的逻辑处理，正常查询不记录
		return logger.Silent
	case "info":
		return logger.Info
	default:
		return logger.Silent
	}
}
