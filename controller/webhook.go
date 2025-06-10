package controller

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/khouwdevin/gitomatically/config"
	"github.com/khouwdevin/gitomatically/middleware"
)

type RepositoryStruct struct {
	HtmlUrl string `json:"html_url"`
}

type GithubResponse struct {
	Repository RepositoryStruct `json:"repository"`
	Ref        string           `json:"ref"`
}

var (
	Server *http.Server
)

func NewServer() error {
	if Server != nil {
		ShutdownServer()
	}

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Welcome to gitomatically!"})
	})

	router.POST("/webhook", middleware.GithubAuthorization(), WebhookController)

	slog.Info("MAIN Gin running")

	Server = &http.Server{
		Addr:    fmt.Sprintf(":%v", os.Getenv("PORT")),
		Handler: router,
	}

	go func() {
		if err := Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error(fmt.Sprintf("Gin server error %v", err))
		}
	}()

	return nil
}

func ShutdownServer() error {
	if Server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := Server.Shutdown(ctx)

	if err != nil {
		return err
	}

	Server = nil

	slog.Info("WEBHOOK Server is off")

	return nil
}

func WebhookController(c *gin.Context) {
	event := c.GetHeader("X-GitHub-Event")

	if event != "push" {
		slog.Debug("WEBHOOK Not a push event, return not continue the process")
		return
	}

	var response GithubResponse

	if err := c.BindJSON(&response); err != nil {
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook receive"})

	var currentRepo config.RepositoryConfig

	htmlUrl := response.Repository.HtmlUrl

	for _, repository := range config.Settings.Repositories {
		if repository.Url == htmlUrl {
			currentRepo = repository
			break
		}
	}

	if reflect.DeepEqual(currentRepo, config.RepositoryConfig{}) {
		slog.Debug("WEBHOOK Current repo is empty, return not continue the process")
		return
	}

	branch := strings.Split(response.Ref, "/")[2]

	if branch != currentRepo.Branch {
		slog.Debug(fmt.Sprintf("WEBHOOK Not the expected %v branch from response %v branch, skip pull and run commands", currentRepo.Branch, branch))
		return
	}

	git := exec.Command("git", "pull")
	git.Dir = currentRepo.Path
	git.Env = os.Environ()

	stdout, err := git.Output()
	output := strings.TrimSpace(string(stdout))

	if err != nil {
		slog.Error(fmt.Sprintf("WEBHOOK Failed to pull branch %v", err))
		slog.Error(fmt.Sprintf("WEBHOOK Git output %v", output))

		return
	}

	for _, command := range currentRepo.Commands {
		slog.Debug(fmt.Sprintf("WEBHOOK Running %v", command))

		arrCommand := strings.Split(command, " ")

		cmd := exec.Command(arrCommand[0], arrCommand[1:]...)
		cmd.Dir = currentRepo.Path
		cmd.Env = os.Environ()

		_, err := cmd.Output()

		if err != nil {
			slog.Error(fmt.Sprintf("WEBHOOK Failed to run build command %v", arrCommand))
			return
		}
	}
}
