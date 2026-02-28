package core

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Exporter 指标导出器接口
type Exporter interface {
	// 基本信息
	Name() string
	Description() string

	// Prometheus Collector接口
	Collect(ch chan<- prometheus.Metric)
	Describe(ch chan<- *prometheus.Desc)

	// 生命周期管理
	Start() error
	Stop() error

	GetPlugins() []Plugin
}

type ExporterRegistry struct {
	exporters map[string]Exporter
}

// NewExporterRegistry 创建Exporter注册表
func NewExporterRegistry() *ExporterRegistry {
	return &ExporterRegistry{
		exporters: make(map[string]Exporter),
	}
}

func (r *ExporterRegistry) Register(exporter Exporter) {
	r.exporters[exporter.Name()] = exporter
}

func (r *ExporterRegistry) GetAllExporters() []Exporter {
	var exporters []Exporter
	for _, exporter := range r.exporters {
		exporters = append(exporters, exporter)
	}
	return exporters
}

// Global exporter registry
var (
	GlobalExporterRegistry = NewExporterRegistry()
)
