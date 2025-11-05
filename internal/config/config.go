package config

import (
	"encoding/json"
	"os"
)

// dsn 结构体
type DatabaseConfig struct {
	DSN string `json:"dsn"`
}

// JWT 结构体
type JWTConfig struct {
	Secret string `json:"secret"`
}

// database 结构体
type Config struct {
	Database DatabaseConfig `json:"database"`
	JWT      JWTConfig      `json:"jwt"`
}

// 全局指针 Cfg，用于存储最终加载的配置
var Cfg *Config

// 加载数据库配置
func LoadConfig() error {
	// 扩展性保留，实际并没有 APP_ENV 配置文件
	env := os.Getenv("APP_ENV")
	// 选择加载什么配置文件
	if env == "" {
		env = "azure"
	}

	// 加载config.json文件
	file, err := os.ReadFile("config/config.json")
	if err != nil {
		return err
	}

	// 解析 JSON 文件，映射到 allConfigs 结构体
	var allConfigs map[string]Config
	if err := json.Unmarshal(file, &allConfigs); err != nil {
		return err
	}

	// 获取指定环境的配置
	envConfig, ok := allConfigs[env]
	if !ok {
		return &ConfigError{Env: env}
	}

	Cfg = &envConfig
	return nil
}

// 自定义错误类型
type ConfigError struct {
	Env string
}

func (e *ConfigError) Error() string {
	return "configuration for environment '" + e.Env + "' not found"
}
