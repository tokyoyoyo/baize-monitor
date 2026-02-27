package plugins

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Plugin 插件接口
type Plugin interface {
	Name() string
	Description() string
	Tool() string
	Parameters() []string
	Interval() time.Duration
	Execute() (interface{}, error)
	LastExecutionStatus() ExecutionStatus
	LastExecutionTime() time.Time
	Enabled() bool
	SetEnabled(bool)

	// Prometheus Collector接口
	Collect(ch chan<- prometheus.Metric)
	Describe(ch chan<- *prometheus.Desc)
}

// ExecutionStatus 执行状态类型
type ExecutionStatus string

const (
	StatusPending  ExecutionStatus = "pending"
	StatusRunning  ExecutionStatus = "running"
	StatusSuccess  ExecutionStatus = "success"
	StatusFailed   ExecutionStatus = "failed"
	StatusDisabled ExecutionStatus = "disabled"
)

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

// GetPlugin 获取插件
func (pr *PluginRegistry) GetPlugin(name string) (Plugin, bool) {
	plugin, exists := pr.plugins[name]
	return plugin, exists
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
