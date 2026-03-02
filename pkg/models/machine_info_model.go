package models

import (
	"time"
)

// MachineInfo 机器信息数据模型（极简版）
// 用于存储服务端上传的管理IP对应的机器信息
type MachineInfo struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"` 
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"` 
	
	// 基础信息
	TargetIP string `json:"target_ip" gorm:"size:45;index"` // 服务端上传的管理IP地址
	Hostname string `json:"hostname" gorm:"size:255"`     // 机器主机名
	
	// 网络信息
	IPAddresses map[string][]string `json:"ip_addresses" gorm:"type:jsonb;not null;default:'{}';serializer:json"` // 网卡名称到IP地址列表的映射，使用JSONB存储
}

// TableName 设置表名
func (MachineInfo) TableName() string {
	return "machine_info"
}