package domain

import "time"

// TODO: Make invalid states unrepresentable through strong typing

// ScanType represents different scanning domains
type ScanType string

const (
	ScanTypeNixStore ScanType = "nix_store"
	ScanTypeHomebrew  ScanType = "homebrew"
	ScanTypeSystem    ScanType = "system"
	ScanTypeTemp      ScanType = "temp_files"
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
type RiskLevel int

const (
	RiskLow RiskLevel = iota
	RiskMedium
	RiskHigh
	RiskCritical
)

// IsValid checks if risk level is valid
func (rl RiskLevel) IsValid() bool {
	return rl >= RiskLow && rl <= RiskCritical
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

// ScanRequest represents scanning command
type ScanRequest struct {
	Type      ScanType `json:"type"`
	Recursive bool      `json:"recursive"`
	Limit     int       `json:"limit"`
}

// ScanResult represents successful scan outcome
type ScanResult struct {
	TotalBytes     int64         `json:"total_bytes"`
	TotalItems     int           `json:"total_items"`
	ScannedPaths   []string      `json:"scanned_paths"`
	ScanTime       time.Duration `json:"scan_time"`
	ScannedAt      time.Time     `json:"scanned_at"`
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
	FreedBytes    int64         `json:"freed_bytes"`
	ItemsRemoved  int            `json:"items_removed"`
	ItemsFailed   int            `json:"items_failed"`
	CleanTime     time.Duration  `json:"clean_time"`
	CleanedAt     time.Time      `json:"cleaned_at"`
	Strategy      string         `json:"strategy"`
}