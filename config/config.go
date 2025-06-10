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
		_, err := os.Stat(repository.Path)

		repoName := filepath.Base(repository.Path)

		if err == nil {
			_, err := os.Stat(filepath.Join(repository.Path, ".git"))

			if err == nil {
				continue
			} else if os.IsNotExist(err) {
				slog.Info(fmt.Sprintf("CONFIG Cloning %v", repository.Url))
				err = os.RemoveAll(repository.Path)

				if err != nil {
					return err
				}

				git := exec.Command("git", "clone", repository.Clone, repoName)
				git.Dir = filepath.Dir(repository.Path)
				git.Env = os.Environ()

				_, err := git.Output()

				if err != nil {
					return err
				}

				for _, command := range repository.Commands {
					slog.Debug(fmt.Sprintf("CONFIG Running %v", command))

					arrCommand := strings.Split(command, " ")

					cmd := exec.Command(arrCommand[0], arrCommand[1:]...)
					cmd.Dir = repository.Path
					cmd.Env = os.Environ()

					_, err := cmd.Output()

					if err != nil {
						return errors.New("failed to run command")
					}
				}
			} else {
				return err
			}
		} else if os.IsNotExist(err) {
			slog.Info(fmt.Sprintf("CONFIG Adding %v", repository.Url))

			dir := filepath.Dir(repository.Path)
			dirPerms := os.FileMode(0755)
			err := os.MkdirAll(dir, dirPerms)

			if err != nil {
				return err
			}

			git := exec.Command("git", "clone", repository.Clone, repoName)
			git.Dir = dir
			git.Env = os.Environ()

			_, err = git.Output()

			if err != nil {
				defer os.RemoveAll(repository.Path)
				return err
			}

			for _, command := range repository.Commands {
				slog.Debug(fmt.Sprintf("CONFIG Running %v", command))

				arrCommand := strings.Split(command, " ")

				cmd := exec.Command(arrCommand[0], arrCommand[1:]...)
				cmd.Dir = repository.Path
				cmd.Env = os.Environ()

				_, err := cmd.Output()

				if err != nil {
					return errors.New("failed to run command")
				}
			}
		} else {
			return err
		}
	}

	return nil
}
