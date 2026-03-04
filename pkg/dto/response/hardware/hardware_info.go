package response

// HardwareInfoResponse 硬件信息响应 - 用于直接传输硬件收集数据
type HardwareInfoResponse struct {
	// 详细硬件信息
	CPU               *CPUResponse               `json:"cpu"`                // CPU信息
	Memory            *MemoryResponse            `json:"memory"`             // 内存总信息
	MemoryModules     []MemoryModuleResponse     `json:"memory_modules"`     // 内存条列表
	Disks             []DiskResponse             `json:"disks"`              // 磁盘列表
	Partitions        []DiskPartitionResponse    `json:"partitions"`         // 分区列表
	NetworkInterfaces []NetworkInterfaceResponse `json:"network_interfaces"` // 网络接口列表

	// 硬件摘要信息（可选）
	Summary *HardwareSummaryResponse `json:"summary,omitempty"` // 硬件摘要

	// 数据版本和校验
	Version  string `json:"version"`  // 数据格式版本
	Checksum string `json:"checksum"` // 数据校验和
}
