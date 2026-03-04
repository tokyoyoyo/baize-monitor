package collectors

import (
	"fmt"
	"net"
	"strings"

	hardwareRequest "baize-monitor/pkg/dto/request/hardware"
)

// NetworkCollector 网络接口信息收集器
type NetworkCollector struct{}

// NewNetworkCollector 创建网络接口信息收集器
func NewNetworkCollector() *NetworkCollector {
	return &NetworkCollector{}
}

// Collect 收集网络接口信息并填充到 hardwareInfo
func (n *NetworkCollector) Collect(hardwareInfo *hardwareRequest.HardwareInfoUploadRequest) {
	// 获取网络接口信息
	interfaces, err := net.Interfaces()
	if err != nil {
		// 采集失败，记录一个失败的网络接口条目
		hardwareInfo.NetworkInterfaces = append(hardwareInfo.NetworkInterfaces, hardwareRequest.NetworkInterfaceCreateRequest{
			Success: false,
			Message: fmt.Sprintf("failed to get network interfaces: %v", err),
		})
		return
	}

	for _, iface := range interfaces {
		networkInterface := hardwareRequest.NetworkInterfaceCreateRequest{
			Success:      true,
			Message:      "collected successfully",
			Name:         iface.Name,
			MACAddress:   iface.HardwareAddr.String(),
			IPAddress:    "",
			SubnetMask:   "",
			Gateway:      "",
			Speed:        "Unknown",
			Duplex:       "Unknown",
			MTU:          iface.MTU,
			Type:         "Unknown",
			IsVirtual:    false,
			Manufacturer: "Unknown",
			Driver:       "",
			PCIAddress:   "",
		}

		// 判断是否为虚拟接口
		if strings.Contains(iface.Name, "docker") ||
			strings.Contains(iface.Name, "veth") ||
			strings.Contains(iface.Name, "br-") {
			networkInterface.IsVirtual = true
		}

		// 获取 IP 地址
		addrs, err := iface.Addrs()
		if err != nil {
			networkInterface.Success = false
			networkInterface.Message = fmt.Sprintf("%s; failed to get addresses: %v", networkInterface.Message, err)
		} else {
			for _, addr := range addrs {
				if ipNet, ok := addr.(*net.IPNet); ok {
					if ipNet.IP.To4() != nil {
						networkInterface.IPAddress = ipNet.IP.String()
						networkInterface.SubnetMask = ipNet.Mask.String()
						break
					}
				}
			}
		}

		hardwareInfo.NetworkInterfaces = append(hardwareInfo.NetworkInterfaces, networkInterface)
	}
}
