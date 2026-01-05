package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Theme    string `yaml:"theme"`
	Language string `yaml:"language"`
	ZenMode  bool   `yaml:"zen_mode"`
	DBPath   string `yaml:"db_path"`
}

func GetDataDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".kata"), nil
}

func DefaultConfig() Config {
	dataDir, _ := GetDataDir()
	// Fallback to local if home dir fails
	dbPath := "kata.db"
	if dataDir != "" {
		dbPath = filepath.Join(dataDir, "kata.db")
	}

	return Config{
		Theme:    "default",
		Language: "go",
		ZenMode:  false,
		DBPath:   dbPath,
	}
}

func GetConfigPath() (string, error) {
	dataDir, err := GetDataDir()
	if err != nil {
		return "", err
	}
	
	configFile := filepath.Join(dataDir, "config.yaml")
	return configFile, nil
}

func Load() (Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return DefaultConfig(), err
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return DefaultConfig(), nil
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
	if cfg.Language == "" {
		cfg.Language = "go"
	}
	if cfg.DBPath == "" {
		def := DefaultConfig()
		cfg.DBPath = def.DBPath
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
