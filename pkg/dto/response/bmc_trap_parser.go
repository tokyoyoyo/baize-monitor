// pkg/dto/response/bmc_trap_parser.go
package response

import (
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
	AlterLevelOID     string `json:"alter_level_oid"`
	AlterContentOID   string `json:"alter_content_oid"`
	AlterTimeOID      string `json:"alter_time_oid"`
	AlterComponentOID string `json:"alter_component_oid"`

	EnableAutoClose bool   `json:"enable_auto_close"`
	AlterIndexOID   string `json:"alter_index_oid,omitempty"`
	AlterStatusOID  string `json:"alter_status_oid,omitempty"`

	EnableContactInterComponentAlters  bool   `json:"enable_contact_inter_component_alters"`
	ContactInterComponentIdentifierOID string `json:"contact_inter_component_identifier_oid,omitempty"`

	TimeFormat string `json:"time_format"`

	// 映射配置
	LevelMappings         map[string]string   `json:"level_mappings"`
	StatusMappings        map[string]string   `json:"status_mappings"`
	EnableProductNameList []string            `json:"enable_product_name_list"`
	EnableHostNameList    []string            `json:"enable_host_name_list"`
	ComponentMappings     map[string][]string `json:"component_mappings"`

	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
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
