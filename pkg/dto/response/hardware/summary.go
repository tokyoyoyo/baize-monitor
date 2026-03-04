package response

// HardwareSummaryResponse 硬件信息摘要响应
type HardwareSummaryResponse struct {
	// CPU信息
	CPUModel string `json:"cpu_model"` // CPU型号
	CPUCores int    `json:"cpu_cores"` // CPU核心数

	// 内存信息
	TotalMemory int64 `json:"total_memory"` // 总内存（字节）
	MemorySlots int   `json:"memory_slots"` // 内存插槽数

	// 磁盘信息
	DiskCount int   `json:"disk_count"` // 磁盘数量
	TotalDisk int64 `json:"total_disk"` // 总磁盘容量（字节）

	// 网络信息
	NetworkCount int `json:"network_count"` // 网络接口数量
}
