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
	if Env.GITHUB_WEBHOOK_SECRET != "" {
		err := godotenv.Overload(".env")

		if err != nil {
			return err
		}
	}

	err := godotenv.Load()

	if err != nil {
		return err
	}

	GITHUB_WEBHOOK_SECRET := os.Getenv("GITHUB_WEBHOOK_SECRET")
	GIN_MODE := os.Getenv("GIN_MODE")
	LOG_LEVEL := os.Getenv("LOG_LEVEL")
	PORT := os.Getenv("PORT")

	if GIN_MODE == "" {
		return errors.New("GIN_MODE env variable is required")
	} else if LOG_LEVEL == "" {
		return errors.New("LOG_LEVEL env variable is required")
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
