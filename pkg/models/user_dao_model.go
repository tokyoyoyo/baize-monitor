package models

import (
	"time"
)

// User 用户模型
type User struct {
	ID           int64           `json:"id" gorm:"primaryKey;autoIncrement" column:"id"`
	Username     string          `json:"username" gorm:"uniqueIndex;not null" column:"username"`
	PasswordHash string          `json:"-" gorm:"not null" column:"password_hash"`
	IsAdmin      bool            `json:"is_admin" gorm:"default:false" column:"is_admin"`
	IsActive     bool            `json:"is_active" gorm:"default:true" column:"is_active"`
	IsDeleted    bool            `json:"is_deleted" gorm:"default:false" column:"is_deleted"`
	CreatedAt    time.Time       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    time.Time       `json:"updated_at" gorm:"column:updated_at"`
	Permissions  map[string]bool `json:"permissions" gorm:"type:jsonb;not null;default:'{}';serializer:json" column:"permissions"` // 存储功能模块权限的JSONB字段
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}
