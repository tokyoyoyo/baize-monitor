package collectors

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v3/disk"

	hardwareRequest "baize-monitor/pkg/dto/request/hardware"
)

// DiskCollector 磁盘信息收集器
type DiskCollector struct{}

// NewDiskCollector 创建磁盘信息收集器
func NewDiskCollector() *DiskCollector {
	return &DiskCollector{}
}

// Collect 收集磁盘信息并填充到 hardwareInfo
func (d *DiskCollector) Collect(hardwareInfo *hardwareRequest.HardwareInfoRequest) {
	diskRequest := hardwareRequest.DiskRequest{
		Content: make([]hardwareRequest.DiskInfo, 0),
		Summary: hardwareRequest.DiskSummary{},
	}

	// 根据操作系统选择不同的采集方式
	var diskErr error
	if runtime.GOOS == "linux" {
		diskErr = d.collectDisksLinux(&diskRequest)
	} else {
		diskErr = d.collectDisksGeneric(&diskRequest)
	}

	// 处理磁盘采集错误
	if diskErr != nil {
		diskRequest.Success = false
		diskRequest.Message = fmt.Sprintf("disk collection failed: %v", diskErr)
	}

	// 收集分区信息（所有系统通用）
	if partitionErr := d.collectPartitions(&diskRequest); partitionErr != nil {
		if diskRequest.Success {
			diskRequest.Success = false
			diskRequest.Message = fmt.Sprintf("%s; partition collection failed: %v", diskRequest.Message, partitionErr)
		}
	}

	// 如果采集成功但没有设置 Success 字段
	if !diskRequest.Success && diskRequest.Message == "" {
		diskRequest.Success = true
		diskRequest.Message = "collected successfully"
	}

	// 检查是否有磁盘
	if len(diskRequest.Content) == 0 && diskRequest.Message == "" {
		diskRequest.Success = false
		diskRequest.Message = "no disk devices found"
	}

	hardwareInfo.Disks = diskRequest
}

// collectDisksLinux 在 Linux 上收集磁盘信息
func (d *DiskCollector) collectDisksLinux(diskRequest *hardwareRequest.DiskRequest) error {
	// 使用 lsblk 获取磁盘列表和基本信息
	if err := d.collectLsblkInfo(diskRequest); err != nil {
		return err
	}

	// 使用 smartctl 获取更详细的信息
	for i := range diskRequest.Content {
		d.enhanceWithSmartctl(&diskRequest.Content[i])
	}
	return nil
}

// collectLsblkInfo 使用 lsblk 获取磁盘信息
func (d *DiskCollector) collectLsblkInfo(diskRequest *hardwareRequest.DiskRequest) error {
	// 使用 lsblk 获取 JSON 格式的输出
	cmd := exec.Command("lsblk", "-J", "-o", "NAME,SIZE,TYPE,MODEL,VENDOR,ROTA,SERIAL,WWN")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("lsblk failed: %w", err)
	}

	// 解析 JSON 输出
	var lsblkOutput struct {
		BlockDevices []struct {
			Name   string `json:"name"`
			Size   string `json:"size"`
			Type   string `json:"type"`
			Model  string `json:"model"`
			Vendor string `json:"vendor"`
			Rota   bool   `json:"rota"`
			Serial string `json:"serial"`
			WWN    string `json:"wwn"`
		} `json:"blockdevices"`
	}

	if err := json.Unmarshal(output, &lsblkOutput); err != nil {
		return fmt.Errorf("failed to parse lsblk output: %w", err)
	}

	for _, device := range lsblkOutput.BlockDevices {
		// 只处理磁盘类型（不是分区）
		if device.Type != "disk" {
			continue
		}

		diskInfo := hardwareRequest.DiskInfo{
			Success:      true,
			Message:      "collected via lsblk",
			Device:       "/dev/" + device.Name,
			Model:        device.Model,
			Type:         d.getDiskType(device.Rota),
			SerialNumber: device.Serial,
			Manufacturer: device.Vendor,
			Product:      device.Model,
			Vendor:       device.Vendor,
		}

		// 解析容量
		diskInfo.Size = d.parseSize(device.Size)

		diskRequest.Content = append(diskRequest.Content, diskInfo)

		// 更新摘要
		diskRequest.Summary.TotalCount++
		diskRequest.Summary.TotalSize += diskInfo.Size
		if diskInfo.Type == "SSD" {
			diskRequest.Summary.SSDCount++
		} else {
			diskRequest.Summary.HDDCount++
		}
	}
	return nil
}

