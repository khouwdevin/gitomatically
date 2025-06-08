package watcher

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/khouwdevin/gitomatically/config"
	"github.com/khouwdevin/gitomatically/controller"
)

func ConfigWatcher() {
	configFilePath := "config.yaml"

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		watcher.Close()
		slog.Error(fmt.Sprintf("WATCHER Watcher error %v", err))
		return
	}

	absConfigPath, err := filepath.Abs(configFilePath)

	if err != nil {
		watcher.Close()
		slog.Error(fmt.Sprintf("WATCHER Get absolute path error %v", err))
		return
	}

	err = watcher.Add(absConfigPath)

	if err != nil {
		watcher.Close()
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
					slog.Info("WATCHER Config file change detected, reinitialize config")
					err := config.InitializeConfig()

					if err != nil {
						slog.Error(fmt.Sprintf("WATCHER Reinitialize config error %v", err))
						return
					}

					if config.Settings.Preference.Cron {
						err := controller.ShutdownServer()

						if err != nil {
							slog.Error(fmt.Sprintf("WATCHER Shutdown server error %v", err))
							return
						}

						controller.ChangeCron()
					} else {
						err := controller.NewServer()

						if err != nil {
							slog.Error(fmt.Sprintf("WATCHER Start server error %v", err))
							return
						}

						controller.StopCron()
					}
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
				if event.Op&(fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0 {
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
