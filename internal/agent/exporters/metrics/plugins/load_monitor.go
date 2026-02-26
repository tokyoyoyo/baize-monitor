package plugins

import (
	"github.com/prometheus/client_golang/prometheus"
	"baize-monitor/internal/agent/plugins"
	"time"
)

// LoadMonitor 负载监控插件
type LoadMonitor struct {
	name              string
	description       string
	tool              string
	parameters        []string
	interval          time.Duration
	enabled           bool
	lastExecutionTime time.Time
	lastExecutionStatus plugins.ExecutionStatus
}

// NewLoadMonitor 创建负载监控插件
func NewLoadMonitor() *LoadMonitor {
	return &LoadMonitor{
		name:                "load_monitor",
		description:         "Monitor system load average",
		tool:                "uptime",
		parameters:          []string{},
		interval:            30 * time.Second,
		enabled:             true,
		lastExecutionStatus: plugins.StatusPending,
	}
}

// Name 获取插件名称
func (lm *LoadMonitor) Name() string {
	return lm.name
}

// Description 获取插件描述
func (lm *LoadMonitor) Description() string {
	return lm.description
}

// Tool 获取工具名称
func (lm *LoadMonitor) Tool() string {
	return lm.tool
}

// Parameters 获取参数
func (lm *LoadMonitor) Parameters() []string {
	return lm.parameters
}

// Interval 获取执行间隔
func (lm *LoadMonitor) Interval() time.Duration {
	return lm.interval
}

// Execute 执行插件
func (lm *LoadMonitor) Execute() (interface{}, error) {
	lm.lastExecutionStatus = plugins.StatusRunning
	lm.lastExecutionTime = time.Now()
	
	// 简化实现，实际应该解析uptime输出
	result := map[string]interface{}{
		"load_averages": map[string]float64{
			"1min":  1.2,
			"5min":  0.8,
			"15min": 0.6,
		},
		"collected_at": time.Now().Unix(),
	}
	
	lm.lastExecutionStatus = plugins.StatusSuccess
	return result, nil
}

// LastExecutionStatus 获取最后执行状态
func (lm *LoadMonitor) LastExecutionStatus() plugins.ExecutionStatus {
	return lm.lastExecutionStatus
}

// LastExecutionTime 获取最后执行时间
func (lm *LoadMonitor) LastExecutionTime() time.Time {
	return lm.lastExecutionTime
}

// Enabled 获取启用状态
func (lm *LoadMonitor) Enabled() bool {
	return lm.enabled
}

// SetEnabled 设置启用状态
func (lm *LoadMonitor) SetEnabled(enabled bool) {
	lm.enabled = enabled
}

// Describe 描述指标
func (lm *LoadMonitor) Describe(ch chan<- *prometheus.Desc) {
	// 插件特定指标描述
}

// Collect 收集指标
func (lm *LoadMonitor) Collect(ch chan<- prometheus.Metric) {
	// 插件特定指标收集
}