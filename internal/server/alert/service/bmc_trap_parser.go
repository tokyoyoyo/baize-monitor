package service

import (
	"baize-monitor/internal/server/alert/repository"
	"baize-monitor/pkg/dto/request"
	"baize-monitor/pkg/dto/response"
	"baize-monitor/pkg/models"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	// 只验证OID格式但不提取
	oidFormatRegex = regexp.MustCompile(`^\d+(?:\.\d+)+$`)
)

type BMCTrapParserService interface {
	Create(*request.BMCTrapParserCreate) (int, error)
	CheckParserNameExist(string, bool, int64) (bool, error)
	Update(*request.BMCTrapParserUpdate) (int, error)
	Delete(id int64) (int, error)
	Activate(id int64) (int, error)
	Deactivate(id int64) (int, error)
	List(*request.BMCTrapParserFilter) (response.BMCTrapParserListResult, int, error)
}

type BMCTrapParserServiceImp struct {
	repo repository.BMCTrapParserRepository
}

func NewBMCTrapParserServiceImp(repo repository.BMCTrapParserRepository) *BMCTrapParserServiceImp {
	return &BMCTrapParserServiceImp{repo: repo}
}

func (s *BMCTrapParserServiceImp) validateOID(oid, fieldName string) error {
	if strings.TrimSpace(oid) == "" {
		return fmt.Errorf("%s不能为空", fieldName)
	}

	if len(oid) > 500 {
		return fmt.Errorf("%s长度不能超过500个字符", fieldName)
	}

	if !oidFormatRegex.MatchString(oid) {
		return fmt.Errorf("%s格式不正确,必须是数字点分隔的格式", fieldName)
	}

	return nil
}

