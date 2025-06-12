package watcher

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

var (
	watcherMutex sync.RWMutex
)

func NewWatcher(filePath string, wg *sync.WaitGroup, quit chan os.Signal) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		watcher.Close()
		return nil, errors.New(fmt.Sprintf("Watcher error %v", err))
	}

	_, err = os.Stat(filePath)

	if os.IsNotExist(err) {
		watcher.Close()
		return nil, err
	}

	absConfigPath, err := filepath.Abs(filePath)

	if err != nil {
		watcher.Close()
		return nil, errors.New(fmt.Sprintf("Get absolute path error %v", err))
	}

	err = watcher.Add(absConfigPath)

	if err != nil {
		watcher.Close()
		return nil, errors.New(fmt.Sprintf("WATCHER Add file to watcher error %v", err))
	}

	w := &Watcher{
		watcher: watcher,
		quit:    quit,
		wg:      wg,
		Self: Self{
			Path:     absConfigPath,
			Timer:    time.AfterFunc(100*time.Millisecond, func() {}),
			StopChan: make(chan struct{}),
		},
	}

	return w, nil
}

func (w *Watcher) Stop() error {
	close(w.Self.StopChan)
	return w.watcher.Close()
}

func (w *Watcher) Run(callback func(w *Watcher)) {
	w.wg.Add(1)

	go func() {
		for {
			select {
			case event, ok := <-w.watcher.Events:
				if !ok {
					slog.Error("WATCHER Watcher events channel is closed")
					w.quit <- syscall.SIGTERM

					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0 {
					go callback(w)
				} else if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
					slog.Error("WATCHER File is removed or renamed")
					w.quit <- syscall.SIGTERM

					return
				}

			case err, ok := <-w.watcher.Errors:
				if !ok {
					slog.Error("WATCHER Watcher errors channel is closed")
					w.quit <- syscall.SIGTERM

					return
				}
				if err != nil {
					slog.Error(fmt.Sprintf("WATCHER Watcher error %v", err))
					w.quit <- syscall.SIGTERM

					return
				}

			case <-w.Self.StopChan:
				w.wg.Done()

				return
			}
		}
	}()
}
