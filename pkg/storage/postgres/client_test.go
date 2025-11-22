package storage

import (
	"baize-monitor/pkg/config"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// Test models
type TestUser struct {
	ID        uint      `gorm:"primarykey"`
	Name      string    `gorm:"size:100;not null"`
	Email     string    `gorm:"size:255;uniqueIndex;not null"`
	Age       int       `gorm:"not null;default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type TestProduct struct {
	ID          uint      `gorm:"primarykey"`
	Name        string    `gorm:"size:200;not null"`
	Price       float64   `gorm:"type:decimal(10,2);not null"`
	Description string    `gorm:"type:text"`
	IsActive    bool      `gorm:"not null;default:true"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}

// Mock test configuration
func getMockTestConfig() *config.PostGresConfig {
	return &config.PostGresConfig{
		Host:      "localhost",
		Port:      5432,
		User:      "postgres",
		Password:  "qwer1234",
		Database:  "baize_test",
		SSLMode:   "disable",
		MaxConns:  10,
		IdleConns: 5,
	}
}

// Test main entry point
func TestMain(m *testing.M) {
	// Check database connection before running
	cfg := getMockTestConfig()
	client, err := NewClient(cfg)
	if err != nil {
		panic("Unable to connect to test database: " + err.Error())
	}
	defer client.Close()

	// Run tests
	m.Run()
}

func TestNewClient(t *testing.T) {
	cfg := getMockTestConfig()
	client, err := NewClient(cfg)

	assert.NoError(t, err, "Should successfully create database client")
	assert.NotNil(t, client, "Client should not be nil")
	assert.NotNil(t, client.DB, "Database connection should not be nil")

	defer client.Close()

	// Test health check
	err = client.HealthCheck()
	assert.NoError(t, err, "Health check should pass")
}

func TestClient_HealthCheck(t *testing.T) {
	cfg := getMockTestConfig()
	client, err := NewClient(cfg)
	require.NoError(t, err)
	defer client.Close()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Normal connection health check",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.HealthCheck()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_AutoMigrate(t *testing.T) {
	cfg := getMockTestConfig()
	client, err := NewClient(cfg)
	require.NoError(t, err)
	defer client.Close()

	// Test auto migrate single model
	err = client.AutoMigrate(&TestUser{})
	assert.NoError(t, err, "Single model migration should succeed")

	// Test auto migrate multiple models
	err = client.AutoMigrate(&TestUser{}, &TestProduct{})
	assert.NoError(t, err, "Multiple models migration should succeed")

	// Verify tables were actually created
	var tableExists bool
	err = client.DB.Raw(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'test_users'
		)
	`).Scan(&tableExists).Error
	assert.NoError(t, err)
	assert.True(t, tableExists, "test_users table should exist")

	// Clean up test tables
	client.DB.Migrator().DropTable(&TestUser{}, &TestProduct{})
}

func TestClient_WithTransaction(t *testing.T) {
	cfg := getMockTestConfig()
	client, err := NewClient(cfg)
	require.NoError(t, err)
	defer client.Close()

	// Create table first
	err = client.AutoMigrate(&TestUser{})
	require.NoError(t, err)
	defer client.DB.Migrator().DropTable(&TestUser{})

	t.Run("Transaction commit", func(t *testing.T) {
		err := client.WithTransaction(func(tx *gorm.DB) error {
			// Create user in transaction
			user := TestUser{
				Name:  "Transaction User",
				Email: "transaction@test.com",
				Age:   25,
			}
			if err := tx.Create(&user).Error; err != nil {
				return err
			}
			return nil
		})

		assert.NoError(t, err, "Transaction should commit successfully")

		// Verify data was saved
		var count int64
		client.DB.Model(&TestUser{}).Where("email = ?", "transaction@test.com").Count(&count)
		assert.Equal(t, int64(1), count, "Should find the user created in transaction")
	})

	t.Run("Transaction rollback", func(t *testing.T) {
		// Clear data first
		client.DB.Exec("DELETE FROM test_users")

		err := client.WithTransaction(func(tx *gorm.DB) error {
			// Create first user - should succeed
			user1 := TestUser{
				Name:  "User 1",
				Email: "user1@test.com",
				Age:   30,
			}
			if err := tx.Create(&user1).Error; err != nil {
				return err
			}

			// Create second user - intentionally cause error (duplicate email)
			user2 := TestUser{
				Name:  "User 2",
				Email: "user1@test.com", // Duplicate email, should cause error
				Age:   35,
			}
			if err := tx.Create(&user2).Error; err != nil {
				return err // Return error to trigger rollback
			}

			return nil
		})

		assert.Error(t, err, "Transaction should fail due to duplicate email")

		// Verify data was rolled back (no users saved)
		var count int64
		client.DB.Model(&TestUser{}).Count(&count)
		assert.Equal(t, int64(0), count, "Should have no data after transaction rollback")
	})
}

func TestClient_GetStats(t *testing.T) {
	cfg := getMockTestConfig()
	client, err := NewClient(cfg)
	require.NoError(t, err)
	defer client.Close()

	stats := client.GetStats()

	// Verify returned stats contain expected fields
	assert.Contains(t, stats, "max_open_connections")
	assert.Contains(t, stats, "open_connections")
	assert.Contains(t, stats, "in_use")
	assert.Contains(t, stats, "idle")
	assert.Contains(t, stats, "wait_count")

	// Verify numeric types
	maxOpenConns, ok := stats["max_open_connections"].(int)
	assert.True(t, ok, "max_open_connections should be integer")
	assert.Equal(t, cfg.MaxConns, maxOpenConns, "Maximum connections should match config")

	openConns, ok := stats["open_connections"].(int)
	assert.True(t, ok, "open_connections should be integer")
	assert.GreaterOrEqual(t, openConns, 0, "Open connections should be greater than or equal to 0")
}

func TestClient_InitDatabase(t *testing.T) {
	cfg := getMockTestConfig()
	client, err := NewClient(cfg)
	require.NoError(t, err)
	defer client.Close()

	// Test database initialization
	err = client.InitDatabase(&TestUser{}, &TestProduct{})
	assert.NoError(t, err, "Database initialization should succeed")

	// Verify tables were created
	var userTableExists, productTableExists bool

	// Check test_users table
	err = client.DB.Raw(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'test_users'
		)
	`).Scan(&userTableExists).Error
	assert.NoError(t, err)
	assert.True(t, userTableExists, "test_users table should exist")

	// Check test_products table
	err = client.DB.Raw(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'test_products'
		)
	`).Scan(&productTableExists).Error
	assert.NoError(t, err)
	assert.True(t, productTableExists, "test_products table should exist")

	// Cleanup
	client.DB.Migrator().DropTable(&TestUser{}, &TestProduct{})
}

func TestClient_Close(t *testing.T) {
	cfg := getMockTestConfig()
	client, err := NewClient(cfg)
	require.NoError(t, err)

	// Test closing connection
	err = client.Close()
	assert.NoError(t, err, "Closing connection should succeed")

	// Test health check should fail after closing
	err = client.HealthCheck()
	assert.Error(t, err, "Health check should fail after closing connection")
}

func TestClient_ConcurrentOperations(t *testing.T) {
	cfg := getMockTestConfig()
	client, err := NewClient(cfg)
	require.NoError(t, err)
	defer client.Close()

	// Create test table
	err = client.AutoMigrate(&TestUser{})
	require.NoError(t, err)
	defer client.DB.Migrator().DropTable(&TestUser{})

	// Concurrently create users
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			user := TestUser{
				Name:  "Concurrent User",
				Email: fmt.Sprintf("concurrent%d@test.com", id),
				Age:   id,
			}
			err := client.DB.Create(&user).Error
			assert.NoError(t, err, "Concurrent user creation should succeed")
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify connection pool stats
	stats := client.GetStats()
	t.Logf("Connection pool stats after concurrent operations: %+v", stats)
}

func TestClient_ConnectionPool(t *testing.T) {
	cfg := getMockTestConfig()
	cfg.MaxConns = 5
	cfg.IdleConns = 2

	client, err := NewClient(cfg)
	require.NoError(t, err)
	defer client.Close()

	// Create test table
	err = client.AutoMigrate(&TestUser{})
	require.NoError(t, err)
	defer client.DB.Migrator().DropTable(&TestUser{})

	// Perform some operations to observe connection pool behavior
	for i := 0; i < 10; i++ {
		var user TestUser
		client.DB.First(&user) // This will acquire and release connections in the pool
	}

	stats := client.GetStats()
	t.Logf("Connection pool stats: %+v", stats)

	// Verify connection pool configuration
	maxOpenConns, _ := stats["max_open_connections"].(int)
	assert.Equal(t, cfg.MaxConns, maxOpenConns, "Maximum connections should match config")
}

// Error scenario tests - using invalid configurations
func TestClient_InvalidConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.PostGresConfig
		expectError bool
	}{
		{
			name: "Invalid hostname",
			config: &config.PostGresConfig{
				Host:     "invalid-host-that-does-not-exist",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				Database: "test",
				SSLMode:  "disable",
			},
			expectError: true,
		},
		{
			name: "Invalid port",
			config: &config.PostGresConfig{
				Host:     "localhost",
				Port:     9999, // Non-existent port
				User:     "postgres",
				Password: "password",
				Database: "test",
				SSLMode:  "disable",
			},
			expectError: true,
		},
		{
			name: "Invalid database name",
			config: &config.PostGresConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				Database: "non_existent_database_12345",
				SSLMode:  "disable",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if tt.expectError {
				assert.Error(t, err, "Should return error")
				assert.Nil(t, client, "Client should be nil")
			} else {
				assert.NoError(t, err, "Should not return error")
				if client != nil {
					client.Close()
				}
			}
		})
	}
}

// Basic CRUD operations test
func TestClient_BasicCRUD(t *testing.T) {
	cfg := getMockTestConfig()
	client, err := NewClient(cfg)
	require.NoError(t, err)
	defer client.Close()

	// Create table
	err = client.AutoMigrate(&TestUser{})
	require.NoError(t, err)
	defer client.DB.Migrator().DropTable(&TestUser{})

	t.Run("Create record", func(t *testing.T) {
		user := TestUser{
			Name:  "Test User",
			Email: "test@example.com",
			Age:   30,
		}
		err := client.DB.Create(&user).Error
		assert.NoError(t, err)
		assert.NotZero(t, user.ID, "Should have ID after creation")

		// Verify record exists
		var foundUser TestUser
		err = client.DB.First(&foundUser, user.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, user.Name, foundUser.Name)
		assert.Equal(t, user.Email, foundUser.Email)
	})

	t.Run("Update record", func(t *testing.T) {
		var user TestUser
		err := client.DB.First(&user).Error
		require.NoError(t, err)

		// Update record
		newName := "Updated User"
		err = client.DB.Model(&user).Update("name", newName).Error
		assert.NoError(t, err)

		// Verify update
		var updatedUser TestUser
		err = client.DB.First(&updatedUser, user.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, newName, updatedUser.Name)
	})

	t.Run("Delete record", func(t *testing.T) {
		var user TestUser
		err := client.DB.First(&user).Error
		require.NoError(t, err)

		// Delete record
		err = client.DB.Delete(&user).Error
		assert.NoError(t, err)

		// Verify deletion
		var deletedUser TestUser
		err = client.DB.First(&deletedUser, user.ID).Error
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})
}
