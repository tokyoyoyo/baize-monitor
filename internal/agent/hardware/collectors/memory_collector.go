package collectors

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v3/mem"

	hardwareRequest "baize-monitor/pkg/dto/request/hardware"
)

// MemoryCollector 内存信息收集器
type MemoryCollector struct{}

// NewMemoryCollector 创建内存信息收集器
func NewMemoryCollector() *MemoryCollector {
	return &MemoryCollector{}
}

// Collect 收集内存信息并填充到 hardwareInfo
func (m *MemoryCollector) Collect(hardwareInfo *hardwareRequest.HardwareInfoUploadRequest) {
	// 初始化内存信息
	hardwareInfo.Memory = &hardwareRequest.MemoryCreateRequest{}

	// 获取虚拟内存信息
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		hardwareInfo.Memory.Success = false
		hardwareInfo.Memory.Message = fmt.Sprintf("failed to get memory info: %v", err)
		return
	}

	// 设置总内存信息
	hardwareInfo.Memory.TotalSize = int64(memInfo.Total)

	// 根据操作系统选择不同的采集方式
	var moduleErr error
	if runtime.GOOS == "linux" {
		// 在 Linux 上使用 dmidecode 获取详细信息
		moduleErr = m.collectMemoryModulesLinux(hardwareInfo)
	} else {
		// 其他系统使用通用方法
		moduleErr = m.collectMemoryModulesGeneric(hardwareInfo)
	}

	// 处理内存模块采集错误
	if moduleErr != nil {
		// 如果已收集到内存模块信息，将失败信息记录到主内存信息中
		if len(hardwareInfo.MemoryModules) > 0 {
			hardwareInfo.Memory.Success = false
			hardwareInfo.Memory.Message = fmt.Sprintf("partial collection failed: %v", moduleErr)
			// 将失败信息也更新到所有内存模块
			for i := range hardwareInfo.MemoryModules {
				hardwareInfo.MemoryModules[i].Success = false
				hardwareInfo.MemoryModules[i].Message = fmt.Sprintf("%s; collection failed: %v", hardwareInfo.MemoryModules[i].Message, moduleErr)
			}
		} else {
			// 没有收集到内存模块信息，记录失败到主内存信息
			hardwareInfo.Memory.Success = false
			hardwareInfo.Memory.Message = fmt.Sprintf("memory module collection failed: %v", moduleErr)
		}
	}
}

// collectMemoryModulesLinux 在 Linux 上使用 dmidecode 收集内存条信息
func (m *MemoryCollector) collectMemoryModulesLinux(hardwareInfo *hardwareRequest.HardwareInfoUploadRequest) error {
	// 使用 dmidecode 获取内存信息
	cmd := exec.Command("dmidecode", "-t", "memory")
	output, err := cmd.Output()
	if err != nil {
		// dmidecode 需要 root 权限，如果没有权限则使用通用方法
		return fmt.Errorf("dmidecode failed: %v", err)
	}

	// 解析 dmidecode 输出
	modules := parseDmiDecodeMemory(string(output))

	// 统计信息
	var totalSlots int
	var memoryType string
	var memorySpeed string
	typeCount := make(map[string]int)
	speedCount := make(map[string]int)

	for i := range modules {
		// 设置内存条执行状态
		modules[i].Success = true
		modules[i].Message = "collected via dmidecode"
		hardwareInfo.MemoryModules = append(hardwareInfo.MemoryModules, modules[i])

		if modules[i].Size > 0 {
			totalSlots++
			typeCount[modules[i].Type]++
			speedCount[modules[i].Speed]++
		}
	}

	// 确定主要内存类型和速度（出现次数最多的）
	maxTypeCount := 0
	for t, count := range typeCount {
		if count > maxTypeCount {
			maxTypeCount = count
			memoryType = t
		}
	}

	maxSpeedCount := 0
	for s, count := range speedCount {
		if count > maxSpeedCount {
			maxSpeedCount = count
			memorySpeed = s
		}
	}

	if memoryType == "" {
		memoryType = "Unknown"
	}
	if memorySpeed == "" {
		memorySpeed = "Unknown"
	}

	hardwareInfo.Memory.Type = memoryType
	hardwareInfo.Memory.Speed = memorySpeed
	hardwareInfo.Memory.Slots = totalSlots
	hardwareInfo.Memory.Success = true
	hardwareInfo.Memory.Message = "collected via dmidecode"
	return nil
}

