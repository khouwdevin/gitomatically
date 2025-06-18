package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func createTempYAMLFile(filePath string, fileContent Config) error {
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

func createTempSSH(filePath string) (string, error) {
	tempSSHPath := filepath.Join(filePath, "ssh_key")

	err := os.WriteFile(tempSSHPath, []byte("dummy"), 0600)
	if err != nil {
		return "", err
	}

	return tempSSHPath, nil
}

func TestInitializeConfigSuccess(t *testing.T) {
	t.Cleanup(func() {
		Settings = Config{}
	})

	sshPath, err := createTempSSH(t.TempDir())

	if err != nil {
		t.Error("Error creating temp ssh")
	}

	fileContent := Config{
		Preference: PreferenceSettings{
			PrivateKey: sshPath,
			Cron:       false,
			Spec:       "*/30 * * * * *",
		},
		Repositories: map[string]RepositoryConfig{
			"gitomatically": {
				Url:      "https://github.com/khouwdevin/gitomatically",
				Clone:    "git@github.com:khouwdevin/gitomatically.git",
				Branch:   "master",
				Path:     filepath.Join(t.TempDir(), "gitomatically"),
				Commands: []string{},
			},
		},
	}
	filePath := filepath.Join(t.TempDir(), "config.yaml")

	err = createTempYAMLFile(filePath, fileContent)

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

	sshPath, err := createTempSSH(t.TempDir())

	if err != nil {
		t.Error("Error creating temp ssh")
	}

	fileContent := Config{
		Preference: PreferenceSettings{
			PrivateKey: sshPath,
		},
		Repositories: map[string]RepositoryConfig{
			"stalker-bot": {
				Url:      "https://github.com/khouwdevin/gitomatically",
				Clone:    "git@github.com:khouwdevin/gitomatically.git",
				Branch:   "master",
				Path:     filepath.Join(t.TempDir(), "gitomatically"),
				Commands: []string{},
			},
		},
	}

	filePath := filepath.Join(t.TempDir(), "config.yaml")

	err = createTempYAMLFile(filePath, fileContent)

	if err != nil {
		t.Error("Cannot write temporary config file")
	}

	err = InitializeConfig(filePath)

	assert.NoError(t, err, "InitializeConfig should not return an error")
}

func TestInitializeConfigRepositoryError(t *testing.T) {
	t.Cleanup(func() {
		Settings = Config{}
	})

	sshPath, err := createTempSSH(t.TempDir())

	if err != nil {
		t.Error("Error creating temp ssh")
	}

	fileContent := Config{
		Preference: PreferenceSettings{
			PrivateKey: sshPath,
		},
		Repositories: map[string]RepositoryConfig{},
	}

	filePath := filepath.Join(t.TempDir(), "config.yaml")

	err = createTempYAMLFile(filePath, fileContent)

	if err != nil {
		t.Error("Cannot write temporary config file")
	}

	err = InitializeConfig(filePath)

	assert.Error(t, err, "InitializeConfig should return an error")
	assert.Contains(t, err.Error(), "there is no repository in config.", "Error message should indicate private_key variable is not exist.")
}

func TestInitializetConfigPrivateKeyMissing(t *testing.T) {
	t.Cleanup(func() {
		Settings = Config{}
	})

	fileContent := Config{
		Repositories: map[string]RepositoryConfig{
			"stalker-bot": {
				Url:      "https://github.com/khouwdevin/gitomatically",
				Clone:    "git@github.com:khouwdevin/gitomatically.git",
				Branch:   "master",
				Path:     filepath.Join(t.TempDir(), "gitomatically"),
				Commands: []string{},
			},
		},
	}

	filePath := filepath.Join(t.TempDir(), "config.yaml")

	err := createTempYAMLFile(filePath, fileContent)

	if err != nil {
		t.Error("Cannot write temporary config file")
	}

	err = InitializeConfig(filePath)

	assert.Error(t, err, "InitializeConfig should return an error")
	assert.Contains(t, err.Error(), "private key is not found.", "Error message should indicate private_key variable is not exist.")
}

func TestInitializetConfigDurationMissing(t *testing.T) {
	t.Cleanup(func() {
		Settings = Config{}
	})

	sshPath, err := createTempSSH(t.TempDir())

	if err != nil {
		t.Error("Error creating temp ssh")
	}

	fileContent := Config{
		Preference: PreferenceSettings{
			PrivateKey: sshPath,
			Cron:       true,
		},
		Repositories: map[string]RepositoryConfig{
			"stalker-bot": {
				Url:      "https://github.com/khouwdevin/gitomatically",
				Clone:    "git@github.com:khouwdevin/gitomatically.git",
				Branch:   "master",
				Path:     filepath.Join(t.TempDir(), "gitomatically"),
				Commands: []string{},
			},
		},
	}
	filePath := filepath.Join(t.TempDir(), "config.yaml")

	err = createTempYAMLFile(filePath, fileContent)

	if err != nil {
		t.Error("Cannot write temporary config file")
	}

	err = InitializeConfig(filePath)

	assert.Error(t, err, "InitializeConfig should return an error")
	assert.Contains(t, err.Error(), "duration value is required.", "Error message should indicate spec variable is not exist.")
}

func TestPreStart(t *testing.T) {
	t.Skip("To make this test work, you need to provide an SSH key that is registered with GitHub.")
	t.Cleanup(func() {
		Settings = Config{}
	})

	Settings = Config{
		Preference: PreferenceSettings{
			PrivateKey: "~/.ssh/id_ed25519",
		},
		Repositories: map[string]RepositoryConfig{
			"stalker-bot": {
				Url:      "https://github.com/khouwdevin/gitomatically",
				Clone:    "git@github.com:khouwdevin/gitomatically.git",
				Branch:   "master",
				Path:     filepath.Join(t.TempDir(), "gitomatically"),
				Commands: []string{},
			},
		},
	}

	err := PreStart()

	assert.NoError(t, err, "Prestart should not return an error")
}
