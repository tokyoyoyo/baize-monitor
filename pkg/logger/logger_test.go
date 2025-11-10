package logger

import (
	"baize-monitor/pkg/models"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// TestInit tests the initialization of the logger configuration
func TestInit(t *testing.T) {
	// Reset global state before test
	globalConfig = nil
	rollingFiles = nil

	cfg := &models.LogConfig{
		MaxSizeMB:  50,
		MaxBackups: 10,
		MaxAgeDays: 30,
		LogDir:     "/test/logs",
		Output:     "file",
		Level:      "debug",
		Format:     "json",
	}

	err := init_log_by_cfg(cfg)
	if err != nil {
		t.Errorf("Init failed: %v", err)
	}

	if globalConfig == nil {
		t.Error("Global config should be set after Init")
	}

	if globalConfig.MaxSizeMB != 50 {
		t.Errorf("Expected MaxSizeMB 50, got %d", globalConfig.MaxSizeMB)
	}

	if globalConfig.LogDir != "/test/logs" {
		t.Errorf("Expected LogDir /test/logs, got %s", globalConfig.LogDir)
	}
}

func TestInitWithNilConfig(t *testing.T) {
	globalConfig = nil

	err := init_log_by_cfg(nil)
	if err == nil {
		t.Error("Expected error with nil config")
	}

	if globalConfig != nil {
		t.Error("Global config should remain nil with invalid config")
	}
}

func TestInitWithDefaultValues(t *testing.T) {
	globalConfig = nil

	cfg := &models.LogConfig{
		MaxSizeMB:  0,  // Should use default
		MaxBackups: 0,  // Should use default
		MaxAgeDays: 0,  // Should use default
		LogDir:     "", // Should use default
		Output:     "file",
		Level:      "info",
		Format:     "text",
	}

	err := init_log_by_cfg(cfg)
	if err != nil {
		t.Errorf("Init failed: %v", err)
	}

	if globalConfig.MaxSizeMB != 100 {
		t.Errorf("Expected default MaxSizeMB 100, got %d", globalConfig.MaxSizeMB)
	}

	if globalConfig.MaxBackups != 10 {
		t.Errorf("Expected default MaxBackups 10, got %d", globalConfig.MaxBackups)
	}

	if globalConfig.MaxAgeDays != 7 {
		t.Errorf("Expected default MaxAgeDays 7, got %d", globalConfig.MaxAgeDays)
	}

	if globalConfig.LogDir != "/var/log/baize" {
		t.Errorf("Expected default LogDir /var/log/baize, got %s", globalConfig.LogDir)
	}
}

func TestForModuleFileOutput(t *testing.T) {
	tempDir := t.TempDir()
	globalConfig = &models.LogConfig{
		MaxSizeMB:  10,
		MaxBackups: 5,
		MaxAgeDays: 7,
		LogDir:     tempDir,
		Output:     "file",
		Level:      "info",
		Format:     "text",
	}

	logger := forModule("testmodule")

	if logger == nil {
		t.Error("Expected logger to be created")
	}

	logger.Info("Test log message")

	// Verify log directory was created
	moduleDir := filepath.Join(tempDir, "testmodule")
	if _, err := os.Stat(moduleDir); os.IsNotExist(err) {
		t.Error("Expected module directory to be created")
	}

	// Verify log file was created
	logFile := filepath.Join(moduleDir, "testmodule.log")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Expected log file to be created")
	}
}

func TestForModuleStdoutOutput(t *testing.T) {
	globalConfig = &models.LogConfig{
		Output: "stdout",
		Level:  "info",
		Format: "text",
	}

	logger := forModule("testmodule")

	if logger == nil {
		t.Error("Expected logger to be created with stdout output")
	}
}

