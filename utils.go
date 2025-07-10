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

func IsNewUpdate(r *git.Repository, o *git.PullOptions) (bool, error) {
	err := r.Fetch(&git.FetchOptions{
		RemoteName: o.RemoteName,
		Auth:       o.Auth,
		Force:      o.Force,
	})

	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			return false, err
		} else {
			return false, err
		}
	}

	headRef, err := r.Head()

	if err != nil {
		slog.Debug("ISNEWUPDATE Error get head")
		return false, err
	}

	remoteRef, err := r.Reference(o.ReferenceName, true)

	if err != nil {
		slog.Debug(fmt.Sprintf("ISNEWUPDATE Error get reference name %v", o.ReferenceName))
		return false, err
	}

	if headRef.Hash() != remoteRef.Hash() {
		return true, nil
	}

	return false, nil
}

func BackupUntrackedFiles(w *git.Worktree, repositoryPath string) (string, error) {
	gitStatus, err := w.Status()

	if err != nil {
		slog.Debug("BACKUPUNTRACKEDFILES Error get worktree status")
		return "", err
	}

	if len(gitStatus) == 0 {
		return "", nil
	}

	tempDirPath, err := os.MkdirTemp("", "git-untracked-backup-")

	for file, status := range gitStatus {
		if status.Staging == git.Untracked {
			data, err := os.ReadFile(fmt.Sprintf("%v/%v", repositoryPath, file))

			if err != nil {
				slog.Debug(fmt.Sprintf("BACKUPUNTRACKEDFILES Error read file %v/%v", repositoryPath, file))
				return "", err
			}

			err = os.WriteFile(fmt.Sprintf("%v/%v", tempDirPath, file), data, 0755)

			if err != nil {
				slog.Debug(fmt.Sprintf("BACKUPUNTRACKEDFILES Error write file %v/%v", tempDirPath, file))
				return "", err
			}
		}
	}

	return tempDirPath, nil
}

func ReturnUntrackedFiles(tempDirPath string, repositoryPath string) error {
	files, err := os.ReadDir(tempDirPath)

	if err != nil {
		slog.Debug(fmt.Sprintf("RETURNUNTRACKEDFILES Error read dir %v", tempDirPath))
		return err
	}

	for _, file := range files {
		data, err := os.ReadFile(fmt.Sprintf("%v/%v", tempDirPath, file.Name()))

		if err != nil {
			slog.Debug(fmt.Sprintf("RETURNUNTRACKEDFILES Error read file %v/%v", tempDirPath, file.Name()))
			return err
		}

		err = os.WriteFile(fmt.Sprintf("%v/%v", repositoryPath, file.Name()), data, 0755)

		if err != nil {
			slog.Debug(fmt.Sprintf("RETURNUNTRACKEDFILES Error write file %v/%v", tempDirPath, file.Name()))
			return err
		}
	}

	err = os.RemoveAll(tempDirPath)

	if err != nil {
		slog.Debug(fmt.Sprintf("RETURNUNTRACKEDFILES Error remove all %v", tempDirPath))
		return err
	}

	return nil
}

func GitClone(repository RepositoryConfig) error {
	slog.Debug(fmt.Sprintf("GITCLONE Clone %v start", repository.Url))
	err := os.RemoveAll(repository.Path)

	if err != nil {
		slog.Debug(fmt.Sprintf("GITCLONE Error remove all files from path %v", repository.Path))
		return err
	}

	dir := filepath.Dir(repository.Path)
	dirPerms := os.FileMode(0755)
	err = os.MkdirAll(dir, dirPerms)

	if err != nil {
		slog.Debug(fmt.Sprintf("GITCLONE Error create path %v", repository.Path))
		return err
	}

	publicKeys, err := ssh.NewPublicKeysFromFile("git", Settings.Preference.PrivateKey, Settings.Preference.Paraphrase)

	if err != nil {
		slog.Debug(fmt.Sprintf("GITCLONE Error get public keys from file %v", Settings.Preference.PrivateKey))
		return err
	}

	_, err = git.PlainClone(repository.Path, false, &git.CloneOptions{
		Auth: publicKeys,
		URL:  repository.Clone,
	})

	slog.Debug(fmt.Sprintf("GITCLONE Error do plain clone %v", repository.Url))

	return err
}

func GitPull(repository RepositoryConfig) error {
	slog.Debug(fmt.Sprintf("GITPULL Pull %v start", repository.Url))
	publicKeys, err := ssh.NewPublicKeysFromFile("git", Settings.Preference.PrivateKey, Settings.Preference.Paraphrase)

	if err != nil {
		slog.Debug("GITPULL Error get public keys from file")
		return err
	}

	r, err := git.PlainOpen(repository.Path)

	if err != nil {
		slog.Debug(fmt.Sprintf("GITPULL Error do plain open %v", repository.Path))
		return err
	}

	w, err := r.Worktree()

	if err != nil {
		slog.Debug("GITPULL Error get worktree")
		return err
	}

	pullOption := &git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: plumbing.NewBranchReferenceName(repository.Branch),
		Auth:          publicKeys,
		Force:         false,
	}

	isNewUpdate, err := IsNewUpdate(r, pullOption)

	if err != nil {
		return err
	}

	if isNewUpdate {
		tempDirPath, err := BackupUntrackedFiles(w, repository.Path)

		if err != nil {
			return err
		}

		err = w.Pull(pullOption)

		if err != nil {
			slog.Debug(fmt.Sprintf("GITPULL Error pull repository %v", repository.Url))
			return err
		}

		if len(tempDirPath) > 0 {
			err = ReturnUntrackedFiles(tempDirPath, repository.Path)

			if err != nil {
				return err
			}

			return nil
		}
	}

	return nil
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
		Settings = Config{}
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

		if prevConfig.Preference.Cron == Settings.Preference.Cron ||
			prevConfig.Preference.Spec == Settings.Preference.Spec {
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
