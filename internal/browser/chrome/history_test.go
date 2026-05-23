package chrome

import (
	"testing"
	"time"
)

func TestLocalTimeToChromeMicros(t *testing.T) {
	ts := time.Date(2026, 5, 22, 9, 15, 23, 0, time.FixedZone("JST", 9*60*60))
	got := LocalTimeToChromeMicros(ts)
	back := ChromeMicrosToTime(got).In(ts.Location())
	if !back.Equal(ts) {
		t.Fatalf("round trip = %s, want %s", back, ts)
	}
}

func TestExtractDomain(t *testing.T) {
	if got := ExtractDomain("https://example.com/path"); got != "example.com" {
		t.Fatalf("ExtractDomain() = %q", got)
	}
	if got := ExtractDomain("://bad"); got != "" {
		t.Fatalf("ExtractDomain() = %q", got)
	}
}
