package watcher

import "sync"

var (
	isChangingSetting bool
	ControllerGroup   sync.WaitGroup
	controllerMutex   sync.RWMutex
)

func UpdateSettingStatus(currentStatus bool) {
	controllerMutex.Lock()
	defer controllerMutex.Unlock()

	isChangingSetting = currentStatus
}

func GetSettingStatus() bool {
	controllerMutex.RLock()
	defer controllerMutex.RUnlock()

	return isChangingSetting
}