// validateRequiredOIDs 如果oid存在就验证。先判断是否为空决定要不要跳过，再清除空格，避免遗漏多个空格的情况
func (s *BMCTrapParserServiceImp) validateRequiredOIDs(req *request.BMCTrapParserCreate) error {
	// 验证四个必需的OID
	requiredOIDs := []struct {
		oid       string
		fieldName string
	}{
		{strings.TrimSpace(req.AlertLevelOID), "告警级别OID"},
		{strings.TrimSpace(req.AlertContentOID), "告警内容OID"},
		{strings.TrimSpace(req.AlertTimeOID), "告警时间OID"},
		{strings.TrimSpace(req.AlertComponentOID), "告警组件OID"},
	}

	for _, item := range requiredOIDs {
		if item.oid == "" {
			continue
		}
		if err := s.validateOID(item.oid, item.fieldName); err != nil {
			return err
		}
	}

	// 验证可选的OID（如果启用自动关闭）
	if req.EnableAutoClose {
		requiredOIDs = []struct {
			oid       string
			fieldName string
		}{
			{strings.TrimSpace(req.AlertIndexOID), "告警索引OID"},
			{strings.TrimSpace(req.AlertStatusOID), "告警状态OID"},
		}

		for _, item := range requiredOIDs {
			if item.oid == "" {
				continue
			}
			if err := s.validateOID(item.oid, item.fieldName); err != nil {
				return err
			}
		}

	}

	// 验证组件间关联OID
	if req.EnableContactInterComponentAlerts {
		if req.ContactInterComponentIdentifierOID != "" {
			if err := s.validateOID(strings.TrimSpace(req.ContactInterComponentIdentifierOID), "组件间关联OID"); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *BMCTrapParserServiceImp) validateTimeFormat(timeFormat string) error {
	if strings.TrimSpace(timeFormat) == "" {
		return errors.New("时间格式不能为空")
	}
	if len(timeFormat) > 200 {
		return errors.New("时间格式长度不能超过200个字符")
	}

	now := time.Now().UTC().Truncate(time.Second)
	formatted := now.Format(timeFormat)

	parsed, err := time.Parse(timeFormat, formatted)
	if err != nil {
		return fmt.Errorf("时间格式无效，无法解析自身生成的字符串: %w", err)
	}

	if !parsed.Truncate(time.Second).Equal(now) {
		return fmt.Errorf("时间格式 round-trip 失败：期望 %v,得到 %v", now, parsed.Truncate(time.Second))
	}

	return nil
}

func (s *BMCTrapParserServiceImp) validateMappings(req *request.BMCTrapParserCreate) error {
	// 验证级别映射
	if len(req.LevelMappings) == 0 {
		return errors.New("告警级别映射不能为空")
	}

	for key, value := range req.LevelMappings {
		if strings.TrimSpace(key) == "" {
			return errors.New("级别映射的键不能为空")
		}
		// 验证值是否是有效的告警级别
		if !isValidAlertLevel(value) {
			return fmt.Errorf("无效的告警级别：%s（键：%s）", value, key)
		}
	}

	// 如果启用自动关闭，验证状态映射
	if req.EnableAutoClose {
		if len(req.StatusMappings) != 2 {
			return errors.New("启用自动关闭时，告警状态映射不能为空")
		}

		for key, value := range req.StatusMappings {
			if strings.TrimSpace(key) == "" {
				return errors.New("状态映射的键不能为空")
			}
			if !isValidAlertStatus(value) {
				return fmt.Errorf("无效的告警状态：%s（键：%s）", value, key)
			}
		}
	}

	// 验证组件映射
	if len(req.ComponentMappings) != 0 {
		for component, identifiers := range req.ComponentMappings {
			if !isValidAlertComponent(component) {
				return fmt.Errorf("组件'%s'不合法", component)
			}
			if len(identifiers) == 0 {
				return fmt.Errorf("组件'%s'的标识符列表不能为空", component)
			}
			for _, identifier := range identifiers {
				if strings.TrimSpace(identifier) == "" {
					return fmt.Errorf("组件'%s'的标识符不能为空", component)
				}
			}
		}
	}
	return nil
}

func (s *BMCTrapParserServiceImp) Create(req *request.BMCTrapParserCreate) (int, error) {
	trimmedParserName := strings.TrimSpace(req.ParserName)
	if trimmedParserName == "" {
		return http.StatusBadRequest, fmt.Errorf("parserName不能为空")
	}
	exist, err := s.CheckParserNameExist(trimmedParserName, true, 0)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("parserName已存在")
	}
	if exist {
		return http.StatusBadRequest, fmt.Errorf("parserName已存在")
	}

	if err := s.validateRequiredOIDs(req); err != nil {
		return http.StatusBadRequest, err
	}
	if err := s.validateTimeFormat(req.TimeFormat); err != nil {
		return http.StatusBadRequest, err
	}
	if err := s.validateMappings(req); err != nil {
		return http.StatusBadRequest, err
	}

	parser, err := req.ToParserRecord()
	if err != nil {
		return http.StatusBadRequest, err
	}

	// 3. 调用Repository保存
	err = s.repo.Create(parser)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateParserName) {
			return http.StatusBadRequest, fmt.Errorf("parserName已存在")
		}
		return http.StatusInternalServerError, fmt.Errorf("保存解析器失败: %v", err)
	}
	return http.StatusCreated, nil
}

func (s *BMCTrapParserServiceImp) CheckParserNameExist(vendorName string, cteateModel bool, id int64) (bool, error) {
	return s.repo.IsParserNameExist(vendorName, cteateModel, id)
}

// Update 更新Trap解析器
func (s *BMCTrapParserServiceImp) Update(req *request.BMCTrapParserUpdate) (int, error) {
	// 检查记录是否存在
	parser, err := s.repo.FindByID(req.ID)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("获取解析器失败: %v", err)
	}
	if parser == nil {
		return http.StatusBadRequest, errors.New("解析器不存在")
	}

	// 验证请求（转换为Create请求进行验证
	req.UpdateParserRecord(parser)

	rc := new(request.BMCTrapParserCreate)
	rc.FromBMCTrapParser(parser)

	trimmedParserName := strings.TrimSpace(parser.ParserName)
	if trimmedParserName == "" {
		return http.StatusBadRequest, fmt.Errorf("parserName不能为空")
	}
	exist, err := s.CheckParserNameExist(trimmedParserName, true, 0)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("update failed: %v", err)
	}
	if exist {
		return http.StatusBadRequest, fmt.Errorf("parserName已存在")
	}
	if err := s.validateRequiredOIDs(rc); err != nil {
		return http.StatusBadRequest, err
	}
	if err := s.validateTimeFormat(rc.TimeFormat); err != nil {
		return http.StatusBadRequest, err
	}
	if err := s.validateMappings(rc); err != nil {
		return http.StatusBadRequest, err
	}

	// 更新
	if err := s.repo.Update(parser); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("更新失败: %v", err)
	}

	return http.StatusOK, nil
}

