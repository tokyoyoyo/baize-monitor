package collectors

import (
	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusMetricsCollector Prometheus指标收集器接口
type PrometheusMetricsCollector interface {
	prometheus.Collector
}

// Registry 收集器注册表
type Registry struct {
	collectors []PrometheusMetricsCollector
}

// NewRegistry 创建收集器注册表
func NewRegistry() *Registry {
	return &Registry{
		collectors: make([]PrometheusMetricsCollector, 0),
	}
}

// Register 注册收集器
func (r *Registry) Register(collector PrometheusMetricsCollector) {
	r.collectors = append(r.collectors, collector)
}

// GetAll 获取所有收集器
func (r *Registry) GetAll() []PrometheusMetricsCollector {
	return r.collectors
}

// AutoDiscover 自动发现并注册收集器
func (r *Registry) AutoDiscover() {
	// 只注册内存收集器（简化版）
	r.Register(NewMemoryCollector())
	// 这里可以添加更多收集器
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