package collectors

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/cpu"
)

// CPUCollector CPU 指标收集器
type CPUCollector struct {
	usageDesc *prometheus.Desc
}

// NewCPUCollector 创建 CPU 指标收集器
func NewCPUCollector() *CPUCollector {
	return &CPUCollector{
		usageDesc: prometheus.NewDesc(
			"baize_metrics_cpu_usage_percent",
			"CPU usage percentage",
			nil, nil,
		),
	}
}

// Describe 描述指标
func (c *CPUCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.usageDesc
}

// Collect 收集指标
func (c *CPUCollector) Collect(ch chan<- prometheus.Metric) {
	// 获取 CPU 使用率（所有核心的平均值）
	percent, err := cpu.Percent(0, false)
	if err == nil && len(percent) > 0 {
		ch <- prometheus.MustNewConstMetric(
			c.usageDesc,
			prometheus.GaugeValue,
			percent[0],
		)
	}
}
