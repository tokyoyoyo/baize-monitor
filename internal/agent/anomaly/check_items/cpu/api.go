package cpu

import (
	"baize-monitor/internal/agent/anomaly/check_items"
	"baize-monitor/pkg/dto/request"
	"baize-monitor/pkg/models"
	"time"
)

// Detector interface for CPU anomaly detection
var detectors = make([]check_items.Detector, 0)

func RegisterDetector(detector check_items.Detector) {
	detectors = append(detectors, detector)
}

func GetAllDetectors() []check_items.Detector {
	return detectors
}

func CreateAnomalyResult(detectorType string, level, message string, extraData map[string]interface{}) request.AnomalyResult {
	return request.AnomalyResult{
		CheckType: string(models.CheckTypeCPU),
		CheckItem: detectorType,
		Success:   false,
		Level:     level,
		Message:   message,
		ExtraData: extraData,
		CheckedAt: time.Now(),
	}
}

func NewCPUChecker() check_items.AnomalyChecker {
	return &CPUChecker{
		detectors: GetAllDetectors(),
	}
}

type CPUChecker struct {
	detectors []check_items.Detector
}

func (c *CPUChecker) CheckType() string {
	return string(models.CheckTypeCPU)
}

func (c *CPUChecker) GetAllDetectors() []check_items.Detector {
	return c.detectors
}
