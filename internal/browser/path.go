package browser

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

func ValidateRelativeProfilePath(profilePath string) (string, error) {
	if profilePath == "" {
		return "", errors.New("profile path must not be empty")
	}

	clean := filepath.Clean(profilePath)
	if filepath.IsAbs(clean) || clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid profile path %q", profilePath)
	}

	return clean, nil
}
