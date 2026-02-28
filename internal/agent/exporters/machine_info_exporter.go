package exporters

import (
	"baize-monitor/internal/agent/core"
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

// MachineInfoExporter 机器信息Exporter
type MachineInfoExporter struct {
	name             string
	description      string
	plugins          []core.Plugin
	hostnameDesc     *prometheus.Desc
	ipAddressDesc    *prometheus.Desc
	biosInfoDesc     *prometheus.Desc
	systemVendorDesc *prometheus.Desc
}

// NewMachineInfoExporter 创建机器信息Exporter
func newMachineInfoExporter() *MachineInfoExporter {
	return &MachineInfoExporter{
		name:        "machine_info",
		description: "Machine information collector",
		plugins:     make([]core.Plugin, 0),
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
func (e *MachineInfoExporter) Name() string {
	return e.name
}

// Description 获取Exporter描述
func (e *MachineInfoExporter) Description() string {
	return e.description
}

// Start 启动Exporter
func (e *MachineInfoExporter) Start() error {
	log.Printf("Machine info exporter started")
	return nil
}

// Stop 停止Exporter
func (e *MachineInfoExporter) Stop() error {
	log.Printf("Machine info exporter stopped")
	return nil
}

// RegisterPlugin 注册插件
func (e *MachineInfoExporter) RegisterPlugin(plugin core.Plugin) error {
	e.plugins = append(e.plugins, plugin)
	return nil
}

// GetPlugins 获取所有插件
func (e *MachineInfoExporter) GetPlugins() []core.Plugin {
	return e.plugins
}

// Describe 描述指标
func (e *MachineInfoExporter) Describe(ch chan<- *prometheus.Desc) {
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
func (e *MachineInfoExporter) Collect(ch chan<- prometheus.Metric) {
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
func (e *MachineInfoExporter) collectMachineInfo(ch chan<- prometheus.Metric) {
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
