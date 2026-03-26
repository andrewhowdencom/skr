package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	LockFileName    = "config.lock"
	AltLockFileName = ".skr.lock"
)

type LockFile struct {
	Skills map[string]string `yaml:"skills"`
}

func NewLockFile() *LockFile {
	return &LockFile{
		Skills: make(map[string]string),
	}
}

// LoadLock reads the lockfile from a specific path
func LoadLock(path string) (*LockFile, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		slog.Debug("lock file not found", "path", path)
		return NewLockFile(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read lock file %s: %w", path, err)
	}

	var l LockFile
	if err := yaml.Unmarshal(data, &l); err != nil {
		return nil, fmt.Errorf("failed to parse lock file %s: %w", path, err)
	}

	if l.Skills == nil {
		l.Skills = make(map[string]string)
	}

	return &l, nil
}

// FindLockFile traverses upwards from startDir looking for .skr.lock or config.lock
func FindLockFile(startDir string) (string, error) {
	dir := startDir
	for i := 0; i < 100; i++ {
		legacyPath := filepath.Join(dir, AltLockFileName)
		if _, err := os.Stat(legacyPath); err == nil {
			return legacyPath, nil
		}

		xdgPath := filepath.Join(dir, LockFileName)
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

// Save persists the lockfile to the given path
func (l *LockFile) SaveTo(path string) error {
	data, err := yaml.Marshal(l)
	if err != nil {
		return fmt.Errorf("failed to marshal lock file: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write lock file to %s: %w", path, err)
	}

	return nil
}
