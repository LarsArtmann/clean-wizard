package format

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
)

// Size formats bytes for human reading using IEC binary prefixes.
func Size(bytes int64) string {
	if bytes < 0 {
		return "0 B"
	}
	return humanize.IBytes(uint64(bytes))
}

// Duration formats duration for human reading.
func Duration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.0f ns", float64(d.Nanoseconds()))
	}
	if d < time.Second {
		return fmt.Sprintf("%.1f ms", float64(d.Nanoseconds())/1e6)
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1f s", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1f m", d.Minutes())
	}
	return fmt.Sprintf("%.1f h", d.Hours())
}

// Date formats date for human reading.
func Date(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	return t.Format("2006-01-02")
}

// DateTime formats date and time for human reading.
func DateTime(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	return t.Format("2006-01-02 15:04:05")
}

// Number formats number with thousand separators.
func Number(n int64) string {
	return humanize.Comma(n)
}
