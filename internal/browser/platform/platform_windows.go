//go:build windows

package platform

import "errors"

func currentDirectories(getenv EnvLookupFunc, _ HomeDirFunc) (Directories, error) {
	localAppData, err := resolveWindowsLocalAppData(getenv)
	if err != nil {
		return Directories{}, err
	}
	return Directories{LocalAppData: localAppData}, nil
}

func resolveWindowsLocalAppData(getenv EnvLookupFunc) (string, error) {
	localAppData := getenv("LOCALAPPDATA")
	if localAppData == "" {
		return "", errors.New("failed to resolve local app data directory: LOCALAPPDATA is not set")
	}
	return localAppData, nil
}
