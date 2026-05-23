package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/dayflower/export-ua-history/internal/browser"
)

func TestWriteReportOmitsTimezone(t *testing.T) {
	report := Report{
		Browser:      "chrome",
		StartDate:    "2026-05-22T00:00:00+09:00",
		EndDate:      "2026-05-22T23:59:59+09:00",
		TotalEntries: 1,
		Entries: []ReportEntry{{
			Timestamp:  "2026-05-22T09:15:23+09:00",
			URL:        "https://example.com",
			Title:      "Example",
			VisitCount: 3,
			Domain:     "example.com",
			Browser:    "chrome",
		}},
	}

	var buf bytes.Buffer
	if err := WriteReport(&buf, report); err != nil {
		t.Fatalf("WriteReport() error = %v", err)
	}

	got := buf.String()
	if strings.Contains(got, "timezone") {
		t.Fatalf("unexpected timezone field: %s", got)
	}
	if !strings.HasSuffix(got, "\n") {
		t.Fatalf("output missing trailing newline")
	}
}

func TestFromHistoryEntries(t *testing.T) {
	entries := []browser.HistoryEntry{{Browser: "chrome"}}
	got := FromHistoryEntries(entries)
	if len(got) != 1 || got[0].Browser != "chrome" {
		t.Fatalf("FromHistoryEntries() = %#v", got)
	}
}

func TestNewReport(t *testing.T) {
	rng := browser.ExportRange{
		StartLocal:               time.Date(2026, 5, 22, 0, 0, 0, 0, time.FixedZone("JST", 9*60*60)),
		DisplayEndLocalInclusive: time.Date(2026, 5, 22, 23, 59, 59, 0, time.FixedZone("JST", 9*60*60)),
	}
	report := NewReport("chrome", rng, []browser.HistoryEntry{{Browser: "chrome"}})
	if report.StartDate != "2026-05-22T00:00:00+09:00" {
		t.Fatalf("StartDate = %q", report.StartDate)
	}
	if report.EndDate != "2026-05-22T23:59:59+09:00" {
		t.Fatalf("EndDate = %q", report.EndDate)
	}
	if report.TotalEntries != 1 {
		t.Fatalf("TotalEntries = %d", report.TotalEntries)
	}
}
