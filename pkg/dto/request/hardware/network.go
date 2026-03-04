package request

// NetworkInterfaceRequest 网络接口信息请求
type NetworkInterfaceRequest struct {
	// 执行状态
	Success bool   `json:"success"` // 采集是否成功
	Message string `json:"message"` // 执行消息（失败时记录错误信息）

	Content []NetworkInterfaceInfo `json:"content"` // 网络接口信息列表
	Summary NetworkInterfaceSummary `json:"summary"` // 摘要信息
}

// NetworkInterfaceInfo 网络接口信息
type NetworkInterfaceInfo struct {
	// 执行状态
	Success bool   `json:"success"` // 采集是否成功
	Message string `json:"message"` // 执行消息（失败时记录错误信息）

	// 接口信息
	Name       string `json:"name" binding:"required"`        // 接口名称
	MACAddress string `json:"mac_address" binding:"required"` // MAC 地址
	IPAddress  string `json:"ip_address"`                     // IP 地址
	SubnetMask string `json:"subnet_mask"`                    // 子网掩码
	Gateway    string `json:"gateway"`                        // 网关

	// 硬件规格
	Speed     string `json:"speed"`      // 连接速度
	Duplex    string `json:"duplex"`     // 双工模式
	MTU       int    `json:"mtu"`        // 最大传输单元
	Type      string `json:"type"`       // 接口类型（ethernet、wireless 等）
	IsVirtual bool   `json:"is_virtual"` // 是否虚拟接口

	// 制造商信息
	Manufacturer string `json:"manufacturer"` // 制造商
	Driver       string `json:"driver"`       // 驱动程序
	PCIAddress   string `json:"pci_address"`  // PCI 地址
}

// NetworkInterfaceSummary 网络接口摘要信息
type NetworkInterfaceSummary struct {
	TotalCount   int `json:"total_count"`   // 接口总数
	PhysicalCount int `json:"physical_count"` // 物理接口数
	VirtualCount  int `json:"virtual_count"`  // 虚拟接口数
	ActiveCount   int `json:"active_count"`   // 活跃接口数
}
