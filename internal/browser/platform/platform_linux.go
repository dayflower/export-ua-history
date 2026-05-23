//go:build linux

package platform

import (
	"fmt"
	"path/filepath"
)

func currentDirectories(getenv EnvLookupFunc, userHomeDir HomeDirFunc) (Directories, error) {
	configHome, err := resolveLinuxConfigHome(getenv, userHomeDir)
	if err != nil {
		return Directories{}, err
	}
	return Directories{ConfigHome: configHome}, nil
}

func resolveLinuxConfigHome(getenv EnvLookupFunc, userHomeDir HomeDirFunc) (string, error) {
	configHome := getenv("XDG_CONFIG_HOME")
	if configHome != "" {
		return configHome, nil
	}

	home, err := userHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve home directory: %w", err)
	}
	return filepath.Join(home, ".config"), nil
}
