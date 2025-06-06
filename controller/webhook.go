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
	c.JSON(http.StatusOK, gin.H{"message": "Webhook receive"})

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
		slog.Error(fmt.Sprintf("WEBHOOK Failed to pull branch %v", err))
		return
	}

	for _, command := range currentRepo.Commands {
		slog.Debug(fmt.Sprintf("Running %v", command))

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
