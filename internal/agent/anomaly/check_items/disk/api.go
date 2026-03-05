package disk

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
		CheckType: string(models.CheckTypeDisk),
		CheckItem: detectorType,
		Success:   false,
		Level:     level,
		Message:   message,
		ExtraData: extraData,
		CheckedAt: time.Now(),
	}
}

func NewDiskChecker() check_items.AnomalyChecker {
	return &DiskChecker{
		detectors: GetAllDetectors(),
	}
}

type DiskChecker struct {
	detectors []check_items.Detector
}

func (d *DiskChecker) CheckType() string {
	return string(models.CheckTypeDisk)
}

func (d *DiskChecker) GetAllDetectors() []check_items.Detector {
	return d.detectors
}
