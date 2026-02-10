// pkg/dto/request/bmc_trap_parser.go
package request

import (
	"baize-monitor/pkg/constants"
	"baize-monitor/pkg/models"
	"errors"
	"fmt"
	"strings"
)

// BMCTrapParserCreate 创建解析器请求
type BMCTrapParserCreate struct {
	ParserName string `json:"parser_name" binding:"required,min=1,max=200"`
	VendorName string `json:"vendor_name" binding:"required,min=1,max=200"`

	// 核心OID字段
	AlertLevelOID     string `json:"alert_level_oid" binding:"required,min=5,max=500"`
	AlertContentOID   string `json:"alert_content_oid" binding:"required,min=5,max=500"`
	AlertTimeOID      string `json:"alert_time_oid" binding:"required,min=5,max=500"`
	AlertComponentOID string `json:"alert_component_oid" binding:"required,min=5,max=500"`

	// 自动关闭配置
	EnableAutoClose bool   `json:"enable_auto_close"`
	AlertIndexOID   string `json:"alert_index_oid" binding:"required_if=EnableAutoClose true,max=500"`
	AlertStatusOID  string `json:"alert_status_oid" binding:"required_if=EnableAutoClose true,max=500"`

	// 组件间关联配置
	EnableContactInterComponentAlerts  bool   `json:"enable_contact_inter_component_alerts"`
	ContactInterComponentIdentifierOID string `json:"contact_inter_component_identifier_oid" binding:"required_if=EnableContactInterComponentAlerts true,max=500"`

	TimeFormat string `json:"time_format" binding:"required,max=200"`

	// 映射配置
	LevelMappings         map[string]string   `json:"level_mappings"`
	StatusMappings        map[string]string   `json:"status_mappings"`
	EnableProductNameList []string            `json:"enable_product_name_list"`
	EnableHostNameList    []string            `json:"enable_host_name_list"`
	ComponentMappings     map[string][]string `json:"component_mappings"`

	Description string `json:"description" binding:"max=1000"`
}

func extractEnterpriseNumberWithRegex(oid string) (string, error) {
	// 查找匹配
	matches := constants.EnterpriseOIDRegex.FindStringSubmatch(oid)
	if len(matches) < 2 {
		return "", errors.New("不是标准的企业OID格式")
	}

	// matches[0] 是整个匹配，matches[1] 是第一个捕获组（企业编号）
	return matches[1], nil
}

func (req *BMCTrapParserCreate) extractAndValidateVendorCode() (string, error) {
	// 从四个OID中提取厂商编码，确保它们一致
	oids := []string{
		req.AlertLevelOID,
		req.AlertContentOID,
		req.AlertTimeOID,
		req.AlertComponentOID,
	}

	var vendorCodes []string

	// 从每个OID中尝试解析厂商编码
	for _, oid := range oids {
		vendorCode, err := extractEnterpriseNumberWithRegex(oid)
		if err != nil {
			return "", fmt.Errorf("解析oid:%s失败, err:%v", oid, err)
		}
		vendorCodes = append(vendorCodes, vendorCode)
	}
	// 检查所有OID解析出的厂商编码是否一致
	firstVendorCode := vendorCodes[0]
	for i, code := range vendorCodes {
		if code != firstVendorCode {
			return "", fmt.Errorf("OID厂商编码不一致: %s(%s) != %s(%s)",
				oids[0], firstVendorCode, oids[i], code)
		}
	}

	return firstVendorCode, nil
}

