package core

import (
	"baize-monitor/pkg/constants"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Plugin 插件接口
type Plugin interface {
	Name() string
	Description() string
	Tools() []string
	Parameters() []string
	Interval() time.Duration
	Execute() (interface{}, error)
	LastExecutionStatus() constants.ExecutionStatus
	LastExecutionTime() time.Time
	Enabled() bool
	SetEnabled(bool)

	// Prometheus Collector接口
	Collect(ch chan<- prometheus.Metric)
	Describe(ch chan<- *prometheus.Desc)
}

// PluginRegistry 插件注册表
type PluginRegistry struct {
	plugins map[string]Plugin
}

// NewPluginRegistry 创建插件注册表
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		plugins: make(map[string]Plugin),
	}
}

// Register 注册插件
func (pr *PluginRegistry) Register(plugin Plugin) {
	pr.plugins[plugin.Name()] = plugin
}

// GetAllPlugins 获取所有插件
func (pr *PluginRegistry) GetAllPlugins() []Plugin {
	plugins := make([]Plugin, 0, len(pr.plugins))
	for _, plugin := range pr.plugins {
		plugins = append(plugins, plugin)
	}
	return plugins
}

// Global plugin registries for each exporter type
var (
	MachineInfoPlugins = NewPluginRegistry()
	HardwarePlugins    = NewPluginRegistry()
	MetricsPlugins     = NewPluginRegistry()
	AnomalyPlugins     = NewPluginRegistry()
)
