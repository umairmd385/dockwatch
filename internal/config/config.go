package config

import (
	"flag"
	"time"
)

// AppConfig holds the configuration for the application
type AppConfig struct {
	WatchDir string
	Interval time.Duration
	LogLevel string
}

// LoadConfig parses command-line flags and returns the AppConfig
func LoadConfig() *AppConfig {
	cfg := &AppConfig{}

	flag.StringVar(&cfg.WatchDir, "watch-dir", "/secrets", "Directory to watch for file changes")
	flag.DurationVar(&cfg.Interval, "interval", 5*time.Second, "Debounce interval for file changes")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "Logging level (debug, info, warn, error)")

	flag.Parse()

	return cfg
}
