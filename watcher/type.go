package watcher

import (
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Self struct {
	Path     string
	Timer    *time.Timer
	StopChan chan struct{}
}

type Watcher struct {
	watcher *fsnotify.Watcher
	quit    chan os.Signal
	wg      *sync.WaitGroup
	Self    Self
}
