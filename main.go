package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/khouwdevin/gitomatically/config"
	"github.com/khouwdevin/gitomatically/controller"
	"github.com/khouwdevin/gitomatically/env"
	"github.com/khouwdevin/gitomatically/middleware"
	"github.com/robfig/cron"
)

func main() {
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

	if config.Settings.Preference.Cron {
		ccron := cron.New()

		err := ccron.AddFunc(config.Settings.Preference.Spec, controller.CronController)

		if err != nil {
			slog.Error(fmt.Sprintf("CRON Error adding job %v", err))
			return
		}

		ccron.Start()

		slog.Info("CRON Cron jobs started")

		select {}
	} else {
		router := gin.Default()

		router.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Welcome to gitomatically!"})
		})

		router.POST("/webhook", middleware.GithubAuthorization(), controller.WebhookController)

		slog.Info("MAIN Gin running")

		router.Run(fmt.Sprintf(":%v", env.Env.PORT))
	}
}
