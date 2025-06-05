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

func InitializeConfig() error {
	yamlFile, err := os.ReadFile("config.yaml")

	if err != nil {
		return err
	}

	err = yaml.Unmarshal(yamlFile, &Settings)

	if err != nil {
		return err
	}

	err = PreStart()

	if err != nil {
		return err
	}

	return nil
}

func PreStart() error {
	for _, repository := range Settings.Repositories {
		_, err := os.Stat(repository.Path)

		repoName := filepath.Base(repository.Path)

		if err == nil {
			_, err := os.Stat(repository.Path + ".git")

			if err == nil {
				continue
			} else if os.IsNotExist(err) {
				slog.Info(fmt.Sprintf("Cloning %v", repository.Url))
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

				for _, buildCommands := range repository.BuildCommands {
					build := strings.Split(buildCommands, " ")

					cmd := exec.Command(build[0], build[1:]...)
					cmd.Dir = repository.Path
					cmd.Env = os.Environ()

					_, err := cmd.Output()

					if err != nil {
						return errors.New("Failed to run build command")
					}
				}
			} else {
				return err
			}
		} else if os.IsNotExist(err) {
			slog.Info(fmt.Sprintf("Adding %v", repository.Url))

			dirPerms := os.FileMode(0755)
			err := os.MkdirAll(repository.Path, dirPerms)

			if err != nil {
				return err
			}

			git := exec.Command("git", "clone", repository.Clone, repoName)
			git.Dir = filepath.Dir(repository.Path)
			git.Env = os.Environ()

			_, err = git.Output()

			if err != nil {
				defer os.RemoveAll(repository.Path)
				return err
			}

			for _, buildCommands := range repository.BuildCommands {
				build := strings.Split(buildCommands, " ")

				cmd := exec.Command(build[0], build[1:]...)
				cmd.Dir = repository.Path
				cmd.Env = os.Environ()

				_, err := cmd.Output()

				if err != nil {
					return errors.New("Failed to run build command")
				}
			}
		} else {
			return err
		}
	}

	return nil
}
