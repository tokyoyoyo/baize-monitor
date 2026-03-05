package network

import (
	"baize-monitor/internal/agent/anomaly/types"
	"baize-monitor/pkg/models"
)

// detectors stores all registered network detectors
var detectors = make([]types.Detector, 0)

// RegisterDetector registers a network detector
func RegisterDetector(detector types.Detector) {
	detectors = append(detectors, detector)
}

// GetAllDetectors returns all registered detectors
func GetAllDetectors() []types.Detector {
	return detectors
}

// NewNetworkChecker creates a new network checker
func NewNetworkChecker() types.AnomalyChecker {
	return &NetworkChecker{
		detectors: GetAllDetectors(),
	}
}

// NetworkChecker implements types.AnomalyChecker for Network
type NetworkChecker struct {
	detectors []types.Detector
}

// CheckType returns the check type for Network
func (n *NetworkChecker) CheckType() string {
	return string(models.CheckTypeNetwork)
}

// GetAllDetectors returns all detectors for this checker
func (n *NetworkChecker) GetAllDetectors() []types.Detector {
	return n.detectors
}
