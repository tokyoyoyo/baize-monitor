package request

// HardwareInfoUploadRequest 硬件信息上传请求
type HardwareInfoUploadRequest struct {
	// 收集时间戳
	CollectedAt string `json:"collected_at" binding:"required"` // 收集时间，ISO8601 格式

	// CPU 信息
	CPUs []CPURequest `json:"cpu" binding:"required"` // CPU 信息列表

	// 内存信息
	Memory *MemoryRequest `json:"memory" binding:"required"` // 内存总信息

	// 磁盘信息
	Disks []DiskRequest `json:"disks" binding:"required"` // 磁盘列表

	// 网络接口信息
	NetworkInterfaces []NetworkInterfaceRequest `json:"network_interfaces" binding:"required"` // 网络接口列表
}
