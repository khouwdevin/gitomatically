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

type GithubResponse struct {
	Repository struct {
		HtmlUrl string `json:"html_url"`
	}
	Ref string
}

func WebhookController(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Webhook receive"})

	event := c.GetHeader("X-GitHub-Event")

	if event != "push" {
		slog.Debug("WEBHOOK Not a push event")
		return
	}

	var response GithubResponse

	if err := c.BindJSON(&response); err != nil {
		return
	}

	var currentRepo config.RepositoryConfig

	htmlUrl := response.Repository.HtmlUrl

	for _, repository := range config.Settings.Repositories {
		if repository.Url == htmlUrl {
			currentRepo = repository
			break
		}
	}

	if reflect.DeepEqual(currentRepo, config.RepositoryConfig{}) {
		return
	}

	branch := strings.Split(response.Ref, "/")[2]

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
