package network

import (
	"baize-monitor/internal/agent/anomaly/check_items"
	"baize-monitor/pkg/dto/request"
	"baize-monitor/pkg/models"
	"time"
)

var detectors = make([]check_items.Detector, 0)

func RegisterDetector(detector check_items.Detector) {
	detectors = append(detectors, detector)
}

func GetAllDetectors() []check_items.Detector {
	return detectors
}

func CreateAnomalyResult(detectorType string, level, message string, extraData map[string]interface{}) request.AnomalyResult {
	return request.AnomalyResult{
		CheckType: string(models.CheckTypeNetwork),
		CheckItem: detectorType,
		Success:   false,
		Level:     level,
		Message:   message,
		ExtraData: extraData,
		CheckedAt: time.Now(),
	}
}

func NewNetworkChecker() check_items.AnomalyChecker {
	return &NetworkChecker{
		detectors: GetAllDetectors(),
	}
}

type NetworkChecker struct {
	detectors []check_items.Detector
}

func (n *NetworkChecker) CheckType() string {
	return string(models.CheckTypeNetwork)
}

func (n *NetworkChecker) GetAllDetectors() []check_items.Detector {
	return n.detectors
}
