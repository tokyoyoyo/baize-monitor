package response

// DiskResponse 磁盘信息响应
type DiskResponse struct {
	// 磁盘规格
	Device       string `json:"device"`        // 设备路径（如/dev/sda）
	Model        string `json:"model"`         // 磁盘型号
	Type         string `json:"type"`          // 磁盘类型（SSD、HDD等）
	Size         int64  `json:"size"`          // 总容量（字节）
	SerialNumber string `json:"serial_number"` // 序列号
	Firmware     string `json:"firmware"`      // 固件版本
	RPM          int    `json:"rpm"`           // 转速（RPM，仅机械硬盘）

	// 制造商信息
	Manufacturer string `json:"manufacturer"` // 制造商
	Product      string `json:"product"`      // 产品系列
	Vendor       string `json:"vendor"`       // 供应商
}

// DiskPartitionResponse 磁盘分区信息响应
type DiskPartitionResponse struct {
	// 分区信息
	Device     string `json:"device"`      // 分区设备路径
	DiskDevice string `json:"disk_device"` // 所属磁盘设备路径
	Size       int64  `json:"size"`        // 分区大小（字节）
	Type       string `json:"type"`        // 分区类型
	Filesystem string `json:"filesystem"`  // 文件系统
	Label      string `json:"label"`       // 标签
	IsBootable bool   `json:"is_bootable"` // 是否可启动
}
