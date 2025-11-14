package domain

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"strings"
	"time"
)

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
	return string(rl)
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

// IsHigherThan returns true if this risk level is higher than the comparison
func (rl RiskLevel) IsHigherThan(other RiskLevel) bool {
	riskOrder := map[RiskLevel]int{
		RiskLow:      1,
		RiskMedium:   2,
		RiskHigh:     3,
		RiskCritical: 4,
	}

	return riskOrder[rl] > riskOrder[other]
}

// IsHigherOrEqualThan returns true if this risk level is higher or equal than the comparison
func (rl RiskLevel) IsHigherOrEqualThan(other RiskLevel) bool {
	riskOrder := map[RiskLevel]int{
		RiskLow:      1,
		RiskMedium:   2,
		RiskHigh:     3,
		RiskCritical: 4,
	}

	return riskOrder[rl] >= riskOrder[other]
}

// MarshalYAML implements yaml.Marshaler interface
func (rl RiskLevel) MarshalYAML() (any, error) {
	return string(rl), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface
func (rl *RiskLevel) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}

	switch strings.ToUpper(s) {
	case "LOW":
		*rl = RiskLow
	case "MEDIUM":
		*rl = RiskMedium
	case "HIGH":
		*rl = RiskHigh
	case "CRITICAL":
		*rl = RiskCritical
	default:
		return fmt.Errorf("invalid risk level: %s", s)
	}
	return nil
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

// ScanRequest represents scanning command
type ScanRequest struct {
	Type      ScanType `json:"type"`
	Recursive bool     `json:"recursive"`
	Limit     int      `json:"limit"`
}

// Validate returns errors for invalid scan request
func (sr ScanRequest) Validate() error {
	if !sr.Type.IsValid() {
		return fmt.Errorf("Invalid scan type: %s", sr.Type)
	}
	if sr.Limit < 0 {
		return fmt.Errorf("Limit cannot be negative, got: %d", sr.Limit)
	}
	return nil
}

// ScanItem represents item found during scanning
type ScanItem struct {
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	Created  time.Time `json:"created"`
	ScanType ScanType  `json:"scan_type"`
}

// CleanRequest represents cleaning command
type CleanRequest struct {
	Items    []ScanItem `json:"items"`
	Strategy string     `json:"strategy"` // "aggressive", "conservative", "dry-run"
}

// Validate returns errors for invalid clean request
func (cr CleanRequest) Validate() error {
	if cr.Strategy != "aggressive" && cr.Strategy != "conservative" && cr.Strategy != "dry-run" {
		return fmt.Errorf("Invalid strategy: %s", cr.Strategy)
	}
	if len(cr.Items) == 0 {
		return fmt.Errorf("Items cannot be empty")
	}
	return nil
}

// ScanResult represents successful scan outcome
type ScanResult struct {
	TotalBytes   int64         `json:"total_bytes"`
	TotalItems   int           `json:"total_items"`
	ScannedPaths []string      `json:"scanned_paths"`
	ScanTime     time.Duration `json:"scan_time"`
	ScannedAt    time.Time     `json:"scanned_at"`
}

// IsValid checks if scan result is valid
func (sr ScanResult) IsValid() bool {
	return sr.TotalBytes >= 0 && sr.TotalItems >= 0 && sr.ScanTime >= 0 && !sr.ScannedAt.IsZero()
}

// Validate returns errors for invalid scan result
func (sr ScanResult) Validate() error {
	if sr.TotalBytes < 0 {
		return fmt.Errorf("TotalBytes cannot be negative, got: %d", sr.TotalBytes)
	}
	if sr.TotalItems < 0 {
		return fmt.Errorf("TotalItems cannot be negative, got: %d", sr.TotalItems)
	}
	if sr.ScanTime < 0 {
		return fmt.Errorf("ScanTime cannot be negative, got: %d", sr.ScanTime)
	}
	if sr.ScannedAt.IsZero() {
		return fmt.Errorf("ScannedAt cannot be zero")
	}
	return nil
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
