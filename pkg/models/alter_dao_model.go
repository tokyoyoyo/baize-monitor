package models

import "time"

// Alert 告警实例模型
type Alert struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 基础信息
	TrapOID    string `json:"trap_oid" gorm:"size:200;index"`
	SourceIP   string `json:"source_ip" gorm:"size:45;index"`
	VendorCode string `json:"vendor_code" gorm:"size:50;index"`
	VendorName string `json:"vendor_name" gorm:"size:100"`

	// 标准化后的告警信息
	AlertLevel  AlertLevel `json:"alert_level" gorm:"size:100;index"`
	AlertTime   string     `json:"alert_time" gorm:"size:100;index"`
	Component   string     `json:"component" gorm:"size:100;index"`
	ComponentID string     `json:"component_id" gorm:"size:100;index"` // 硬件序列号等标识
	Content     string     `json:"content" gorm:"type:text"`

	// 原始信息（用于调试和审计）
	RawData  []byte `json:"raw_data" gorm:"type:text"`
	ParserID int64  `json:"parser_id" gorm:"index"`

	EnableAutoClose bool        `json:"enable_auto_close"`                  // 是否启用自动关闭
	AlertIndex      string      `json:"alert_index" gorm:"size:200;index"`  // 告警索引OID
	AlertStatus     AlertStatus `json:"alert_status" gorm:"size:200;index"` // 告警状态OID

	EnableContactInterComponentAlerts bool   `json:"enable_contact_inter_component_alerts"`
	IdentifierOfTheSameComponent      string `json:"identifier_of_the_same_component" gorm:"size:200;index"`

	// 关联信息
	MachineSN   string `json:"machine_sn" gorm:"size:100;index"`
	HostName    string `json:"host_name" gorm:"size:100;index"`
	ProductName string `json:"product_name" gorm:"size:100;index"`
}
