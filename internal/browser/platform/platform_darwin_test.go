//go:build darwin

package platform

import (
	"errors"
	"strings"
	"testing"
)

func TestCurrentDirectoriesDarwin(t *testing.T) {
	dirs, err := currentDirectories(func(string) string { return "" }, func() (string, error) {
		return "/Users/test", nil
	})
	if err != nil {
		t.Fatalf("currentDirectories() error = %v", err)
	}
	if dirs.HomeDir != "/Users/test" {
		t.Fatalf("HomeDir = %q", dirs.HomeDir)
	}
}

func TestCurrentDirectoriesDarwinHomeDirError(t *testing.T) {
	_, err := currentDirectories(func(string) string { return "" }, func() (string, error) {
		return "", errors.New("boom")
	})
	if err == nil || !strings.Contains(err.Error(), "failed to resolve home directory") {
		t.Fatalf("unexpected error: %v", err)
	}
}
