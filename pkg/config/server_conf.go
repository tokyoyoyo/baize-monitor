package config

import (
	"time"
)

// ResponseManagerConfig response manager configuration
type ResponseManagerConfig struct {
	EngineFactoryConfig *ResponseEngineFactoryConfig
}

// ResponseEngineFactoryConfig response engine factory configuration
type ResponseEngineFactoryConfig struct {
	V1Config  *V1EngineConfig
	V2cConfig *V2cEngineConfig
	V3Config  *V3EngineConfig
}

// V1EngineConfig V1 engine configuration
type V1EngineConfig struct {
	ReadCommunity      string
	ReadWriteCommunity string
	Enabled            bool
}

// V2cEngineConfig V2c engine configuration
type V2cEngineConfig struct {
	ReadCommunity      string
	ReadWriteCommunity string
	Enabled            bool
}

// V3EngineConfig V3 engine configuration
type V3EngineConfig struct {
	Enabled        bool
	UserName       string
	MsgFlags       string
	AuthProtocol   string
	PrivProtocol   string
	PrivPassphrase string
	AuthPassphrase string
}

// RedisConfig Redis configuration
type RedisConfig struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Password string `yaml:"password" json:"password"`
	DB       int    `yaml:"db" json:"db"`
}

type SNMPServerConfig struct {
	ReceiverConf    *ReceiverConfig
	TrapHandlerConf *TrapHandlerConfig
	MidChannelSize  int
}

type ReceiverConfig struct {
	Port uint16
}

type TrapHandlerConfig struct { // Trap handler configuration
	WorkerCount       int
	LockTimeout       int
	ProcessingTimeout int
}

type PostGresConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	User      string `yaml:"user"`
	Password  string `yaml:"password"`
	Database  string `yaml:"database"`
	SSLMode   string `yaml:"ssl_mode"`
	MaxConns  int    `yaml:"max_conns"`
	IdleConns int    `yaml:"idle_conns"`
}

type AdminServerConfig struct {
	Addr string
}

type AlertServerConfig struct {
	Addr string
}

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

type ServerConfig struct {
	GinEnv            string
	AdminServerConfig *AdminServerConfig
	AlertServerConfig *AlertServerConfig
	PostGresConfig    *PostGresConfig
	SNMPServerConfig  *SNMPServerConfig
}

func LoadServerConfig() (*ServerConfig, error) {
	return &ServerConfig{}, nil
}

func LoadTestMockServerConfig() (*ServerConfig, error) {
	return &ServerConfig{
		PostGresConfig: &PostGresConfig{
			Host:      "localhost",
			Port:      5432,
			User:      "postgres",
			Password:  "qwer1234",
			Database:  "baize_test",
			SSLMode:   "disable",
			MaxConns:  10,
			IdleConns: 5,
		},
	}, nil
}
