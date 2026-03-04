package collectors

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v3/cpu"

	hardwareRequest "baize-monitor/pkg/dto/request/hardware"
)

// CPUCollector CPU 信息收集器
type CPUCollector struct{}

// NewCPUCollector 创建 CPU 信息收集器
func NewCPUCollector() *CPUCollector {
	return &CPUCollector{}
}

// Collect 收集 CPU 信息并填充到 hardwareInfo
func (c *CPUCollector) Collect(hardwareInfo *hardwareRequest.HardwareInfoUploadRequest) {
	// 根据操作系统选择不同的采集方式
	if runtime.GOOS == "linux" {
		// 在 Linux 上优先使用 dmidecode 获取详细信息
		dmidecodeErr := c.collectCPULinux(hardwareInfo)
		if dmidecodeErr != nil {
			// 如果 dmidecode 失败，回退到 gopsutil
			gopsutilErr := c.collectCPUGeneric(hardwareInfo)
			if gopsutilErr != nil {
				// 两者都失败，检查是否已收集到 CPU 信息
				if len(hardwareInfo.CPUs) > 0 {
					// 已收集到 CPU 信息，将失败信息添加到所有 CPU
					failureMsg := fmt.Sprintf("partial collection failed: dmidecode failed(%v), gopsutil also failed(%v)", dmidecodeErr, gopsutilErr)
					for i := range hardwareInfo.CPUs {
						hardwareInfo.CPUs[i].Success = false
						hardwareInfo.CPUs[i].Message = failureMsg
					}
				} else {
					// 没有收集到 CPU 信息，创建专门的失败记录
					hardwareInfo.CPUs = append(hardwareInfo.CPUs, hardwareRequest.CPUCreateRequest{
						Success: false,
						Message: fmt.Sprintf("failed to collect CPU info: dmidecode failed(%v), gopsutil also failed(%v)", dmidecodeErr, gopsutilErr),
					})
				}
			}
		}
		return
	}

	// 其他系统使用通用方法
	if err := c.collectCPUGeneric(hardwareInfo); err != nil {
		// 采集失败，检查是否已收集到 CPU 信息
		if len(hardwareInfo.CPUs) > 0 {
			// 已收集到 CPU 信息，将失败信息添加到所有 CPU
			failureMsg := fmt.Sprintf("partial collection failed: %v", err)
			for i := range hardwareInfo.CPUs {
				hardwareInfo.CPUs[i].Success = false
				hardwareInfo.CPUs[i].Message = failureMsg
			}
		} else {
			// 没有收集到 CPU 信息，创建专门的失败记录
			hardwareInfo.CPUs = append(hardwareInfo.CPUs, hardwareRequest.CPUCreateRequest{
				Success: false,
				Message: fmt.Sprintf("failed to collect CPU info: %v", err),
			})
		}
	}
}

// collectCPULinux 在 Linux 上使用 dmidecode 收集 CPU 信息
func (c *CPUCollector) collectCPULinux(hardwareInfo *hardwareRequest.HardwareInfoUploadRequest) error {
	// 使用 dmidecode 获取处理器信息
	cmd := exec.Command("dmidecode", "-t", "processor")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("dmidecode failed: %w", err)
	}

	// 解析 dmidecode 输出
	cpus := parseDmiDecodeCPU(string(output))

	if len(cpus) == 0 {
		return fmt.Errorf("no CPU info from dmidecode")
	}

	// 设置成功状态
	for i := range cpus {
		cpus[i].Success = true
		cpus[i].Message = "collected via dmidecode"
	}

	hardwareInfo.CPUs = cpus
	return nil
}

