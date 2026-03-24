package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/dockwatch/internal/config"
	"github.com/dockwatch/internal/docker"
	"github.com/dockwatch/internal/reloader"
	"github.com/dockwatch/internal/watcher"
	"github.com/dockwatch/pkg/logger"
)

func main() {
	cfg := config.LoadConfig()

	if err := logger.InitLogger(cfg.LogLevel); err != nil {
		os.Exit(1)
	}

	logger.Log.Info("Starting Dockwatch...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		logger.Log.Info("Serving metrics on :9090/metrics")
		if err := http.ListenAndServe(":9090", nil); err != nil {
			logger.Log.Errorf("Metrics server failed: %v", err)
		}
	}()

	dockerClient, err := docker.NewClient()
	if err != nil {
		logger.Log.Fatalf("Failed to initialize Docker client: %v", err)
	}

	watchEngine, err := watcher.NewWatcher(cfg.Interval)
	if err != nil {
		logger.Log.Fatalf("Failed to initialize Watcher Engine: %v", err)
	}
	defer watchEngine.Stop()

	err = watchEngine.AddPath(cfg.WatchDir)
	if err != nil {
		logger.Log.Fatalf("Failed to watch directory %s: %v", cfg.WatchDir, err)
	}

	watchEngine.Start()

	reloadEngine := reloader.NewEngine(dockerClient, watchEngine, 10*time.Second)
	go reloadEngine.Start(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	logger.Log.Infof("Received signal: %v. Shutting down gracefully...", sig)
}
