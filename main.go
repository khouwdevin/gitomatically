package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/khouwdevin/gitomatically/config"
	"github.com/khouwdevin/gitomatically/controller"
	"github.com/khouwdevin/gitomatically/env"
	"github.com/khouwdevin/gitomatically/watcher"
)

func main() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	slog.SetLogLoggerLevel(slog.LevelInfo)

	err := env.InitializeEnv()

	if err != nil {
		slog.Error(fmt.Sprintf("MAIN Initialize env error %v", err))
		return
	}

	slog.SetLogLoggerLevel(slog.Level(env.Env.LOG_LEVEL))

	slog.Info("MAIN Powered by khouwdevin.com")

	slog.Info("MAIN Initialize env")

	slog.Info("MAIN Initialize config")
	err = config.InitializeConfig()

	if err != nil {
		slog.Error(fmt.Sprintf("MAIN Initialize config error %v", err))
		return
	}

	watcher.ConfigWatcher()
	watcher.EnvWatcher()

	if config.Settings.Preference.Cron {
		err := controller.NewCron()

		if err != nil {
			slog.Error(fmt.Sprintf("CRON Create new cron error %v", err))
		}

		<-quit

		err = controller.StopCron()

		if err != nil {
			slog.Error(fmt.Sprintf("CRON Stopping cron error %v", err))
		}
	} else {
		err := controller.NewServer()

		if err != nil {
			slog.Error(fmt.Sprintf("MAIN Server error %v", err))
		}

		<-quit

		err = controller.ShutdownServer()

		if err != nil {
			slog.Error(fmt.Sprintf("MAIN Shutdown server error %v", err))
		}
	}
}
