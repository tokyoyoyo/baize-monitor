package types

import "baize-monitor/pkg/dto/request"

// CheckerRegistry manages all anomaly checkers
type CheckerRegistry struct {
	checkers []AnomalyChecker
}

// NewCheckerRegistry creates a new checker registry
func NewCheckerRegistry() *CheckerRegistry {
	return &CheckerRegistry{
		checkers: make([]AnomalyChecker, 0),
	}
}

// Register adds a checker to the registry
func (r *CheckerRegistry) Register(collector AnomalyChecker) {
	r.checkers = append(r.checkers, collector)
}

// CheckerALL runs all detectors and returns results
func (r *CheckerRegistry) CheckerALL() request.AnomalyResultRequest {
	anomalyReq := &request.AnomalyResultRequest{
		Content: make([]request.AnomalyResult, 0),
	}
	for _, checker := range r.checkers {
		for _, detector := range checker.GetAllDetectors() {
			result, err := detector.Detect()
			if err != nil {
				// 检测出错，记录错误信息
				result.Success = false
				result.Message = err.Error()
				result.CheckItem = detector.CheckItemName()
				result.CheckType = checker.CheckType()
				anomalyReq.Content = append(anomalyReq.Content, result)
				continue
			}
			anomalyReq.Content = append(anomalyReq.Content, result)
		}
	}
	return *anomalyReq
}