func (req *BMCTrapParserCreate) ToParserRecord() (*models.BMCTrapParser, error) {
	// 提取厂商编码
	vc, err := req.extractAndValidateVendorCode()
	if err != nil {
		return nil, err
	}

	parser := &models.BMCTrapParser{
		ParserName: strings.TrimSpace(req.ParserName),
		VendorName: strings.TrimSpace(req.VendorName),
		VendorCode: vc,

		AlertLevelOID:     strings.TrimSpace(req.AlertLevelOID),
		AlertContentOID:   strings.TrimSpace(req.AlertContentOID),
		AlertTimeOID:      strings.TrimSpace(req.AlertTimeOID),
		AlertComponentOID: strings.TrimSpace(req.AlertComponentOID),

		EnableAutoClose: req.EnableAutoClose,
		AlertIndexOID:   strings.TrimSpace(req.AlertIndexOID),
		AlertStatusOID:  strings.TrimSpace(req.AlertStatusOID),

		EnableContactInterComponentAlerts:  req.EnableContactInterComponentAlerts,
		ContactInterComponentIdentifierOID: strings.TrimSpace(req.ContactInterComponentIdentifierOID),

		Description: strings.TrimSpace(req.Description),
	}

	// 处理映射字段
	parser.LevelMappings = make(map[string]models.AlertLevel)
	for k, v := range req.LevelMappings {
		parser.LevelMappings[strings.TrimSpace(k)] = models.AlertLevel(strings.TrimSpace(v))
	}

	parser.StatusMappings = make(map[string]models.TrapStatus)
	for k, v := range req.StatusMappings {
		parser.StatusMappings[strings.TrimSpace(k)] = models.TrapStatus(strings.TrimSpace(v))
	}

	if req.ComponentMappings != nil {
		parser.ComponentMappings = make(map[string][]string)
		for k, v := range req.ComponentMappings {
			cleanKey := strings.TrimSpace(k)
			if cleanKey == "" {
				continue
			}
			parser.ComponentMappings[cleanKey] = trimStringSlice(v)
		}
	}

	if req.EnableProductNameList != nil {
		parser.EnableProductNameList = trimStringSlice(req.EnableProductNameList)
	}

	if req.EnableHostNameList != nil {
		parser.EnableHostNameList = trimStringSlice(req.EnableHostNameList)
	}

	return parser, nil

}

func (rc *BMCTrapParserCreate) FromBMCTrapParser(parser *models.BMCTrapParser) {
	rc.ParserName = parser.ParserName
	rc.VendorName = parser.VendorName
	rc.AlertLevelOID = parser.AlertLevelOID
	rc.AlertContentOID = parser.AlertContentOID
	rc.AlertTimeOID = parser.AlertTimeOID
	rc.AlertComponentOID = parser.AlertComponentOID
	rc.EnableAutoClose = parser.EnableAutoClose
	rc.AlertIndexOID = parser.AlertIndexOID
	rc.AlertStatusOID = parser.AlertStatusOID
	rc.EnableContactInterComponentAlerts = parser.EnableContactInterComponentAlerts
	rc.ContactInterComponentIdentifierOID = parser.ContactInterComponentIdentifierOID
	rc.Description = parser.Description

	if parser.LevelMappings != nil {
		rc.LevelMappings = make(map[string]string)
		for k, v := range parser.LevelMappings {
			rc.LevelMappings[k] = string(v)
		}
	} else {
		rc.LevelMappings = nil
	}

	if parser.StatusMappings != nil {
		rc.StatusMappings = make(map[string]string)
		for k, v := range parser.StatusMappings {
			rc.StatusMappings[k] = string(v)
		}
	} else {
		rc.StatusMappings = nil
	}

	rc.ComponentMappings = parser.ComponentMappings
	rc.EnableProductNameList = parser.EnableProductNameList
	rc.EnableHostNameList = parser.EnableHostNameList

}

// BMCTrapParserUpdate 更新解析器请求
type BMCTrapParserUpdate struct {
	ID int64 `json:"id" binding:"required"`

	ParserName *string `json:"parser_name,omitempty" binding:"omitempty,max=200"`
	VendorName *string `json:"vendor_name,omitempty" binding:"omitempty,max=200"`

	// 核心OID字段 - 使用指针表示可选更新
	AlertLevelOID     *string `json:"alert_level_oid,omitempty" binding:"omitempty,max=500"`
	AlertContentOID   *string `json:"alert_content_oid,omitempty" binding:"omitempty,max=500"`
	AlertTimeOID      *string `json:"alert_time_oid,omitempty" binding:"omitempty,max=500"`
	AlertComponentOID *string `json:"alert_component_oid,omitempty" binding:"omitempty,max=500"`

	EnableAutoClose *bool   `json:"enable_auto_close,omitempty"`
	AlertIndexOID   *string `json:"alert_index_oid,omitempty" binding:"omitempty,max=500"`
	AlertStatusOID  *string `json:"alert_status_oid,omitempty" binding:"omitempty,max=500"`

	EnableContactInterComponentAlerts  *bool   `json:"enable_contact_inter_component_alerts,omitempty"`
	ContactInterComponentIdentifierOID *string `json:"contact_inter_component_identifier_oid,omitempty" binding:"omitempty,max=500"`

	TimeFormat *string `json:"time_format,omitempty" binding:"omitempty,max=200"`

	// 映射配置 - 使用指针表示可选更新
	LevelMappings         *map[string]string   `json:"level_mappings,omitempty"`
	StatusMappings        *map[string]string   `json:"status_mappings,omitempty"`
	EnableProductNameList *[]string            `json:"enable_product_name_list,omitempty"`
	EnableHostNameList    *[]string            `json:"enable_host_name_list,omitempty"`
	ComponentMappings     *map[string][]string `json:"component_mappings,omitempty"`

	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
}

