package machine_info

import (
	"baize-monitor/pkg/dto/request"
	"baize-monitor/pkg/models"
	"net"
	"os"
	"time"
)

// MachineInfo 机器信息收集器
type MachineInfo struct {
	enabled bool
}

// NewMachineInfo 创建机器信息收集器
func NewMachineInfo() *MachineInfo {
	return &MachineInfo{
		enabled: true,
	}
}

// GetData 获取机器信息数据（实现 dataProvider 接口）
func (m *MachineInfo) GetData() (interface{}, error) {
	if !m.enabled {
		return nil, nil
	}

	// 创建数据实例，各个方法填入自己负责的字段
	info := &request.MachineInfoRequest{
		Success: true,
		Message: "success",
		Content: models.MachineInfo{},
	}

	// 填入主机名
	hostname, err := m.getHostname()
	if err != nil {
		info.Success = false
		info.Message = "Failed to get hostname: " + err.Error()
	}
	info.Content.Hostname = hostname

	// 填入 IP 地址信息
	ipAddresses, err := m.getIPAddresses()
	if err != nil {
		info.Success = false
		if info.Message != "success" {
			info.Message += "; Failed to get IP addresses: " + err.Error()
		} else {
			info.Message = "Failed to get IP addresses: " + err.Error()
		}
	}
	info.Content.IPAddresses = ipAddresses

	// 填入收集时间
	info.CollectedAt = time.Now()

	return info, nil
}

// Start 启动（空实现，保持接口一致）
func (m *MachineInfo) Start() error {
	return nil
}

// Stop 停止（空实现，保持接口一致）
func (m *MachineInfo) Stop() error {
	return nil
}

// getHostname 获取主机名
func (m *MachineInfo) getHostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown", err
	}
	return hostname, nil
}

// getIPAddresses 获取 IP 地址信息
func (m *MachineInfo) getIPAddresses() (map[string][]string, error) {
	result := make(map[string][]string)

	nics, err := net.Interfaces()
	if err != nil {
		return result, err
	}

	for _, nic := range nics {
		// 只收集启用的接口
		if nic.Flags&net.FlagUp == 0 {
			continue
		}

		var ips []string
		addrs, err := nic.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				ips = append(ips, ipnet.IP.String())
			}
		}

		if len(ips) > 0 {
			result[nic.Name] = ips
		}
	}

	return result, nil
}
