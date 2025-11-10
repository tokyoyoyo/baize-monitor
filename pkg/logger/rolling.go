package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type rollingFile struct {
	filename   string
	maxSize    int64
	maxBackups int
	maxAge     time.Duration
	mu         sync.Mutex
	file       *os.File
	size       int64
}

func newRollingFile(filename string, maxSizeMB, maxBackups, maxAgeDays int) *rollingFile {
	return &rollingFile{
		filename:   filename,
		maxSize:    int64(maxSizeMB) * 1024 * 1024,
		maxBackups: maxBackups,
		maxAge:     time.Duration(maxAgeDays) * 24 * time.Hour,
	}
}

func (r *rollingFile) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.file == nil {
		if err := r.open(); err != nil {
			return 0, err
		}
	}

	if r.size+int64(len(p)) >= r.maxSize {
		r.rotate()
	}

	n, err = r.file.Write(p)
	r.size += int64(n)
	return n, err
}

func (r *rollingFile) open() error {
	dir := filepath.Dir(r.filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log dir %s: %w", dir, err)
	}
	file, err := os.OpenFile(r.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	r.file = file
	stat, _ := file.Stat()
	r.size = stat.Size()
	return nil
}

func (r *rollingFile) rotate() {
	r.file.Close()
	now := time.Now().UTC().Format("2006-01-02T15-04-05.000")
	backup := r.filename + "." + now
	os.Rename(r.filename, backup)
	r.cleanup()
	r.open()
}

func (r *rollingFile) cleanup() {
	dir := filepath.Dir(r.filename)
	base := filepath.Base(r.filename)
	files, _ := os.ReadDir(dir)
	var backups []string
	for _, f := range files {
		if strings.HasPrefix(f.Name(), base+".") {
			backups = append(backups, filepath.Join(dir, f.Name()))
		}
	}
	sort.Strings(backups)

	// Delete files exceeding maxBackups
	if len(backups) > r.maxBackups {
		for _, f := range backups[:len(backups)-r.maxBackups] {
			os.Remove(f)
		}
	}

	// Delete expired files
	cutoff := time.Now().Add(-r.maxAge)
	for _, f := range backups {
		if info, err := os.Stat(f); err == nil {
			if info.ModTime().Before(cutoff) {
				os.Remove(f)
			}
		}
	}
}

func (r *rollingFile) Sync() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.file != nil {
		return r.file.Sync()
	}
	return nil
}
