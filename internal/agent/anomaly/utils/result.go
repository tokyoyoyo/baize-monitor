package utils

import (
	"baize-monitor/pkg/dto/request"
	"baize-monitor/pkg/models"
	"time"
)

// CreateAnomalyResult creates an anomaly result with the given parameters
func CreateAnomalyResult(checkType models.CheckType, detectorType string, level models.AnomalyLevel, message string, extraData map[string]interface{}) request.AnomalyResult {
	return request.AnomalyResult{
		CheckType: string(checkType),
		CheckItem: detectorType,
		Success:   false,
		Level:     string(level),
		Message:   message,
		ExtraData: extraData,
		CheckedAt: time.Now(),
	}
}
