package logger

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"baize-monitor/pkg/config"

	"gorm.io/gorm/logger"
)

// Add test-specific configuration setting function at the beginning of the test file
func setupTestConfig() {
	globalConfig = &config.LogConfig{
		Output: "stdout",
		Level:  "info",
		Format: "text",
		GORM: config.GORMLogConfig{
			Enabled:                   true,
			Level:                     "info",
			SlowThreshold:             100 * time.Millisecond,
			IgnoreRecordNotFoundError: true,
		},
	}
}

func TestGetGormLogger(t *testing.T) {
	// Reset singleton state
	gormLoggerOnce = sync.Once{}
	gormLoggerInstance = nil

	setupTestConfig()

	// Test singleton pattern for getting GORM logger
	logger1 := GetGormLogger()
	logger2 := GetGormLogger()

	if logger1 != logger2 {
		t.Error("Expected same instance from GetGormLogger()")
	}
}

func TestGormLogger_LogMode(t *testing.T) {
	setupTestConfig()

	gormLogger := GetGormLogger()
	newLogger := gormLogger.LogMode(logger.Info)

	if newLogger == nil {
		t.Error("Expected new logger instance")
	}
}

func TestGetGormSlogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"silent", slog.LevelError + 1},
		{"error", slog.LevelError},
		{"warn", slog.LevelWarn},
		{"info", slog.LevelInfo},
		{"unknown", slog.LevelWarn}, // default case
	}

	for _, test := range tests {
		result := getGormSlogLevel(test.input)
		if result != test.expected {
			t.Errorf("For input '%s', expected '%v' but got '%v'", test.input, test.expected, result)
		}
	}
}

func TestGormLogger_Info(t *testing.T) {
	setupTestConfig()

	l := newGormLogger().(*gormLogger)

	// Test that Info method does not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Info method panicked: %v", r)
			}
		}()
		l.Info(context.Background(), "test message")
	}()
}

func TestGormLogger_Warn(t *testing.T) {
	setupTestConfig()

	l := newGormLogger().(*gormLogger)

	// Test that Warn method does not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Warn method panicked: %v", r)
			}
		}()
		l.Warn(context.Background(), "test warning")
	}()
}

func TestGormLogger_Error(t *testing.T) {
	setupTestConfig()

	l := newGormLogger().(*gormLogger)

	// Test that Error method does not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Error method panicked: %v", r)
			}
		}()
		l.Error(context.Background(), "test error")
	}()
}

func TestGormLogger_Trace(t *testing.T) {
	setupTestConfig()

	l := newGormLogger().(*gormLogger)

	// Test that Trace method does not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Trace method panicked: %v", r)
			}
		}()

		begin := time.Now()
		fc := func() (string, int64) {
			return "SELECT * FROM users", 10
		}

		l.Trace(context.Background(), begin, fc, nil)
	}()
}

