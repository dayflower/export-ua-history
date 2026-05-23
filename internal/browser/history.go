package browser

import "time"

type ExportRange struct {
	StartLocal               time.Time
	EndLocalExclusive        time.Time
	DisplayEndLocalInclusive time.Time
}

type HistoryEntry struct {
	Timestamp  time.Time
	URL        string
	Title      string
	VisitCount int
	Domain     string
	Browser    string
}
