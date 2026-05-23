//go:build !darwin && !linux && !windows

package chrome

import (
	"fmt"
	"runtime"

	"github.com/dayflower/export-ua-history/internal/browser/platform"
)

func defaultPlatformPaths(_ platform.Directories) (Paths, error) {
	return Paths{}, fmt.Errorf("unsupported platform %q", runtime.GOOS)
}
