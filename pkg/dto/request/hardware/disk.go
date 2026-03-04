package request

// DiskRequest 磁盘信息请求
type DiskRequest struct {
	// 执行状态
	Success bool   `json:"success"` // 采集是否成功
	Message string `json:"message"` // 执行消息（失败时记录错误信息）

	Content []DiskInfo      `json:"content"` // 磁盘信息列表
	Summary DiskSummary     `json:"summary"` // 摘要信息
}

// DiskInfo 磁盘信息
type DiskInfo struct {
	// 执行状态
	Success bool   `json:"success"` // 采集是否成功
	Message string `json:"message"` // 执行消息（失败时记录错误信息）

	// 磁盘规格
	Device       string `json:"device" binding:"required"` // 设备路径（如/dev/sda）
	Model        string `json:"model" binding:"required"`  // 磁盘型号
	Type         string `json:"type" binding:"required"`   // 磁盘类型（SSD、HDD 等）
	Size         int64  `json:"size" binding:"required"`   // 总容量（字节）
	SerialNumber string `json:"serial_number"`             // 序列号
	Firmware     string `json:"firmware"`                  // 固件版本
	RPM          int    `json:"rpm"`                       // 转速（RPM，仅机械硬盘）

	// 制造商信息
	Manufacturer string `json:"manufacturer"` // 制造商
	Product      string `json:"product"`      // 产品系列
	Vendor       string `json:"vendor"`       // 供应商

	// 分区列表
	Partitions []DiskPartitionInfo `json:"partitions"` // 分区列表
}

// DiskPartitionInfo 磁盘分区信息
type DiskPartitionInfo struct {
	// 执行状态
	Success bool   `json:"success"` // 采集是否成功
	Message string `json:"message"` // 执行消息（失败时记录错误信息）

	// 分区信息
	Device     string `json:"device" binding:"required"`      // 分区设备路径
	DiskDevice string `json:"disk_device" binding:"required"` // 所属磁盘设备路径
	Size       int64  `json:"size" binding:"required"`        // 分区大小（字节）
	Type       string `json:"type"`                           // 分区类型
	Filesystem string `json:"filesystem"`                     // 文件系统
	Label      string `json:"label"`                          // 标签
	IsBootable bool   `json:"is_bootable"`                    // 是否可启动
}

// DiskSummary 磁盘摘要信息
type DiskSummary struct {
	TotalCount      int   `json:"total_count"`      // 磁盘总数
	TotalSize       int64 `json:"total_size"`       // 总容量（字节）
	SSDCount        int   `json:"ssd_count"`        // SSD 数量
	HDDCount        int   `json:"hdd_count"`        // HDD 数量
	TotalPartitions int   `json:"total_partitions"` // 分区总数
}
