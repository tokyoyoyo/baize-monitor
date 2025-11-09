package models

type LogConfig struct {
	Level      string `yaml:"level"`       // debug, info, warn, error
	Format     string `yaml:"format"`      // file,text, json
	Output     string `yaml:"output"`      // stdout, file
	LogDir     string `yaml:"log_dir"`     // e.g., /var/log/baize
	MaxSizeMB  int    `yaml:"max_size"`    // per file, default 100
	MaxBackups int    `yaml:"max_backups"` // default 5
	MaxAgeDays int    `yaml:"max_age"`     // default 7
}

// DefaultConfig returns a Config struct populated with default values.
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level:      "info",
		Format:     "text",
		Output:     "stdout",
		LogDir:     "/var/log/baize",
		MaxSizeMB:  100,
		MaxBackups: 10,
		MaxAgeDays: 7,
	}
}

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
