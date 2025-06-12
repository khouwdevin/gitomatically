package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	Settings Config
)

func InitializeConfig(filePath string) error {
	yamlFile, err := os.ReadFile(filePath)

	if err != nil {
		return err
	}

	err = yaml.Unmarshal(yamlFile, &Settings)

	if err != nil {
		return err
	}

	if len(Settings.Repositories) == 0 {
		return errors.New("there is no repository in config.")
	}
	if Settings.Preference.Cron && Settings.Preference.Spec == "" {
		return errors.New("duration value is required.")
	}

	return nil
}

func PreStart() error {
	for _, repository := range Settings.Repositories {
		_, err := os.Stat(filepath.Join(repository.Path, ".git"))

		repoName := filepath.Base(repository.Path)

		if err == nil {
			slog.Debug(fmt.Sprintf("CONFIG Pulling %v", repository.Url))

			git := exec.Command("git", "pull")
			git.Dir = repository.Path
			git.Env = os.Environ()

			stdout, err := git.Output()
			output := strings.TrimSpace(string(stdout))

			if err != nil {
				slog.Debug(fmt.Sprintf("CONFIG Git clone err output %v", string(output)))
				return err
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
		} else if os.IsNotExist(err) {
			slog.Info(fmt.Sprintf("CONFIG Cloning %v", repository.Url))
			err = os.RemoveAll(repository.Path)

			if err != nil {
				return err
			}

			dir := filepath.Dir(repository.Path)
			dirPerms := os.FileMode(0755)
			err := os.MkdirAll(dir, dirPerms)

			if err != nil {
				return err
			}

			git := exec.Command("git", "clone", repository.Clone, repoName)
			git.Dir = dir
			git.Env = os.Environ()

			output, err := git.Output()

			if err != nil {
				slog.Debug(fmt.Sprintf("CONFIG Git clone err output %v", string(output)))
				return err
			}

			for _, command := range repository.Commands {
				slog.Debug(fmt.Sprintf("CONFIG Running %v", command))

				arrCommand := strings.Split(command, " ")

				cmd := exec.Command(arrCommand[0], arrCommand[1:]...)
				cmd.Dir = repository.Path
				cmd.Env = os.Environ()

				output, err := cmd.Output()

				if err != nil {
					slog.Debug(fmt.Sprintf("CONFIG Command err output %v", string(output)))
					return errors.New("failed to run command")
				}
			}
		} else {
			return err
		}
	}

	return nil
}