// enhanceWithSmartctl 使用 smartctl 增强磁盘信息
func (d *DiskCollector) enhanceWithSmartctl(diskInfo *hardwareRequest.DiskInfo) {
	// 使用 smartctl 获取详细信息
	cmd := exec.Command("smartctl", "-i", diskInfo.Device)
	output, err := cmd.Output()
	if err != nil {
		// smartctl 可能失败（如没有权限或设备不支持），记录但不中断
		diskInfo.Message = fmt.Sprintf("%s; smartctl failed: %v", diskInfo.Message, err)
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	// 正则表达式
	modelFamilyRegex := regexp.MustCompile(`Model Family:\s*(.+)`)
	deviceModelRegex := regexp.MustCompile(`Device Model:\s*(.+)`)
	serialRegex := regexp.MustCompile(`Serial Number:\s*(.+)`)
	firmwareRegex := regexp.MustCompile(`Firmware Version:\s*(.+)`)
	userCapacityRegex := regexp.MustCompile(`User Capacity:\s*\[?([\d,]+)\s*bytes\]?`)
	rotationRateRegex := regexp.MustCompile(`\s*Rotation Rate:\s*(.+)`)

	for scanner.Scan() {
		line := scanner.Text()

		// 解析型号家族
		if matches := modelFamilyRegex.FindStringSubmatch(line); len(matches) > 1 {
			family := strings.TrimSpace(matches[1])
			if family != "" && diskInfo.Manufacturer == "" {
				diskInfo.Manufacturer = family
			}
		}

		// 解析设备型号
		if matches := deviceModelRegex.FindStringSubmatch(line); len(matches) > 1 {
			model := strings.TrimSpace(matches[1])
			if model != "" {
				diskInfo.Model = model
				diskInfo.Product = model
			}
		}

		// 解析序列号
		if matches := serialRegex.FindStringSubmatch(line); len(matches) > 1 {
			serial := strings.TrimSpace(matches[1])
			if serial != "" && diskInfo.SerialNumber == "" {
				diskInfo.SerialNumber = serial
			}
		}

		// 解析固件版本
		if matches := firmwareRegex.FindStringSubmatch(line); len(matches) > 1 {
			diskInfo.Firmware = strings.TrimSpace(matches[1])
		}

		// 解析用户容量
		if matches := userCapacityRegex.FindStringSubmatch(line); len(matches) > 1 {
			capacityStr := strings.ReplaceAll(matches[1], ",", "")
			if capacity, err := strconv.ParseInt(capacityStr, 10, 64); err == nil {
				diskInfo.Size = capacity
			}
		}

		// 解析转速
		if matches := rotationRateRegex.FindStringSubmatch(line); len(matches) > 1 {
			rotationRate := strings.TrimSpace(matches[1])
			if rotationRate == "Solid State Device" {
				diskInfo.Type = "SSD"
				diskInfo.RPM = 0
			} else if rpm, err := strconv.Atoi(strings.TrimSuffix(rotationRate, " rpm")); err == nil {
				diskInfo.RPM = rpm
				diskInfo.Type = "HDD"
			}
		}
	}

	diskInfo.Message = diskInfo.Message + "; enhanced via smartctl"
}

// getDiskType 根据旋转属性获取磁盘类型
func (d *DiskCollector) getDiskType(rota bool) string {
	if rota {
		return "HDD"
	}
	return "SSD"
}

// parseSize 解析容量字符串为字节数
func (d *DiskCollector) parseSize(sizeStr string) int64 {
	if sizeStr == "" {
		return 0
	}

	// 移除空格
	sizeStr = strings.TrimSpace(sizeStr)

	// 尝试直接解析数字
	if size, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
		return size
	}

	// 解析带单位的容量
	re := regexp.MustCompile(`^([\d.]+)\s*([KMGTPE]?i?B?)$`)
	matches := re.FindStringSubmatch(sizeStr)
	if len(matches) < 3 {
		return 0
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0
	}

	unit := strings.ToUpper(matches[2])
	multiplier := float64(1)

	switch {
	case strings.HasPrefix(unit, "K"):
		multiplier = 1024
	case strings.HasPrefix(unit, "M"):
		multiplier = 1024 * 1024
	case strings.HasPrefix(unit, "G"):
		multiplier = 1024 * 1024 * 1024
	case strings.HasPrefix(unit, "T"):
		multiplier = 1024 * 1024 * 1024 * 1024
	case strings.HasPrefix(unit, "P"):
		multiplier = 1024 * 1024 * 1024 * 1024 * 1024
	case strings.HasPrefix(unit, "E"):
		multiplier = 1024 * 1024 * 1024 * 1024 * 1024 * 1024
	}

	return int64(value * multiplier)
}

