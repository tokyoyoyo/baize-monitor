package storage

import (
	"baize-monitor/pkg/config"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisDistributedLocker DistributedLocker base on redis
type RedisDistributedLocker struct {
	client *redis.Client
}

// NewRedisDistributedLocker
func NewRedisDistributedLocker(config *config.RedisConfig) (DistributedLockerInterface, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	// connect test
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisDistributedLocker{
		client: client,
	}, nil
}

// AcquireLock
func (r *RedisDistributedLocker) AcquireLock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	result := r.client.SetNX(ctx, key, "locked", expiration)
	if result.Err() != nil {
		return false, result.Err()
	}
	return result.Val(), nil
}

// ReleaseLock
func (r *RedisDistributedLocker) ReleaseLock(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// GenerateTrapLockKey 	generate lock key for trap data
func (r *RedisDistributedLocker) GenerateTrapLockKey(trapData []byte) string {
	hash := sha256.Sum256(trapData)
	return "trap_lock:" + hex.EncodeToString(hash[:])
}

// Close close redis client
func (r *RedisDistributedLocker) Close() error {
	return r.client.Close()
}
