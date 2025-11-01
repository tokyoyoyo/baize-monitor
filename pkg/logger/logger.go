package logger

import (
	"baize-monitor/pkg/models"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	globalConfig *models.LogConfig
	rollingFiles []*rollingFile
	mutex        sync.Mutex
)

var (
	Snmp_logger = getLogger("snmp")
)

func loadLoggerConfig() (*models.LogConfig, error) {
	// TODO: load logger config from file
	return nil, fmt.Errorf("not implemented")
}

func getLogger(model string) *slog.Logger {
	init_logger()
	return forModule(model)
}

func init_logger() {
	if globalConfig != nil {
		return
	}
	conf, err := loadLoggerConfig()
	if err != nil {
		conf = models.DefaultLogConfig()
	}

	err = init_log_by_cfg(conf)
	if err != nil {
		panic(err)
	}
}

// init initializes the global logging configuration (does not create specific logger)
func init_log_by_cfg(cfg *models.LogConfig) error {
	if cfg == nil {
		return fmt.Errorf("invalid logger config")
	}
	// Set default values
	if cfg.MaxSizeMB <= 0 {
		cfg.MaxSizeMB = 100
	}
	if cfg.MaxBackups <= 0 {
		cfg.MaxBackups = 10
	}
	if cfg.MaxAgeDays <= 0 {
		cfg.MaxAgeDays = 7
	}
	if cfg.LogDir == "" {
		cfg.LogDir = "/var/log/baize"
	}

	globalConfig = cfg
	return nil
}

// forModule returns a module-specific logger instance
// moduleName will be used as subdirectory and filename, e.g. "snmp" → /var/log/baize/snmp/snmp.log
func forModule(moduleName string) *slog.Logger {
	if globalConfig == nil {
		panic("logger not initialized, call logger.Init() first")
	}

	var writer io.Writer

	if globalConfig.Output == "file" {
		logPath := filepath.Join(globalConfig.LogDir, moduleName, moduleName+".log")
		rolling := newRollingFile(
			logPath,
			globalConfig.MaxSizeMB,
			globalConfig.MaxBackups,
			globalConfig.MaxAgeDays,
		)
		writer = io.MultiWriter(rolling, os.Stdout)
	} else {
		writer = os.Stdout
	}

	var level slog.Level
	switch strings.ToLower(globalConfig.Level) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var handler slog.Handler
	if globalConfig.Format == "json" {
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewTextHandler(writer, &slog.HandlerOptions{Level: level})
	}

	return slog.New(handler).With("module", moduleName)
}

// Sync flushes all log data to disk
func Sync() error {
	mutex.Lock()
	defer mutex.Unlock()

	for _, rf := range rollingFiles {
		if err := rf.Sync(); err != nil {
			return err
		}
	}
	return nil
}
