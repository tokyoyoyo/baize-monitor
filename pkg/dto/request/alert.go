package request

import (
	"baize-monitor/pkg/models"
	"time"
)

type AlertUpdate struct {
	ID     int64              `json:"id"`
	Status models.AlertStatus `json:"alert_status"`
}

// AlertFilter 定义了告警列表查询的过滤条件
type AlertFilter struct {
	// 统一使用 json tag，因为数据从 Body 中解析，如果前端传了就用传的值，没传就用 default
	Page     int `json:"page" binding:"min=1" default:"1"`               // 不传则用默认值
	PageSize int `json:"page_size" binding:"min=1,max=100" default:"10"` // 不传则用默认值

	// 查询条件：使用指针。如果前端不传，指针为 nil，后端可跳过该条件
	ParserID            *int64     `json:"parser_id,omitempty"`              // 精确匹配: 解析器ID
	Status              *string    `json:"alert_status,omitempty"`           // 精确匹配: 告警状态 (使用 string 避免值拷贝，omitempty 配合指针)
	TrapOID             *string    `json:"trap_oid,omitempty"`               // 精确匹配: 告警的OID
	TrapStatus          *string    `json:"trap_status,omitempty"`            // 精确匹配: trap报文的状态
	TrapIndex           *string    `json:"trap_index,omitempty"`             // 精确匹配: trap报文的标号
	SourceIP            *string    `json:"source_ip,omitempty"`              // 精确匹配: 告警来源IP
	VendorCode          *string    `json:"vendor_code,omitempty"`            // 精确匹配: 厂商代码
	VendorName          *string    `json:"vendor_name,omitempty"`            // 模糊匹配: 厂商名称
	AlertLevel          *string    `json:"alert_level,omitempty"`            // 精确匹配: 告警级别
	AlertTimeRangeStart *time.Time `json:"alert_time_range_start,omitempty"` // 范围查询: 告警时间开始
	AlertTimeRangeEnd   *time.Time `json:"alert_time_range_end,omitempty"`   // 范围查询: 告警时间结束
	Component           *string    `json:"component,omitempty"`              // 精确匹配: 告警组件
	Content             *string    `json:"content,omitempty"`                // 模糊匹配: 告警内容
}
