package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"baize-monitor/internal/agent/plugins" // 使用相对独立的插件包
	"log"
)

// Exporter 指标监控Exporter
type Exporter struct {
	name               string
	description        string
	plugins            []plugins.Plugin
	cpuUsageDesc       *prometheus.Desc
	memoryUsageDesc    *prometheus.Desc
	diskUsageDesc      *prometheus.Desc
	networkTrafficDesc *prometheus.Desc
	loadAverageDesc    *prometheus.Desc
}

// NewExporter 创建指标监控Exporter
func NewExporter() *Exporter {
	return &Exporter{
		name:        "metrics",
		description: "System metrics collector",
		plugins:     make([]plugins.Plugin, 0),
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

// Name 获取Exporter名称
func (e *Exporter) Name() string {
	return e.name
}

// Description 获取Exporter描述
func (e *Exporter) Description() string {
	return e.description
}

// Start 启动Exporter
func (e *Exporter) Start() error {
	log.Printf("Metrics exporter started")
	return nil
}

// Stop 停止Exporter
func (e *Exporter) Stop() error {
	log.Printf("Metrics exporter stopped")
	return nil
}

// RegisterPlugin 注册插件
func (e *Exporter) RegisterPlugin(plugin plugins.Plugin) error {
	e.plugins = append(e.plugins, plugin)
	return nil
}

// GetPlugins 获取所有插件
func (e *Exporter) GetPlugins() []plugins.Plugin {
	return e.plugins
}

// Describe 描述指标
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
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
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
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
func (e *Exporter) collectSystemMetrics(ch chan<- prometheus.Metric) {
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