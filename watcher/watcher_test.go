package watcher

import (
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var defaultContent = "PORT=3000"

func createTempFile(filePath string, fileContent string) error {
	err := os.WriteFile(filePath, []byte(fileContent), 0644)

	if err != nil {
		return err
	}

	return nil
}

func TestWatcherCreation(t *testing.T) {
	var wg sync.WaitGroup
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	filePath := filepath.Join(t.TempDir(), ".env")

	err := createTempFile(filePath, defaultContent)

	if err != nil {
		t.Errorf("Error creating env file %v", err)
	}

	envWatcher, err := NewWatcher(filePath, &wg, quit)

	defer func() {
		err := envWatcher.Stop()

		assert.NoError(t, err, "Stop watcher should not return an error")
	}()

	assert.NoError(t, err, "New watcher should not return an error")
}

func TestWatcherCreationError(t *testing.T) {
	var wg sync.WaitGroup
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	filePath := filepath.Join(t.TempDir(), ".env")

	err := createTempFile(filePath, defaultContent)

	if err != nil {
		t.Errorf("Error creating env file %v", err)
	}

	_, err = NewWatcher(filepath.Join(t.TempDir(), "local.env"), &wg, quit)

	assert.Error(t, err, "New watcher should return an error for not exist file")
}

func TestNewWatcher(t *testing.T) {
	var wg sync.WaitGroup
	var fileString string

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	filePath := filepath.Join(t.TempDir(), ".env")

	err := createTempFile(filePath, defaultContent)

	if err != nil {
		t.Errorf("Error creating env file %v", err)
	}

	envWatcher, err := NewWatcher(filePath, &wg, quit)

	if err != nil {
		t.Errorf("Creating watcher error %v", err)
	}

	envWatcher.Run(func(w *Watcher) {
		if w.Self.Timer != nil {
			w.Self.Timer.Stop()
		}

		w.Self.Timer = time.AfterFunc(100*time.Millisecond, func() {
			fileByte, err := os.ReadFile(w.Self.Path)

			if err != nil {
				t.Errorf("Error reading env file %v", err)
			}

			fileString = string(fileByte)

			w.Stop()
		})
	})

	fileContent := "PORT=5000"

	err = createTempFile(filePath, fileContent)

	if err != nil {
		t.Errorf("Error overwriting env file %v", err)
	}

	<-envWatcher.Self.StopChan

	assert.NotEqual(t, defaultContent, fileString, "File value should be change")
}
