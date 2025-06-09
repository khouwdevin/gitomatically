package watcher

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/khouwdevin/gitomatically/config"
	"github.com/khouwdevin/gitomatically/controller"
)

var (
	configTimer *time.Timer
)

func configDebouncedEvents(quit chan os.Signal) {
	if configTimer != nil {
		configTimer.Stop()
	}

	configTimer = time.AfterFunc(100*time.Millisecond, func() {
		slog.Info("WATCHER Config file change detected, reinitialize config")
		err := config.InitializeConfig("config.yaml")

		if err != nil {
			slog.Error(fmt.Sprintf("WATCHER Reinitialize config error %v", err))
			quit <- syscall.SIGTERM

			return
		}

		if config.Settings.Preference.Cron {
			err := controller.ShutdownServer()

			if err != nil {
				slog.Error(fmt.Sprintf("WATCHER Shutdown server error %v", err))
				quit <- syscall.SIGTERM

				return
			}

			controller.ChangeCron()
		} else {
			err := controller.NewServer()

			if err != nil {
				slog.Error(fmt.Sprintf("WATCHER Start server error %v", err))
				quit <- syscall.SIGTERM

				return
			}

			controller.StopCron()
		}
	})
}

func ConfigWatcher(stopChan chan struct{}, wg *sync.WaitGroup, quit chan os.Signal) error {
	configFilePath := "config.yaml"

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		watcher.Close()
		return errors.New(fmt.Sprintf("Watcher error %v", err))
	}

	absConfigPath, err := filepath.Abs(configFilePath)

	if err != nil {
		watcher.Close()
		return errors.New(fmt.Sprintf("Get absolute path error %v", err))
	}

	err = watcher.Add(absConfigPath)

	if err != nil {
		watcher.Close()
		return errors.New(fmt.Sprintf("WATCHER Add file to watcher error %v", err))
	}

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
					quit <- syscall.SIGTERM

					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0 {
					configDebouncedEvents(quit)
				} else if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
					slog.Error("WATCHER File is removed or renamed")
					quit <- syscall.SIGTERM

					return
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					slog.Error("WATCHER Watcher errors channel is closed")
					quit <- syscall.SIGTERM

					return
				}
				if err != nil {
					slog.Error(fmt.Sprintf("WATCHER Watcher error %v", err))
					quit <- syscall.SIGTERM

					return
				}

			case <-stopChan:
				wg.Done()

				return
			}
		}
	}()

	return nil
}
