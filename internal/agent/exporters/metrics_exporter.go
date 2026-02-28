package exporters

import (
	"baize-monitor/internal/agent/core"
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

// MetricsExporter 指标监控Exporter
type MetricsExporter struct {
	name               string
	description        string
	plugins            []core.Plugin
	cpuUsageDesc       *prometheus.Desc
	memoryUsageDesc    *prometheus.Desc
	diskUsageDesc      *prometheus.Desc
	networkTrafficDesc *prometheus.Desc
	loadAverageDesc    *prometheus.Desc
}

// NewMetricsExporter 创建指标监控Exporter
func newMetricsExporter() *MetricsExporter {
	return &MetricsExporter{
		name:        "metrics",
		description: "System metrics collector",
		plugins:     make([]core.Plugin, 0),
		cpuUsageDesc: prometheus.NewDesc(
			"baize_metrics_cpu_usage_percent",
			"CPU usage percentage",
			nil, nil,
		),
		memoryUsageDesc: prometheus.NewDesc(
			"baize_metrics_memory_usage_percent",
			"Memory usage percentage",
			nil, nil,
		),
		diskUsageDesc: prometheus.NewDesc(
			"baize_metrics_disk_usage_percent",
			"Disk usage percentage",
			[]string{"mountpoint"}, nil,
		),
		networkTrafficDesc: prometheus.NewDesc(
			"baize_metrics_network_traffic_bytes_total",
			"Network traffic in bytes",
			[]string{"interface", "direction"}, nil,
		),
		loadAverageDesc: prometheus.NewDesc(
			"baize_metrics_load_average",
			"System load average",
			[]string{"period"}, nil,
		),
	}
}

// Name 获取MetricsExporter名称
func (e *MetricsExporter) Name() string {
	return e.name
}

// Description 获取MetricsExporter描述
func (e *MetricsExporter) Description() string {
	return e.description
}

// Start 启动Exporter
func (e *MetricsExporter) Start() error {
	log.Printf("Metrics exporter started")
	return nil
}

// Stop 停止Exporter
func (e *MetricsExporter) Stop() error {
	log.Printf("Metrics exporter stopped")
	return nil
}

// RegisterPlugin 注册插件
func (e *MetricsExporter) RegisterPlugin(plugin core.Plugin) error {
	e.plugins = append(e.plugins, plugin)
	return nil
}

// GetPlugins 获取所有插件
func (e *MetricsExporter) GetPlugins() []core.Plugin {
	return e.plugins
}

// Describe 描述指标
func (e *MetricsExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.cpuUsageDesc
	ch <- e.memoryUsageDesc
	ch <- e.diskUsageDesc
	ch <- e.networkTrafficDesc
	ch <- e.loadAverageDesc

	// 描述插件指标
	for _, plugin := range e.plugins {
		plugin.Describe(ch)
	}
}

// Collect 收集指标
func (e *MetricsExporter) Collect(ch chan<- prometheus.Metric) {
	// 收集系统指标
	e.collectSystemMetrics(ch)

	// 收集插件指标
	for _, plugin := range e.plugins {
		if plugin.Enabled() {
			plugin.Collect(ch)
		}
	}
}

// collectSystemMetrics 收集系统指标
func (e *MetricsExporter) collectSystemMetrics(ch chan<- prometheus.Metric) {
	// 简化实现，实际应该获取真实的系统指标
	ch <- prometheus.MustNewConstMetric(
		e.cpuUsageDesc,
		prometheus.GaugeValue,
		25.5,
	)

	ch <- prometheus.MustNewConstMetric(
		e.memoryUsageDesc,
		prometheus.GaugeValue,
		60.2,
	)

	ch <- prometheus.MustNewConstMetric(
		e.diskUsageDesc,
		prometheus.GaugeValue,
		75.8,
		"/",
	)

	ch <- prometheus.MustNewConstMetric(
		e.loadAverageDesc,
		prometheus.GaugeValue,
		1.2,
		"1min",
	)
}
