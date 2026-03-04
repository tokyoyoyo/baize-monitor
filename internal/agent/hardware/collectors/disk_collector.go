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
func (d *DiskCollector) Collect(hardwareInfo *hardwareRequest.HardwareInfoUploadRequest) {
	// 根据操作系统选择不同的采集方式
	var diskErr error
	if runtime.GOOS == "linux" {
		// 在 Linux 上使用 lsblk 和 smartctl 获取详细信息
		diskErr = d.collectDisksLinux(hardwareInfo)
	} else {
		// 其他系统使用通用方法
		diskErr = d.collectDisksGeneric(hardwareInfo)
	}

	// 处理磁盘采集错误
	if diskErr != nil {
		// 如果已收集到磁盘信息，将失败信息添加到所有磁盘
		if len(hardwareInfo.Disks) > 0 {
			failureMsg := fmt.Sprintf("partial collection failed: %v", diskErr)
			for i := range hardwareInfo.Disks {
				hardwareInfo.Disks[i].Success = false
				hardwareInfo.Disks[i].Message = failureMsg
			}
		} else {
			// 没有收集到磁盘信息，创建专门的失败记录
			hardwareInfo.Disks = append(hardwareInfo.Disks, hardwareRequest.DiskCreateRequest{
				Success: false,
				Message: fmt.Sprintf("disk collection failed: %v", diskErr),
			})
		}
	} else if len(hardwareInfo.Disks) == 0 {
		// 采集成功但没有找到任何磁盘
		hardwareInfo.Disks = append(hardwareInfo.Disks, hardwareRequest.DiskCreateRequest{
			Success: false,
			Message: "no disk devices found",
		})
	}

	// 收集分区信息（所有系统通用）
	if err := d.collectPartitions(hardwareInfo); err != nil {
		// 分区采集失败，将失败信息添加到所有磁盘（如果存在），否则添加一个专门的记录
		if len(hardwareInfo.Disks) > 0 {
			partitionFailureMsg := fmt.Sprintf("partition collection failed: %v", err)
			for i := range hardwareInfo.Disks {
				hardwareInfo.Disks[i].Success = false
				hardwareInfo.Disks[i].Message = fmt.Sprintf("%s; %s", hardwareInfo.Disks[i].Message, partitionFailureMsg)
			}
		} else {
			// 没有磁盘记录时，添加一个失败记录
			hardwareInfo.Disks = append(hardwareInfo.Disks, hardwareRequest.DiskCreateRequest{
				Success: false,
				Message: fmt.Sprintf("partition collection failed: %v", err),
			})
		}
	}
}

// collectDisksLinux 在 Linux 上收集磁盘信息
func (d *DiskCollector) collectDisksLinux(hardwareInfo *hardwareRequest.HardwareInfoUploadRequest) error {
	// 使用 lsblk 获取磁盘列表和基本信息
	if err := d.collectLsblkInfo(hardwareInfo); err != nil {
		return err
	}

	// 使用 smartctl 获取更详细的信息
	for i := range hardwareInfo.Disks {
		d.enhanceWithSmartctl(&hardwareInfo.Disks[i], hardwareInfo)
	}
	return nil
}

// collectLsblkInfo 使用 lsblk 获取磁盘信息
func (d *DiskCollector) collectLsblkInfo(hardwareInfo *hardwareRequest.HardwareInfoUploadRequest) error {
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

		diskInfo := hardwareRequest.DiskCreateRequest{
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

		hardwareInfo.Disks = append(hardwareInfo.Disks, diskInfo)
	}
	return nil
}

// enhanceWithSmartctl 使用 smartctl 增强磁盘信息
func (d *DiskCollector) enhanceWithSmartctl(diskInfo *hardwareRequest.DiskCreateRequest, hardwareInfo *hardwareRequest.HardwareInfoUploadRequest) {
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
func (d *DiskCollector) collectDisksGeneric(hardwareInfo *hardwareRequest.HardwareInfoUploadRequest) error {
	// 获取磁盘 IO 统计信息来获取磁盘列表
	ioCounters, err := disk.IOCounters()
	if err != nil {
		return fmt.Errorf("failed to get disk io counters: %w", err)
	}

	// 收集磁盘信息
	for name := range ioCounters {
		diskInfo := hardwareRequest.DiskCreateRequest{
			Success:      true,
			Message:      "collected via gopsutil",
			Device:       "/dev/" + name,
			Model:        name,
			Type:         "Unknown",
			Size:         0,
			SerialNumber: "",
			Firmware:     "",
			RPM:          0,
			Manufacturer: "Unknown",
			Product:      "",
			Vendor:       "",
		}

		hardwareInfo.Disks = append(hardwareInfo.Disks, diskInfo)
	}
	return nil
}

// collectPartitions 收集分区信息
func (d *DiskCollector) collectPartitions(hardwareInfo *hardwareRequest.HardwareInfoUploadRequest) error {
	// 获取磁盘分区信息
	partitions, err := disk.Partitions(false)
	if err != nil {
		return fmt.Errorf("failed to get disk partitions: %w", err)
	}

	// 收集分区信息
	for _, partition := range partitions {
		diskPartition := hardwareRequest.DiskPartitionCreateRequest{
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

		hardwareInfo.Partitions = append(hardwareInfo.Partitions, diskPartition)
	}

	return nil
}
