package models

import (
	"time"
)

type AnomalyResult struct {
	ID        int64     `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`

	TargetIP string `json:"target_ip" gorm:"column:target_ip;size:45;index;not null"`

	CheckType string `json:"check_type" gorm:"column:check_type;size:50;index;not null"`
	CheckItem string `json:"check_item" gorm:"column:check_item;size:100;index;not null"`
	Success   bool   `json:"success" gorm:"column:success;index;not null"`
	Level     string `json:"level" gorm:"column:level;size:50;index"`
	Message   string `json:"message" gorm:"column:message;type:text"`

	ExtraData map[string]interface{} `json:"extra_data" gorm:"column:extra_data;type:jsonb;not null;default:'{}';serializer:json"`

	CheckedAt time.Time `json:"checked_at" gorm:"column:checked_at;index"`
}

func (AnomalyResult) TableName() string {
	return "anomaly_results"
}

type AnomalyLevel string

const (
	AnomalyLevelInfo     AnomalyLevel = "info"
	AnomalyLevelWarning  AnomalyLevel = "warning"
	AnomalyLevelCritical AnomalyLevel = "critical"
)

type CheckType string

const (
	CheckTypeCPU     CheckType = "cpu"
	CheckTypeMemory  CheckType = "memory"
	CheckTypeDisk    CheckType = "disk"
	CheckTypeNetwork CheckType = "network"
	CheckTypeProcess CheckType = "process"
)
