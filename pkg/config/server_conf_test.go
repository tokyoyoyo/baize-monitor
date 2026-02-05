// pkg/config/server_config_test.go

package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

const configPath = "../../config/server.yaml"

var defaultCfg = &ServerConfig{
	GinDebug:          false,
	AdminServerConfig: &AdminServerConfig{Port: 8080},
	AlertServerConfig: &AlertServerConfig{Port: 8081},
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
	SNMPServerConfig: &SNMPServerConfig{
		ReceiverConf:    &ReceiverConfig{Port: 8888},
		TrapHandlerConf: &TrapHandlerConfig{WorkerCount: 5, LockTimeout: 30, ProcessingTimeout: 60},
		MidChannelSize:  100,
	},
	RedisConfig: &RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "qwer1234",
		DB:       0,
	},
	LogConfig: &LogConfig{
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
			SlowThreshold:             1 * time.Second,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      false,
			Colorful:                  false,
		},
	},
	ResponseManagerConfig: &ResponseManagerConfig{
		EngineFactoryConfig: &ResponseEngineFactoryConfig{
			V1Config: &V1EngineConfig{
				ReadCommunity: "public", ReadWriteCommunity: "private", Enabled: true,
			},
			V2cConfig: &V2cEngineConfig{
				ReadCommunity: "public", ReadWriteCommunity: "private", Enabled: true,
			},
			V3Config: &V3EngineConfig{
				Enabled:        false,
				UserName:       "admin",
				MsgFlags:       "AuthPriv",
				AuthProtocol:   "SHA",
				PrivProtocol:   "AES",
				AuthPassphrase: "auth_passphrase",
				PrivPassphrase: "priv_passphrase",
			},
		},
	},
}

func TestGenServerConfig(t *testing.T) {

	serverYAML := map[string]interface{}{
		"server_config": defaultCfg,
	}

	data, _ := yaml.Marshal(serverYAML)
	os.WriteFile(configPath, data, 0644)

	t.Run("Test load ServerConfig", func(t *testing.T) {

		actualCfg, err := LoadServerConfig(configPath)
		assert.NoError(t, err)
		assert.NotNil(t, actualCfg)

		// 直接比较两个 Go 结构体！
		assert.Equal(t, defaultCfg, actualCfg, "Loaded config does not match expected default config")
	})

}
