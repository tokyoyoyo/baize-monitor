package models

type LogConfig struct {
	Level      string `yaml:"level"`       // debug, info, warn, error
	Format     string `yaml:"format"`      // file,text, json
	Output     string `yaml:"output"`      // stdout, file
	LogDir     string `yaml:"log_dir"`     // e.g., /var/log/baize
	MaxSizeMB  int    `yaml:"max_size"`    // per file, default 100
	MaxBackups int    `yaml:"max_backups"` // default 5
	MaxAgeDays int    `yaml:"max_age"`     // default 7
}

// DefaultConfig returns a Config struct populated with default values.
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level:      "info",
		Format:     "text",
		Output:     "stdout",
		LogDir:     "/var/log/baize",
		MaxSizeMB:  100,
		MaxBackups: 10,
		MaxAgeDays: 7,
	}
}
