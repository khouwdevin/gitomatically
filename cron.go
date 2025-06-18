package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/khouwdevin/gitomatically/watcher"
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

	err := Ccron.AddFunc(Settings.Preference.Spec, CronController)

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
	if watcher.GetSettingStatus() {
		return
	}

	watcher.ControllerGroup.Add(1)

	slog.Debug("CRON Rerun all config")

	for _, repository := range Settings.Repositories {
		publicKeys, err := ssh.NewPublicKeysFromFile("git", Settings.Preference.PrivateKey, Settings.Preference.Paraphrase)

		if err != nil {
			slog.Error(fmt.Sprintf("CRON New public keys from file error %v", err))
		}

		r, err := git.PlainOpen(repository.Path)

		w, err := r.Worktree()

		if err != nil {
			slog.Error(fmt.Sprintf("CRON Get work tree error %v", err))
		}

		err = w.Pull(&git.PullOptions{RemoteName: "origin", Auth: publicKeys, Progress: os.Stdout})

		if err != nil {
			if err == git.NoErrAlreadyUpToDate {
				slog.Debug(fmt.Sprintf("CRON %v is up to date, continue to next repository", repository.Url))
				continue
			} else {
				slog.Debug(fmt.Sprintf("CRON Git pull err output %v", err))
			}
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

	watcher.ControllerGroup.Done()
}
