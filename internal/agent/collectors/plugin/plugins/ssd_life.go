package plugins

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/disk"
)

// SSDLifePlugin SSD寿命检测插件
type SSDLifePlugin struct {
	// 基础指标
	ssdHealth         *prometheus.Desc
	ssdTemperature    *prometheus.Desc
	ssdPowerOnHours   *prometheus.Desc
	ssdPowerCycleCount *prometheus.Desc
	
	// 寿命相关指标
	ssdWearLevelingCount *prometheus.Desc
	ssdReallocatedSectorCount *prometheus.Desc
	ssdPendingSectorCount *prometheus.Desc
	ssdUncorrectableSectorCount *prometheus.Desc
	
	// 性能指标
	ssdReadErrorRate  *prometheus.Desc
	ssdSeekErrorRate  *prometheus.Desc
	ssdSpinRetryCount *prometheus.Desc
	
	// 容量指标
	ssdTotalBytes     *prometheus.Desc
	ssdUsedBytes      *prometheus.Desc
	ssdAvailableBytes *prometheus.Desc
	
	// 告警指标
	ssdHealthStatus   *prometheus.Desc
	ssdWearPercentage *prometheus.Desc
	
	// 配置参数
	warningThreshold  int // 健康评分警告阈值 (0-100)
	criticalThreshold int // 健康评分危险阈值 (0-100)
	temperatureWarning float64 // 温度警告阈值 (摄氏度)
	
	checkInterval     time.Duration
	isInitialized     bool
	lastError         error
}

// SMARTAttribute SMART属性结构
type SMARTAttribute struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	CurrentValue int    `json:"current_value"`
	WorstValue  int     `json:"worst_value"`
	Threshold   int     `json:"threshold"`
	RawValue    int64   `json:"raw_value"`
	Status      string  `json:"status"` // OK, WARNING, CRITICAL
}

// SSDInfo SSD信息结构
type SSDInfo struct {
	Device        string            `json:"device"`
	Model         string            `json:"model"`
	SerialNumber  string            `json:"serial_number"`
	Firmware      string            `json:"firmware"`
	Capacity      uint64            `json:"capacity"`
	HealthScore   int               `json:"health_score"` // 0-100
	Temperature   float64           `json:"temperature"`
	PowerOnHours  int64             `json:"power_on_hours"`
	Attributes    []SMARTAttribute  `json:"attributes"`
}

// NewSSDLifePlugin 创建SSD寿命检测插件
func NewSSDLifePlugin() *SSDLifePlugin {
	const subsystem = "ssd"
	
	return &SSDLifePlugin{
		// 基础指标
		ssdHealth: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "health_score"),
			"SSD health score (0-100)",
			[]string{"device", "model", "serial"}, nil,
		),
		ssdTemperature: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "temperature_celsius"),
			"SSD temperature in Celsius",
			[]string{"device", "model", "serial"}, nil,
		),
		ssdPowerOnHours: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "power_on_hours_total"),
			"Total power on hours",
			[]string{"device", "model", "serial"}, nil,
		),
		ssdPowerCycleCount: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "power_cycle_count_total"),
			"Total power cycle count",
			[]string{"device", "model", "serial"}, nil,
		),
		
		// 寿命相关指标
		ssdWearLevelingCount: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "wear_leveling_count"),
			"Wear leveling count",
			[]string{"device", "model", "serial"}, nil,
		),
		ssdReallocatedSectorCount: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "reallocated_sector_count"),
			"Reallocated sector count",
			[]string{"device", "model", "serial"}, nil,
		),
		ssdPendingSectorCount: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "pending_sector_count"),
			"Pending sector count",
			[]string{"device", "model", "serial"}, nil,
		),
		ssdUncorrectableSectorCount: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "uncorrectable_sector_count"),
			"Uncorrectable sector count",
			[]string{"device", "model", "serial"}, nil,
		),
		
		// 性能指标
		ssdReadErrorRate: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "read_error_rate"),
			"Read error rate",
			[]string{"device", "model", "serial"}, nil,
		),
		ssdSeekErrorRate: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "seek_error_rate"),
			"Seek error rate",
			[]string{"device", "model", "serial"}, nil,
		),
		ssdSpinRetryCount: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "spin_retry_count"),
			"Spin retry count",
			[]string{"device", "model", "serial"}, nil,
		),
		
		// 容量指标
		ssdTotalBytes: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "total_bytes"),
			"Total SSD capacity in bytes",
			[]string{"device", "model", "serial"}, nil,
		),
		ssdUsedBytes: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "used_bytes"),
			"Used SSD space in bytes",
			[]string{"device", "model", "serial"}, nil,
		),
		ssdAvailableBytes: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "available_bytes"),
			"Available SSD space in bytes",
			[]string{"device", "model", "serial"}, nil,
		),
		
		// 告警指标
		ssdHealthStatus: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "health_status"),
			"SSD health status (0=normal, 1=warning, 2=critical)",
			[]string{"device", "model", "serial", "level"}, nil,
		),
		ssdWearPercentage: prometheus.NewDesc(
			prometheus.BuildFQName("baize", subsystem, "wear_percentage"),
			"SSD wear percentage",
			[]string{"device", "model", "serial"}, nil,
		),
		
		// 默认配置
		warningThreshold:  70,
		criticalThreshold: 40,
		temperatureWarning: 60.0,
		checkInterval:     60 * time.Second,
	}
}