func (ru *BMCTrapParserUpdate) UpdateParserRecord(parser *models.BMCTrapParser) {

	if ru.ParserName != nil {
		parser.ParserName = *ru.ParserName
	}
	if ru.VendorName != nil {
		parser.VendorName = *ru.VendorName
	}

	if ru.AlertLevelOID != nil {
		parser.AlertLevelOID = *ru.AlertLevelOID
	}
	if ru.AlertContentOID != nil {
		parser.AlertContentOID = *ru.AlertContentOID
	}
	if ru.AlertTimeOID != nil {
		parser.AlertTimeOID = *ru.AlertTimeOID
	}
	if ru.AlertComponentOID != nil {
		parser.AlertComponentOID = *ru.AlertComponentOID
	}

	if ru.EnableAutoClose != nil {
		parser.EnableAutoClose = *ru.EnableAutoClose
	}
	if ru.AlertIndexOID != nil {
		parser.AlertIndexOID = *ru.AlertIndexOID
	}
	if ru.AlertStatusOID != nil {
		parser.AlertStatusOID = *ru.AlertStatusOID
	}

	if ru.EnableContactInterComponentAlerts != nil {
		parser.EnableContactInterComponentAlerts = *ru.EnableContactInterComponentAlerts
	}
	if ru.ContactInterComponentIdentifierOID != nil {
		parser.ContactInterComponentIdentifierOID = *ru.ContactInterComponentIdentifierOID
	}

	if ru.LevelMappings != nil {
		parser.LevelMappings = make(map[string]models.AlertLevel)
		for k, v := range *ru.LevelMappings {
			parser.LevelMappings[strings.TrimSpace(k)] = models.AlertLevel(strings.TrimSpace(v))
		}
	}
	if ru.StatusMappings != nil {
		parser.StatusMappings = make(map[string]models.TrapStatus)
		for k, v := range *ru.StatusMappings {
			parser.StatusMappings[strings.TrimSpace(k)] = models.TrapStatus(strings.TrimSpace(v))
		}
	}

	if ru.EnableProductNameList != nil {
		parser.EnableProductNameList = *ru.EnableProductNameList
	}
	if ru.EnableHostNameList != nil {
		parser.EnableHostNameList = *ru.EnableHostNameList
	}
	if ru.ComponentMappings != nil {
		parser.ComponentMappings = *ru.ComponentMappings
	}

	if ru.Description != nil {
		parser.Description = *ru.Description
	}
}

type BMCTrapParserFilter struct {
	// 统一使用 json tag，因为数据从 Body 中解析，如果前端传了就用传的值，没传就用 default
	Page     int `json:"page" binding:"min=1" default:"1"`
	PageSize int `json:"page_size" binding:"min=1,max=100" default:"10"`

	// 查询条件：使用指针。如果前端不传，指针为 nil，后端可跳过该条件
	ParserName  *string `json:"parser_name,omitempty"` // 模糊匹配 (nil或空字符串时忽略)
	VendorCode  *string `json:"vendor_code,omitempty"` // 精确匹配 (nil或空字符串时忽略)
	VendorName  *string `json:"vendor_name,omitempty"` // 模糊匹配 (nil或空字符串时忽略)
	IsActive    *bool   `json:"is_active,omitempty"`   // 精确匹配 (nil时忽略)
	Description *string `json:"description,omitempty"` // 模糊匹配 (nil或空字符串时忽略)
}

func trimStringSlice(slice []string) []string {
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
