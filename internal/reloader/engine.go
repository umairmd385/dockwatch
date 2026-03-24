package reloader

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dockwatch/internal/docker"
	"github.com/dockwatch/internal/watcher"
	"github.com/dockwatch/pkg/logger"
)

type Engine struct {
	dockerClient *docker.Client
	watcher      *watcher.Watcher
	cooldowns    sync.Map // struct: containerID -> time.Time
	cooldownDur  time.Duration
}

func NewEngine(dc *docker.Client, w *watcher.Watcher, cooldown time.Duration) *Engine {
	return &Engine{
		dockerClient: dc,
		watcher:      w,
		cooldownDur:  cooldown,
	}
}

func (e *Engine) Start(ctx context.Context) {
	logger.Log.Info("Starting Reload Engine...")

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Stopping Reload Engine...")
			return
		case event := <-e.watcher.Events:
			logger.Log.Infof("Processing event for file: %s", event.Path)
			FileChangesTotal.WithLabelValues(event.Path).Inc()
			e.processEvent(ctx, event.Path)
		}
	}
}

func (e *Engine) processEvent(ctx context.Context, changedPath string) {
	containers, err := e.dockerClient.ListContainers(ctx)
	if err != nil {
		logger.Log.Errorf("Failed to list containers: %v", err)
		return
	}

	for _, c := range containers {
		cfg := docker.ParseLabels(c.Labels)
		if !cfg.Enabled {
			continue
		}

		matched := false
		for _, watchPath := range cfg.Watch {
			cleanWatchPath := filepath.Clean(watchPath)
			cleanChangedPath := filepath.Clean(changedPath)

			if cleanChangedPath == cleanWatchPath || strings.HasSuffix(cleanChangedPath, cleanWatchPath) {
				matched = true
				break
			}
		}

		if matched {
			e.triggerReload(ctx, c.ID, c.Names[0], cfg)
		}
	}
}

func (e *Engine) triggerReload(ctx context.Context, containerID, containerName string, cfg docker.ReloaderConfig) {
	if e.inCooldown(containerID) {
		logger.Log.Warnf("Skipping reload for container %s (in cooldown)", containerName)
		return
	}
	e.setCooldown(containerID)

	logger.Log.Infof("Triggering %s on container %s", cfg.Mode, containerName)

	var err error
	action := cfg.Mode

	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		switch cfg.Mode {
		case docker.ModeRestart:
			err = e.dockerClient.Restart(ctx, containerID)
		case docker.ModeSignal:
			err = e.dockerClient.Signal(ctx, containerID, cfg.Signal)
		case docker.ModeHTTP:
			err = e.dockerClient.ExecHTTP(ctx, containerID, cfg.Endpoint)
		default:
			logger.Log.Warnf("Unknown reload mode: %s on container %s", cfg.Mode, containerName)
			return
		}

		if err == nil {
			break
		}
		logger.Log.Warnf("Failed to trigger %s on %s (attempt %d/%d): %v", cfg.Mode, containerName, i+1, maxRetries, err)
		time.Sleep(2 * time.Second)
	}

	status := "success"
	if err != nil {
		status = "error"
		logger.Log.Errorf("Final failure triggering %s on %s: %v", cfg.Mode, containerName, err)
	}

	ReloadsTotal.WithLabelValues(containerName, action, status).Inc()
}

func (e *Engine) inCooldown(containerID string) bool {
	if val, ok := e.cooldowns.Load(containerID); ok {
		lastReload := val.(time.Time)
		if time.Since(lastReload) < e.cooldownDur {
			return true
		}
	}
	return false
}

func (e *Engine) setCooldown(containerID string) {
	e.cooldowns.Store(containerID, time.Now())
}