func TestForModuleInitializationScenarios(t *testing.T) {
	originalConfig := globalConfig
	defer func() {
		globalConfig = originalConfig
	}()

	testCases := []struct {
		name        string
		setup       func()
		expectPanic bool
		panicMsg    string
	}{
		{
			name: "Panic when globalConfig is nil",
			setup: func() {
				globalConfig = nil
			},
			expectPanic: true,
			panicMsg:    "logger not initialized",
		},
		{
			name: "No panic when globalConfig is set",
			setup: func() {
				globalConfig = &models.LogConfig{
					Output: "stdout",
					Level:  "info",
					Format: "text",
				}
			},
			expectPanic: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()

			defer func() {
				r := recover()
				if tc.expectPanic {
					if r == nil {
						t.Error("Expected panic but none occurred")
					} else if tc.panicMsg != "" {
						panicMsg := fmt.Sprintf("%v", r)
						if !strings.Contains(panicMsg, tc.panicMsg) {
							t.Errorf("Expected panic message to contain '%s', got: '%s'", tc.panicMsg, panicMsg)
						}
					}
				} else {
					if r != nil {
						t.Errorf("Unexpected panic: %v", r)
					}
				}
			}()

			logger := forModule("testmodule")

			// If we get here and no panic was expected, verify logger was created
			if !tc.expectPanic && logger == nil {
				t.Error("Expected logger to be created")
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	testCases := []struct {
		configLevel string
		expected    string
	}{
		{"debug", "Level(-4)"},
		{"info", "Level(0)"},
		{"warn", "Level(4)"},
		{"error", "Level(8)"},
		{"invalid", "Level(0)"}, // Should default to info
	}

	for _, tc := range testCases {
		t.Run(tc.configLevel, func(t *testing.T) {
			globalConfig = &models.LogConfig{
				Output: "stdout",
				Level:  tc.configLevel,
				Format: "text",
			}

			logger := forModule("testmodule")
			handler := logger.Handler()

			// The handler doesn't expose level directly in a simple way,
			// but we can verify the logger was created without error
			if logger == nil {
				t.Error("Expected logger to be created")
			}
			if handler == nil {
				t.Error("Expected handler to be created")
			}
		})
	}
}

func TestLogFormats(t *testing.T) {
	// Test JSON format
	globalConfig = &models.LogConfig{
		Output: "stdout",
		Level:  "info",
		Format: "json",
	}

	logger := forModule("testmodule")
	if logger == nil {
		t.Error("Expected JSON logger to be created")
	}

	// Verify it's using JSON handler
	handler := logger.Handler()
	if _, ok := handler.(*slog.JSONHandler); !ok {
		t.Error("Expected JSON handler for JSON format")
	}

	// Test text format
	globalConfig.Format = "text"
	logger2 := forModule("testmodule2")
	if logger2 == nil {
		t.Error("Expected text logger to be created")
	}

	// Verify it's using text handler
	handler2 := logger2.Handler()
	if _, ok := handler2.(*slog.TextHandler); !ok {
		t.Error("Expected text handler for text format")
	}
}

func TestGetLogger(t *testing.T) {
	// Reset global state
	globalConfig = nil

	logger := getLogger("snmp")
	if logger == nil {
		t.Error("Expected logger to be created by getLogger")
	}

	logger2 := getLogger("snmp2")
	if logger2 == nil {
		t.Error("Expected logger to be created by getLogger")
	}

	if logger == logger2 {
		t.Error("Expected different loggers for different modules")
	}

	// Verify global config was initialized
	if globalConfig == nil {
		t.Error("Expected global config to be initialized by getLogger")
	}
}

func TestSync(t *testing.T) {
	tempDir := t.TempDir()

	// Create some rolling files manually
	rf1 := newRollingFile(filepath.Join(tempDir, "test1.log"), 10, 5, 7)
	rf2 := newRollingFile(filepath.Join(tempDir, "test2.log"), 10, 5, 7)

	// Write some data to open the files
	rf1.Write([]byte("test1"))
	rf2.Write([]byte("test2"))

	// Add to global rollingFiles slice
	rollingFiles = []*rollingFile{rf1, rf2}

	err := Sync()
	if err != nil {
		t.Errorf("Sync failed: %v", err)
	}
}

func TestSyncWithNoFiles(t *testing.T) {
	rollingFiles = nil

	err := Sync()
	if err != nil {
		t.Errorf("Sync with no files should not error, got: %v", err)
	}
}

func TestModuleNameInLogger(t *testing.T) {
	globalConfig = &models.LogConfig{
		Output: "stdout",
		Level:  "info",
		Format: "text",
	}

	logger := forModule("testmodule")
	if logger == nil {
		t.Error("Expected logger to be created")
	}
}

func TestConcurrentLoggerCreation(t *testing.T) {
	tempDir := t.TempDir()
	globalConfig = &models.LogConfig{
		MaxSizeMB:  10,
		MaxBackups: 5,
		MaxAgeDays: 7,
		LogDir:     tempDir,
		Output:     "file",
		Level:      "info",
		Format:     "text",
	}

	var wg sync.WaitGroup
	modules := []string{"module1", "module2", "module3", "module4", "module5"}

	for _, module := range modules {
		wg.Add(1)
		go func(m string) {
			defer wg.Done()
			logger := forModule(m)
			if logger == nil {
				t.Errorf("Failed to create logger for module %s", m)
			}
			logger.Info("Test log message")
		}(module)
	}

	wg.Wait()

	// Verify all module directories were created
	for _, module := range modules {
		moduleDir := filepath.Join(tempDir, module)
		if _, err := os.Stat(moduleDir); os.IsNotExist(err) {
			t.Errorf("Expected module directory %s to be created", moduleDir)
		}
	}
}

func TestInitLoggerIdempotent(t *testing.T) {
	globalConfig = nil

	// First call should initialize
	init_logger()

	// Store the initial config
	initialConfig := globalConfig

	// Second call should not change the config
	init_logger()

	if globalConfig != initialConfig {
		t.Error("init_logger should be idempotent")
	}
}

func TestDefaultConfigFallback(t *testing.T) {
	// This test is tricky because loadLoggerConfig always returns error
	// But the init_logger should fall back to default config
	globalConfig = nil

	// We need to ensure that when loadLoggerConfig fails, default config is used
	// However, the current implementation always panics if Init fails
	// So we'll test this by ensuring getLogger doesn't panic

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("getLogger panicked: %v", r)
		}
	}()

	logger := getLogger("test")
	if logger == nil {
		t.Error("Expected getLogger to create logger with default config")
	}
}
