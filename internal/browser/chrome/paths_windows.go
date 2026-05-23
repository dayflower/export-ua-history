//go:build windows

package chrome

import (
	"path/filepath"

	"github.com/dayflower/export-ua-history/internal/browser/platform"
)

func defaultPlatformPaths(directories platform.Directories) (Paths, error) {
	userDataDir := filepath.Join(directories.LocalAppData, "Google", "Chrome", "User Data")
	return Paths{
		UserDataDir: userDataDir,
		LocalState:  filepath.Join(userDataDir, localStateName),
	}, nil
}
