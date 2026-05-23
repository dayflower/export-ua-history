package chrome

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dayflower/export-ua-history/internal/browser/platform"
)

func TestLoadProfiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Local State")
	content := `{"profile":{"info_cache":{"Profile 1":{"name":"Work"},"Default":{"name":"Personal"}}}}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	profiles, err := LoadProfiles(path)
	if err != nil {
		t.Fatalf("LoadProfiles() error = %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("len(profiles) = %d", len(profiles))
	}
	if profiles[0].Path != "Default" || profiles[1].Path != "Profile 1" {
		t.Fatalf("profiles order = %#v", profiles)
	}
}

func TestDefaultPlatformPathsFromDarwinDirectories(t *testing.T) {
	paths, err := defaultPlatformPaths(platform.Directories{HomeDir: "/Users/test"})
	if err != nil {
		t.Fatalf("defaultPlatformPaths() error = %v", err)
	}
	if paths.UserDataDir == "" || paths.LocalState == "" {
		t.Fatalf("unexpected empty paths: %#v", paths)
	}
}

func TestHistoryDBPathValidation(t *testing.T) {
	_, err := HistoryDBPath("../bad")
	if err == nil || !strings.Contains(err.Error(), "invalid profile path") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPathAccessErrorIsPlatformNeutral(t *testing.T) {
	err := pathAccessError("/tmp/test", "failed", errors.New("boom"))
	if strings.Contains(err.Error(), "macOS") {
		t.Fatalf("unexpected macOS-specific hint: %v", err)
	}
}
