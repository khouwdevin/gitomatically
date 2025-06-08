package watcher

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/khouwdevin/gitomatically/controller"
	"github.com/khouwdevin/gitomatically/env"
)

func EnvWatcher() {
	envFilePath := ".env"

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		slog.Error(fmt.Sprintf("WATCHER Watcher error %v", err))
		return
	}

	absConfigPath, err := filepath.Abs(envFilePath)

	if err != nil {
		slog.Error(fmt.Sprintf("WATCHER Get absolute path error %v", err))
		return
	}

	err = watcher.Add(absConfigPath)

	if err != nil {
		slog.Error(fmt.Sprintf("WATCHER Add file to watcher error %v", err))
		return
	}

	debouncedEvents := make(chan struct{}, 1)

	go func() {
		var timer *time.Timer
		for {
			select {
			case <-debouncedEvents:
				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(100*time.Millisecond, func() {
					slog.Info("WATCHER Env file change detected, reinitialize WATCHER")
					err := env.InitializeEnv()

					if err != nil {
						slog.Error(fmt.Sprintf("WATCHER Reinitialize env error %v", err))
						return
					}

					err = controller.NewServer()

					if err != nil {
						slog.Error(fmt.Sprintf("WATCHER Restart server error %v", err))
						return
					}

					slog.SetLogLoggerLevel(slog.Level(env.Env.LOG_LEVEL))
				})
			}
		}
	}()

	go func() {
		defer func() {
			slog.Debug("WATCHER Watcher goroutine stopped.")
			watcher.Close()
		}()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					slog.Error("WATCHER Watcher events channel is closed")
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					select {
					case debouncedEvents <- struct{}{}:
					default:
					}
				} else if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
					slog.Error("WATCHER File is removed or renamed")
					return
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					slog.Error("WATCHER Watcher errors channel is closed")
					return
				}
				if err != nil {
					slog.Error(fmt.Sprintf("WATCHER Watcher error %v", err))
				}

			}
		}
	}()
}
