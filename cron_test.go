package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func defaultConfig(dirPath string) Config {
	return Config{
		Preference: PreferenceSettings{
			Cron: false,
			Spec: "*/30 * * * * *",
		},
		Repositories: map[string]RepositoryConfig{
			"gitomatically": {
				Url:      "https://github.com/khouwdevin/gitomatically",
				Clone:    "git@github.com:khouwdevin/gitomatically.git",
				Branch:   "master",
				Path:     dirPath,
				Commands: []string{},
			},
		},
	}
}

func TestCronCreation(t *testing.T) {
	Settings = defaultConfig(t.TempDir())

	err := NewCron()

	defer func() {
		err = StopCron()

		assert.NoError(t, err, "Stop cron should not return an error")
	}()

	assert.NoError(t, err, "New cron should not return an error")
}

func TestChangeCron(t *testing.T) {
	Settings = defaultConfig(t.TempDir())

	err := NewCron()

	defer func() {
		err = StopCron()

		assert.NoError(t, err, "Stop cron should not return an error")
	}()

	err = ChangeCron()

	assert.NoError(t, err, "New cron should not return an error")
}