// Name 实现Plugin接口
func (p *SSDLifePlugin) Name() string {
	return "ssd_life"
}

// Description 实现Plugin接口
func (p *SSDLifePlugin) Description() string {
	return "Monitors SSD health, lifespan and potential failure risks using SMART data"
}

// Interval 实现Plugin接口
func (p *SSDLifePlugin) Interval() time.Duration {
	return p.checkInterval
}

// Setup 实现Plugin接口
func (p *SSDLifePlugin) Setup() error {
	// 检查smartctl是否可用
	if !p.isSmartctlAvailable() {
		p.lastError = fmt.Errorf("smartctl not available, please install smartmontools")
		return p.lastError
	}
	
	p.isInitialized = true
	fmt.Println("SSD life plugin initialized successfully")
	return nil
}

// Teardown 实现Plugin接口
func (p *SSDLifePlugin) Teardown() error {
	p.isInitialized = false
	fmt.Println("SSD life plugin cleaned up")
	return nil
}

// Collect 实现Plugin接口
func (p *SSDLifePlugin) Collect(ch chan<- prometheus.Metric) error {
	if !p.isInitialized {
		return fmt.Errorf("plugin not initialized")
	}
	
	// 获取所有磁盘设备
	devices, err := p.getSSDDevices()
	if err != nil {
		p.lastError = fmt.Errorf("failed to get SSD devices: %w", err)
		return p.lastError
	}
	
	// 收集每个SSD的信息
	for _, device := range devices {
		if err := p.collectSSDInfo(device, ch); err != nil {
			fmt.Printf("Warning: failed to collect SSD info for %s: %v\n", device, err)
			continue
		}
	}
	
	p.lastError = nil
	return nil
}

// isSmartctlAvailable 检查smartctl是否可用
func (p *SSDLifePlugin) isSmartctlAvailable() bool {
	_, err := exec.LookPath("smartctl")
	return err == nil
}

// getSSDDevices 获取SSD设备列表
func (p *SSDLifePlugin) getSSDDevices() ([]string, error) {
	var devices []string
	
	// 方法1: 使用lsblk命令
	if lsblkDevices, err := p.getDevicesFromLsblk(); err == nil {
		devices = append(devices, lsblkDevices...)
	}
	
	// 方法2: 检查/dev目录下的常见SSD设备
	commonDevices := []string{"/dev/sda", "/dev/sdb", "/dev/sdc", "/dev/nvme0n1"}
	for _, dev := range commonDevices {
		if p.isBlockDevice(dev) && p.isSSD(dev) {
			devices = append(devices, dev)
		}
	}
	
	// 去重
	uniqueDevices := make(map[string]bool)
	var result []string
	for _, dev := range devices {
		if !uniqueDevices[dev] {
			uniqueDevices[dev] = true
			result = append(result, dev)
		}
	}
	
	return result, nil
}