// Delete 删除Trap解析器
func (s *BMCTrapParserServiceImp) Delete(id int64) (int, error) {
	// 检查记录是否存在
	parser, err := s.repo.FindByID(id)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("获取解析器失败: %v", err)
	}
	if parser == nil {
		return http.StatusBadRequest, errors.New("解析器不存在")
	}

	// 执行软删除
	if err := s.repo.SoftDelete(id); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("删除Trap解析器失败: %v", err)
	}
	return http.StatusOK, nil
}

func (s *BMCTrapParserServiceImp) Activate(id int64) (int, error) {
	// 检查记录是否存在
	parser, err := s.repo.FindByID(id)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("获取解析器失败: %v", err)
	}
	if parser == nil {
		return http.StatusBadRequest, errors.New("解析器不存在")
	}
	if parser.IsActive {
		return http.StatusBadRequest, errors.New("解析器已处于激活状态")
	}
	if err := s.repo.Activate(id); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("激活Trap解析器失败: %v", err)
	}
	return http.StatusOK, nil
}

func (s *BMCTrapParserServiceImp) Deactivate(id int64) (int, error) {
	// 检查记录是否存在
	parser, err := s.repo.FindByID(id)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("获取解析器失败: %v", err)
	}
	if parser == nil {
		return http.StatusBadRequest, errors.New("解析器不存在")
	}
	if !parser.IsActive {
		return http.StatusBadRequest, errors.New("解析器已处于非激活状态")
	}
	if err := s.repo.Deactivate(id); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("激活Trap解析器失败: %v", err)
	}
	return http.StatusOK, nil
}

func (s *BMCTrapParserServiceImp) List(filter *request.BMCTrapParserFilter) (response.BMCTrapParserListResult, int, error) {
	var resp response.BMCTrapParserListResult
	parsers, total, err := s.repo.List(filter)
	if err != nil {
		return response.BMCTrapParserListResult{}, http.StatusInternalServerError, fmt.Errorf("获取列表失败: %v", err)
	}

	resp.Total = total
	resp.Page = filter.Page
	resp.PageSize = filter.PageSize
	resp.List = make([]*response.BMCTrapParserResponse, len(parsers))
	for i, parser := range parsers {
		resp.List[i] = parser.ToResponse()
	}
	return resp, http.StatusOK, nil

}

// ============ 辅助函数 ============

// isValidAlertLevel 验证告警级别是否有效
func isValidAlertLevel(level string) bool {
	switch models.AlertLevel(level) {
	case models.Critical,
		models.Info,
		models.Warning,
		models.Notification:
		return true
	default:
		return false
	}
}

// isValidAlertStatus 验证告警状态是否有效
func isValidAlertStatus(status string) bool {
	switch models.AlertStatus(status) {
	case models.Asserted,
		models.Deasserted:
		return true
	default:
		return false
	}
}

var componentSet = getComponentSet()

func getComponentSet() map[models.AlertComponent]struct{} {
	cs := make(map[models.AlertComponent]struct{})
	for _, bmd := range models.BMCAlertComponentDescriptions {
		cs[bmd.Component] = struct{}{}
		for _, sub := range bmd.SubComponentsDescription {
			cs[sub.Component] = struct{}{}
		}
	}
	return cs
}

func isValidAlertComponent(component string) bool {
	c := models.AlertComponent(component)
	_, exists := componentSet[c]
	return exists
}
