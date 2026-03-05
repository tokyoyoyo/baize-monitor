package types

import "baize-monitor/pkg/dto/request"

// Detector interface for anomaly detection
type Detector interface {
	CheckItemName() string
	Detect() (request.AnomalyResult, error)
}

// AnomalyChecker interface for anomaly checkers
type AnomalyChecker interface {
	CheckType() string
	GetAllDetectors() []Detector
}
