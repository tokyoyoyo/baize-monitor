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

// BaseExporter 基础Exporter结构
type BaseExporter struct {
	name        string
	description string
	plugins     []plugins.Plugin
}

// NewBaseExporter 创建基础Exporter
func NewBaseExporter(name, description string) *BaseExporter {
	return &BaseExporter{
		name:        name,
		description: description,
		plugins:     make([]plugins.Plugin, 0),
	}
}

// Name 获取Exporter名称
func (be *BaseExporter) Name() string {
	return be.name
}

// Description 获取Exporter描述
func (be *BaseExporter) Description() string {
	return be.description
}

// RegisterPlugin 注册插件
func (be *BaseExporter) RegisterPlugin(plugin plugins.Plugin) error {
	be.plugins = append(be.plugins, plugin)
	return nil
}

// GetPlugins 获取所有插件
func (be *BaseExporter) GetPlugins() []plugins.Plugin {
	return be.plugins
}
