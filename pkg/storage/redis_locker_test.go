package storage

import (
	"baize-monitor/pkg/config"
	"context"
	"testing"
	"time"
)

// Redis test configuration constants
const (
	testRedisHost     = "localhost"
	testRedisPort     = 6379
	testRedisPassword = "qwer1234"
	testRedisDB       = 0
)

// getTestRedisConfig returns Redis configuration for testing
func getTestRedisConfig() *config.RedisConfig {
	return &config.RedisConfig{
		Host:     testRedisHost,
		Port:     testRedisPort,
		Password: testRedisPassword,
		DB:       testRedisDB,
	}
}

// TestRedisDistributedLocker_AcquireLock tests the lock acquisition functionality
func TestRedisDistributedLocker_AcquireLock(t *testing.T) {
	// Use unified configuration function to ensure password is included
	config := getTestRedisConfig()

	locker, err := NewRedisDistributedLocker(config)
	if err != nil {
		t.Fatalf("Skipping test: cannot connect to Redis: %v", err)
	}
	defer locker.Close()

	ctx := context.Background()
	key := "test_lock_key"
	expiration := 5 * time.Second

	// Test acquiring lock
	acquired, err := locker.AcquireLock(ctx, key, expiration)
	if err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}
	if !acquired {
		t.Error("Expected to acquire lock but failed")
	}

	// Test trying to acquire the same lock again
	acquired2, err := locker.AcquireLock(ctx, key, expiration)
	if err != nil {
		t.Fatalf("Failed to try acquiring lock again: %v", err)
	}
	if acquired2 {
		t.Error("Expected to fail acquiring same lock but succeeded")
	}

	// Clean up
	err = locker.ReleaseLock(ctx, key)
	if err != nil {
		t.Errorf("Failed to release lock: %v", err)
	}
}

// TestRedisDistributedLocker_ReleaseLock tests the lock release functionality
func TestRedisDistributedLocker_ReleaseLock(t *testing.T) {
	// Use unified configuration function
	config := getTestRedisConfig()

	locker, err := NewRedisDistributedLocker(config)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to Redis: %v", err)
	}
	defer locker.Close()

	ctx := context.Background()
	key := "test_release_lock_key"
	expiration := 5 * time.Second

	// First acquire the lock
	acquired, err := locker.AcquireLock(ctx, key, expiration)
	if err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}
	if !acquired {
		t.Fatal("Failed to acquire lock")
	}

	// Release the lock
	err = locker.ReleaseLock(ctx, key)
	if err != nil {
		t.Fatalf("Failed to release lock: %v", err)
	}

	// Now we should be able to acquire it again
	acquired2, err := locker.AcquireLock(ctx, key, expiration)
	if err != nil {
		t.Fatalf("Failed to acquire lock after release: %v", err)
	}
	if !acquired2 {
		t.Error("Expected to acquire lock after release but failed")
	}

	// Clean up
	err = locker.ReleaseLock(ctx, key)
	if err != nil {
		t.Errorf("Failed to release lock: %v", err)
	}
}

// TestRedisDistributedLocker_GenerateTrapLockKey tests the trap lock key generation
func TestRedisDistributedLocker_GenerateTrapLockKey(t *testing.T) {
	// Create instance directly, no Redis connection needed for this test
	locker := &RedisDistributedLocker{}

	testData := []byte("test trap data for hashing")
	expectedPrefix := "trap_lock:"

	key := locker.GenerateTrapLockKey(testData)

	if key == "" {
		t.Error("Generated key is empty")
	}

	if len(key) <= len(expectedPrefix) {
		t.Errorf("Generated key is too short. Key: %s", key)
	}

	if key[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("Generated key doesn't have expected prefix. Got: %s", key)
	}

	// Test that same data produces same key
	key2 := locker.GenerateTrapLockKey(testData)
	if key != key2 {
		t.Error("Same data should produce same key")
	}

	// Test that different data produces different key
	differentData := []byte("different test trap data")
	key3 := locker.GenerateTrapLockKey(differentData)
	if key == key3 {
		t.Error("Different data should produce different key")
	}

	// Test with empty data
	emptyKey := locker.GenerateTrapLockKey([]byte{})
	if emptyKey == "" {
		t.Error("Empty data should still generate a key")
	}
	if emptyKey[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("Empty data key doesn't have expected prefix. Got: %s", emptyKey)
	}
}

// TestNewRedisDistributedLocker tests the creation of Redis distributed locker
func TestNewRedisDistributedLocker(t *testing.T) {
	// Test successful creation
	config := getTestRedisConfig()

	locker, err := NewRedisDistributedLocker(config)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to Redis: %v", err)
	}

	if locker == nil {
		t.Error("Expected locker instance, got nil")
	}

	// Clean up
	err = locker.Close()
	if err != nil {
		t.Errorf("Error closing locker: %v", err)
	}

	// Test failure with invalid config - wrong port
	invalidConfig := getTestRedisConfig()
	invalidConfig.Port = 8090

	_, err = NewRedisDistributedLocker(invalidConfig)
	if err == nil {
		t.Error("Expected error with invalid Redis config, got nil")
	}

	// Test failure with wrong password
	wrongPasswordConfig := getTestRedisConfig()
	wrongPasswordConfig.Password = "wrong_password"

	_, err = NewRedisDistributedLocker(wrongPasswordConfig)
	if err == nil {
		t.Error("Expected error with wrong password, got nil")
	}
}

// TestRedisDistributedLocker_ConcurrentLock tests concurrent lock acquisition
func TestRedisDistributedLocker_ConcurrentLock(t *testing.T) {
	config := getTestRedisConfig()

	locker, err := NewRedisDistributedLocker(config)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to Redis: %v", err)
	}
	defer locker.Close()

	ctx := context.Background()
	lockKey := "test_concurrent_lock"
	expiration := 5 * time.Second

	// Use channel to collect concurrent test results
	results := make(chan bool, 3)

	// Start 3 goroutines to attempt acquiring the same lock simultaneously
	for i := 0; i < 3; i++ {
		go func(id int) {
			localLocker, err := NewRedisDistributedLocker(config)
			if err != nil {
				results <- false
				return
			}
			defer localLocker.Close()

			acquired, err := localLocker.AcquireLock(ctx, lockKey, expiration)
			if err != nil {
				results <- false
				return
			}
			results <- acquired

			// If acquired successfully, release after a short delay
			if acquired {
				time.Sleep(100 * time.Millisecond)
				localLocker.ReleaseLock(ctx, lockKey)
			}
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < 3; i++ {
		if <-results {
			successCount++
		}
	}

	// Only one goroutine should successfully acquire the lock
	if successCount != 1 {
		t.Errorf("Expected exactly 1 successful lock acquisition, got %d", successCount)
	}
}
