package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds app configuration.
type Config struct {
	Token   string `json:"token"`
	APIBase string `json:"api_base"`
	LogPath string `json:"log_path"`
}

const (
	defaultAPIBase = "https://api.github.com"
)

// Load returns config from file plus environment overrides.
func Load() (Config, error) {
	cfg := Config{
		APIBase: defaultAPIBase,
		LogPath: filepath.Join(defaultConfigDir(), "actions.log"),
	}

	if data, err := os.ReadFile(configPath()); err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("parse config: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return cfg, fmt.Errorf("read config: %w", err)
	}

	// Apply defaults if missing.
	if cfg.APIBase == "" {
		cfg.APIBase = defaultAPIBase
	}
	if cfg.LogPath == "" {
		cfg.LogPath = filepath.Join(defaultConfigDir(), "actions.log")
	}

	// Environment overrides.
	if envToken := os.Getenv("GITHUB_TOKEN"); envToken != "" {
		cfg.Token = envToken
	}
	if envBase := os.Getenv("GITHUB_API_BASE"); envBase != "" {
		cfg.APIBase = envBase
	}

	expandedLog, err := expandPath(cfg.LogPath)
	if err != nil {
		return cfg, fmt.Errorf("log path: %w", err)
	}
	cfg.LogPath = expandedLog

	return cfg, nil
}

// EnsureLogDir creates the directory for the log file if needed.
func EnsureLogDir(logPath string) error {
	dir := filepath.Dir(logPath)
	return os.MkdirAll(dir, 0o755)
}

// ConfigPath returns the default config file location.
func configPath() string {
	return filepath.Join(defaultConfigDir(), "config.json")
}

func defaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".github-fork-manager"
	}
	return filepath.Join(home, ".github-fork-manager")
}

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, strings.TrimPrefix(path, "~")), nil
	}
	return path, nil
}
