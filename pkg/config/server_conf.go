package config

import (
	"fmt"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

// ResponseManagerConfig response manager configuration
type ResponseManagerConfig struct {
	EngineFactoryConfig *ResponseEngineFactoryConfig `mapstructure:"engine_factory_config" yaml:"engine_factory_config"`
}

// ResponseEngineFactoryConfig response engine factory configuration
type ResponseEngineFactoryConfig struct {
	V1Config  *V1EngineConfig  `mapstructure:"v1_config" yaml:"v1_config"`
	V2cConfig *V2cEngineConfig `mapstructure:"v2c_config" yaml:"v2c_config"`
	V3Config  *V3EngineConfig  `mapstructure:"v3_config" yaml:"v3_config"`
}

// V1EngineConfig V1 engine configuration
type V1EngineConfig struct {
	ReadCommunity      string `mapstructure:"read_community" yaml:"read_community"`
	ReadWriteCommunity string `mapstructure:"read_write_community" yaml:"read_write_community"`
	Enabled            bool   `mapstructure:"enabled" yaml:"enabled"`
}

// V2cEngineConfig V2c engine configuration
type V2cEngineConfig struct {
	ReadCommunity      string `mapstructure:"read_community" yaml:"read_community"`
	ReadWriteCommunity string `mapstructure:"read_write_community" yaml:"read_write_community"`
	Enabled            bool   `mapstructure:"enabled" yaml:"enabled"`
}

// V3EngineConfig V3 engine configuration
type V3EngineConfig struct {
	Enabled        bool   `mapstructure:"enabled" yaml:"enabled"`
	UserName       string `mapstructure:"user_name" yaml:"user_name"`
	MsgFlags       string `mapstructure:"msg_flags" yaml:"msg_flags"`
	AuthProtocol   string `mapstructure:"auth_protocol" yaml:"auth_protocol"`
	PrivProtocol   string `mapstructure:"priv_protocol" yaml:"priv_protocol"`
	PrivPassphrase string `mapstructure:"priv_passphrase" yaml:"priv_passphrase"`
	AuthPassphrase string `mapstructure:"auth_passphrase" yaml:"auth_passphrase"`
}

type SNMPServerConfig struct {
	ReceiverConf    *ReceiverConfig    `mapstructure:"receiver_conf" yaml:"receiver_conf"`
	TrapHandlerConf *TrapHandlerConfig `mapstructure:"trap_handler_conf" yaml:"trap_handler_conf"`
	MidChannelSize  int                `mapstructure:"mid_channel_size" yaml:"mid_channel_size"`
}

type ReceiverConfig struct {
	Port uint16 `mapstructure:"port" yaml:"port"`
}

type TrapHandlerConfig struct { // Trap handler configuration
	WorkerCount       int `mapstructure:"worker_count" yaml:"worker_count"`
	LockTimeout       int `mapstructure:"lock_timeout" yaml:"lock_timeout"`
	ProcessingTimeout int `mapstructure:"processing_timeout" yaml:"processing_timeout"`
}

type PostGresConfig struct {
	Host      string `mapstructure:"host" yaml:"host"`
	Port      int    `mapstructure:"port" yaml:"port"`
	User      string `mapstructure:"user" yaml:"user"`
	Password  string `mapstructure:"password" yaml:"password"`
	Database  string `mapstructure:"database" yaml:"database"`
	SSLMode   string `mapstructure:"ssl_mode" yaml:"ssl_mode"`
	MaxConns  int    `mapstructure:"max_conns" yaml:"max_conns"`
	IdleConns int    `mapstructure:"idle_conns" yaml:"idle_conns"`
}

type AdminServerConfig struct {
	Port int `mapstructure:"port" yaml:"port"`
}

type AlertServerConfig struct {
	Port int `mapstructure:"port" yaml:"port"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level" yaml:"level"`
	Format     string `mapstructure:"format" yaml:"format"` // text, json
	Output     string `mapstructure:"output" yaml:"output"`
	LogDir     string `mapstructure:"log_dir" yaml:"log_dir"`
	MaxSizeMB  int    `mapstructure:"max_size_mb" yaml:"max_size_mb"`
	MaxBackups int    `mapstructure:"max_backups" yaml:"max_backups"`
	MaxAgeDays int    `mapstructure:"max_age_days" yaml:"max_age_days"`

	// GORM 专用配置
	GORM GORMLogConfig `mapstructure:"gorm" yaml:"gorm"`
}

// GORMLogConfig GORM 日志专用配置
type GORMLogConfig struct {
	Enabled                   bool          `mapstructure:"enabled" yaml:"enabled"`
	Level                     string        `mapstructure:"level" yaml:"level"` // silent, error, warn, info
	SlowThreshold             time.Duration `mapstructure:"slow_threshold" yaml:"slow_threshold"`
	IgnoreRecordNotFoundError bool          `mapstructure:"ignore_record_not_found_error" yaml:"ignore_record_not_found_error"`
	ParameterizedQueries      bool          `mapstructure:"parameterized_queries" yaml:"parameterized_queries"`
	Colorful                  bool          `mapstructure:"colorful" yaml:"colorful"`
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
	GinDebug              bool                   `mapstructure:"gin_debug" yaml:"gin_debug"`
	AdminServerConfig     *AdminServerConfig     `mapstructure:"admin_server_config" yaml:"admin_server_config"`
	AlertServerConfig     *AlertServerConfig     `mapstructure:"alert_server_config" yaml:"alert_server_config"`
	PostGresConfig        *PostGresConfig        `mapstructure:"post_gres_config" yaml:"post_gres_config"`
	SNMPServerConfig      *SNMPServerConfig      `mapstructure:"snmp_server_config" yaml:"snmp_server_config"`
	LogConfig             *LogConfig             `mapstructure:"log_config" yaml:"log_config"`
	ResponseManagerConfig *ResponseManagerConfig `mapstructure:"response_manager_config" yaml:"response_manager_config"`
}

func LoadServerConfig(configPath string) (*ServerConfig, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// 进入 server_config 子配置块
	sub := v.Sub("server_config")
	if sub == nil {
		return nil, fmt.Errorf("'server_config' section not found in config")
	}

	var cfg ServerConfig
	err := sub.Unmarshal(&cfg, func(c *mapstructure.DecoderConfig) {
		c.ZeroFields = true
		c.TagName = "mapstructure"
	})
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal server_config: %w", err)
	}

	return &cfg, nil
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
