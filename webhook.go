package main

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
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/khouwdevin/gitomatically/watcher"
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

	router.POST("/webhook", GithubAuthorization(), WebhookController)

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
	if watcher.GetSettingStatus() {
		return
	}

	watcher.ControllerGroup.Add(1)

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

	var currentRepo RepositoryConfig

	htmlUrl := response.Repository.HtmlUrl

	for _, repository := range Settings.Repositories {
		if repository.Url == htmlUrl {
			currentRepo = repository
			break
		}
	}

	if reflect.DeepEqual(currentRepo, RepositoryConfig{}) {
		slog.Debug("WEBHOOK Current repo is empty, return not continue the process")
		return
	}

	branch := strings.Split(response.Ref, "/")[2]

	if branch != currentRepo.Branch {
		slog.Debug(fmt.Sprintf("WEBHOOK Not the expected %v branch from response %v branch, skip pull and run commands", currentRepo.Branch, branch))
		return
	}

	publicKeys, err := ssh.NewPublicKeysFromFile("git", Settings.Preference.PrivateKey, Settings.Preference.Paraphrase)

	if err != nil {
		slog.Error(fmt.Sprintf("WEBHOOK New public keys from file error %v", err))
	}

	r, err := git.PlainOpen(currentRepo.Path)

	w, err := r.Worktree()

	if err != nil {
		slog.Error(fmt.Sprintf("WEBHOOK Get work tree error %v", err))
	}

	err = w.Pull(&git.PullOptions{RemoteName: "origin", Auth: publicKeys, Progress: os.Stdout})

	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			slog.Debug(fmt.Sprintf("WEBHOOK %v is up to date", currentRepo.Url))
			return
		} else {
			slog.Error(fmt.Sprintf("WEBHOOK Failed to pull branch %v", err))
			return
		}
	}

	for _, command := range currentRepo.Commands {
		slog.Debug(fmt.Sprintf("WEBHOOK Running %v", command))

		arrCommand := strings.Split(command, " ")

		cmd := exec.Command(arrCommand[0], arrCommand[1:]...)
		cmd.Dir = currentRepo.Path
		cmd.Env = os.Environ()

		output, err := cmd.Output()

		if err != nil {
			slog.Debug(fmt.Sprintf("WEBHOOK Command err output %v", string(output)))
			slog.Error(fmt.Sprintf("WEBHOOK Failed to run build command %v", arrCommand))
			return
		}
	}

	watcher.ControllerGroup.Done()
}
