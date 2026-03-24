package docker

import (
	"testing"
)

func TestParseLabels(t *testing.T) {
	tests := []struct {
		name     string
		labels   map[string]string
		expected ReloaderConfig
	}{
		{
			name:   "empty labels",
			labels: map[string]string{},
			expected: ReloaderConfig{
				Enabled: false,
				Mode:    ModeRestart,
				Signal:  "SIGHUP",
			},
		},
		{
			name: "fully configured",
			labels: map[string]string{
				LabelEnable:   "true",
				LabelMode:     ModeSignal,
				LabelSignal:   "SIGUSR1",
				LabelWatch:    "/app/config.json, /app/.env",
				LabelEndpoint: "http://localhost:8080/reload",
			},
			expected: ReloaderConfig{
				Enabled:  true,
				Mode:     ModeSignal,
				Signal:   "SIGUSR1",
				Watch:    []string{"/app/config.json", "/app/.env"},
				Endpoint: "http://localhost:8080/reload",
			},
		},
		{
			name: "enabled with defaults",
			labels: map[string]string{
				LabelEnable: "True",
			},
			expected: ReloaderConfig{
				Enabled: true,
				Mode:    ModeRestart,
				Signal:  "SIGHUP",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ParseLabels(tt.labels)
			if cfg.Enabled != tt.expected.Enabled {
				t.Errorf("expected Enabled %v, got %v", tt.expected.Enabled, cfg.Enabled)
			}
			if cfg.Mode != tt.expected.Mode {
				t.Errorf("expected Mode %v, got %v", tt.expected.Mode, cfg.Mode)
			}
			if cfg.Signal != tt.expected.Signal {
				t.Errorf("expected Signal %v, got %v", tt.expected.Signal, cfg.Signal)
			}
			if cfg.Endpoint != tt.expected.Endpoint {
				t.Errorf("expected Endpoint %v, got %v", tt.expected.Endpoint, cfg.Endpoint)
			}
			if len(cfg.Watch) != len(tt.expected.Watch) {
				t.Errorf("expected Watch len %d, got %d", len(tt.expected.Watch), len(cfg.Watch))
			} else {
				for i, v := range cfg.Watch {
					if v != tt.expected.Watch[i] {
						t.Errorf("expected Watch path %v, got %v", tt.expected.Watch[i], v)
					}
				}
			}
		})
	}
}
