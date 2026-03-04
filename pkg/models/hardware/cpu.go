package hardware

import (
	request "baize-monitor/pkg/dto/request/hardware"
	"net"
	"time"
)

// CPU CPU 信息表模型（外层结构，用于采集状态管理）
type CPU struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`

	// 基本信息 - 使用 TargetIP 关联 machine_info 表
	TargetIP net.IP `json:"target_ip" gorm:"column:target_ip;type:inet;index;not null;"`
	// 执行状态
	Success bool   `json:"success" gorm:"column:success;not null;default:true;"` // 采集是否成功
	Message string `json:"message" gorm:"column:message;type:text;"`             // 执行消息（失败时记录错误信息）

	CPUInfo []request.CPUInfo `json:"cpu_info" gorm:"type:jsonb;serializer:json;column:cpu_info;"` // CPU 信息列表，使用 JSONB 存储

	// 摘要信息
	TotalCores   int `json:"total_cores" gorm:"column:total_cores;not null;default:0;"`     // 总物理核心数
	TotalThreads int `json:"total_threads" gorm:"column:total_threads;not null;default:0;"` // 总逻辑线程数
}

// TableName 设置表名
func (CPU) TableName() string {
	return "hardware_cpu"
}
