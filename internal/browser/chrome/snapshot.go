package chrome

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dayflower/export-ua-history/internal/browser"
	"github.com/ncruces/go-sqlite3"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

func ExportHistory(historyPath string, rng browser.ExportRange) ([]browser.HistoryEntry, error) {
	if _, err := os.Stat(historyPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("History database missing: %s", historyPath)
		}
		return nil, browser.FormatAccessError(historyPath, "failed to access History database", err)
	}

	snapshotPath, cleanup, err := snapshotHistory(historyPath)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	return queryEntries(snapshotPath, rng)
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

func vacuumIntoSQL(path string) string {
	return fmt.Sprintf("VACUUM INTO '%s'", browser.EscapeSQLiteString(path))
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
		if err := browser.CopyFileIfExists(pair[0], pair[1]); err != nil {
			return err
		}
	}
	return nil
}
