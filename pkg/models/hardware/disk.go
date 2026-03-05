package hardware

import (
	"net"
	"time"
)

// Disk 磁盘信息表模型（外层结构，用于采集状态管理）
type Disk struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`

	// 基本信息 - 使用 TargetIP 关联 machine_info 表
	TargetIP net.IP `json:"target_ip" gorm:"column:target_ip;type:inet;index;not null;"` // 管理 IP 地址，使用 PostgreSQL inet 类型

	// 执行状态
	Success bool   `json:"success" gorm:"column:success;not null;default:true;"` // 采集是否成功
	Message string `json:"message" gorm:"column:message;type:text;"`             // 执行消息（失败时记录错误信息）

	// 相关字段，直接存结构体，会导致检索不方便

	// 摘要信息
	TotalCount      int   `json:"total_count" gorm:"column:total_count;not null;default:0;"`           // 磁盘总数
	TotalSize       int64 `json:"total_size" gorm:"column:total_size;not null;default:0;"`             // 总容量（字节）
	SSDCount        int   `json:"ssd_count" gorm:"column:ssd_count;not null;default:0;"`               // SSD 数量
	HDDCount        int   `json:"hdd_count" gorm:"column:hdd_count;not null;default:0;"`               // HDD 数量
	TotalPartitions int   `json:"total_partitions" gorm:"column:total_partitions;not null;default:0;"` // 分区总数
}

// TableName 设置表名
func (Disk) TableName() string {
	return "hardware_disk"
}
