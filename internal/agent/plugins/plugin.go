package plugins

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

// Plugin 插件接口
type Plugin interface {
	// 基本信息
	Name() string
	Description() string
	Tool() string
	Parameters() []string
	Interval() time.Duration
	
	// 执行相关
	Execute() (interface{}, error)
	LastExecutionStatus() ExecutionStatus
	LastExecutionTime() time.Time
	Enabled() bool
	SetEnabled(bool)
	
	// Prometheus Collector接口
	Collect(ch chan<- prometheus.Metric)
	Describe(ch chan<- *prometheus.Desc)
}

// ExecutionStatus 执行状态类型
type ExecutionStatus string

const (
	StatusPending   ExecutionStatus = "pending"
	StatusRunning   ExecutionStatus = "running"
	StatusSuccess   ExecutionStatus = "success"
	StatusFailed    ExecutionStatus = "failed"
	StatusDisabled  ExecutionStatus = "disabled"
)