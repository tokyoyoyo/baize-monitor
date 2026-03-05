package request

import (
	"baize-monitor/pkg/models"
	"time"
)

// MachineInfoRequest 机器信息请求DTO（极简版）
// 用于Agent向服务端上报机器信息，关联服务端上传的管理IP
type MachineInfoRequest struct {
	// 收集时间戳
	CollectedAt time.Time `json:"collected_at" binding:"required"` // 收集时间，ISO8601 格式

	// 执行状态
	Success bool   `json:"success" gorm:"column:success;not null;default:true;"` // 采集是否成功
	Message string `json:"message" gorm:"column:message;type:text;"`             // 执行消息（失败时记录错误信息）

	Content models.MachineInfo `json:"content"` // 机器信息内容

}
