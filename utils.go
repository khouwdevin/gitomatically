package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/khouwdevin/gitomatically/watcher"
)

func EnvDebouncedEvents(w *watcher.Watcher) {
	if w.Self.Timer != nil {
		w.Self.Timer.Stop()
	}

	w.Self.Timer = time.AfterFunc(100*time.Millisecond, func() {
		watcher.UpdateSettingStatus(true)
		watcher.ControllerGroup.Wait()

		slog.Info("WATCHER Env file change detected, reinitialize env")

		prevPort := os.Getenv("PORT")
		err := InitializeEnv(w.Self.Path)
		currentPort := os.Getenv("PORT")

		if err != nil {
			slog.Error(fmt.Sprintf("WATCHER Reinitialize env error %v", err))
			w.Quit <- syscall.SIGTERM

			return
		}

		if !Settings.Preference.Cron && prevPort != currentPort {
			err = NewServer()

			if err != nil {
				slog.Error(fmt.Sprintf("WATCHER Restart server error %v", err))
				w.Quit <- syscall.SIGTERM

				return
			}
		}

		LOG_LEVEL_INT, err := strconv.Atoi(os.Getenv("LOG_LEVEL"))

		if err != nil {
			slog.Error(fmt.Sprintf("MAIN Error convert string to int %v", err))
			return
		}

		slog.SetLogLoggerLevel(slog.Level(LOG_LEVEL_INT))

		watcher.UpdateSettingStatus(false)
	})
}

func ConfigDebouncedEvents(w *watcher.Watcher) {
	if w.Self.Timer != nil {
		w.Self.Timer.Stop()
	}

	w.Self.Timer = time.AfterFunc(100*time.Millisecond, func() {
		watcher.UpdateSettingStatus(true)
		watcher.ControllerGroup.Wait()

		slog.Info("WATCHER Config file change detected, reinitialize config")

		prevConfig := Settings
		err := InitializeConfig(w.Self.Path)

		if err != nil {
			slog.Error(fmt.Sprintf("WATCHER Reinitialize config error %v", err))
			w.Quit <- syscall.SIGTERM

			return
		}

		err = PreStart()

		if err != nil {
			slog.Error(fmt.Sprintf("WATCHER Rerun prestart error %v", err))
			w.Quit <- syscall.SIGTERM

			return
		}

		if prevConfig.Preference.Cron == Settings.Preference.Cron &&
			(!Settings.Preference.Cron || prevConfig.Preference.Spec == Settings.Preference.Spec) {
			return
		}

		if Settings.Preference.Cron {
			err := ShutdownServer()

			if err != nil {
				slog.Error(fmt.Sprintf("WATCHER Shutdown server error %v", err))
				w.Quit <- syscall.SIGTERM

				return
			}

			ChangeCron()
		} else {
			StopCron()

			err := NewServer()

			if err != nil {
				slog.Error(fmt.Sprintf("WATCHER Start server error %v", err))
				w.Quit <- syscall.SIGTERM

				return
			}
		}

		watcher.UpdateSettingStatus(false)
	})
}
