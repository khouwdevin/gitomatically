package env

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

var (
	Env EnvType
)

func InitializeEnv() error {
	err := godotenv.Load()

	if err != nil {
		return err
	}

	GITHUB_WEBHOOK_SECRET := os.Getenv("GITHUB_WEBHOOK_SECRET")
	GIN_MODE := os.Getenv("GIN_MODE")

	if GITHUB_WEBHOOK_SECRET == "" || GIN_MODE == "" {
		return errors.New("Env variables are not complete")
	}

	Env = EnvType{GITHUB_WEBHOOK_SECRET: GITHUB_WEBHOOK_SECRET, GIN_MODE: GIN_MODE}

	return nil
}
