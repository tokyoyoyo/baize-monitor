package metrics

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"

	"baize-monitor/internal/agent/metrics/collectors"
	"baize-monitor/internal/agent/metrics/registry"
)

// Metrics 指标监控模块
type Metrics struct {
	collectorRegistry *registry.Registry
}

// New 创建指标监控模块
func New() *Metrics {
	registry := registry.NewRegistry()

	// 注册采集器
	registry.Register(collectors.NewMemoryCollector())
	registry.Register(collectors.NewCPUCollector())

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
