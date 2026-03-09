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

type MemoryCollector struct{}

func NewMemoryCollector() *MemoryCollector {
	return &MemoryCollector{}
}

func (m *MemoryCollector) Collect(hardwareInfo *hardwareRequest.HardwareInfoRequest) {
	memoryRequest := hardwareRequest.MemoryRequest{
		Content: make([]hardwareRequest.MemoryModuleInfo, 0),
		Summary: hardwareRequest.MemorySummary{},
	}

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		memoryRequest.Success = false
		memoryRequest.Message = fmt.Sprintf("failed to get memory info: %v", err)
		hardwareInfo.Memory = memoryRequest
		return
	}

	memoryRequest.Summary.TotalSize = int64(memInfo.Total)

	if runtime.GOOS == "linux" {
		moduleErr := m.collectMemoryModulesLinux(&memoryRequest)
		if moduleErr != nil {
			memoryRequest.Success = false
			memoryRequest.Message = fmt.Sprintf("partial collection failed: %v", moduleErr)
		}
	} else {
		m.collectMemoryModulesGeneric(&memoryRequest)
	}

	if !memoryRequest.Success && memoryRequest.Message == "" {
		memoryRequest.Success = true
		memoryRequest.Message = "collected successfully"
	}

	hardwareInfo.Memory = memoryRequest
}

func (m *MemoryCollector) collectMemoryModulesLinux(memoryRequest *hardwareRequest.MemoryRequest) error {
	cmd := exec.Command("dmidecode", "-t", "memory")
	output, err := cmd.Output()
	if err != nil {
		m.collectMemoryModulesGeneric(memoryRequest)
		return nil
	}

	modules := parseDmiDecodeMemory(string(output))

	var memoryType string
	var memorySpeed string
	typeCount := make(map[string]int)
	speedCount := make(map[string]int)

	for i := range modules {
		modules[i].Success = true
		modules[i].Message = "collected via dmidecode"
		memoryRequest.Content = append(memoryRequest.Content, modules[i])

		if modules[i].Size > 0 {
			memoryRequest.Summary.UsedSlots++
			typeCount[modules[i].Type]++
			speedCount[modules[i].Speed]++
		}
	}

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

	memoryRequest.Summary.Type = memoryType
	memoryRequest.Summary.Speed = memorySpeed
	memoryRequest.Summary.TotalSlots = memoryRequest.Summary.UsedSlots
	return nil
}

func parseDmiDecodeMemory(output string) []hardwareRequest.MemoryModuleInfo {
	var modules []hardwareRequest.MemoryModuleInfo
	var currentModule *hardwareRequest.MemoryModuleInfo

	scanner := bufio.NewScanner(strings.NewReader(output))
	inMemoryDevice := false

	slotRegex := regexp.MustCompile(`\s*Locator:\s*(.+)`)
	sizeRegex := regexp.MustCompile(`\s*Size:\s*(\d+)\s*MB`)
	typeRegex := regexp.MustCompile(`\s*Type:\s*(DDR\w*)`)
	speedRegex := regexp.MustCompile(`\s*Speed:\s*(\d+)\s*MT/s`)
	manufacturerRegex := regexp.MustCompile(`\s*Manufacturer:\s*(.+)`)
	serialRegex := regexp.MustCompile(`\s*Serial Number:\s*(.+)`)
	partNumberRegex := regexp.MustCompile(`\s*Part Number:\s*(.+)`)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "Memory Device") && strings.Contains(line, "Handle") {
			if currentModule != nil {
				modules = append(modules, *currentModule)
			}
			currentModule = &hardwareRequest.MemoryModuleInfo{
				Type:  "Unknown",
				Speed: "Unknown",
			}
			inMemoryDevice = true
			continue
		}

		if !inMemoryDevice || currentModule == nil {
			continue
		}

		if matches := slotRegex.FindStringSubmatch(line); len(matches) > 1 {
			currentModule.Slot = strings.TrimSpace(matches[1])
		}

		if matches := sizeRegex.FindStringSubmatch(line); len(matches) > 1 {
			if sizeMB, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
				currentModule.Size = sizeMB * 1024 * 1024
			}
		}

		if matches := typeRegex.FindStringSubmatch(line); len(matches) > 1 {
			currentModule.Type = strings.TrimSpace(matches[1])
		}

		if matches := speedRegex.FindStringSubmatch(line); len(matches) > 1 {
			currentModule.Speed = matches[1] + " MT/s"
		}

		if matches := manufacturerRegex.FindStringSubmatch(line); len(matches) > 1 {
			manufacturer := strings.TrimSpace(matches[1])
			if manufacturer != "" && manufacturer != "Not Specified" {
				currentModule.Manufacturer = manufacturer
			}
		}

		if matches := serialRegex.FindStringSubmatch(line); len(matches) > 1 {
			serial := strings.TrimSpace(matches[1])
			if serial != "" && serial != "Not Specified" {
				currentModule.SerialNumber = serial
			}
		}

		if matches := partNumberRegex.FindStringSubmatch(line); len(matches) > 1 {
			partNumber := strings.TrimSpace(matches[1])
			if partNumber != "" && partNumber != "Not Specified" {
				currentModule.PartNumber = partNumber
			}
		}
	}

	if currentModule != nil {
		modules = append(modules, *currentModule)
	}

	return modules
}

func (m *MemoryCollector) collectMemoryModulesGeneric(memoryRequest *hardwareRequest.MemoryRequest) {
	memoryRequest.Summary.Type = "Unknown"
	memoryRequest.Summary.Speed = "Unknown"
	memoryRequest.Summary.TotalSlots = 0
	memoryRequest.Summary.UsedSlots = 0
	if memoryRequest.Message == "" {
		memoryRequest.Success = true
		memoryRequest.Message = "collected via generic method (limited info)"
	}
}
