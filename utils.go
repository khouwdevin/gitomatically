package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/khouwdevin/gitomatically/watcher"
)

func GitClone(repository RepositoryConfig) error {
	err := os.RemoveAll(repository.Path)

	if err != nil {
		return err
	}

	dir := filepath.Dir(repository.Path)
	dirPerms := os.FileMode(0755)
	err = os.MkdirAll(dir, dirPerms)

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

	return err
}

func GitPull(repository RepositoryConfig) error {
	publicKeys, err := ssh.NewPublicKeysFromFile("git", Settings.Preference.PrivateKey, Settings.Preference.Paraphrase)

	if err != nil {
		return err
	}

	r, err := git.PlainOpen(repository.Path)

	w, err := r.Worktree()

	if err != nil {
		return err
	}

	err = w.Pull(&git.PullOptions{RemoteName: "origin", ReferenceName: plumbing.ReferenceName(repository.Branch), Force: false, Auth: publicKeys, Progress: os.Stdout})

	return err
}

func EnvDebouncedEvents(w *watcher.Watcher) {
	if w.Self.Timer != nil {
		w.Self.Timer.Stop()
	}

	w.Self.Timer = time.AfterFunc(100*time.Millisecond, func() {
		watcher.UpdateSettingStatus(true)
		watcher.ControllerGroup.Wait()

		slog.Info("WATCHER Env file change detected, reinitialize env")

		prevPort := os.Getenv("PORT")
		err := InitializeEnv(w.Self.Path)
		currentPort := os.Getenv("PORT")

		if err != nil {
			slog.Error(fmt.Sprintf("WATCHER Reinitialize env error %v", err))
			w.Quit <- syscall.SIGTERM

			return
		}

		if !Settings.Preference.Cron && prevPort != currentPort {
			err = NewServer()

			if err != nil {
				slog.Error(fmt.Sprintf("WATCHER Restart server error %v", err))
				w.Quit <- syscall.SIGTERM

				return
			}
		}

		LOG_LEVEL_INT, err := strconv.Atoi(os.Getenv("LOG_LEVEL"))

		if err != nil {
			slog.Error(fmt.Sprintf("MAIN Error convert string to int %v", err))
			return
		}

		slog.SetLogLoggerLevel(slog.Level(LOG_LEVEL_INT))

		watcher.UpdateSettingStatus(false)
	})
}

func ConfigDebouncedEvents(w *watcher.Watcher) {
	if w.Self.Timer != nil {
		w.Self.Timer.Stop()
	}

	w.Self.Timer = time.AfterFunc(100*time.Millisecond, func() {
		watcher.UpdateSettingStatus(true)
		watcher.ControllerGroup.Wait()

		slog.Info("WATCHER Config file change detected, reinitialize config")

		prevConfig := Settings
		err := InitializeConfig(w.Self.Path)

		if err != nil {
			slog.Error(fmt.Sprintf("WATCHER Reinitialize config error %v", err))
			w.Quit <- syscall.SIGTERM

			return
		}

		err = PreStart()

		if err != nil {
			slog.Error(fmt.Sprintf("WATCHER Rerun prestart error %v", err))
			w.Quit <- syscall.SIGTERM

			return
		}

		if prevConfig.Preference.Cron == Settings.Preference.Cron &&
			(!Settings.Preference.Cron || prevConfig.Preference.Spec == Settings.Preference.Spec) {
			return
		}

		if Settings.Preference.Cron {
			err := ShutdownServer()

			if err != nil {
				slog.Error(fmt.Sprintf("WATCHER Shutdown server error %v", err))
				w.Quit <- syscall.SIGTERM

				return
			}

			ChangeCron()
		} else {
			StopCron()

			err := NewServer()

			if err != nil {
				slog.Error(fmt.Sprintf("WATCHER Start server error %v", err))
				w.Quit <- syscall.SIGTERM

				return
			}
		}

		watcher.UpdateSettingStatus(false)
	})
}
