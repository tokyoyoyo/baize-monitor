package registry

import (
	"github.com/prometheus/client_golang/prometheus"

	"baize-monitor/internal/agent/metrics/types"
)

// Registry 采集器注册表
type Registry struct {
	collectors []types.PrometheusMetricsCollector
}

// NewRegistry 创建采集器注册表
func NewRegistry() *Registry {
	return &Registry{
		collectors: make([]types.PrometheusMetricsCollector, 0),
	}
}

// Register 注册采集器
func (r *Registry) Register(collector types.PrometheusMetricsCollector) {
	r.collectors = append(r.collectors, collector)
}

// GetAll 获取所有采集器
func (r *Registry) GetAll() []types.PrometheusMetricsCollector {
	return r.collectors
}

// Describe 描述所有指标
func (r *Registry) Describe(ch chan<- *prometheus.Desc) {
	for _, collector := range r.collectors {
		collector.Describe(ch)
	}
}

// Collect 收集所有指标
func (r *Registry) Collect(ch chan<- prometheus.Metric) {
	for _, collector := range r.collectors {
		collector.Collect(ch)
	}
}
