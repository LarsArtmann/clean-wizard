package types

import (
	"fmt"
	"time"
)

// DEPRECATED: Use internal/domain.RiskLevel instead
// TODO: Migrate all consumers and remove
type RiskLevel string

const (
	RiskLow    RiskLevel = "low"
	RiskMedium RiskLevel = "medium"
	RiskHigh   RiskLevel = "high"
)

// String returns string representation of RiskLevel
func (r RiskLevel) String() string {
	return string(r)
}

// Color returns color representation of RiskLevel
func (r RiskLevel) Color() string {
	switch r {
	case RiskLow:
		return "green"
	case RiskMedium:
		return "yellow"
	case RiskHigh:
		return "red"
	default:
		return "white"
	}
}

// Icon returns icon representation of RiskLevel
func (r RiskLevel) Icon() string {
	switch r {
	case RiskLow:
		return "✅"
	case RiskMedium:
		return "⚡"
	case RiskHigh:
		return "⚠️"
	default:
		return "❓"
	}
}

// FormatSize formats a size in GB to a human-readable string
// TODO: Move to internal/format package
func FormatSize(sizeGB float64) string {
	if sizeGB < 1 {
		return fmt.Sprintf("%.0f MB", sizeGB*1024)
	}
	if sizeGB < 1024 {
		return fmt.Sprintf("%.1f GB", sizeGB)
	}
	return fmt.Sprintf("%.1f TB", sizeGB/1024)
}

// FormatDuration formats a duration to a human-readable string
// TODO: Move to internal/format package
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}
