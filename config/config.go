package config

import (
	"log"
	"os"
	"strconv"
)

var GlobalAppConfig = &AppConfig{}

func init() {
	c, err := NewAppConfigFromEnv()

	if err != nil {
		panic(err.Error())
	}

	GlobalAppConfig = &c
}

type AppConfig struct {
	Port          int
	MaxFileSize   int64
	AuthServerUrl string
}

func NewAppConfigFromEnv() (AppConfig, error) {
	port, err := strconv.Atoi(os.Getenv("SERVER_PORT"))
	maxFileSize, err := strconv.ParseInt(os.Getenv("MAX_FILE_SIZE"), 10, 64)
	authServerUrl := os.Getenv("AUTH_SERVER_URL")

	if err != nil {
		return AppConfig{}, nil
	}

	if !areEnvValid(authServerUrl) {
		return AppConfig{}, nil
	}

	config := AppConfig{}

	config.Port = port
	config.MaxFileSize = int64(maxFileSize)
	config.AuthServerUrl = authServerUrl

	log.Printf("%v", config)

	return config, err
}

func areEnvValid(envs ...string) bool {
	for _, env := range envs {
		if env == "" {
			return false
		}
	}
	return true
}
