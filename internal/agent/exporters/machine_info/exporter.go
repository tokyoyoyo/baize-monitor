package machine_info

import (
	"github.com/prometheus/client_golang/prometheus"
	"baize-monitor/internal/agent/plugins"
	"log"
)

// Exporter 机器信息Exporter
type Exporter struct {
	name            string
	description     string
	plugins         []plugins.Plugin
	hostnameDesc    *prometheus.Desc
	ipAddressDesc   *prometheus.Desc
	biosInfoDesc    *prometheus.Desc
	systemVendorDesc *prometheus.Desc
}

// NewExporter 创建机器信息Exporter
func NewExporter() *Exporter {
	return &Exporter{
		name:        "machine_info",
		description: "Machine information collector",
		plugins:     make([]plugins.Plugin, 0),
		hostnameDesc: prometheus.NewDesc(
			"baize_machine_hostname",
			"Machine hostname",
			[]string{"hostname"}, nil,
		),
		ipAddressDesc: prometheus.NewDesc(
			"baize_machine_ip_address",
			"Machine IP addresses",
			[]string{"interface", "ip"}, nil,
		),
		biosInfoDesc: prometheus.NewDesc(
			"baize_machine_bios_info",
			"Machine BIOS information",
			[]string{"vendor", "version", "date"}, nil,
		),
		systemVendorDesc: prometheus.NewDesc(
			"baize_machine_system_vendor",
			"Machine system vendor information",
			[]string{"manufacturer", "product_name", "serial_number"}, nil,
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
	log.Printf("Machine info exporter started")
	return nil
}

// Stop 停止Exporter
func (e *Exporter) Stop() error {
	log.Printf("Machine info exporter stopped")
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
	ch <- e.hostnameDesc
	ch <- e.ipAddressDesc
	ch <- e.biosInfoDesc
	ch <- e.systemVendorDesc
	
	// 描述插件指标
	for _, plugin := range e.plugins {
		plugin.Describe(ch)
	}
}

// Collect 收集指标
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	// 收集机器基本信息
	e.collectMachineInfo(ch)
	
	// 收集插件指标
	for _, plugin := range e.plugins {
		if plugin.Enabled() {
			plugin.Collect(ch)
		}
	}
}

// collectMachineInfo 收集机器基本信息
func (e *Exporter) collectMachineInfo(ch chan<- prometheus.Metric) {
	// 简化实现，实际应该获取真实的机器信息
	ch <- prometheus.MustNewConstMetric(
		e.hostnameDesc,
		prometheus.GaugeValue,
		1,
		"default-hostname",
	)
	
	ch <- prometheus.MustNewConstMetric(
		e.biosInfoDesc,
		prometheus.GaugeValue,
		1,
		"unknown-vendor",
		"unknown-version",
		"2024-01-01",
	)
	
	ch <- prometheus.MustNewConstMetric(
		e.systemVendorDesc,
		prometheus.GaugeValue,
		1,
		"unknown-manufacturer",
		"unknown-product",
		"unknown-serial",
	)
}