// collectDisksGeneric 通用方法收集磁盘信息
func (d *DiskCollector) collectDisksGeneric(diskRequest *hardwareRequest.DiskRequest) error {
	// 获取磁盘 IO 统计信息来获取磁盘列表
	ioCounters, err := disk.IOCounters()
	if err != nil {
		return fmt.Errorf("failed to get disk io counters: %w", err)
	}

	// 收集磁盘信息
	for name, ioCounter := range ioCounters {
		diskInfo := hardwareRequest.DiskInfo{
			Success:      true,
			Message:      "collected via gopsutil",
			Device:       "/dev/" + name,
			Model:        name,
			Type:         "Unknown",
			Size:         int64(ioCounter.ReadBytes + ioCounter.WriteBytes), // 估算值
			SerialNumber: "",
			Firmware:     "",
			RPM:          0,
			Manufacturer: "Unknown",
			Product:      "",
			Vendor:       "",
		}

		diskRequest.Content = append(diskRequest.Content, diskInfo)
		diskRequest.Summary.TotalCount++
		diskRequest.Summary.TotalSize += diskInfo.Size
	}
	return nil
}

// collectPartitions 收集分区信息并添加到对应的磁盘对象中
func (d *DiskCollector) collectPartitions(diskRequest *hardwareRequest.DiskRequest) error {
	// 获取磁盘分区信息
	partitions, err := disk.Partitions(false)
	if err != nil {
		return fmt.Errorf("failed to get disk partitions: %w", err)
	}

	// 收集分区信息
	for _, partition := range partitions {
		diskPartition := hardwareRequest.DiskPartitionInfo{
			Success:    true,
			Message:    "collected via gopsutil",
			Device:     partition.Device,
			DiskDevice: partition.Mountpoint,
			Size:       0, // 需要额外获取
			Type:       partition.Fstype,
			Filesystem: partition.Fstype,
			Label:      "",
			IsBootable: false,
		}

		// 获取分区使用率以计算大小
		usage, err := disk.Usage(partition.Mountpoint)
		if err == nil {
			diskPartition.Size = int64(usage.Total)
		} else {
			diskPartition.Message = fmt.Sprintf("collected via gopsutil; failed to get size: %v", err)
		}

		// 找到对应的磁盘并添加分区
		for i := range diskRequest.Content {
			if d.isPartitionOfDisk(partition.Device, diskRequest.Content[i].Device) {
				diskRequest.Content[i].Partitions = append(diskRequest.Content[i].Partitions, diskPartition)
				diskRequest.Summary.TotalPartitions++
				break
			}
		}
	}

	return nil
}

// isPartitionOfDisk 判断分区是否属于某个磁盘
func (d *DiskCollector) isPartitionOfDisk(partitionDevice, diskDevice string) bool {
	// 简化判断：如果分区设备名包含磁盘设备名，则认为属于该磁盘
	// 例如：/dev/sda1 属于 /dev/sda
	if partitionDevice == diskDevice {
		return true
	}

	// 去除 /dev/ 前缀进行比较
	partitionName := strings.TrimPrefix(partitionDevice, "/dev/")
	diskName := strings.TrimPrefix(diskDevice, "/dev/")

	return strings.HasPrefix(partitionName, diskName)
}
