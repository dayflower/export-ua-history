//go:build linux

package platform

import (
	"path/filepath"
	"testing"
)

func TestResolveLinuxConfigHomeWithXDGConfigHome(t *testing.T) {
	configHome, err := resolveLinuxConfigHome(func(key string) string {
		if key == "XDG_CONFIG_HOME" {
			return "/tmp/xdg"
		}
		return ""
	}, func() (string, error) {
		return "/home/test", nil
	})
	if err != nil {
		t.Fatalf("resolveLinuxConfigHome() error = %v", err)
	}
	if configHome != "/tmp/xdg" {
		t.Fatalf("configHome = %q", configHome)
	}
}

func TestResolveLinuxConfigHomeFallback(t *testing.T) {
	configHome, err := resolveLinuxConfigHome(func(string) string { return "" }, func() (string, error) {
		return "/home/test", nil
	})
	if err != nil {
		t.Fatalf("resolveLinuxConfigHome() error = %v", err)
	}
	want := filepath.Join("/home/test", ".config")
	if configHome != want {
		t.Fatalf("configHome = %q, want %q", configHome, want)
	}
}
