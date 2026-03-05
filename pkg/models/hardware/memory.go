package hardware

import (
	"net"
	"time"
)

// Memory 内存信息表模型（外层结构，用于采集状态管理）
type Memory struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`

	// 基本信息 - 使用 TargetIP 关联 machine_info 表
	TargetIP net.IP `json:"target_ip" gorm:"column:target_ip;type:inet;index;not null"` // 管理 IP 地址，使用 PostgreSQL inet 类型

	// 执行状态
	Success bool   `json:"success" gorm:"column:success;not null;default:true"` // 采集是否成功
	Message string `json:"message" gorm:"column:message;type:text"`             // 执行消息（失败时记录错误信息）

	// 相关字段，直接存结构体，会导致检索不方便

	// 摘要信息
	TotalSize  int64  `json:"total_size"`                            // 总容量（字节）
	Type       string `json:"type" gorm:"column:type;size:100"`      // 内存类型（DDR4、DDR5 等）
	Speed      string `json:"speed" gorm:"column:speed;size:100"`    // 内存频率
	TotalSlots int    `json:"total_slots" gorm:"column:total_slots"` // 插槽总数
	UsedSlots  int    `json:"used_slots" gorm:"column:used_slots"`   // 已用插槽数
}

// TableName 设置表名
func (Memory) TableName() string {
	return "hardware_memory"
}
