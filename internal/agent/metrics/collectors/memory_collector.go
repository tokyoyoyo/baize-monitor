package collectors

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/mem"
)

// MemoryCollector 内存指标收集器
type MemoryCollector struct {
	usageDesc *prometheus.Desc
	totalDesc *prometheus.Desc
	usedDesc  *prometheus.Desc
	freeDesc  *prometheus.Desc
}

// NewMemoryCollector 创建内存指标收集器
func NewMemoryCollector() *MemoryCollector {
	return &MemoryCollector{
		usageDesc: prometheus.NewDesc(
			"baize_metrics_memory_usage_percent",
			"Memory usage percentage",
			nil, nil,
		),
		totalDesc: prometheus.NewDesc(
			"baize_metrics_memory_total_bytes",
			"Total memory in bytes",
			nil, nil,
		),
		usedDesc: prometheus.NewDesc(
			"baize_metrics_memory_used_bytes",
			"Used memory in bytes",
			nil, nil,
		),
		freeDesc: prometheus.NewDesc(
			"baize_metrics_memory_free_bytes",
			"Free memory in bytes",
			nil, nil,
		),
	}
}

// Describe 描述指标
func (m *MemoryCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- m.usageDesc
	ch <- m.totalDesc
	ch <- m.usedDesc
	ch <- m.freeDesc
}

// Collect 收集指标
func (m *MemoryCollector) Collect(ch chan<- prometheus.Metric) {
	// 使用 gopsutil 收集物理内存信息
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		// 内存使用率
		ch <- prometheus.MustNewConstMetric(
			m.usageDesc,
			prometheus.GaugeValue,
			memInfo.UsedPercent,
		)
		// 总内存
		ch <- prometheus.MustNewConstMetric(
			m.totalDesc,
			prometheus.GaugeValue,
			float64(memInfo.Total),
		)
		// 已用内存
		ch <- prometheus.MustNewConstMetric(
			m.usedDesc,
			prometheus.GaugeValue,
			float64(memInfo.Used),
		)
		// 空闲内存
		ch <- prometheus.MustNewConstMetric(
			m.freeDesc,
			prometheus.GaugeValue,
			float64(memInfo.Free),
		)
	}
}
