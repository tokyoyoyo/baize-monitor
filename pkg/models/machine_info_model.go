package models

import (
	"time"
)

// MachineInfo 机器信息数据模型（极简版）
// 用于存储服务端上传的管理IP对应的机器信息
type MachineInfos struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`

	CollectedAt time.Time `json:"collected_at" binding:"required"`

	// 执行状态
	Success bool   `json:"success" gorm:"column:success;not null;default:true;"` // 采集是否成功
	Message string `json:"message" gorm:"column:message;type:text;"`             // 执行消息（失败时记录错误信息）

	Content MachineInfo `json:"content" gorm:"column:content;type:jsonb;not null;default:'{}';serializer:json;"` // 机器信息内容
}

// TableName 设置表名
func (MachineInfos) TableName() string {
	return "machine_infos"
}

type MachineInfo struct {
	Hostname    string              `json:"hostname" binding:"required,min=1,max=255"`
	IPAddresses map[string][]string `json:"ip_addresses"`
}
