package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 与 config.yaml 结构对应
type Config struct {
	Config struct {
		Server   ServerConfig   `yaml:"server"`
		Database DatabaseConfig `yaml:"database"`
		Platform PlatformConfig `yaml:"platform"`
		User     UserConfig     `yaml:"user_config"`
	} `yaml:"config"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DatabaseConfig struct {
	Type     string `yaml:"type"` // mysql（后续可扩展其他数据库）
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

type PlatformConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type UserConfig struct {
	DefaultMoney       float64 `yaml:"default_money"`
	DefaultWarehouse   int64   `yaml:"default_warehouse"`
	DefaultColdStorage int64   `yaml:"default_cold_storage"`
}

// Load 读取并解析配置文件，填充默认值
func Load(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("无法读取配置文件 %s: %w", path, err)
	}
	var c Config
	if err := yaml.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 默认值
	if c.Config.Server.Host == "" {
		c.Config.Server.Host = "0.0.0.0"
	}
	if c.Config.Server.Port == 0 {
		c.Config.Server.Port = 8080
	}
	if c.Config.Database.Type == "" {
		c.Config.Database.Type = "mysql"
	}
	if c.Config.Database.Port == 0 {
		c.Config.Database.Port = 3306
	}
	if c.Config.Database.Host == "" {
		c.Config.Database.Host = "localhost"
	}
	return &c, nil
}
