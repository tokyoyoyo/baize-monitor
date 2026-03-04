package request

// MemoryRequest 内存信息请求
type MemoryRequest struct {
	// 执行状态
	Success bool   `json:"success"` // 采集是否成功
	Message string `json:"message"` // 执行消息（失败时记录错误信息）

	Content []MemoryModuleInfo `json:"content"` // 内存条信息列表
	Summary MemorySummary      `json:"summary"` // 摘要信息
}

// MemoryModuleInfo 内存条信息
type MemoryModuleInfo struct {
	// 执行状态
	Success bool   `json:"success"` // 采集是否成功
	Message string `json:"message"` // 执行消息（失败时记录错误信息）

	// 内存条规格
	Slot         string `json:"slot" binding:"required"`         // 插槽位置
	Size         int64  `json:"size" binding:"required"`         // 容量（字节）
	Type         string `json:"type" binding:"required"`         // 类型
	Speed        string `json:"speed" binding:"required"`        // 频率
	Manufacturer string `json:"manufacturer" binding:"required"` // 制造商
	SerialNumber string `json:"serial_number"`                   // 序列号
	PartNumber   string `json:"part_number"`                     // 零件号
}

// MemorySummary 内存摘要信息
type MemorySummary struct {
	TotalSize  int64  `json:"total_size"`  // 总容量（字节）
	Type       string `json:"type"`        // 内存类型（DDR4、DDR5 等）
	Speed      string `json:"speed"`       // 内存频率
	TotalSlots int    `json:"total_slots"` // 插槽总数
	UsedSlots  int    `json:"used_slots"`  // 已用插槽数
}
