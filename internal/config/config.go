// Package config manages ignr configuration settings.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	_ "go.seanlatimer.dev/ignr/internal/xdginit"
)

const (
	configDirName  = "ignr"
	configFileName = "config.json"
)

type Config struct {
	DefaultOutput    string `json:"default_output"`
	UserTemplatePath string `json:"user_template_path"`
}

func GetConfigDir() (string, error) {
	return filepath.Join(xdg.ConfigHome, configDirName), nil
}

func GetConfigPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

func GetUserTemplatePath() (string, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return "", err
	}

	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(dir, "templates")
	if strings.TrimSpace(cfg.UserTemplatePath) != "" {
		path = cfg.UserTemplatePath
	}

	if err := os.MkdirAll(path, 0o755); err != nil {
		return "", fmt.Errorf("create user templates dir: %w", err)
	}
	return path, nil
}

func GetPresetsPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create config dir: %w", err)
	}

	path := filepath.Join(dir, "presets.yaml")
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("check presets file: %w", err)
		}
		if err := os.WriteFile(path, []byte("presets: []\n"), 0o644); err != nil {
			return "", fmt.Errorf("create presets file: %w", err)
		}
	}
	return path, nil
}

func LoadConfig() (Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

func SaveConfig(cfg Config) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}
