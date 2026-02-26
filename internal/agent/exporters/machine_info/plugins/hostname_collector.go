package plugins

import (
	"github.com/prometheus/client_golang/prometheus"
	"baize-monitor/internal/agent/plugins"
	"os"
	"time"
)

// HostnameCollector 主机名收集插件
type HostnameCollector struct {
	name              string
	description       string
	tool              string
	parameters        []string
	interval          time.Duration
	enabled           bool
	lastExecutionTime time.Time
	lastExecutionStatus plugins.ExecutionStatus
}

// NewHostnameCollector 创建主机名收集插件
func NewHostnameCollector() *HostnameCollector {
	return &HostnameCollector{
		name:                "hostname_collector",
		description:         "Collect machine hostname information",
		tool:                "hostname",
		parameters:          []string{},
		interval:            60 * time.Second,
		enabled:             true,
		lastExecutionStatus: plugins.StatusPending,
	}
}

// Name 获取插件名称
func (hc *HostnameCollector) Name() string {
	return hc.name
}

// Description 获取插件描述
func (hc *HostnameCollector) Description() string {
	return hc.description
}

// Tool 获取工具名称
func (hc *HostnameCollector) Tool() string {
	return hc.tool
}

// Parameters 获取参数
func (hc *HostnameCollector) Parameters() []string {
	return hc.parameters
}

// Interval 获取执行间隔
func (hc *HostnameCollector) Interval() time.Duration {
	return hc.interval
}

// Execute 执行插件
func (hc *HostnameCollector) Execute() (interface{}, error) {
	hc.lastExecutionStatus = plugins.StatusRunning
	hc.lastExecutionTime = time.Now()
	
	hostname, err := os.Hostname()
	if err != nil {
		hc.lastExecutionStatus = plugins.StatusFailed
		return nil, err
	}
	
	result := map[string]interface{}{
		"hostname": hostname,
		"collected_at": time.Now().Unix(),
	}
	
	hc.lastExecutionStatus = plugins.StatusSuccess
	return result, nil
}

// LastExecutionStatus 获取最后执行状态
func (hc *HostnameCollector) LastExecutionStatus() plugins.ExecutionStatus {
	return hc.lastExecutionStatus
}

// LastExecutionTime 获取最后执行时间
func (hc *HostnameCollector) LastExecutionTime() time.Time {
	return hc.lastExecutionTime
}

// Enabled 获取启用状态
func (hc *HostnameCollector) Enabled() bool {
	return hc.enabled
}

// SetEnabled 设置启用状态
func (hc *HostnameCollector) SetEnabled(enabled bool) {
	hc.enabled = enabled
}

// Describe 描述指标
func (hc *HostnameCollector) Describe(ch chan<- *prometheus.Desc) {
	// 插件特定指标描述
}

// Collect 收集指标
func (hc *HostnameCollector) Collect(ch chan<- prometheus.Metric) {
	// 插件特定指标收集
	// 这里可以根据执行结果生成Prometheus指标
}