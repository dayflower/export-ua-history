package browser

import (
	"errors"
	"fmt"
)

func ResolveProfile(profiles []Profile, profileName, profilePath, defaultProfilePath string) (Profile, error) {
	if profileName != "" && profilePath != "" {
		return Profile{}, errors.New("--profile and --profile-path are mutually exclusive")
	}

	if profilePath != "" {
		for _, profile := range profiles {
			if profile.Path == profilePath {
				return profile, nil
			}
		}
		return Profile{}, fmt.Errorf("profile path %q not found", profilePath)
	}

	if profileName != "" {
		var matches []Profile
		for _, profile := range profiles {
			if profile.Name == profileName {
				matches = append(matches, profile)
			}
		}
		switch len(matches) {
		case 0:
			return Profile{}, fmt.Errorf("profile %q not found", profileName)
		case 1:
			return matches[0], nil
		default:
			return Profile{}, fmt.Errorf("profile name %q is ambiguous; use --profile-path instead", profileName)
		}
	}

	for _, profile := range profiles {
		if profile.Path == defaultProfilePath {
			return profile, nil
		}
	}
	return Profile{}, fmt.Errorf("%s profile not found", defaultProfilePath)
}
