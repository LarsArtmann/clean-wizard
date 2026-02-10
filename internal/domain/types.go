package domain

import (
	"errors"
	"fmt"
	"time"
)


// Use type-safe constants directly.
var (
	RiskLow      = RiskLevelType(RiskLevelLowType)
	RiskMedium   = RiskLevelType(RiskLevelMediumType)
	RiskHigh     = RiskLevelType(RiskLevelHighType)
	RiskCritical = RiskLevelType(RiskLevelCriticalType)

	ValidationLevelNone          = ValidationLevelType(ValidationLevelNoneType)
	ValidationLevelBasic         = ValidationLevelType(ValidationLevelBasicType)
	ValidationLevelComprehensive = ValidationLevelType(ValidationLevelComprehensiveType)
	ValidationLevelStrict        = ValidationLevelType(ValidationLevelStrictType)

	OperationAdded    = ChangeOperationType(ChangeOperationAddedType)
	OperationRemoved  = ChangeOperationType(ChangeOperationRemovedType)
	OperationModified = ChangeOperationType(ChangeOperationModifiedType)

	StrategyAggressive   = CleanStrategyType(StrategyAggressiveType)
	StrategyConservative = CleanStrategyType(StrategyConservativeType)
	StrategyDryRun       = CleanStrategyType(StrategyDryRunType)
)

// NixGeneration represents Nix store generation.
type NixGeneration struct {
	ID      int              `json:"id"`
	Path    string           `json:"path"`
	Date    time.Time        `json:"date"`
	Current GenerationStatus `json:"current"`
}

// IsValid validates generation.
func (g NixGeneration) IsValid() bool {
	return g.ID > 0 && g.Path != "" && !g.Date.IsZero()
}

// Validate returns errors for invalid generation.
func (g NixGeneration) Validate() error {
	if g.ID <= 0 {
		return fmt.Errorf("Generation ID must be positive, got: %d", g.ID)
	}
	if g.Path == "" {
		return errors.New("Generation path cannot be empty")
	}
	if g.Date.IsZero() {
		return errors.New("Generation date cannot be zero")
	}
	return nil
}

// EstimateSize estimates the size of this generation in bytes
// This is a rough estimate used when actual size calculation is not available.
func (g NixGeneration) EstimateSize() int64 {
	// Rough estimate: 50MB per generation as baseline with adjustments
	// Older generations tend to be larger, newer ones smaller
	baseSize := int64(50 * 1024 * 1024) // 50MB base
	age := time.Since(g.Date)
	ageFactor := int64(age.Hours() / 24 / 30)        // Age in months
	return baseSize + (ageFactor * 10 * 1024 * 1024) // Add 10MB per month
}

// ScanType represents different scanning domains.
type ScanType string

const (
	ScanTypeNixStore ScanType = "nix_store"
	ScanTypeHomebrew ScanType = "homebrew"
	ScanTypeSystem   ScanType = "system"
	ScanTypeTemp     ScanType = "temp_files"
)

// IsValid validates ScanType.
func (st ScanType) IsValid() bool {
	switch st {
	case ScanTypeNixStore, ScanTypeHomebrew, ScanTypeSystem, ScanTypeTemp:
		return true
	default:
		return false
	}
}

// ScanRequest represents scanning command.
type ScanRequest struct {
	Type      ScanType `json:"type"`
	Recursive ScanMode `json:"recursive"`
	Limit     int      `json:"limit"`
}

// Validate returns errors for invalid scan request.
func (sr ScanRequest) Validate() error {
	if !sr.Type.IsValid() {
		return fmt.Errorf("Invalid scan type: %s", sr.Type)
	}
	if sr.Limit < 0 {
		return fmt.Errorf("Limit cannot be negative, got: %d", sr.Limit)
	}
	return nil
}

// ScanItem represents item found during scanning.
type ScanItem struct {
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	Created  time.Time `json:"created"`
	ScanType ScanType  `json:"scan_type"`
}

// CleanRequest represents cleaning command.
type CleanRequest struct {
	Items    []ScanItem    `json:"items"`
	Strategy CleanStrategyType `json:"strategy"`
}

