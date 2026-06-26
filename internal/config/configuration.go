package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	envHome        = "SAURON_HOME"
	defaultHomeDir = ".sauron"
)

// Configuration is Sauron's resolved runtime configuration.
type Configuration struct {
	// HomeDirectory is Sauron's own state root ($SAURON_HOME or ~/.sauron).
	HomeDirectory string
	// UserHomeDirectory is the user's real home, the root the provider artifact
	// directories (.claude, .zencoder) live under.
	UserHomeDirectory string
}

// New resolves the Configuration from the environment.
func New() (Configuration, error) {
	home, err := GetHomeDirectory()
	if err != nil {
		return Configuration{}, err
	}
	userHome, err := os.UserHomeDir()
	if err != nil {
		return Configuration{}, fmt.Errorf("resolve user home directory: %w", err)
	}
	return Configuration{HomeDirectory: home, UserHomeDirectory: userHome}, nil
}

// GetHomeDirectory resolves Sauron's home: $SAURON_HOME when set, else ~/.sauron.
func GetHomeDirectory() (string, error) {
	home := os.Getenv(envHome)
	if home != "" {
		return home, nil
	}
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(userHome, defaultHomeDir), nil
}