// parseDmiDecodeCPU 解析 dmidecode CPU 输出
func parseDmiDecodeCPU(output string) []hardwareRequest.CPUCreateRequest {
	var cpus []hardwareRequest.CPUCreateRequest
	var currentCPU *hardwareRequest.CPUCreateRequest

	scanner := bufio.NewScanner(strings.NewReader(output))
	inProcessor := false

	// 正则表达式
	socketRegex := regexp.MustCompile(`\s*Socket Designation:\s*(.+)`)
	typeRegex := regexp.MustCompile(`\s*Type:\s*(.+)`)
	familyRegex := regexp.MustCompile(`\s*Family:\s*(.+)`)
	manufacturerRegex := regexp.MustCompile(`\s*Manufacturer:\s*(.+)`)
	idRegex := regexp.MustCompile(`\s*ID:\s*(.+)`)
	signatureRegex := regexp.MustCompile(`\s*Signature:\s*(.+)`)
	flagsRegex := regexp.MustCompile(`\s*Flags:\s*(.+)`)
	versionRegex := regexp.MustCompile(`\s*Version:\s*(.+)`)
	maxSpeedRegex := regexp.MustCompile(`\s*Max Speed:\s*(.+)`)
	currentSpeedRegex := regexp.MustCompile(`\s*Current Speed:\s*(.+)`)
	coreCountRegex := regexp.MustCompile(`\s*Core Count:\s*(\d+)`)
	coreEnabledRegex := regexp.MustCompile(`\s*Core Enabled:\s*(\d+)`)
	threadCountRegex := regexp.MustCompile(`\s*Thread Count:\s*(\d+)`)

	for scanner.Scan() {
		line := scanner.Text()

		// 检测处理器段开始
		if strings.Contains(line, "Processor Information") && strings.Contains(line, "Handle") {
			if currentCPU != nil {
				cpus = append(cpus, *currentCPU)
			}
			currentCPU = &hardwareRequest.CPUCreateRequest{
				Architecture: "Unknown",
				BaseSpeed:    "Unknown",
			}
			inProcessor = true
			continue
		}

		if !inProcessor || currentCPU == nil {
			continue
		}

		// 解析插槽
		if matches := socketRegex.FindStringSubmatch(line); len(matches) > 1 {
			// Socket 信息可以存储在 Family 或其他字段
		}

		// 解析类型
		if matches := typeRegex.FindStringSubmatch(line); len(matches) > 1 {
			cpuType := strings.TrimSpace(matches[1])
			if cpuType != "" && cpuType != "Central Processor" {
				currentCPU.Family = cpuType
			}
		}

		// 解析家族
		if matches := familyRegex.FindStringSubmatch(line); len(matches) > 1 {
			family := strings.TrimSpace(matches[1])
			if family != "" && family != "Unknown" {
				currentCPU.Family = family
			}
		}

		// 解析制造商
		if matches := manufacturerRegex.FindStringSubmatch(line); len(matches) > 1 {
			manufacturer := strings.TrimSpace(matches[1])
			if manufacturer != "" && manufacturer != "Not Specified" {
				currentCPU.Vendor = manufacturer
			}
		}

		// 解析 ID (包含 stepping 等信息)
		if matches := idRegex.FindStringSubmatch(line); len(matches) > 1 {
			id := strings.TrimSpace(matches[1])
			if id != "" {
				currentCPU.Stepping = id
			}
		}

		// 解析版本（型号名称）
		if matches := versionRegex.FindStringSubmatch(line); len(matches) > 1 {
			version := strings.TrimSpace(matches[1])
			if version != "" && version != "Not Specified" {
				currentCPU.Model = version
			}
		}

		// 解析最大频率
		if matches := maxSpeedRegex.FindStringSubmatch(line); len(matches) > 1 {
			maxSpeed := strings.TrimSpace(matches[1])
			if maxSpeed != "" && maxSpeed != "Unknown" {
				currentCPU.BaseSpeed = maxSpeed
			}
		}

		// 解析当前频率
		if matches := currentSpeedRegex.FindStringSubmatch(line); len(matches) > 1 {
			currentSpeed := strings.TrimSpace(matches[1])
			if currentSpeed != "" && currentSpeed != "Unknown" && currentCPU.BaseSpeed == "Unknown" {
				currentCPU.BaseSpeed = currentSpeed
			}
		}

		// 解析核心数
		if matches := coreCountRegex.FindStringSubmatch(line); len(matches) > 1 {
			if cores, err := strconv.Atoi(matches[1]); err == nil {
				currentCPU.Cores = cores
			}
		}

		// 解析启用核心数
		if matches := coreEnabledRegex.FindStringSubmatch(line); len(matches) > 1 {
			// 可以记录但当前结构没有对应字段
		}

		// 解析线程数
		if matches := threadCountRegex.FindStringSubmatch(line); len(matches) > 1 {
			if threads, err := strconv.Atoi(matches[1]); err == nil {
				currentCPU.Threads = threads
			}
		}

		// 解析特性标志
		if matches := flagsRegex.FindStringSubmatch(line); len(matches) > 1 {
			flags := strings.TrimSpace(matches[1])
			if flags != "" {
				currentCPU.Flags = flags
			}
		}

		// 解析签名
		if matches := signatureRegex.FindStringSubmatch(line); len(matches) > 1 {
			signature := strings.TrimSpace(matches[1])
			if signature != "" {
				currentCPU.Flags = signature
			}
		}
	}

	// 添加最后一个 CPU
	if currentCPU != nil {
		cpus = append(cpus, *currentCPU)
	}

	return cpus
}

// collectCPUGeneric 使用 gopsutil 收集 CPU 信息
func (c *CPUCollector) collectCPUGeneric(hardwareInfo *hardwareRequest.HardwareInfoUploadRequest) error {
	// 获取 CPU 信息
	cpuInfos, err := cpu.Info()
	if err != nil {
		return fmt.Errorf("failed to get CPU info: %w", err)
	}

	if len(cpuInfos) == 0 {
		return fmt.Errorf("no CPU info available")
	}

	// 遍历所有 CPU（支持多 CPU 系统）
	for _, cpuInfo := range cpuInfos {
		// 获取物理核心数
		physicalCores, err := cpu.Counts(false)
		if err != nil {
			physicalCores = int(cpuInfo.Cores)
		}

		// 获取逻辑核心数
		logicalCores, err := cpu.Counts(true)
		if err != nil {
			logicalCores = int(cpuInfo.Cores)
		}

		cpuRequest := hardwareRequest.CPUCreateRequest{
			Success:  true,
			Message:  "collected via gopsutil",
			Model:    cpuInfo.ModelName,
			Vendor:   cpuInfo.VendorID,
			Family:   cpuInfo.Family,
			Stepping: fmt.Sprintf("%d", cpuInfo.Stepping),
			Flags:    fmt.Sprintf("%v", cpuInfo.Flags),
			Cores:    physicalCores,
			Threads:  logicalCores,
		}

		// 处理频率信息
		if cpuInfo.Mhz > 0 {
			cpuRequest.BaseSpeed = fmt.Sprintf("%.2f MHz", cpuInfo.Mhz)
		} else {
			cpuRequest.BaseSpeed = "Unknown"
		}

		// 缓存大小信息（gopsutil 未直接提供，设置为未知）
		cpuRequest.CacheSizeL1 = "Unknown"
		cpuRequest.CacheSizeL2 = "Unknown"
		cpuRequest.CacheSizeL3 = "Unknown"

		// 架构信息（gopsutil 未直接提供，设置为未知）
		cpuRequest.Architecture = "Unknown"

		hardwareInfo.CPUs = append(hardwareInfo.CPUs, cpuRequest)
	}

	return nil
}
