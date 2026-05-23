package chrome

import (
	"database/sql"
	"fmt"

	"github.com/dayflower/export-ua-history/internal/browser"
)

type historyRow struct {
	VisitTime  int64
	RawURL     string
	Title      string
	VisitCount int
}

func queryEntries(snapshotPath string, rng browser.ExportRange) ([]browser.HistoryEntry, error) {
	db, err := sql.Open("sqlite3", browser.SQLiteReadOnlyURI(snapshotPath))
	if err != nil {
		return nil, fmt.Errorf("failed to open snapshot database: %w", err)
	}
	defer db.Close()

	const query = `
SELECT
  v.visit_time,
  u.url,
  COALESCE(u.title, ''),
  u.visit_count
FROM visits AS v
JOIN urls AS u ON u.id = v.url
WHERE v.visit_time >= ? AND v.visit_time < ?
ORDER BY v.visit_time ASC;
`

	startMicros := LocalTimeToChromeMicros(rng.StartLocal)
	endMicros := LocalTimeToChromeMicros(rng.EndLocalExclusive)

	rows, err := db.Query(query, startMicros, endMicros)
	if err != nil {
		return nil, fmt.Errorf("failed to query snapshot database: %w", err)
	}
	defer rows.Close()

	var entries []browser.HistoryEntry
	for rows.Next() {
		row, err := scanHistoryRow(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, mapHistoryEntry(row, rng))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while reading history rows: %w", err)
	}

	return entries, nil
}

func scanHistoryRow(rows *sql.Rows) (historyRow, error) {
	var row historyRow
	if err := rows.Scan(&row.VisitTime, &row.RawURL, &row.Title, &row.VisitCount); err != nil {
		return historyRow{}, fmt.Errorf("failed to scan history row: %w", err)
	}
	return row, nil
}

func mapHistoryEntry(row historyRow, rng browser.ExportRange) browser.HistoryEntry {
	timestamp := ChromeMicrosToTime(row.VisitTime).In(rng.StartLocal.Location())
	return browser.HistoryEntry{
		Timestamp:  timestamp,
		URL:        row.RawURL,
		Title:      row.Title,
		VisitCount: row.VisitCount,
		Domain:     browser.ExtractDomain(row.RawURL),
		Browser:    "chrome",
	}
}
