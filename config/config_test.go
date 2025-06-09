package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func createTempFile(filePath string, fileContent Config) error {
	yamlBytes, err := yaml.Marshal(fileContent)

	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, yamlBytes, 0644)

	if err != nil {
		return err
	}

	return nil
}

func TestInitializeConfigSuccess(t *testing.T) {
	t.Cleanup(func() {
		Settings = Config{}
	})

	fileContent := Config{
		Preference: CronSettings{
			Cron: false,
			Spec: "*/30 * * * * *",
		},
		Repositories: map[string]RepositoryConfig{
			"stalker-bot": {
				Url:      "https://github.com/khouwdevin/gitomatically",
				Clone:    "git@github.com:khouwdevin/gitomatically.git",
				Branch:   "master",
				Path:     t.TempDir(),
				Commands: []string{},
			},
		},
	}
	filePath := filepath.Join(t.TempDir(), "config.yaml")

	err := createTempFile(filePath, fileContent)

	if err != nil {
		t.Error("Cannot write temporary config file")
	}

	err = InitializeConfig(filePath)

	assert.NoError(t, err, "InitializeConfig should not return an error")
}

func TestInitializeDefaultConfig(t *testing.T) {
	t.Cleanup(func() {
		Settings = Config{}
	})

	fileContent := Config{
		Repositories: map[string]RepositoryConfig{
			"stalker-bot": {
				Url:      "https://github.com/khouwdevin/gitomatically",
				Clone:    "git@github.com:khouwdevin/gitomatically.git",
				Branch:   "master",
				Path:     t.TempDir(),
				Commands: []string{},
			},
		},
	}
	filePath := filepath.Join(t.TempDir(), "config.yaml")

	err := createTempFile(filePath, fileContent)

	if err != nil {
		t.Error("Cannot write temporary config file")
	}

	err = InitializeConfig(filePath)

	assert.NoError(t, err, "InitializeConfig should not return an error")
}

func TestInitializeConfigError(t *testing.T) {
	t.Cleanup(func() {
		Settings = Config{}
	})

	fileContent := Config{
		Repositories: map[string]RepositoryConfig{},
	}
	filePath := filepath.Join(t.TempDir(), "config.yaml")

	err := createTempFile(filePath, fileContent)

	if err != nil {
		t.Error("Cannot write temporary config file")
	}

	err = InitializeConfig(filePath)

	assert.Error(t, err, "InitializeConfig should return an error")
	assert.Contains(t, err.Error(), "there is no repository in config.", "Error message should indicate no repository is provided.")
}

func TestInitializetConfigDurationMissing(t *testing.T) {
	t.Cleanup(func() {
		Settings = Config{}
	})

	fileContent := Config{
		Preference: CronSettings{
			Cron: true,
		},
		Repositories: map[string]RepositoryConfig{
			"stalker-bot": {
				Url:      "https://github.com/khouwdevin/gitomatically",
				Clone:    "git@github.com:khouwdevin/gitomatically.git",
				Branch:   "master",
				Path:     t.TempDir(),
				Commands: []string{},
			},
		},
	}
	filePath := filepath.Join(t.TempDir(), "config.yaml")

	err := createTempFile(filePath, fileContent)

	if err != nil {
		t.Error("Cannot write temporary config file")
	}

	err = InitializeConfig(filePath)

	assert.Error(t, err, "InitializeConfig should return an error")
	assert.Contains(t, err.Error(), "duration value is required.", "Error message should indicate spec variable is not exist.")
}
