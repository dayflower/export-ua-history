//go:build linux

package chrome

import (
	"path/filepath"

	"github.com/dayflower/export-ua-history/internal/browser/platform"
)

func defaultPlatformPaths(directories platform.Directories) (Paths, error) {
	userDataDir := filepath.Join(directories.ConfigHome, "google-chrome")
	return Paths{
		UserDataDir: userDataDir,
		LocalState:  filepath.Join(userDataDir, localStateName),
	}, nil
}
