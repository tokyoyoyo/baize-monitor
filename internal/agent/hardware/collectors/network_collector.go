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
	networkRequest := hardwareRequest.NetworkInterfaceRequest{
		Content: make([]hardwareRequest.NetworkInterfaceInfo, 0),
		Summary: hardwareRequest.NetworkInterfaceSummary{},
	}

	// 获取网络接口信息
	interfaces, err := net.Interfaces()
	if err != nil {
		networkRequest.Success = false
		networkRequest.Message = fmt.Sprintf("failed to get network interfaces: %v", err)
		hardwareInfo.NetworkInterfaces = append(hardwareInfo.NetworkInterfaces, networkRequest)
		return
	}

	for _, iface := range interfaces {
		networkInterface := hardwareRequest.NetworkInterfaceInfo{
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
			networkRequest.Summary.VirtualCount++
		} else {
			networkRequest.Summary.PhysicalCount++
		}

		// 获取 IP 地址
		addrs, err := iface.Addrs()
		if err != nil {
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

		networkRequest.Content = append(networkRequest.Content, networkInterface)
		networkRequest.Summary.TotalCount++
	}

	// 如果采集成功但没有设置 Success 字段
	if !networkRequest.Success && networkRequest.Message == "" {
		networkRequest.Success = true
		networkRequest.Message = "collected successfully"
	}

	hardwareInfo.NetworkInterfaces = append(hardwareInfo.NetworkInterfaces, networkRequest)
}
