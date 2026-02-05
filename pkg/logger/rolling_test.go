package logger

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewRollingFile(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "test.log")

	rf := newRollingFile(filename, 10, 5, 7)

	if rf.filename != filename {
		t.Errorf("Expected filename %s, got %s", filename, rf.filename)
	}

	if rf.maxSize != 10*1024*1024 {
		t.Errorf("Expected maxSize %d, got %d", 10*1024*1024, rf.maxSize)
	}

	if rf.maxBackups != 5 {
		t.Errorf("Expected maxBackups %d, got %d", 5, rf.maxBackups)
	}

	expectedMaxAge := 7 * 24 * time.Hour
	if rf.maxAge != expectedMaxAge {
		t.Errorf("Expected maxAge %v, got %v", expectedMaxAge, rf.maxAge)
	}
}

func TestRollingFileWrite(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "test.log")

	rf := newRollingFile(filename, 1, 3, 1) // 1MB max size

	data := []byte("Hello, World!")
	n, err := rf.Write(data)

	if err != nil {
		t.Errorf("Write failed: %v", err)
	}

	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}

	// Verify file was created and contains our data
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("Failed to read file: %v", err)
	}

	if string(content) != string(data) {
		t.Errorf("File content mismatch")
	}
}

func TestRollingFileRotation(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "test.log")

	// Small max size to trigger rotation quickly
	rf := newRollingFile(filename, 10, 3, 1)

	// Write enough data to trigger rotation
	largeData := make([]byte, 20*1024*1024) // 20 bytes
	_, err := rf.Write(largeData)
	if err != nil {
		t.Errorf("Write failed: %v", err)
	}

	// Check if backup files were created
	matches, err := filepath.Glob(filename + ".*")
	if err != nil {
		t.Errorf("Failed to glob backup files: %v", err)
	}

	if len(matches) == 2 {
		t.Error("Expected backup files to be created")
	}
}

func TestRollingFileCleanup(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "test.log")

	// Create a rolling file with small backup limit
	rf := newRollingFile(filename, 5, 2, 1) // Keep only 2 backups

	// Create some mock backup files
	backups := []string{
		filename + ".2023-01-01T00-00-00.000",
		filename + ".2023-01-02T00-00-00.000",
		filename + ".2023-01-03T00-00-00.000",
	}

	for _, backup := range backups {
		f, err := os.Create(backup)
		if err != nil {
			t.Fatalf("Failed to create test backup: %v", err)
		}
		f.Close()
	}

	// Run cleanup
	rf.cleanup()

	// Check if only the newest 2 backups remain
	matches, err := filepath.Glob(filename + ".*")
	if err != nil {
		t.Errorf("Failed to glob backup files: %v", err)
	}

	if len(matches) != 2 {
		t.Errorf("Expected max 2 backups, got %d", len(matches))
	}
}

func TestRollingFileSync(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "test.log")

	rf := newRollingFile(filename, 10, 5, 7)

	// Sync without file opened should not error
	err := rf.Sync()
	if err != nil {
		t.Errorf("Sync without file should not error, got: %v", err)
	}

	// Write some data to open the file
	_, err = rf.Write([]byte("test"))
	if err != nil {
		t.Errorf("Write failed: %v", err)
	}

	// Now sync should work
	err = rf.Sync()
	if err != nil {
		t.Errorf("Sync failed: %v", err)
	}
}

func TestRollingFileConcurrentWrites(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "test.log")

	rf := newRollingFile(filename, 501, 1, 1)

	// Write concurrently from multiple goroutines
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				data := make([]byte, 0.5*1024*1024)
				_, err := rf.Write(data)
				if err != nil {
					t.Errorf("Concurrent write failed: %v", err)
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all data was written
	info, err := os.Stat(filename)
	if err != nil {
		t.Errorf("Failed to stat file: %v", err)
	}

	if info.Size() != 500*1024*1024 {
		t.Error("Incorrect file size")
	}
}

func TestRollingFileDirectoryCreation(t *testing.T) {
	tempDir := t.TempDir()
	// Create a nested directory path
	filename := filepath.Join(tempDir, "nested", "dir", "test.log")

	rf := newRollingFile(filename, 10, 5, 7)

	// This should create the directory structure
	_, err := rf.Write([]byte("test data"))
	if err != nil {
		t.Errorf("Write should create directories, got error: %v", err)
	}

	// Verify directory was created
	dir := filepath.Dir(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("Expected directory to be created")
	}
}

func TestRollingFileOpenError(t *testing.T) {
	// Try to create file in non-writable location (root directory on Unix)
	if os.Getuid() == 0 {
		t.Skip("Test skipped when running as root")
	}

	filename := "/proc/invalid/test.log"
	rf := newRollingFile(filename, 10, 5, 7)

	_, err := rf.Write([]byte("test"))
	if err == nil {
		t.Error("Expected error when writing to invalid location")
	}
}

func TestRollingFileClose(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "test.log")

	rf := newRollingFile(filename, 10, 5, 7)

	// Write some data to open the file
	_, err := rf.Write([]byte("test"))
	if err != nil {
		t.Errorf("Write failed: %v", err)
	}

	// The file should be closed during rotation
	rf.rotate()

	// Verify we can write again (file should be reopened)
	_, err = rf.Write([]byte("more test"))
	if err != nil {
		t.Errorf("Write after rotation failed: %v", err)
	}
}
