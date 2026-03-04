package response

// MemoryResponse 内存信息响应
type MemoryResponse struct {
	// 物理内存规格
	TotalSize int64  `json:"total_size"` // 总容量（字节）
	Type      string `json:"type"`       // 内存类型（DDR4、DDR5等）
	Speed     string `json:"speed"`      // 内存频率
	Slots     int    `json:"slots"`      // 插槽总数
}

// MemoryModuleResponse 内存条信息响应
type MemoryModuleResponse struct {
	// 内存条规格
	Slot         string `json:"slot"`          // 插槽位置
	Size         int64  `json:"size"`          // 容量（字节）
	Type         string `json:"type"`          // 类型
	Speed        string `json:"speed"`         // 频率
	Manufacturer string `json:"manufacturer"`  // 制造商
	SerialNumber string `json:"serial_number"` // 序列号
	PartNumber   string `json:"part_number"`   // 零件号
}
