package app

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/dayflower/export-ua-history/internal/browser"
	"github.com/dayflower/export-ua-history/internal/output"
)

func TestRunnerRunExport(t *testing.T) {
	now := time.Date(2026, 5, 23, 10, 0, 0, 0, time.FixedZone("JST", 9*60*60))
	var gotHistoryPath string
	var gotReport output.Report

	runner := Runner{
		LoadProfiles: func() ([]browser.Profile, error) {
			return []browser.Profile{{Name: "Personal", Path: "Default"}}, nil
		},
		ResolveProfile: func(profiles []browser.Profile, profileName, profilePath, defaultProfilePath string) (browser.Profile, error) {
			if defaultProfilePath != "Default" {
				t.Fatalf("defaultProfilePath = %q", defaultProfilePath)
			}
			return profiles[0], nil
		},
		HistoryDBPath: func(profilePath string) (string, error) {
			return "/tmp/History", nil
		},
		ExportHistory: func(historyPath string, rng browser.ExportRange) ([]browser.HistoryEntry, error) {
			gotHistoryPath = historyPath
			return []browser.HistoryEntry{{
				Timestamp:  now,
				URL:        "https://example.com",
				Title:      "Example",
				VisitCount: 1,
				Domain:     "example.com",
				Browser:    "chrome",
			}}, nil
		},
		WriteProfiles: func(io.Writer, []browser.Profile) error { return nil },
		WriteReport: func(_ io.Writer, report output.Report) error {
			gotReport = report
			return nil
		},
		WriteReportFile: func(string, output.Report) error { return nil },
		FormatVersion:   output.FormatVersion,
		Version:         "test",
	}

	var stdout bytes.Buffer
	if err := runner.Run([]string{"export", "--date", "2026-05-22"}, &stdout, &stdout, now); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if gotHistoryPath != "/tmp/History" {
		t.Fatalf("historyPath = %q", gotHistoryPath)
	}
	if gotReport.TotalEntries != 1 || gotReport.Browser != "chrome" {
		t.Fatalf("report = %#v", gotReport)
	}
}

func TestRunnerRunExportProfileResolutionError(t *testing.T) {
	now := time.Date(2026, 5, 23, 10, 0, 0, 0, time.FixedZone("JST", 9*60*60))
	runner := Runner{
		LoadProfiles: func() ([]browser.Profile, error) {
			return []browser.Profile{{Name: "Personal", Path: "Default"}}, nil
		},
		ResolveProfile: func([]browser.Profile, string, string, string) (browser.Profile, error) {
			return browser.Profile{}, errors.New("profile not found")
		},
		HistoryDBPath:   func(string) (string, error) { return "", nil },
		ExportHistory:   func(string, browser.ExportRange) ([]browser.HistoryEntry, error) { return nil, nil },
		WriteProfiles:   func(io.Writer, []browser.Profile) error { return nil },
		WriteReport:     func(io.Writer, output.Report) error { return nil },
		WriteReportFile: func(string, output.Report) error { return nil },
		FormatVersion:   output.FormatVersion,
		Version:         "test",
	}

	err := runner.Run([]string{"export", "--date", "2026-05-22"}, &bytes.Buffer{}, &bytes.Buffer{}, now)
	if err == nil || !strings.Contains(err.Error(), "profile not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunnerRunExportHistoryMissing(t *testing.T) {
	now := time.Date(2026, 5, 23, 10, 0, 0, 0, time.FixedZone("JST", 9*60*60))
	runner := Runner{
		LoadProfiles: func() ([]browser.Profile, error) {
			return []browser.Profile{{Name: "Personal", Path: "Default"}}, nil
		},
		ResolveProfile: func(profiles []browser.Profile, profileName, profilePath, defaultProfilePath string) (browser.Profile, error) {
			return profiles[0], nil
		},
		HistoryDBPath: func(string) (string, error) {
			return "/tmp/History", nil
		},
		ExportHistory: func(string, browser.ExportRange) ([]browser.HistoryEntry, error) {
			return nil, errors.New("History database missing: /tmp/History")
		},
		WriteProfiles:   func(io.Writer, []browser.Profile) error { return nil },
		WriteReport:     func(io.Writer, output.Report) error { return nil },
		WriteReportFile: func(string, output.Report) error { return nil },
		FormatVersion:   output.FormatVersion,
		Version:         "test",
	}

	err := runner.Run([]string{"export", "--date", "2026-05-22"}, &bytes.Buffer{}, &bytes.Buffer{}, now)
	if err == nil || !strings.Contains(err.Error(), "History database missing") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunnerRunListProfilesEmpty(t *testing.T) {
	runner := Runner{
		LoadProfiles: func() ([]browser.Profile, error) { return nil, nil },
		ResolveProfile: func([]browser.Profile, string, string, string) (browser.Profile, error) {
			return browser.Profile{}, nil
		},
		HistoryDBPath:   func(string) (string, error) { return "", nil },
		ExportHistory:   func(string, browser.ExportRange) ([]browser.HistoryEntry, error) { return nil, nil },
		WriteProfiles:   func(io.Writer, []browser.Profile) error { return nil },
		WriteReport:     func(io.Writer, output.Report) error { return nil },
		WriteReportFile: func(string, output.Report) error { return nil },
		FormatVersion:   output.FormatVersion,
		Version:         "test",
	}

	err := runner.Run([]string{"list-profiles"}, &bytes.Buffer{}, &bytes.Buffer{}, time.Now())
	if err == nil || !strings.Contains(err.Error(), "no Chrome profiles found") {
		t.Fatalf("unexpected error: %v", err)
	}
}
