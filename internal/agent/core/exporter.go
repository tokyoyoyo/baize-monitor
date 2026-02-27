package core

import (
	"baize-monitor/internal/agent/plugins"

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

	// 插件管理
	RegisterPlugin(plugin plugins.Plugin) error
	GetPlugins() []plugins.Plugin
}

// ExecutionStatus 执行状态类型
type ExecutionStatus string

const (
	StatusPending  ExecutionStatus = "pending"
	StatusRunning  ExecutionStatus = "running"
	StatusSuccess  ExecutionStatus = "success"
	StatusFailed   ExecutionStatus = "failed"
	StatusDisabled ExecutionStatus = "disabled"
)
