package types

import (
	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusMetricsCollector Prometheus 指标采集器接口
type PrometheusMetricsCollector interface {
	prometheus.Collector
}
