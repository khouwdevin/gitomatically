package env

import (
	"errors"
	"os"
	"strconv"

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
	LOG_LEVEL := os.Getenv("LOG_LEVEL")
	PORT := os.Getenv("PORT")

	if GITHUB_WEBHOOK_SECRET == "" || GIN_MODE == "" || LOG_LEVEL == "" {
		return errors.New("Env variables are not complete")
	}
	if PORT == "" {
		PORT = "8080"
	}

	LOG_LEVEL_INT, err := strconv.Atoi(LOG_LEVEL)

	if err != nil {
		return err
	}

	Env = EnvType{GITHUB_WEBHOOK_SECRET: GITHUB_WEBHOOK_SECRET, GIN_MODE: GIN_MODE, LOG_LEVEL: LOG_LEVEL_INT, PORT: PORT}

	return nil
}
