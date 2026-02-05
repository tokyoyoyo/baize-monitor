package service

import (
	"baize-monitor/internal/server/alert/repository"
	"baize-monitor/pkg/dto/request"
	"baize-monitor/pkg/dto/response"
	"baize-monitor/pkg/models"
	"errors"
	"fmt"
	"net/http"

	"gorm.io/gorm"
)

type BMCTrapService interface {
	Process(tm *models.TrapMessage) (int, error)
	List(filter *request.AlertFilter) (response.AlertListResult, int, error)
	Update(req request.AlertUpdate) (int, error)
}

type BMCTrapServiceImp struct {
	pc   *BMCTrapParserCache
	repo repository.AlertRepo
}

func NewBMCTrapServiceImp(pc *BMCTrapParserCache, repo repository.AlertRepo) *BMCTrapServiceImp {
	return &BMCTrapServiceImp{
		pc:   pc,
		repo: repo,
	}
}

func (s *BMCTrapServiceImp) Process(tm *models.TrapMessage) (int, error) {
	p, err := s.pc.FindParser(tm)
	if err != nil {
		// 获取解析器失败，走兜底策略
		alert := &models.Alert{
			SourceIP:    tm.SourceIP.String(),
			SourceType:  tm.SourceType,
			AlertStatus: models.AlertStatusActive,
			AlertLevel:  models.AlertLevelWarning,
			AlertTime:   tm.ReceivedAt,
			Component:   string(models.AlertComponentUnknown),
			VariableMap: tm.VariableMap,
			RawData:     tm.RawData,
		}
		err := s.repo.Create(alert)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("fail to create alert, err: %v", err)
		}
		return http.StatusOK, nil
	}

	alert, err := p.Parse(tm)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("fail to process bmc trap, err: %v", err)
	}

	if alert.EnableAutoClose && alert.TrapStatus == models.TrapStatusDeasserted {
		active := string(models.AlertStatusActive)
		page := 1
		PageNum := 10
		filter := &request.AlertFilter{
			Page:      page,
			PageSize:  PageNum,
			SourceIP:  &alert.SourceIP,
			Status:    &active,
			TrapIndex: &alert.TrapIndex,
		}
		res, err := s.repo.List(filter)
		if err != nil {
			//TODO 关闭的报文处理失败只需要记日志
			return http.StatusOK, nil
		}
		if res.Total > 0 {
			if len(res.List) == 1 {
				alertRecord := res.List[0]
				alertRecord.AlertStatus = models.AlertStatusCleared
				err = s.repo.Update(alertRecord)
				if err != nil {
					// TODO 告警自动关闭失败,记日志
				}
				return http.StatusOK, nil
			}
		}
		//TODO 如果是关闭报文，查出对应的告警，处理告警自动关闭，不管成功与否都要结束执行
		return http.StatusOK, nil
	}

	alert.AlertStatus = models.AlertStatusActive
	// 新的告警报文，直接落库
	err = s.repo.Create(alert)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("fail to create alert, err: %v", err)
	}
	return http.StatusOK, nil
}

func (s *BMCTrapServiceImp) List(filter *request.AlertFilter) (response.AlertListResult, int, error) {
	res, err := s.repo.List(filter)
	if err != nil {
		return response.AlertListResult{}, http.StatusInternalServerError, fmt.Errorf("fail to list alert, err: %v", err)
	}
	return res, http.StatusOK, nil
}

func (s *BMCTrapServiceImp) Update(req request.AlertUpdate) (int, error) {
	alertRecord, err := s.repo.FindByID(req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 返回 400 状态码和预期的错误消息
			return http.StatusBadRequest, errors.New("alert not found")
		}
		return http.StatusInternalServerError, fmt.Errorf("fail to find alert, err: %v", err)
	}

	alertRecord.AlertStatus = req.Status
	err = s.repo.Update(alertRecord)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("fail to update alert, err: %v", err)
	}
	return http.StatusOK, nil
}
