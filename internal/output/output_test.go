package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/dayflower/export-ua-history/internal/browser/chrome"
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
	entries := []chrome.Entry{{Browser: "chrome"}}
	got := FromHistoryEntries(entries)
	if len(got) != 1 || got[0].Browser != "chrome" {
		t.Fatalf("FromHistoryEntries() = %#v", got)
	}
}
