package disk

import (
	"baize-monitor/internal/agent/anomaly/types"
	"baize-monitor/pkg/models"
)

// detectors stores all registered disk detectors
var detectors = make([]types.Detector, 0)

// RegisterDetector registers a disk detector
func RegisterDetector(detector types.Detector) {
	detectors = append(detectors, detector)
}

// GetAllDetectors returns all registered detectors
func GetAllDetectors() []types.Detector {
	return detectors
}

// NewDiskChecker creates a new disk checker
func NewDiskChecker() types.AnomalyChecker {
	return &DiskChecker{
		detectors: GetAllDetectors(),
	}
}

// DiskChecker implements types.AnomalyChecker for Disk
type DiskChecker struct {
	detectors []types.Detector
}

// CheckType returns the check type for Disk
func (d *DiskChecker) CheckType() string {
	return string(models.CheckTypeDisk)
}

// GetAllDetectors returns all detectors for this checker
func (d *DiskChecker) GetAllDetectors() []types.Detector {
	return d.detectors
}
