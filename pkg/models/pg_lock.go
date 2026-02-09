package models

import "time"

// DistributedLockRecord represents a lock record in the database
type DistributedLockRecord struct {
	Key        string    `gorm:"primaryKey;column:key"`
	Holder     string    `gorm:"column:holder"`      // Lock holder identifier
	AcquiredAt time.Time `gorm:"column:acquired_at"` // Lock acquisition time
	ExpiresAt  time.Time `gorm:"column:expires_at"`  // Lock expiration time
	UpdatedAt  time.Time `gorm:"column:updated_at"`  // Last update time
}

// TableName specifies the table name for the lock records
func (DistributedLockRecord) TableName() string {
	return "distributed_locks"
}
