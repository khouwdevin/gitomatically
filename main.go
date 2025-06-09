package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/khouwdevin/gitomatically/config"
	"github.com/khouwdevin/gitomatically/controller"
	"github.com/khouwdevin/gitomatically/env"
	"github.com/khouwdevin/gitomatically/watcher"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

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

	configStopChan := make(chan struct{})
	envStopChan := make(chan struct{})

	err = watcher.ConfigWatcher(configStopChan, &wg, quit)

	if err != nil {
		slog.Error(fmt.Sprintf("MAIN %v", err))
		return
	}

	err = watcher.EnvWatcher(envStopChan, &wg, quit)

	if err != nil {
		slog.Error(fmt.Sprintf("MAIN %v", err))
		return
	}

	if config.Settings.Preference.Cron {
		err := controller.NewCron()

		if err != nil {
			slog.Error(fmt.Sprintf("MAIN Create new cron error %v", err))
			return
		}
	} else {
		err := controller.NewServer()

		if err != nil {
			slog.Error(fmt.Sprintf("MAIN Server error %v", err))
			return
		}
	}

	<-quit

	if config.Settings.Preference.Cron {
		err = controller.StopCron()

		if err != nil {
			slog.Error(fmt.Sprintf("MAIN Stopping cron error %v", err))
		}
	} else {
		err = controller.ShutdownServer()

		if err != nil {
			slog.Error(fmt.Sprintf("MAIN Shutdown server error %v", err))
		}
	}

	close(configStopChan)
	close(envStopChan)

	slog.Debug("MAIN Closing watcher for config and env channel")

	wg.Wait()
}
