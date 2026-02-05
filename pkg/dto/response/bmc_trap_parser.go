package response

import (
	"baize-monitor/pkg/models"
	"time"
)

// BMCTrapParserResponse 解析器详情响应
type BMCTrapParserResponse struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ParserName string `json:"parser_name"`
	VendorCode string `json:"vendor_code"`
	VendorName string `json:"vendor_name"`

	// 核心OID字段
	AlertLevelOID     string `json:"alert_level_oid"`
	AlertContentOID   string `json:"alert_content_oid"`
	AlertTimeOID      string `json:"alert_time_oid"`
	AlertComponentOID string `json:"alert_component_oid"`

	EnableAutoClose bool   `json:"enable_auto_close"`
	AlertIndexOID   string `json:"alert_index_oid,omitempty"`
	AlertStatusOID  string `json:"alert_status_oid,omitempty"`

	EnableContactInterComponentAlerts  bool   `json:"enable_contact_inter_component_alerts"`
	ContactInterComponentIdentifierOID string `json:"contact_inter_component_identifier_oid,omitempty"`

	// 映射配置
	LevelMappings         map[string]string   `json:"level_mappings"`
	StatusMappings        map[string]string   `json:"status_mappings"`
	EnableProductNameList []string            `json:"enable_product_name_list"`
	EnableHostNameList    []string            `json:"enable_host_name_list"`
	ComponentMappings     map[string][]string `json:"component_mappings"`

	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

func (res *BMCTrapParserResponse) FromParserDao(p *models.BMCTrapParser) *BMCTrapParserResponse {
	res.ID = p.ID
	res.CreatedAt = p.CreatedAt
	res.UpdatedAt = p.UpdatedAt

	res.ParserName = p.ParserName
	res.VendorName = p.VendorName
	res.VendorCode = p.VendorCode

	res.AlertLevelOID = p.AlertLevelOID
	res.AlertContentOID = p.AlertContentOID
	res.AlertTimeOID = p.AlertTimeOID
	res.AlertComponentOID = p.AlertComponentOID

	res.EnableAutoClose = p.EnableAutoClose
	res.AlertIndexOID = p.AlertIndexOID
	res.AlertStatusOID = p.AlertStatusOID

	res.EnableContactInterComponentAlerts = p.EnableContactInterComponentAlerts
	res.ContactInterComponentIdentifierOID = p.ContactInterComponentIdentifierOID

	res.LevelMappings = make(map[string]string)
	for k, v := range p.LevelMappings {
		res.LevelMappings[k] = string(v)
	}
	res.StatusMappings = make(map[string]string)
	for k, v := range p.StatusMappings {
		res.StatusMappings[k] = string(v)
	}

	res.EnableProductNameList = p.EnableProductNameList
	res.EnableHostNameList = p.EnableHostNameList
	res.ComponentMappings = p.ComponentMappings

	res.Description = p.Description
	res.IsActive = p.IsActive

	return res
}

// BMCTrapParserCreateResponse 创建解析器响应
type BMCTrapParserCreateResponse struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
}

type BMCTrapParserActiveResponse struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
}

type BMCTrapParserDeactiveResponse struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
}

// BMCTrapParserUpdateResponse 更新解析器响应
type BMCTrapParserUpdateResponse struct {
	Message string `json:"message"`
}

// BMCTrapParserDeleteResponse 删除解析器响应
type BMCTrapParserDeleteResponse struct {
	Message string `json:"message"`
}

type BMCTrapParserListResult struct {
	List     []*BMCTrapParserResponse `json:"list"`
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
}
