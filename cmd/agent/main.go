package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/viper"

	"baize-monitor/internal/agent"
)

func main() {
	fmt.Println("Starting BaiZe Agent with new architecture...")

	// 加载配置
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建Agent实例
	agent := agent.NewAgent(config)

	// 运行Agent
	if err := agent.Run(); err != nil {
		log.Fatalf("Agent failed: %v", err)
	}
}

// loadConfig 加载配置
func loadConfig() (*agent.AgentConfig, error) {
	// 设置默认配置
	viper.SetDefault("server_url", "http://localhost:9988")
	viper.SetDefault("heartbeat_interval", 30)
	viper.SetDefault("metrics_port", 9100)
	viper.SetDefault("node_name", getHostname())
	viper.SetDefault("log_level", "info")

	// 读取配置文件
	viper.SetConfigName("agent")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// 读取环境变量
	viper.AutomaticEnv()

	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found, using defaults")
		} else {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	} else {
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}

	// 创建配置结构体
	config := &agent.AgentConfig{
		ServerURL:         viper.GetString("server_url"),
		HeartbeatInterval: time.Duration(viper.GetInt64("heartbeat_interval")) * time.Second,
		MetricsPort:       viper.GetInt("metrics_port"),
		NodeName:          viper.GetString("node_name"),
		LogLevel:          viper.GetString("log_level"),
	}

	// 打印配置信息
	fmt.Printf("Agent configuration loaded:\n")
	fmt.Printf("  Server URL: %s\n", config.ServerURL)
	fmt.Printf("  Heartbeat Interval: %v\n", config.HeartbeatInterval)
	fmt.Printf("  Metrics Port: %d\n", config.MetricsPort)
	fmt.Printf("  Node Name: %s\n", config.NodeName)

	return config, nil
}

// getHostname 获取主机名
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown-host"
	}
	return hostname
}
