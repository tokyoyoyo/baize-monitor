package storage

import (
	"context"
	"time"
)

// DistributedLocker interface for distributed locking mechanisms
type DistributedLockerInterface interface {
	// AcquireLock acquires a distributed lock for a given key
	AcquireLock(ctx context.Context, key string, expiration time.Duration) (bool, error)
	// ReleaseLock release the lock
	ReleaseLock(ctx context.Context, key string) error
	// GenerateTrapLockKey generate lock key for trap data
	GenerateTrapLockKey(trapData []byte) string
	// Close close redis client
	Close() error
}