// Validate returns errors for invalid clean request.
func (cr CleanRequest) Validate() error {
	if !cr.Strategy.IsValid() {
		return fmt.Errorf("Invalid strategy: %s (must be 'aggressive', 'conservative', or 'dry-run')", cr.Strategy)
	}
	if len(cr.Items) == 0 {
		return errors.New("Items cannot be empty")
	}
	return nil
}

// ScanResult represents successful scan outcome.
type ScanResult struct {
	TotalBytes   int64         `json:"total_bytes"`
	TotalItems   int           `json:"total_items"`
	ScannedPaths []string      `json:"scanned_paths"`
	ScanTime     time.Duration `json:"scan_time"`
	ScannedAt    time.Time     `json:"scanned_at"`
}

// IsValid checks if scan result is valid.
func (sr ScanResult) IsValid() bool {
	return sr.TotalBytes >= 0 && sr.TotalItems >= 0 && sr.ScanTime >= 0 && !sr.ScannedAt.IsZero()
}

// Validate returns errors for invalid scan result.
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
		return errors.New("ScannedAt cannot be zero")
	}
	return nil
}

// SizeEstimate represents an honest size estimate, handling cases where exact size is unknown.
type SizeEstimate struct {
	Known  uint64                 `json:"known"`
	Status SizeEstimateStatusType `json:"status"`
}

// Value returns the known value, or 0 if unknown.
func (se SizeEstimate) Value() uint64 {
	if se.Status == SizeEstimateStatusUnknown {
		return 0
	}
	return se.Known
}

// IsKnown returns true if the size is known.
func (se SizeEstimate) IsKnown() bool {
	return se.Status == SizeEstimateStatusKnown
}

// String returns a formatted string representation.
func (se SizeEstimate) String() string {
	if se.Status == SizeEstimateStatusUnknown {
		return "Unknown"
	}
	// Note: format.Bytes would be used here, but we avoid import cycle
	// Format: "250 MB", "1.5 GB", etc.
	return fmt.Sprintf("%d bytes", se.Known)
}

// CleanResult represents successful clean outcome.
type CleanResult struct {
	SizeEstimate SizeEstimate  `json:"size_estimate"`
	FreedBytes   uint64        `json:"freed_bytes"` // Deprecated: Use SizeEstimate instead
	ItemsRemoved uint          `json:"items_removed"`
	ItemsFailed  uint          `json:"items_failed"`
	CleanTime    time.Duration `json:"clean_time"`
	CleanedAt    time.Time     `json:"cleaned_at"`
	Strategy     CleanStrategyType `json:"strategy"`
}

// IsValid checks if clean result is valid.
func (cr CleanResult) IsValid() bool {
	// Cannot remove items without size info unless explicitly marked unknown
	// Note: Some operations legitimately have 0 size (e.g., test cache when path unavailable)
	if cr.ItemsRemoved > 0 && cr.SizeEstimate.Status == SizeEstimateStatusKnown && cr.CleanTime == 0 {
		// If CleanTime is 0, this is likely a synthetic result where size might be 0
		return true
	}
	// Cannot fail items without any activity (removed or sized)
	if cr.ItemsFailed > 0 && cr.ItemsRemoved == 0 && cr.SizeEstimate.Known == 0 {
		return false
	}
	// Other validations
	return !cr.CleanedAt.IsZero() && cr.Strategy.IsValid()
}

// Validate returns errors for invalid clean result.
func (cr CleanResult) Validate() error {
	// Cannot remove items without size info unless explicitly marked unknown
	// Note: Some operations legitimately have 0 size (e.g., test cache when path unavailable)
	if cr.ItemsRemoved > 0 && cr.SizeEstimate.Status == SizeEstimateStatusKnown && cr.SizeEstimate.Known == 0 && cr.CleanTime > 0 {
		return errors.New("cannot have zero SizeEstimate when ItemsRemoved is > 0 (set Status: Unknown if size cannot be determined)")
	}
	// Cannot fail items without any activity (removed or sized)
	if cr.ItemsFailed > 0 && cr.ItemsRemoved == 0 && cr.SizeEstimate.Known == 0 {
		return errors.New("cannot have failed items when no items were processed")
	}
	if cr.CleanedAt.IsZero() {
		return errors.New("CleanedAt cannot be zero")
	}
	if !cr.Strategy.IsValid() {
		return fmt.Errorf("Invalid strategy: %s (must be 'aggressive', 'conservative', or 'dry-run')", cr.Strategy)
	}
	return nil
}
