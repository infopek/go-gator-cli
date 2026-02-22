package config

import (
	"encoding/json"
	"fmt"
	"os"
	"io"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbURL 			string	`json:"db_url"`
	CurrentUserName string	`json:"current_user_name"`
}

func Read() (Config, error) {
	jsonFilePath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	jsonFile, _ := os.Open(jsonFilePath)
	defer jsonFile.Close()

	bytes, _ := io.ReadAll(jsonFile)
	var config Config
	json.Unmarshal(bytes, &config)

	return config, nil
}

func (config Config) SetUser(username string) error {
	config.CurrentUserName = username
	if err := write(config); err != nil {
		return err
	}
	return nil
}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	filePath := filepath.Join(homeDir, configFileName)
	if _, err := os.Stat(filePath); err != nil {
		return "", err
	}
	return filePath, nil
}

func write(config Config) error {
 	bytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	filePath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	configFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer configFile.Close()

	n, err := configFile.Write(bytes)
	if err != nil {
		return err
	}
	fmt.Printf("Written %d bytes to %s\n", n, filePath)

	return nil
}