// getDevicesFromLsblk 从lsblk获取设备列表
func (p *SSDLifePlugin) getDevicesFromLsblk() ([]string, error) {
	cmd := exec.Command("lsblk", "-d", "-o", "NAME,TYPE,RM", "-n")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	var devices []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.Fields(scanner.Text())
		if len(line) >= 3 && line[1] == "disk" && line[2] == "0" {
			device := "/dev/" + line[0]
			if p.isSSD(device) {
				devices = append(devices, device)
			}
		}
	}
	
	return devices, nil
}

// isBlockDevice 检查是否为块设备
func (p *SSDLifePlugin) isBlockDevice(device string) bool {
	_, err := os.Stat(device)
	return err == nil
}

// isSSD 检查是否为SSD
func (p *SSDLifePlugin) isSSD(device string) bool {
	// 使用smartctl检查设备类型
	cmd := exec.Command("smartctl", "-i", device)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	
	// 检查输出中是否包含SSD相关关键词
	outputStr := strings.ToLower(string(output))
	ssdKeywords := []string{"ssd", "nvme", "solid state"}
	
	for _, keyword := range ssdKeywords {
		if strings.Contains(outputStr, keyword) {
			return true
		}
	}
	
	return false
}

// collectSSDInfo 收集SSD信息
func (p *SSDLifePlugin) collectSSDInfo(device string, ch chan<- prometheus.Metric) error {
	// 获取基本信息
	info, err := p.getSSDInfo(device)
	if err != nil {
		return fmt.Errorf("failed to get SSD info: %w", err)
	}
	
	// 获取使用情况
	usage, err := p.getDiskUsage(device)
	if err != nil {
		fmt.Printf("Warning: failed to get disk usage for %s: %v\n", device, err)
	}
	
	// 发送指标
	labels := []string{info.Device, info.Model, info.SerialNumber}
	
	// 基础指标
	ch <- prometheus.MustNewConstMetric(p.ssdHealth, prometheus.GaugeValue, float64(info.HealthScore), labels...)
	ch <- prometheus.MustNewConstMetric(p.ssdTemperature, prometheus.GaugeValue, info.Temperature, labels...)
	ch <- prometheus.MustNewConstMetric(p.ssdPowerOnHours, prometheus.CounterValue, float64(info.PowerOnHours), labels...)
	
	// 容量指标
	if usage != nil {
		ch <- prometheus.MustNewConstMetric(p.ssdTotalBytes, prometheus.GaugeValue, float64(usage.Total), labels...)
		ch <- prometheus.MustNewConstMetric(p.ssdUsedBytes, prometheus.GaugeValue, float64(usage.Used), labels...)
		ch <- prometheus.MustNewConstMetric(p.ssdAvailableBytes, prometheus.GaugeValue, float64(usage.Free), labels...)
	}
	
	// SMART属性指标
	for _, attr := range info.Attributes {
		p.collectSMARTAttribute(attr, labels, ch)
	}
	
	// 健康状态指标
	p.collectHealthStatus(info, labels, ch)
	
	return nil
}

// getSSDInfo 获取SSD详细信息
func (p *SSDLifePlugin) getSSDInfo(device string) (*SSDInfo, error) {
	info := &SSDInfo{
		Device: device,
	}
	
	// 使用smartctl获取详细信息
	cmd := exec.Command("smartctl", "-a", device)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("smartctl failed: %w", err)
	}
	
	outputStr := string(output)
	
	// 解析基本信息
	info.Model = p.parseField(outputStr, "Device Model:")
	info.SerialNumber = p.parseField(outputStr, "Serial Number:")
	info.Firmware = p.parseField(outputStr, "Firmware Version:")
	
	// 解析容量
	if capacityStr := p.parseField(outputStr, "User Capacity:"); capacityStr != "" {
		re := regexp.MustCompile(`\[([0-9,]+) bytes\]`)
		if matches := re.FindStringSubmatch(capacityStr); len(matches) > 1 {
			capacity, _ := strconv.ParseUint(strings.ReplaceAll(matches[1], ",", ""), 10, 64)
			info.Capacity = capacity
		}
	}
	
	// 解析SMART属性
	info.Attributes = p.parseSMARTAttributes(outputStr)
	
	// 计算健康评分
	info.HealthScore = p.calculateHealthScore(info.Attributes)
	
	// 获取温度
	info.Temperature = p.parseTemperature(outputStr)
	
	// 获取通电时间
	info.PowerOnHours = p.parsePowerOnHours(outputStr)
	
	return info, nil
}

