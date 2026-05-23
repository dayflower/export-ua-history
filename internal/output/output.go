package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/dayflower/export-ua-history/internal/browser"
	"github.com/dayflower/export-ua-history/internal/browser/chrome"
)

type Report struct {
	Browser      string        `json:"browser"`
	StartDate    string        `json:"start_date"`
	EndDate      string        `json:"end_date"`
	TotalEntries int           `json:"total_entries"`
	Entries      []ReportEntry `json:"entries"`
}

type ReportEntry struct {
	Timestamp  string `json:"timestamp"`
	URL        string `json:"url"`
	Title      string `json:"title"`
	VisitCount int    `json:"visit_count"`
	Domain     string `json:"domain"`
	Browser    string `json:"browser"`
}

func FromHistoryEntries(entries []chrome.Entry) []ReportEntry {
	reportEntries := make([]ReportEntry, 0, len(entries))
	for _, entry := range entries {
		reportEntries = append(reportEntries, ReportEntry{
			Timestamp:  entry.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			URL:        entry.URL,
			Title:      entry.Title,
			VisitCount: entry.VisitCount,
			Domain:     entry.Domain,
			Browser:    entry.Browser,
		})
	}
	return reportEntries
}

func WriteReport(w io.Writer, report Report) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode JSON output: %w", err)
	}
	data = append(data, '\n')
	_, err = w.Write(data)
	return err
}

func WriteReportFile(path string, report Report) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("output file creation failure: %w", err)
	}
	defer file.Close()

	return WriteReport(file, report)
}

func WriteProfiles(w io.Writer, profiles []browser.Profile) error {
	for _, profile := range profiles {
		if _, err := fmt.Fprintf(w, "%s\t%s\n", profile.Name, profile.Path); err != nil {
			return err
		}
	}
	return nil
}

func FormatVersion(name, version string) string {
	return fmt.Sprintf("%s %s", name, version)
}
