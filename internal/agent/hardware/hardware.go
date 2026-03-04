package hardware

import (
	"time"

	hardwareRequest "baize-monitor/pkg/dto/request/hardware"
	"baize-monitor/internal/agent/hardware/collectors"
)

// Hardware 硬件信息收集器
type Hardware struct {
	registry *collectors.Registry
	enabled  bool
}

// New 创建硬件信息收集器
func New() *Hardware {
	registry := collectors.NewRegistry()
	registry.AutoDiscover()
	return &Hardware{
		registry: registry,
		enabled:  true,
	}
}

// GetData 获取硬件信息数据（实现 dataProvider 接口）
func (h *Hardware) GetData() (interface{}, error) {
	if !h.enabled {
		return nil, nil
	}

	hardwareInfo := h.registry.CollectAll()
	
	// 设置收集时间戳
	hardwareInfo.CollectedAt = time.Now().Format(time.RFC3339)
	
	// 生成摘要信息
	hardwareInfo.Summary = h.generateSummary(hardwareInfo)

	return hardwareInfo, nil
}

// Start 启动（空实现，保持接口一致）
func (h *Hardware) Start() error {
	return nil
}

// Stop 停止（空实现，保持接口一致）
func (h *Hardware) Stop() error {
	return nil
}

// generateSummary 生成硬件摘要信息
func (h *Hardware) generateSummary(info *hardwareRequest.HardwareInfoUploadRequest) hardwareRequest.HardwareSummary {
	summary := hardwareRequest.HardwareSummary{}

	// CPU 摘要
	if len(info.CPUs) > 0 {
		summary.CPUModel = info.CPUs[0].Model
		totalCores := 0
		for _, cpu := range info.CPUs {
			totalCores += cpu.Cores
		}
		summary.CPUCores = totalCores
	}

	// 内存摘要
	if info.Memory != nil {
		summary.TotalMemory = info.Memory.TotalSize
		summary.MemorySlots = info.Memory.Slots
	}

	// 磁盘摘要
	summary.DiskCount = len(info.Disks)
	totalDisk := int64(0)
	for _, disk := range info.Disks {
		totalDisk += disk.Size
	}
	summary.TotalDisk = totalDisk

	// 网络摘要
	summary.NetworkCount = len(info.NetworkInterfaces)

	// 统计信息
	summary.ModuleCount = len(info.MemoryModules)
	summary.PartitionCount = len(info.Partitions)

	return summary
}
