package types

import "time"

// OperationResult is type-safe result for operations
type OperationResult struct {
	Success      bool          `json:"success"`
	FreedBytes   int64         `json:"freed_bytes"`
	Duration     time.Duration `json:"duration"`
	ErrorMessage string        `json:"error_message,omitempty"`
}

// ScanResult is type-safe result for scanning
type ScanResult struct {
	CleanableItems int           `json:"cleanable_items"`
	TotalBytes     int64         `json:"total_bytes"`
	ScanTime       time.Duration `json:"scan_time"`
	Items          []ScanItem    `json:"items"`
}

// ScanItem represents a single cleanable item
type ScanItem struct {
	Path         string    `json:"path"`
	Size         int64     `json:"size"`
	Type         CleanType `json:"type"`
	LastAccessed time.Time `json:"last_accessed"`
}

// CleanType is type-safe cleaning type
type CleanType string

const (
	CleanTypeTempFiles    CleanType = "temp_files"
	CleanTypePackageCache CleanType = "package_cache"
	CleanTypeNixStore     CleanType = "nix_store"
	CleanTypeHomebrew     CleanType = "homebrew"
)

// IsValid checks if CleanType is valid
func (ct CleanType) IsValid() bool {
	switch ct {
	case CleanTypeTempFiles, CleanTypePackageCache, CleanTypeNixStore, CleanTypeHomebrew:
		return true
	default:
		return false
	}
}

// String returns string representation
func (ct CleanType) String() string {
	return string(ct)
}
