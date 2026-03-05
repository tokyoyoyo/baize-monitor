package check_items

import (
	"baize-monitor/pkg/dto/request"
)

type Detector interface {
	DetectorType() string
	Detect() (request.AnomalyResult, error)
}

type AnomalyChecker interface {
	CheckType() string
	GetAllDetectors() []Detector
}

type CheckerRegistry struct {
	checkers []AnomalyChecker
}

func NewCheckerRegistry() *CheckerRegistry {
	return &CheckerRegistry{
		checkers: make([]AnomalyChecker, 0),
	}
}

func (r *CheckerRegistry) Register(collector AnomalyChecker) {
	r.checkers = append(r.checkers, collector)
}

func (r *CheckerRegistry) AutoDiscover() {
	// 检测器会在 anomaly.New() 时通过空白导入自动注册
	// 这里不需要手动注册具体的 Checker
}

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
				result.CheckItem = detector.DetectorType()
				anomalyReq.Content = append(anomalyReq.Content, result)
				continue
			} 
			anomalyReq.Content = append(anomalyReq.Content, result)
		}
	}
	return *anomalyReq
}
