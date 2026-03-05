package memory

import (
	"baize-monitor/internal/agent/anomaly/types"
	"baize-monitor/pkg/models"
)

// detectors stores all registered memory detectors
var detectors = make([]types.Detector, 0)

// RegisterDetector registers a memory detector
func RegisterDetector(detector types.Detector) {
	detectors = append(detectors, detector)
}

// GetAllDetectors returns all registered detectors
func GetAllDetectors() []types.Detector {
	return detectors
}

// NewMemoryChecker creates a new memory checker
func NewMemoryChecker() types.AnomalyChecker {
	return &MemoryChecker{
		detectors: GetAllDetectors(),
	}
}

// MemoryChecker implements types.AnomalyChecker for Memory
type MemoryChecker struct {
	detectors []types.Detector
}

// CheckType returns the check type for Memory
func (m *MemoryChecker) CheckType() string {
	return string(models.CheckTypeMemory)
}

// GetAllDetectors returns all detectors for this checker
func (m *MemoryChecker) GetAllDetectors() []types.Detector {
	return m.detectors
}
