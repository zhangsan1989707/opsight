package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config holds all agent configuration.
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Collector CollectorConfig `yaml:"collector"`
	Logging   LoggingConfig   `yaml:"logging"`
}

// ServerConfig defines the connection to the Opsight backend.
type ServerConfig struct {
	URL    string `yaml:"url"`
	APIKey string `yaml:"api_key"`
}

// CollectorConfig controls metric collection behaviour.
type CollectorConfig struct {
	IntervalSeconds int  `yaml:"interval_seconds"`
	CPU             bool `yaml:"cpu"`
	Memory          bool `yaml:"memory"`
	Disk            bool `yaml:"disk"`
	Network         bool `yaml:"network"`
	Load            bool `yaml:"load"`
}

// LoggingConfig controls log output.
type LoggingConfig struct {
	Level string `yaml:"level"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Server: ServerConfig{
			URL:    "http://localhost:80",
			APIKey: "",
		},
		Collector: CollectorConfig{
			IntervalSeconds: 10,
			CPU:             true,
			Memory:          true,
			Disk:            true,
			Network:         true,
			Load:            true,
		},
		Logging: LoggingConfig{
			Level: "info",
		},
	}
}

// LoadConfig reads agent.yaml from the executable directory and applies
// environment variable overrides.
func LoadConfig() (Config, error) {
	cfg := DefaultConfig()

	// Determine config file path (same directory as executable).
	exePath, err := os.Executable()
	if err != nil {
		return cfg, fmt.Errorf("cannot determine executable path: %w", err)
	}
	configPath := filepath.Join(filepath.Dir(exePath), "agent.yaml")

	// Read and parse YAML if it exists.
	data, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return cfg, fmt.Errorf("read config file: %w", err)
		}
		// Config file doesn't exist, use defaults.
		fmt.Printf("[INFO] %s not found, using defaults and env vars\n", configPath)
	} else {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("parse config file: %w", err)
		}
	}

	// Environment variable overrides.
	if v := os.Getenv("OPSIGHT_SERVER_URL"); v != "" {
		cfg.Server.URL = v
	}
	if v := os.Getenv("OPSIGHT_API_KEY"); v != "" {
		cfg.Server.APIKey = v
	}
	if v := os.Getenv("OPSIGHT_INTERVAL"); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil && n > 0 {
			cfg.Collector.IntervalSeconds = n
		} else {
			fmt.Printf("[WARN] invalid OPSIGHT_INTERVAL=%q, using default\n", v)
		}
	}

	return cfg, nil
}
