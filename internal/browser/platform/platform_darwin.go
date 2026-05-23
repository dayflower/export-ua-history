//go:build darwin

package platform

import "fmt"

func currentDirectories(_ EnvLookupFunc, userHomeDir HomeDirFunc) (Directories, error) {
	home, err := userHomeDir()
	if err != nil {
		return Directories{}, fmt.Errorf("failed to resolve home directory: %w", err)
	}
	return Directories{HomeDir: home}, nil
}
