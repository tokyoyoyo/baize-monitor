package hardware

import (
	request "baize-monitor/pkg/dto/request/hardware"
	"net"
	"time"
)

// NetworkInterface 网络接口信息表模型（外层结构，用于采集状态管理）
type NetworkInterface struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`

	// 基本信息 - 使用 TargetIP 关联 machine_info 表
	TargetIP net.IP `json:"target_ip" gorm:"column:target_ip;type:inet;index;not null"` // 管理 IP 地址，使用 PostgreSQL inet 类型

	// 执行状态
	Success bool   `json:"success" gorm:"column:success;not null;default:true"` // 采集是否成功
	Message string `json:"message" gorm:"column:message;type:text"`             // 执行消息（失败时记录错误信息）

	Content []request.NetworkInterfaceInfo `json:"content" gorm:"column:content;type:jsonb;serializer:json;not null"` // 网络接口详细信息列表

	// 摘要信息
	TotalCount    int `json:"total_count" gorm:"column:total_count"`       // 接口总数
	PhysicalCount int `json:"physical_count" gorm:"column:physical_count"` // 物理接口数
	VirtualCount  int `json:"virtual_count" gorm:"column:virtual_count"`   // 虚拟接口数
	ActiveCount   int `json:"active_count" gorm:"column:active_count"`     // 活跃接口数
}

// TableName 设置表名
func (NetworkInterface) TableName() string {
	return "hardware_network_interface"
}
