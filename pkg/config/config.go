package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const ConfigFileName = ".skr.yaml"

type Config struct {
	Agent struct {
		Type string `yaml:"type"`
	} `yaml:"agent"`
	Skills []string `yaml:"skills"`
}

// Load looks for .skr.yaml in the directory dir (defaults to current dir)
func Load(dir string) (*Config, error) {
	if dir == "" {
		d, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
		dir = d
	}

	configPath := filepath.Join(dir, ConfigFileName)

	// Check if file exists, if not, return empty config but no error?
	// Or should we error? The user requested .skr.yaml allows us to specify config.
	// Let's return error if not found ONLY if called explicitly, but here we might want discovery.
	// For now, simple Load logic.

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return &Config{}, nil // Return empty default config if not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", ConfigFileName, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", ConfigFileName, err)
	}

	return &cfg, nil
}

// Save persists the config to .skr.yaml in dir
func (c *Config) Save(dir string) error {
	if dir == "" {
		d, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		dir = d
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := filepath.Join(dir, ConfigFileName)
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", ConfigFileName, err)
	}

	return nil
}
