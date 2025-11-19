package domain

import (
	"fmt"
	"time"
)

// Backward compatibility aliases - delegate to type-safe enums
type (
	RiskLevel       = RiskLevelType
	ValidationLevel = ValidationLevelType
	ChangeOperation = ChangeOperationType
)

// Backward compatibility constants - point to type-safe enums
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
)

// CleanStrategy represents cleaning strategy with type safety
type CleanStrategy = CleanStrategyType

// Backward compatibility constants - delegate to type-safe enums
var (
	StrategyAggressive   = CleanStrategyType(StrategyAggressiveType)
	StrategyConservative = CleanStrategyType(StrategyConservativeType)
	StrategyDryRun       = CleanStrategyType(StrategyDryRunType)
)

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

// EstimateSize estimates the size of this generation in bytes
// This is a rough estimate used when actual size calculation is not available
func (g NixGeneration) EstimateSize() int64 {
	// Rough estimate: 50MB per generation as baseline with adjustments
	// Older generations tend to be larger, newer ones smaller
	baseSize := int64(50 * 1024 * 1024) // 50MB base
	age := time.Since(g.Date)
	ageFactor := int64(age.Hours() / 24 / 30)        // Age in months
	return baseSize + (ageFactor * 10 * 1024 * 1024) // Add 10MB per month
}

// DEPRECATED: Use ScanTypeType from type_safe_enums.go for type safety
// ScanType represents different scanning domains (LEGACY - use ScanTypeType)
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
	Items    []ScanItem    `json:"items"`
	Strategy CleanStrategy `json:"strategy"`
}

// Validate returns errors for invalid clean request
func (cr CleanRequest) Validate() error {
	if !cr.Strategy.IsValid() {
		return fmt.Errorf("Invalid strategy: %s (must be 'aggressive', 'conservative', or 'dry-run')", cr.Strategy)
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
	FreedBytes   uint64        `json:"freed_bytes"`
	ItemsRemoved uint          `json:"items_removed"`
	ItemsFailed  uint          `json:"items_failed"`
	CleanTime    time.Duration `json:"clean_time"`
	CleanedAt    time.Time     `json:"cleaned_at"`
	Strategy     CleanStrategy `json:"strategy"`
}

// IsValid checks if clean result is valid
func (cr CleanResult) IsValid() bool {
	// Cannot remove items without freeing bytes
	if cr.ItemsRemoved > 0 && cr.FreedBytes == 0 {
		return false
	}
	// Cannot fail items without any activity
	if cr.ItemsFailed > 0 && cr.ItemsRemoved == 0 && cr.FreedBytes == 0 {
		return false
	}
	// Other validations
	return cr.CleanedAt.IsZero() == false && cr.Strategy.IsValid()
}

// Validate returns errors for invalid clean result
func (cr CleanResult) Validate() error {
	// Cannot remove items without freeing bytes
	if cr.ItemsRemoved > 0 && cr.FreedBytes == 0 {
		return fmt.Errorf("cannot have zero FreedBytes when ItemsRemoved is > 0")
	}
	// Cannot fail items without any activity (removed or freed)
	if cr.ItemsFailed > 0 && cr.ItemsRemoved == 0 && cr.FreedBytes == 0 {
		return fmt.Errorf("cannot have failed items when no items were processed")
	}
	if cr.CleanedAt.IsZero() {
		return fmt.Errorf("CleanedAt cannot be zero")
	}
	if !cr.Strategy.IsValid() {
		return fmt.Errorf("Invalid strategy: %s (must be 'aggressive', 'conservative', or 'dry-run')", cr.Strategy)
	}
	return nil
}
