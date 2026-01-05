package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Theme   string `yaml:"theme"`
	ZenMode bool   `yaml:"zen_mode"`
	DBPath  string `yaml:"db_path"`
}

func DefaultConfig() Config {
	return Config{
		Theme:   "default",
		ZenMode: false,
		DBPath:  "kata.db",
	}
}

func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	
	configDir := filepath.Join(homeDir, ".config", "kata")
	configFile := filepath.Join(configDir, "config.yaml")
	
	return configFile, nil
}

func Load() (Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return DefaultConfig(), err
	}

	// If config doesn't exist, return defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultConfig(), err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), err
	}

	// Validate and set defaults for empty fields
	if cfg.Theme == "" {
		cfg.Theme = "default"
	}
	if cfg.DBPath == "" {
		cfg.DBPath = "kata.db"
	}

	return cfg, nil
}

func Save(cfg Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (c *Config) SetTheme(theme string) error {
	c.Theme = theme
	return Save(*c)
}

func (c *Config) SetZenMode(enabled bool) error {
	c.ZenMode = enabled
	return Save(*c)
}
