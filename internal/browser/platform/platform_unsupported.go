//go:build !darwin && !linux && !windows

package platform

import (
	"fmt"
	"runtime"
)

func currentDirectories(_ EnvLookupFunc, _ HomeDirFunc) (Directories, error) {
	return Directories{}, fmt.Errorf("unsupported platform %q", runtime.GOOS)
}
