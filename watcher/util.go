package watcher

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/khouwdevin/gitomatically/config"
	"github.com/khouwdevin/gitomatically/controller"
	"github.com/khouwdevin/gitomatically/env"
)

func EnvDebouncedEvents(w *Watcher) {
	if w.Self.Timer != nil {
		w.Self.Timer.Stop()
	}

	w.Self.Timer = time.AfterFunc(100*time.Millisecond, func() {
		controller.UpdateSettingStatus(true)
		controller.ControllerGroup.Wait()

		slog.Info("WATCHER Env file change detected, reinitialize env")

		prevPort := os.Getenv("PORT")
		err := env.InitializeEnv(w.Self.Path)
		currentPort := os.Getenv("PORT")

		if err != nil {
			slog.Error(fmt.Sprintf("WATCHER Reinitialize env error %v", err))
			w.quit <- syscall.SIGTERM

			return
		}

		if !config.Settings.Preference.Cron && prevPort != currentPort {
			err = controller.NewServer()

			if err != nil {
				slog.Error(fmt.Sprintf("WATCHER Restart server error %v", err))
				w.quit <- syscall.SIGTERM

				return
			}
		}

		LOG_LEVEL_INT, err := strconv.Atoi(os.Getenv("LOG_LEVEL"))

		if err != nil {
			slog.Error(fmt.Sprintf("MAIN Error convert string to int %v", err))
			return
		}

		slog.SetLogLoggerLevel(slog.Level(LOG_LEVEL_INT))

		controller.UpdateSettingStatus(false)
	})
}

func ConfigDebouncedEvents(w *Watcher) {
	if w.Self.Timer != nil {
		w.Self.Timer.Stop()
	}

	w.Self.Timer = time.AfterFunc(100*time.Millisecond, func() {
		controller.UpdateSettingStatus(true)
		controller.ControllerGroup.Wait()

		slog.Info("WATCHER Config file change detected, reinitialize config")

		prevConfig := config.Settings
		err := config.InitializeConfig(w.Self.Path)

		if err != nil {
			slog.Error(fmt.Sprintf("WATCHER Reinitialize config error %v", err))
			w.quit <- syscall.SIGTERM

			return
		}

		err = config.PreStart()

		if err != nil {
			slog.Error(fmt.Sprintf("WATCHER Rerun prestart error %v", err))
			w.quit <- syscall.SIGTERM

			return
		}

		if prevConfig.Preference.Cron == config.Settings.Preference.Cron &&
			(!config.Settings.Preference.Cron || prevConfig.Preference.Spec == config.Settings.Preference.Spec) {
			return
		}

		if config.Settings.Preference.Cron {
			err := controller.ShutdownServer()

			if err != nil {
				slog.Error(fmt.Sprintf("WATCHER Shutdown server error %v", err))
				w.quit <- syscall.SIGTERM

				return
			}

			controller.ChangeCron()
		} else {
			controller.StopCron()

			err := controller.NewServer()

			if err != nil {
				slog.Error(fmt.Sprintf("WATCHER Start server error %v", err))
				w.quit <- syscall.SIGTERM

				return
			}
		}

		controller.UpdateSettingStatus(false)
	})
}
