package chrome

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/dayflower/export-ua-history/internal/browser"
)

func TestLocalTimeToChromeMicros(t *testing.T) {
	ts := time.Date(2026, 5, 22, 9, 15, 23, 0, time.FixedZone("JST", 9*60*60))
	got := LocalTimeToChromeMicros(ts)
	back := ChromeMicrosToTime(got).In(ts.Location())
	if !back.Equal(ts) {
		t.Fatalf("round trip = %s, want %s", back, ts)
	}
}

func TestMapHistoryEntry(t *testing.T) {
	rng := browser.ExportRange{
		StartLocal: time.Date(2026, 5, 22, 0, 0, 0, 0, time.FixedZone("JST", 9*60*60)),
	}
	entry := mapHistoryEntry(historyRow{
		VisitTime:  LocalTimeToChromeMicros(time.Date(2026, 5, 22, 9, 15, 23, 0, rng.StartLocal.Location())),
		RawURL:     "https://example.com/path",
		Title:      "Example",
		VisitCount: 3,
	}, rng)
	if entry.Domain != "example.com" {
		t.Fatalf("Domain = %q", entry.Domain)
	}
	if entry.Timestamp.Format(time.RFC3339) != "2026-05-22T09:15:23+09:00" {
		t.Fatalf("Timestamp = %s", entry.Timestamp.Format(time.RFC3339))
	}
}

func TestFormatBrowserAccessError(t *testing.T) {
	err := browser.FormatAccessError("/tmp/test", "failed", errors.New("boom"))
	if !strings.Contains(err.Error(), "hint:") {
		t.Fatalf("unexpected error: %v", err)
	}
}
