package chrome

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dayflower/export-ua-history/internal/cli"
	"github.com/ncruces/go-sqlite3"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

var chromeEpochStart = time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)

type Entry struct {
	Timestamp  time.Time
	URL        string
	Title      string
	VisitCount int
	Domain     string
	Browser    string
}

func ExportHistory(historyPath string, rng cli.ResolvedRange) ([]Entry, error) {
	if _, err := os.Stat(historyPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("History database missing: %s", historyPath)
		}
		return nil, accessError(historyPath, "failed to access History database", err)
	}

	snapshotPath, cleanup, err := snapshotHistory(historyPath)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	return queryEntries(snapshotPath, rng)
}

func LocalTimeToChromeMicros(t time.Time) int64 {
	return t.UTC().UnixMicro() - chromeEpochStart.UnixMicro()
}

func ChromeMicrosToTime(value int64) time.Time {
	return time.UnixMicro(chromeEpochStart.UnixMicro() + value).UTC()
}

func ExtractDomain(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return parsed.Hostname()
}

func snapshotHistory(historyPath string) (string, func(), error) {
	tmpDir, err := os.MkdirTemp("", "export-ua-history-*")
	if err != nil {
		return "", nil, fmt.Errorf("snapshot creation failure: %w", err)
	}

	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}

	snapshotPath := filepath.Join(tmpDir, "History.snapshot")
	source, err := sqlite3.Open(historyPath)
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("snapshot creation failure: %w", err)
	}
	defer source.Close()

	if err := source.Exec(vacuumIntoSQL(snapshotPath)); err != nil {
		if isDatabaseLocked(err) {
			// Chrome often keeps History locked while running, so fall back to copying
			// the database and its WAL/SHM sidecars into a temporary snapshot set.
			if err := copyHistoryFiles(historyPath, snapshotPath); err != nil {
				cleanup()
				return "", nil, fmt.Errorf("snapshot creation failure: %w", err)
			}
			return snapshotPath, cleanup, nil
		}
		cleanup()
		return "", nil, fmt.Errorf("snapshot creation failure: %w", err)
	}

	return snapshotPath, cleanup, nil
}

func queryEntries(snapshotPath string, rng cli.ResolvedRange) ([]Entry, error) {
	db, err := sql.Open("sqlite3", sqliteReadOnlyURI(snapshotPath))
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

	var entries []Entry
	for rows.Next() {
		var (
			visitTime  int64
			rawURL     string
			title      string
			visitCount int
		)
		if err := rows.Scan(&visitTime, &rawURL, &title, &visitCount); err != nil {
			return nil, fmt.Errorf("failed to scan history row: %w", err)
		}

		timestamp := ChromeMicrosToTime(visitTime).In(rng.StartLocal.Location())
		entries = append(entries, Entry{
			Timestamp:  timestamp,
			URL:        rawURL,
			Title:      title,
			VisitCount: visitCount,
			Domain:     ExtractDomain(rawURL),
			Browser:    "chrome",
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while reading history rows: %w", err)
	}

	return entries, nil
}

func sqliteReadOnlyURI(path string) string {
	uri := &url.URL{Scheme: "file", Path: path}
	query := uri.Query()
	query.Set("mode", "ro")
	uri.RawQuery = query.Encode()
	return uri.String()
}

func vacuumIntoSQL(path string) string {
	return fmt.Sprintf("VACUUM INTO '%s'", escapeSQLiteString(path))
}

func escapeSQLiteString(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

func isDatabaseLocked(err error) bool {
	return err != nil && strings.Contains(err.Error(), "database is locked")
}

func copyHistoryFiles(historyPath, snapshotPath string) error {
	pairs := [][2]string{
		{historyPath, snapshotPath},
		{historyPath + "-wal", snapshotPath + "-wal"},
		{historyPath + "-shm", snapshotPath + "-shm"},
	}

	for _, pair := range pairs {
		if err := copyFileIfExists(pair[0], pair[1]); err != nil {
			return err
		}
	}
	return nil
}

func copyFileIfExists(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return accessError(src, "failed to copy Chrome history snapshot source", err)
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return accessError(src, "failed to stat Chrome history snapshot source", err)
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return fmt.Errorf("failed to create snapshot file %s: %w", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("failed to copy snapshot file %s: %w", src, err)
	}

	return nil
}

func accessError(path, action string, err error) error {
	return fmt.Errorf("%s: %s: %w\nhint: the invoking terminal or parent process may not have sufficient permission to access the browser profile directory", action, path, err)
}
