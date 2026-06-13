// Package config resolves and provides Sauron's runtime configuration.
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
	HomeDirectory string
}

// New resolves the Configuration from the environment.
func New() (Configuration, error) {
	home, err := GetHomeDirectory()
	if err != nil {
		return Configuration{}, err
	}
	return Configuration{HomeDirectory: home}, nil
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
