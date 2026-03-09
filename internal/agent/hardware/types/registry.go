package types

import (
	hardwareRequest "baize-monitor/pkg/dto/request/hardware"
	"time"
)

// CollectorRegistry 收集器注册表
type CollectorRegistry struct {
	collectors []HardwareCollector
}

// NewCollectorRegistry 创建收集器注册表
func NewCollectorRegistry() *CollectorRegistry {
	return &CollectorRegistry{
		collectors: make([]HardwareCollector, 0),
	}
}

// Register 注册收集器
func (r *CollectorRegistry) Register(collector HardwareCollector) {
	r.collectors = append(r.collectors, collector)
}

// CollectAll 收集所有硬件信息
func (r *CollectorRegistry) CollectAll() *hardwareRequest.HardwareInfoRequest {
	hardwareInfo := &hardwareRequest.HardwareInfoRequest{}

	for _, collector := range r.collectors {
		collector.Collect(hardwareInfo)
	}

	return hardwareInfo
}

// CollectAllWithTimestamp 收集所有硬件信息并添加时间戳
func (r *CollectorRegistry) CollectAllWithTimestamp() *hardwareRequest.HardwareInfoRequest {
	hardwareInfo := r.CollectAll()
	hardwareInfo.CollectedAt = time.Now()
	return hardwareInfo
}
