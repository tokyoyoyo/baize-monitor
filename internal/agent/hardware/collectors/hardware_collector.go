package collectors

import (
	hardwareRequest "baize-monitor/pkg/dto/request/hardware"
)

// HardwareCollector 硬件信息收集器接口
type HardwareCollector interface {
	// Collect 收集硬件信息并直接填充到 hardwareInfo
	// 采集过程中的异常信息应记录到对应部件的 Success 和 Message 字段，而不是返回 error 中断整个流程
	Collect(hardwareInfo *hardwareRequest.HardwareInfoUploadRequest)
}

// Registry 收集器注册表
type Registry struct {
	collectors []HardwareCollector
}

// NewRegistry 创建收集器注册表
func NewRegistry() *Registry {
	return &Registry{
		collectors: make([]HardwareCollector, 0),
	}
}

// Register 注册收集器
func (r *Registry) Register(collector HardwareCollector) {
	r.collectors = append(r.collectors, collector)
}

// GetAll 获取所有收集器
func (r *Registry) GetAll() []HardwareCollector {
	return r.collectors
}

// AutoDiscover 自动发现并注册收集器
func (r *Registry) AutoDiscover() {
	// 注册所有硬件收集器
	r.Register(NewCPUCollector())
	r.Register(NewMemoryCollector())
	r.Register(NewDiskCollector())
	r.Register(NewNetworkCollector())
}

// CollectAll 收集所有硬件信息
// 即使个别采集器失败，也会继续执行其他采集器，并将异常信息记录到对应部件的 Success 和 Message 字段
func (r *Registry) CollectAll() *hardwareRequest.HardwareInfoUploadRequest {
	hardwareInfo := &hardwareRequest.HardwareInfoUploadRequest{}

	for _, collector := range r.collectors {
		collector.Collect(hardwareInfo)
	}

	return hardwareInfo
}
