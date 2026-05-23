package cli

import (
	"strings"
	"testing"
	"time"
)

func TestParseExportCommand(t *testing.T) {
	now := time.Date(2026, 5, 23, 10, 0, 0, 0, time.FixedZone("JST", 9*60*60))

	result, err := Parse([]string{"export", "--date", "2026-05-22"}, now)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if result.Command != CommandExport {
		t.Fatalf("Command = %q, want %q", result.Command, CommandExport)
	}
	if got := result.Options.Range.StartLocal.Format(time.RFC3339); got != "2026-05-22T00:00:00+09:00" {
		t.Fatalf("StartLocal = %s", got)
	}
}

func TestParseImplicitExportCommand(t *testing.T) {
	now := time.Date(2026, 5, 23, 10, 0, 0, 0, time.FixedZone("JST", 9*60*60))

	result, err := Parse([]string{"--date", "2026-05-22"}, now)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if result.Command != CommandExport {
		t.Fatalf("Command = %q, want %q", result.Command, CommandExport)
	}
}

func TestKnownCommandsAreNotImplicitExport(t *testing.T) {
	now := time.Now()
	for _, args := range [][]string{{"list-profiles"}, {"version"}} {
		result, err := Parse(args, now)
		if err != nil {
			t.Fatalf("Parse(%v) error = %v", args, err)
		}
		if result.Command != args[0] {
			t.Fatalf("Parse(%v) command = %q", args, result.Command)
		}
	}
}

func TestHelpCommandRequestsHelp(t *testing.T) {
	_, err := Parse([]string{"help"}, time.Now())
	if err != ErrHelpRequested {
		t.Fatalf("Parse(help) error = %v", err)
	}
}

func TestHelpCommandForSubcommandRequestsHelp(t *testing.T) {
	_, err := Parse([]string{"help", "export"}, time.Now())
	if err != ErrHelpRequested {
		t.Fatalf("Parse(help export) error = %v", err)
	}
}

func TestHelpTextIncludesAllCommands(t *testing.T) {
	help := HelpText([]string{"help"})
	for _, snippet := range []string{
		"export-ua-history export [options]",
		"export-ua-history list-profiles",
		"export-ua-history version",
		"export-ua-history help [command]",
	} {
		if !strings.Contains(help, snippet) {
			t.Fatalf("missing %q in help text:\n%s", snippet, help)
		}
	}
}

func TestExportHelpMentionsRelatedCommands(t *testing.T) {
	help := HelpText([]string{"export", "--help"})
	for _, snippet := range []string{
		"export-ua-history list-profiles",
		"export-ua-history version",
		"export-ua-history help [command]",
	} {
		if !strings.Contains(help, snippet) {
			t.Fatalf("missing %q in export help:\n%s", snippet, help)
		}
	}
}

func TestUnknownCommand(t *testing.T) {
	_, err := Parse([]string{"unknown"}, time.Now())
	if err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDateValidation(t *testing.T) {
	now := time.Date(2026, 5, 23, 10, 0, 0, 0, time.FixedZone("JST", 9*60*60))
	_, err := Parse([]string{"export", "--date", "2026-05-22", "--start-date", "2026-05-22", "--end-date", "2026-05-23"}, now)
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIncompleteDateRange(t *testing.T) {
	_, err := Parse([]string{"export", "--start-date", "2026-05-22"}, time.Now())
	if err == nil || !strings.Contains(err.Error(), "must be supplied together") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTimeValidation(t *testing.T) {
	_, err := Parse([]string{"export", "--date", "2026-05-22", "--start-time", "13:00", "--end-time", "12:00"}, time.Now())
	if err == nil || !strings.Contains(err.Error(), "strictly earlier") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDisplayEndLocalInclusiveForWholeDay(t *testing.T) {
	now := time.Date(2026, 5, 23, 10, 0, 0, 0, time.FixedZone("JST", 9*60*60))
	result, err := Parse([]string{"export", "--date", "2026-05-22"}, now)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got := result.Options.Range.DisplayEndLocalInclusive.Format(time.RFC3339); got != "2026-05-22T23:59:59+09:00" {
		t.Fatalf("DisplayEndLocalInclusive = %s", got)
	}
}

func TestDisplayEndLocalInclusiveForTimeWindow(t *testing.T) {
	now := time.Date(2026, 5, 23, 10, 0, 0, 0, time.FixedZone("JST", 9*60*60))
	result, err := Parse([]string{"export", "--date", "2026-05-22", "--start-time", "12:00", "--end-time", "13:00"}, now)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got := result.Options.Range.DisplayEndLocalInclusive.Format(time.RFC3339); got != "2026-05-22T12:59:59+09:00" {
		t.Fatalf("DisplayEndLocalInclusive = %s", got)
	}
}
