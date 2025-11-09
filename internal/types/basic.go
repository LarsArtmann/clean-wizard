package types

import (
	"context"
	"fmt"
	"time"
)

// RiskLevel represents the risk level of a cleaning operation
type RiskLevel string

const (
	RiskLow    RiskLevel = "low"
	RiskMedium RiskLevel = "medium"
	RiskHigh   RiskLevel = "high"
)

// ScanResult represents the result of scanning a system component
type LegacyScanResult struct {
	Name        string  `json:"name" yaml:"name"`
	SizeGB      float64 `json:"size_gb" yaml:"size_gb"`
	Description string  `json:"description" yaml:"description"`
	Cleanable   bool    `json:"cleanable" yaml:"cleanable"`
}

// LegacyScanResults represents all scan results
type LegacyScanResults struct {
	TotalSizeGB float64      `json:"total_size_gb" yaml:"total_size_gb"`
	Results     []ScanResult `json:"results" yaml:"results"`
	Timestamp   time.Time    `json:"timestamp" yaml:"timestamp"`
}

// CleanupOperation represents a single cleanup operation
type CleanupOperation struct {
	Name        string         `json:"name" yaml:"name"`
	Description string         `json:"description" yaml:"description"`
	RiskLevel   RiskLevel      `json:"risk_level" yaml:"risk_level"`
	Enabled     bool           `json:"enabled" yaml:"enabled"`
	Settings    map[string]any `json:"settings,omitempty" yaml:"settings,omitempty"`
}

// Profile represents a cleaning profile
type Profile struct {
	Name        string             `json:"name" yaml:"name"`
	Description string             `json:"description" yaml:"description"`
	Operations  []CleanupOperation `json:"operations" yaml:"operations"`
	CreatedAt   time.Time          `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" yaml:"updated_at"`
}

// Config represents the application configuration
type Config struct {
	Version      string             `json:"version" yaml:"version"`
	SafeMode     bool               `json:"safe_mode" yaml:"safe_mode"`
	DryRun       bool               `json:"dry_run" yaml:"dry_run"`
	Verbose      bool               `json:"verbose" yaml:"verbose"`
	Backup       bool               `json:"backup" yaml:"backup"`
	MaxDiskUsage int                `json:"max_disk_usage_percent" yaml:"max_disk_usage_percent"`
	Protected    []string           `json:"protected_paths" yaml:"protected_paths"`
	Profiles     map[string]Profile `json:"profiles" yaml:"profiles"`
}

// CleanupResult represents the result of a cleanup operation
type CleanupResult struct {
	Operation   string        `json:"operation" yaml:"operation"`
	Success     bool          `json:"success" yaml:"success"`
	SizeFreedGB float64       `json:"size_freed_gb" yaml:"size_freed_gb"`
	Error       error         `json:"error,omitempty" yaml:"error,omitempty"`
	Duration    time.Duration `json:"duration" yaml:"duration"`
}

// CleanupResults represents all cleanup results
type CleanupResults struct {
	TotalSizeFreedGB float64         `json:"total_size_freed_gb" yaml:"total_size_freed_gb"`
	Results          []CleanupResult `json:"results" yaml:"results"`
	StartTime        time.Time       `json:"start_time" yaml:"start_time"`
	EndTime          time.Time       `json:"end_time" yaml:"end_time"`
	Duration         time.Duration   `json:"duration" yaml:"duration"`
}

// Scanner interface for different system scanners
type Scanner interface {
	Scan(ctx context.Context) (*LegacyScanResults, error)
	Name() string
}

// Cleaner interface for different system cleaners
type Cleaner interface {
	Clean(ctx context.Context, operation CleanupOperation) (*CleanupResult, error)
	Name() string
	RiskLevel() RiskLevel
}

// String returns the string representation of RiskLevel
func (r RiskLevel) String() string {
	return string(r)
}

// Color returns the color representation of RiskLevel
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

// Icon returns the icon representation of RiskLevel
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
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}
