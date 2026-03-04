package response

// NetworkInterfaceResponse 网络接口信息响应
type NetworkInterfaceResponse struct {
	// 接口信息
	Name       string `json:"name"`        // 接口名称
	MACAddress string `json:"mac_address"` // MAC地址
	IPAddress  string `json:"ip_address"`  // IP地址
	SubnetMask string `json:"subnet_mask"` // 子网掩码
	Gateway    string `json:"gateway"`     // 网关

	// 硬件规格
	Speed     string `json:"speed"`      // 连接速度
	Duplex    string `json:"duplex"`     // 双工模式
	MTU       int    `json:"mtu"`        // 最大传输单元
	Type      string `json:"type"`       // 接口类型（ethernet、wireless等）
	IsVirtual bool   `json:"is_virtual"` // 是否虚拟接口

	// 制造商信息
	Manufacturer string `json:"manufacturer"` // 制造商
	Driver       string `json:"driver"`       // 驱动程序
	PCIAddress   string `json:"pci_address"`  // PCI地址
}
