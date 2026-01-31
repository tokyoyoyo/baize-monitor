package models

import "time"

// Alert 告警实例模型
type Alert struct {
	ID          int64       `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	ParserID    int64       `json:"parser_id" gorm:"index"`
	AlertStatus AlertStatus `json:"alert_status" gorm:"size:100;index"`

	// 基础信息
	TrapOID    string         `json:"trap_oid" gorm:"size:200;index"`
	SourceIP   string         `json:"source_ip" gorm:"size:45;index"`
	SourceType TrapSourceType `json:"source_type" gorm:"size:100;index"`
	VendorCode string         `json:"vendor_code" gorm:"size:50;index"`
	VendorName string         `json:"vendor_name" gorm:"size:100"`

	// 标准化后的告警信息
	AlertLevel AlertLevel `json:"alert_level" gorm:"size:100;index"`
	AlertTime  time.Time  `json:"alert_time" gorm:"index"`
	Component  string     `json:"component" gorm:"size:100;index"`
	Content    string     `json:"content" gorm:"type:text"`

	// 原始信息
	RawData     []byte            `json:"raw_data" gorm:"type:text"`
	TrapRawTime string            `json:"trap_raw_time" gorm:"size:100"`
	VariableMap map[string]string `json:"variable_map" gorm:"type:json"` // OID -> value mapping for easy access

	EnableAutoClose bool       `json:"enable_auto_close"`                 // 是否启用自动关闭
	TrapIndex       string     `json:"trap_index" gorm:"size:200;index"`  // 告警索引OID
	TrapStatus      TrapStatus `json:"trap_status" gorm:"size:200;index"` // 告警状态OID

	EnableContactInterComponentAlerts bool   `json:"enable_contact_inter_component_alerts"`
	IdentifierOfTheSameComponent      string `json:"identifier_of_the_same_component" gorm:"size:200;index"`
}

// 告警记录状态枚举类型
type AlertStatus string

const (
	AlertStatusActive  AlertStatus = "ACTIVE"
	AlertStatusCleared AlertStatus = "CLEARED"
)
