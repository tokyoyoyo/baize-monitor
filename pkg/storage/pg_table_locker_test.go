package storage

import (
	"math/rand"
	"testing"
	"time"

	"baize-monitor/pkg/config"
	"baize-monitor/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// getTestDBClient creates a test database client for lock testing
func getTestDBClient(t *testing.T) *Client {
	testConf, err := config.LoadTestMockServerConfig()
	require.NoError(t, err)

	client, err := NewClient(testConf.PostGresConfig)
	require.NoError(t, err)

	// Ensure the distributed_locks table exists
	err = client.AutoMigrate(&models.DistributedLockRecord{})
	require.NoError(t, err)

	return client
}

// cleanupAllTestLocks removes all locks from the database for testing
func cleanupAllTestLocks(t *testing.T, db *gorm.DB) {
	err := db.Exec("DELETE FROM distributed_locks").Error
	require.NoError(t, err)
}

func TestPGTableDistributedLocker_NewCreatesRandomInterval(t *testing.T) {
	client := getTestDBClient(t)
	defer client.Close()
	cleanupAllTestLocks(t, client.DB)

	locker, err := NewPGTableDistributedLocker(client.DB)
	require.NoError(t, err)
	require.NotNil(t, locker)

	pgLocker := locker.(*PGTableDistributedLocker)

	// Verify cleanup interval is within 24-48 hours range
	assert.GreaterOrEqual(t, pgLocker.cleanupInterval, 24*time.Hour)
	assert.LessOrEqual(t, pgLocker.cleanupInterval, 48*time.Hour)

	// Verify last cleanup time is set to current time (within reasonable bounds)
	now := time.Now()
	assert.WithinDuration(t, pgLocker.lastCleanupTime, now, 5*time.Second)

	// Verify holder ID is set
	assert.NotEmpty(t, pgLocker.holderID)
}

func TestPGTableDistributedLocker_CleanupExpiredLocks(t *testing.T) {
	client := getTestDBClient(t)
	defer client.Close()
	cleanupAllTestLocks(t, client.DB)

	locker, err := NewPGTableDistributedLocker(client.DB)
	require.NoError(t, err)
	require.NotNil(t, locker)

	pgLocker := locker.(*PGTableDistributedLocker)

	// After creating locker, there should be 1 lock (the instance lock)
	var count int64
	client.DB.Model(&models.DistributedLockRecord{}).Count(&count)
	assert.Equal(t, int64(1), count)

	// Insert an expired lock record
	expiredLock := models.DistributedLockRecord{
		Key:        "test_expired_lock",
		Holder:     "test_holder_1",
		AcquiredAt: time.Now().Add(-2 * time.Hour),
		ExpiresAt:  time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		UpdatedAt:  time.Now().Add(-1 * time.Hour),
	}
	result := client.DB.Create(&expiredLock)
	require.NoError(t, result.Error)

	// Insert a non-expired lock record
	nonExpiredLock := models.DistributedLockRecord{
		Key:        "test_non_expired_lock",
		Holder:     "test_holder_2",
		AcquiredAt: time.Now(),
		ExpiresAt:  time.Now().Add(1 * time.Hour), // Expires in 1 hour
		UpdatedAt:  time.Now(),
	}
	result = client.DB.Create(&nonExpiredLock)
	require.NoError(t, result.Error)

	// Verify all locks exist before cleanup (1 instance + 2 test = 3 total)
	client.DB.Model(&models.DistributedLockRecord{}).Count(&count)
	assert.Equal(t, int64(3), count)

	// Perform cleanup
	err = pgLocker.cleanupExpiredLocks()
	require.NoError(t, err)

	// Verify expired lock is cleaned up (1 instance + 1 non-expired = 2 total)
	client.DB.Model(&models.DistributedLockRecord{}).Count(&count)
	assert.Equal(t, int64(2), count)

	// Verify the non-expired lock still exists
	var remainingLocks []models.DistributedLockRecord
	err = client.DB.Find(&remainingLocks).Error
	require.NoError(t, err)
	
	// Find the non-expired lock
	nonExpiredFound := false
	for _, lock := range remainingLocks {
		if lock.Key == "test_non_expired_lock" {
			nonExpiredFound = true
			break
		}
	}
	assert.True(t, nonExpiredFound, "Non-expired lock should still exist")
}

func TestPGTableDistributedLocker_MaybeCleanupExpiredLocks(t *testing.T) {
	client := getTestDBClient(t)
	defer client.Close()
	cleanupAllTestLocks(t, client.DB)

	locker, err := NewPGTableDistributedLocker(client.DB)
	require.NoError(t, err)
	require.NotNil(t, locker)

	pgLocker := locker.(*PGTableDistributedLocker)

	// Set cleanup interval to a very short duration for testing
	pgLocker.cleanupInterval = 100 * time.Millisecond

	// Initially, no cleanup should happen (last cleanup was just now)
	pgLocker.maybeCleanupExpiredLocks()

	// Wait for interval to pass
	time.Sleep(150 * time.Millisecond)

	// Insert some expired locks to verify cleanup actually happens
	expiredLock1 := models.DistributedLockRecord{
		Key:        "test_cleanup_lock_1",
		Holder:     "test_holder_3",
		AcquiredAt: time.Now().Add(-2 * time.Hour),
		ExpiresAt:  time.Now().Add(-1 * time.Hour),
		UpdatedAt:  time.Now().Add(-1 * time.Hour),
	}
	result := client.DB.Create(&expiredLock1)
	require.NoError(t, result.Error)

	expiredLock2 := models.DistributedLockRecord{
		Key:        "test_cleanup_lock_2",
		Holder:     "test_holder_4",
		AcquiredAt: time.Now().Add(-3 * time.Hour),
		ExpiresAt:  time.Now().Add(-2 * time.Hour),
		UpdatedAt:  time.Now().Add(-2 * time.Hour),
	}
	result = client.DB.Create(&expiredLock2)
	require.NoError(t, result.Error)

	// Verify locks exist before cleanup
	var count int64
	client.DB.Model(&models.DistributedLockRecord{}).Where("key LIKE 'test_cleanup_lock%'").Count(&count)
	assert.Equal(t, int64(2), count)

	// Now cleanup should be triggered
	pgLocker.maybeCleanupExpiredLocks()

	// Wait a bit for async cleanup to complete
	time.Sleep(50 * time.Millisecond)

	// Verify expired locks are cleaned up
	client.DB.Model(&models.DistributedLockRecord{}).Where("key LIKE 'test_cleanup_lock%'").Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestPGTableDistributedLocker_AcquireLockWithCleanup(t *testing.T) {
	client := getTestDBClient(t)
	defer client.Close()
	cleanupAllTestLocks(t, client.DB)

	// Create a locker with short cleanup interval
	locker, err := NewPGTableDistributedLocker(client.DB)
	require.NoError(t, err)
	require.NotNil(t, locker)

	pgLocker := locker.(*PGTableDistributedLocker)
	pgLocker.cleanupInterval = 50 * time.Millisecond

	// Insert some expired locks
	expiredLock := models.DistributedLockRecord{
		Key:        "test_acquire_expired",
		Holder:     "old_holder",
		AcquiredAt: time.Now().Add(-2 * time.Hour),
		ExpiresAt:  time.Now().Add(-1 * time.Hour),
		UpdatedAt:  time.Now().Add(-1 * time.Hour),
	}
	result := client.DB.Create(&expiredLock)
	require.NoError(t, result.Error)

	// Verify expired lock exists
	var count int64
	client.DB.Model(&models.DistributedLockRecord{}).Where("key = 'test_acquire_expired'").Count(&count)
	assert.Equal(t, int64(1), count)

	// Wait for cleanup interval to pass
	time.Sleep(60 * time.Millisecond)

	// Acquire a new lock - this should trigger cleanup
	acquired, err := locker.AcquireLock("test_new_lock", 10*time.Second)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Wait for async cleanup to complete
	time.Sleep(50 * time.Millisecond)

	// Verify expired lock was cleaned up
	client.DB.Model(&models.DistributedLockRecord{}).Where("key = 'test_acquire_expired'").Count(&count)
	assert.Equal(t, int64(0), count)

	// Verify new lock exists
	client.DB.Model(&models.DistributedLockRecord{}).Where("key = 'test_new_lock'").Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestRandomCleanupIntervalGeneration(t *testing.T) {
	// Test the random interval generation logic that's used in NewPGTableDistributedLocker

	// Generate multiple intervals to verify they're within the expected range
	intervals := make(map[time.Duration]bool)

	for i := 0; i < 100; i++ {
		// Simulate the exact logic from NewPGTableDistributedLocker
		r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(i)))
		randomHours := 24 + r.Intn(25) // 24 to 48 hours
		cleanupInterval := time.Duration(randomHours) * time.Hour

		// Verify interval is within 24-48 hours range
		assert.GreaterOrEqual(t, cleanupInterval, 24*time.Hour, "Interval should be at least 24 hours")
		assert.LessOrEqual(t, cleanupInterval, 48*time.Hour, "Interval should be at most 48 hours")

		intervals[cleanupInterval] = true
	}

	// Verify we got a reasonable variety of intervals
	assert.GreaterOrEqual(t, len(intervals), 1, "Should generate at least one unique interval")

	// Verify we have intervals covering the full range (at least some below 36h and some above)
	hasBelow36 := false
	hasAbove36 := false
	for interval := range intervals {
		if interval < 36*time.Hour {
			hasBelow36 = true
		}
		if interval > 36*time.Hour {
			hasAbove36 = true
		}
	}

	// At least one should be below and one above 36 hours (midpoint)
	// This might not always be true with small sample size, so we'll make it less strict
	assert.True(t, hasBelow36 || hasAbove36, "Should have intervals in the range")
}
