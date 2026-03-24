package watcher

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/dockwatch/pkg/logger"
	"github.com/fsnotify/fsnotify"
)

type Event struct {
	Path string
}

type Watcher struct {
	fw         *fsnotify.Watcher
	interval   time.Duration
	Events     chan Event
	fileHashes map[string]string
	timers     map[string]*time.Timer
	mu         sync.Mutex
	done       chan struct{}
}

func NewWatcher(interval time.Duration) (*Watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}

	w := &Watcher{
		fw:         fw,
		interval:   interval,
		Events:     make(chan Event, 100),
		fileHashes: make(map[string]string),
		timers:     make(map[string]*time.Timer),
		done:       make(chan struct{}),
	}

	return w, nil
}

func (w *Watcher) AddPath(path string) error {
	info, err := os.Stat(path)
	if err == nil && !info.IsDir() {
		hash, _ := CalculateHash(path)
		w.mu.Lock()
		w.fileHashes[path] = hash
		w.mu.Unlock()
	}

	err = w.fw.Add(path)
	if err != nil {
		return fmt.Errorf("failed to add path %s to watcher: %w", path, err)
	}
	logger.Log.Debugf("Added watch for path: %s", path)
	return nil
}

func (w *Watcher) Start() {
	go func() {
		for {
			select {
			case event, ok := <-w.fw.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					w.handleEvent(event.Name)
				}
			case err, ok := <-w.fw.Errors:
				if !ok {
					return
				}
				logger.Log.Errorf("Watcher error: %v", err)
			case <-w.done:
				return
			}
		}
	}()
}

func (w *Watcher) Stop() error {
	close(w.done)
	return w.fw.Close()
}

func (w *Watcher) handleEvent(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if timer, exists := w.timers[path]; exists {
		timer.Stop()
	}

	w.timers[path] = time.AfterFunc(w.interval, func() {
		w.processChange(path)
	})
}

func (w *Watcher) processChange(path string) {
	hash, err := CalculateHash(path)
	if err != nil {
		logger.Log.Warnf("Failed to calculate hash for %s: %v", path, err)
		return
	}

	w.mu.Lock()
	oldHash := w.fileHashes[path]
	if hash != oldHash {
		w.fileHashes[path] = hash

		// Run channel dispatch out of lock to prevent blocking
		w.mu.Unlock()
		logger.Log.Infof("Detected genuine change in file: %s", path)
		w.Events <- Event{Path: path}
		w.mu.Lock()
	} else {
		logger.Log.Debugf("File %s written but hash is unchanged", path)
	}
	delete(w.timers, path)
	w.mu.Unlock()
}

// CalculateHash calculates the SHA256 hash of a file
func CalculateHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
