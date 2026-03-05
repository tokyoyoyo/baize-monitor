package memory

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
		CheckType: string(models.CheckTypeMemory),
		CheckItem: detectorType,
		Success:   false,
		Level:     level,
		Message:   message,
		ExtraData: extraData,
		CheckedAt: time.Now(),
	}
}

func NewMemoryChecker() check_items.AnomalyChecker {
	return &MemoryChecker{
		detectors: GetAllDetectors(),
	}
}

type MemoryChecker struct {
	detectors []check_items.Detector
}

func (m *MemoryChecker) CheckType() string {
	return string(models.CheckTypeMemory)
}

func (m *MemoryChecker) GetAllDetectors() []check_items.Detector {
	return m.detectors
}
