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

type CPUCollector struct{}

func NewCPUCollector() *CPUCollector {
	return &CPUCollector{}
}

func (c *CPUCollector) Collect(hardwareInfo *hardwareRequest.HardwareInfoRequest) {
	cpuRequest := hardwareRequest.CPURequest{
		Content: make([]hardwareRequest.CPUInfo, 0),
		Summary: hardwareRequest.CPUSummary{},
	}

	if runtime.GOOS == "linux" {
		dmidecodeErr := c.collectCPULinux(&cpuRequest)
		if dmidecodeErr != nil {
			gopsutilErr := c.collectCPUGeneric(&cpuRequest)
			if gopsutilErr != nil {
				cpuRequest.Success = false
				cpuRequest.Message = fmt.Sprintf("failed to collect CPU info: dmidecode failed(%v), gopsutil also failed(%v)", dmidecodeErr, gopsutilErr)
			}
		}
	} else {
		if err := c.collectCPUGeneric(&cpuRequest); err != nil {
			cpuRequest.Success = false
			cpuRequest.Message = fmt.Sprintf("failed to collect CPU info: %v", err)
		}
	}

	if !cpuRequest.Success && cpuRequest.Message == "" {
		cpuRequest.Success = true
		cpuRequest.Message = "collected successfully"
	}

	hardwareInfo.CPUs = cpuRequest
}

func (c *CPUCollector) collectCPULinux(cpuRequest *hardwareRequest.CPURequest) error {
	cmd := exec.Command("dmidecode", "-t", "processor")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("dmidecode failed: %w", err)
	}

	cpus := parseDmiDecodeCPU(string(output))

	if len(cpus) == 0 {
		return fmt.Errorf("no CPU info from dmidecode")
	}

	cpuRequest.Content = cpus

	totalCores := 0
	totalThreads := 0
	for _, cpu := range cpus {
		totalCores += cpu.Cores
		totalThreads += cpu.Threads
	}
	cpuRequest.Summary.TotalCores = totalCores
	cpuRequest.Summary.TotalThreads = totalThreads

	return nil
}

func parseDmiDecodeCPU(output string) []hardwareRequest.CPUInfo {
	var cpus []hardwareRequest.CPUInfo
	var currentCPU *hardwareRequest.CPUInfo

	scanner := bufio.NewScanner(strings.NewReader(output))
	inProcessor := false

	maxSpeedRegex := regexp.MustCompile(`\s*Max Speed:\s*(.+)`)
	currentSpeedRegex := regexp.MustCompile(`\s*Current Speed:\s*(.+)`)
	coreCountRegex := regexp.MustCompile(`\s*Core Count:\s*(\d+)`)
	threadCountRegex := regexp.MustCompile(`\s*Thread Count:\s*(\d+)`)
	versionRegex := regexp.MustCompile(`\s*Version:\s*(.+)`)
	manufacturerRegex := regexp.MustCompile(`\s*Manufacturer:\s*(.+)`)
	familyRegex := regexp.MustCompile(`\s*Family:\s*(.+)`)
	idRegex := regexp.MustCompile(`\s*ID:\s*(.+)`)
	flagsRegex := regexp.MustCompile(`\s*Flags:\s*(.+)`)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "Processor Information") && strings.Contains(line, "Handle") {
			if currentCPU != nil {
				cpus = append(cpus, *currentCPU)
			}
			currentCPU = &hardwareRequest.CPUInfo{
				Architecture: "Unknown",
				BaseSpeed:    "Unknown",
			}
			inProcessor = true
			continue
		}

		if !inProcessor || currentCPU == nil {
			continue
		}

		if matches := versionRegex.FindStringSubmatch(line); len(matches) > 1 {
			version := strings.TrimSpace(matches[1])
			if version != "" && version != "Not Specified" {
				currentCPU.Model = version
			}
		}

		if matches := manufacturerRegex.FindStringSubmatch(line); len(matches) > 1 {
			manufacturer := strings.TrimSpace(matches[1])
			if manufacturer != "" && manufacturer != "Not Specified" {
				currentCPU.Vendor = manufacturer
			}
		}

		if matches := familyRegex.FindStringSubmatch(line); len(matches) > 1 {
			family := strings.TrimSpace(matches[1])
			if family != "" && family != "Unknown" {
				currentCPU.Family = family
			}
		}

		if matches := idRegex.FindStringSubmatch(line); len(matches) > 1 {
			id := strings.TrimSpace(matches[1])
			if id != "" {
				currentCPU.Stepping = id
			}
		}

		if matches := maxSpeedRegex.FindStringSubmatch(line); len(matches) > 1 {
			maxSpeed := strings.TrimSpace(matches[1])
			if maxSpeed != "" && maxSpeed != "Unknown" {
				currentCPU.BaseSpeed = maxSpeed
			}
		}

		if matches := currentSpeedRegex.FindStringSubmatch(line); len(matches) > 1 {
			currentSpeed := strings.TrimSpace(matches[1])
			if currentSpeed != "" && currentCPU.BaseSpeed == "Unknown" {
				currentCPU.BaseSpeed = currentSpeed
			}
		}

		if matches := coreCountRegex.FindStringSubmatch(line); len(matches) > 1 {
			if cores, err := strconv.Atoi(matches[1]); err == nil {
				currentCPU.Cores = cores
			}
		}

		if matches := threadCountRegex.FindStringSubmatch(line); len(matches) > 1 {
			if threads, err := strconv.Atoi(matches[1]); err == nil {
				currentCPU.Threads = threads
			}
		}

		if matches := flagsRegex.FindStringSubmatch(line); len(matches) > 1 {
			flags := strings.TrimSpace(matches[1])
			if flags != "" {
				currentCPU.Flags = flags
			}
		}
	}

	if currentCPU != nil {
		cpus = append(cpus, *currentCPU)
	}

	return cpus
}

func (c *CPUCollector) collectCPUGeneric(cpuRequest *hardwareRequest.CPURequest) error {
	cpuInfos, err := cpu.Info()
	if err != nil {
		return fmt.Errorf("failed to get CPU info: %w", err)
	}

	if len(cpuInfos) == 0 {
		return fmt.Errorf("no CPU info available")
	}

	physicalCores, err := cpu.Counts(false)
	if err != nil {
		physicalCores = 0
	}

	logicalCores, err := cpu.Counts(true)
	if err != nil {
		logicalCores = 0
	}

	for _, cpuInfo := range cpuInfos {
		cpuInfoRequest := hardwareRequest.CPUInfo{
			Model:    cpuInfo.ModelName,
			Vendor:   cpuInfo.VendorID,
			Family:   cpuInfo.Family,
			Stepping: fmt.Sprintf("%d", cpuInfo.Stepping),
			Flags:    fmt.Sprintf("%v", cpuInfo.Flags),
			Cores:    physicalCores,
			Threads:  logicalCores,
		}

		if cpuInfo.Mhz > 0 {
			cpuInfoRequest.BaseSpeed = fmt.Sprintf("%.2f MHz", cpuInfo.Mhz)
		} else {
			cpuInfoRequest.BaseSpeed = "Unknown"
		}

		cpuInfoRequest.CacheSizeL1 = "Unknown"
		cpuInfoRequest.CacheSizeL2 = "Unknown"
		cpuInfoRequest.CacheSizeL3 = "Unknown"
		cpuInfoRequest.Architecture = "Unknown"

		cpuRequest.Content = append(cpuRequest.Content, cpuInfoRequest)
	}

	cpuRequest.Summary.TotalCores = physicalCores
	cpuRequest.Summary.TotalThreads = logicalCores

	return nil
}
