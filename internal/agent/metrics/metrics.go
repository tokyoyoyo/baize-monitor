package metrics

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"

	"baize-monitor/internal/agent/metrics/collectors"
)

// Metrics 指标监控模块
type Metrics struct {
	collectorRegistry *collectors.Registry
}

// New 创建指标监控模块
func New() *Metrics {
	registry := collectors.NewRegistry()
	// 自动发现并注册收集器
	registry.AutoDiscover()

	return &Metrics{
		collectorRegistry: registry,
	}
}

// Start 启动 Metrics
func (m *Metrics) Start() error {
	log.Printf("Metrics module started")
	return nil
}

// Stop 停止 Metrics
func (m *Metrics) Stop() error {
	log.Printf("Metrics module stopped")
	return nil
}

// Describe 描述指标
func (m *Metrics) Describe(ch chan<- *prometheus.Desc) {
	m.collectorRegistry.Describe(ch)
}

// Collect 收集指标
func (m *Metrics) Collect(ch chan<- prometheus.Metric) {
	m.collectorRegistry.Collect(ch)
}
