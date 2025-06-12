package controller

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/khouwdevin/gitomatically/config"
	"github.com/robfig/cron"
)

var (
	Ccron     *cron.Cron
	cronMutex sync.RWMutex
)

func NewCron() error {
	cronMutex.Lock()
	defer cronMutex.Unlock()

	if Ccron != nil {
		slog.Debug("CRON Stopping existing cron instance before creating new one")
		Ccron.Stop()
	}

	Ccron = cron.New()

	err := Ccron.AddFunc(config.Settings.Preference.Spec, CronController)

	if err != nil {
		return err
	}

	Ccron.Start()

	slog.Info("CRON Cron jobs started")

	return nil
}

func ChangeCron() error {
	if Ccron == nil {
		NewCron()
		return nil
	}

	StopCron()
	NewCron()

	return nil
}

func StopCron() error {
	cronMutex.Lock()
	defer cronMutex.Unlock()

	if Ccron == nil {
		return nil
	}

	Ccron.Stop()

	Ccron = nil

	slog.Info("CRON Stopping current cron")

	return nil
}

func CronController() {
	if GetSettingStatus() {
		return
	}

	ControllerGroup.Add(1)

	slog.Debug("CRON Rerun all config")

	for _, repository := range config.Settings.Repositories {
		git := exec.Command("git", "pull")
		git.Dir = repository.Path
		git.Env = os.Environ()

		stdout, err := git.Output()
		output := strings.TrimSpace(string(stdout))

		if err != nil {
			slog.Error(fmt.Sprintf("CRON Git pull error %v", err))
			slog.Error(fmt.Sprintf("CRON Git output %v", output))

			continue
		}

		if strings.Contains(output, "Already up to date.") {
			slog.Debug(fmt.Sprintf("CRON %v is up to date, continue to next repository", repository.Url))
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
				break
			}
		}
	}

	slog.Debug("CRON Cron finished")

	ControllerGroup.Done()
}
