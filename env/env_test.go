package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createTempFile(filePath string, fileContent string) error {
	err := os.WriteFile(filePath, []byte(fileContent), 0644)

	if err != nil {
		return err
	}

	return nil
}

func TestInitializeEnvSuccess(t *testing.T) {
	t.Cleanup(func() {
		os.Unsetenv("GITHUB_WEBHOOK_SECRET")
		os.Unsetenv("GIN_MODE")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("PORT")
	})

	fileContent := `
GITHUB_WEBHOOK_SECRET=supersecret
GIN_MODE=release
LOG_LEVEL=0
PORT=3000
	`
	filePath := filepath.Join(t.TempDir(), ".env")

	err := createTempFile(filePath, fileContent)

	if err != nil {
		t.Error("Cannot write temporary env file")
	}

	err = InitializeEnv(filePath)

	assert.NoError(t, err, "InitializeEnv should not return an error")
	assert.Equal(t, "supersecret", Env.GITHUB_WEBHOOK_SECRET, "GITHUB_WEBHOOK_SECRET should be set")
	assert.Equal(t, "release", Env.GIN_MODE, "GIN_MODE should be set")
	assert.Equal(t, 0, Env.LOG_LEVEL, "LOG_LEVEL should be parsed correctly")
	assert.Equal(t, "3000", Env.PORT, "PORT should be set correctly")
}

func TestInitializeEnvDefaultPort(t *testing.T) {
	t.Cleanup(func() {
		os.Unsetenv("GITHUB_WEBHOOK_SECRET")
		os.Unsetenv("GIN_MODE")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("PORT")
	})

	fileContent := "LOG_LEVEL=0"
	filePath := filepath.Join(t.TempDir(), ".env")

	err := createTempFile(filePath, fileContent)

	if err != nil {
		t.Error("Cannot write temporary env file")
	}

	err = InitializeEnv(filePath)

	assert.NoError(t, err, "InitializeEnv should not return an error with default port")
	assert.Equal(t, "8080", Env.PORT, "PORT should default to 8080")
}

func TestInitializeEnvMissingRequiredVars(t *testing.T) {
	t.Cleanup(func() {
		os.Unsetenv("GITHUB_WEBHOOK_SECRET")
		os.Unsetenv("GIN_MODE")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("PORT")
	})

	fileContent := ""
	filePath := filepath.Join(t.TempDir(), ".env")

	err := createTempFile(filePath, fileContent)

	if err != nil {
		t.Error("Cannot write temporary env file")
	}

	err = InitializeEnv(filePath)

	assert.Error(t, err, "Expected an error for missing LOG_LEVEL")
	assert.Contains(t, err.Error(), "LOG_LEVEL env variable is required", "Error message should indicate LOG_LEVEL variable is not exist.")
}

func TestInitializeEnvInvalidLogLevel(t *testing.T) {
	t.Cleanup(func() {
		os.Unsetenv("GITHUB_WEBHOOK_SECRET")
		os.Unsetenv("GIN_MODE")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("PORT")
	})

	fileContent := "LOG_LEVEL='not-a-number'"
	filePath := filepath.Join(t.TempDir(), ".env")

	err := createTempFile(filePath, fileContent)

	if err != nil {
		t.Error("Cannot write temporary env file")
	}

	err = InitializeEnv(filePath)

	assert.Error(t, err, "Expected an error for invalid LOG_LEVEL")
	assert.Contains(t, err.Error(), "strconv.Atoi", "Error message should indicate conversion failure")
}

func TestInitializeEnvGodotenvOverload(t *testing.T) {
	t.Cleanup(func() {
		os.Unsetenv("GITHUB_WEBHOOK_SECRET")
		os.Unsetenv("GIN_MODE")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("PORT")
	})

	fileContent := "LOG_LEVEL=0"
	filePath := filepath.Join(t.TempDir(), ".env")

	err := createTempFile(filePath, fileContent)

	if err != nil {
		t.Error("Cannot write temporary env file")
	}

	err = InitializeEnv(filePath)

	fileContent = "LOG_LEVEL=-4"

	err = createTempFile(filePath, fileContent)

	if err != nil {
		t.Error("Cannot write temporary env file")
	}

	err = InitializeEnv(filePath)

	assert.NoError(t, err)
	assert.Equal(t, -4, Env.LOG_LEVEL, "LOG_LEVEL should be overloaded")
}
