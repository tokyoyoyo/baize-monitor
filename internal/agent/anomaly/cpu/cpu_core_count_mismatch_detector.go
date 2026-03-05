package cpu

import (
	"baize-monitor/internal/agent/anomaly/utils"
	"baize-monitor/pkg/dto/request"
	"baize-monitor/pkg/models"

	gopsutilcpu "github.com/shirou/gopsutil/v3/cpu"
)

type cpuCoreCountMismatchDetector struct {
	expectedCores int
}

func init() {
	RegisterDetector(&cpuCoreCountMismatchDetector{
		expectedCores: 0,
	})
}

func (d *cpuCoreCountMismatchDetector) CheckItemName() string {
	return "cpu_core_count_mismatch"
}

func (d *cpuCoreCountMismatchDetector) Detect() (request.AnomalyResult, error) {
	cpuInfos, err := gopsutilcpu.Info()
	if err != nil {
		return request.AnomalyResult{}, err
	}

	if len(cpuInfos) == 0 {
		return request.AnomalyResult{}, nil
	}

	physicalCores, err := gopsutilcpu.Counts(false)
	if err != nil {
		return request.AnomalyResult{}, err
	}

	logicalCores, err := gopsutilcpu.Counts(true)
	if err != nil {
		return request.AnomalyResult{}, err
	}

	if d.expectedCores > 0 && physicalCores != d.expectedCores {
		return utils.CreateAnomalyResult(
			models.CheckTypeCPU,
			d.CheckItemName(),
			models.AnomalyLevelWarning,
			"CPU core count mismatch detected",
			map[string]interface{}{
				"expected_cores":        d.expectedCores,
				"actual_physical_cores": physicalCores,
				"actual_logical_cores":  logicalCores,
				"cpu_count":             len(cpuInfos),
				"issue":                 "core_count_mismatch",
			},
		), nil
	}

	if logicalCores < physicalCores {
		return utils.CreateAnomalyResult(
			models.CheckTypeCPU,
			d.CheckItemName(),
			models.AnomalyLevelWarning,
			"Logical cores less than physical cores - possible CPU degradation",
			map[string]interface{}{
				"physical_cores": physicalCores,
				"logical_cores":  logicalCores,
				"issue":          "hyperthreading_disabled",
			},
		), nil
	}

	return request.AnomalyResult{}, nil
}
