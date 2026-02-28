package exporters

import (
	"baize-monitor/internal/agent/core"
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

// HardwareExporter 硬件信息Exporter
type HardwareExporter struct {
	name            string
	description     string
	plugins         []core.Plugin
	cpuInfoDesc     *prometheus.Desc
	memoryInfoDesc  *prometheus.Desc
	diskInfoDesc    *prometheus.Desc
	networkInfoDesc *prometheus.Desc
}

// 创建硬件信息Exporter
func newHardwareExporter() *HardwareExporter {
	return &HardwareExporter{
		name:        "hardware",
		description: "Hardware information collector",
		plugins:     make([]core.Plugin, 0),
		cpuInfoDesc: prometheus.NewDesc(
			"baize_hardware_cpu_info",
			"CPU hardware information",
			[]string{"model", "cores", "threads"}, nil,
		),
		memoryInfoDesc: prometheus.NewDesc(
			"baize_hardware_memory_info_bytes",
			"Memory hardware information",
			[]string{"type", "speed"}, nil,
		),
		diskInfoDesc: prometheus.NewDesc(
			"baize_hardware_disk_info_bytes",
			"Disk hardware information",
			[]string{"device", "model", "type"}, nil,
		),
		networkInfoDesc: prometheus.NewDesc(
			"baize_hardware_network_info",
			"Network hardware information",
			[]string{"interface", "mac_address", "speed"}, nil,
		),
	}
}

// Name 获取HardwareExporter名称
func (e *HardwareExporter) Name() string {
	return e.name
}

// Description 获取HardwareExporter描述
func (e *HardwareExporter) Description() string {
	return e.description
}

// Start 启动Exporter
func (e *HardwareExporter) Start() error {
	log.Printf("Hardware exporter started")
	return nil
}

// Stop 停止Exporter
func (e *HardwareExporter) Stop() error {
	log.Printf("Hardware exporter stopped")
	return nil
}

// RegisterPlugin 注册插件
func (e *HardwareExporter) RegisterPlugin(plugin core.Plugin) error {
	e.plugins = append(e.plugins, plugin)
	return nil
}

// GetPlugins 获取所有插件
func (e *HardwareExporter) GetPlugins() []core.Plugin {
	return e.plugins
}

// Describe 描述指标
func (e *HardwareExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.cpuInfoDesc
	ch <- e.memoryInfoDesc
	ch <- e.diskInfoDesc
	ch <- e.networkInfoDesc

	// 描述插件指标
	for _, plugin := range e.plugins {
		plugin.Describe(ch)
	}
}

// Collect 收集指标
func (e *HardwareExporter) Collect(ch chan<- prometheus.Metric) {
	// 收集硬件基本信息
	e.collectHardwareInfo(ch)

	// 收集插件指标
	for _, plugin := range e.plugins {
		if plugin.Enabled() {
			plugin.Collect(ch)
		}
	}
}

// collectHardwareInfo 收集硬件基本信息
func (e *HardwareExporter) collectHardwareInfo(ch chan<- prometheus.Metric) {
	// 简化实现，实际应该获取详细的硬件信息
	ch <- prometheus.MustNewConstMetric(
		e.cpuInfoDesc,
		prometheus.GaugeValue,
		1,
		"unknown-model",
		"unknown-cores",
		"unknown-threads",
	)

	ch <- prometheus.MustNewConstMetric(
		e.memoryInfoDesc,
		prometheus.GaugeValue,
		1,
		"unknown-type",
		"unknown-speed",
	)

	ch <- prometheus.MustNewConstMetric(
		e.diskInfoDesc,
		prometheus.GaugeValue,
		1,
		"unknown-device",
		"unknown-model",
		"unknown-type",
	)

	ch <- prometheus.MustNewConstMetric(
		e.networkInfoDesc,
		prometheus.GaugeValue,
		1,
		"unknown-interface",
		"unknown-mac",
		"unknown-speed",
	)
}
