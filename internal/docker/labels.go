package docker

import (
	"strings"
)

const (
	LabelEnable   = "reloader.enable"
	LabelMode     = "reloader.mode"
	LabelSignal   = "reloader.signal"
	LabelEndpoint = "reloader.endpoint"
	LabelWatch    = "reloader.watch"

	ModeRestart = "restart"
	ModeSignal  = "signal"
	ModeHTTP    = "http"
)

// ReloaderConfig holds the parsed reloader labels for a container
type ReloaderConfig struct {
	Enabled  bool
	Mode     string
	Signal   string
	Endpoint string
	Watch    []string
}

// ParseLabels parses a map of container labels into a ReloaderConfig
func ParseLabels(labels map[string]string) ReloaderConfig {
	cfg := ReloaderConfig{
		Enabled: false,
		Mode:    ModeRestart, // Default mode
		Signal:  "SIGHUP",    // Default signal
	}

	if val, ok := labels[LabelEnable]; ok && strings.ToLower(val) == "true" {
		cfg.Enabled = true
	}

	if val, ok := labels[LabelMode]; ok && val != "" {
		cfg.Mode = val
	}

	if val, ok := labels[LabelSignal]; ok && val != "" {
		cfg.Signal = val
	}

	if val, ok := labels[LabelEndpoint]; ok && val != "" {
		cfg.Endpoint = val
	}

	if val, ok := labels[LabelWatch]; ok && val != "" {
		paths := strings.Split(val, ",")
		for _, p := range paths {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				cfg.Watch = append(cfg.Watch, trimmed)
			}
		}
	}

	return cfg
}
