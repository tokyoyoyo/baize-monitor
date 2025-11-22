// pkg/models/log_config.go
package models

import "time"

// LogConfig 日志配置
type LogConfig struct {
	Level      string `json:"level" yaml:"level"`
	Format     string `json:"format" yaml:"format"`
	Output     string `json:"output" yaml:"output"`
	LogDir     string `json:"log_dir" yaml:"log_dir"`
	MaxSizeMB  int    `json:"max_size_mb" yaml:"max_size_mb"`
	MaxBackups int    `json:"max_backups" yaml:"max_backups"`
	MaxAgeDays int    `json:"max_age_days" yaml:"max_age_days"`

	// GORM 专用配置
	GORM GORMLogConfig `json:"gorm" yaml:"gorm"`
}

// GORMLogConfig GORM 日志专用配置
type GORMLogConfig struct {
	Enabled                   bool          `json:"enabled" yaml:"enabled"`
	Level                     string        `json:"level" yaml:"level"` // silent, error, warn, info
	SlowThreshold             time.Duration `json:"slow_threshold" yaml:"slow_threshold"`
	IgnoreRecordNotFoundError bool          `json:"ignore_record_not_found_error" yaml:"ignore_record_not_found_error"`
	ParameterizedQueries      bool          `json:"parameterized_queries" yaml:"parameterized_queries"`
	Colorful                  bool          `json:"colorful" yaml:"colorful"`
}

// DefaultLogConfig
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level:      "info",
		Format:     "text",
		Output:     "file",
		LogDir:     "/var/log/baize",
		MaxSizeMB:  100,
		MaxBackups: 10,
		MaxAgeDays: 7,
		GORM: GORMLogConfig{
			Enabled:                   true,
			Level:                     "warn",
			SlowThreshold:             time.Second,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      false,
			Colorful:                  false,
		},
	}
}
