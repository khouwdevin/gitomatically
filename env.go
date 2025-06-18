package main

import (
	"errors"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func InitializeEnv(filePath string) error {
	if os.Getenv("GITHUB_WEBHOOK_SECRET") != "" {
		err := godotenv.Overload(filePath)

		if err != nil {
			return err
		}
	} else {
		err := godotenv.Load(filePath)

		if err != nil {
			return err
		}
	}

	GITHUB_WEBHOOK_SECRET := os.Getenv("GITHUB_WEBHOOK_SECRET")
	GIN_MODE := os.Getenv("GIN_MODE")
	LOG_LEVEL := os.Getenv("LOG_LEVEL")
	PORT := os.Getenv("PORT")

	if LOG_LEVEL == "" {
		return errors.New("LOG_LEVEL env variable is required")
	}

	if PORT == "" {
		os.Setenv("PORT", "8080")
	}
	if GIN_MODE == "" {
		os.Setenv("GIN_MODE", "release")
	}
	if GITHUB_WEBHOOK_SECRET == "" {
		os.Setenv("GITHUB_WEBHOOK_SECRET", "empty")
	}

	_, err := strconv.Atoi(LOG_LEVEL)

	if err != nil {
		return err
	}

	return nil
}
