package domain

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// FormatDuration converts time.Duration to normalized string format
func FormatDuration(d time.Duration) string {
	if d == 0 {
		return "0"
	}

	// Normalize to the most appropriate unit
	if d < time.Minute {
		return d.String()
	}

	if d < time.Hour {
		minutes := int(d.Minutes())
		return fmt.Sprintf("%dm", minutes)
	}

	if d < 24*time.Hour {
		hours := int(d.Hours())
		return fmt.Sprintf("%dh", hours)
	}

	// For durations >= 24h, use days
	days := int(d.Hours()) / 24
	remainingHours := int(d.Hours()) % 24

	if remainingHours == 0 {
		return fmt.Sprintf("%dd", days)
	}
	return fmt.Sprintf("%dd%dh", days, remainingHours)
}

// ParseCustomDuration parses human-readable duration formats like "7d", "24h", "30m"
// and converts them to Go time.Duration
func ParseCustomDuration(durationStr string) (time.Duration, error) {
	// Trim whitespace first
	durationStr = strings.TrimSpace(durationStr)

	if durationStr == "" {
		return 0, fmt.Errorf("empty duration string")
	}

	// Check if it's already a valid Go duration
	if goDuration, err := time.ParseDuration(durationStr); err == nil {
		return goDuration, nil
	}

	// Custom parsing for "d" (days) unit
	if strings.HasSuffix(durationStr, "d") {
		return parseDaysDuration(durationStr)
	}

	// Try other custom formats
	return parseComplexDuration(durationStr)
}

// parseDaysDuration parses duration with "d" suffix
func parseDaysDuration(durationStr string) (time.Duration, error) {
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)d$`)
	matches := re.FindStringSubmatch(durationStr)
	if len(matches) != 2 {
		return 0, fmt.Errorf("invalid days duration format: %s", durationStr)
	}

	days, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid days value: %v", err)
	}

	// Convert days to hours (24 hours per day)
	hours := days * 24
	goDurationStr := fmt.Sprintf("%.0fh", hours)

	return time.ParseDuration(goDurationStr)
}

// parseComplexDuration parses combined durations like "7d12h"
func parseComplexDuration(durationStr string) (time.Duration, error) {
	// This could be extended to support complex formats
	// For now, focus on the most common case: days only
	return 0, fmt.Errorf("unsupported duration format: %s (supported formats: '7d', '24h', '30m', '1h30m')", durationStr)
}

// ValidateCustomDuration checks if a duration string is valid for this system
func ValidateCustomDuration(durationStr string) error {
	if durationStr == "" {
		return fmt.Errorf("duration cannot be empty")
	}

	_, err := ParseCustomDuration(durationStr)
	return err
}
