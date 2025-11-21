package format

import (
	"fmt"
	"time"
)

// Size formats bytes for human reading
func Size(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Duration formats duration for human reading
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

// Date formats date for human reading
func Date(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	return t.Format("2006-01-02")
}

// DateTime formats date and time for human reading
func DateTime(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	return t.Format("2006-01-02 15:04:05")
}

// Number formats number with thousand separators
func Number(n int64) string {
	s := fmt.Sprintf("%d", n)

	// Add commas for thousands
	var result []rune
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, r)
	}

	return string(result)
}
