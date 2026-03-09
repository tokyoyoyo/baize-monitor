package types

import (
	hardwareRequest "baize-monitor/pkg/dto/request/hardware"
)

// HardwareCollector 硬件收集器接口
type HardwareCollector interface {
	// Collect 收集硬件信息
	Collect(hardwareInfo *hardwareRequest.HardwareInfoRequest)
}
