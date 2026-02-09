package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"gorm.io/gorm"
)

// DistributedLocker interface for distributed locking mechanisms
type DistributedLockerInterface interface {
	// AcquireLock acquires a distributed lock for a given key
	AcquireLock(key string, expiration time.Duration) (bool, error)
	// GenerateTrapLockKey generate lock key for trap data
	GenerateTrapLockKey(trapData []byte) string
}

// PGTableDistributedLocker DistributedLocker based on PostgreSQL table
type PGTableDistributedLocker struct {
	db       *gorm.DB
	holderID string
	// lastCleanupTime tracks when the last cleanup was performed
	lastCleanupTime time.Time
	cleanupMutex    sync.Mutex // Protects lastCleanupTime updates
	// cleanupInterval is the random interval between 24-48 hours for cleanup
	cleanupInterval time.Duration
}

// NewPGTableDistributedLocker creates a new PostgreSQL table-based distributed locker
func NewPGTableDistributedLocker(db *gorm.DB) (DistributedLockerInterface, error) {
	// Generate a unique holder ID for this instance
	holderID := fmt.Sprintf("instance_%d_%d", time.Now().Unix(), time.Now().Nanosecond())

	// Create a new random source with current time as seed
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	randomHours := 24 + r.Intn(25)
	cleanupInterval := time.Duration(randomHours) * time.Hour

	locker := &PGTableDistributedLocker{
		db:              db,
		holderID:        holderID,
		lastCleanupTime: time.Now(), // Initialize with current time
		cleanupInterval: cleanupInterval,
	}

	_, err := locker.AcquireLock(holderID, 10*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock for instance: %w", err)
	}

	return locker, nil
}

// cleanupExpiredLocks removes all expired locks from the table
func (p *PGTableDistributedLocker) cleanupExpiredLocks() error {
	now := time.Now()
	result := p.db.Exec(`
		DELETE FROM distributed_locks 
		WHERE expires_at <= $1
	`, now)

	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired locks: %w", result.Error)
	}

	return nil
}

// maybeCleanupExpiredLocks performs lazy cleanup of expired locks if needed
// This is called during AcquireLock to avoid blocking the main operation
func (p *PGTableDistributedLocker) maybeCleanupExpiredLocks() {
	p.cleanupMutex.Lock()
	defer p.cleanupMutex.Unlock()

	// Check if the cleanup interval has passed since last cleanup
	if time.Since(p.lastCleanupTime) > p.cleanupInterval {
		// Update last cleanup time immediately to prevent multiple concurrent cleanups
		p.lastCleanupTime = time.Now()

		// Start a goroutine to perform the cleanup asynchronously
		go func() {
			if err := p.cleanupExpiredLocks(); err != nil {
				// Log the error but don't fail the main operation
				// In a real implementation, you would use a proper logger
				fmt.Printf("Warning: failed to cleanup expired locks: %v\n", err)
			}
		}()
	}
}

// AcquireLock acquires a distributed lock for a given key with automatic expiration
// This implementation includes lazy cleanup of expired locks
func (p *PGTableDistributedLocker) AcquireLock(key string, expiration time.Duration) (bool, error) {
	// Perform lazy cleanup of expired locks if needed
	p.maybeCleanupExpiredLocks()

	now := time.Now()
	expiresAt := now.Add(expiration)

	// Try to insert a new lock record or update an expired one
	result := p.db.Exec(`
		INSERT INTO distributed_locks (key, holder, acquired_at, expires_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (key) DO UPDATE
		SET holder = $2, acquired_at = $3, expires_at = $4, updated_at = $5
		WHERE distributed_locks.expires_at <= $3 OR distributed_locks.holder = $2
	`, key, p.holderID, now, expiresAt, now)

	if result.Error != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", result.Error)
	}

	// Check if we actually acquired the lock
	return result.RowsAffected > 0, nil
}

// GenerateTrapLockKey generates lock key for trap data (same as Redis implementation)
func (p *PGTableDistributedLocker) GenerateTrapLockKey(trapData []byte) string {
	hash := sha256.Sum256(trapData)
	return "trap_lock:" + hex.EncodeToString(hash[:])
}
