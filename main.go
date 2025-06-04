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
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelInfo)

	slog.Info("Powered by khouwdevin.com")

	slog.Info("Initialize env")
	err := env.InitializeEnv()

	if err != nil {
		slog.Error(fmt.Sprintf("Initialize env error %v", err))
		return
	}

	slog.Info("Initialize config")
	err = config.InitializeConfig()

	if err != nil {
		slog.Error(fmt.Sprintf("Initialize config error %v", err))
		return
	}

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Welcome to gitomatically!"})
	})

	router.POST("/webhook", middleware.GithubAuthorization(), controller.WebhookController)

	slog.Info("Gin running")

	router.Run(":8080")
}