// parseField 解析字段值
func (p *SSDLifePlugin) parseField(output, fieldName string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, fieldName) {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

// parseSMARTAttributes 解析SMART属性
func (p *SSDLifePlugin) parseSMARTAttributes(output string) []SMARTAttribute {
	var attributes []SMARTAttribute
	
	lines := strings.Split(output, "\n")
	re := regexp.MustCompile(`^\s*(\d+)\s+(.*?)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)`)
	
	for _, line := range lines {
		if matches := re.FindStringSubmatch(line); len(matches) >= 7 {
			id, _ := strconv.Atoi(matches[1])
			current, _ := strconv.Atoi(matches[3])
			worst, _ := strconv.Atoi(matches[4])
			threshold, _ := strconv.Atoi(matches[5])
			raw, _ := strconv.ParseInt(matches[6], 10, 64)
			
			attr := SMARTAttribute{
				ID:          id,
				Name:        strings.TrimSpace(matches[2]),
				CurrentValue: current,
				WorstValue:  worst,
				Threshold:   threshold,
				RawValue:    raw,
				Status:      p.getAttributeStatus(current, threshold),
			}
			
			attributes = append(attributes, attr)
		}
	}
	
	return attributes
}

// getAttributeStatus 获取属性状态
func (p *SSDLifePlugin) getAttributeStatus(current, threshold int) string {
	if current <= threshold {
		return "CRITICAL"
	} else if current <= threshold+10 {
		return "WARNING"
	}
	return "OK"
}

// calculateHealthScore 计算健康评分
func (p *SSDLifePlugin) calculateHealthScore(attributes []SMARTAttribute) int {
	score := 100
	
	for _, attr := range attributes {
		switch attr.ID {
		case 5: // Reallocated Sector Count
			if attr.RawValue > 0 {
				score -= int(attr.RawValue) * 5
			}
		case 184: // End-to-End Error
			if attr.CurrentValue < 100 {
				score -= (100 - attr.CurrentValue) * 2
			}
		case 187: // Reported Uncorrectable Errors
			if attr.RawValue > 0 {
				score -= int(attr.RawValue) * 10
			}
		case 197: // Current Pending Sector Count
			if attr.RawValue > 0 {
				score -= int(attr.RawValue) * 3
			}
		case 198: // Offline Uncorrectable
			if attr.RawValue > 0 {
				score -= int(attr.RawValue) * 8
			}
		}
	}
	
	// 确保评分在0-100范围内
	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}
	
	return score
}

// parseTemperature 解析温度
func (p *SSDLifePlugin) parseTemperature(output string) float64 {
	re := regexp.MustCompile(`Temperature:\s+([0-9.]+)`)
	if matches := re.FindStringSubmatch(output); len(matches) > 1 {
		if temp, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return temp
		}
	}
	return 0.0
}

// parsePowerOnHours 解析通电时间
func (p *SSDLifePlugin) parsePowerOnHours(output string) int64 {
	re := regexp.MustCompile(`Power_On_Hours\s+\d+\s+\d+\s+\d+\s+(\d+)`)
	if matches := re.FindStringSubmatch(output); len(matches) > 1 {
		if hours, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
			return hours
		}
	}
	return 0
}

// getDiskUsage 获取磁盘使用情况
func (p *SSDLifePlugin) getDiskUsage(device string) (*disk.UsageStat, error) {
	// 获取挂载点
	mountPoint := p.getMountPoint(device)
	if mountPoint == "" {
		return nil, fmt.Errorf("no mount point found for device %s", device)
	}
	
	return disk.Usage(mountPoint)
}

// getMountPoint 获取设备挂载点
func (p *SSDLifePlugin) getMountPoint(device string) string {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return ""
	}
	
	for _, partition := range partitions {
		if strings.HasPrefix(partition.Device, device) {
			return partition.Mountpoint
		}
	}
	
	return ""
}