// parseDmiDecodeMemory 解析 dmidecode 内存输出
func parseDmiDecodeMemory(output string) []hardwareRequest.MemoryModuleCreateRequest {
	var modules []hardwareRequest.MemoryModuleCreateRequest
	var currentModule *hardwareRequest.MemoryModuleCreateRequest

	scanner := bufio.NewScanner(strings.NewReader(output))
	inMemoryDevice := false

	// 正则表达式
	slotRegex := regexp.MustCompile(`\s*Locator:\s*(.+)`)
	sizeRegex := regexp.MustCompile(`\s*Size:\s*(\d+)\s*MB`)
	typeRegex := regexp.MustCompile(`\s*Type:\s*(DDR\w*)`)
	speedRegex := regexp.MustCompile(`\s*Speed:\s*(\d+)\s*MT/s`)
	manufacturerRegex := regexp.MustCompile(`\s*Manufacturer:\s*(.+)`)
	serialRegex := regexp.MustCompile(`\s*Serial Number:\s*(.+)`)
	partNumberRegex := regexp.MustCompile(`\s*Part Number:\s*(.+)`)

	for scanner.Scan() {
		line := scanner.Text()

		// 检测内存设备段开始
		if strings.Contains(line, "Memory Device") && strings.Contains(line, "Handle") {
			if currentModule != nil {
				modules = append(modules, *currentModule)
			}
			currentModule = &hardwareRequest.MemoryModuleCreateRequest{
				Type:  "Unknown",
				Speed: "Unknown",
			}
			inMemoryDevice = true
			continue
		}

		if !inMemoryDevice || currentModule == nil {
			continue
		}

		// 解析插槽位置
		if matches := slotRegex.FindStringSubmatch(line); len(matches) > 1 {
			currentModule.Slot = strings.TrimSpace(matches[1])
		}

		// 解析容量
		if matches := sizeRegex.FindStringSubmatch(line); len(matches) > 1 {
			if sizeMB, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
				currentModule.Size = sizeMB * 1024 * 1024 // 转换为字节
			}
		}

		// 解析类型
		if matches := typeRegex.FindStringSubmatch(line); len(matches) > 1 {
			currentModule.Type = strings.TrimSpace(matches[1])
		}

		// 解析速度
		if matches := speedRegex.FindStringSubmatch(line); len(matches) > 1 {
			currentModule.Speed = matches[1] + " MT/s"
		}

		// 解析制造商
		if matches := manufacturerRegex.FindStringSubmatch(line); len(matches) > 1 {
			manufacturer := strings.TrimSpace(matches[1])
			// 过滤掉空值或 "Not Specified"
			if manufacturer != "" && manufacturer != "Not Specified" {
				currentModule.Manufacturer = manufacturer
			}
		}

		// 解析序列号
		if matches := serialRegex.FindStringSubmatch(line); len(matches) > 1 {
			serial := strings.TrimSpace(matches[1])
			if serial != "" && serial != "Not Specified" {
				currentModule.SerialNumber = serial
			}
		}

		// 解析零件号
		if matches := partNumberRegex.FindStringSubmatch(line); len(matches) > 1 {
			partNumber := strings.TrimSpace(matches[1])
			if partNumber != "" && partNumber != "Not Specified" {
				currentModule.PartNumber = partNumber
			}
		}
	}

	// 添加最后一个模块
	if currentModule != nil {
		modules = append(modules, *currentModule)
	}

	return modules
}

// collectMemoryModulesGeneric 通用方法收集内存条信息
func (m *MemoryCollector) collectMemoryModulesGeneric(hardwareInfo *hardwareRequest.HardwareInfoUploadRequest) error {
	// 尝试读取 /proc/meminfo 获取基本信息
	hardwareInfo.Memory.Type = "Unknown"
	hardwareInfo.Memory.Speed = "Unknown"
	hardwareInfo.Memory.Slots = 0
	hardwareInfo.Memory.Success = true
	hardwareInfo.Memory.Message = "collected via generic method (limited info)"
	return nil
}
