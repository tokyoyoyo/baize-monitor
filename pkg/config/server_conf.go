package config

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
	QueueSize         int
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
