package domain

import (
	"fmt"
	"time"
)

// TODO: Make invalid states unrepresentable through strong typing

// ScanType represents different scanning domains
type ScanType string

const (
	ScanTypeNixStore ScanType = "nix_store"
	ScanTypeHomebrew ScanType = "homebrew"
	ScanTypeSystem   ScanType = "system"
	ScanTypeTemp     ScanType = "temp_files"
)

// IsValid validates ScanType
func (st ScanType) IsValid() bool {
	switch st {
	case ScanTypeNixStore, ScanTypeHomebrew, ScanTypeSystem, ScanTypeTemp:
		return true
	default:
		return false
	}
}

// RiskLevel represents operation risk with type safety
type RiskLevel string

const (
	RiskLow      RiskLevel = "LOW"
	RiskMedium   RiskLevel = "MEDIUM"
	RiskHigh     RiskLevel = "HIGH"
	RiskCritical RiskLevel = "CRITICAL"
)

// IsValid checks if risk level is valid
func (rl RiskLevel) IsValid() bool {
	switch rl {
	case RiskLow, RiskMedium, RiskHigh, RiskCritical:
		return true
	default:
		return false
	}
}

// String returns string representation
func (rl RiskLevel) String() string {
	switch rl {
	case RiskLow:
		return "LOW"
	case RiskMedium:
		return "MEDIUM"
	case RiskHigh:
		return "HIGH"
	case RiskCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// Icon returns emoji for risk level
func (rl RiskLevel) Icon() string {
	switch rl {
	case RiskLow:
		return "🟢"
	case RiskMedium:
		return "🟡"
	case RiskHigh:
		return "🟠"
	case RiskCritical:
		return "🔴"
	default:
		return "⚪"
	}
}

// NixGeneration represents Nix store generation
type NixGeneration struct {
	ID      int       `json:"id"`
	Path    string    `json:"path"`
	Date    time.Time `json:"date"`
	Current bool      `json:"current"`
}

// IsValid validates generation
func (g NixGeneration) IsValid() bool {
	return g.ID > 0 && g.Path != "" && !g.Date.IsZero()
}

// Validate returns errors for invalid generation
func (g NixGeneration) Validate() error {
	if g.ID <= 0 {
		return fmt.Errorf("Generation ID must be positive, got: %d", g.ID)
	}
	if g.Path == "" {
		return fmt.Errorf("Generation path cannot be empty")
	}
	if g.Date.IsZero() {
		return fmt.Errorf("Generation date cannot be zero")
	}
	return nil
}

// ScanRequest represents scanning command
type ScanRequest struct {
	Type      ScanType `json:"type"`
	Recursive bool     `json:"recursive"`
	Limit     int      `json:"limit"`
}

// ScanResult represents successful scan outcome
type ScanResult struct {
	TotalBytes   int64         `json:"total_bytes"`
	TotalItems   int           `json:"total_items"`
	ScannedPaths []string      `json:"scanned_paths"`
	ScanTime     time.Duration `json:"scan_time"`
	ScannedAt    time.Time     `json:"scanned_at"`
}

// ScanItem represents single scannable item
type ScanItem struct {
	Path         string    `json:"path"`
	Size         int64     `json:"size"`
	Type         ScanType  `json:"type"`
	LastAccessed time.Time `json:"last_accessed"`
	Protectable  bool      `json:"protectable"`
}

// CleanRequest represents cleaning command
type CleanRequest struct {
	Items    []ScanItem `json:"items"`
	Strategy string     `json:"strategy"` // "aggressive", "conservative", "dry-run"
}

// CleanResult represents successful clean outcome
type CleanResult struct {
	FreedBytes   int64         `json:"freed_bytes"`
	ItemsRemoved int           `json:"items_removed"`
	ItemsFailed  int           `json:"items_failed"`
	CleanTime    time.Duration `json:"clean_time"`
	CleanedAt    time.Time     `json:"cleaned_at"`
	Strategy     string        `json:"strategy"`
}

// IsValid checks if clean result is valid
func (cr CleanResult) IsValid() bool {
	return cr.FreedBytes >= 0 && cr.ItemsRemoved >= 0 && cr.CleanedAt.IsZero() == false
}

// Validate returns errors for invalid clean result
func (cr CleanResult) Validate() error {
	if cr.FreedBytes < 0 {
		return fmt.Errorf("FreedBytes cannot be negative, got: %d", cr.FreedBytes)
	}
	if cr.ItemsRemoved < 0 {
		return fmt.Errorf("ItemsRemoved cannot be negative, got: %d", cr.ItemsRemoved)
	}
	if cr.CleanedAt.IsZero() {
		return fmt.Errorf("CleanedAt cannot be zero")
	}
	return nil
}
