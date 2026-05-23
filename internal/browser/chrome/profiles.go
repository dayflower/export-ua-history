package chrome

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/dayflower/export-ua-history/internal/browser"
	"github.com/dayflower/export-ua-history/internal/browser/platform"
)

const (
	localStateName = "Local State"
	historyDBName  = "History"
)

type Paths struct {
	UserDataDir string
	LocalState  string
}

type localState struct {
	Profile struct {
		InfoCache map[string]struct {
			Name string `json:"name"`
		} `json:"info_cache"`
	} `json:"profile"`
}

func DefaultPaths() (Paths, error) {
	directories, err := platform.DefaultDirectories()
	if err != nil {
		return Paths{}, err
	}
	return defaultPlatformPaths(directories)
}

func LoadProfilesFromDefaultLocation() ([]browser.Profile, error) {
	localStatePath, err := LocalStatePath()
	if err != nil {
		return nil, err
	}
	return LoadProfiles(localStatePath)
}

func LoadProfiles(localStatePath string) ([]browser.Profile, error) {
	data, err := os.ReadFile(localStatePath)
	if err != nil {
		return nil, pathAccessError(localStatePath, "failed to read Chrome Local State", err)
	}

	var state localState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse Chrome Local State %q: %w", localStatePath, err)
	}

	profiles := make([]browser.Profile, 0, len(state.Profile.InfoCache))
	for path, info := range state.Profile.InfoCache {
		profiles = append(profiles, browser.Profile{Name: info.Name, Path: path})
	}

	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Path < profiles[j].Path
	})
	return profiles, nil
}

func UserDataDir() (string, error) {
	paths, err := DefaultPaths()
	if err != nil {
		return "", err
	}
	return paths.UserDataDir, nil
}

func LocalStatePath() (string, error) {
	paths, err := DefaultPaths()
	if err != nil {
		return "", err
	}
	return paths.LocalState, nil
}

func HistoryDBPath(profilePath string) (string, error) {
	userDataDir, err := UserDataDir()
	if err != nil {
		return "", err
	}
	clean, err := browser.ValidateRelativeProfilePath(profilePath)
	if err != nil {
		return "", err
	}

	return filepath.Join(userDataDir, clean, historyDBName), nil
}

func pathAccessError(path, action string, err error) error {
	return browser.FormatAccessError(path, action, err)
}
