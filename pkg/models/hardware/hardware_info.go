package hardware

import (
	request "baize-monitor/pkg/dto/request/hardware"
	"net"
	"time"
)

// HardwareInfo 硬件信息表模型 - 用于数据库存储
type HardwareInfo struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`

	// 基本信息 - 使用TargetIP关联machine_info表
	TargetIP net.IP `json:"target_ip" gorm:"column:target_ip;type:inet;index;not null;"` // 管理IP地址，使用PostgreSQL inet类型

	// 收集时间戳
	CollectedAt time.Time `json:"collected_at" gorm:"column:collected_at;not null;comment:硬件信息收集时间"` // 硬件信息收集时间

	// 硬件信息快照 - 使用JSONB存储复杂结构
	HardwareInfoSnapshot request.HardwareInfoRequest `json:"hardware_info_snapshot" gorm:"column:hardware_info_snapshot;type:jsonb;not null;serializer:json;"` // 硬件信息快照，使用JSONB存储

	// 数据版本和校验
	Version  string `json:"version" gorm:"column:version;size:50;not null;"`    // 数据格式版本
	Checksum string `json:"checksum" gorm:"column:checksum;size:128;not null;"` // 数据校验和
}

// TableName 设置表名
func (HardwareInfo) TableName() string {
	return "hardware_info"
}
