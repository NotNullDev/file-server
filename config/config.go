package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
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

func ParseEnvFiles(ignoreOnNonExistingEnvFile bool, envFilePath ...string) (map[string]string, error) {
	var result map[string]string = make(map[string]string)

	for _, filePath := range envFilePath {
		file, err := os.Open(filePath)

		if err != nil {
			if ignoreOnNonExistingEnvFile {
				continue
			}
			return nil, err
		}

		scanner := bufio.NewScanner(file)

		var lines []string

		for {
			lines = append(lines, scanner.Text())
			if !scanner.Scan() {
				break
			}
		}

		for lineNumber, line := range lines {

			commentRegex := regexp.MustCompile("#.+$")

			comment := commentRegex.Find([]byte(line))

			line = strings.Replace(line, string(comment), "", -1)
			line = strings.TrimSpace(line)

			if line == "" {
				continue
			}

			splitted := strings.Split(line, "=")

			if len(splitted) != 2 {
				return nil, fmt.Errorf("could not parse line %d in file [%s]", lineNumber, filePath)
			}

			envKey := splitted[0]
			envKey = strings.Replace(envKey, "export", "", -1)
			envKey = strings.TrimSpace(envKey)

			envVal := splitted[1]
			envVal = strings.Replace(envVal, "\"", "", -1)
			envVal = strings.Replace(envVal, "'", "", -1)
			envVal = strings.Replace(envVal, "`", "", -1)
			envVal = strings.TrimSpace(envVal)

			if envKey == "" {
				return nil, fmt.Errorf("key at line %d is empty in file [%s]", lineNumber, filePath)
			}

			if envVal == "" {
				return nil, fmt.Errorf("value at line %d is empty in file [%s]", lineNumber, filePath)
			}

			result[envKey] = envVal
		}
	}

	return result, nil
}
