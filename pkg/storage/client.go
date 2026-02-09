package storage

import (
	"baize-monitor/pkg/config"
	"baize-monitor/pkg/logger"
	"baize-monitor/pkg/models"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Client struct {
	DB *gorm.DB
}

func NewClient(cfg *config.PostGresConfig) (*Client, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)

	// Use our custom GORM logging configuration
	gormConfig := &gorm.Config{
		Logger: logger.GetGormLogger(),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get generic database object sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Set connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxConns)
	sqlDB.SetMaxIdleConns(cfg.IdleConns)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	return &Client{DB: db}, nil
}

func (c *Client) Close() error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// HealthCheck Database health check
func (c *Client) HealthCheck() error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return err
	}

	err = sqlDB.Ping()
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// AutoMigrate Auto migrate table structure
func (c *Client) AutoMigrate(models ...interface{}) error {

	err := c.DB.AutoMigrate(models...)
	if err != nil {
		return fmt.Errorf("auto migration failed: %w", err)
	}

	return nil
}

// WithTransaction Transaction support
func (c *Client) WithTransaction(fn func(tx *gorm.DB) error) error {

	err := c.DB.Transaction(func(tx *gorm.DB) error {
		err := fn(tx)
		if err != nil {
			return err
		}
		return nil
	})

	return err
}

// GetStats Get database statistics
func (c *Client) GetStats() map[string]interface{} {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration,
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}

// InitDatabase Initialize database (including auto migration)
func (c *Client) InitDatabase(models ...interface{}) error {

	if err := c.AutoMigrate(models...); err != nil {
		return err
	}

	return nil
}

func GetTableModelsList() []interface{} {
	return []interface{}{
		&models.BMCTrapParser{},
		&models.Alert{},
		&models.User{},
		&models.DistributedLockRecord{},
	}
}
