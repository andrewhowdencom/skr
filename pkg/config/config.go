package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	ConfigFileName = "config.yaml" // XDG style inside skr dir
	AltConfigName  = ".skr.yaml"   // Legacy/Local
)

var KnownAgents = map[string]func(home string) string{
	"standard":    func(home string) string { return filepath.Join(home, ".config", "agent", "skills") },
	"antigravity": func(home string) string { return filepath.Join(home, ".antigravity", "skills") },
	"roocode":     func(home string) string { return filepath.Join(home, ".roocode", "skills") },
	"gemini":      func(home string) string { return filepath.Join(home, ".gemini", "skills") },
	"claude":      func(home string) string { return filepath.Join(home, ".claude", "skills") },
}

type Config struct {
	Agents []string `yaml:"agents"`
	Skills []string `yaml:"skills"`
}

func (c *Config) Merge(other *Config) {
	if other == nil {
		return
	}

	// Merge skills (append unique?)
	c.Skills = append(c.Skills, other.Skills...)

	// Merge Agents (append unique)
	for _, agent := range other.Agents {
		found := false
		for _, existing := range c.Agents {
			if existing == agent {
				found = true
				break
			}
		}
		if !found {
			c.Agents = append(c.Agents, agent)
		}
	}
}

// FindConfigFile traverses upwards from startDir looking for .skr.yaml or config.yaml
func FindConfigFile(startDir string) (string, error) {
	dir := startDir
	for i := 0; i < 100; i++ {
		// Check for .skr.yaml
		legacyPath := filepath.Join(dir, AltConfigName)
		if _, err := os.Stat(legacyPath); err == nil {
			return legacyPath, nil
		}

		// Check for config.yaml (only if in strict skr dir? or just in root?)
		// The design says "XDG style inside skr dir" for global, but for local?
		// Assuming we just look for these files in the project roots.
		xdgPath := filepath.Join(dir, ConfigFileName)
		if _, err := os.Stat(xdgPath); err == nil {
			return xdgPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", os.ErrNotExist
}

// Load reads configuration from a specific file path
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		slog.Debug("config file not found", "path", path)
		return &Config{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config %s: %w", path, err)
	}

	return &cfg, nil
}

// LoadMerged loads global XDG config and merges it with local config found by traversing up from dir.
func LoadMerged(startDir string) (*Config, error) {
	// 1. Load Global (XDG)
	configDir, err := os.UserConfigDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config")
	}

	globalConfigPath := filepath.Join(configDir, "skr", ConfigFileName)
	globalCfg, err := Load(globalConfigPath)
	if err != nil {
		// Warn?
		slog.Debug("failed to load global config", "path", globalConfigPath, "error", err)
	}

	// 2. Load Local
	// Find config file traversing up
	if startDir == "" {
		startDir, _ = os.Getwd()
	}

	localConfigPath, err := FindConfigFile(startDir)
	if err == nil {
		localCfg, err := Load(localConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load local config: %w", err)
		}
		globalCfg.Merge(localCfg)
	} else {
		slog.Debug("no local config found in hierarchy", "startDir", startDir)
	}

	return globalCfg, nil
}

// Save persists the config to .skr.yaml in dir
// Save persists the config to .skr.yaml in dir
func (c *Config) Save(dir string) error {
	if dir == "" {
		d, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		dir = d
	}

	// Default to .skr.yaml for Save (usually local)
	configPath := filepath.Join(dir, AltConfigName)
	return c.SaveTo(configPath)
}

// SaveTo writes the config to a specific file path
func (c *Config) SaveTo(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config to %s: %w", path, err)
	}

	return nil
}
