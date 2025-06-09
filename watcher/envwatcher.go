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
	"github.com/khouwdevin/gitomatically/controller"
	"github.com/khouwdevin/gitomatically/env"
)

var (
	envTimer *time.Timer
)

func envDebouncedEvents(quit chan os.Signal) {
	if envTimer != nil {
		envTimer.Stop()
	}

	envTimer = time.AfterFunc(100*time.Millisecond, func() {
		slog.Info("WATCHER Env file change detected, reinitialize env")
		err := env.InitializeEnv()

		if err != nil {
			slog.Error(fmt.Sprintf("WATCHER Reinitialize env error %v", err))
			quit <- syscall.SIGTERM

			return
		}

		err = controller.NewServer()

		if err != nil {
			slog.Error(fmt.Sprintf("WATCHER Restart server error %v", err))
			quit <- syscall.SIGTERM

			return
		}

		slog.SetLogLoggerLevel(slog.Level(env.Env.LOG_LEVEL))
	})
}

func EnvWatcher(stopChan chan struct{}, wg *sync.WaitGroup, quit chan os.Signal) error {
	envFilePath := ".env"

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		watcher.Close()
		return errors.New(fmt.Sprintf("Watcher error %v", err))
	}

	absConfigPath, err := filepath.Abs(envFilePath)

	if err != nil {
		watcher.Close()
		return errors.New(fmt.Sprintf("Get absolute path error %v", err))
	}

	err = watcher.Add(absConfigPath)

	if err != nil {
		watcher.Close()
		return errors.New(fmt.Sprintf("Add file to watcher error %v", err))
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
				if event.Op&fsnotify.Write == fsnotify.Write {
					envDebouncedEvents(quit)
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
