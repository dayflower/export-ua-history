package chrome

import "time"

var chromeEpochStart = time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)

func LocalTimeToChromeMicros(t time.Time) int64 {
	return t.UTC().UnixMicro() - chromeEpochStart.UnixMicro()
}

func ChromeMicrosToTime(value int64) time.Time {
	return time.UnixMicro(chromeEpochStart.UnixMicro() + value).UTC()
}
