//go:build windows

package platform

import (
	"strings"
	"testing"
)

func TestResolveWindowsLocalAppData(t *testing.T) {
	localAppData, err := resolveWindowsLocalAppData(func(key string) string {
		if key == "LOCALAPPDATA" {
			return `C:\Users\Test\AppData\Local`
		}
		return ""
	})
	if err != nil {
		t.Fatalf("resolveWindowsLocalAppData() error = %v", err)
	}
	if localAppData != `C:\Users\Test\AppData\Local` {
		t.Fatalf("localAppData = %q", localAppData)
	}
}

func TestResolveWindowsLocalAppDataError(t *testing.T) {
	_, err := resolveWindowsLocalAppData(func(string) string { return "" })
	if err == nil || !strings.Contains(err.Error(), "LOCALAPPDATA") {
		t.Fatalf("unexpected error: %v", err)
	}
}
