//go:build darwin

package chrome

import (
	"path/filepath"

	"github.com/dayflower/export-ua-history/internal/browser/platform"
)

func defaultPlatformPaths(directories platform.Directories) (Paths, error) {
	userDataDir := filepath.Join(directories.HomeDir, "Library", "Application Support", "Google", "Chrome")
	return Paths{
		UserDataDir: userDataDir,
		LocalState:  filepath.Join(userDataDir, localStateName),
	}, nil
}
