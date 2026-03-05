package memory

import (
	"bufio"
	"os/exec"
	"strconv"
	"strings"

	"baize-monitor/internal/agent/anomaly/utils"
	"baize-monitor/pkg/dto/request"
	"baize-monitor/pkg/models"
)

type memoryECCErrorDetector struct {
	errorThreshold int
}

func init() {
	RegisterDetector(&memoryECCErrorDetector{
		errorThreshold: 1,
	})
}

func (d *memoryECCErrorDetector) CheckItemName() string {
	return "memory_ecc_error"
}

func (d *memoryECCErrorDetector) Detect() (request.AnomalyResult, error) {
	cmd := exec.Command("edac-util", "-v")
	output, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("cat", "/sys/devices/system/edac/mc/mc*/ece_count")
		output, err = cmd.Output()
		if err != nil {
			return request.AnomalyResult{}, nil
		}
	}

	totalErrors := 0
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "ecc_correctable_errors") ||
			strings.Contains(line, "ece_count") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				if count, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
					totalErrors += count
				}
			}
		}
	}

	if totalErrors >= d.errorThreshold {
		return utils.CreateAnomalyResult(
			models.CheckTypeMemory,
			d.CheckItemName(),
			models.AnomalyLevelWarning,
			"Memory ECC correctable errors detected",
			map[string]interface{}{
				"total_ecc_errors": totalErrors,
				"threshold":        d.errorThreshold,
				"issue":            "accumulated_ecc_errors",
			},
		), nil
	}

	return request.AnomalyResult{}, nil
}
