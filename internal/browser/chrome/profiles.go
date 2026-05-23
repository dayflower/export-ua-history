package chrome

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

func ResolveProfile(profiles []browser.Profile, profileName, profilePath string) (browser.Profile, error) {
	if profileName != "" && profilePath != "" {
		return browser.Profile{}, errors.New("--profile and --profile-path are mutually exclusive")
	}

	if profilePath != "" {
		for _, profile := range profiles {
			if profile.Path == profilePath {
				return profile, nil
			}
		}
		return browser.Profile{}, fmt.Errorf("profile path %q not found", profilePath)
	}

	if profileName != "" {
		var matches []browser.Profile
		for _, profile := range profiles {
			if profile.Name == profileName {
				matches = append(matches, profile)
			}
		}
		switch len(matches) {
		case 0:
			return browser.Profile{}, fmt.Errorf("profile %q not found", profileName)
		case 1:
			return matches[0], nil
		default:
			return browser.Profile{}, fmt.Errorf("profile name %q is ambiguous; use --profile-path instead", profileName)
		}
	}

	for _, profile := range profiles {
		if profile.Path == "Default" {
			return profile, nil
		}
	}
	return browser.Profile{}, errors.New("Default profile not found")
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
	if profilePath == "" {
		return "", errors.New("profile path must not be empty")
	}

	clean := filepath.Clean(profilePath)
	if filepath.IsAbs(clean) || clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid profile path %q", profilePath)
	}

	return filepath.Join(userDataDir, clean, historyDBName), nil
}

func pathAccessError(path, action string, err error) error {
	return fmt.Errorf("%s: %s: %w\nhint: the invoking terminal or parent process may not have sufficient permission to access the browser profile directory", action, path, err)
}
