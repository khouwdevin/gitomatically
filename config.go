package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

type PreferenceSettings struct {
	PrivateKey string `yaml:"private_key"`
	Paraphrase string `yaml:"paraphrase"`
	Cron       bool   `yaml:"cron"`
	Spec       string `yaml:"spec"`
}

type RepositoryConfig struct {
	Url      string   `yaml:"url"`
	Clone    string   `yaml:"clone"`
	Branch   string   `yaml:"branch"`
	Path     string   `yaml:"path"`
	Commands []string `yaml:"commands"`
}

type Config struct {
	Preference   PreferenceSettings          `yaml:"preference"`
	Repositories map[string]RepositoryConfig `yaml:"repositories"`
}

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

	if Settings.Preference.PrivateKey == "" {
		return errors.New("private key is not found.")
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

		if err == nil {
			slog.Debug(fmt.Sprintf("CONFIG Pulling %v", repository.Url))

			publicKeys, err := ssh.NewPublicKeysFromFile("git", Settings.Preference.PrivateKey, Settings.Preference.Paraphrase)

			if err != nil {
				return err
			}

			r, err := git.PlainOpen(repository.Path)

			w, err := r.Worktree()

			if err != nil {
				return err
			}

			err = w.Pull(&git.PullOptions{RemoteName: "origin", Auth: publicKeys, Progress: os.Stdout})

			if err != nil {
				if err == git.NoErrAlreadyUpToDate {
					slog.Debug(fmt.Sprintf("CONFIG %v is up to date, continue to next repository", repository.Url))
					continue
				} else {
					slog.Debug(fmt.Sprintf("CONFIG Git clone err output %v", err))
					return err
				}
			}

			for _, command := range repository.Commands {
				slog.Debug(fmt.Sprintf("CONFIG Running %v", command))

				arrCommand := strings.Split(command, " ")

				cmd := exec.Command(arrCommand[0], arrCommand[1:]...)
				cmd.Dir = repository.Path
				cmd.Env = os.Environ()

				_, err := cmd.Output()

				if err != nil {
					slog.Error("CONFIG Failed to run command")
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

			publicKeys, err := ssh.NewPublicKeysFromFile("git", Settings.Preference.PrivateKey, Settings.Preference.Paraphrase)

			if err != nil {
				return err
			}

			_, err = git.PlainClone(repository.Path, false, &git.CloneOptions{
				Auth: publicKeys,
				URL:  repository.Clone,
			})

			if err != nil {
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
