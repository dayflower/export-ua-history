package browser

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractDomain(t *testing.T) {
	if got := ExtractDomain("https://example.com/path"); got != "example.com" {
		t.Fatalf("ExtractDomain() = %q", got)
	}
	if got := ExtractDomain("://bad"); got != "" {
		t.Fatalf("ExtractDomain() = %q", got)
	}
}

func TestResolveProfileByName(t *testing.T) {
	profile, err := ResolveProfile([]Profile{{Name: "Personal", Path: "Default"}}, "Personal", "", "Default")
	if err != nil {
		t.Fatalf("ResolveProfile() error = %v", err)
	}
	if profile.Path != "Default" {
		t.Fatalf("profile.Path = %q", profile.Path)
	}
}

func TestResolveProfileByPath(t *testing.T) {
	profile, err := ResolveProfile([]Profile{{Name: "Work", Path: "Profile 1"}}, "", "Profile 1", "Default")
	if err != nil {
		t.Fatalf("ResolveProfile() error = %v", err)
	}
	if profile.Name != "Work" {
		t.Fatalf("profile.Name = %q", profile.Name)
	}
}

func TestResolveProfileAmbiguousName(t *testing.T) {
	_, err := ResolveProfile([]Profile{{Name: "Same", Path: "Default"}, {Name: "Same", Path: "Profile 1"}}, "Same", "", "Default")
	if err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveDefaultProfileMissing(t *testing.T) {
	_, err := ResolveProfile([]Profile{{Name: "Work", Path: "Profile 1"}}, "", "", "Default")
	if err == nil || !strings.Contains(err.Error(), "Default profile not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRelativeProfilePath(t *testing.T) {
	clean, err := ValidateRelativeProfilePath("Profile 1")
	if err != nil {
		t.Fatalf("ValidateRelativeProfilePath() error = %v", err)
	}
	if clean != "Profile 1" {
		t.Fatalf("clean = %q", clean)
	}
}

func TestValidateRelativeProfilePathRejectsInvalid(t *testing.T) {
	for _, value := range []string{"", "..", "../x"} {
		if _, err := ValidateRelativeProfilePath(value); err == nil {
			t.Fatalf("expected error for %q", value)
		}
	}
}

func TestSQLiteReadOnlyURI(t *testing.T) {
	got := SQLiteReadOnlyURI("/tmp/test.db")
	if !strings.Contains(got, "mode=ro") {
		t.Fatalf("SQLiteReadOnlyURI() = %q", got)
	}
}

func TestEscapeSQLiteString(t *testing.T) {
	if got := EscapeSQLiteString("a'b"); got != "a''b" {
		t.Fatalf("EscapeSQLiteString() = %q", got)
	}
}

func TestCopyFileIfExists(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	if err := os.WriteFile(src, []byte("hello"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := CopyFileIfExists(src, dst); err != nil {
		t.Fatalf("CopyFileIfExists() error = %v", err)
	}
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("copied = %q", string(data))
	}
}

func TestCopyFileIfExistsMissingSource(t *testing.T) {
	if err := CopyFileIfExists("/tmp/does-not-exist-codex", "/tmp/unused"); err != nil {
		t.Fatalf("CopyFileIfExists() error = %v", err)
	}
}

func TestFormatAccessErrorHasHint(t *testing.T) {
	err := FormatAccessError("/tmp/test", "failed", errors.New("boom"))
	if !strings.Contains(err.Error(), "hint:") {
		t.Fatalf("unexpected error: %v", err)
	}
}
