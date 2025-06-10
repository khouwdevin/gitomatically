package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
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

	err := env.InitializeEnv(".env")

	if err != nil {
		slog.Error(fmt.Sprintf("MAIN Initialize env error %v", err))
		return
	}

	LOG_LEVEL_INT, err := strconv.Atoi(os.Getenv("LOG_LEVEL"))

	if err != nil {
		slog.Error(fmt.Sprintf("MAIN Error convert string to int %v", err))
		return
	}

	slog.SetLogLoggerLevel(slog.Level(LOG_LEVEL_INT))

	slog.Info("MAIN Powered by khouwdevin.com")

	slog.Info("MAIN Initialize env")

	slog.Info("MAIN Initialize config")
	err = config.InitializeConfig("config.yaml")

	if err != nil {
		slog.Error(fmt.Sprintf("MAIN Initialize config error %v", err))
		return
	}

	err = config.PreStart()

	if err != nil {
		slog.Error(fmt.Sprintf("MAIN Prestart error %v", err))
	}

	if !config.Settings.Preference.Cron && os.Getenv("GITHUB_WEBHOOK_SECRET") == "" {
		slog.Error("MAIN Github webhook secret is required")
		return
	}

	configStopChan := make(chan struct{})
	envStopChan := make(chan struct{})

	err = watcher.ConfigWatcher("config.yaml", configStopChan, &wg, quit)

	if err != nil {
		slog.Error(fmt.Sprintf("MAIN %v", err))
		return
	}

	err = watcher.EnvWatcher(".env", envStopChan, &wg, quit)

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
