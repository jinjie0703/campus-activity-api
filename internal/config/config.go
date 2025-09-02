package config

import (
	"encoding/json"
	"os"
)

type DatabaseConfig struct {
	DSN string `json:"dsn"`
}

type Config struct {
	Database DatabaseConfig `json:"database"`
}

var Cfg *Config

func LoadConfig() error {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "azure"
	}

	file, err := os.ReadFile("config/config.json")
	if err != nil {
		return err
	}

	var allConfigs map[string]Config
	if err := json.Unmarshal(file, &allConfigs); err != nil {
		return err
	}

	envConfig, ok := allConfigs[env]
	if !ok {
		return &ConfigError{Env: env}
	}

	Cfg = &envConfig
	return nil
}

type ConfigError struct {
	Env string
}

func (e *ConfigError) Error() string {
	return "configuration for environment '" + e.Env + "' not found"
}
