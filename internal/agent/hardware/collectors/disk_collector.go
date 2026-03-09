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

type DiskCollector struct{}

func NewDiskCollector() *DiskCollector {
	return &DiskCollector{}
}

func (d *DiskCollector) Collect(hardwareInfo *hardwareRequest.HardwareInfoRequest) {
	diskRequest := hardwareRequest.DiskRequest{
		Content: make([]hardwareRequest.DiskInfo, 0),
		Summary: hardwareRequest.DiskSummary{},
	}

	var diskErr error
	if runtime.GOOS == "linux" {
		diskErr = d.collectDisksLinux(&diskRequest)
	} else {
		diskErr = d.collectDisksGeneric(&diskRequest)
	}

	if diskErr != nil {
		diskRequest.Success = false
		diskRequest.Message = fmt.Sprintf("disk collection failed: %v", diskErr)
	}

	if partitionErr := d.collectPartitions(&diskRequest); partitionErr != nil {
		if diskRequest.Success {
			diskRequest.Success = false
			diskRequest.Message = fmt.Sprintf("%s; partition collection failed: %v", diskRequest.Message, partitionErr)
		}
	}

	if !diskRequest.Success && diskRequest.Message == "" {
		diskRequest.Success = true
		diskRequest.Message = "collected successfully"
	}

	if len(diskRequest.Content) == 0 && diskRequest.Message == "" {
		diskRequest.Success = false
		diskRequest.Message = "no disk devices found"
	}

	hardwareInfo.Disks = diskRequest
}

func (d *DiskCollector) collectDisksLinux(diskRequest *hardwareRequest.DiskRequest) error {
	if err := d.collectLsblkInfo(diskRequest); err != nil {
		return err
	}

	for i := range diskRequest.Content {
		d.enhanceWithSmartctl(&diskRequest.Content[i])
	}
	return nil
}

func (d *DiskCollector) collectLsblkInfo(diskRequest *hardwareRequest.DiskRequest) error {
	cmd := exec.Command("lsblk", "-J", "-o", "NAME,SIZE,TYPE,MODEL,VENDOR,ROTA,SERIAL,WWN")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("lsblk failed: %w", err)
	}

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

		diskInfo.Size = d.parseSize(device.Size)

		diskRequest.Content = append(diskRequest.Content, diskInfo)

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

func (d *DiskCollector) enhanceWithSmartctl(diskInfo *hardwareRequest.DiskInfo) {
	cmd := exec.Command("smartctl", "-i", diskInfo.Device)
	output, err := cmd.Output()
	if err != nil {
		diskInfo.Message = fmt.Sprintf("%s; smartctl failed: %v", diskInfo.Message, err)
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	modelFamilyRegex := regexp.MustCompile(`Model Family:\s*(.+)`)
	deviceModelRegex := regexp.MustCompile(`Device Model:\s*(.+)`)
	serialRegex := regexp.MustCompile(`Serial Number:\s*(.+)`)
	firmwareRegex := regexp.MustCompile(`Firmware Version:\s*(.+)`)
	userCapacityRegex := regexp.MustCompile(`User Capacity:\s*\[?([\d,]+)\s*bytes\]?`)
	rotationRateRegex := regexp.MustCompile(`\s*Rotation Rate:\s*(.+)`)

	for scanner.Scan() {
		line := scanner.Text()

		if matches := modelFamilyRegex.FindStringSubmatch(line); len(matches) > 1 {
			family := strings.TrimSpace(matches[1])
			if family != "" && diskInfo.Manufacturer == "" {
				diskInfo.Manufacturer = family
			}
		}

		if matches := deviceModelRegex.FindStringSubmatch(line); len(matches) > 1 {
			model := strings.TrimSpace(matches[1])
			if model != "" {
				diskInfo.Model = model
				diskInfo.Product = model
			}
		}

		if matches := serialRegex.FindStringSubmatch(line); len(matches) > 1 {
			serial := strings.TrimSpace(matches[1])
			if serial != "" && diskInfo.SerialNumber == "" {
				diskInfo.SerialNumber = serial
			}
		}

		if matches := firmwareRegex.FindStringSubmatch(line); len(matches) > 1 {
			diskInfo.Firmware = strings.TrimSpace(matches[1])
		}

		if matches := userCapacityRegex.FindStringSubmatch(line); len(matches) > 1 {
			capacityStr := strings.ReplaceAll(matches[1], ",", "")
			if capacity, err := strconv.ParseInt(capacityStr, 10, 64); err == nil {
				diskInfo.Size = capacity
			}
		}

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

func (d *DiskCollector) getDiskType(rota bool) string {
	if rota {
		return "HDD"
	}
	return "SSD"
}

func (d *DiskCollector) parseSize(sizeStr string) int64 {
	if sizeStr == "" {
		return 0
	}

	sizeStr = strings.TrimSpace(sizeStr)

	if size, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
		return size
	}

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

func (d *DiskCollector) collectDisksGeneric(diskRequest *hardwareRequest.DiskRequest) error {
	ioCounters, err := disk.IOCounters()
	if err != nil {
		return fmt.Errorf("failed to get disk io counters: %w", err)
	}

	for name, ioCounter := range ioCounters {
		diskInfo := hardwareRequest.DiskInfo{
			Success:      true,
			Message:      "collected via gopsutil",
			Device:       "/dev/" + name,
			Model:        name,
			Type:         "Unknown",
			Size:         int64(ioCounter.ReadBytes + ioCounter.WriteBytes),
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

func (d *DiskCollector) collectPartitions(diskRequest *hardwareRequest.DiskRequest) error {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return fmt.Errorf("failed to get disk partitions: %w", err)
	}

	for _, partition := range partitions {
		diskPartition := hardwareRequest.DiskPartitionInfo{
			Success:    true,
			Message:    "collected via gopsutil",
			Device:     partition.Device,
			DiskDevice: partition.Mountpoint,
			Size:       0,
			Type:       partition.Fstype,
			Filesystem: partition.Fstype,
			Label:      "",
			IsBootable: false,
		}

		usage, err := disk.Usage(partition.Mountpoint)
		if err == nil {
			diskPartition.Size = int64(usage.Total)
		} else {
			diskPartition.Message = fmt.Sprintf("collected via gopsutil; failed to get size: %v", err)
		}

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

func (d *DiskCollector) isPartitionOfDisk(partitionDevice, diskDevice string) bool {
	if partitionDevice == diskDevice {
		return true
	}

	partitionName := strings.TrimPrefix(partitionDevice, "/dev/")
	diskName := strings.TrimPrefix(diskDevice, "/dev/")

	return strings.HasPrefix(partitionName, diskName)
}
