package controller

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/khouwdevin/gitomatically/config"
)

func CronController() {
	slog.Debug("CRON Rerun all config")

	for _, repository := range config.Settings.Repositories {
		git := exec.Command("git", "pull")
		git.Dir = repository.Path
		git.Env = os.Environ()

		stdout, err := git.Output()

		if err != nil {
			slog.Error(fmt.Sprintf("CRON Git pull error %v", err))
		}

		output := strings.TrimSpace(string(stdout))

		if strings.Contains(output, "Already up to date.") {
			slog.Debug("CRON Git is up to date, continue to next repository")
			continue
		}

		for _, command := range repository.Commands {
			slog.Debug(fmt.Sprintf("CRON Running %v", command))

			arrCommand := strings.Split(command, " ")

			cmd := exec.Command(arrCommand[0], arrCommand[1:]...)
			cmd.Dir = repository.Path
			cmd.Env = os.Environ()

			_, err := cmd.Output()

			if err != nil {
				slog.Error("CRON Failed to run command")
			}
		}
	}

	slog.Debug("CRON Cron finished")
}