// collectSMARTAttribute 收集SMART属性指标
func (p *SSDLifePlugin) collectSMARTAttribute(attr SMARTAttribute, labels []string, ch chan<- prometheus.Metric) {
	switch attr.ID {
	case 5: // Reallocated Sector Count
		ch <- prometheus.MustNewConstMetric(
			p.ssdReallocatedSectorCount,
			prometheus.GaugeValue,
			float64(attr.RawValue),
			labels...,
		)
	case 184: // End-to-End Error
		ch <- prometheus.MustNewConstMetric(
			p.ssdReadErrorRate,
			prometheus.GaugeValue,
			float64(attr.CurrentValue),
			labels...,
		)
	case 187: // Reported Uncorrectable Errors
		ch <- prometheus.MustNewConstMetric(
			p.ssdUncorrectableSectorCount,
			prometheus.GaugeValue,
			float64(attr.RawValue),
			labels...,
		)
	case 194: // Temperature
		// 已经单独处理
	case 197: // Current Pending Sector Count
		ch <- prometheus.MustNewConstMetric(
			p.ssdPendingSectorCount,
			prometheus.GaugeValue,
			float64(attr.RawValue),
			labels...,
		)
	case 233: // Wear Leveling Count (Intel SSDs)
		ch <- prometheus.MustNewConstMetric(
			p.ssdWearLevelingCount,
			prometheus.GaugeValue,
			float64(attr.RawValue),
			labels...,
		)
	}
}

// collectHealthStatus 收集健康状态
func (p *SSDLifePlugin) collectHealthStatus(info *SSDInfo, labels []string, ch chan<- prometheus.Metric) {
	var statusLevel string
	var statusValue float64
	
	switch {
	case info.HealthScore <= p.criticalThreshold:
		statusLevel = "critical"
		statusValue = 2
		fmt.Printf("CRITICAL: SSD %s health score (%d) below critical threshold (%d)\n",
			info.Device, info.HealthScore, p.criticalThreshold)
	case info.HealthScore <= p.warningThreshold:
		statusLevel = "warning"
		statusValue = 1
		fmt.Printf("WARNING: SSD %s health score (%d) below warning threshold (%d)\n",
			info.Device, info.HealthScore, p.warningThreshold)
	default:
		statusLevel = "normal"
		statusValue = 0
	}
	
	// 发送健康状态指标
	ch <- prometheus.MustNewConstMetric(
		p.ssdHealthStatus,
		prometheus.GaugeValue,
		statusValue,
		append(labels, statusLevel)...,
	)
	
	// 发送磨损百分比（如果有相关属性）
	wearPercent := p.calculateWearPercentage(info.Attributes)
	ch <- prometheus.MustNewConstMetric(
		p.ssdWearPercentage,
		prometheus.GaugeValue,
		wearPercent,
		labels...,
	)
	
	// 温度警告
	if info.Temperature >= p.temperatureWarning {
		fmt.Printf("WARNING: SSD %s temperature (%.1f°C) exceeds warning threshold (%.1f°C)\n",
			info.Device, info.Temperature, p.temperatureWarning)
	}
}

// calculateWearPercentage 计算磨损百分比
func (p *SSDLifePlugin) calculateWearPercentage(attributes []SMARTAttribute) float64 {
	for _, attr := range attributes {
		// Intel SSD磨损等级属性
		if attr.ID == 233 {
			// Intel SSD的磨损等级通常是0-100，其中100表示全新
			return 100.0 - float64(attr.CurrentValue)
		}
		// Samsung SSD磨损等级属性
		if attr.ID == 177 {
			return float64(attr.CurrentValue)
		}
	}
	return 0.0
}

// GetWarningThreshold 获取警告阈值
func (p *SSDLifePlugin) GetWarningThreshold() int {
	return p.warningThreshold
}

// SetWarningThreshold 设置警告阈值
func (p *SSDLifePlugin) SetWarningThreshold(threshold int) {
	p.warningThreshold = threshold
}

// GetCriticalThreshold 获取危险阈值
func (p *SSDLifePlugin) GetCriticalThreshold() int {
	return p.criticalThreshold
}

// SetCriticalThreshold 设置危险阈值
func (p *SSDLifePlugin) SetCriticalThreshold(threshold int) {
	p.criticalThreshold = threshold
}

// GetLastError 获取最后错误
func (p *SSDLifePlugin) GetLastError() error {
	return p.lastError
}