package anomaly

import (
	"baize-monitor/internal/agent/plugins" // 使用相对独立的插件包
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

// Exporter 异常检测Exporter
type Exporter struct {
	name                string
	description         string
	plugins             []plugins.Plugin
	diskHealthDesc      *prometheus.Desc
	temperatureDesc     *prometheus.Desc
	anomalyAlertDesc    *prometheus.Desc
	predictionScoreDesc *prometheus.Desc
}

// NewExporter 创建异常检测Exporter
func NewExporter() *Exporter {
	return &Exporter{
		name:        "anomaly",
		description: "Anomaly detection collector",
		plugins:     make([]plugins.Plugin, 0),
		diskHealthDesc: prometheus.NewDesc(
			"baize_anomaly_disk_health_score",
			"Disk health score (0-100)",
			[]string{"device", "type"}, nil,
		),
		temperatureDesc: prometheus.NewDesc(
			"baize_anomaly_temperature_celsius",
			"Hardware temperature in Celsius",
			[]string{"component"}, nil,
		),
		anomalyAlertDesc: prometheus.NewDesc(
			"baize_anomaly_alert_active",
			"Active anomaly alerts",
			[]string{"type", "severity", "component"}, nil,
		),
		predictionScoreDesc: prometheus.NewDesc(
			"baize_anomaly_prediction_score",
			"Anomaly prediction confidence score",
			[]string{"type", "component"}, nil,
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
	log.Printf("Anomaly detection exporter started")
	return nil
}

// Stop 停止Exporter
func (e *Exporter) Stop() error {
	log.Printf("Anomaly detection exporter stopped")
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
	ch <- e.diskHealthDesc
	ch <- e.temperatureDesc
	ch <- e.anomalyAlertDesc
	ch <- e.predictionScoreDesc

	// 描述插件指标
	for _, plugin := range e.plugins {
		plugin.Describe(ch)
	}
}

// Collect 收集指标
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	// 收集异常检测指标
	e.collectAnomalyMetrics(ch)

	// 收集插件指标
	for _, plugin := range e.plugins {
		if plugin.Enabled() {
			plugin.Collect(ch)
		}
	}
}

// collectAnomalyMetrics 收集异常检测指标
func (e *Exporter) collectAnomalyMetrics(ch chan<- prometheus.Metric) {
	// 简化实现，实际应该进行真正的异常检测
	ch <- prometheus.MustNewConstMetric(
		e.diskHealthDesc,
		prometheus.GaugeValue,
		95,
		"/dev/sda",
		"SSD",
	)

	ch <- prometheus.MustNewConstMetric(
		e.temperatureDesc,
		prometheus.GaugeValue,
		45,
		"cpu",
	)

	ch <- prometheus.MustNewConstMetric(
		e.anomalyAlertDesc,
		prometheus.GaugeValue,
		0,
		"disk_failure",
		"warning",
		"/dev/sda",
	)

	ch <- prometheus.MustNewConstMetric(
		e.predictionScoreDesc,
		prometheus.GaugeValue,
		0.1,
		"failure_prediction",
		"disk",
	)
}
