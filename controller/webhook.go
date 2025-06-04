package controller

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/khouwdevin/gitomatically/config"
)

func WebhookController(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{"message": "Request receive"})

	var payload map[string]any

	if err := c.BindJSON(&payload); err != nil {
		return
	}

	var currentRepo config.RepositoryConfig

	repository := payload["repository"].(map[string]any)
	htmlUrl := repository["html_url"].(string)

	for _, repository := range config.Settings.Repositories {
		if repository.Url == htmlUrl {
			currentRepo = repository
			break
		}
	}

	if reflect.DeepEqual(currentRepo, config.RepositoryConfig{}) {
		return
	}

	ref := payload["ref"].(string)
	branch := strings.Split(ref, "/")[2]

	if branch != currentRepo.Branch {
		return
	}

	git := exec.Command("git", "pull")
	git.Dir = currentRepo.Path
	git.Env = os.Environ()

	_, err := git.Output()

	if err != nil {
		slog.Error(fmt.Sprintf("Failed to pull branch %v", err))
		return
	}

	for _, buildCommands := range currentRepo.BuildCommands {
		build := strings.Split(buildCommands, " ")

		cmd := exec.Command(build[0], build[1:]...)
		cmd.Dir = currentRepo.Path
		cmd.Env = os.Environ()

		_, err := cmd.Output()

		if err != nil {
			slog.Error(fmt.Sprintf("Failed to run build command %v", build))
			return
		}
	}

}
