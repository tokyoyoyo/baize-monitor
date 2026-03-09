package collectors

import (
	"fmt"
	"net"
	"strings"

	hardwareRequest "baize-monitor/pkg/dto/request/hardware"
)

type NetworkCollector struct{}

func NewNetworkCollector() *NetworkCollector {
	return &NetworkCollector{}
}

func (n *NetworkCollector) Collect(hardwareInfo *hardwareRequest.HardwareInfoRequest) {
	networkRequest := hardwareRequest.NetworkInterfaceRequest{
		Content: make([]hardwareRequest.NetworkInterfaceInfo, 0),
		Summary: hardwareRequest.NetworkInterfaceSummary{},
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		networkRequest.Success = false
		networkRequest.Message = fmt.Sprintf("failed to get network interfaces: %v", err)
		hardwareInfo.NetworkInterfaces = networkRequest
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

		if strings.Contains(iface.Name, "docker") ||
			strings.Contains(iface.Name, "veth") ||
			strings.Contains(iface.Name, "br-") {
			networkInterface.IsVirtual = true
			networkRequest.Summary.VirtualCount++
		} else {
			networkRequest.Summary.PhysicalCount++
		}

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

	if !networkRequest.Success && networkRequest.Message == "" {
		networkRequest.Success = true
		networkRequest.Message = "collected successfully"
	}

	hardwareInfo.NetworkInterfaces = networkRequest
}