func TestGormLogger_getTraceLogLevel(t *testing.T) {
	setupTestConfig()

	l := newGormLogger().(*gormLogger)

	// Test log levels under different conditions
	tests := []struct {
		name                 string
		elapsed              time.Duration
		err                  error
		level                string
		expected             logger.LogLevel
		ignoreRecordNotFound bool
		slowThreshold        time.Duration
	}{
		{
			name:          "Silent level - normal query",
			elapsed:       time.Millisecond,
			err:           nil,
			level:         "silent",
			expected:      logger.Silent,
			slowThreshold: 100 * time.Millisecond,
		},
		{
			name:          "Info level - normal query",
			elapsed:       time.Millisecond,
			err:           nil,
			level:         "info",
			expected:      logger.Info,
			slowThreshold: 100 * time.Millisecond,
		},
		{
			name:          "Error level - normal query",
			elapsed:       time.Millisecond,
			err:           nil,
			level:         "error",
			expected:      logger.Silent,
			slowThreshold: 100 * time.Millisecond,
		},
		{
			name:          "Warn level - normal query",
			elapsed:       time.Millisecond,
			err:           nil,
			level:         "warn",
			expected:      logger.Silent,
			slowThreshold: 100 * time.Millisecond,
		},
		{
			name:          "Info level - slow query",
			elapsed:       200 * time.Millisecond, // Exceed threshold
			err:           nil,
			level:         "info",
			expected:      logger.Warn, // Fix: Slow queries return Warn
			slowThreshold: 100 * time.Millisecond,
		},
		{
			name:          "Warn level - slow query",
			elapsed:       200 * time.Millisecond, // Exceed threshold
			err:           nil,
			level:         "warn",
			expected:      logger.Warn, // Slow queries are recorded as warn under warn level
			slowThreshold: 100 * time.Millisecond,
		},
		{
			name:                 "Error with record not found and ignored",
			elapsed:              time.Millisecond,
			err:                  logger.ErrRecordNotFound,
			level:                "error",
			expected:             logger.Silent, // Ignored record not found errors are not logged at error level
			ignoreRecordNotFound: true,
			slowThreshold:        100 * time.Millisecond,
		},
		{
			name:                 "Error with record not found not ignored",
			elapsed:              time.Millisecond,
			err:                  logger.ErrRecordNotFound,
			level:                "error",
			expected:             logger.Error, // Non-ignored record not found errors are recorded as error
			ignoreRecordNotFound: false,
			slowThreshold:        100 * time.Millisecond,
		},
		{
			name:          "Error with other error",
			elapsed:       time.Millisecond,
			err:           errors.New("other error"),
			level:         "error",
			expected:      logger.Error, // Other errors are recorded as error at error level
			slowThreshold: 100 * time.Millisecond,
		},
		{
			name:          "Warn level with error",
			elapsed:       time.Millisecond,
			err:           errors.New("test error"),
			level:         "warn",
			expected:      logger.Error, // Fix: Return Error when there is an error
			slowThreshold: 100 * time.Millisecond,
		},
		{
			name:          "Info level with error",
			elapsed:       time.Millisecond,
			err:           errors.New("test error"),
			level:         "info",
			expected:      logger.Error, // Fix: Return Error when there is an error
			slowThreshold: 100 * time.Millisecond,
		},
		{
			name:          "Error level with slow query",
			elapsed:       200 * time.Millisecond, // Exceed threshold
			err:           nil,
			level:         "error",
			expected:      logger.Warn, // Fix: Slow queries return Warn
			slowThreshold: 100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create new configuration for each test case
			l.config = &config.GORMLogConfig{
				Enabled:                   true,
				Level:                     tt.level,
				SlowThreshold:             tt.slowThreshold,
				IgnoreRecordNotFoundError: tt.ignoreRecordNotFound,
			}

			result := l.getTraceLogLevel(tt.elapsed, tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Test disabled scenario
func TestGormLogger_Disabled(t *testing.T) {
	setupTestConfig()

	l := newGormLogger().(*gormLogger)
	l.config.Enabled = false

	// All methods should not panic when disabled
	ctx := context.Background()

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Method panicked when disabled: %v", r)
			}
		}()

		l.Info(ctx, "test")
		l.Warn(ctx, "test")
		l.Error(ctx, "test")

		begin := time.Now()
		fc := func() (string, int64) { return "SELECT 1", 1 }
		l.Trace(ctx, begin, fc, nil)
	}()
}

// Test slow query threshold
func TestGormLogger_SlowQuery(t *testing.T) {
	setupTestConfig()

	l := newGormLogger().(*gormLogger)
	l.config.SlowThreshold = 50 * time.Millisecond

	tests := []struct {
		name     string
		elapsed  time.Duration
		level    string
		expected logger.LogLevel
	}{
		{
			name:     "Info level - Below threshold",
			elapsed:  25 * time.Millisecond,
			level:    "info",
			expected: logger.Info,
		},
		{
			name:     "Info level - Above threshold",
			elapsed:  75 * time.Millisecond,
			level:    "info",
			expected: logger.Warn, // Fix: Slow queries return Warn
		},
		{
			name:     "Warn level - Below threshold",
			elapsed:  25 * time.Millisecond,
			level:    "warn",
			expected: logger.Silent, // Normal queries are not logged under warn level
		},
		{
			name:     "Warn level - Above threshold",
			elapsed:  75 * time.Millisecond,
			level:    "warn",
			expected: logger.Warn, // Slow queries are recorded as warn under warn level
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l.config.Level = tt.level
			result := l.getTraceLogLevel(tt.elapsed, nil)
			if result != tt.expected {
				t.Errorf("For elapsed %v and level %s, expected %v, got %v", tt.elapsed, tt.level, tt.expected, result)
			}
		})
	}
}

// Test default configuration
func TestGormLogger_DefaultConfig(t *testing.T) {
	// Reset global configuration
	globalConfig = nil
	gormLoggerOnce = sync.Once{}
	gormLoggerInstance = nil

	// Should be able to create logger without panicking
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Creating logger with default config panicked: %v", r)
			}
		}()

		logger := GetGormLogger()
		if logger == nil {
			t.Error("Expected logger to be created with default config")
		}
	}()
}

// Test handling of slow queries in Trace method
func TestGormLogger_TraceSlowQuery(t *testing.T) {
	setupTestConfig()

	l := newGormLogger().(*gormLogger)
	l.config.SlowThreshold = 50 * time.Millisecond

	// Test that slow queries will be logged as warn
	ctx := context.Background()
	begin := time.Now().Add(-100 * time.Millisecond) // Slow query
	fc := func() (string, int64) { return "SELECT * FROM large_table", 1000 }

	// This test mainly verifies no panic occurs, we cannot directly verify actual log output in tests
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Trace with slow query panicked: %v", r)
			}
		}()
		l.Trace(ctx, begin, fc, nil)
	}()
}

// Test error handling
func TestGormLogger_ErrorHandling(t *testing.T) {
	setupTestConfig()

	l := newGormLogger().(*gormLogger)

	ctx := context.Background()
	begin := time.Now()
	fc := func() (string, int64) { return "INSERT INTO users VALUES (1)", 1 }
	testError := errors.New("duplicate key value")

	// Test error scenarios do not panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Trace with error panicked: %v", r)
			}
		}()
		l.Trace(ctx, begin, fc, testError)
	}()
}
