package cpu

import (
	"baize-monitor/internal/agent/anomaly/types"
	"baize-monitor/pkg/models"
)

// detectors stores all registered CPU detectors
var detectors = make([]types.Detector, 0)

// RegisterDetector registers a CPU detector
func RegisterDetector(detector types.Detector) {
	detectors = append(detectors, detector)
}

// GetAllDetectors returns all registered detectors
func GetAllDetectors() []types.Detector {
	return detectors
}

// NewCPUChecker creates a new CPU checker
func NewCPUChecker() types.AnomalyChecker {
	return &CPUChecker{
		detectors: GetAllDetectors(),
	}
}

// CPUChecker implements types.AnomalyChecker for CPU
type CPUChecker struct {
	detectors []types.Detector
}

// CheckType returns the check type for CPU
func (c *CPUChecker) CheckType() string {
	return string(models.CheckTypeCPU)
}

// GetAllDetectors returns all detectors for this checker
func (c *CPUChecker) GetAllDetectors() []types.Detector {
	return c.detectors
}
