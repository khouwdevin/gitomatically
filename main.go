package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/khouwdevin/gitomatically/config"
	"github.com/khouwdevin/gitomatically/controller"
	"github.com/khouwdevin/gitomatically/env"
	"github.com/khouwdevin/gitomatically/middleware"
)

func main() {
	err := env.InitializeEnv()

	if err != nil {
		log.Fatalf("Initialize env error %v", err)
		return
	}

	err = config.InitializeConfig()

	if err != nil {
		log.Fatalf("Initialize config error %v", err)
		return
	}

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Welcome to gitomatically!"})
	})

	router.POST("/webhook", middleware.GithubAuthorization(), controller.WebhookController)

	router.Run(":8080")
}
