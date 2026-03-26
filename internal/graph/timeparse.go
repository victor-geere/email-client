package graph

import "time"

var timeFormats = []string{
	time.RFC3339,
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05.0000000Z",
}

func parseTime(s string) (time.Time, error) {
	var lastErr error
	for _, fmt := range timeFormats {
		t, err := time.Parse(fmt, s)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}
	return time.Time{}, lastErr
}
